# OctoOps

一套前后端一体化的任务调度与作业管理平台。

- 后端：Go + Gin，提供任务调度、RBAC、告警、作业相关 API
- 前端：Vue 3 + TypeScript + TDesign（位于 `web/`）
- 部署：支持 Docker 一体化部署

## 功能概览

- 任务中心：调度器、自定义任务、任务日志
- 数据集成：基于 SeaTunnel 实现流批一体的数据同步与作业编排，通过 [REST API V2](https://seatunnel.incubator.apache.org/docs/engines/zeta/rest-api-v2) 对接执行能力
- 云资源：阿里云相关能力
- 告警体系：告警渠道、告警组、告警模板
- 权限体系：用户、角色、权限（RBAC）

## 快速开始（Docker 推荐）

### 1) 构建镜像

```bash
docker build -t octoops-allinone .
```

### 2) 启动服务

```bash
docker run -d --name octoops -p 8080:8080 octoops-allinone
```

### 3) 初始化 RBAC 与管理员账号

```bash
docker exec -it octoops /app/octoops-init
```

初始化后会创建默认管理员：

- 用户名：`admin`
- 密码：随机生成（执行 `octoops-init` 时输出在日志中）

首次登录后请立即修改密码。

### 4) 访问系统

- 前端页面：`http://<服务器IP>:8080`
- API 入口：`http://<服务器IP>:8080/api`

## 本地开发

### 前置依赖

- Go（建议 1.24+）
- Node.js（建议 20+）
- PostgreSQL
- 可选：SeaTunnel、SMTP 服务、Docker

### 后端开发

1. 复制配置文件并按需修改：

```bash
cp config.yaml.example config.yaml
```

2. 启动后端：

```bash
go run ./cmd/octoops/main.go
```

后端默认监听 `8080` 端口（可通过 `config.yaml` 的 `octoops.server.port` 修改）。

3. 运行后端单元测试：

```bash
go test ./... -v
```

### 前端开发

```bash
cd web
npm install
npm run dev
```

- 前端开发服务默认端口：`3002`
- 开发代理配置在 `web/vite.config.ts`，默认会代理到 `http://127.0.0.1:8080`

## 配置说明

核心配置文件为 `config.yaml`，可基于 `config.yaml.example` 修改。

主要配置项：

- `postgres`：数据库连接信息
- `octoops.auth.jwt_secret`：JWT 密钥
- `octoops.mail`：SMTP 邮件告警配置
- `octoops.server.port`：后端服务监听端口
- `octoops.redis`：可选 Redis 配置（密码找回验证码与限流存储）
- `seatunnel.base_url`：SeaTunnel API 地址
- `octoops.aliyun.aes_key`：阿里云密钥加密用 AES Key（32 字节）

## 项目结构

```text
octoops/
├── cmd/                    # 可执行入口（服务、初始化脚本）
│   ├── octoops/            # 主服务入口
│   └── init-rbac/          # RBAC 初始化入口
├── internal/               # 后端核心代码（api/service/model/scheduler 等）
├── web/                    # 前端工程（Vue3 + TS + TDesign）
├── config.yaml.example     # 配置模板
├── Dockerfile              # 一体化构建与运行镜像
└── README.md
```

## 常见问题（FAQ）

1. 页面 404 或空白怎么办？

- 先确认后端服务已正常启动且监听 `8080`。
- 如果是本地运行后端，请确认 `frontend/` 静态资源存在（Docker 构建会自动生成并复制）。

2. 前端请求接口失败怎么办？

- 检查 `web/vite.config.ts` 中的代理目标地址。
- 检查后端 `8080` 端口与数据库连接是否正常。

3. 如何更新前端页面？

- 修改 `web/` 代码后，重新执行 `npm run build`（或重新构建 Docker 镜像）。

4. 如何更改服务端口？

- 在 `config.yaml` 中设置 `octoops.server.port`，重启服务生效。
- 也可通过环境变量 `OCTOOPS_SERVER_PORT` 覆盖配置值。

## 相关文档

- 前端说明：`web/README.md`
- 前端发布流程：`web/PUBLISH.md`

## 贡献

欢迎提交 Issue 或 Pull Request 来改进项目。
