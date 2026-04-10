# 部署指南

## 快速部署到 Railway (推荐 ⭐)

### 步骤

1. **登录 Railway**
   - 访问 [railway.app](https://railway.app)
   - 使用 GitHub 账号登录

2. **创建项目**
   - 点击 **New Project → Deploy from GitHub repo**
   - 选择仓库 `1liyongjun/smart-chat`
   - Railway 会自动检测 `railway.toml`

3. **配置环境变量**
   - 在项目 Settings → Variables 中添加：
     ```
     MINIMAX_API_KEY=你的MiniMax_API密钥
     ```

4. **部署**
   - Railway 会自动构建 Docker 镜像并部署
   - 等待完成后，你会获得一个 URL：`https://xxx.railway.app`

---

## Docker 部署 (本地/服务器)

### 本地构建

```bash
cd backend

# 构建镜像
docker build -t smart-chat .

# 运行
docker run -d \
  --name smart-chat \
  -p 8000:8000 \
  -v smart-chat-data:/app/data \
  -e MINIMAX_API_KEY=你的API密钥 \
  smart-chat
```

### 使用 Docker Compose (推荐)

```bash
cd backend

# 创建 .env 文件
cp .env.example .env
# 编辑 .env，填入你的 API Key

# 启动服务
docker-compose up -d

# 查看日志
docker-compose logs -f
```

---

## 部署到其他平台

### Render
1. 创建 Render 账号
2. New → Web Service → 连接 GitHub 仓库
3. 设置：
   - Build Command: `go build -o smart-chat .`
   - Start Command: `./smart-chat --port $PORT --db ./data/smart_chat.db`
4. 添加环境变量 `MINIMAX_API_KEY`

### Fly.io
```bash
# 安装 flyctl
curl -L https://fly.io/install.sh | sh

# 登录
fly auth login

# 部署
cd backend
fly launch
fly secrets set MINIMAX_API_KEY=你的密钥
fly deploy
```

### Zeabur (国产推荐)
1. 访问 [zeabur.com](https://zeabur.com)
2. 连接 GitHub 仓库
3. 一键部署，自动生成 URL

---

## 验证部署

部署完成后，访问：

```
https://你的域名/api/health
```

返回以下内容表示成功：
```json
{"status": "ok", "service": "smart-chat"}
```

---

## API 使用

### 聊天接口
```bash
curl -X POST https://你的域名/api/chat \
  -H "Content-Type: application/json" \
  -d '{"message": "你好", "visitor_id": "user001"}'
```

### 添加知识库
```bash
curl -X POST https://你的域名/api/knowledge \
  -H "Content-Type: application/json" \
  -d '{"question": "你们几点开门？", "answer": "早上9点到晚上6点"}'
```

---

## 故障排查

### 数据库权限问题
如果遇到数据库写入错误，确保挂载的 volume 有正确权限：
```bash
chmod 755 ./data
```

### 内存不足
在 `Dockerfile` 中减少 Go 的并发数，或添加环境变量：
```dockerfile
ENV GOMAXPROCS=1
```
