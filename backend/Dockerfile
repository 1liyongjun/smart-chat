# Stage 1: Build
FROM golang:1.21-alpine AS builder

WORKDIR /app

# 安装构建依赖
RUN apk add --no-cache gcc musl-dev

# 复制源码
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# 构建二进制
RUN CGO_ENABLED=1 GOOS=linux go build -o smart-chat .

# Stage 2: Run
FROM alpine:3.19

WORKDIR /app

# 安装运行时依赖（SQLite）
RUN apk add --no-cache ca-certificates sqlite-libs

# 从构建阶段复制二进制
COPY --from=builder /app/smart-chat .
COPY --from=builder /app/data ./data

# 创建数据目录
RUN mkdir -p ./data

# 暴露端口
EXPOSE 8000

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8000/api/health || exit 1

# 启动命令
ENTRYPOINT ["./smart-chat"]
CMD ["--port", "8000", "--db", "./data/smart_chat.db"]
