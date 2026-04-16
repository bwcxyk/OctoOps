package rbac

import (
	"crypto/rand"
	"log"
	"math/big"
	"net/http"
	"octoops/internal/infra/postgres"
	"octoops/internal/middleware"
	"octoops/internal/model/rbac"
	"octoops/internal/utils"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const maxPageSize = 100

const (
	resetCodeTTL            = 5 * time.Minute
	resetCodeResendInterval = 60 * time.Second
	resetCodeSendWindow     = 1 * time.Hour
	resetCodeSendLimit      = 5
	resetCodeMaxAttempts    = 5
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
type ForgotPasswordRequest struct {
	Username    string `json:"username" binding:"required"`
	Email       string `json:"email" binding:"required,email"`
	NewPassword string `json:"new_password" binding:"required"`
}

type resetCodeEntry struct {
	Code           string
	ExpiresAt      time.Time
	VerifyAttempts int
}

type resetRateEntry struct {
	WindowStart time.Time
	SendCount   int
	LastSentAt  time.Time
}

// SendResetCodeRequest 发送验证码请求
type SendResetCodeRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// RegisterUserRoutes 注册用户管理路由
func RegisterUserRoutes(r *gin.RouterGroup) {
	users := r.Group("/users")
	// 忘记密码接口无需登录
	users.POST("/forgot-password", forgotPassword)
	// 发送验证码接口
	users.POST("/send-reset-code", sendResetCode)

	authUsers := r.Group("/users")
	authUsers.Use(middleware.AuthMiddleware())
	{
		authUsers.GET("", middleware.RequirePermission("rbac:user:read"), getUsers)
		authUsers.GET("/:id", middleware.RequirePermission("rbac:user:read"), getUser)
		authUsers.POST("", middleware.RequirePermission("rbac:user:create"), createUser)
		authUsers.PUT("/:id", middleware.RequirePermission("rbac:user:update"), updateUser)
		authUsers.DELETE("/:id", middleware.RequirePermission("rbac:user:delete"), deleteUser)
		authUsers.POST("/:id/roles", middleware.RequirePermission("rbac:user:assign_role"), assignRoles)
		authUsers.DELETE("/:id/roles", middleware.RequirePermission("rbac:user:assign_role"), removeRoles)
		authUsers.POST("/change-password", changePassword)
	}
}

// getUsers 获取用户列表
func getUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	username := c.Query("username")
	email := c.Query("email")
	status := c.Query("status")
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > maxPageSize {
		pageSize = maxPageSize
	}

	query := postgres.DB.Model(&model.User{}).Preload("Roles").Where("is_super_admin = ?", false)

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
	if err := query.Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "统计用户数量失败",
		})
		return
	}

	var users []model.User
	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC, id DESC").Offset(offset).Limit(pageSize).Find(&users).Error; err != nil {
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
	if err := postgres.DB.Preload("Roles").First(&user, id).Error; err != nil {
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
	if err := postgres.DB.Where("username = ?", req.Username).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "用户名已存在",
		})
		return
	}

	// 检查邮箱是否已存在
	if err := postgres.DB.Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "邮箱已存在",
		})
		return
	}

	if err := utils.ValidatePasswordComplexity(req.Password); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": err.Error(),
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
	tx := postgres.DB.Begin()

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

	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "创建用户失败",
		})
		return
	}

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
	if err := postgres.DB.First(&user, id).Error; err != nil {
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
	tx := postgres.DB.Begin()

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

	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "更新用户失败",
		})
		return
	}

	// 重新加载用户信息
	postgres.DB.Preload("Roles").First(&user, user.ID)

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
	tx := postgres.DB.Begin()

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

	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "删除用户失败",
		})
		return
	}

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
	if err := postgres.DB.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "用户不存在",
		})
		return
	}

	// 去重，避免重复 role_id 导致无效写入或唯一键冲突
	uniqueRoleIDs := make([]uint, 0, len(req.RoleIDs))
	roleIDSet := make(map[uint]struct{}, len(req.RoleIDs))
	for _, roleID := range req.RoleIDs {
		if _, exists := roleIDSet[roleID]; exists {
			continue
		}
		roleIDSet[roleID] = struct{}{}
		uniqueRoleIDs = append(uniqueRoleIDs, roleID)
	}
	if len(uniqueRoleIDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "role_ids 不能为空",
		})
		return
	}

	tx := postgres.DB.Begin()

	// 只插入缺失的角色关联，实现幂等
	var existingUserRoles []model.UserRole
	if err := tx.Where("user_id = ? AND role_id IN ?", user.ID, uniqueRoleIDs).Find(&existingUserRoles).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询现有角色关联失败",
		})
		return
	}

	existingRoleSet := make(map[uint]struct{}, len(existingUserRoles))
	for _, ur := range existingUserRoles {
		existingRoleSet[ur.RoleID] = struct{}{}
	}

	userRolesToCreate := make([]model.UserRole, 0, len(uniqueRoleIDs))
	for _, roleID := range uniqueRoleIDs {
		if _, exists := existingRoleSet[roleID]; exists {
			continue
		}
		userRolesToCreate = append(userRolesToCreate, model.UserRole{
			UserID: user.ID,
			RoleID: roleID,
		})
	}

	if len(userRolesToCreate) > 0 {
		if err := tx.Create(&userRolesToCreate).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "分配角色失败",
			})
			return
		}
	}

	if err := tx.Commit().Error; err != nil {
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
	if err := postgres.DB.Where("user_id = ? AND role_id IN ?", id, req.RoleIDs).Delete(&model.UserRole{}).Error; err != nil {
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

	if err := utils.ValidatePasswordComplexity(req.NewPassword); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": err.Error(),
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
	if err := postgres.DB.Model(&user).Update("password", hashedPassword).Error; err != nil {
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
	store, err := GetRecoveryStore()
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"code": 503, "message": "账号恢复服务未配置"})
		return
	}
	email := normalizeEmail(req.Email)
	now := time.Now()
	rateKey := buildResetRateKey(email, c.ClientIP())

	if !allowSendResetCode(store, rateKey, now) {
		c.JSON(http.StatusTooManyRequests, gin.H{
			"code":    429,
			"message": "请求过于频繁，请稍后再试",
		})
		return
	}

	var user model.User
	if err := postgres.DB.Where("email = ?", email).First(&user).Error; err != nil {
		// 防枚举：统一响应，不暴露邮箱是否存在。
		c.JSON(http.StatusOK, gin.H{"code": 200, "message": "如邮箱已注册，验证码已发送"})
		return
	}
	// 生成6位验证码
	code := ""
	for i := 0; i < 6; i++ {
		n, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "验证码生成失败"})
			return
		}
		code += strconv.FormatInt(n.Int64(), 10)
	}
	// 存储验证码，5分钟有效
	if err := store.SetCode(email, resetCodeEntry{
		Code:           code,
		ExpiresAt:      now.Add(resetCodeTTL),
		VerifyAttempts: 0,
	}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "验证码存储失败"})
		return
	}
	subject := "OctoOps 验证码"
	body := "您的验证码为 <b>" + code + "</b>，5分钟内有效。"
	if err := utils.SendMail(utils.MailOptions{
		To:      email,
		Subject: subject,
		Body:    body,
	}); err != nil {
		log.Printf("发送邮件失败: %v", err)
		// 防枚举：统一响应，不暴露邮箱是否存在或邮件系统状态。
		c.JSON(http.StatusOK, gin.H{"code": 200, "message": "如邮箱已注册，验证码已发送"})
		return
	}
	log.Printf("向 %s 发送验证码邮件成功", email)
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "如邮箱已注册，验证码已发送"})
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
	store, err := GetRecoveryStore()
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"code": 503, "message": "账号恢复服务未配置"})
		return
	}
	email := normalizeEmail(req.Email)
	// 校验验证码
	entry, ok, err := store.GetCode(email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "验证码校验失败"})
		return
	}
	if !ok || entry.ExpiresAt.Before(time.Now()) {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "验证码错误或已过期"})
		return
	}
	if entry.Code != req.Code {
		entry.VerifyAttempts++
		if entry.VerifyAttempts >= resetCodeMaxAttempts {
			if err := store.DeleteCode(email); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "验证码校验失败"})
				return
			}
		} else {
			if err := store.SetCode(email, entry); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "验证码校验失败"})
				return
			}
		}
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "验证码错误或已过期"})
		return
	}
	// 验证通过后删除验证码
	if err := store.DeleteCode(email); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "验证码校验失败"})
		return
	}
	var user model.User
	if err := postgres.DB.Where("email = ?", email).First(&user).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "验证码错误或已过期"})
		return
	}
	if err := utils.ValidatePasswordComplexity(req.NewPassword); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}
	hashedPassword, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "密码加密失败"})
		return
	}
	if err := postgres.DB.Model(&user).Update("password", hashedPassword).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "重置密码失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "密码重置成功，请重新登录"})
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func buildResetRateKey(email, ip string) string {
	return email + "|" + ip
}

func allowSendResetCode(store RecoveryStore, key string, now time.Time) bool {
	entry, exists, err := store.GetRate(key)
	if err != nil {
		return false
	}
	if !exists || now.Sub(entry.WindowStart) >= resetCodeSendWindow {
		err := store.SetRate(key, resetRateEntry{
			WindowStart: now,
			SendCount:   1,
			LastSentAt:  now,
		})
		return err == nil
	}

	if now.Sub(entry.LastSentAt) < resetCodeResendInterval {
		return false
	}
	if entry.SendCount >= resetCodeSendLimit {
		return false
	}

	entry.SendCount++
	entry.LastSentAt = now
	return store.SetRate(key, entry) == nil
}
