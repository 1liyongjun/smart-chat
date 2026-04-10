'use client'

import { useState, useRef, useEffect } from 'react'

interface Message {
  role: 'user' | 'assistant'
  content: string
  timestamp?: string
}

interface ChatWidgetProps {
  apiUrl?: string
  title?: string
  placeholder?: string
}

export default function ChatWidget({
  apiUrl = '',
  title = '智能客服',
  placeholder = '输入您的问题...'
}: ChatWidgetProps) {
  const [isOpen, setIsOpen] = useState(false)
  const [messages, setMessages] = useState<Message[]>([])
  const [input, setInput] = useState('')
  const [loading, setLoading] = useState(false)
  const [visitorId] = useState(() => localStorage.getItem('visitor_id') || (() => {
    const id = 'visitor_' + Math.random().toString(36).slice(2, 11)
    localStorage.setItem('visitor_id', id)
    return id
  })())
  const [conversationId, setConversationId] = useState('')
  const messagesEndRef = useRef<HTMLDivElement>(null)

  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' })
  }

  useEffect(() => {
    scrollToBottom()
  }, [messages])

  const sendMessage = async () => {
    if (!input.trim() || loading) return
    
    const userMessage: Message = { role: 'user', content: input }
    setMessages(prev => [...prev, userMessage])
    setInput('')
    setLoading(true)

    try {
      const res = await fetch(`${apiUrl}/api/chat`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ 
          message: input, 
          visitor_id: visitorId,
          conversation_id: conversationId || undefined
        })
      })
      const data = await res.json()
      
      if (data.conversation_id && !conversationId) {
        setConversationId(data.conversation_id)
      }
      
      setMessages(prev => [...prev, { 
        role: 'assistant', 
        content: data.answer 
      }])
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
    <>
      {/* 聊天按钮 */}
      <button
        onClick={() => setIsOpen(!isOpen)}
        style={{
          position: 'fixed',
          bottom: '1.5rem',
          right: '1.5rem',
          width: '3.5rem',
          height: '3.5rem',
          borderRadius: '50%',
          background: 'var(--primary)',
          color: 'white',
          border: 'none',
          cursor: 'pointer',
          boxShadow: '0 4px 12px rgba(79, 70, 229, 0.4)',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          fontSize: '1.5rem',
          zIndex: 9999,
          transition: 'transform 0.2s',
          transform: isOpen ? 'rotate(45deg)' : 'none'
        }}
      >
        {isOpen ? '✕' : '💬'}
      </button>

      {/* 聊天窗口 */}
      {isOpen && (
        <div style={{
          position: 'fixed',
          bottom: '5.5rem',
          right: '1.5rem',
          width: '380px',
          height: '520px',
          background: 'white',
          borderRadius: '1rem',
          boxShadow: '0 8px 30px rgba(0, 0, 0, 0.12)',
          display: 'flex',
          flexDirection: 'column',
          zIndex: 9999,
          overflow: 'hidden'
        }}>
          {/* 头部 */}
          <div style={{
            padding: '1rem 1.25rem',
            background: 'var(--primary)',
            color: 'white',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'space-between'
          }}>
            <div>
              <div style={{ fontWeight: 600 }}>{title}</div>
              <div style={{ fontSize: '0.75rem', opacity: 0.85 }}>基于 MiniMax AI</div>
            </div>
            <button 
              onClick={() => setIsOpen(false)}
              style={{ background: 'none', border: 'none', color: 'white', cursor: 'pointer', fontSize: '1.25rem' }}
            >
              ✕
            </button>
          </div>

          {/* 消息区域 */}
          <div style={{ flex: 1, overflow: 'auto', padding: '1rem' }}>
            {messages.length === 0 && (
              <div style={{ textAlign: 'center', color: '#9CA3AF', marginTop: '2rem' }}>
                <div style={{ fontSize: '2rem', marginBottom: '0.5rem' }}>🤖</div>
                <div>您好！我是智能客服助手</div>
                <div style={{ fontSize: '0.8rem', marginTop: '0.5rem' }}>
                  可以问我关于产品、服务的问题
                </div>
              </div>
            )}
            
            {messages.map((msg, i) => (
              <div key={i} style={{
                display: 'flex',
                justifyContent: msg.role === 'user' ? 'flex-end' : 'flex-start',
                marginBottom: '0.75rem'
              }}>
                <div style={{
                  maxWidth: '80%',
                  padding: '0.75rem 1rem',
                  borderRadius: '1rem',
                  background: msg.role === 'user' ? 'var(--primary)' : '#F3F4F6',
                  color: msg.role === 'user' ? 'white' : '#111827',
                  borderBottomRightRadius: msg.role === 'user' ? '0.25rem' : '1rem',
                  borderBottomLeftRadius: msg.role === 'user' ? '1rem' : '0.25rem',
                  fontSize: '0.9rem',
                  whiteSpace: 'pre-wrap',
                  wordBreak: 'break-word'
                }}>
                  {msg.content}
                </div>
              </div>
            ))}
            
            {loading && (
              <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', color: '#9CA3AF' }}>
                <div className="typing-dot" />
                <span>AI 思考中...</span>
              </div>
            )}
            <div ref={messagesEndRef} />
          </div>

          {/* 输入区域 */}
          <div style={{ padding: '1rem', borderTop: '1px solid #E5E7EB' }}>
            <div style={{ display: 'flex', gap: '0.5rem' }}>
              <input
                type="text"
                value={input}
                onChange={e => setInput(e.target.value)}
                onKeyDown={e => e.key === 'Enter' && !e.shiftKey && (e.preventDefault(), sendMessage())}
                placeholder={placeholder}
                style={{
                  flex: 1,
                  padding: '0.625rem 0.875rem',
                  border: '1px solid #E5E7EB',
                  borderRadius: '9999px',
                  fontSize: '0.875rem',
                  outline: 'none'
                }}
              />
              <button
                onClick={sendMessage}
                disabled={!input.trim() || loading}
                style={{
                  width: '2.5rem',
                  height: '2.5rem',
                  borderRadius: '50%',
                  background: input.trim() && !loading ? 'var(--primary)' : '#D1D5DB',
                  color: 'white',
                  border: 'none',
                  cursor: input.trim() && !loading ? 'pointer' : 'not-allowed',
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  transition: 'background 0.2s'
                }}
              >
                ➤
              </button>
            </div>
          </div>
        </div>
      )}

      <style jsx>{`
        @keyframes typing {
          0%, 60%, 100% { transform: translateY(0); }
          30% { transform: translateY(-4px); }
        }
        .typing-dot::before {
          content: '●';
          animation: typing 1s infinite;
        }
      `}</style>
    </>
  )
}
