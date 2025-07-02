# octoops 一体化部署说明

## 项目简介

本项目为 octoops 的前后端一体化管理平台，包含：
- **后端**：基于 Go 语言开发，提供任务调度、作业管理等 API 服务
- **前端**：基于 Vue3 + Element Plus，提供现代化的可视化管理界面

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
cd octoops-frontend
npm install
npm run dev
```

此时前端会自动代理 API 请求到本地后端（见 `vite.config.js` 配置）

## 目录结构说明

```
octoops/
  ├── api/           # Go 后端 API 相关
  ├── config/        # 配置文件
  ├── db/            # 数据库相关
  ├── main.go        # 后端入口
  ├── web/public/    # 前端构建产物（自动生成）
  ├── octoops-frontend/ # 前端源码
  └── Dockerfile     # 一体化构建脚本
```

## 常见问题 FAQ

1. **如何更新前端页面？**
   - 修改 `octoops-frontend` 目录下代码后，重新构建镜像即可。

2. **如何自定义端口？**
   - 修改 `main.go` 中的 `r.Run(":8080")`，或运行容器时映射其他端口。

3. **访问页面空白或 404？**
   - 请确保前端已正确构建，且 `web/public` 目录下有内容。
   - 通过 Dockerfile 构建时会自动完成此步骤。

4. **API 无法访问？**
   - 检查容器端口映射和防火墙设置。

## 贡献与反馈

如有建议或问题，欢迎提 Issue 或 PR。

---

感谢使用 octoops！ 