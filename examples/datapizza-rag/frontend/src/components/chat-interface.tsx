'use client'

import { useState, useRef, useEffect } from 'react'
import { Message } from '@ai-sdk/react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { ScrollArea } from '@/components/ui/scroll-area'
import { Separator } from '@/components/ui/separator'
import { 
  Menu, 
  Send, 
  Square, 
  RotateCcw,
  Sparkles,
  FileText,
  Globe,
  Github,
  Paperclip,
  Search,
  X,
  Upload
} from 'lucide-react'
import { MessageBubble } from '@/components/message-bubble'
import { TypingIndicator } from '@/components/typing-indicator'
import { ChatInputArea } from '@/components/chat-input-area'
import { cn } from '@/lib/utils'

interface ChatInterfaceProps {
  messages: Message[]
  input: string
  handleInputChange: (e: React.ChangeEvent<HTMLInputElement>) => void
  handleSubmit: (e: React.FormEvent<HTMLFormElement>) => void
  isLoading: boolean
  error: Error | undefined
  reload: () => void
  stop: () => void
  sidebarOpen: boolean
  onToggleSidebar: () => void
}

const QUICK_ACTIONS = [
  {
    icon: FileText,
    label: 'Ask about documents',
    prompt: 'What documents do I have in my knowledge base?'
  },
  {
    icon: Globe,
    label: 'Research on web',
    prompt: 'Research the latest developments in AI and vector databases'
  },
  {
    icon: Github,
    label: 'Analyze code',
    prompt: 'Show me the code structure of my indexed repositories'
  },
  {
    icon: Sparkles,
    label: 'Get insights',
    prompt: 'What are the key insights from my knowledge base?'
  }
]

export function ChatInterface({
  messages,
  input,
  handleInputChange,
  handleSubmit,
  isLoading,
  error,
  reload,
  stop,
  sidebarOpen,
  onToggleSidebar
}: ChatInterfaceProps) {
  const scrollAreaRef = useRef<HTMLDivElement>(null)
  const inputRef = useRef<HTMLInputElement>(null)
  const [showQuickActions, setShowQuickActions] = useState(messages.length === 0)

  // Auto-scroll to bottom when new messages arrive
  useEffect(() => {
    if (scrollAreaRef.current) {
      const scrollElement = scrollAreaRef.current.querySelector('[data-radix-scroll-area-viewport]')
      if (scrollElement) {
        scrollElement.scrollTop = scrollElement.scrollHeight
      }
    }
  }, [messages])

  // Hide quick actions when user starts typing
  useEffect(() => {
    if (messages.length > 0) {
      setShowQuickActions(false)
    }
  }, [messages.length])

  const handleQuickAction = (prompt: string) => {
    const syntheticEvent = {
      target: { value: prompt }
    } as React.ChangeEvent<HTMLInputElement>
    
    handleInputChange(syntheticEvent)
    setShowQuickActions(false)
    
    // Focus input after setting the prompt
    setTimeout(() => {
      inputRef.current?.focus()
    }, 0)
  }

  const onSubmit = (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault()
    if (!input.trim() || isLoading) return
    
    setShowQuickActions(false)
    handleSubmit(e)
  }

  return (
    <div className="flex flex-col h-full">
      {/* Header */}
      <div className="flex items-center justify-between p-4 border-b bg-card/50 backdrop-blur-sm">
        <div className="flex items-center gap-3">
          <Button
            variant="ghost"
            size="sm"
            onClick={onToggleSidebar}
            className="lg:hidden"
          >
            <Menu className="h-4 w-4" />
          </Button>
          
          <div className="flex items-center gap-2">
            <div className="w-8 h-8 rounded-full bg-gradient-to-br from-blue-500 to-purple-600 flex items-center justify-center">
              <Sparkles className="h-4 w-4 text-white" />
            </div>
            <div>
              <h1 className="font-semibold text-lg">VittoriaDB Assistant</h1>
              <p className="text-sm text-muted-foreground">
                Powered by AI • Connected to your knowledge base
              </p>
            </div>
          </div>
        </div>

        <div className="flex items-center gap-2">
          {error && (
            <Button
              variant="outline"
              size="sm"
              onClick={reload}
              className="text-destructive hover:text-destructive"
            >
              <RotateCcw className="h-4 w-4 mr-1" />
              Retry
            </Button>
          )}
          
          {isLoading && (
            <Button
              variant="outline"
              size="sm"
              onClick={stop}
              className="text-orange-600 hover:text-orange-700"
            >
              <Square className="h-4 w-4 mr-1" />
              Stop
            </Button>
          )}
        </div>
      </div>

      {/* Messages Area */}
      <ScrollArea ref={scrollAreaRef} className="flex-1 p-4">
        <div className="max-w-4xl mx-auto space-y-6">
          {/* Welcome Message */}
          {messages.length === 0 && (
            <div className="text-center py-12">
              <div className="w-16 h-16 rounded-full bg-gradient-to-br from-blue-500 to-purple-600 flex items-center justify-center mx-auto mb-4">
                <Sparkles className="h-8 w-8 text-white" />
              </div>
              <h2 className="text-2xl font-bold mb-2">Welcome to Your Personal Assistant</h2>
              <p className="text-muted-foreground mb-8 max-w-md mx-auto">
                I can help you search through your documents, research topics on the web, 
                and analyze code from your indexed repositories.
              </p>
              
              {/* Quick Actions */}
              {showQuickActions && (
                <div className="grid grid-cols-1 md:grid-cols-2 gap-3 max-w-2xl mx-auto">
                  {QUICK_ACTIONS.map((action, index) => (
                    <Button
                      key={index}
                      variant="outline"
                      className="h-auto p-4 text-left justify-start hover:bg-accent/50 transition-colors"
                      onClick={() => handleQuickAction(action.prompt)}
                    >
                      <action.icon className="h-5 w-5 mr-3 text-primary" />
                      <div>
                        <div className="font-medium">{action.label}</div>
                        <div className="text-sm text-muted-foreground mt-1">
                          {action.prompt}
                        </div>
                      </div>
                    </Button>
                  ))}
                </div>
              )}
            </div>
          )}

          {/* Chat Messages */}
          {messages.map((message, index) => (
            <MessageBubble
              key={index}
              message={message}
              isLast={index === messages.length - 1}
            />
          ))}

          {/* Typing Indicator */}
          {isLoading && <TypingIndicator />}

          {/* Error Message */}
          {error && (
            <div className="flex items-center justify-center p-4">
              <div className="bg-destructive/10 text-destructive px-4 py-2 rounded-lg text-sm">
                Error: {error.message}
              </div>
            </div>
          )}
        </div>
      </ScrollArea>

      {/* Enhanced Input Area */}
      <ChatInputArea
        input={input}
        handleInputChange={handleInputChange}
        handleSubmit={handleSubmit}
        isLoading={isLoading}
        inputRef={inputRef}
      />
    </div>
  )
}
