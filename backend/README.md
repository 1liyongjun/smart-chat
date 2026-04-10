# 智能客服后端 (Go + Eino)

基于 Go + 字节跳动 Eino 框架的智能客服后端服务。

## 技术栈

- **Web 框架**: Fiber v2 (高性能 Go Web 框架)
- **AI 框架**: CloudWeGo Eino (字节跳动开源的 LLM 应用开发框架)
- **数据库**: SQLite (通过 mattn/go-sqlite3)
- **AI 模型**: MiniMax-M2.7

## 依赖

```bash
go mod tidy
```

主要依赖：
- `github.com/cloudwego/eino` - Eino AI 框架
- `github.com/cloudwego/eino-ext` - Eino 扩展组件
- `github.com/gofiber/fiber/v2` - Web 框架
- `github.com/mattn/go-sqlite3` - SQLite 驱动

## 配置

通过命令行参数或环境变量配置：

| 参数 | 环境变量 | 默认值 | 说明 |
|------|---------|--------|------|
| `-port` | - | `8000` | 服务端口 |
| `-db` | - | `./data/smart_chat.db` | 数据库路径 |
| `-api-key` | `MINIMAX_API_KEY` | - | MiniMax API Key |
| `-base-url` | `MINIMAX_BASE_URL` | `https://api.minimax.io/v1` | API 地址 |
| `-model` | - | `MiniMax-M2.7` | 模型名称 |
| `-cors` | - | `*` | CORS 允许的源 |

## 运行

```bash
# 安装依赖
go mod tidy

# 运行
go run main.go -api-key="your-api-key"
```

## API 接口

### 健康检查
```
GET /api/health
```

### 知识库管理
```
POST   /api/knowledge       # 添加知识库条目
GET    /api/knowledge       # 获取知识库列表
PUT    /api/knowledge/:id   # 更新知识库条目
DELETE /api/knowledge/:id   # 删除知识库条目
```

### 聊天
```
POST /api/chat          # 普通聊天
POST /api/chat/stream   # 流式聊天 (SSE)
```

### 对话
```
GET /api/conversations           # 获取对话列表
GET /api/conversations/:id       # 获取对话详情
```

### 统计
```
GET /api/stats   # 获取统计数据
```

## Docker 部署

```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=1 GOOS=linux go build -o server .

FROM alpine
RUN apk add --no-cache ca-certificates sqlite-libs
WORKDIR /app
COPY --from=builder /app/server .
COPY --from=builder /app/data ./data
CMD ["./server", "-db", "./data/smart_chat.db"]
```
