# RBAC 权限管理系统

## 概述

本项目实现了基于角色的访问控制（RBAC）系统，提供了完整的用户、角色、权限管理功能。

## 系统架构

### 数据模型

- **User（用户）**: 系统用户，可以分配多个角色
- **Role（角色）**: 权限的集合，用户通过角色获得权限
- **Permission（权限）**: 具体的操作权限，如查看、创建、更新、删除等
- **UserRole（用户角色关联）**: 用户和角色的多对多关系
- **RolePermission（角色权限关联）**: 角色和权限的多对多关系

### 权限设计

权限采用 `资源:操作` 的命名方式，例如：
- `user:read` - 用户查看权限
- `user:create` - 用户创建权限
- `role:update` - 角色更新权限
- `task:execute` - 任务执行权限

## 快速开始

### 1. 初始化RBAC系统

运行初始化脚本创建默认的用户、角色和权限：

```bash
go run scripts/init_rbac.go
```

这将创建：
- 默认管理员用户：`admin` / `admin123`
- 4个默认角色：超级管理员、管理员、操作员、观察者
- 完整的权限体系

### 2. 配置JWT密钥

在 `config.yaml` 中配置JWT密钥：

```yaml
octoops:
  auth:
    jwt_secret: "your-secret-key-change-in-production"
```

或通过环境变量设置：

```bash
export OCTOOPS_AUTH_JWT_SECRET="your-secret-key-change-in-production"
```

### 3. 启动服务

```bash
go run main.go
```

## API 接口

### 认证相关

#### 用户登录
```http
POST /api/auth/login
Content-Type: application/json

{
  "username": "admin",
  "password": "admin123"
}
```

#### 用户注册
```http
POST /api/auth/register
Content-Type: application/json

{
  "username": "newuser",
  "password": "password123",
  "email": "user@example.com",
  "nickname": "新用户"
}
```

#### 获取用户信息
```http
GET /api/auth/profile
Authorization: Bearer <token>
```

#### 获取用户权限
```http
GET /api/auth/permissions
Authorization: Bearer <token>
```

### 用户管理

#### 获取用户列表
```http
GET /api/users?page=1&page_size=10&username=admin
Authorization: Bearer <token>
```

#### 创建用户
```http
POST /api/users
Authorization: Bearer <token>
Content-Type: application/json

{
  "username": "newuser",
  "password": "password123",
  "email": "user@example.com",
  "nickname": "新用户",
  "role_ids": [1, 2]
}
```

#### 更新用户
```http
PUT /api/users/1
Authorization: Bearer <token>
Content-Type: application/json

{
  "email": "newemail@example.com",
  "nickname": "新昵称",
  "status": 1,
  "role_ids": [1, 3]
}
```

#### 删除用户
```http
DELETE /api/users/1
Authorization: Bearer <token>
```

#### 为用户分配角色
```http
POST /api/users/1/roles
Authorization: Bearer <token>
Content-Type: application/json

{
  "role_ids": [1, 2]
}
```

### 角色管理

#### 获取角色列表
```http
GET /api/roles?page=1&page_size=10&name=管理员
Authorization: Bearer <token>
```

#### 创建角色
```http
POST /api/roles
Authorization: Bearer <token>
Content-Type: application/json

{
  "name": "新角色",
  "description": "角色描述",
  "permission_ids": [1, 2, 3]
}
```

#### 更新角色
```http
PUT /api/roles/1
Authorization: Bearer <token>
Content-Type: application/json

{
  "name": "更新后的角色名",
  "description": "更新后的描述",
  "status": 1,
  "permission_ids": [1, 2, 3, 4]
}
```

#### 删除角色
```http
DELETE /api/roles/1
Authorization: Bearer <token>
```

#### 为角色分配权限
```http
POST /api/roles/1/permissions
Authorization: Bearer <token>
Content-Type: application/json

{
  "permission_ids": [1, 2, 3]
}
```

### 权限管理

#### 获取权限列表
```http
GET /api/permissions?page=1&page_size=10&code=user:read
Authorization: Bearer <token>
```

#### 创建权限
```http
POST /api/permissions
Authorization: Bearer <token>
Content-Type: application/json

{
  "name": "新权限",
  "code": "new:permission",
  "description": "权限描述",
  "type": "api",
  "path": "/api/new",
  "method": "GET"
}
```

#### 获取权限树
```http
GET /api/permissions/tree
Authorization: Bearer <token>
```

## 中间件使用

### 认证中间件

```go
// 需要认证的路由
r.GET("/protected", middleware.AuthMiddleware(), handler)
```

### 权限中间件

```go
// 需要特定权限
r.GET("/users", middleware.RequirePermission("user:read"), handler)

// 需要任意一个权限
r.POST("/users", middleware.RequireAnyPermission("user:create", "user:admin"), handler)

// 需要所有权限
r.PUT("/users", middleware.RequireAllPermissions("user:read", "user:update"), handler)
```

### 角色中间件

```go
// 需要特定角色
r.GET("/admin", middleware.RequireRole("超级管理员"), handler)

// 需要任意一个角色
r.GET("/manager", middleware.RequireAnyRole("超级管理员", "管理员"), handler)
```

## 默认角色和权限

### 超级管理员
- 拥有系统所有权限
- 可以管理用户、角色、权限
- 可以访问所有功能模块

### 管理员
- 拥有大部分管理权限
- 不能修改系统配置
- 可以管理用户、角色、权限

### 操作员
- 拥有任务执行和查看权限
- 可以创建和管理任务
- 可以查看和创建告警

### 观察者
- 只有查看权限
- 可以查看任务、告警、用户、角色、权限
- 不能进行任何修改操作

## 安全建议

1. **修改默认密码**: 首次登录后立即修改默认管理员密码
2. **JWT密钥**: 在生产环境中使用强密钥，定期更换
3. **权限最小化**: 遵循最小权限原则，只分配必要的权限
4. **定期审计**: 定期检查用户权限分配情况
5. **HTTPS**: 在生产环境中使用HTTPS传输

## 扩展说明

### 添加新的权限

1. 在数据库中创建新的权限记录
2. 为相关角色分配新权限
3. 在API中使用权限中间件进行保护

### 自定义权限检查

```go
// 在业务逻辑中检查权限
user := middleware.GetCurrentUser(c)
if middleware.HasPermission(user, "custom:permission") {
    // 执行有权限的操作
}
```

### 权限缓存

对于高并发场景，建议实现权限缓存机制：

```go
// 示例：使用Redis缓存用户权限
func GetUserPermissionsCached(userID uint) []string {
    // 从缓存获取权限
    // 如果缓存不存在，从数据库查询并缓存
}
```

## 故障排除

### 常见问题

1. **JWT解析失败**: 检查JWT密钥配置是否正确
2. **权限验证失败**: 确认用户角色和权限分配正确
3. **数据库连接失败**: 检查数据库配置和连接

### 调试模式

在开发环境中可以启用调试模式：

```go
gin.SetMode(gin.DebugMode)
```

### 日志查看

系统会记录详细的权限验证日志，可以通过日志排查问题。 