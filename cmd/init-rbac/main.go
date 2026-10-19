package main

import (
	"fmt"
	"log"
	"os"

	"octoops/internal/config"
	"octoops/internal/infra/postgres"
	"octoops/internal/model/rbac"
	"octoops/internal/pkg/crypto"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// db 本次初始化使用的数据库句柄，在事务中执行，任一步骤失败整体回滚
var db *gorm.DB

// findPermissionByCode 按 code 查询权限；第二个返回值为是否存在，查询出错时返回 error
func findPermissionByCode(code string) (model.Permission, bool, error) {
	var perm model.Permission
	if err := db.Where("code = ?", code).Limit(1).Find(&perm).Error; err != nil {
		return model.Permission{}, false, fmt.Errorf("查询权限 %s 失败: %w", code, err)
	}
	if perm.Code == "" {
		return model.Permission{}, false, nil
	}
	return perm, true, nil
}

func main() {
	// 初始化配置和数据库
	if err := config.InitConfig(); err != nil {
		log.Fatalf("初始化配置失败: %v", err)
	}
	if err := postgres.Init(config.PostgresDSN); err != nil {
		log.Fatalf("初始化数据库失败: %v", err)
	}
	if err := postgres.Migrate(); err != nil {
		log.Fatalf("数据库迁移失败: %v", err)
	}

	log.Println("开始初始化RBAC系统...")

	tx := postgres.DB.Begin()
	if tx.Error != nil {
		log.Fatalf("开启事务失败: %v", tx.Error)
	}
	db = tx
	if err := initRBAC(); err != nil {
		tx.Rollback()
		log.Fatalf("初始化RBAC系统失败，已回滚: %v", err)
	}
	if err := tx.Commit().Error; err != nil {
		log.Fatalf("提交事务失败: %v", err)
	}

	log.Println("RBAC系统初始化完成！")
}

// initRBAC 只新增和更新，不删除任何已有数据：权限树同步 -> 默认角色及授权补齐 -> 默认管理员
func initRBAC() error {
	// 1. 创建或更新权限树
	permissions, err := createMenuLikePermissions()
	if err != nil {
		return fmt.Errorf("同步权限失败: %w", err)
	}

	// 2. 创建默认角色，并增量补齐角色权限
	roles, err := createDefaultRoles(permissions)
	if err != nil {
		return fmt.Errorf("初始化角色失败: %w", err)
	}

	// 3. 创建默认管理员（已存在则跳过，不覆盖密码、不自动提权）
	if err := createDefaultAdmin(); err != nil {
		return fmt.Errorf("初始化默认管理员失败: %w", err)
	}

	log.Printf("RBAC 初始化完成：同步了 %d 个权限，检查了 %d 个默认角色", len(permissions), len(roles))
	return nil
}

// createMenuLikePermissions 创建与菜单结构一致的权限树
func createMenuLikePermissions() (map[string]*model.Permission, error) {
	// 1. 一级菜单
	menuDefs := []struct {
		Name, Code, Desc, Path string
		OrderNum               int
	}{
		{"阿里云", "aliyun", "阿里云相关", "/aliyun", 1},
		{"Seatunnel", "seatunnel", "Seatunnel相关", "/seatunnel", 2},
		{"任务中心", "task", "任务管理相关", "/task", 3},
		{"告警管理", "notify", "消息通知相关", "/alert", 4},
		{"系统管理", "rbac", "系统管理相关", "/rbac", 5},
		{"系统设置", "system", "系统设置相关", "/system", 6},
	}
	menuMap := make(map[string]*model.Permission)
	for _, m := range menuDefs {
		perm, err := upsertPermission(model.Permission{
			Name:        m.Name,
			Code:        m.Code,
			Description: m.Desc,
			Type:        "menu",
			Path:        m.Path,
			OrderNum:    m.OrderNum,
			Method:      "",
			Status:      1,
			ParentID:    0,
		})
		if err != nil {
			return nil, err
		}
		menuMap[m.Code] = &perm
	}

	// 2. 二级菜单
	subMenuDefs := []struct {
		Name, Code, Desc, Parent, Path string
		OrderNum                       int
	}{
		// 阿里云
		{"ECS安全组", "aliyun:ecs_sg", "ECS安全组", "aliyun", "/aliyun/ecs-security-group", 1},
		// ETL调度
		{"实时数据集成", "etl:stream", "实时数据集成", "seatunnel", "/seatunnel/stream", 1},
		{"离线数据集成", "etl:batch", "离线数据集成", "seatunnel", "/seatunnel/batch", 2},
		// 任务管理
		{"调度器", "task:scheduler", "调度器", "task", "/task/scheduler", 1},
		{"任务日志", "task:log", "任务日志", "task", "/task/log", 2},
		// 消息通知
		{"告警组管理", "notify:group", "告警组管理", "notify", "/alert/group", 1},
		{"告警模板", "notify:template", "告警模板", "notify", "/alert/template", 2},
		{"告警渠道", "notify:channel", "告警渠道", "notify", "/alert/channel", 3},
		// 权限管理
		{"用户管理", "rbac:user", "用户管理", "rbac", "/rbac/user", 1},
		{"角色管理", "rbac:role", "角色管理", "rbac", "/rbac/role", 2},
		{"权限管理", "rbac:permission", "权限管理", "rbac", "/rbac/permission", 3},
		// 系统设置
		{"基本设置", "system:setting", "基本设置", "system", "/system/setting", 1},
	}
	subMenuMap := make(map[string]*model.Permission)
	for _, s := range subMenuDefs {
		parent, ok := menuMap[s.Parent]
		if !ok {
			return nil, fmt.Errorf("二级菜单 %s 的父级 %s 不存在", s.Code, s.Parent)
		}
		perm, err := upsertPermission(model.Permission{
			Name:        s.Name,
			Code:        s.Code,
			Description: s.Desc,
			Type:        "menu",
			Path:        s.Path,
			OrderNum:    s.OrderNum,
			Method:      "",
			Status:      1,
			ParentID:    parent.ID,
		})
		if err != nil {
			return nil, err
		}
		subMenuMap[s.Code] = &perm
	}

	// 3. 三级操作权限（以系统管理为例，可按需扩展到其他菜单）
	permissions := []model.Permission{
		// 阿里云
		// ECS安全组权限
		{Name: "查看", Code: "aliyun:ecs_sg:read", Description: "查看ECS安全组", Type: "api", Path: "/api/aliyun/ecs-sg", Method: "GET", Status: 1, ParentID: subMenuMap["aliyun:ecs_sg"].ID},
		{Name: "创建", Code: "aliyun:ecs_sg:create", Description: "创建ECS安全组", Type: "api", Path: "/api/aliyun/ecs-sg", Method: "POST", Status: 1, ParentID: subMenuMap["aliyun:ecs_sg"].ID},
		{Name: "更新", Code: "aliyun:ecs_sg:update", Description: "更新ECS安全组", Type: "api", Path: "/api/aliyun/ecs-sg/:id", Method: "PUT", Status: 1, ParentID: subMenuMap["aliyun:ecs_sg"].ID},
		{Name: "删除", Code: "aliyun:ecs_sg:delete", Description: "删除ECS安全组", Type: "api", Path: "/api/aliyun/ecs-sg/:id", Method: "DELETE", Status: 1, ParentID: subMenuMap["aliyun:ecs_sg"].ID},
		{Name: "同步", Code: "aliyun:ecs_sg:sync", Description: "同步ECS安全组", Type: "api", Path: "/api/aliyun/ecs-sg/:id/sync", Method: "POST", Status: 1, ParentID: subMenuMap["aliyun:ecs_sg"].ID},
		// Seatunnel
		// stream权限
		{Name: "查看", Code: "etl:stream:read", Description: "查看实时数据集成列表/详情", Type: "api", Path: "/api/seatunnel/stream", Method: "GET", Status: 1, ParentID: subMenuMap["etl:stream"].ID},
		{Name: "创建", Code: "etl:stream:create", Description: "创建实时数据集成", Type: "api", Path: "/api/seatunnel/stream", Method: "POST", Status: 1, ParentID: subMenuMap["etl:stream"].ID},
		{Name: "更新", Code: "etl:stream:update", Description: "更新实时数据集成", Type: "api", Path: "/api/seatunnel/stream/:id", Method: "PUT", Status: 1, ParentID: subMenuMap["etl:stream"].ID},
		{Name: "删除", Code: "etl:stream:delete", Description: "删除实时数据集成", Type: "api", Path: "/api/seatunnel/stream/:id", Method: "DELETE", Status: 1, ParentID: subMenuMap["etl:stream"].ID},
		{Name: "提交作业", Code: "etl:stream:submit", Description: "提交实时数据集成作业", Type: "api", Path: "/api/seatunnel/tasks/:id/start", Method: "POST", Status: 1, ParentID: subMenuMap["etl:stream"].ID},
		{Name: "停止作业", Code: "etl:stream:stop", Description: "停止实时数据集成作业", Type: "api", Path: "/api/seatunnel/tasks/:id/stop", Method: "POST", Status: 1, ParentID: subMenuMap["etl:stream"].ID},
		{Name: "同步状态", Code: "etl:stream:sync_status", Description: "同步实时数据集成作业状态", Type: "api", Path: "/api/seatunnel/tasks/sync-status", Method: "POST", Status: 1, ParentID: subMenuMap["etl:stream"].ID},
		// batch权限
		{Name: "查看", Code: "etl:batch:read", Description: "查看离线数据集成列表/详情", Type: "api", Path: "/api/seatunnel/batch", Method: "GET", Status: 1, ParentID: subMenuMap["etl:batch"].ID},
		{Name: "创建", Code: "etl:batch:create", Description: "创建离线数据集成", Type: "api", Path: "/api/seatunnel/batch", Method: "POST", Status: 1, ParentID: subMenuMap["etl:batch"].ID},
		{Name: "更新", Code: "etl:batch:update", Description: "更新离线数据集成", Type: "api", Path: "/api/seatunnel/batch/:id", Method: "PUT", Status: 1, ParentID: subMenuMap["etl:batch"].ID},
		{Name: "删除", Code: "etl:batch:delete", Description: "删除离线数据集成", Type: "api", Path: "/api/seatunnel/batch/:id", Method: "DELETE", Status: 1, ParentID: subMenuMap["etl:batch"].ID},
		{Name: "手动执行", Code: "etl:batch:submit", Description: "提交离线数据集成作业", Type: "api", Path: "/api/seatunnel/tasks/:id/start", Method: "POST", Status: 1, ParentID: subMenuMap["etl:batch"].ID},
		// 任务管理
		// 调度器权限
		{Name: "查看状态", Code: "task:scheduler:status", Description: "获取调度器状态", Type: "api", Path: "/api/task/scheduler/status", Method: "GET", Status: 1, ParentID: subMenuMap["task:scheduler"].ID},
		{Name: "重新加载", Code: "task:scheduler:reload", Description: "重新加载调度器", Type: "api", Path: "/api/task/scheduler/reload", Method: "POST", Status: 1, ParentID: subMenuMap["task:scheduler"].ID},
		{Name: "启动", Code: "task:scheduler:start", Description: "启动调度器", Type: "api", Path: "/api/task/scheduler/start", Method: "POST", Status: 1, ParentID: subMenuMap["task:scheduler"].ID},
		{Name: "停止", Code: "task:scheduler:stop", Description: "停止调度器", Type: "api", Path: "/api/task/scheduler/stop", Method: "POST", Status: 1, ParentID: subMenuMap["task:scheduler"].ID},

		// 任务日志权限
		{Name: "查看", Code: "task:log:read", Description: "查看任务日志", Type: "api", Path: "/api/task/log", Method: "GET", Status: 1, ParentID: subMenuMap["task:log"].ID},
		// 告警管理
		// 告警组权限
		{Name: "查看", Code: "notify:group:read", Description: "查看告警组", Type: "api", Path: "/api/alert/group", Method: "GET", Status: 1, ParentID: subMenuMap["notify:group"].ID},
		{Name: "创建", Code: "notify:group:create", Description: "创建告警组", Type: "api", Path: "/api/alert/group", Method: "POST", Status: 1, ParentID: subMenuMap["notify:group"].ID},
		{Name: "更新", Code: "notify:group:update", Description: "更新告警组", Type: "api", Path: "/api/alert/group/:id", Method: "PUT", Status: 1, ParentID: subMenuMap["notify:group"].ID},
		{Name: "删除", Code: "notify:group:delete", Description: "删除告警组", Type: "api", Path: "/api/alert/group/:id", Method: "DELETE", Status: 1, ParentID: subMenuMap["notify:group"].ID},
		{Name: "查看成员", Code: "notify:group:member:read", Description: "查看告警组成员", Type: "api", Path: "/api/alert/group/:id/members", Method: "GET", Status: 1, ParentID: subMenuMap["notify:group"].ID},
		{Name: "添加成员", Code: "notify:group:member:create", Description: "添加告警组成员", Type: "api", Path: "/api/alert/group/:id/members", Method: "POST", Status: 1, ParentID: subMenuMap["notify:group"].ID},
		{Name: "删除成员", Code: "notify:group:member:delete", Description: "删除告警组成员", Type: "api", Path: "/api/alert/group/:id/members/:member_id", Method: "DELETE", Status: 1, ParentID: subMenuMap["notify:group"].ID},
		// 告警模板权限
		{Name: "查看", Code: "notify:template:read", Description: "查看告警模板", Type: "api", Path: "/api/alert/template", Method: "GET", Status: 1, ParentID: subMenuMap["notify:template"].ID},
		{Name: "创建", Code: "notify:template:create", Description: "创建告警模板", Type: "api", Path: "/api/alert/template", Method: "POST", Status: 1, ParentID: subMenuMap["notify:template"].ID},
		{Name: "更新", Code: "notify:template:update", Description: "更新告警模板", Type: "api", Path: "/api/alert/template/:id", Method: "PUT", Status: 1, ParentID: subMenuMap["notify:template"].ID},
		{Name: "删除", Code: "notify:template:delete", Description: "删除告警模板", Type: "api", Path: "/api/alert/template/:id", Method: "DELETE", Status: 1, ParentID: subMenuMap["notify:template"].ID},
		// 告警渠道权限
		{Name: "查看", Code: "notify:channel:read", Description: "查看告警渠道", Type: "api", Path: "/api/alert/channel", Method: "GET", Status: 1, ParentID: subMenuMap["notify:channel"].ID},
		{Name: "创建", Code: "notify:channel:create", Description: "创建告警渠道", Type: "api", Path: "/api/alert/channel", Method: "POST", Status: 1, ParentID: subMenuMap["notify:channel"].ID},
		{Name: "更新", Code: "notify:channel:update", Description: "更新告警渠道", Type: "api", Path: "/api/alert/channel/:id", Method: "PUT", Status: 1, ParentID: subMenuMap["notify:channel"].ID},
		{Name: "删除", Code: "notify:channel:delete", Description: "删除告警渠道", Type: "api", Path: "/api/alert/channel/:id", Method: "DELETE", Status: 1, ParentID: subMenuMap["notify:channel"].ID},
		{Name: "测试", Code: "notify:channel:test", Description: "测试告警渠道", Type: "api", Path: "/api/alert/channel/:id/test", Method: "POST", Status: 1, ParentID: subMenuMap["notify:channel"].ID},
		// 用户管理
		{Name: "查看", Code: "rbac:user:read", Description: "查看用户", Type: "api", Path: "/api/users", Method: "GET", Status: 1, ParentID: subMenuMap["rbac:user"].ID},
		{Name: "创建", Code: "rbac:user:create", Description: "创建用户", Type: "api", Path: "/api/users", Method: "POST", Status: 1, ParentID: subMenuMap["rbac:user"].ID},
		{Name: "更新", Code: "rbac:user:update", Description: "更新用户", Type: "api", Path: "/api/users/:id", Method: "PUT", Status: 1, ParentID: subMenuMap["rbac:user"].ID},
		{Name: "删除", Code: "rbac:user:delete", Description: "删除用户", Type: "api", Path: "/api/users/:id", Method: "DELETE", Status: 1, ParentID: subMenuMap["rbac:user"].ID},
		{Name: "分配角色", Code: "rbac:user:assign_role", Description: "给用户分配或移除角色", Type: "api", Path: "/api/users/:id/roles", Method: "POST", Status: 1, ParentID: subMenuMap["rbac:user"].ID},
		// 角色管理
		{Name: "查看", Code: "rbac:role:read", Description: "查看角色", Type: "api", Path: "/api/roles", Method: "GET", Status: 1, ParentID: subMenuMap["rbac:role"].ID},
		{Name: "创建", Code: "rbac:role:create", Description: "创建角色", Type: "api", Path: "/api/roles", Method: "POST", Status: 1, ParentID: subMenuMap["rbac:role"].ID},
		{Name: "更新", Code: "rbac:role:update", Description: "更新角色", Type: "api", Path: "/api/roles/:id", Method: "PUT", Status: 1, ParentID: subMenuMap["rbac:role"].ID},
		{Name: "删除", Code: "rbac:role:delete", Description: "删除角色", Type: "api", Path: "/api/roles/:id", Method: "DELETE", Status: 1, ParentID: subMenuMap["rbac:role"].ID},
		{Name: "分配权限", Code: "rbac:role:assign_permission", Description: "给角色分配或移除权限", Type: "api", Path: "/api/roles/:id/permissions", Method: "POST", Status: 1, ParentID: subMenuMap["rbac:role"].ID},
		// 权限管理
		{Name: "查看", Code: "rbac:permission:read", Description: "查看权限", Type: "api", Path: "/api/permissions", Method: "GET", Status: 1, ParentID: subMenuMap["rbac:permission"].ID},
		{Name: "创建", Code: "rbac:permission:create", Description: "创建权限", Type: "api", Path: "/api/permissions", Method: "POST", Status: 1, ParentID: subMenuMap["rbac:permission"].ID},
		{Name: "更新", Code: "rbac:permission:update", Description: "更新权限", Type: "api", Path: "/api/permissions/:id", Method: "PUT", Status: 1, ParentID: subMenuMap["rbac:permission"].ID},
		{Name: "删除", Code: "rbac:permission:delete", Description: "删除权限", Type: "api", Path: "/api/permissions/:id", Method: "DELETE", Status: 1, ParentID: subMenuMap["rbac:permission"].ID},
		// 系统设置（update 同时保护 PUT /settings 与 POST /settings/mail/test，这里记录主路由）
		{Name: "查看", Code: "system:setting:read", Description: "查看系统设置", Type: "api", Path: "/api/system/settings", Method: "GET", Status: 1, ParentID: subMenuMap["system:setting"].ID},
		{Name: "更新", Code: "system:setting:update", Description: "更新系统设置", Type: "api", Path: "/api/system/settings", Method: "PUT", Status: 1, ParentID: subMenuMap["system:setting"].ID},
	}

	permissionMap := make(map[string]*model.Permission)
	// 一级菜单
	for k, v := range menuMap {
		permissionMap[k] = v
	}
	// 二级菜单
	for k, v := range subMenuMap {
		permissionMap[k] = v
	}
	// 三级操作权限
	for i := range permissions {
		perm, err := upsertPermission(permissions[i])
		if err != nil {
			return nil, err
		}
		permissionMap[perm.Code] = &perm
	}
	return permissionMap, nil
}

// upsertPermission 按 code 创建或更新权限，返回库中的最新记录；任何数据库错误都会返回 error
func upsertPermission(perm model.Permission) (model.Permission, error) {
	existing, found, err := findPermissionByCode(perm.Code)
	if err != nil {
		return model.Permission{}, err
	}
	if !found {
		if err := db.Create(&perm).Error; err != nil {
			return model.Permission{}, fmt.Errorf("创建权限 %s 失败: %w", perm.Code, err)
		}
		return perm, nil
	}
	// 只更新代码维护的字段，不动 status：后台停用过的权限不因重启被重新启用
	if err := db.Model(&existing).Updates(map[string]interface{}{
		"name":        perm.Name,
		"description": perm.Description,
		"type":        perm.Type,
		"path":        perm.Path,
		"order_num":   perm.OrderNum,
		"method":      perm.Method,
		"parent_id":   perm.ParentID,
	}).Error; err != nil {
		return model.Permission{}, fmt.Errorf("更新权限 %s 失败: %w", perm.Code, err)
	}
	if err := db.First(&existing, existing.ID).Error; err != nil {
		return model.Permission{}, fmt.Errorf("回读权限 %s 失败: %w", perm.Code, err)
	}
	return existing, nil
}

// createDefaultRoles 创建默认角色并对齐权限
func createDefaultRoles(permissions map[string]*model.Permission) (map[string]*model.Role, error) {
	roles := []model.Role{
		{
			Name:        "管理员",
			Description: "拥有所有管理权限",
			Status:      1,
		},
		{
			Name:        "操作员",
			Description: "拥有任务执行和查看权限",
			Status:      1,
		},
		{
			Name:        "观察者",
			Description: "只有查看权限",
			Status:      1,
		},
	}

	roleMap := make(map[string]*model.Role)

	for i := range roles {
		r := roles[i]
		var existing model.Role
		if err := db.Where("name = ?", r.Name).Limit(1).Find(&existing).Error; err != nil {
			return nil, fmt.Errorf("查询角色 %s 失败: %w", r.Name, err)
		}
		if existing.Name != "" {
			// 只同步描述，保留数据库中的角色状态（后台停用过的角色不因重启被重新启用）
			if err := db.Model(&existing).Updates(map[string]interface{}{
				"description": r.Description,
			}).Error; err != nil {
				return nil, fmt.Errorf("更新角色 %s 失败: %w", r.Name, err)
			}
			existing.Description = r.Description
			roleMap[r.Name] = &existing
			continue
		}

		if err := db.Create(&r).Error; err != nil {
			return nil, fmt.Errorf("创建角色 %s 失败: %w", r.Name, err)
		}
		log.Printf("创建角色: %s", r.Name)
		roleMap[r.Name] = &r
	}

	// 为角色分配权限
	if err := assignPermissionsToRoles(roleMap, permissions); err != nil {
		return nil, err
	}

	return roleMap, nil
}

// assignPermissionsToRoles 为角色分配权限
func assignPermissionsToRoles(roles map[string]*model.Role, permissions map[string]*model.Permission) error {
	// 超级管理员：无需分配权限，通过用户表的 IsSuperAdmin 字段标记

	// 管理员：所有权限
	if admin, exists := roles["管理员"]; exists {
		codes := make([]string, 0, len(permissions))
		for code := range permissions {
			codes = append(codes, code)
		}
		if err := assignRolePermissions("管理员", admin, permissions, codes); err != nil {
			return err
		}
	}

	// 操作员：仅新增/修改/查询权限
	if operator, exists := roles["操作员"]; exists {
		operatorPermissions := []string{
			// Seatunnel API
			"etl:stream", "etl:stream:read", "etl:stream:create", "etl:stream:update",
			"etl:batch", "etl:batch:read", "etl:batch:create", "etl:batch:update",
			"notify:group:read",
			// 任务中心 API
			"task:scheduler", "task:scheduler:status",
			"task:log", "task:log:read",
		}
		if err := assignRolePermissions("操作员", operator, permissions, operatorPermissions); err != nil {
			return err
		}
	}

	// 观察者：只有查看权限
	if observer, exists := roles["观察者"]; exists {
		observerPermissions := []string{
			// Seatunnel
			"etl:stream", "etl:stream:read",
			"etl:batch", "etl:batch:read",
			"notify:group:read",
			// 任务管理
			"task:log", "task:log:read",
			// 系统设置
			"system:setting", "system:setting:read",
		}
		if err := assignRolePermissions("观察者", observer, permissions, observerPermissions); err != nil {
			return err
		}
	}
	return nil
}

// assignRolePermissions 补齐角色权限：只新增缺失的关联，不删除任何已有授权
func assignRolePermissions(roleName string, role *model.Role, permissions map[string]*model.Permission, codes []string) error {
	if role == nil {
		return fmt.Errorf("角色 %s 为空", roleName)
	}
	if len(codes) == 0 {
		return fmt.Errorf("角色 %s 的目标权限列表为空", roleName)
	}
	rolePermissions := make([]model.RolePermission, 0, len(codes))
	seen := make(map[string]struct{}, len(codes))
	for _, code := range codes {
		// 避免配置列表重复
		if _, ok := seen[code]; ok {
			continue
		}
		seen[code] = struct{}{}

		permission, exists := permissions[code]
		if !exists || permission == nil {
			return fmt.Errorf("角色 %s 的目标权限 %s 不存在", roleName, code)
		}
		rolePermissions = append(rolePermissions, model.RolePermission{
			RoleID:       role.ID,
			PermissionID: permission.ID,
		})
	}
	// 只补齐缺失关联；role_permissions 以 (role_id, permission_id) 为主键，重复关联直接忽略
	if err := db.Clauses(clause.OnConflict{DoNothing: true}).Create(&rolePermissions).Error; err != nil {
		return fmt.Errorf("为角色 %s 补齐权限失败: %w", roleName, err)
	}
	log.Printf("已检查角色 %s 的 %d 个默认权限", roleName, len(rolePermissions))
	return nil
}

// createDefaultAdmin 创建默认管理员用户
func createDefaultAdmin() error {
	// 检查管理员用户是否已存在
	var existingUser model.User
	if err := db.Where("username = ?", "admin").Limit(1).Find(&existingUser).Error; err != nil {
		return fmt.Errorf("查询管理员用户失败: %w", err)
	}
	if existingUser.Username != "" {
		if !existingUser.IsSuperAdmin {
			// 不自动提升已有账号，交由人工确认，避免意外提权
			log.Printf("警告：admin 用户已存在但不是超级管理员，请人工确认其权限状态")
			return nil
		}
		log.Println("管理员用户已存在，跳过创建")
		return nil
	}

	// 从环境变量获取管理员密码
	initialPassword := os.Getenv("ADMIN_PASSWORD")
	if initialPassword == "" {
		return fmt.Errorf("未设置环境变量 ADMIN_PASSWORD，无法创建默认管理员")
	}

	if len(initialPassword) < 8 {
		return fmt.Errorf("ADMIN_PASSWORD 长度不能少于8位")
	}

	// 加密密码
	hashedPassword, err := crypto.HashPassword(initialPassword)
	if err != nil {
		return fmt.Errorf("密码加密失败: %w", err)
	}

	// 创建超级管理员用户
	admin := model.User{
		Username:     "admin",
		Password:     hashedPassword,
		Email:        "admin@example.com",
		Nickname:     "系统管理员",
		Status:       1,
		IsSuperAdmin: true,
	}

	if err := db.Create(&admin).Error; err != nil {
		return fmt.Errorf("创建管理员用户失败: %w", err)
	}

	log.Println("创建默认超级管理员用户成功！")
	log.Println("用户名: admin")
	return nil
}
