import type { Metadata } from 'next'
import './globals.scss'

export const metadata: Metadata = {
  title: 'Voice Diary - Голосовой дневник с AI',
  description: 'Записывайте голосовые заметки с автоматической транскрипцией и AI-анализом',
}

export default function RootLayout({
  children,
}: {
  children: React.ReactNode
}) {
  return (
    <html lang="ru">
      <body>{children}</body>
    </html>
  )
}
