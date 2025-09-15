import type { Metadata } from 'next'
import { Inter } from 'next/font/google'
import './globals.css'
import { cn } from '@/lib/utils'
import { ThemeProvider } from '@/components/theme-provider'

const inter = Inter({ subsets: ['latin'] })

export const metadata: Metadata = {
  title: 'VittoriaDB RAG Assistant',
  description: 'Intelligent document processing, web research, and code indexing powered by VittoriaDB',
  keywords: ['AI', 'RAG', 'Vector Database', 'VittoriaDB', 'Chat', 'Document Processing'],
  authors: [{ name: 'VittoriaDB Team' }],
}

export const viewport = {
  width: 'device-width',
  initialScale: 1,
}

export default function RootLayout({
  children,
}: {
  children: React.ReactNode
}) {
  return (
    <html lang="en" suppressHydrationWarning>
      <body className={cn(
        inter.className,
        "min-h-screen bg-background font-sans antialiased"
      )}>
        <ThemeProvider
          attribute="class"
          defaultTheme="system"
          enableSystem
          disableTransitionOnChange
        >
          <div className="relative flex min-h-screen flex-col">
            <div className="flex-1">
              {children}
            </div>
          </div>
        </ThemeProvider>
      </body>
    </html>
  )
}
