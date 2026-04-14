package main

import (
	"log"
	"octoops/internal/config"
	"octoops/internal/db"
	"octoops/internal/model/rbac"
	"octoops/internal/utils"
)

func main() {
	// 初始化配置和数据库
	config.InitConfig()
	db.Init()

	log.Println("开始初始化RBAC系统...")

	// 创建与菜单一致的权限树
	permissions := createMenuLikePermissions()

	// 创建默认角色
	roles := createDefaultRoles(permissions)

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
		{"仪表盘", "dashboard", "仪表盘页面", "/dashboard", 0},
		{"阿里云", "aliyun", "阿里云相关", "/aliyun", 1},
		{"Seatunnel", "seatunnel", "Seatunnel相关", "/seatunnel", 2},
		{"任务中心", "task", "任务管理相关", "/task", 3},
		{"告警管理", "notify", "消息通知相关", "/alert", 4},
		{"系统管理", "rbac", "系统管理相关", "/rbac", 5},
	}
	menuMap := make(map[string]*model.Permission)
	for _, m := range menuDefs {
		perm := model.Permission{
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
		db.DB.Where("code = ?", perm.Code).FirstOrCreate(&perm)
		menuMap[m.Code] = &perm
	}

	// 2. 二级菜单
	subMenuDefs := []struct {
		Name, Code, Desc, Parent, Path string
		OrderNum                       int
	}{
		// 仪表盘
		{"基础仪表盘", "dashboard:base", "基础仪表盘", "dashboard", "/dashboard/base", 1},
		{"详情仪表盘", "dashboard:detail", "详情仪表盘", "dashboard", "/dashboard/detail", 2},
		// 阿里云
		{"ECS安全组", "aliyun:ecs_sg", "ECS安全组", "aliyun", "/aliyun/ecs-security-group", 1},
		// ETL调度
		{"实时数据集成", "etl:stream", "实时数据集成", "seatunnel", "/seatunnel/stream", 1},
		{"离线数据集成", "etl:batch", "离线数据集成", "seatunnel", "/seatunnel/batch", 2},
		// 任务管理
		{"调度器", "task:scheduler", "调度器", "task", "/task/scheduler", 1},
		{"自定义任务", "task:schedule", "自定义任务", "task", "/task/custom", 2},
		{"任务日志", "task:log", "任务日志", "task", "/task/log", 3},
		// 消息通知
		{"告警组管理", "notify:group", "告警组管理", "notify", "/alert/group", 1},
		{"告警模板", "notify:template", "告警模板", "notify", "/alert/template", 2},
		{"告警渠道", "notify:channel", "告警渠道", "notify", "/alert/channel", 3},
		// 权限管理
		{"用户管理", "rbac:user", "用户管理", "rbac", "/rbac/user", 1},
		{"角色管理", "rbac:role", "角色管理", "rbac", "/rbac/role", 2},
		{"权限管理", "rbac:permission", "权限管理", "rbac", "/rbac/permission", 3},
	}
	subMenuMap := make(map[string]*model.Permission)
	for _, s := range subMenuDefs {
		perm := model.Permission{
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
		db.DB.Where("code = ?", perm.Code).FirstOrCreate(&perm)
		subMenuMap[s.Code] = &perm
	}

	// 3. 三级操作权限（以系统管理为例，可按需扩展到其他菜单）
	permissions := []model.Permission{
		// 阿里云
		{Name: "ECS安全组查看", Code: "aliyun:ecs_sg:read", Description: "查看ECS安全组", Type: "api", Path: "/api/aliyun/ecs-sg", Method: "GET", Status: 1, ParentID: subMenuMap["aliyun:ecs_sg"].ID},
		{Name: "ECS安全组创建", Code: "aliyun:ecs_sg:create", Description: "创建ECS安全组", Type: "api", Path: "/api/aliyun/ecs-sg", Method: "POST", Status: 1, ParentID: subMenuMap["aliyun:ecs_sg"].ID},
		{Name: "ECS安全组更新", Code: "aliyun:ecs_sg:update", Description: "更新ECS安全组", Type: "api", Path: "/api/aliyun/ecs-sg/:id", Method: "PUT", Status: 1, ParentID: subMenuMap["aliyun:ecs_sg"].ID},
		{Name: "ECS安全组删除", Code: "aliyun:ecs_sg:delete", Description: "删除ECS安全组", Type: "api", Path: "/api/aliyun/ecs-sg/:id", Method: "DELETE", Status: 1, ParentID: subMenuMap["aliyun:ecs_sg"].ID},
		// Seatunnel
		{Name: "实时数据集成查看", Code: "etl:stream:read", Description: "查看实时数据集成", Type: "api", Path: "/api/seatunnel/stream", Method: "GET", Status: 1, ParentID: subMenuMap["etl:stream"].ID},
		{Name: "实时数据集成创建", Code: "etl:stream:create", Description: "创建实时数据集成", Type: "api", Path: "/api/seatunnel/stream", Method: "POST", Status: 1, ParentID: subMenuMap["etl:stream"].ID},
		{Name: "实时数据集成更新", Code: "etl:stream:update", Description: "更新实时数据集成", Type: "api", Path: "/api/seatunnel/stream/:id", Method: "PUT", Status: 1, ParentID: subMenuMap["etl:stream"].ID},
		{Name: "实时数据集成删除", Code: "etl:stream:delete", Description: "删除实时数据集成", Type: "api", Path: "/api/seatunnel/stream/:id", Method: "DELETE", Status: 1, ParentID: subMenuMap["etl:stream"].ID},
		{Name: "实时数据集成提交作业", Code: "etl:stream:submit", Description: "提交实时数据集成作业", Type: "api", Path: "/api/submit-job", Method: "POST", Status: 1, ParentID: subMenuMap["etl:stream"].ID},
		{Name: "实时数据集成停止作业", Code: "etl:stream:stop", Description: "停止实时数据集成作业", Type: "api", Path: "/api/stop-job", Method: "POST", Status: 1, ParentID: subMenuMap["etl:stream"].ID},
		{Name: "实时数据集成同步作业状态", Code: "etl:stream:sync_status", Description: "同步实时数据集成作业状态", Type: "api", Path: "/api/sync-job-status", Method: "POST", Status: 1, ParentID: subMenuMap["etl:stream"].ID},
		{Name: "离线数据集成查看", Code: "etl:batch:read", Description: "查看离线数据集成", Type: "api", Path: "/api/seatunnel/batch", Method: "GET", Status: 1, ParentID: subMenuMap["etl:batch"].ID},
		{Name: "离线数据集成创建", Code: "etl:batch:create", Description: "创建离线数据集成", Type: "api", Path: "/api/seatunnel/batch", Method: "POST", Status: 1, ParentID: subMenuMap["etl:batch"].ID},
		{Name: "离线数据集成更新", Code: "etl:batch:update", Description: "更新离线数据集成", Type: "api", Path: "/api/seatunnel/batch/:id", Method: "PUT", Status: 1, ParentID: subMenuMap["etl:batch"].ID},
		{Name: "离线数据集成删除", Code: "etl:batch:delete", Description: "删除离线数据集成", Type: "api", Path: "/api/seatunnel/batch/:id", Method: "DELETE", Status: 1, ParentID: subMenuMap["etl:batch"].ID},
		{Name: "离线数据集成手动执行", Code: "etl:batch:submit", Description: "手动执行离线数据集成作业（提交作业）", Type: "api", Path: "/api/submit-job", Method: "POST", Status: 1, ParentID: subMenuMap["etl:batch"].ID},
		// 任务管理
		{Name: "获取调度器状态", Code: "task:scheduler:status", Description: "获取调度器状态", Type: "api", Path: "/api/task/scheduler/status", Method: "GET", Status: 1, ParentID: subMenuMap["task:scheduler"].ID},
		{Name: "重新加载调度器", Code: "task:scheduler:reload", Description: "重新加载调度器", Type: "api", Path: "/api/task/scheduler/reload", Method: "POST", Status: 1, ParentID: subMenuMap["task:scheduler"].ID},
		{Name: "启动调度器", Code: "task:scheduler:start", Description: "启动调度器", Type: "api", Path: "/api/task/scheduler/start", Method: "POST", Status: 1, ParentID: subMenuMap["task:scheduler"].ID},
		{Name: "停止调度器", Code: "task:scheduler:stop", Description: "停止调度器", Type: "api", Path: "/api/task/scheduler/stop", Method: "POST", Status: 1, ParentID: subMenuMap["task:scheduler"].ID},
		{Name: "自定义任务查看", Code: "task:schedule:read", Description: "查看自定义任务", Type: "api", Path: "/api/task/custom", Method: "GET", Status: 1, ParentID: subMenuMap["task:schedule"].ID},
		{Name: "任务日志查看", Code: "task:log:read", Description: "查看任务日志", Type: "api", Path: "/api/task/log", Method: "GET", Status: 1, ParentID: subMenuMap["task:log"].ID},
		// 告警管理
		{Name: "告警组管理查看", Code: "notify:group:read", Description: "查看告警组", Type: "api", Path: "/api/alert/group", Method: "GET", Status: 1, ParentID: subMenuMap["notify:group"].ID},
		{Name: "告警组管理创建", Code: "notify:group:create", Description: "创建告警组", Type: "api", Path: "/api/alert/group", Method: "POST", Status: 1, ParentID: subMenuMap["notify:group"].ID},
		{Name: "告警组管理更新", Code: "notify:group:update", Description: "更新告警组", Type: "api", Path: "/api/alert/group/:id", Method: "PUT", Status: 1, ParentID: subMenuMap["notify:group"].ID},
		{Name: "告警组管理删除", Code: "notify:group:delete", Description: "删除告警组", Type: "api", Path: "/api/alert/group/:id", Method: "DELETE", Status: 1, ParentID: subMenuMap["notify:group"].ID},
		{Name: "告警模板查看", Code: "notify:template:read", Description: "查看告警模板", Type: "api", Path: "/api/alert/template", Method: "GET", Status: 1, ParentID: subMenuMap["notify:template"].ID},
		{Name: "告警模板创建", Code: "notify:template:create", Description: "创建告警模板", Type: "api", Path: "/api/alert/template", Method: "POST", Status: 1, ParentID: subMenuMap["notify:template"].ID},
		{Name: "告警模板更新", Code: "notify:template:update", Description: "更新告警模板", Type: "api", Path: "/api/alert/template/:id", Method: "PUT", Status: 1, ParentID: subMenuMap["notify:template"].ID},
		{Name: "告警模板删除", Code: "notify:template:delete", Description: "删除告警模板", Type: "api", Path: "/api/alert/template/:id", Method: "DELETE", Status: 1, ParentID: subMenuMap["notify:template"].ID},
		{Name: "告警渠道查看", Code: "notify:channel:read", Description: "查看告警渠道", Type: "api", Path: "/api/alert/channel", Method: "GET", Status: 1, ParentID: subMenuMap["notify:channel"].ID},
		{Name: "告警渠道创建", Code: "notify:channel:create", Description: "创建告警渠道", Type: "api", Path: "/api/alert/channel", Method: "POST", Status: 1, ParentID: subMenuMap["notify:channel"].ID},
		{Name: "告警渠道更新", Code: "notify:channel:update", Description: "更新告警渠道", Type: "api", Path: "/api/alert/channel/:id", Method: "PUT", Status: 1, ParentID: subMenuMap["notify:channel"].ID},
		{Name: "告警渠道删除", Code: "notify:channel:delete", Description: "删除告警渠道", Type: "api", Path: "/api/alert/channel/:id", Method: "DELETE", Status: 1, ParentID: subMenuMap["notify:channel"].ID},
		// 用户管理
		{Name: "用户查看", Code: "rbac:user:read", Description: "查看用户", Type: "api", Path: "/api/users", Method: "GET", Status: 1, ParentID: subMenuMap["rbac:user"].ID},
		{Name: "用户创建", Code: "rbac:user:create", Description: "创建用户", Type: "api", Path: "/api/users", Method: "POST", Status: 1, ParentID: subMenuMap["rbac:user"].ID},
		{Name: "用户更新", Code: "rbac:user:update", Description: "更新用户", Type: "api", Path: "/api/users/:id", Method: "PUT", Status: 1, ParentID: subMenuMap["rbac:user"].ID},
		{Name: "用户删除", Code: "rbac:user:delete", Description: "删除用户", Type: "api", Path: "/api/users/:id", Method: "DELETE", Status: 1, ParentID: subMenuMap["rbac:user"].ID},
		// 角色管理
		{Name: "角色查看", Code: "rbac:role:read", Description: "查看角色", Type: "api", Path: "/api/roles", Method: "GET", Status: 1, ParentID: subMenuMap["rbac:role"].ID},
		{Name: "角色创建", Code: "rbac:role:create", Description: "创建角色", Type: "api", Path: "/api/roles", Method: "POST", Status: 1, ParentID: subMenuMap["rbac:role"].ID},
		{Name: "角色更新", Code: "rbac:role:update", Description: "更新角色", Type: "api", Path: "/api/roles/:id", Method: "PUT", Status: 1, ParentID: subMenuMap["rbac:role"].ID},
		{Name: "角色删除", Code: "rbac:role:delete", Description: "删除角色", Type: "api", Path: "/api/roles/:id", Method: "DELETE", Status: 1, ParentID: subMenuMap["rbac:role"].ID},
		// 权限管理
		{Name: "权限查看", Code: "rbac:permission:read", Description: "查看权限", Type: "api", Path: "/api/permissions", Method: "GET", Status: 1, ParentID: subMenuMap["rbac:permission"].ID},
		{Name: "权限创建", Code: "rbac:permission:create", Description: "创建权限", Type: "api", Path: "/api/permissions", Method: "POST", Status: 1, ParentID: subMenuMap["rbac:permission"].ID},
		{Name: "权限更新", Code: "rbac:permission:update", Description: "更新权限", Type: "api", Path: "/api/permissions/:id", Method: "PUT", Status: 1, ParentID: subMenuMap["rbac:permission"].ID},
		{Name: "权限删除", Code: "rbac:permission:delete", Description: "删除权限", Type: "api", Path: "/api/permissions/:id", Method: "DELETE", Status: 1, ParentID: subMenuMap["rbac:permission"].ID},
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
		var existing model.Permission
		if err := db.DB.Where("code = ?", p.Code).First(&existing).Error; err == nil {
			log.Printf("权限 %s 已存在，跳过创建", p.Code)
			permissionMap[p.Code] = &existing
			continue
		}
		if err := db.DB.Create(p).Error; err != nil {
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
		if err := db.DB.Where("name = ?", r.Name).First(&existing).Error; err == nil {
			log.Printf("角色 %s 已存在，跳过创建", r.Name)
			roleMap[r.Name] = &existing
			continue
		}

		// 创建角色
		if err := db.DB.Create(&r).Error; err != nil {
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
		var rolePermissions []model.RolePermission
		for _, p := range permissions {
			rolePermissions = append(rolePermissions, model.RolePermission{
				RoleID:       admin.ID,
				PermissionID: p.ID,
			})
		}
		if len(rolePermissions) > 0 {
			db.DB.Create(&rolePermissions)
			log.Printf("为管理员分配了 %d 个权限", len(rolePermissions))
		}
	}

	// 操作员：任务和告警相关权限（仅分配三级API权限）
	if operator, exists := roles["操作员"]; exists {
		operatorPermissions := []string{
			// 仪表盘
			"dashboard:base", "dashboard:detail",
			// Seatunnel
			"etl:stream:read", "etl:stream:create", "etl:stream:update", "etl:stream:submit", "etl:stream:stop", "etl:stream:sync_status",
			"etl:batch:read", "etl:batch:create", "etl:batch:update", "etl:batch:submit",
			// 任务管理
			"task:scheduler:status", "task:schedule:read", "task:log:read",
		}
		var rolePermissions []model.RolePermission
		for _, code := range operatorPermissions {
			if p, exists := permissions[code]; exists {
				rolePermissions = append(rolePermissions, model.RolePermission{
					RoleID:       operator.ID,
					PermissionID: p.ID,
				})
			}
		}
		if len(rolePermissions) > 0 {
			db.DB.Create(&rolePermissions)
			log.Printf("为操作员分配了 %d 个权限", len(rolePermissions))
		}
	}

	// 观察者：只有查看权限（仅分配三级API权限）
	if observer, exists := roles["观察者"]; exists {
		observerPermissions := []string{
			// 阿里云
			"aliyun:ecs_sg:read",
			// Seatunnel
			"etl:stream:read", "etl:batch:read",
			// 任务管理
			"task:scheduler:status", "task:schedule:read", "task:log:read",
			// 告警管理
			"notify:group:read", "notify:template:read", "notify:channel:read",
			// 系统管理
			"rbac:user:read", "rbac:role:read", "rbac:permission:read",
		}
		var rolePermissions []model.RolePermission
		for _, code := range observerPermissions {
			if p, exists := permissions[code]; exists {
				rolePermissions = append(rolePermissions, model.RolePermission{
					RoleID:       observer.ID,
					PermissionID: p.ID,
				})
			}
		}
		if len(rolePermissions) > 0 {
			db.DB.Create(&rolePermissions)
			log.Printf("为观察者分配了 %d 个权限", len(rolePermissions))
		}
	}
}

// createDefaultAdmin 创建默认管理员用户
func createDefaultAdmin(roles map[string]*model.Role) {
	// 检查管理员用户是否已存在
	var existingUser model.User
	if err := db.DB.Where("username = ?", "admin").First(&existingUser).Error; err == nil {
		log.Println("管理员用户已存在，跳过创建")
		return
	}

	// 加密密码
	hashedPassword, err := utils.HashPassword("admin123")
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

	if err := db.DB.Create(&admin).Error; err != nil {
		log.Printf("创建管理员用户失败: %v", err)
		return
	}

	log.Println("创建默认超级管理员用户成功！")
	log.Println("用户名: admin")
	log.Println("密码: admin123")
	log.Println("请及时修改默认密码！")
}
