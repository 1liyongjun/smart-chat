# 智能客服系统

基于 MiniMax AI 的知识库问答客服系统，支持网页嵌入聊天气泡和管理后台。

## 功能特性

- 🤖 **AI 智能问答** - 基于 MiniMax-M2.7 模型
- 📚 **知识库管理** - 添加、编辑、删除问答知识
- 💬 **访客聊天页面** - 独立的访客聊天界面
- 🔗 **嵌入组件** - 可嵌入到任意网站的聊天气泡组件
- 📊 **数据统计** - 查看对话量、访客数等统计数据

## 快速开始

### 后端

```bash
cd backend
pip install -r requirements.txt
export MINIMAX_API_KEY="your-api-key"
python main.py
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

## 部署

### 后端部署
推荐使用 Railway、Render、Vercel Serverless Functions 或 Docker

### 前端部署
推荐使用 Vercel、Netlify 或 Cloudflare Pages

## 嵌入网站

在您的网站 HTML 底部添加：

```html
<script src="https://your-domain.com/widget.js"></script>
```

## 技术栈

- **后端**: Python FastAPI + SQLite
- **前端**: Next.js 14 + TypeScript
- **AI**: MiniMax-M2.7
