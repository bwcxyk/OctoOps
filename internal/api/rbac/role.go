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

// CreateRoleRequest 创建角色请求
type CreateRoleRequest struct {
	Name          string `json:"name" binding:"required"`
	Description   string `json:"description"`
	PermissionIDs []uint `json:"permission_ids"`
}

// UpdateRoleRequest 更新角色请求
type UpdateRoleRequest struct {
	Name          string `json:"name"`
	Description   string `json:"description"`
	Status        *int   `json:"status"`
	PermissionIDs []uint `json:"permission_ids"`
}

// RegisterRoleRoutes 注册角色管理路由
func RegisterRoleRoutes(r *gin.RouterGroup) {
	roles := r.Group("/roles")
	roles.Use(middleware.AuthMiddleware())
	{
		roles.GET("", middleware.RequirePermission("rbac:role:read"), getRoles)
		roles.GET("/:id", middleware.RequirePermission("rbac:role:read"), getRole)
		roles.POST("", middleware.RequirePermission("rbac:role:create"), createRole)
		roles.PUT("/:id", middleware.RequirePermission("rbac:role:update"), updateRole)
		roles.DELETE("/:id", middleware.RequirePermission("rbac:role:delete"), deleteRole)
		roles.POST("/:id/permissions", middleware.RequirePermission("rbac:role:assign_permission"), assignPermissions)
		roles.DELETE("/:id/permissions", middleware.RequirePermission("rbac:role:assign_permission"), removePermissions)
	}
}

// getRoles 获取角色列表
func getRoles(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	name := c.Query("name")
	status := c.Query("status")

	query := db.DB.Model(&model.Role{}).Preload("Permissions")

	// 添加查询条件
	if name != "" {
		query = query.Where("name LIKE ?", "%"+name+"%")
	}
	if status != "" {
		if statusInt, err := strconv.Atoi(status); err == nil {
			query = query.Where("status = ?", statusInt)
		}
	}

	var total int64
	query.Count(&total)

	var roles []model.Role
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Find(&roles).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取角色列表失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data": gin.H{
			"roles":      roles,
			"total":      total,
			"page":       page,
			"page_size":  pageSize,
			"total_page": (int(total) + pageSize - 1) / pageSize,
		},
	})
}

// getRole 获取单个角色
func getRole(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "角色ID格式错误",
		})
		return
	}

	var role model.Role
	if err := db.DB.Preload("Permissions").First(&role, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "角色不存在",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "获取角色信息失败",
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data":    role,
	})
}

// createRole 创建角色
func createRole(c *gin.Context) {
	var req CreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请求参数错误",
			"error":   err.Error(),
		})
		return
	}

	// 检查角色名是否已存在
	var existingRole model.Role
	if err := db.DB.Where("name = ?", req.Name).First(&existingRole).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "角色名已存在",
		})
		return
	}

	// 开始事务
	tx := db.DB.Begin()

	// 创建角色
	role := model.Role{
		Name:        req.Name,
		Description: req.Description,
		Status:      1,
	}

	if err := tx.Create(&role).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "创建角色失败",
		})
		return
	}

	// 分配权限
	if len(req.PermissionIDs) > 0 {
		var rolePermissions []model.RolePermission
		for _, permissionID := range req.PermissionIDs {
			rolePermissions = append(rolePermissions, model.RolePermission{
				RoleID:       role.ID,
				PermissionID: permissionID,
			})
		}
		if err := tx.Create(&rolePermissions).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "分配权限失败",
			})
			return
		}
	}

	tx.Commit()

	// 重新加载角色信息
	db.DB.Preload("Permissions").First(&role, role.ID)

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "创建成功",
		"data":    role,
	})
}

// updateRole 更新角色
func updateRole(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "角色ID格式错误",
		})
		return
	}

	var req UpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请求参数错误",
			"error":   err.Error(),
		})
		return
	}

	// 检查角色是否存在
	var role model.Role
	if err := db.DB.First(&role, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "角色不存在",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "获取角色信息失败",
			})
		}
		return
	}

	// 开始事务
	tx := db.DB.Begin()

	// 更新角色信息
	updates := make(map[string]interface{})
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}

	if len(updates) > 0 {
		if err := tx.Model(&role).Updates(updates).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "更新角色信息失败",
			})
			return
		}
	}

	// 更新权限
	if req.PermissionIDs != nil {
		// 删除现有权限
		if err := tx.Where("role_id = ?", role.ID).Delete(&model.RolePermission{}).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "删除角色权限失败",
			})
			return
		}

		// 分配新权限
		if len(req.PermissionIDs) > 0 {
			var rolePermissions []model.RolePermission
			for _, permissionID := range req.PermissionIDs {
				rolePermissions = append(rolePermissions, model.RolePermission{
					RoleID:       role.ID,
					PermissionID: permissionID,
				})
			}
			if err := tx.Create(&rolePermissions).Error; err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{
					"code":    500,
					"message": "分配权限失败",
				})
				return
			}
		}
	}

	tx.Commit()

	// 重新加载角色信息
	db.DB.Preload("Permissions").First(&role, role.ID)

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "更新成功",
		"data":    role,
	})
}

// deleteRole 删除角色
func deleteRole(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "角色ID格式错误",
		})
		return
	}

	// 开始事务
	tx := db.DB.Begin()

	// 删除角色权限关联
	if err := tx.Where("role_id = ?", id).Delete(&model.RolePermission{}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "删除角色权限失败",
		})
		return
	}

	// 删除用户角色关联
	if err := tx.Where("role_id = ?", id).Delete(&model.UserRole{}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "删除用户角色失败",
		})
		return
	}

	// 删除角色
	if err := tx.Delete(&model.Role{}, id).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "删除角色失败",
		})
		return
	}

	tx.Commit()

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "删除成功",
	})
}

// assignPermissions 分配权限
func assignPermissions(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "角色ID格式错误",
		})
		return
	}

	var req struct {
		PermissionIDs []uint `json:"permission_ids" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请求参数错误",
		})
		return
	}

	// 检查角色是否存在
	var role model.Role
	if err := db.DB.First(&role, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "角色不存在",
		})
		return
	}

	// 分配权限
	var rolePermissions []model.RolePermission
	for _, permissionID := range req.PermissionIDs {
		rolePermissions = append(rolePermissions, model.RolePermission{
			RoleID:       role.ID,
			PermissionID: permissionID,
		})
	}

	if err := db.DB.Create(&rolePermissions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "分配权限失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "分配权限成功",
	})
}

// removePermissions 移除权限
func removePermissions(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "角色ID格式错误",
		})
		return
	}

	var req struct {
		PermissionIDs []uint `json:"permission_ids" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请求参数错误",
		})
		return
	}

	// 移除权限
	if err := db.DB.Where("role_id = ? AND permission_id IN ?", id, req.PermissionIDs).Delete(&model.RolePermission{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "移除权限失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "移除权限成功",
	})
}
