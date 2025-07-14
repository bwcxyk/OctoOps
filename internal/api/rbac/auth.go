package api

import (
	"net/http"
	"octoops/internal/db"
	"octoops/internal/middleware"
	"octoops/internal/model/rbac"
	"octoops/internal/pkg/jwt"
	"octoops/internal/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// RegisterRequest 注册请求
type RegisterRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Nickname string `json:"nickname"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	Token        string   `json:"token"`
	User         UserInfo `json:"user"`
	Roles        []string `json:"roles"`
	Permissions  []string `json:"permissions"`
}

// UserInfo 用户信息
type UserInfo struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
	Status   int    `json:"status"`
}

// RegisterAuthRoutes 注册认证路由
func RegisterAuthRoutes(r *gin.RouterGroup) {
	auth := r.Group("/auth")
	{
		auth.POST("/login", login)
		auth.POST("/register", register)
		auth.GET("/profile", middleware.AuthMiddleware(), getProfile)
		auth.POST("/logout", middleware.AuthMiddleware(), logout)
		auth.GET("/permissions", middleware.AuthMiddleware(), getUserPermissions)
	}
}

// login 用户登录
func login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请求参数错误",
			"error":   err.Error(),
		})
		return
	}

	// 查询用户
	var user model.User
	if err := db.DB.Preload("Roles").Where("username = ?", req.Username).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "用户名或密码错误",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "数据库查询错误",
			})
		}
		return
	}

	// 验证密码
	if !utils.VerifyPassword(req.Password, user.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "用户名或密码错误",
		})
		return
	}

	// 检查用户状态
	if user.Status != 1 {
		c.JSON(http.StatusForbidden, gin.H{
			"code":    403,
			"message": "用户已被禁用",
		})
		return
	}

	// 获取用户角色名称
	var roleNames []string
	for _, role := range user.Roles {
		if role.Status == 1 {
			roleNames = append(roleNames, role.Name)
		}
	}

	// 生成JWT token
	token, err := jwt.GenerateToken(user.ID, user.Username, roleNames)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "生成token失败",
		})
		return
	}

	// 获取用户权限
	permissions := middleware.GetUserPermissions(&user)

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "登录成功",
		"data": LoginResponse{
			Token:       token,
			User:        UserInfo{
				ID:       user.ID,
				Username: user.Username,
				Email:    user.Email,
				Nickname: user.Nickname,
				Avatar:   user.Avatar,
				Status:   user.Status,
			},
			Roles:       roleNames,
			Permissions: permissions,
		},
	})
}

// register 用户注册
func register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请求参数错误",
			"error":   err.Error(),
		})
		return
	}

	// 检查用户名是否已存在
	var existingUser model.User
	if err := db.DB.Where("username = ?", req.Username).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "用户名已存在",
		})
		return
	}

	// 检查邮箱是否已存在
	if err := db.DB.Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "邮箱已存在",
		})
		return
	}

	// 加密密码
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "密码加密失败",
		})
		return
	}

	// 创建用户
	user := model.User{
		Username: req.Username,
		Password: hashedPassword,
		Email:    req.Email,
		Nickname: req.Nickname,
		Status:   1,
	}

	if err := db.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "创建用户失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "注册成功",
		"data": UserInfo{
			ID:       user.ID,
			Username: user.Username,
			Email:    user.Email,
			Nickname: user.Nickname,
			Avatar:   user.Avatar,
			Status:   user.Status,
		},
	})
}

// getProfile 获取用户信息
func getProfile(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "用户未认证",
		})
		return
	}

	// 重新加载用户信息
	if err := db.DB.Preload("Roles").First(user, user.ID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取用户信息失败",
		})
		return
	}

	// 获取用户角色和权限
	roleNames := middleware.GetUserRoles(user)
	permissions := middleware.GetUserPermissions(user)

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data": gin.H{
			"user": UserInfo{
				ID:       user.ID,
				Username: user.Username,
				Email:    user.Email,
				Nickname: user.Nickname,
				Avatar:   user.Avatar,
				Status:   user.Status,
			},
			"roles":       roleNames,
			"permissions": permissions,
		},
	})
}

// logout 用户登出
func logout(c *gin.Context) {
	// 在实际应用中，可以将token加入黑名单
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "登出成功",
	})
}

// getUserPermissions 获取用户权限
func getUserPermissions(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "用户未认证",
		})
		return
	}

	permissions := middleware.GetUserPermissions(user)
	roles := middleware.GetUserRoles(user)

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data": gin.H{
			"permissions": permissions,
			"roles":       roles,
		},
	})
} 