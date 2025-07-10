package api

import (
	"log"
	"math/rand"
	"net/http"
	"octoops/internal/db"
	"octoops/internal/middleware"
	"octoops/internal/model/rbac"
	"octoops/internal/utils"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// CreateUserRequest 创建用户请求
type CreateUserRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Nickname string `json:"nickname"`
	RoleIDs  []uint `json:"role_ids"`
}

// UpdateUserRequest 更新用户请求
type UpdateUserRequest struct {
	Email    string `json:"email"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
	Status   *int   `json:"status"`
	RoleIDs  []uint `json:"role_ids"`
}

// ChangePasswordRequest 修改密码请求
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required"`
}

// ForgotPasswordRequest 忘记密码请求
// 用户名+邮箱+新密码
// 生产环境建议加验证码
//
type ForgotPasswordRequest struct {
	Username    string `json:"username" binding:"required"`
	Email       string `json:"email" binding:"required,email"`
	NewPassword string `json:"new_password" binding:"required"`
}

// 简单内存验证码存储（生产建议用 Redis）
var resetCodeStore = struct {
	sync.RWMutex
	m map[string]resetCodeEntry
}{m: make(map[string]resetCodeEntry)}

type resetCodeEntry struct {
	Code      string
	ExpiresAt time.Time
}

// SendResetCodeRequest 发送验证码请求
//
type SendResetCodeRequest struct {
	Email    string `json:"email" binding:"required,email"`
}

// RegisterUserRoutes 注册用户管理路由
func RegisterUserRoutes(r *gin.RouterGroup) {
	users := r.Group("/users")
	users.Use(middleware.AuthMiddleware())
	{
		users.GET("", middleware.RequirePermission("rbac:user:read"), getUsers)
		users.GET("/:id", middleware.RequirePermission("rbac:user:read"), getUser)
		users.POST("", middleware.RequirePermission("rbac:user:create"), createUser)
		users.PUT("/:id", middleware.RequirePermission("rbac:user:update"), updateUser)
		users.DELETE("/:id", middleware.RequirePermission("rbac:user:delete"), deleteUser)
		users.POST("/:id/roles", middleware.RequirePermission("rbac:user:assign_role"), assignRoles)
		users.DELETE("/:id/roles", middleware.RequirePermission("rbac:user:assign_role"), removeRoles)
		users.POST("/change-password", changePassword)
	}
	// 忘记密码接口无需登录
	r.POST("/users/forgot-password", forgotPassword)
	// 发送验证码接口
	r.POST("/users/send-reset-code", sendResetCode)
}

// getUsers 获取用户列表
func getUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	username := c.Query("username")
	email := c.Query("email")
	status := c.Query("status")

	query := db.DB.Model(&model.User{}).Preload("Roles")

	// 添加查询条件
	if username != "" {
		query = query.Where("username LIKE ?", "%"+username+"%")
	}
	if email != "" {
		query = query.Where("email LIKE ?", "%"+email+"%")
	}
	if status != "" {
		if statusInt, err := strconv.Atoi(status); err == nil {
			query = query.Where("status = ?", statusInt)
		}
	}

	var total int64
	query.Count(&total)

	var users []model.User
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取用户列表失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data": gin.H{
			"users":      users,
			"total":      total,
			"page":       page,
			"page_size":  pageSize,
			"total_page": (int(total) + pageSize - 1) / pageSize,
		},
	})
}

// getUser 获取单个用户
func getUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "用户ID格式错误",
		})
		return
	}

	var user model.User
	if err := db.DB.Preload("Roles").First(&user, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "用户不存在",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "获取用户信息失败",
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data":    user,
	})
}

// createUser 创建用户
func createUser(c *gin.Context) {
	var req CreateUserRequest
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

	// 开始事务
	tx := db.DB.Begin()

	// 创建用户
	user := model.User{
		Username: req.Username,
		Password: hashedPassword,
		Email:    req.Email,
		Nickname: req.Nickname,
		Status:   1,
	}

	if err := tx.Create(&user).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "创建用户失败",
		})
		return
	}

	// 分配角色
	if len(req.RoleIDs) > 0 {
		var userRoles []model.UserRole
		for _, roleID := range req.RoleIDs {
			userRoles = append(userRoles, model.UserRole{
				UserID: user.ID,
				RoleID: roleID,
			})
		}
		if err := tx.Create(&userRoles).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "分配角色失败",
			})
			return
		}
	}

	tx.Commit()

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "创建成功",
		"data":    user,
	})
}

// updateUser 更新用户
func updateUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "用户ID格式错误",
		})
		return
	}

	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请求参数错误",
			"error":   err.Error(),
		})
		return
	}

	// 检查用户是否存在
	var user model.User
	if err := db.DB.First(&user, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "用户不存在",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "获取用户信息失败",
			})
		}
		return
	}

	// 开始事务
	tx := db.DB.Begin()

	// 更新用户信息
	updates := make(map[string]interface{})
	if req.Email != "" {
		updates["email"] = req.Email
	}
	if req.Nickname != "" {
		updates["nickname"] = req.Nickname
	}
	if req.Avatar != "" {
		updates["avatar"] = req.Avatar
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}

	if len(updates) > 0 {
		if err := tx.Model(&user).Updates(updates).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "更新用户信息失败",
			})
			return
		}
	}

	// 更新角色
	if req.RoleIDs != nil {
		// 删除现有角色
		if err := tx.Where("user_id = ?", user.ID).Delete(&model.UserRole{}).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "删除用户角色失败",
			})
			return
		}

		// 分配新角色
		if len(req.RoleIDs) > 0 {
			var userRoles []model.UserRole
			for _, roleID := range req.RoleIDs {
				userRoles = append(userRoles, model.UserRole{
					UserID: user.ID,
					RoleID: roleID,
				})
			}
			if err := tx.Create(&userRoles).Error; err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{
					"code":    500,
					"message": "分配角色失败",
				})
				return
			}
		}
	}

	tx.Commit()

	// 重新加载用户信息
	db.DB.Preload("Roles").First(&user, user.ID)

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "更新成功",
		"data":    user,
	})
}

// deleteUser 删除用户
func deleteUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "用户ID格式错误",
		})
		return
	}

	// 开始事务
	tx := db.DB.Begin()

	// 删除用户角色关联
	if err := tx.Where("user_id = ?", id).Delete(&model.UserRole{}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "删除用户角色失败",
		})
		return
	}

	// 删除用户
	if err := tx.Delete(&model.User{}, id).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "删除用户失败",
		})
		return
	}

	tx.Commit()

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "删除成功",
	})
}

// assignRoles 分配角色
func assignRoles(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "用户ID格式错误",
		})
		return
	}

	var req struct {
		RoleIDs []uint `json:"role_ids" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请求参数错误",
		})
		return
	}

	// 检查用户是否存在
	var user model.User
	if err := db.DB.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "用户不存在",
		})
		return
	}

	// 分配角色
	var userRoles []model.UserRole
	for _, roleID := range req.RoleIDs {
		userRoles = append(userRoles, model.UserRole{
			UserID: user.ID,
			RoleID: roleID,
		})
	}

	if err := db.DB.Create(&userRoles).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "分配角色失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "分配角色成功",
	})
}

// removeRoles 移除角色
func removeRoles(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "用户ID格式错误",
		})
		return
	}

	var req struct {
		RoleIDs []uint `json:"role_ids" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请求参数错误",
		})
		return
	}

	// 移除角色
	if err := db.DB.Where("user_id = ? AND role_id IN ?", id, req.RoleIDs).Delete(&model.UserRole{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "移除角色失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "移除角色成功",
	})
}

// changePassword 修改密码
func changePassword(c *gin.Context) {
	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请求参数错误",
		})
		return
	}

	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "用户未认证",
		})
		return
	}

	// 验证旧密码
	if !utils.VerifyPassword(req.OldPassword, user.Password) {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "旧密码错误",
		})
		return
	}

	// 加密新密码
	hashedPassword, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "密码加密失败",
		})
		return
	}

	// 更新密码
	if err := db.DB.Model(&user).Update("password", hashedPassword).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "修改密码失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "修改密码成功",
	})
}

// sendResetCode 发送邮箱验证码
func sendResetCode(c *gin.Context) {
	var req SendResetCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请求参数错误", "error": err.Error()})
		return
	}
	var user model.User
	if err := db.DB.Where("email = ?", req.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "邮箱未注册"})
		return
	}
	// 生成6位验证码
	code := ""
	rand.Seed(time.Now().UnixNano())
	for i := 0; i < 6; i++ {
		code += string('0' + rand.Intn(10))
	}
	// 存储验证码，5分钟有效
	resetCodeStore.Lock()
	resetCodeStore.m[req.Email] = resetCodeEntry{Code: code, ExpiresAt: time.Now().Add(5 * time.Minute)}
	resetCodeStore.Unlock()
	// TODO: 发送邮件（此处仅打印，生产环境请集成邮件服务）
	subject := "OctoOps 验证码"
	body := "您的验证码为 <b>" + code + "</b>，5分钟内有效。"
	if err := utils.SendMail(utils.MailOptions{
		To:      req.Email,
		Subject: subject,
		Body:    body,
	}); err != nil {
		log.Printf("发送邮件失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "邮件发送失败"})
		return
	}
	log.Printf("向 %s 发送验证码: %s", req.Email, code)
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "验证码已发送到邮箱"})
}

// forgotPassword 忘记密码
func forgotPassword(c *gin.Context) {
	type ForgotPasswordRequest struct {
		Email       string `json:"email" binding:"required,email"`
		Code        string `json:"code" binding:"required"`
		NewPassword string `json:"new_password" binding:"required"`
	}
	var req ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请求参数错误", "error": err.Error()})
		return
	}
	// 校验验证码
	resetCodeStore.RLock()
	entry, ok := resetCodeStore.m[req.Email]
	resetCodeStore.RUnlock()
	if !ok || entry.ExpiresAt.Before(time.Now()) || entry.Code != req.Code {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "验证码错误或已过期"})
		return
	}
	// 验证通过后删除验证码
	resetCodeStore.Lock()
	delete(resetCodeStore.m, req.Email)
	resetCodeStore.Unlock()
	var user model.User
	if err := db.DB.Where("email = ?", req.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "邮箱未注册"})
		return
	}
	hashedPassword, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "密码加密失败"})
		return
	}
	if err := db.DB.Model(&user).Update("password", hashedPassword).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "重置密码失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "密码重置成功，请重新登录"})
}
