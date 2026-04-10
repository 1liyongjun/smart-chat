import type { Metadata } from 'next'
import './globals.css'

export const metadata: Metadata = {
  title: '智能客服系统',
  description: '基于 MiniMax AI 的知识库问答客服',
}

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="zh-CN">
      <body>{children}</body>
    </html>
  )
}
