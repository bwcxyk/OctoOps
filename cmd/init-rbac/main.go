package main

import (
	"log"
	"octoops/internal/config"
	"octoops/internal/infra/postgres"
	"octoops/internal/model/rbac"
	"octoops/internal/pkg/crypto"
	"os"

	"gorm.io/gorm/clause"
)

// findPermissionByCode 按 code 查询权限，不存在返回 false（用 Find 避免 gorm 打印 record not found）
func findPermissionByCode(code string) (model.Permission, bool) {
	var perm model.Permission
	if err := postgres.DB.Where("code = ?", code).Limit(1).Find(&perm).Error; err != nil || perm.Code == "" {
		return model.Permission{}, false
	}
	return perm, true
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

	// 创建与菜单一致的权限树
	permissions := createMenuLikePermissions()

	// 创建默认角色并对齐权限
	roles := createDefaultRoles(permissions)

	// 回收脚本中已删除的权限：默认只提示，带 --prune 参数才实际删除
	pruneStalePermissions(permissions)

	// 创建默认管理员用户
	createDefaultAdmin(roles)

	log.Println("RBAC系统初始化完成！")
}

// createMenuLikePermissions 创建与菜单结构一致的权限树
func createMenuLikePermissions() map[string]*model.Permission {
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
		var perm model.Permission
		perm = model.Permission{
			Name:        m.Name,
			Code:        m.Code,
			Description: m.Desc,
			Type:        "menu",
			Path:        m.Path,
			OrderNum:    m.OrderNum,
			Method:      "",
			Status:      1,
			ParentID:    0,
		}
		if existing, ok := findPermissionByCode(perm.Code); ok {
			postgres.DB.Model(&existing).Updates(map[string]interface{}{
				"name":        perm.Name,
				"description": perm.Description,
				"type":        perm.Type,
				"path":        perm.Path,
				"order_num":   perm.OrderNum,
				"method":      perm.Method,
				"status":      perm.Status,
				"parent_id":   perm.ParentID,
			})
			postgres.DB.First(&perm, existing.ID)
		} else {
			postgres.DB.Create(&perm)
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
		var perm model.Permission
		perm = model.Permission{
			Name:        s.Name,
			Code:        s.Code,
			Description: s.Desc,
			Type:        "menu",
			Path:        s.Path,
			OrderNum:    s.OrderNum,
			Method:      "",
			Status:      1,
			ParentID:    menuMap[s.Parent].ID,
		}
		if existing, ok := findPermissionByCode(perm.Code); ok {
			postgres.DB.Model(&existing).Updates(map[string]interface{}{
				"name":        perm.Name,
				"description": perm.Description,
				"type":        perm.Type,
				"path":        perm.Path,
				"order_num":   perm.OrderNum,
				"method":      perm.Method,
				"status":      perm.Status,
				"parent_id":   perm.ParentID,
			})
			postgres.DB.First(&perm, existing.ID)
		} else {
			postgres.DB.Create(&perm)
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
		// 系统设置
		{Name: "查看", Code: "system:setting:read", Description: "查看系统任务设置", Type: "api", Path: "/api/system/tasks", Method: "GET", Status: 1, ParentID: subMenuMap["system:setting"].ID},
		{Name: "更新", Code: "system:setting:update", Description: "更新系统任务设置", Type: "api", Path: "/api/system/tasks/:type", Method: "PUT", Status: 1, ParentID: subMenuMap["system:setting"].ID},
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
		p := &permissions[i]
		if existing, ok := findPermissionByCode(p.Code); ok {
			if err := postgres.DB.Model(&existing).Updates(map[string]interface{}{
				"name":        p.Name,
				"description": p.Description,
				"type":        p.Type,
				"path":        p.Path,
				"method":      p.Method,
				"status":      p.Status,
				"parent_id":   p.ParentID,
			}).Error; err != nil {
				log.Printf("更新权限 %s 失败: %v", p.Code, err)
				continue
			}
			postgres.DB.First(&existing, existing.ID)
			log.Printf("更新权限: %s", p.Code)
			permissionMap[p.Code] = &existing
			continue
		}
		if err := postgres.DB.Create(p).Error; err != nil {
			log.Printf("创建权限 %s 失败: %v", p.Code, err)
			continue
		}
		log.Printf("创建权限: %s", p.Code)
		permissionMap[p.Code] = p
	}
	return permissionMap
}

// createDefaultRoles 创建默认角色
func createDefaultRoles(permissions map[string]*model.Permission) map[string]*model.Role {
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

	for _, r := range roles {
		// 检查角色是否已存在
		var existing model.Role
		if err := postgres.DB.Where("name = ?", r.Name).Limit(1).Find(&existing).Error; err == nil && existing.Name != "" {
			log.Printf("角色 %s 已存在，跳过创建", r.Name)
			roleMap[r.Name] = &existing
			continue
		}

		// 创建角色
		if err := postgres.DB.Create(&r).Error; err != nil {
			log.Printf("创建角色 %s 失败: %v", r.Name, err)
			continue
		}
		log.Printf("创建角色: %s", r.Name)
		roleMap[r.Name] = &r
	}

	// 为角色分配权限
	assignPermissionsToRoles(roleMap, permissions)

	return roleMap
}

// assignPermissionsToRoles 为角色分配权限
func assignPermissionsToRoles(roles map[string]*model.Role, permissions map[string]*model.Permission) {
	// 超级管理员：无需分配权限，通过用户表的 IsSuperAdmin 字段标记

	// 管理员：所有权限
	if admin, exists := roles["管理员"]; exists {
		codes := make([]string, 0, len(permissions))
		for code := range permissions {
			codes = append(codes, code)
		}
		assignRolePermissions("管理员", admin, permissions, codes)
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
			"system:setting", "system:setting:read", "system:setting:update",
		}
		assignRolePermissions("操作员", operator, permissions, operatorPermissions)
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
		assignRolePermissions("观察者", observer, permissions, observerPermissions)
	}
}

// assignRolePermissions 对齐角色权限：缺失的补上，已不再授权的删除（脚本可重复执行）
func assignRolePermissions(roleName string, role *model.Role, permissions map[string]*model.Permission, codes []string) {
	rolePermissions := make([]model.RolePermission, 0, len(codes))
	ids := make([]uint, 0, len(codes))
	for _, code := range codes {
		if p, exists := permissions[code]; exists {
			rolePermissions = append(rolePermissions, model.RolePermission{
				RoleID:       role.ID,
				PermissionID: p.ID,
			})
			ids = append(ids, p.ID)
		}
	}
	// 删除该角色上不在本次授权列表内的关联
	tx := postgres.DB.Where("role_id = ?", role.ID)
	if len(ids) > 0 {
		tx = tx.Where("permission_id NOT IN ?", ids)
	}
	if err := tx.Delete(&model.RolePermission{}).Error; err != nil {
		log.Printf("清理%s的多余权限失败: %v", roleName, err)
		return
	}
	if len(rolePermissions) == 0 {
		log.Printf("为%s对齐了 0 个权限", roleName)
		return
	}
	if err := postgres.DB.Clauses(clause.OnConflict{DoNothing: true}).Create(&rolePermissions).Error; err != nil {
		log.Printf("为%s分配权限失败: %v", roleName, err)
		return
	}
	log.Printf("为%s对齐了 %d 个权限", roleName, len(rolePermissions))
}

// pruneStalePermissions 回收脚本定义中已删除的权限，默认只打印，带 --prune 时删除
func pruneStalePermissions(permissions map[string]*model.Permission) {
	var all []model.Permission
	if err := postgres.DB.Find(&all).Error; err != nil {
		log.Printf("查询已有权限失败: %v", err)
		return
	}
	stale := make([]model.Permission, 0)
	for _, p := range all {
		if _, ok := permissions[p.Code]; !ok {
			stale = append(stale, p)
		}
	}
	if len(stale) == 0 {
		return
	}
	prune := len(os.Args) > 1 && os.Args[1] == "--prune"
	if !prune {
		for _, p := range stale {
			log.Printf("提示：权限 %s（%s）已不在脚本定义中，如需删除请带 --prune 参数执行", p.Code, p.Name)
		}
		return
	}
	staleIDs := make([]uint, 0, len(stale))
	for _, p := range stale {
		log.Printf("回收权限: %s（%s）", p.Code, p.Name)
		staleIDs = append(staleIDs, p.ID)
	}
	// 先解除角色关联，再删除权限（子权限在前，避免父级外键约束）
	if err := postgres.DB.Where("permission_id IN ?", staleIDs).Delete(&model.RolePermission{}).Error; err != nil {
		log.Printf("解除失效权限的角色关联失败: %v", err)
		return
	}
	if err := postgres.DB.Where("id IN ? AND parent_id <> 0", staleIDs).Delete(&model.Permission{}).Error; err != nil {
		log.Printf("删除失效子权限失败: %v", err)
		return
	}
	if err := postgres.DB.Where("id IN ?", staleIDs).Delete(&model.Permission{}).Error; err != nil {
		log.Printf("删除失效权限失败: %v", err)
		return
	}
	log.Printf("已回收 %d 个失效权限", len(stale))
}

// createDefaultAdmin 创建默认管理员用户
func createDefaultAdmin(roles map[string]*model.Role) {
	// 检查管理员用户是否已存在
	var existingUser model.User
	if err := postgres.DB.Where("username = ?", "admin").Limit(1).Find(&existingUser).Error; err == nil && existingUser.Username != "" {
		log.Println("管理员用户已存在，跳过创建")
		return
	}

	// 从环境变量获取管理员密码
	initialPassword := os.Getenv("ADMIN_PASSWORD")
	if initialPassword == "" {
		log.Printf("请设置环境变量 ADMIN_PASSWORD")
		return
	}

	if len(initialPassword) < 8 {
		log.Printf("密码长度不能少于8位")
		return
	}

	// 加密密码
	hashedPassword, err := crypto.HashPassword(initialPassword)
	if err != nil {
		log.Printf("密码加密失败: %v", err)
		return
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

	if err := postgres.DB.Create(&admin).Error; err != nil {
		log.Printf("创建管理员用户失败: %v", err)
		return
	}

	log.Println("创建默认超级管理员用户成功！")
	log.Println("用户名: admin")
}
