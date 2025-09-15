'use client'

import { Message } from '@ai-sdk/react'
import { User, Bot, Copy, Check } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { useState } from 'react'
import { cn } from '@/lib/utils'
import ReactMarkdown from 'react-markdown'
import { Prism as SyntaxHighlighter } from 'react-syntax-highlighter'
import { oneDark } from 'react-syntax-highlighter/dist/esm/styles/prism'

interface MessageBubbleProps {
  message: Message
  isLast?: boolean
}

export function MessageBubble({ message, isLast }: MessageBubbleProps) {
  const [copied, setCopied] = useState(false)
  const isUser = message.role === 'user'

  const handleCopy = async () => {
    await navigator.clipboard.writeText(message.content)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }

  return (
    <div className={cn(
      "flex gap-3 group",
      isUser ? "justify-end" : "justify-start"
    )}>
      {!isUser && (
        <div className="w-8 h-8 rounded-full bg-gradient-to-br from-blue-500 to-purple-600 flex items-center justify-center flex-shrink-0">
          <Bot className="h-4 w-4 text-white" />
        </div>
      )}
      
      <div className={cn(
        "max-w-[80%] rounded-lg px-4 py-3 relative",
        isUser 
          ? "bg-primary text-primary-foreground ml-12" 
          : "bg-muted"
      )}>
        <div className="prose prose-sm max-w-none dark:prose-invert">
          {isUser ? (
            <p className="m-0">{message.content}</p>
          ) : (
            <ReactMarkdown
              components={{
                code({ node, inline, className, children, ...props }) {
                  const match = /language-(\w+)/.exec(className || '')
                  return !inline && match ? (
                    <SyntaxHighlighter
                      style={oneDark}
                      language={match[1]}
                      PreTag="div"
                      className="rounded-md !mt-2 !mb-2"
                      {...props}
                    >
                      {String(children).replace(/\n$/, '')}
                    </SyntaxHighlighter>
                  ) : (
                    <code className={cn(
                      "relative rounded bg-muted px-[0.3rem] py-[0.2rem] font-mono text-sm font-semibold",
                      className
                    )} {...props}>
                      {children}
                    </code>
                  )
                },
                p({ children }) {
                  return <p className="mb-2 last:mb-0">{children}</p>
                },
                ul({ children }) {
                  return <ul className="mb-2 ml-4 list-disc">{children}</ul>
                },
                ol({ children }) {
                  return <ol className="mb-2 ml-4 list-decimal">{children}</ol>
                },
                li({ children }) {
                  return <li className="mb-1">{children}</li>
                },
                h1({ children }) {
                  return <h1 className="text-lg font-semibold mb-2">{children}</h1>
                },
                h2({ children }) {
                  return <h2 className="text-base font-semibold mb-2">{children}</h2>
                },
                h3({ children }) {
                  return <h3 className="text-sm font-semibold mb-2">{children}</h3>
                },
                blockquote({ children }) {
                  return <blockquote className="border-l-4 border-border pl-4 italic">{children}</blockquote>
                },
              }}
            >
              {message.content}
            </ReactMarkdown>
          )}
        </div>
        
        {/* Copy button */}
        <Button
          variant="ghost"
          size="sm"
          className={cn(
            "absolute top-2 right-2 h-6 w-6 p-0 opacity-0 group-hover:opacity-100 transition-opacity",
            isUser ? "text-primary-foreground/70 hover:text-primary-foreground" : ""
          )}
          onClick={handleCopy}
        >
          {copied ? (
            <Check className="h-3 w-3" />
          ) : (
            <Copy className="h-3 w-3" />
          )}
        </Button>
      </div>
      
      {isUser && (
        <div className="w-8 h-8 rounded-full bg-gradient-to-br from-green-500 to-blue-600 flex items-center justify-center flex-shrink-0">
          <User className="h-4 w-4 text-white" />
        </div>
      )}
    </div>
  )
}
