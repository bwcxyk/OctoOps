package api

import (
	"net/http"
	"octoops/internal/db"
	"octoops/internal/middleware"
	"octoops/internal/model/rbac"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// CreatePermissionRequest 创建权限请求
type CreatePermissionRequest struct {
	Name        string `json:"name" binding:"required"`
	Code        string `json:"code" binding:"required"`
	Description string `json:"description"`
	Type        string `json:"type"`   // menu:菜单权限 api:接口权限
	Path        string `json:"path"`   // 菜单路径或API路径
	Method      string `json:"method"` // HTTP方法，如: GET, POST, PUT, DELETE
}

// UpdatePermissionRequest 更新权限请求
type UpdatePermissionRequest struct {
	Name        string `json:"name"`
	Code        string `json:"code"`
	Description string `json:"description"`
	Type        string `json:"type"`
	Path        string `json:"path"`
	Method      string `json:"method"`
	Status      *int   `json:"status"`
}

// RegisterPermissionRoutes 注册权限管理路由
func RegisterPermissionRoutes(r *gin.RouterGroup) {
	permissions := r.Group("/permissions")
	permissions.Use(middleware.AuthMiddleware())
	{
		permissions.GET("", middleware.RequirePermission("rbac:permission:read"), getPermissions)
		permissions.GET("/:id", middleware.RequirePermission("rbac:permission:read"), getPermission)
		permissions.POST("", middleware.RequirePermission("rbac:permission:create"), createPermission)
		permissions.PUT("/:id", middleware.RequirePermission("rbac:permission:update"), updatePermission)
		permissions.DELETE("/:id", middleware.RequirePermission("rbac:permission:delete"), deletePermission)
		permissions.GET("/tree", middleware.RequirePermission("rbac:permission:read"), getPermissionTree)
	}

	// 注册菜单树接口
	r.GET("/menus", middleware.AuthMiddleware(), getUserMenus)
}

// getPermissions 获取权限列表
func getPermissions(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	name := c.Query("name")
	code := c.Query("code")
	type_ := c.Query("type")
	status := c.Query("status")

	query := db.DB.Model(&model.Permission{})

	// 添加查询条件
	if name != "" {
		query = query.Where("name LIKE ?", "%"+name+"%")
	}
	if code != "" {
		query = query.Where("code LIKE ?", "%"+code+"%")
	}
	if type_ != "" {
		query = query.Where("type = ?", type_)
	}
	if status != "" {
		if statusInt, err := strconv.Atoi(status); err == nil {
			query = query.Where("status = ?", statusInt)
		}
	}

	var total int64
	query.Count(&total)

	var permissions []model.Permission
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Find(&permissions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取权限列表失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data": gin.H{
			"permissions": permissions,
			"total":       total,
			"page":        page,
			"page_size":   pageSize,
			"total_page":  (int(total) + pageSize - 1) / pageSize,
		},
	})
}

// getPermission 获取单个权限
func getPermission(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "权限ID格式错误",
		})
		return
	}

	var permission model.Permission
	if err := db.DB.First(&permission, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "权限不存在",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "获取权限信息失败",
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data":    permission,
	})
}

// createPermission 创建权限
func createPermission(c *gin.Context) {
	var req CreatePermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请求参数错误",
			"error":   err.Error(),
		})
		return
	}

	// 检查权限代码是否已存在
	var existingPermission model.Permission
	if err := db.DB.Where("code = ?", req.Code).First(&existingPermission).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "权限代码已存在",
		})
		return
	}

	// 创建权限
	permission := model.Permission{
		Name:        req.Name,
		Code:        req.Code,
		Description: req.Description,
		Type:        req.Type,
		Path:        req.Path,
		Method:      req.Method,
		Status:      1,
	}

	if err := db.DB.Create(&permission).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "创建权限失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "创建成功",
		"data":    permission,
	})
}

// updatePermission 更新权限
func updatePermission(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "权限ID格式错误",
		})
		return
	}

	var req UpdatePermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请求参数错误",
			"error":   err.Error(),
		})
		return
	}

	// 检查权限是否存在
	var permission model.Permission
	if err := db.DB.First(&permission, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "权限不存在",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "获取权限信息失败",
			})
		}
		return
	}

	// 如果更新了权限代码，检查是否与其他权限冲突
	if req.Code != "" && req.Code != permission.Code {
		var existingPermission model.Permission
		if err := db.DB.Where("code = ? AND id != ?", req.Code, id).First(&existingPermission).Error; err == nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    400,
				"message": "权限代码已存在",
			})
			return
		}
	}

	// 更新权限信息
	updates := make(map[string]interface{})
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Code != "" {
		updates["code"] = req.Code
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}
	if req.Type != "" {
		updates["type"] = req.Type
	}
	if req.Path != "" {
		updates["path"] = req.Path
	}
	if req.Method != "" {
		updates["method"] = req.Method
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}

	if len(updates) > 0 {
		if err := db.DB.Model(&permission).Updates(updates).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "更新权限信息失败",
			})
			return
		}
	}

	// 重新加载权限信息
	db.DB.First(&permission, permission.ID)

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "更新成功",
		"data":    permission,
	})
}

// deletePermission 删除权限
func deletePermission(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "权限ID格式错误",
		})
		return
	}

	// 开始事务
	tx := db.DB.Begin()

	// 删除角色权限关联
	if err := tx.Where("permission_id = ?", id).Delete(&model.RolePermission{}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "删除角色权限失败",
		})
		return
	}

	// 删除权限
	if err := tx.Delete(&model.Permission{}, id).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "删除权限失败",
		})
		return
	}

	tx.Commit()

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "删除成功",
	})
}

// PermissionTreeNode 权限树节点
type PermissionTreeNode struct {
	ID          uint                 `json:"id"`
	Name        string               `json:"name"`
	Code        string               `json:"code"`
	Description string               `json:"description"`
	Type        string               `json:"type"`
	Path        string               `json:"path"`
	Method      string               `json:"method"`
	Status      int                  `json:"status"`
	Children    []PermissionTreeNode `json:"children"`
}

// getPermissionTree 获取权限树（基于 parent_id 递归）
func getPermissionTree(c *gin.Context) {
	var permissions []model.Permission
	if err := db.DB.Where("status = 1").Find(&permissions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取权限列表失败",
		})
		return
	}

	tree := buildPermissionTreeByParentID(permissions, 0)
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data":    tree,
	})
}

// buildPermissionTreeByParentID 基于 parent_id 递归组装权限树
func buildPermissionTreeByParentID(permissions []model.Permission, parentID uint) []PermissionTreeNode {
	var tree []PermissionTreeNode
	for _, p := range permissions {
		if p.ParentID == parentID {
			node := PermissionTreeNode{
				ID:          p.ID,
				Name:        p.Name,
				Code:        p.Code,
				Description: p.Description,
				Type:        p.Type,
				Path:        p.Path,
				Method:      p.Method,
				Status:      p.Status,
				Children:    buildPermissionTreeByParentID(permissions, p.ID),
			}
			tree = append(tree, node)
		}
	}
	return tree
}

// MenuNode 菜单节点结构
type MenuNode struct {
	Name     string     `json:"name"`
	Code     string     `json:"code"`
	Path     string     `json:"path,omitempty"`
	Icon     string     `json:"icon,omitempty"`
	Children []MenuNode `json:"children,omitempty"`
	Permission string   `json:"permission,omitempty"`
}

// getUserMenus 返回当前用户有权限的菜单树
func getUserMenus(c *gin.Context) {
	user, _ := c.Get("user")
	userModel, _ := user.(*model.User)

	// 查询用户所有权限
	var roles []model.Role
	db.DB.Model(userModel).Preload("Permissions").Association("Roles").Find(&roles)
	permMap := make(map[string]bool)
	for _, role := range roles {
		for _, p := range role.Permissions {
			permMap[p.Code] = true
		}
	}

	// 查询所有菜单权限
	var menus []model.Permission
	db.DB.Where("type = ? AND status = 1", "menu").Order("order_num ASC, id ASC").Find(&menus)

	menuTree := buildMenuTree(menus, 0, permMap)
	c.JSON(200, gin.H{"code": 200, "data": menuTree})
}

// buildMenuTree 递归组装菜单树，只保留有权限的菜单
func buildMenuTree(menus []model.Permission, parentID uint, userPerms map[string]bool) []MenuNode {
	var tree []MenuNode
	for _, m := range menus {
		if m.ParentID == parentID {
			children := buildMenuTree(menus, m.ID, userPerms)
			// 叶子节点需有权限，overview 特殊放行
			if len(children) == 0 && m.Code != "overview" && !userPerms[m.Code] {
				continue
			}
			node := MenuNode{
				Name: m.Name,
				Code: m.Code,
				Path: m.Path,
				Children: children,
				Permission: m.Code, // 可选，前端可用
			}
			tree = append(tree, node)
		}
	}
	return tree
}
