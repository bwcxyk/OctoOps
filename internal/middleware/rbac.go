package middleware

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"octoops/internal/db"
	"octoops/internal/model/rbac"
)

// RequirePermission 要求特定权限的中间件
func RequirePermission(permissionCode string) gin.HandlerFunc {
	return func(c *gin.Context) {
		user := GetCurrentUser(c)
		if user == nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "用户未认证",
			})
			c.Abort()
			return
		}

		// 检查用户是否有指定权限
		if !HasPermission(user, permissionCode) {
			c.JSON(http.StatusForbidden, gin.H{
				"code":    403,
				"message": "权限不足",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireAnyPermission 要求任意一个权限的中间件
func RequireAnyPermission(permissionCodes ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		user := GetCurrentUser(c)
		if user == nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "用户未认证",
			})
			c.Abort()
			return
		}

		// 检查用户是否有任意一个指定权限
		for _, code := range permissionCodes {
			if HasPermission(user, code) {
				c.Next()
				return
			}
		}

		c.JSON(http.StatusForbidden, gin.H{
			"code":    403,
			"message": "权限不足",
		})
		c.Abort()
	}
}

// RequireAllPermissions 要求所有权限的中间件
func RequireAllPermissions(permissionCodes ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		user := GetCurrentUser(c)
		if user == nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "用户未认证",
			})
			c.Abort()
			return
		}

		// 检查用户是否有所有指定权限
		for _, code := range permissionCodes {
			if !HasPermission(user, code) {
				c.JSON(http.StatusForbidden, gin.H{
					"code":    403,
					"message": "权限不足",
				})
				c.Abort()
				return
			}
		}

		c.Next()
	}
}

// RequireRole 要求特定角色的中间件
func RequireRole(roleName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		user := GetCurrentUser(c)
		if user == nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "用户未认证",
			})
			c.Abort()
			return
		}

		// 检查用户是否有指定角色
		if !HasRole(user, roleName) {
			c.JSON(http.StatusForbidden, gin.H{
				"code":    403,
				"message": "角色权限不足",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireAnyRole 要求任意一个角色的中间件
func RequireAnyRole(roleNames ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		user := GetCurrentUser(c)
		if user == nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "用户未认证",
			})
			c.Abort()
			return
		}

		// 检查用户是否有任意一个指定角色
		for _, roleName := range roleNames {
			if HasRole(user, roleName) {
				c.Next()
				return
			}
		}

		c.JSON(http.StatusForbidden, gin.H{
			"code":    403,
			"message": "角色权限不足",
		})
		c.Abort()
	}
}

// HasPermission 检查用户是否有指定权限
func HasPermission(user *model.User, permissionCode string) bool {
	// 重新加载用户的角色和权限信息
	var userWithRoles model.User
	if err := db.DB.Preload("Roles.Permissions").First(&userWithRoles, user.ID).Error; err != nil {
		return false
	}

	for _, role := range userWithRoles.Roles {
		if role.Status != 1 {
			continue // 跳过禁用的角色
		}
		for _, permission := range role.Permissions {
			if permission.Status == 1 && permission.Code == permissionCode {
				return true
			}
		}
	}
	return false
}

// HasRole 检查用户是否有指定角色
func HasRole(user *model.User, roleName string) bool {
	for _, role := range user.Roles {
		if role.Status == 1 && role.Name == roleName {
			return true
		}
	}
	return false
}

// GetUserPermissions 获取用户所有权限
func GetUserPermissions(user *model.User) []string {
	var permissions []string
	var userWithRoles model.User

	if err := db.DB.Preload("Roles.Permissions").First(&userWithRoles, user.ID).Error; err != nil {
		return permissions
	}

	permissionMap := make(map[string]bool)
	for _, role := range userWithRoles.Roles {
		if role.Status != 1 {
			continue
		}
		for _, permission := range role.Permissions {
			if permission.Status == 1 {
				permissionMap[permission.Code] = true
			}
		}
	}

	for code := range permissionMap {
		permissions = append(permissions, code)
	}
	return permissions
}

// GetUserRoles 获取用户所有角色
func GetUserRoles(user *model.User) []string {
	var roles []string
	for _, role := range user.Roles {
		if role.Status == 1 {
			roles = append(roles, role.Name)
		}
	}
	return roles
}
