# --------- 前端构建阶段 ---------
FROM node:20 AS frontend-build

WORKDIR /app/web

COPY web/package*.json ./
RUN npm install

COPY web/ ./
RUN npm run build

# --------- 后端构建阶段 ---------
FROM golang:1.23 AS backend-build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

COPY --from=frontend-build /app/public /app/public

# 复制 example 文件为 config.yaml
COPY config.yaml.example /app/config.yaml

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o octoops ./cmd/octoops/main.go

# --------- 运行阶段 ---------
FROM debian:bullseye-slim

WORKDIR /app

RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*

COPY --from=backend-build /app/octoops /app/
COPY --from=backend-build /app/config.yaml /app/config.yaml
COPY --from=backend-build /app/public /app/public

EXPOSE 8080

CMD ["./octoops"]