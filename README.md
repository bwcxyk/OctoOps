# octoops 一体化部署说明

## 项目简介

octoops 是一套前后端一体化的任务调度与作业管理平台，包含：

- **后端**：Go 语言开发，支持任务调度、RBAC 权限、作业日志、告警等
- **前端**：Vue3 + Element Plus，现代化管理界面

---

## 快速开始

### 1. 构建镜像

确保已安装 Docker，且当前目录为项目根目录。

```bash
# 构建一体化镜像
# 推荐使用国内源加速（如有需要）
docker build -t octoops-allinone .
```

### 2. 运行容器

```bash
docker run -d -p 8080:8080 --name octoops octoops-allinone
```

- 访问 `http://<服务器IP>:8080` 即可进入 octoops 管理平台
- 所有 API 也通过该端口对外提供

### 3. 前端开发调试（可选）

如需本地开发前端，可单独运行：

```bash
cd web
npm install
npm run dev
```
前端会自动代理 API 请求到本地后端（见 `web/vite.config.js` 配置）

---

## 目录结构说明

```
octoops/
├── cmd/                  # 后端启动入口（如 cmd/octoops/main.go）
├── internal/             # 后端核心代码
│   ├── api/              # 路由与 API 控制器（rbac, seatunnel, alert, aliyun, custom_task等）
│   ├── config/           # 配置加载
│   ├── db/               # 数据库初始化
│   ├── middleware/       # Gin 中间件（如认证、RBAC）
│   ├── model/            # 数据模型（rbac, seatunnel, alert, aliyun, custom_task等）
│   ├── pkg/              # JWT等通用包
│   ├── scheduler/        # 调度核心逻辑
│   ├── service/          # 业务服务层（alert, aliyun, seatunnel等）
│   └── utils/            # 工具函数（加密、邮件等）
├── web/                  # 前端源码（Vue3 + Element Plus）
│   ├── src/
│   │   ├── api/          # 前端 API 封装
│   │   ├── components/   # 通用组件
│   │   ├── layouts/      # 页面布局
│   │   ├── store/        # 状态管理
│   │   ├── utils/        # 前端工具
│   │   ├── views/        # 业务页面（rbac、seatunnel、alert等）
│   │   └── main.js       # 前端入口
│   ├── public/           # 前端静态资源
│   ├── package.json      # 前端依赖
│   └── vite.config.js    # 前端构建配置
├── config.yaml           # 主配置文件
├── config.yaml.example   # 配置模板
├── Dockerfile            # 一体化构建脚本
├── go.mod                # Go 依赖
├── go.sum
└── README.md             # 项目说明
```

---

## 常见问题 FAQ

1. **如何更新前端页面？**
   - 修改 `web` 目录下代码后，重新构建镜像即可。

2. **如何自定义端口？**
   - 修改 `cmd/octoops/main.go` 中的监听端口，或运行容器时映射其他端口。

3. **访问页面空白或 404？**
   - 请确保前端已正确构建，且 `web/public` 目录下有内容。
   - 通过 Dockerfile 构建时会自动完成此步骤。

4. **API 无法访问？**
   - 检查容器端口映射和防火墙设置。

---

## 贡献与反馈

如有建议或问题，欢迎提 Issue 或 PR。

---

感谢使用 octoops！ 