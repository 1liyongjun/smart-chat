'use client'

import { useState, useEffect, useRef } from 'react'
import ChatWidget from '@/components/ChatWidget'

const API_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8000'

interface Message {
  role: 'user' | 'assistant'
  content: string
}

export default function WidgetPage() {
  const [messages, setMessages] = useState<Message[]>([])
  const [input, setInput] = useState('')
  const [loading, setLoading] = useState(false)
  const [visitorId] = useState(() => localStorage.getItem('widget_visitor_id') || (() => {
    const id = 'v_' + Math.random().toString(36).slice(2, 11)
    localStorage.setItem('widget_visitor_id', id)
    return id
  })())
  const [conversationId, setConversationId] = useState('')
  const messagesEndRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [messages])

  const sendMessage = async () => {
    if (!input.trim() || loading) return
    
    const userMessage: Message = { role: 'user', content: input }
    setMessages(prev => [...prev, userMessage])
    const currentInput = input
    setInput('')
    setLoading(true)

    try {
      const res = await fetch(`${API_URL}/api/chat`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ 
          message: currentInput, 
          visitor_id: visitorId,
          conversation_id: conversationId || undefined
        })
      })
      const data = await res.json()
      
      if (data.conversation_id && !conversationId) {
        setConversationId(data.conversation_id)
      }
      
      setMessages(prev => [...prev, { role: 'assistant', content: data.answer }])
    } catch (err) {
      setMessages(prev => [...prev, { 
        role: 'assistant', 
        content: '抱歉，服务暂时不可用，请稍后重试。' 
      }])
    } finally {
      setLoading(false)
    }
  }

  return (
    <div style={{ minHeight: '100vh', display: 'flex', flexDirection: 'column' }}>
      {/* 头部 */}
      <header style={{
        background: 'linear-gradient(135deg, #667EEA 0%, #764BA2 100%)',
        color: 'white',
        padding: '2rem'
      }}>
        <div style={{ maxWidth: '800px', margin: '0 auto', textAlign: 'center' }}>
          <h1 style={{ fontSize: '2rem', fontWeight: 700, marginBottom: '0.5rem' }}>🤖 智能客服中心</h1>
          <p style={{ opacity: 0.9 }}>基于 MiniMax AI 的知识库问答系统</p>
        </div>
      </header>

      {/* 聊天区域 */}
      <main style={{ flex: 1, maxWidth: '800px', margin: '0 auto', width: '100%', padding: '2rem', display: 'flex', flexDirection: 'column' }}>
        <div style={{
          flex: 1,
          border: '1px solid #E5E7EB',
          borderRadius: '1rem',
          overflow: 'hidden',
          display: 'flex',
          flexDirection: 'column',
          background: 'white'
        }}>
          {/* 消息列表 */}
          <div style={{ flex: 1, overflow: 'auto', padding: '1.5rem', minHeight: '400px' }}>
            {messages.length === 0 && (
              <div style={{ textAlign: 'center', color: '#9CA3AF', marginTop: '3rem' }}>
                <div style={{ fontSize: '3rem', marginBottom: '1rem' }}>💬</div>
                <div style={{ fontSize: '1.1rem', marginBottom: '0.5rem' }}>欢迎来到智能客服</div>
                <div style={{ fontSize: '0.9rem' }}>
                  我可以回答关于产品、服务的问题<br/>
                  输入您的问题开始对话
                </div>
              </div>
            )}
            
            {messages.map((msg, i) => (
              <div key={i} style={{
                display: 'flex',
                justifyContent: msg.role === 'user' ? 'flex-end' : 'flex-start',
                marginBottom: '1rem'
              }}>
                <div style={{
                  maxWidth: '75%',
                  padding: '1rem 1.25rem',
                  borderRadius: '1rem',
                  background: msg.role === 'user' ? 'linear-gradient(135deg, #667EEA 0%, #764BA2 100%)' : '#F3F4F6',
                  color: msg.role === 'user' ? 'white' : '#111827',
                  borderBottomRightRadius: msg.role === 'user' ? '0.25rem' : '1rem',
                  borderBottomLeftRadius: msg.role === 'user' ? '1rem' : '0.25rem',
                  fontSize: '0.95rem',
                  lineHeight: 1.6,
                  whiteSpace: 'pre-wrap',
                  wordBreak: 'break-word',
                  boxShadow: '0 1px 3px rgba(0,0,0,0.1)'
                }}>
                  {msg.content}
                </div>
              </div>
            ))}
            
            {loading && (
              <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', color: '#9CA3AF', padding: '0.5rem 0' }}>
                <div style={{
                  width: '8px', height: '8px', borderRadius: '50%', background: '#9CA3AF',
                  animation: 'typing 1s infinite'
                }} />
                <div style={{
                  width: '8px', height: '8px', borderRadius: '50%', background: '#9CA3AF',
                  animation: 'typing 1s infinite 0.2s'
                }} />
                <div style={{
                  width: '8px', height: '8px', borderRadius: '50%', background: '#9CA3AF',
                  animation: 'typing 1s infinite 0.4s'
                }} />
                <span style={{ marginLeft: '0.5rem' }}>AI 思考中...</span>
              </div>
            )}
            <div ref={messagesEndRef} />
          </div>

          {/* 输入框 */}
          <div style={{ padding: '1rem 1.5rem', borderTop: '1px solid #E5E7EB', background: '#FAFAFA' }}>
            <div style={{ display: 'flex', gap: '0.75rem' }}>
              <input
                type="text"
                value={input}
                onChange={e => setInput(e.target.value)}
                onKeyDown={e => e.key === 'Enter' && !e.shiftKey && (e.preventDefault(), sendMessage())}
                placeholder="输入您的问题，按 Enter 发送..."
                style={{
                  flex: 1,
                  padding: '0.875rem 1.25rem',
                  border: '1px solid #E5E7EB',
                  borderRadius: '9999px',
                  fontSize: '0.95rem',
                  outline: 'none',
                  transition: 'border-color 0.2s'
                }}
                onFocus={e => e.target.style.borderColor = '#667EEA'}
                onBlur={e => e.target.style.borderColor = '#E5E7EB'}
              />
              <button
                onClick={sendMessage}
                disabled={!input.trim() || loading}
                style={{
                  padding: '0.875rem 1.5rem',
                  borderRadius: '9999px',
                  background: input.trim() && !loading ? 'linear-gradient(135deg, #667EEA 0%, #764BA2 100%)' : '#D1D5DB',
                  color: 'white',
                  border: 'none',
                  cursor: input.trim() && !loading ? 'pointer' : 'not-allowed',
                  fontSize: '0.95rem',
                  fontWeight: 500,
                  transition: 'all 0.2s'
                }}
              >
                发送
              </button>
            </div>
          </div>
        </div>

        {/* 底部信息 */}
        <div style={{ textAlign: 'center', marginTop: '1rem', color: '#9CA3AF', fontSize: '0.8rem' }}>
          Powered by MiniMax AI · 智能客服系统
        </div>
      </main>

      <style jsx global>{`
        @keyframes typing {
          0%, 100% { transform: translateY(0); opacity: 0.5; }
          50% { transform: translateY(-4px); opacity: 1; }
        }
      `}</style>
    </div>
  )
}
