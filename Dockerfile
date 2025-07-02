# --------- 前端构建阶段 ---------
FROM node:18 AS frontend-build

WORKDIR /app/frontend

COPY octoops-frontend/package*.json ./
RUN npm install

COPY octoops-frontend/ ./
RUN npm run build

# --------- 后端构建阶段 ---------
FROM golang:1.21 AS backend-build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# 将前端构建产物复制到后端静态目录
RUN rm -rf web/public && \
    mkdir -p web/public && \
    cp -r octoops-frontend/dist/* web/public/

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o octoops

# --------- 运行阶段 ---------
FROM debian:bullseye-slim

WORKDIR /app

COPY --from=backend-build /app/octoops /app/
COPY --from=backend-build /app/web/public /app/web/public

EXPOSE 8080

CMD ["./octoops"] 