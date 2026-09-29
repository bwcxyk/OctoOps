# --------- 前端构建阶段 ---------
FROM node:24.21.0 AS frontend-build

WORKDIR /app/web

RUN corepack enable

COPY web/package.json web/pnpm-lock.yaml web/pnpm-workspace.yaml ./
RUN pnpm install --frozen-lockfile

COPY web/ ./
RUN pnpm run build

# --------- 后端构建阶段 ---------
FROM golang:1.27.1 AS backend-build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# 将前端构建产物复制到嵌入目录，go:embed 会在编译时打包进二进制
COPY --from=frontend-build /app/web/dist /app/internal/assets/dist

# 复制 example 文件为 config.yaml
COPY config.yaml.example /app/config.yaml

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -tags embed_frontend -o octoops ./cmd/octoops/main.go
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o octoops-init ./cmd/init-rbac/main.go

# --------- 运行阶段 ---------
FROM debian:trixie-slim
ENV TZ=Asia/Shanghai

WORKDIR /app

RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*

COPY --from=backend-build /app/octoops /app/
COPY --from=backend-build /app/octoops-init /app/
COPY --from=backend-build /app/config.yaml /app/config.yaml

EXPOSE 8080

CMD ["./octoops"]
