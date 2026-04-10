# 智能客服系统

基于 MiniMax AI 的知识库问答客服系统，支持网页嵌入聊天气泡和管理后台。

## 技术架构

| 层级 | 技术栈 |
|------|--------|
| **前端** | Next.js 14 + TypeScript |
| **后端** | Go + Fiber v2 + Eino 框架 |
| **AI** | MiniMax-M2.7 (字节跳动 Eino 框架) |
| **数据库** | SQLite |

## 功能特性

- 🤖 **AI 智能问答** - 基于 MiniMax-M2.7 模型 + Eino 框架编排
- 📚 **知识库管理** - 添加、编辑、删除问答知识
- 💬 **访客聊天页面** - 独立的访客聊天界面
- 🔗 **嵌入组件** - 可嵌入到任意网站的聊天气泡组件
- 📊 **数据统计** - 查看对话量、访客数等统计数据
- 🌊 **流式响应** - 支持 SSE 流式输出

## 快速开始

### 后端 (Go + Eino)

```bash
cd backend

# 安装依赖
go mod tidy

# 运行
go run main.go -api-key="your-minimax-api-key"
```

后端运行在 http://localhost:8000

### 前端

```bash
cd frontend
npm install
npm run dev
```

前端运行在 http://localhost:3000

- 访客页面: http://localhost:3000/widget
- 管理后台: http://localhost:3000/admin

## 目录结构

```
smart-chat/
├── backend/
│   ├── main.go
│   ├── go.mod
│   ├── README.md
│   └── internal/
│       ├── eino/minimax.go
│       ├── handler/handler.go
│       └── store/store.go
└── frontend/
    ├── package.json
    ├── next.config.js
    ├── tsconfig.json
    └── src/
        ├── app/
        │   ├── admin/page.tsx
        │   ├── widget/page.tsx
        │   ├── layout.tsx
        │   ├── page.tsx
        │   └── globals.css
        └── components/
            └── ChatWidget.tsx
```

## 关于 Eino 框架

[Eino](https://github.com/cloudwego/eino) 是字节跳动开源的 Go 语言 LLM 应用开发框架：

- **组件化**: ChatModel、Tool、Retriever 等标准化组件
- **强类型**: 编译时校验，运行时更稳定
- **高性能**: 基于 Go 语言，高并发处理
- **编排能力**: 支持 Chain、Graph 等复杂工作流编排

本项目使用 Eino 的 ChatModel 接口封装了 MiniMax API，实现了:
- `Generate()` - 普通对话生成
- `Stream()` - 流式对话生成 (SSE)

## 部署

### 后端部署
- Docker / Kubernetes
- Railway / Render (需要构建)

### 前端部署
- Vercel (推荐)
- Netlify
- Cloudflare Pages

## 嵌入网站

在您的网站 HTML 底部添加：

```html
<script src="https://your-domain.com/widget.js"></script>
```
