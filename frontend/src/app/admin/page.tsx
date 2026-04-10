'use client'

import { useState, useEffect } from 'react'
import Link from 'next/link'

const API_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8000'

interface Stats {
  knowledge_count: number
  conversation_count: number
  visitor_count: number
}

interface KnowledgeItem {
  id: number
  question: string
  answer: string
  category: string
  tags: string
  usage_count: number
  created_at: string
}

interface Conversation {
  id: string
  visitor_id: string
  messages: string
  created_at: string
  updated_at: string
}

export default function AdminDashboard() {
  const [stats, setStats] = useState<Stats>({ knowledge_count: 0, conversation_count: 0, visitor_count: 0 })
  const [knowledgeList, setKnowledgeList] = useState<KnowledgeItem[]>([])
  const [conversations, setConversations] = useState<Conversation[]>([])
  const [activeTab, setActiveTab] = useState<'dashboard' | 'knowledge' | 'conversations'>('dashboard')
  const [editingItem, setEditingItem] = useState<KnowledgeItem | null>(null)
  const [newQA, setNewQA] = useState({ question: '', answer: '', category: 'general', tags: '' })
  const [loading, setLoading] = useState(false)

  const fetchStats = async () => {
    try {
      const res = await fetch(`${API_URL}/api/stats`)
      const data = await res.json()
      setStats(data)
    } catch (e) { console.error('Failed to fetch stats:', e) }
  }

  const fetchKnowledge = async () => {
    try {
      const res = await fetch(`${API_URL}/api/knowledge?page_size=50`)
      const data = await res.json()
      setKnowledgeList(data.items || [])
    } catch (e) { console.error('Failed to fetch knowledge:', e) }
  }

  const fetchConversations = async () => {
    try {
      const res = await fetch(`${API_URL}/api/conversations?page_size=50`)
      const data = await res.json()
      setConversations(data.items || [])
    } catch (e) { console.error('Failed to fetch conversations:', e) }
  }

  useEffect(() => {
    fetchStats()
    fetchKnowledge()
    fetchConversations()
  }, [])

  const handleAddKnowledge = async () => {
    if (!newQA.question || !newQA.answer) return
    setLoading(true)
    try {
      await fetch(`${API_URL}/api/knowledge`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          question: newQA.question,
          answer: newQA.answer,
          category: newQA.category,
          tags: newQA.tags.split(',').map(t => t.trim()).filter(Boolean)
        })
      })
      setNewQA({ question: '', answer: '', category: 'general', tags: '' })
      fetchKnowledge()
      fetchStats()
    } catch (e) { console.error('Failed to add:', e) }
    setLoading(false)
  }

  const handleUpdateKnowledge = async () => {
    if (!editingItem) return
    setLoading(true)
    try {
      await fetch(`${API_URL}/api/knowledge/${editingItem.id}`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          question: editingItem.question,
          answer: editingItem.answer,
          category: editingItem.category,
          tags: editingItem.tags ? JSON.parse(editingItem.tags) : []
        })
      })
      setEditingItem(null)
      fetchKnowledge()
    } catch (e) { console.error('Failed to update:', e) }
    setLoading(false)
  }

  const handleDeleteKnowledge = async (id: number) => {
    if (!confirm('确定删除？')) return
    try {
      await fetch(`${API_URL}/api/knowledge/${id}`, { method: 'DELETE' })
      fetchKnowledge()
      fetchStats()
    } catch (e) { console.error('Failed to delete:', e) }
  }

  const parseMessages = (msgStr: string) => {
    try {
      const msgs = JSON.parse(msgStr)
      return Array.isArray(msgs) ? msgs : []
    } catch {
      return []
    }
  }

  return (
    <div style={{ minHeight: '100vh', background: '#F9FAFB' }}>
      {/* 顶部导航 */}
      <header style={{ background: 'white', borderBottom: '1px solid #E5E7EB', padding: '0 2rem' }}>
        <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', maxWidth: '1200px', margin: '0 auto', height: '4rem' }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: '2rem' }}>
            <h1 style={{ fontSize: '1.25rem', fontWeight: 700, color: '#4F46E5' }}>🤖 智能客服管理后台</h1>
            <nav style={{ display: 'flex', gap: '1.5rem' }}>
              {(['dashboard', 'knowledge', 'conversations'] as const).map(tab => (
                <button
                  key={tab}
                  onClick={() => setActiveTab(tab)}
                  style={{
                    background: 'none',
                    border: 'none',
                    padding: '0.5rem 0',
                    cursor: 'pointer',
                    fontSize: '0.9rem',
                    fontWeight: activeTab === tab ? 600 : 400,
                    color: activeTab === tab ? '#4F46E5' : '#6B7280',
                    borderBottom: activeTab === tab ? '2px solid #4F46E5' : '2px solid transparent'
                  }}
                >
                  {tab === 'dashboard' ? '📊 总览' : tab === 'knowledge' ? '📚 知识库' : '💬 对话记录'}
                </button>
              ))}
            </nav>
          </div>
          <Link href="/widget" style={{ color: '#6B7280', fontSize: '0.875rem' }}>访客页面 →</Link>
        </div>
      </header>

      {/* 内容区 */}
      <main style={{ maxWidth: '1200px', margin: '2rem auto', padding: '0 2rem' }}>
        
        {/* 总览 */}
        {activeTab === 'dashboard' && (
          <div>
            <h2 style={{ fontSize: '1.25rem', fontWeight: 600, marginBottom: '1.5rem' }}>数据总览</h2>
            <div style={{ display: 'grid', gridTemplateColumns: 'repeat(3, 1fr)', gap: '1.5rem' }}>
              <div className="card">
                <div style={{ color: '#6B7280', fontSize: '0.875rem' }}>知识库条目</div>
                <div style={{ fontSize: '2.5rem', fontWeight: 700, color: '#4F46E5', marginTop: '0.5rem' }}>{stats.knowledge_count}</div>
              </div>
              <div className="card">
                <div style={{ color: '#6B7280', fontSize: '0.875rem' }}>对话总数</div>
                <div style={{ fontSize: '2.5rem', fontWeight: 700, color: '#10B981', marginTop: '0.5rem' }}>{stats.conversation_count}</div>
              </div>
              <div className="card">
                <div style={{ color: '#6B7280', fontSize: '0.875rem' }}>独立访客</div>
                <div style={{ fontSize: '2.5rem', fontWeight: 700, color: '#F59E0B', marginTop: '0.5rem' }}>{stats.visitor_count}</div>
              </div>
            </div>

            <div className="card" style={{ marginTop: '1.5rem' }}>
              <h3 style={{ fontSize: '1rem', fontWeight: 600, marginBottom: '1rem' }}>快速开始</h3>
              <div style={{ display: 'grid', gridTemplateColumns: 'repeat(2, 1fr)', gap: '1rem' }}>
                <div style={{ padding: '1rem', background: '#F9FAFB', borderRadius: '0.5rem' }}>
                  <div style={{ fontWeight: 600 }}>1. 嵌入聊天气泡</div>
                  <div style={{ fontSize: '0.8rem', color: '#6B7280', marginTop: '0.25rem' }}>在网站 HTML 底部加入代码</div>
                  <code style={{ display: 'block', marginTop: '0.5rem', padding: '0.5rem', background: '#1F2937', color: '#10B981', borderRadius: '0.25rem', fontSize: '0.75rem' }}>
                    {`<script src="https://your-domain.com/widget.js"><\/script>`}
                  </code>
                </div>
                <div style={{ padding: '1rem', background: '#F9FAFB', borderRadius: '0.5rem' }}>
                  <div style={{ fontWeight: 600 }}>2. 添加知识库</div>
                  <div style={{ fontSize: '0.8rem', color: '#6B7280', marginTop: '0.25rem' }}>在知识库页面添加常见问答</div>
                </div>
              </div>
            </div>
          </div>
        )}

        {/* 知识库管理 */}
        {activeTab === 'knowledge' && (
          <div>
            <h2 style={{ fontSize: '1.25rem', fontWeight: 600, marginBottom: '1.5rem' }}>知识库管理</h2>
            
            {/* 添加表单 */}
            <div className="card" style={{ marginBottom: '1.5rem' }}>
              <h3 style={{ fontSize: '1rem', fontWeight: 600, marginBottom: '1rem' }}>添加新条目</h3>
              <div style={{ display: 'grid', gap: '1rem' }}>
                <div>
                  <label style={{ display: 'block', fontSize: '0.875rem', color: '#6B7280', marginBottom: '0.25rem' }}>问题</label>
                  <input
                    className="input"
                    placeholder="用户可能会问的问题"
                    value={newQA.question}
                    onChange={e => setNewQA({ ...newQA, question: e.target.value })}
                  />
                </div>
                <div>
                  <label style={{ display: 'block', fontSize: '0.875rem', color: '#6B7280', marginBottom: '0.25rem' }}>回答</label>
                  <textarea
                    className="input"
                    placeholder="AI 应该如何回答"
                    rows={3}
                    value={newQA.answer}
                    onChange={e => setNewQA({ ...newQA, answer: e.target.value })}
                  />
                </div>
                <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '1rem' }}>
                  <div>
                    <label style={{ display: 'block', fontSize: '0.875rem', color: '#6B7280', marginBottom: '0.25rem' }}>分类</label>
                    <select
                      className="input"
                      value={newQA.category}
                      onChange={e => setNewQA({ ...newQA, category: e.target.value })}
                    >
                      <option value="general">通用</option>
                      <option value="product">产品</option>
                      <option value="service">服务</option>
                      <option value="technical">技术</option>
                      <option value="other">其他</option>
                    </select>
                  </div>
                  <div>
                    <label style={{ display: 'block', fontSize: '0.875rem', color: '#6B7280', marginBottom: '0.25rem' }}>标签（逗号分隔）</label>
                    <input
                      className="input"
                      placeholder="标签1, 标签2"
                      value={newQA.tags}
                      onChange={e => setNewQA({ ...newQA, tags: e.target.value })}
                    />
                  </div>
                </div>
                <button className="btn btn-primary" onClick={handleAddKnowledge} disabled={loading}>
                  {loading ? '添加中...' : '➕ 添加条目'}
                </button>
              </div>
            </div>

            {/* 列表 */}
            <div style={{ display: 'grid', gap: '1rem' }}>
              {knowledgeList.map(item => (
                <div key={item.id} className="card" style={{ padding: '1rem' }}>
                  <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
                    <div style={{ flex: 1 }}>
                      <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', marginBottom: '0.5rem' }}>
                        <span className="badge badge-primary">{item.category}</span>
                        <span style={{ fontSize: '0.75rem', color: '#9CA3AF' }}>使用 {item.usage_count} 次</span>
                      </div>
                      <div style={{ fontWeight: 600, marginBottom: '0.25rem' }}>Q: {item.question}</div>
                      <div style={{ color: '#4B5563', fontSize: '0.9rem' }}>A: {item.answer}</div>
                    </div>
                    <div style={{ display: 'flex', gap: '0.5rem' }}>
                      <button className="btn btn-secondary" onClick={() => setEditingItem(item)}>编辑</button>
                      <button className="btn btn-secondary" onClick={() => handleDeleteKnowledge(item.id)} style={{ color: '#EF4444' }}>删除</button>
                    </div>
                  </div>
                </div>
              ))}
              {knowledgeList.length === 0 && (
                <div style={{ textAlign: 'center', color: '#9CA3AF', padding: '3rem' }}>
                  暂无知识库条目，请添加
                </div>
              )}
            </div>
          </div>
        )}

        {/* 对话记录 */}
        {activeTab === 'conversations' && (
          <div>
            <h2 style={{ fontSize: '1.25rem', fontWeight: 600, marginBottom: '1.5rem' }}>对话记录</h2>
            <div style={{ display: 'grid', gap: '1rem' }}>
              {conversations.map(conv => {
                const messages = parseMessages(conv.messages)
                return (
                  <div key={conv.id} className="card" style={{ padding: '1rem' }}>
                    <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '0.75rem' }}>
                      <div>
                        <span style={{ fontWeight: 600 }}>访客: {conv.visitor_id.slice(0, 12)}...</span>
                        <span style={{ marginLeft: '1rem', fontSize: '0.75rem', color: '#9CA3AF' }}>
                          {new Date(conv.created_at).toLocaleString('zh-CN')}
                        </span>
                      </div>
                      <span className="badge badge-success">{messages.length} 条消息</span>
                    </div>
                    <div style={{ display: 'grid', gap: '0.5rem' }}>
                      {messages.slice(-4).map((msg: any, i: number) => (
                        <div key={i} style={{ fontSize: '0.85rem', padding: '0.5rem', background: '#F9FAFB', borderRadius: '0.25rem' }}>
                          <span style={{ fontWeight: 600, color: msg.role === 'user' ? '#4F46E5' : '#10B981' }}>
                            {msg.role === 'user' ? '👤' : '🤖'}:
                          </span>
                          <span style={{ marginLeft: '0.5rem' }}>{msg.content.slice(0, 100)}{msg.content.length > 100 ? '...' : ''}</span>
                        </div>
                      ))}
                    </div>
                  </div>
                )
              })}
              {conversations.length === 0 && (
                <div style={{ textAlign: 'center', color: '#9CA3AF', padding: '3rem' }}>
                  暂无对话记录
                </div>
              )}
            </div>
          </div>
        )}
      </main>

      {/* 编辑弹窗 */}
      {editingItem && (
        <div style={{
          position: 'fixed', inset: 0, background: 'rgba(0,0,0,0.5)', display: 'flex',
          alignItems: 'center', justifyContent: 'center', zIndex: 1000
        }}>
          <div className="card" style={{ width: '500px', maxHeight: '80vh', overflow: 'auto' }}>
            <h3 style={{ fontSize: '1.1rem', fontWeight: 600, marginBottom: '1rem' }}>编辑知识库条目</h3>
            <div style={{ display: 'grid', gap: '1rem' }}>
              <div>
                <label style={{ display: 'block', fontSize: '0.875rem', marginBottom: '0.25rem' }}>问题</label>
                <input
                  className="input"
                  value={editingItem.question}
                  onChange={e => setEditingItem({ ...editingItem, question: e.target.value })}
                />
              </div>
              <div>
                <label style={{ display: 'block', fontSize: '0.875rem', marginBottom: '0.25rem' }}>回答</label>
                <textarea
                  className="input"
                  rows={3}
                  value={editingItem.answer}
                  onChange={e => setEditingItem({ ...editingItem, answer: e.target.value })}
                />
              </div>
              <div>
                <label style={{ display: 'block', fontSize: '0.875rem', marginBottom: '0.25rem' }}>分类</label>
                <select
                  className="input"
                  value={editingItem.category}
                  onChange={e => setEditingItem({ ...editingItem, category: e.target.value })}
                >
                  <option value="general">通用</option>
                  <option value="product">产品</option>
                  <option value="service">服务</option>
                  <option value="technical">技术</option>
                </select>
              </div>
              <div style={{ display: 'flex', gap: '0.5rem', justifyContent: 'flex-end' }}>
                <button className="btn btn-secondary" onClick={() => setEditingItem(null)}>取消</button>
                <button className="btn btn-primary" onClick={handleUpdateKnowledge} disabled={loading}>
                  {loading ? '保存中...' : '保存'}
                </button>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
