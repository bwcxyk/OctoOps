# OctoOps Web

OctoOps 的前端工程，基于 `Vue 3`、`TypeScript`、`Vite`、`Pinia`、`TDesign`。

## 环境要求

- Node.js `>= 18.18.0`（建议 Node 20+）
- npm（仓库默认使用 `package-lock.json`）

## 快速开始

```bash
npm install
npm run dev
```

- 默认启动地址：`http://localhost:3002`
- 默认代理后端：`http://127.0.0.1:8080`

## 常用命令

```bash
# 开发
npm run dev
npm run dev:linux
npm run dev:mock

# 构建
npm run build
npm run build:test
npm run build:site

# 预览
npm run preview

# 代码检查
npm run lint
npm run lint:fix
npm run stylelint
npm run stylelint:fix
```

## 环境变量

环境变量由 `.env*` 文件管理（例如 `.env.development`、`.env.test`）。

- `VITE_BASE_URL`：应用部署基础路径
- `VITE_API_URL_PREFIX`：接口前缀（默认 `/api`）
- `VITE_API_URL`：后端服务地址
- `VITE_IS_REQUEST_PROXY`：是否启用请求代理标记

开发模式下的代理配置见 `vite.config.ts`：

- 代理前缀：`VITE_API_URL_PREFIX`
- 代理目标：`VITE_API_URL`（为空时回退 `http://127.0.0.1:8080`）

## 目录说明

```text
web/
├── src/
│   ├── api/              # 接口封装
│   ├── components/       # 通用组件
│   ├── layouts/          # 布局
│   ├── pages/            # 页面
│   ├── router/           # 路由
│   ├── store/            # 状态管理
│   └── style/            # 样式与主题变量
├── public/               # 静态资源
├── mock/                 # Mock 数据
├── vite.config.ts        # Vite 配置
├── package.json
└── PUBLISH.md            # 发布流程
```

## 发布

发布流程请查看 [PUBLISH.md](./PUBLISH.md)。
