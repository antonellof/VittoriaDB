'use client'

import { Bot } from 'lucide-react'

export function TypingIndicator() {
  return (
    <div className="flex gap-3">
      <div className="w-8 h-8 rounded-full bg-gradient-to-br from-blue-500 to-purple-600 flex items-center justify-center flex-shrink-0">
        <Bot className="h-4 w-4 text-white" />
      </div>
      
      <div className="bg-muted rounded-lg px-4 py-3">
        <div className="flex items-center gap-1">
          <div className="w-2 h-2 bg-muted-foreground/40 rounded-full animate-bounce [animation-delay:-0.3s]"></div>
          <div className="w-2 h-2 bg-muted-foreground/40 rounded-full animate-bounce [animation-delay:-0.15s]"></div>
          <div className="w-2 h-2 bg-muted-foreground/40 rounded-full animate-bounce"></div>
        </div>
      </div>
    </div>
  )
}
