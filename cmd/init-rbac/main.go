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
		OrderNum int
	}{
		{"系统总览", "overview", "系统总览页面", "/overview", 1},
		{"阿里云", "aliyun", "阿里云相关", "", 2},
		{"ETL调度", "etl", "ETL调度相关", "", 3},
		{"任务管理", "task", "任务管理相关", "", 4},
		{"作业日志", "tasklog", "作业日志相关", "/tasklog", 5},
		{"消息通知", "notify", "消息通知相关", "", 6},
		{"系统管理", "rbac", "系统管理相关", "", 7},
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
		OrderNum int
	}{
		// 阿里云
		{"ECS安全组", "aliyun:ecs_sg", "ECS安全组", "aliyun", "/ecs-security-group", 1},
		// ETL调度
		{"离线数据集成", "etl:batch", "离线数据集成", "etl", "/batchtask", 1},
		{"实时数据集成", "etl:stream", "实时数据集成", "etl", "/streamtask", 2},
		// 任务管理
		{"定时任务", "task:schedule", "定时任务", "task", "/task/timer", 1},
		{"调度器", "task:scheduler", "调度器", "task", "/scheduler", 2},
		// 消息通知
		{"告警模板", "notify:template", "告警模板", "notify", "/alert-template", 1},
		{"告警渠道", "notify:channel", "告警渠道", "notify", "/alert-channel", 2},
		{"告警组管理", "notify:group", "告警组管理", "notify", "/alert-group", 3},
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
			Name:        "超级管理员",
			Description: "拥有系统所有权限",
			Status:      1,
		},
		{
			Name:        "管理员",
			Description: "拥有大部分管理权限",
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
	// 超级管理员：所有权限
	if superAdmin, exists := roles["超级管理员"]; exists {
		var rolePermissions []model.RolePermission
		for _, p := range permissions {
			rolePermissions = append(rolePermissions, model.RolePermission{
				RoleID:       superAdmin.ID,
				PermissionID: p.ID,
			})
		}
		if len(rolePermissions) > 0 {
			db.DB.Create(&rolePermissions)
			log.Printf("为超级管理员分配了 %d 个权限", len(rolePermissions))
		}
	}

	// 管理员：除系统配置外的所有权限
	if admin, exists := roles["管理员"]; exists {
		var rolePermissions []model.RolePermission
		for code, p := range permissions {
			if code != "system:read" && code != "system:update" {
				rolePermissions = append(rolePermissions, model.RolePermission{
					RoleID:       admin.ID,
					PermissionID: p.ID,
				})
			}
		}
		if len(rolePermissions) > 0 {
			db.DB.Create(&rolePermissions)
			log.Printf("为管理员分配了 %d 个权限", len(rolePermissions))
		}
	}

	// 操作员：任务和告警相关权限
	if operator, exists := roles["操作员"]; exists {
		operatorPermissions := []string{
			"task:read", "task:create", "task:update", "task:execute",
			"alert:read", "alert:create", "alert:update",
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

	// 观察者：只有查看权限
	if observer, exists := roles["观察者"]; exists {
		observerPermissions := []string{
			"task:read", "alert:read", "user:read", "role:read", "permission:read",
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

	// 创建管理员用户
	admin := model.User{
		Username: "admin",
		Password: hashedPassword,
		Email:    "admin@example.com",
		Nickname: "系统管理员",
		Status:   1,
	}

	if err := db.DB.Create(&admin).Error; err != nil {
		log.Printf("创建管理员用户失败: %v", err)
		return
	}

	// 为管理员分配超级管理员角色
	if superAdmin, exists := roles["超级管理员"]; exists {
		userRole := model.UserRole{
			UserID: admin.ID,
			RoleID: superAdmin.ID,
		}
		if err := db.DB.Create(&userRole).Error; err != nil {
			log.Printf("为管理员分配角色失败: %v", err)
		} else {
			log.Println("为管理员分配了超级管理员角色")
		}
	}

	log.Println("创建默认管理员用户成功！")
	log.Println("用户名: admin")
	log.Println("密码: admin123")
	log.Println("请及时修改默认密码！")
}
