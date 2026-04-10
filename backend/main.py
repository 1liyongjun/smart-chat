"""
智能客服后端 - FastAPI + SQLite + MiniMax AI
"""
import os
import json
import sqlite3
import uuid
import time
from datetime import datetime, timedelta
from typing import Optional, List
from contextlib import contextmanager

from fastapi import FastAPI, HTTPException, Request
from fastapi.middleware.cors import CORSMiddleware
from fastapi.responses import JSONResponse
from pydantic import BaseModel
import requests

# ============== 配置 ==============
DATABASE_PATH = os.path.join(os.path.dirname(__file__), "data", "smart_chat.db")
MINIMAX_API_KEY = os.environ.get("MINIMAX_API_KEY", "")
MINIMAX_BASE_URL = "https://api.minimax.io/v1"
MODEL_NAME = "MiniMax-M2.7"

# ============== 数据库初始化 ==============
def init_db():
    """初始化数据库"""
    os.makedirs(os.path.dirname(DATABASE_PATH), exist_ok=True)
    conn = sqlite3.connect(DATABASE_PATH)
    c = conn.cursor()
    
    # 知识库表
    c.execute("""
        CREATE TABLE IF NOT EXISTS knowledge_base (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            question TEXT NOT NULL,
            answer TEXT NOT NULL,
            category TEXT DEFAULT 'general',
            tags TEXT DEFAULT '[]',
            usage_count INTEGER DEFAULT 0,
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
            updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
        )
    """)
    
    # 对话记录表
    c.execute("""
        CREATE TABLE IF NOT EXISTS conversations (
            id TEXT PRIMARY KEY,
            visitor_id TEXT NOT NULL,
            messages TEXT DEFAULT '[]',
            status TEXT DEFAULT 'active',
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
            updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
        )
    """)
    
    # 访客表
    c.execute("""
        CREATE TABLE IF NOT EXISTS visitors (
            id TEXT PRIMARY KEY,
            name TEXT,
            email TEXT,
            metadata TEXT DEFAULT '{}',
            first_visit TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
            last_visit TIMESTAMP DEFAULT CURRENT_TIMESTAMP
        )
    """)
    
    conn.commit()
    conn.close()
    print(f"✅ 数据库初始化完成: {DATABASE_PATH}")

@contextmanager
def get_db():
    """数据库连接上下文管理器"""
    conn = sqlite3.connect(DATABASE_PATH)
    conn.row_factory = sqlite3.Row
    try:
        yield conn
    finally:
        conn.close()

# ============== Pydantic 模型 ==============
class QAItem(BaseModel):
    question: str
    answer: str
    category: str = "general"
    tags: List[str] = []

class ChatRequest(BaseModel):
    message: str
    visitor_id: Optional[str] = None
    conversation_id: Optional[str] = None

class ChatResponse(BaseModel):
    answer: str
    conversation_id: str
    sources: List[str] = []

# ============== MiniMax AI 调用 ==============
def call_minimax(prompt: str, context: str = "") -> str:
    """调用 MiniMax AI"""
    if not MINIMAX_API_KEY:
        return "⚠️ AI服务未配置，请联系管理员设置 MINIMAX_API_KEY 环境变量"
    
    system_prompt = f"""你是一个专业的智能客服助手。请根据提供的知识库内容回答用户问题。

知识库内容：
{context}

回答要求：
1. 如果知识库中有相关内容，优先使用知识库内容回答
2. 如果知识库中没有相关信息，礼貌地说明暂时无法回答，并引导用户联系人工客服
3. 回答要简洁、专业、友好
4. 不要编造知识库中没有的信息
"""
    
    try:
        response = requests.post(
            f"{MINIMAX_BASE_URL}/chat/completions",
            headers={
                "Authorization": f"Bearer {MINIMAX_API_KEY}",
                "Content-Type": "application/json"
            },
            json={
                "model": MODEL_NAME,
                "messages": [
                    {"role": "system", "content": system_prompt},
                    {"role": "user", "content": prompt}
                ],
                "temperature": 0.7,
                "max_tokens": 1024
            },
            timeout=30
        )
        result = response.json()
        if "choices" in result and len(result["choices"]) > 0:
            return result["choices"][0]["message"]["content"]
        return "❌ AI响应格式错误"
    except Exception as e:
        return f"❌ AI服务调用失败: {str(e)}"

def search_knowledge_base(query: str, top_k: int = 3) -> List[dict]:
    """搜索知识库"""
    with get_db() as conn:
        c = conn.cursor()
        # 简单的关键词匹配搜索
        keywords = query.split()
        if keywords:
            pattern = "%" + "%".join(keywords) + "%"
            c.execute("""
                SELECT * FROM knowledge_base 
                WHERE question LIKE ? OR answer LIKE ?
                ORDER BY usage_count DESC
                LIMIT ?
            """, (pattern, pattern, top_k))
        else:
            c.execute("SELECT * FROM knowledge_base ORDER BY usage_count DESC LIMIT ?", (top_k,))
        
        rows = c.fetchall()
        return [dict(row) for row in rows]

# ============== FastAPI 应用 ==============
app = FastAPI(title="智能客服 API", version="1.0.0")

# CORS 配置
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

@app.on_event("startup")
async def startup():
    init_db()
    print("🚀 智能客服服务已启动")

# ============== API 路由 ==============

# 健康检查
@app.get("/health")
async def health():
    return {"status": "ok", "service": "smart-chat"}

# --- 知识库管理 ---

@app.post("/api/knowledge")
async def add_knowledge(qa: QAItem):
    """添加知识库条目"""
    with get_db() as conn:
        c = conn.cursor()
        c.execute("""
            INSERT INTO knowledge_base (question, answer, category, tags)
            VALUES (?, ?, ?, ?)
        """, (qa.question, qa.answer, qa.category, json.dumps(qa.tags)))
        conn.commit()
        return {"id": c.lastrowid, "message": "添加成功"}

@app.get("/api/knowledge")
async def list_knowledge(category: Optional[str] = None, page: int = 1, page_size: int = 20):
    """获取知识库列表"""
    with get_db() as conn:
        c = conn.cursor()
        if category:
            c.execute("""
                SELECT * FROM knowledge_base 
                WHERE category = ?
                ORDER BY updated_at DESC
                LIMIT ? OFFSET ?
            """, (category, page_size, (page-1)*page_size))
        else:
            c.execute("""
                SELECT * FROM knowledge_base 
                ORDER BY updated_at DESC
                LIMIT ? OFFSET ?
            """, (page_size, (page-1)*page_size))
        
        items = [dict(row) for row in c.fetchall()]
        
        c.execute("SELECT COUNT(*) FROM knowledge_base")
        total = c.fetchone()[0]
        
        return {"items": items, "total": total, "page": page, "page_size": page_size}

@app.put("/api/knowledge/{item_id}")
async def update_knowledge(item_id: int, qa: QAItem):
    """更新知识库条目"""
    with get_db() as conn:
        c = conn.cursor()
        c.execute("""
            UPDATE knowledge_base 
            SET question = ?, answer = ?, category = ?, tags = ?, updated_at = CURRENT_TIMESTAMP
            WHERE id = ?
        """, (qa.question, qa.answer, qa.category, json.dumps(qa.tags), item_id))
        conn.commit()
        if c.rowcount == 0:
            raise HTTPException(status_code=404, message="条目不存在")
        return {"message": "更新成功"}

@app.delete("/api/knowledge/{item_id}")
async def delete_knowledge(item_id: int):
    """删除知识库条目"""
    with get_db() as conn:
        c = conn.cursor()
        c.execute("DELETE FROM knowledge_base WHERE id = ?", (item_id,))
        conn.commit()
        if c.rowcount == 0:
            raise HTTPException(status_code=404, message="条目不存在")
        return {"message": "删除成功"}

# --- 聊天接口 ---

@app.post("/api/chat", response_model=ChatResponse)
async def chat(req: ChatRequest):
    """处理用户消息"""
    visitor_id = req.visitor_id or str(uuid.uuid4())
    conversation_id = req.conversation_id or str(uuid.uuid4())
    
    # 搜索知识库
    kb_results = search_knowledge_base(req.message)
    
    # 构建上下文
    context = ""
    sources = []
    if kb_results:
        for i, item in enumerate(kb_results, 1):
            context += f"\n【相关问答 {i}】\n问题: {item['question']}\n回答: {item['answer']}\n"
            sources.append(f"Q: {item['question']}")
        
        # 更新使用次数
        with get_db() as conn:
            c = conn.cursor()
            for item in kb_results:
                c.execute("UPDATE knowledge_base SET usage_count = usage_count + 1 WHERE id = ?", (item['id'],))
            conn.commit()
    
    # 调用 AI
    answer = call_minimax(req.message, context)
    
    # 保存对话记录
    with get_db() as conn:
        c = conn.cursor()
        
        # 获取现有消息
        c.execute("SELECT messages FROM conversations WHERE id = ?", (conversation_id,))
        row = c.fetchone()
        messages = json.loads(row[0]) if row and row[0] else [] if row else []
        
        # 添加新消息
        messages.append({
            "role": "user",
            "content": req.message,
            "timestamp": datetime.now().isoformat()
        })
        messages.append({
            "role": "assistant", 
            "content": answer,
            "timestamp": datetime.now().isoformat()
        })
        
        # 更新或创建对话
        c.execute("""
            INSERT OR REPLACE INTO conversations (id, visitor_id, messages, updated_at)
            VALUES (?, ?, ?, CURRENT_TIMESTAMP)
        """, (conversation_id, visitor_id, json.dumps(messages)))
        conn.commit()
    
    return ChatResponse(
        answer=answer,
        conversation_id=conversation_id,
        sources=sources
    )

@app.get("/api/conversations/{conversation_id}")
async def get_conversation(conversation_id: str):
    """获取对话历史"""
    with get_db() as conn:
        c = conn.cursor()
        c.execute("SELECT * FROM conversations WHERE id = ?", (conversation_id,))
        row = c.fetchone()
        if not row:
            raise HTTPException(status_code=404, message="对话不存在")
        return dict(row)

@app.get("/api/conversations")
async def list_conversations(page: int = 1, page_size: int = 50):
    """获取对话列表"""
    with get_db() as conn:
        c = conn.cursor()
        c.execute("""
            SELECT id, visitor_id, status, created_at, updated_at,
                   json_array_length(messages) as message_count
            FROM conversations
            ORDER BY updated_at DESC
            LIMIT ? OFFSET ?
        """, (page_size, (page-1)*page_size))
        items = [dict(row) for row in c.fetchall()]
        
        c.execute("SELECT COUNT(*) FROM conversations")
        total = c.fetchone()[0]
        
        return {"items": items, "total": total}

# --- 统计接口 ---

@app.get("/api/stats")
async def get_stats():
    """获取统计数据"""
    with get_db() as conn:
        c = conn.cursor()
        
        c.execute("SELECT COUNT(*) FROM knowledge_base")
        kb_count = c.fetchone()[0]
        
        c.execute("SELECT COUNT(*) FROM conversations")
        conv_count = c.fetchone()[0]
        
        c.execute("SELECT COUNT(DISTINCT visitor_id) FROM conversations")
        visitor_count = c.fetchone()[0]
        
        return {
            "knowledge_count": kb_count,
            "conversation_count": conv_count,
            "visitor_count": visitor_count
        }

# ============== 启动 ==============
if __name__ == "__main__":
    import uvicorn
    port = int(os.environ.get("PORT", 8000))
    uvicorn.run(app, host="0.0.0.0", port=port)
