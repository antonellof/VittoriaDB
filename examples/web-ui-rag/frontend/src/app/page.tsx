'use client'

import { useChat } from '@ai-sdk/react'
import { useState, useEffect, useCallback } from 'react'
import { flushSync } from 'react-dom'
import { useDropzone } from 'react-dropzone'
import {
  Branch,
  BranchMessages,
  BranchNext,
  BranchPage,
  BranchPrevious,
  BranchSelector,
} from '@/components/ai-elements/branch'
import {
  Conversation,
  ConversationContent,
  ConversationScrollButton,
} from '@/components/ai-elements/conversation'
import { Message, MessageContent } from '@/components/ai-elements/message'
import {
  PromptInput,
  PromptInputButton,
  type PromptInputMessage,
  PromptInputTextarea,
  PromptInputToolbar,
  PromptInputTools,
} from '@/components/ai-elements/prompt-input'
import {
  ChainOfThought,
  ChainOfThoughtContent,
  ChainOfThoughtHeader,
  ChainOfThoughtStep,
  ChainOfThoughtSearchResults,
  ChainOfThoughtSearchResult,
} from '@/components/ai-elements/chain-of-thought'
import { Response } from '@/components/ai-elements/response'
import {
  Source,
  Sources,
  SourcesContent,
  SourcesTrigger,
} from '@/components/ai-elements/sources'
import { Suggestion, Suggestions } from '@/components/ai-elements/suggestion'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { Switch } from '@/components/ui/switch'
import {
  AudioWaveformIcon,
  BarChartIcon,
  BoxIcon,
  CameraIcon,
  Code,
  FileIcon,
  GlobeIcon,
  GraduationCapIcon,
  ImageIcon,
  NotebookPenIcon,
  PaperclipIcon,
  ScreenShareIcon,
  Settings,
  Menu,
  Database,
  Activity,
  Send,
  ArrowUp,
  Brain,
  ChevronDown,
  ChevronDownIcon,
  ChevronRight,
  Search,
  SearchIcon,
  ExternalLink,
  BookOpen,
  Plus,
  MessageSquare,
  Save,
  ArrowDown
} from 'lucide-react'
import { nanoid } from 'nanoid'
import { useTheme } from 'next-themes'
import { Toaster } from 'react-hot-toast'
import { useStatsStore } from '@/store'
import { SettingsPanel } from '@/components/settings-panel'
import { DataSourcesPanel } from '@/components/data-sources-panel'
import { apiClient } from '@/lib/api'
import { useWebSocketNotifications } from '@/hooks/useWebSocketNotifications'
import { cn } from '@/lib/utils'
import toast from 'react-hot-toast'

type MessageType = {
  key: string
  from: 'user' | 'assistant'
  sources?: { href: string; title: string }[]
  versions: {
    id: string
    content: string
  }[]
  reasoning?: {
    content: string
    duration: number
    steps?: Array<{
      icon?: any
      label: string
      status: 'active' | 'complete' | 'pending'
      searchResults?: Array<{
        title?: string
        url?: string
        type?: 'web_search' | 'knowledge_base'
        favicon?: string
      }>
    }>
  }
  avatar: string
  name: string
  isReasoningComplete?: boolean
  isContentComplete?: boolean
  isReasoningStreaming?: boolean
}

const suggestions = [
  { icon: FileIcon, text: "What documents do I have in my knowledge base?", color: "#76d0eb" },
  { icon: GlobeIcon, text: "Research the latest developments in AI", color: "#ea8444" },
  { icon: Code, text: "Show me code from my indexed repositories", color: "#6c71ff" },
  { icon: BarChartIcon, text: "Analyze my uploaded documents", color: "#76d0eb" },
  { icon: GraduationCapIcon, text: "Get insights from my knowledge base", color: "#76d0eb" },
  { icon: null, text: "More" },
]

// Icon mapping for ChainOfThought steps
const getStepIcon = (iconName: string) => {
  const iconMap: { [key: string]: any } = {
    'Search': SearchIcon,
    'Database': Database,
    'Brain': Brain,
  }
  return iconMap[iconName] || Brain
}

export default function Home() {
  const [text, setText] = useState('')
  const [useWebSearch, setUseWebSearch] = useState(false)
  const [settingsOpen, setSettingsOpen] = useState(false)
  const [dataSourcesOpen, setDataSourcesOpen] = useState(false)
  const [sidebarOpen, setSidebarOpen] = useState(true)
  const [messages, setMessages] = useState<MessageType[]>([])
  const [status, setStatus] = useState<'submitted' | 'streaming' | 'ready' | 'error'>('ready')
  const [showScrollButton, setShowScrollButton] = useState(false)
  
  const { theme, setTheme } = useTheme()
  const { stats, health, fetchStats, fetchHealth } = useStatsStore()
  const { isConnected, processingFiles } = useWebSocketNotifications()

  // Drag & drop for file upload
  const onDrop = useCallback((acceptedFiles: File[]) => {
    handleFileUpload(acceptedFiles)
  }, [])

  const { getRootProps, getInputProps, isDragActive } = useDropzone({
    onDrop,
    accept: {
      'application/pdf': ['.pdf'],
      'application/vnd.openxmlformats-officedocument.wordprocessingml.document': ['.docx'],
      'application/msword': ['.doc'],
      'text/plain': ['.txt'],
      'text/markdown': ['.md'],
      'text/html': ['.html', '.htm'],
    },
    multiple: true,
    noClick: true,
    noKeyboard: true
  })

  // Custom chat state (not using useChat hook)
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<Error | null>(null)
  const [searchProgress, setSearchProgress] = useState<string[]>([])
  
  // Chat session management
  const [currentSessionId, setCurrentSessionId] = useState<string | null>(null)
  const [chatSessions, setChatSessions] = useState<any[]>([])
  const [autoSaveEnabled, setAutoSaveEnabled] = useState(true)
  
  // Sidebar collapsible sections
  const [collectionsExpanded, setCollectionsExpanded] = useState(true)
  const [systemStatusExpanded, setSystemStatusExpanded] = useState(true)  // Expanded by default to show session status
  
  // Scroll down button state
  const [showScrollDown, setShowScrollDown] = useState(false)


  // Create new chat session
  const createNewChatSession = async (title?: string) => {
    try {
      const response = await fetch('http://localhost:8501/chat/sessions', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          title: title || `Chat ${new Date().toLocaleString()}`
        })
      })
      
      if (response.ok) {
        const data = await response.json()
        setCurrentSessionId(data.session.session_id)
        console.log('New chat session created:', data.session.session_id)
        return data.session.session_id
      }
    } catch (error) {
      console.error('Failed to create chat session:', error)
    }
    return null
  }

  // Save current chat to VittoriaDB
  const saveChatHistory = async () => {
    if (!currentSessionId || messages.length === 0 || !autoSaveEnabled) return
    
    try {
      // Convert messages to the format expected by the backend
      const chatMessages = messages.map(msg => ({
        role: msg.from === 'user' ? 'user' : 'assistant',
        content: msg.versions[0]?.content || '',
        timestamp: Date.now() / 1000,
        sources: msg.sources || []
      }))
      
      const response = await fetch('http://localhost:8501/chat/save', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          session_id: currentSessionId,
          messages: chatMessages
        })
      })
      
      if (response.ok) {
        const data = await response.json()
        console.log('Chat history saved:', data.messages_saved, 'messages')
      }
    } catch (error) {
      console.error('Failed to save chat history:', error)
    }
  }

  // Enhanced new chat function
  const startNewChat = () => {
    
    try {
      // Save current chat in background (non-blocking)
      if (currentSessionId && messages.length > 0) {
        // Fire and forget - don't wait for completion
        saveChatHistory().catch(error => {
          console.error('Background save failed:', error)
        })
      }
      
      // Immediately clear current chat and reset to initial state
      setMessages([])
      setSearchProgress([])
      setError(null)
      setIsLoading(false)
      setStatus('ready')  // Reset status to show welcome screen
      setText('')  // Clear input text
      setCurrentSessionId(null)  // Clear session to show "None" status
      
      // Note: New session will be created automatically on first message
      
    } catch (error) {
      console.error('❌ Error starting new chat:', error)
      // Make sure we don't leave the UI in a broken state
      setIsLoading(false)
    }
  }

  // Send message directly to backend using new RAG streaming endpoint
  const sendMessage = async (userMessage: string, options: { webSearch?: boolean } = {}) => {
    try {
      setIsLoading(true)
      setError(null)
      setSearchProgress([])
      
      // Create session if this is the first message
      if (!currentSessionId) {
        const sessionId = await createNewChatSession(`Chat: ${userMessage.slice(0, 30)}...`)
        if (!sessionId) {
          throw new Error('Failed to create chat session')
        }
      }
      
      // Add user message
      const userMsg: MessageType = {
        key: `user-${Date.now()}`,
        from: 'user',
        versions: [{ id: `version-${Date.now()}`, content: userMessage }],
        avatar: '',
        name: 'You',
        isContentComplete: true,
        isReasoningComplete: true
      }
      
      setMessages(prev => [...prev, userMsg])
      
      // Create assistant message placeholder
      const assistantMsg: MessageType = {
        key: `assistant-${Date.now()}`,
        from: 'assistant',
        versions: [{ id: `version-${Date.now()}`, content: '' }],
        avatar: '',
        name: 'VittoriaDB Assistant',
        isContentComplete: false,
        isReasoningComplete: false
      }
      
      setMessages(prev => [...prev, assistantMsg])
      
      // Stream from backend
      const response = await fetch('http://localhost:8501/rag/stream', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          message: userMessage,
          chat_history: [],
          search_collections: ['documents', 'web_research', 'github_code'],
          model: 'gpt-4',
          search_limit: 5,
          web_search: options.webSearch || false
        })
      })

      if (!response.ok) {
        throw new Error('Backend request failed')
      }

      const reader = response.body?.getReader()
      if (!reader) {
        throw new Error('No response body')
      }

      const decoder = new TextDecoder()
      let buffer = ''
      
      while (true) {
        const { done, value } = await reader.read()
        
        if (done) break
        
        buffer += decoder.decode(value, { stream: true })
        const lines = buffer.split('\n')
        buffer = lines.pop() || ''
        
        for (const line of lines) {
          if (line.startsWith('data: ')) {
            try {
              const data = JSON.parse(line.slice(6))
              
              if (data.type === 'reasoning_start') {
                // Start reasoning/thinking - hide loading state
// console.log('🧠 Reasoning start received:', data)
                setIsLoading(false)
                setMessages(prev => 
                  prev.map(msg => 
                    msg.key === assistantMsg.key 
                      ? {
                          ...msg,
                          reasoning: {
                            content: '🧠 Thinking...',
                            duration: 0
                          },
                          isReasoningStreaming: true
                        }
                      : msg
                  )
                )
              } else if (data.type === 'reasoning_update') {
                // Update reasoning content or steps
// console.log('🔄 Reasoning update received:', data)
                setMessages(prev => 
                  prev.map(msg => 
                    msg.key === assistantMsg.key 
                      ? {
                          ...msg,
                          reasoning: {
                            content: data.content || '',
                            duration: 0,
                            steps: data.steps || msg.reasoning?.steps
                          },
                          isReasoningStreaming: true
                        }
                      : msg
                  )
                )
              } else if (data.type === 'reasoning_complete') {
                // Complete reasoning
// console.log('✅ Reasoning complete received:', data)
                setMessages(prev => 
                  prev.map(msg => 
                    msg.key === assistantMsg.key 
                      ? {
                          ...msg,
                          reasoning: {
                            content: data.content || '',
                            duration: 3,
                            steps: data.steps || msg.reasoning?.steps
                          },
                          isReasoningStreaming: false,
                          isReasoningComplete: true
                        }
                      : msg
                  )
                )
              } else if (data.type === 'search_progress') {
                // Add search progress message
                setSearchProgress(prev => [...prev, data.message])
              } else if (data.type === 'content') {
                // Update assistant message content with immediate flush
                flushSync(() => {
                  setMessages(prev => 
                    prev.map(msg => 
                      msg.key === assistantMsg.key 
                        ? {
                            ...msg,
                            versions: msg.versions.map(v => ({
                              ...v,
                              content: v.content + data.content
                            }))
                          }
                        : msg
                    )
                  )
                })
              } else if (data.type === 'done') {
                // Mark as complete
                setMessages(prev => {
                  const updatedMessages = prev.map(msg => 
                    msg.key === assistantMsg.key 
                      ? { ...msg, isContentComplete: true, isReasoningComplete: true }
                      : msg
                  )
                  
                  // Auto-save chat history when conversation completes
                  if (autoSaveEnabled && currentSessionId) {
                    setTimeout(() => saveChatHistory(), 1000) // Save after 1 second delay
                  }
                  
                  return updatedMessages
                })
                setSearchProgress([])
              }
            } catch (parseError) {
              console.error('Parse error:', parseError)
            }
          }
        }
      }
      
    } catch (error: any) {
      console.error('Chat error:', error)
      setError(error)
      toast.error('Chat error: ' + error.message)
    } finally {
      setIsLoading(false)
    }
  }

  // Fetch initial stats
  useEffect(() => {
    
    // Auto-refresh stats every 10 seconds
    const statsInterval = setInterval(fetchStats, 10000)
    const healthInterval = setInterval(fetchHealth, 30000)
    
    return () => {
      clearInterval(statsInterval)
      clearInterval(healthInterval)
    }
  }, [fetchStats, fetchHealth])

  // Auto-scroll to bottom when new messages arrive
  useEffect(() => {
    if (messages.length > 0) {
      setTimeout(() => {
        const conversationElement = document.querySelector('[data-radix-scroll-area-viewport]')
        if (conversationElement) {
          conversationElement.scrollTo({
            top: conversationElement.scrollHeight,
            behavior: 'smooth'
          })
        }
      }, 100)
    }
  }, [messages])

  // Scroll detection for long responses
  useEffect(() => {
    const handleScroll = () => {
      // Find the viewport container that has fixed height from page layout
      let conversationElement = null
      
      // Look for the main conversation viewport - likely has viewport height constraints
      const candidates = [
        // Look for elements with viewport-based heights
        ...document.querySelectorAll('[class*="h-screen"]'),
        ...document.querySelectorAll('[class*="h-full"]'),
        ...document.querySelectorAll('[class*="min-h"]'),
        // Look for elements with overflow auto that have height constraints
        ...document.querySelectorAll('div[style*="overflow: auto"]'),
        // Fallback selectors
        ...document.querySelectorAll('[data-radix-scroll-area-viewport]')
      ]
      
      // Find the element that actually has scrollable content
      for (const element of candidates) {
        const { scrollHeight, clientHeight } = element
        const computedStyle = window.getComputedStyle(element)
        
        // Check if this element has a real height constraint and scrollable content
        const hasRealHeight = 
          computedStyle.height !== 'auto' ||
          element.classList.contains('h-full') ||
          element.classList.contains('h-screen') ||
          element.style.height.includes('%') ||
          element.style.height.includes('vh')
        
        const isActuallyScrollable = scrollHeight > clientHeight
        
        if (hasRealHeight || isActuallyScrollable) {
          conversationElement = element
          break
        }
      }
      
      // Check window scroll instead of element scroll (since that's what we actually scroll)
      const windowScrollTop = document.documentElement.scrollTop || document.body.scrollTop
      const windowScrollHeight = document.documentElement.scrollHeight
      const windowClientHeight = window.innerHeight
      
      // Footer compensation - same calculation as scrollToBottom function
      const footerHeight = 166
      const bufferSpace = -166
      const targetScrollPosition = Math.max(0, windowScrollHeight - windowClientHeight - footerHeight - bufferSpace)
      
      // Check if window content is scrollable
      const isWindowScrollable = windowScrollHeight > windowClientHeight
      
      // Check if we're at/near the target scroll position (where the scroll button takes us)
      const distanceFromTarget = Math.abs(windowScrollTop - targetScrollPosition)
      const isAtTarget = distanceFromTarget < 30 // Within 30px of target position
      
      // Show button when window content is scrollable AND we're not at the target position
      const shouldShowButton = isWindowScrollable && !isAtTarget
        
      // Update scroll down button visibility
      setShowScrollDown(shouldShowButton)
    }

    // Try to find and attach to scroll container
    const findAndAttachScroll = () => {
      let element = null
      
      // First try to find the div with inline style "overflow: auto"
      const allDivs = document.querySelectorAll('div')
      for (const div of allDivs) {
        const style = div.getAttribute('style') || ''
        if (style.includes('overflow: auto') || style.includes('overflow:auto')) {
          element = div
          break
        }
      }
      
      // Fallback to other selectors if not found
      if (!element) {
        const selectors = [
          '[data-radix-scroll-area-viewport]',
          '.conversation-content',
          '.h-full.max-w-4xl'
        ]
        
        for (const selector of selectors) {
          element = document.querySelector(selector)
          if (element) {
            break
          }
        }
      }
      
      if (element) {
        element.addEventListener('scroll', handleScroll)
        // Check initially and frequently for dynamic content
        setTimeout(handleScroll, 50)
        setTimeout(handleScroll, 200)
        setTimeout(handleScroll, 500)
        setTimeout(handleScroll, 1000)
        setTimeout(handleScroll, 2000) // Extra checks for streaming content
        
        return () => element.removeEventListener('scroll', handleScroll)
      }
      
      return () => {}
    }

    const cleanup = findAndAttachScroll()
    
    // Also listen for window resize to check overflow
    const handleResize = () => {
      setTimeout(handleScroll, 100)
    }
    
    window.addEventListener('resize', handleResize)
    
    // Add MutationObserver to detect content changes (streaming responses)
    let observerElement = null
    
    // Find the div with overflow: auto for mutation observation
    const allDivs = document.querySelectorAll('div')
    for (const div of allDivs) {
      const style = div.getAttribute('style') || ''
      if (style.includes('overflow: auto') || style.includes('overflow:auto')) {
        observerElement = div
        break
      }
    }
    
    // Fallback to other elements
    if (!observerElement) {
      observerElement = document.querySelector('[data-radix-scroll-area-viewport]') || 
                       document.querySelector('.conversation-content')
    }
    
    let observer = null
    if (observerElement) {
      observer = new MutationObserver(() => {
        // Content changed, check scroll after a brief delay
        setTimeout(handleScroll, 100)
      })
      
      observer.observe(observerElement, {
        childList: true,
        subtree: true,
        characterData: true
      })
      
    }
    
    return () => {
      cleanup()
      window.removeEventListener('resize', handleResize)
      if (observer) {
        observer.disconnect()
      }
    }
  }, [messages])

  const scrollToBottom = () => {
    // Simple scroll to bottom with fixed pixel offset for footer
    const footerHeight = 166 // Fixed footer with chat input
    const bufferSpace = -166  // Extra space above footer
    const targetScrollTop = Math.max(0, document.documentElement.scrollHeight - window.innerHeight - footerHeight - bufferSpace)
    
    window.scrollTo({
      top: targetScrollTop,
      behavior: 'smooth'
    })
  }

  const handleFileUpload = async (files: File[]) => {
    if (!files || files.length === 0) return

    for (const file of files) {
      const fileId = `upload-${file.name}-${Date.now()}`
      
      try {
        // Phase 1: File upload
        toast.loading(`📤 Uploading ${file.name}...`, { id: fileId })
        
        const uploadStart = Date.now()
        // Upload directly to backend
        const formData = new FormData()
        formData.append('file', file)
        
        const uploadResponse = await fetch('http://localhost:8501/upload', {
          method: 'POST',
          body: formData,
        })
        
        if (!uploadResponse.ok) {
          throw new Error(`Upload failed: ${uploadResponse.statusText}`)
        }
        
        const result = await uploadResponse.json()
        const uploadTime = Date.now() - uploadStart
        
        // Phase 2: Processing feedback
        toast.loading(`⚙️ Processing ${file.name} (${result.chunks_created} chunks created)...`, { id: fileId })
        
        // Wait for processing to complete and verify in VittoriaDB
        let processingComplete = false
        let attempts = 0
        const maxAttempts = 20 // 10 seconds max
        
        while (!processingComplete && attempts < maxAttempts) {
          await new Promise(resolve => setTimeout(resolve, 500))
          
          try {
            // Check if the document appears in search results
            const searchResponse = await fetch('http://localhost:8501/search', {
              method: 'POST',
              headers: { 'Content-Type': 'application/json' },
              body: JSON.stringify({
                query: file.name.replace(/\.[^/.]+$/, ''), // Remove extension for search
                collections: ['documents'],
                limit: 1,
                min_score: 0.1
              })
            })
            
            if (searchResponse.ok) {
              const searchData = await searchResponse.json()
              if (searchData.results && searchData.results.length > 0) {
                // Check if any result is from our uploaded file
                const hasOurFile = searchData.results.some((r: any) => 
                  r.metadata?.filename === file.name ||
                  r.metadata?.content_hash === result.metadata?.content_hash
                )
                
                if (hasOurFile) {
                  processingComplete = true
                }
              }
            }
          } catch (searchError) {
            console.warn('Search verification failed:', searchError)
          }
          
          attempts++
        }
        
        const totalTime = Date.now() - uploadStart
        
        if (processingComplete) {
          toast.success(
            `✅ ${file.name} fully processed and indexed (${result.chunks_created} chunks, ${totalTime}ms)`, 
            { id: fileId, duration: 4000 }
          )
        } else {
          toast.success(
            `✅ ${file.name} uploaded (${result.chunks_created} chunks) - indexing in progress`, 
            { id: fileId, duration: 3000 }
          )
        }

        
      } catch (error: any) {
        toast.error(`❌ Failed to upload ${file.name}: ${error.message}`, { id: fileId })
        console.error('Upload error:', error)
      }
    }
  }

  const handleFileAction = async (action: string) => {
    if (action === 'upload-file') {
      const input = document.createElement('input')
      input.type = 'file'
      input.multiple = true
      input.accept = '.pdf,.docx,.doc,.txt,.md,.html,.htm'
      
      input.onchange = async (e) => {
        const files = Array.from((e.target as HTMLInputElement).files || [])
        
        if (files.length > 0) {
          await handleFileUpload(files)
        }
      }
      
      input.click()
    } else {
      toast.success('File action', { description: action })
    }
  }

  const handleSuggestionClick = (suggestion: string) => {
    setStatus('submitted')
    // Auto-enable web search for research-related suggestions
    const isResearchQuery = suggestion.toLowerCase().includes('research') || 
                           suggestion.toLowerCase().includes('latest') ||
                           suggestion.toLowerCase().includes('developments')
    sendMessage(suggestion, { webSearch: isResearchQuery || useWebSearch })
  }

  const handleSubmit = async (message: PromptInputMessage) => {
    const hasText = Boolean(message.text)
    const hasAttachments = Boolean(message.files?.length)


    if (!(hasText || hasAttachments)) {
      return
    }

    setStatus('submitted')
    
    // Handle file attachments first
    if (message.files?.length) {
      await handleFileUpload(Array.from(message.files))
      
      // If only files were uploaded (no text), show a message
      if (!hasText) {
        toast.success(`📁 Uploaded ${message.files.length} file(s) to your knowledge base`)
        setStatus('ready')
        return
      }
    }
    
    // Send text message
    if (hasText) {
      sendMessage(message.text, { webSearch: useWebSearch })
    }
    
    setText('')
  }

  return (
    <div className="h-screen bg-background flex flex-col">
      {/* Fixed Header */}
      <div className="fixed top-0 left-0 right-0 z-50 border-b bg-card/95 backdrop-blur-sm">
        <div className="flex items-center justify-start p-4">
          <div className="flex items-center gap-3">
            <Button
              variant="ghost"
              size="sm"
              onClick={() => setSidebarOpen(!sidebarOpen)}
            >
              <Menu className="h-4 w-4" />
            </Button>
            
            <div className="flex items-center gap-2">
              <div className="w-8 h-8 rounded-full bg-gradient-to-br from-blue-500 to-purple-600 flex items-center justify-center">
                <span className="text-white font-bold">V</span>
              </div>
              <div>
                <h1 className="font-semibold">Your Personal Assistant</h1>
                <p className="text-xs text-muted-foreground">
                  Powered by VittoriaDB • Connected to your knowledge base
                </p>
              </div>
            </div>
          </div>
        </div>
      </div>

      {/* Main Content Area with top padding for fixed header */}
      <div className="flex flex-1 pt-[73px]">
        {/* Fixed Sidebar */}
        {sidebarOpen && (
          <div className="fixed left-0 top-[73px] bottom-0 w-80 bg-card border-r border-border overflow-y-auto z-40">
            {/* Chat Management Section */}
            <div className="p-4 border-b">
              {/* New Chat Button */}
              <Button
                onClick={() => {
                  startNewChat()
                }}
                className="w-full flex items-center gap-2"
                variant="outline"
                size="sm"
                disabled={false}  // Always enabled for new chat
              >
                <Plus className="h-4 w-4" />
                New Chat
              </Button>
            </div>
            
            {/* Collection Stats */}
            <div className="p-4 space-y-4">
              <div>
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() => setCollectionsExpanded(!collectionsExpanded)}
                  className="w-full justify-start p-2 h-auto mb-3"
                >
                  <div className="flex items-center gap-2 w-full">
                    {collectionsExpanded ? (
                      <ChevronDown className="h-3 w-3" />
                    ) : (
                      <ChevronRight className="h-3 w-3" />
                    )}
                    <Database className="h-4 w-4" />
                    <span className="text-sm font-medium">Collections</span>
                    <div className="flex items-center gap-1 text-xs text-muted-foreground ml-auto">
                      <span>Auto-sync</span>
                    </div>
                  </div>
                </Button>
                
                {collectionsExpanded && stats?.collections ? (
                  <div className="space-y-2">
                    {Object.entries(stats.collections).map(([name, collection]) => (
                      <div key={name} className="bg-muted/50 rounded-lg p-3">
                        <div className="flex items-center justify-between">
                          <span className="text-sm font-medium capitalize">
                            {name.replace('_', ' ')}
                          </span>
                          <span className="text-xs text-muted-foreground">
                            {collection.vector_count} docs
                          </span>
                        </div>
                        <div className="text-xs text-muted-foreground mt-1">
                          {collection.dimensions}D • {collection.index_type?.toUpperCase() || 'FLAT'} • {collection.metric}
                        </div>
                      </div>
                    ))}
                  </div>
                ) : (
                  <div className="text-sm text-muted-foreground">
                    Loading collections...
                  </div>
                )}
              </div>

              {/* System Info */}
              <div>
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() => setSystemStatusExpanded(!systemStatusExpanded)}
                  className="w-full justify-start p-2 h-auto mb-3"
                >
                  <div className="flex items-center gap-2">
                    {systemStatusExpanded ? (
                      <ChevronDown className="h-3 w-3" />
                    ) : (
                      <ChevronRight className="h-3 w-3" />
                    )}
                    <Activity className="h-4 w-4" />
                    <span className="text-sm font-medium">System Status</span>
                  </div>
                </Button>
                
                {systemStatusExpanded && (
                  <div className="space-y-3 text-sm pl-2">
                    <div className="flex justify-between">
                      <span className="text-muted-foreground">Total Vectors:</span>
                      <span>{stats?.total_vectors || 0}</span>
                    </div>
                    
                    <div className="flex justify-between">
                      <span className="text-muted-foreground">OpenAI:</span>
                      <span className={cn(
                        health?.openai_configured ? "text-green-600" : "text-red-600"
                      )}>
                        {health?.openai_configured ? "Configured" : "Not configured"}
                      </span>
                    </div>

                    <div className="flex justify-between">
                      <span className="text-muted-foreground">VittoriaDB:</span>
                      <span className={cn(
                        health?.vittoriadb_connected ? "text-green-600" : "text-red-600"
                      )}>
                        {health?.vittoriadb_connected ? "Connected" : "Disconnected"}
                      </span>
                    </div>
                    
                    {/* Chat Session Status */}
                    <div className="border-t pt-3 space-y-2">
                      <div className="flex justify-between">
                        <span className="text-muted-foreground">Chat Session:</span>
                        <span className={cn(
                          currentSessionId ? "text-green-600" : "text-gray-500"
                        )}>
                          {currentSessionId ? "Active" : "None"}
                        </span>
                      </div>
                      
                      {currentSessionId && (
                        <div className="text-xs text-muted-foreground">
                          ID: {currentSessionId.slice(0, 8)}...
                        </div>
                      )}
                      
                      <div className="flex items-center justify-between">
                        <span className="text-muted-foreground">Auto-save:</span>
                        <Switch
                          checked={autoSaveEnabled}
                          onCheckedChange={setAutoSaveEnabled}
                          size="sm"
                        />
                      </div>
                      
                      {currentSessionId && (
                        <Button
                          onClick={saveChatHistory}
                          variant="ghost"
                          size="sm"
                          className="w-full h-7 text-xs"
                        >
                          <Save className="h-3 w-3 mr-1" />
                          Save Now
                        </Button>
                      )}
                    </div>
                  </div>
                )}
              </div>

              {/* Settings Button in Sidebar */}
              <div className="p-4 border-t mt-auto">
                <Dialog open={settingsOpen} onOpenChange={setSettingsOpen}>
                  <DialogTrigger asChild>
                    <Button variant="outline" className="w-full justify-start" size="sm">
                      <Settings className="h-4 w-4 mr-2" />
                      Settings
                    </Button>
                  </DialogTrigger>
                  <DialogContent className="max-w-4xl max-h-[80vh] overflow-y-auto">
                    <DialogHeader>
                      <DialogTitle>VittoriaDB RAG Settings</DialogTitle>
                    </DialogHeader>
                    <SettingsPanel />
                  </DialogContent>
                </Dialog>
              </div>
            </div>
          </div>
        )}

        {/* Main Chat Container */}
        <div className={cn(
          "flex-1 flex flex-col",
          sidebarOpen ? "ml-80" : "ml-0"
        )}>
          {/* Chat Messages Area with Drag & Drop */}
          <div 
            {...getRootProps()}
            className={cn(
              "flex-1 overflow-hidden relative",
              isDragActive && "bg-primary/5"
            )}
          >
            <input {...getInputProps()} />
            
            {/* Drag Overlay */}
            {isDragActive && (
              <div className="absolute inset-0 z-50 flex items-center justify-center bg-primary/10 border-2 border-dashed border-primary">
                <div className="text-center">
                  <PaperclipIcon className="h-12 w-12 mx-auto mb-4 text-primary" />
                  <p className="text-lg font-medium text-primary">Drop files here to upload</p>
                  <p className="text-sm text-muted-foreground">Supports: PDF, DOCX, TXT, MD, HTML</p>
                </div>
              </div>
            )}
            
            <div className="h-full max-w-4xl mx-auto">
              <Conversation className="h-full">
                <ConversationContent className="p-4 pb-48">{/* Increased bottom padding for fixed input */}
                  {messages.length === 0 && (
                    <div className="text-center py-12">
                      <div className="w-16 h-16 rounded-full bg-gradient-to-br from-blue-500 to-purple-600 flex items-center justify-center mx-auto mb-4">
                        <span className="text-2xl text-white font-bold">V</span>
                      </div>
                      <h2 className="text-2xl font-bold mb-2">Welcome to Your Personal Assistant</h2>
                      <p className="text-muted-foreground mb-8 max-w-md mx-auto">
                        I can help you search through your documents, research topics on the web, 
                        and analyze code from your indexed repositories.
                      </p>
                      
                      {/* Suggestions in main area */}
                      <div className="mt-8 max-w-2xl mx-auto">
                        <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
                          {suggestions.filter(s => s.icon).map(({ icon: Icon, text, color }) => (
                            <button
                              key={text}
                              onClick={() => handleSuggestionClick(text)}
                              className="flex items-center gap-3 p-4 text-left hover:bg-muted/50 rounded-lg transition-colors group border border-border/30"
                            >
                              {Icon && (
                                <div className="flex-shrink-0">
                                  <Icon size={20} style={{ color }} />
                                </div>
                              )}
                              <div className="flex-1">
                                <div className="font-medium text-sm group-hover:text-foreground">
                                  {text}
                                </div>
                              </div>
                            </button>
                          ))}
                        </div>
                      </div>
                    </div>
                  )}

                  {messages.map(({ versions, ...message }) => (
                    <Branch defaultBranch={0} key={message.key}>
                      <BranchMessages>
                        {versions.map((version) => (
                          <Message
                            from={message.from}
                            key={`${message.key}-${version.id}`}
                          >
                            <div>
                              {message.sources?.length && (
                                <Sources>
                                  <SourcesTrigger count={message.sources.length} />
                                  <SourcesContent>
                                    {message.sources.map((source) => (
                                      <Source
                                        href={source.href}
                                        key={source.href}
                                        title={source.title}
                                      />
                                    ))}
                                  </SourcesContent>
                                </Sources>
                              )}
                              
                              {message.reasoning && (
                                <div className="not-prose max-w-prose space-y-4 mb-6">
                                  {message.isReasoningStreaming && (
                                    <div className="flex w-full items-center gap-2 text-muted-foreground text-sm">
                                      <Brain className="size-4" />
                                      <span className="flex-1 text-left">Thinking...</span>
                                    </div>
                                  )}
                                  <div className="space-y-3">
                                    {message.reasoning.steps && message.reasoning.steps.length > 0 ? (
                                      message.reasoning.steps.map((step, index) => (
                                        <ChainOfThoughtStep
                                          key={index}
                                          icon={getStepIcon(step.icon)}
                                          label={step.label}
                                          status={step.status}
                                        >
                                          {step.searchResults && (
                                            <div className="space-y-2">
                                              {step.searchResults.map((result, resultIndex) => (
                                                <div key={resultIndex} className="flex items-start gap-2">
                                                  {result.type === 'web_search' ? (
                                                    <div className="flex items-center gap-2 p-2 rounded-md bg-blue-50 dark:bg-blue-950/30 border border-blue-200 dark:border-blue-800 min-w-0 flex-1">
                                                      {result.favicon && (
                                                        <img 
                                                          src={result.favicon} 
                                                          alt="" 
                                                          className="w-4 h-4 flex-shrink-0"
                                                          onError={(e) => {
                                                            e.currentTarget.style.display = 'none'
                                                          }}
                                                        />
                                                      )}
                                                      <ExternalLink className="w-3 h-3 text-blue-600 dark:text-blue-400 flex-shrink-0" />
                                                      <a 
                                                        href={result.url} 
                                                        target="_blank" 
                                                        rel="noopener noreferrer"
                                                        className="text-sm text-blue-700 dark:text-blue-300 hover:underline truncate"
                                                      >
                                                        {result.title || result.url || `Web Result ${resultIndex + 1}`}
                                                      </a>
                                                    </div>
                                                  ) : (
                                                    <div className="flex items-center gap-2 p-2 rounded-md bg-green-50 dark:bg-green-950/30 border border-green-200 dark:border-green-800 min-w-0 flex-1">
                                                      <BookOpen className="w-3 h-3 text-green-600 dark:text-green-400 flex-shrink-0" />
                                                      <span className="text-sm text-green-700 dark:text-green-300 truncate">
                                                        {result.title || `Source ${resultIndex + 1}`}
                                                      </span>
                                                      {result.url && (
                                                        <span className="text-xs text-green-600 dark:text-green-400 ml-auto flex-shrink-0">
                                                          {result.url}
                                                        </span>
                                                      )}
                                                    </div>
                                                  )}
                                                </div>
                                              ))}
                                            </div>
                                          )}
                                        </ChainOfThoughtStep>
                                      ))
                                    ) : (
                                      <ChainOfThoughtStep
                                        icon={Brain}
                                        label={message.isReasoningStreaming ? "Thinking..." : "Completed thinking"}
                                        status={message.isReasoningStreaming ? "active" : "complete"}
                                      >
                                        <div className="text-sm whitespace-pre-wrap text-muted-foreground">
                                          {message.reasoning.content || "🧠 Processing..."}
                                        </div>
                                      </ChainOfThoughtStep>
                                    )}
                                  </div>
                                </div>
                              )}
                              
                              {(message.from === 'user' ||
                                message.isReasoningComplete ||
                                !message.reasoning) && (
                                <MessageContent
                                  className={cn(
                                    "group-[.is-user]:rounded-[24px] group-[.is-user]:bg-secondary group-[.is-user]:text-foreground",
                                    "group-[.is-assistant]:bg-transparent group-[.is-assistant]:p-0 group-[.is-assistant]:text-foreground"
                                  )}
                                >
                                  <Response>{version.content}</Response>
                                </MessageContent>
                              )}
                            </div>
                          </Message>
                        ))}
                      </BranchMessages>
                      {versions.length > 1 && (
                        <BranchSelector className="px-0" from={message.from}>
                          <BranchPrevious />
                          <BranchPage />
                          <BranchNext />
                        </BranchSelector>
                      )}
                    </Branch>
                  ))}

                  {/* Thinking Mode / Loader with Search Progress */}
                  {isLoading && (
                    <Message from="assistant">
                      <div>
                        <div className="not-prose max-w-prose space-y-4">
                          <div className="flex w-full items-center gap-2 text-muted-foreground text-sm">
                            <Brain className="size-4" />
                            <span className="flex-1 text-left">Thinking...</span>
                          </div>
                          <div className="space-y-3">
                            <ChainOfThoughtStep
                              icon={Brain}
                              label={useWebSearch 
                                ? "Searching your knowledge base and researching on the web..."
                                : "Searching your knowledge base for relevant information..."
                              }
                              status="active"
                            >
                              {/* Live Search Progress */}
                              <div className="text-xs text-muted-foreground space-y-1 max-h-40 overflow-y-auto">
                                {searchProgress.map((progress, index) => (
                                  <div key={index} className="py-0.5 opacity-80">
                                    {progress}
                                  </div>
                                ))}
                              </div>
                            </ChainOfThoughtStep>
                          </div>
                        </div>
                      </div>
                    </Message>
                  )}
                </ConversationContent>
                <ConversationScrollButton />
                
                {/* Central Scroll Down Button - Positioned above footer, centered in main area */}
                {showScrollDown && (
                  <div className={cn(
                    "fixed bottom-44 z-50 pointer-events-auto transition-all duration-300",
                    sidebarOpen 
                      ? "left-1/2 ml-40 transform -translate-x-1/2" // Center of main area when sidebar open
                      : "left-1/2 transform -translate-x-1/2" // Center of full page when sidebar closed
                  )}>
                    <Button
                      onClick={scrollToBottom}
                      size="sm"
                      className="rounded-full w-10 h-10 p-0 shadow-lg bg-white hover:bg-gray-100 backdrop-blur-sm border border-gray-200 transition-all duration-200 hover:scale-105"
                    >
                      <ArrowDown className="h-4 w-4 text-gray-700" />
                    </Button>
                  </div>
                )}

                
                {/* Custom Scroll Down Button */}
                {showScrollButton && (
                  <div className="absolute bottom-24 left-1/2 transform -translate-x-1/2 z-20">
                    <Button
                      onClick={scrollToBottom}
                      className="rounded-full w-10 h-10 p-0 shadow-lg bg-background/90 hover:bg-background border border-border/50 backdrop-blur-sm"
                      variant="outline"
                    >
                      <ChevronDown className="h-4 w-4 text-muted-foreground" />
                    </Button>
                  </div>
                )}
              </Conversation>
            </div>
          </div>
        </div>
      </div>

      {/* Fixed Input Area at Bottom */}
      <div className={cn(
        "fixed bottom-0 left-0 right-0 z-30 border-t bg-background/95 backdrop-blur-sm",
        sidebarOpen ? "pl-80" : "pl-0"
      )}>
        <div className="max-w-4xl mx-auto p-4">
          <PromptInput
            className="divide-y-0 rounded-[28px] max-w-2xl mx-auto"
            onSubmit={handleSubmit}
          >
            <PromptInputTextarea
              className="px-5 md:text-base"
              onChange={(event) => setText(event.target.value)}
              placeholder="Ask me anything about your documents, or request web research..."
              value={text}
            />
            <PromptInputToolbar className="p-2.5">
              <PromptInputTools>
                <Button
                  className="rounded-full border font-medium h-9 px-3"
                  variant="outline"
                  size="sm"
                  onClick={() => setDataSourcesOpen(true)}
                >
                  <Database size={16} />
                  <span className="ml-1">Add Data</span>
                </Button>
                
                <PromptInputButton
                  className={cn(
                    "rounded-full border font-medium",
                    useWebSearch ? "bg-primary text-primary-foreground" : ""
                  )}
                  onClick={() => setUseWebSearch(!useWebSearch)}
                  variant={useWebSearch ? "default" : "outline"}
                >
                  <GlobeIcon size={16} />
                  <span>Web Search</span>
                </PromptInputButton>
              </PromptInputTools>
              
                <PromptInputButton
                  className={cn(
                    "rounded-full font-medium transition-colors",
                    text.trim() && !isLoading
                      ? "bg-black text-white hover:bg-black/90 dark:bg-white dark:text-black dark:hover:bg-white/90"
                      : "text-muted-foreground"
                  )}
                  disabled={!text.trim() || isLoading}
                  type="submit"
                >
                  <ArrowUp size={16} />
                  <span className="sr-only">{isLoading ? 'Sending...' : 'Send'}</span>
                </PromptInputButton>
            </PromptInputToolbar>
          </PromptInput>
        </div>
      </div>
      
      {/* Data Sources Dialog */}
      <Dialog open={dataSourcesOpen} onOpenChange={setDataSourcesOpen}>
        <DialogContent className="max-w-4xl max-h-[80vh] overflow-y-auto">
          <DataSourcesPanel onClose={() => setDataSourcesOpen(false)} />
        </DialogContent>
      </Dialog>

      {/* Toast notifications */}
      <Toaster 
        position="top-right"
        toastOptions={{
          duration: 4000,
          style: {
            background: 'hsl(var(--card))',
            color: 'hsl(var(--card-foreground))',
            border: '1px solid hsl(var(--border))',
          },
        }}
      />
    </div>
  )
}