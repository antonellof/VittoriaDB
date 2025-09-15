'use client'

import { useState, useRef, useCallback } from 'react'
import { useDropzone } from 'react-dropzone'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Badge } from '@/components/ui/badge'
import { Switch } from '@/components/ui/switch'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import { 
  Send, 
  Paperclip, 
  Search, 
  X, 
  Upload,
  Globe,
  FileText,
  Loader2,
  CheckCircle,
  Square,
  Sparkles
} from 'lucide-react'
import { cn, getFileIcon, formatFileSize } from '@/lib/utils'
import { useStatsStore } from '@/store'
import toast from 'react-hot-toast'

interface EnhancedChatInputProps {
  onSubmit: (message: string, options: { webSearch?: boolean, files?: File[] }) => Promise<void>
  isLoading: boolean
  onStop: () => void
}

interface AttachedFile {
  file: File
  id: string
  result?: any
  status?: 'uploading' | 'success' | 'error'
}

export function EnhancedChatInput({ onSubmit, isLoading, onStop }: EnhancedChatInputProps) {
  const [input, setInput] = useState('')
  const [webSearchEnabled, setWebSearchEnabled] = useState(false)
  const [attachedFiles, setAttachedFiles] = useState<AttachedFile[]>([])
  const [isExpanded, setIsExpanded] = useState(false)
  const inputRef = useRef<HTMLTextAreaElement>(null)
  const { fetchStats, setCollectionUpdating } = useStatsStore()

  const onDrop = useCallback(async (acceptedFiles: File[]) => {
    // Add files with uploading status
    const newFiles: AttachedFile[] = acceptedFiles.map(file => ({
      file,
      id: Math.random().toString(36).substring(2),
      status: 'uploading' as const
    }))
    
    setAttachedFiles(prev => [...prev, ...newFiles])
    setIsExpanded(true)
    
    // Mark documents collection as updating
    setCollectionUpdating('documents', true)
    
    // Upload files immediately, one by one
    for (const file of acceptedFiles) {
      try {
        toast.loading(`📁 Uploading ${file.name}...`, { id: file.name })
        
        const formData = new FormData()
        formData.append('file', file)
        
        const response = await fetch('http://localhost:8501/upload', {
          method: 'POST',
          body: formData,
        })
        
        if (!response.ok) {
          throw new Error(`Failed to upload ${file.name}`)
        }
        
        const result = await response.json()
        
        // Update specific file status
        setAttachedFiles(prev => 
          prev.map(attachedFile => 
            attachedFile.file.name === file.name 
              ? { ...attachedFile, status: 'success' as const, result }
              : attachedFile
          )
        )
        
        toast.success(`✅ ${file.name} uploaded and indexed (${result.chunks_created} chunks)`, { id: file.name })
        
      } catch (error: any) {
        // Update file status to error
        setAttachedFiles(prev => 
          prev.map(attachedFile => 
            attachedFile.file.name === file.name 
              ? { ...attachedFile, status: 'error' as const }
              : attachedFile
          )
        )
        
        toast.error(`❌ Failed to upload ${file.name}`, { id: file.name })
        console.error('Upload error:', error)
      }
    }
    
    // Clear updating status and refresh stats
    setCollectionUpdating('documents', false)

  }, [setCollectionUpdating, fetchStats])

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

  const removeFile = (id: string) => {
    setAttachedFiles(prev => prev.filter(f => f.id !== id))
    if (attachedFiles.length === 1) {
      setIsExpanded(false)
    }
  }

  const clearAllFiles = () => {
    setAttachedFiles([])
    setIsExpanded(false)
  }

  const triggerFileUpload = () => {
    const fileInput = document.createElement('input')
    fileInput.type = 'file'
    fileInput.multiple = true
    fileInput.accept = '.pdf,.docx,.doc,.txt,.md,.html,.htm'
    fileInput.onchange = (e) => {
      const files = Array.from((e.target as HTMLInputElement).files || [])
      onDrop(files)
    }
    fileInput.click()
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!input.trim() || isLoading) return

    const message = input.trim()
    
    // Clear input (keep files visible to show they're already processed)
    setInput('')

    // Submit with options (files are already uploaded, so don't include them)
    await onSubmit(message, {
      webSearch: webSearchEnabled
    })
  }

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault()
      handleSubmit(e as any)
    }
  }

  const placeholder = webSearchEnabled 
    ? "Ask a question - I'll research it on the web and answer..."
    : attachedFiles.length > 0
    ? "Ask questions about your uploaded files..."
    : "Ask me anything about your documents..."

  return (
    <div 
      {...getRootProps()}
      className={cn(
        "relative transition-all duration-200 rounded-lg border",
        isDragActive && "border-primary bg-primary/5",
        isExpanded && "bg-muted/20"
      )}
    >
      <input {...getInputProps()} />
      
      {/* Drag Overlay */}
      {isDragActive && (
        <div className="absolute inset-0 z-10 flex items-center justify-center bg-primary/10 rounded-lg border-2 border-dashed border-primary">
          <div className="text-center">
            <Upload className="h-8 w-8 mx-auto mb-2 text-primary" />
            <p className="text-sm font-medium text-primary">Drop files here to attach</p>
            <p className="text-xs text-muted-foreground">Supports: PDF, DOCX, TXT, MD, HTML</p>
          </div>
        </div>
      )}

      {/* Attached Files Display */}
      {isExpanded && attachedFiles.length > 0 && (
        <div className="p-3 border-b bg-muted/30">
          <div className="flex items-center justify-between mb-2">
            <div className="flex items-center gap-2">
              <FileText className="h-4 w-4" />
              <span className="text-sm font-medium">Attached Files</span>
              <Badge variant="secondary">{attachedFiles.length}</Badge>
            </div>
            <Button
              variant="ghost"
              size="sm"
              onClick={clearAllFiles}
              className="h-6 w-6 p-0"
            >
              <X className="h-3 w-3" />
            </Button>
          </div>
          
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-2">
            {attachedFiles.map((attachedFile) => (
              <div
                key={attachedFile.id}
                className="flex items-center gap-2 p-2 rounded bg-background border text-xs"
              >
                <span>{getFileIcon(attachedFile.file.name)}</span>
                
                <div className="flex-1 min-w-0">
                  <div className="font-medium truncate">
                    {attachedFile.file.name}
                  </div>
                  <div className="text-muted-foreground">
                    {formatFileSize(attachedFile.file.size)}
                    {attachedFile.result && (
                      <span className="text-green-600 ml-1">
                        • {attachedFile.result.chunks_created} chunks
                      </span>
                    )}
                  </div>
                </div>

                <div className="flex items-center gap-1">
                  {attachedFile.status === 'uploading' && (
                    <Loader2 className="h-3 w-3 animate-spin text-blue-600" />
                  )}
                  {attachedFile.status === 'success' && (
                    <CheckCircle className="h-3 w-3 text-green-600" />
                  )}
                  {attachedFile.status === 'error' && (
                    <X className="h-3 w-3 text-red-600" />
                  )}
                  
                  <Button
                    variant="ghost"
                    size="sm"
                    className="h-5 w-5 p-0"
                    onClick={() => removeFile(attachedFile.id)}
                  >
                    <X className="h-3 w-3" />
                  </Button>
                </div>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Options Bar */}
      <div className="flex items-center justify-between p-3 border-b">
        <div className="flex items-center gap-4">
          {/* Web Search Toggle */}
          <div className="flex items-center gap-2">
            <Switch
              id="web-search"
              checked={webSearchEnabled}
              onCheckedChange={setWebSearchEnabled}
              disabled={isLoading}
            />
            <Label htmlFor="web-search" className="text-xs flex items-center gap-1 cursor-pointer">
              <Search className="h-3 w-3" />
              Web Research
            </Label>
          </div>
          
          {/* File Attach Button */}
          <Button
            variant="ghost"
            size="sm"
            onClick={triggerFileUpload}
            disabled={isLoading}
            className="text-xs"
          >
            <Paperclip className="h-3 w-3 mr-1" />
            Attach
          </Button>
        </div>

        {/* Status Indicators */}
        <div className="flex items-center gap-2">
          {webSearchEnabled && (
            <Badge variant="default" className="text-xs">
              <Globe className="h-3 w-3 mr-1" />
              Web Search
            </Badge>
          )}
          
          {attachedFiles.length > 0 && (
            <Badge variant="secondary" className="text-xs">
              <FileText className="h-3 w-3 mr-1" />
              {attachedFiles.length} files
            </Badge>
          )}
        </div>
      </div>

      {/* Input Form */}
      <form onSubmit={handleSubmit} className="p-3">
        <div className="flex gap-2">
          <div className="flex-1 relative">
            <Textarea
              ref={inputRef}
              value={input}
              onChange={(e) => setInput(e.target.value)}
              onKeyDown={handleKeyDown}
              placeholder={placeholder}
              className="min-h-[48px] max-h-32 resize-none pr-12"
              disabled={isLoading}
              autoFocus
              rows={1}
            />
            
            {/* Send/Stop Button */}
            <div className="absolute right-2 bottom-2">
              {isLoading ? (
                <Button
                  type="button"
                  size="sm"
                  variant="outline"
                  className="h-8 w-8 p-0"
                  onClick={onStop}
                >
                  <Square className="h-3 w-3" />
                </Button>
              ) : (
                <Button
                  type="submit"
                  size="sm"
                  className="h-8 w-8 p-0"
                  disabled={!input.trim()}
                >
                  <Send className="h-3 w-3" />
                </Button>
              )}
            </div>
          </div>
        </div>
        
        <div className="text-xs text-muted-foreground text-center mt-2 flex items-center justify-center gap-2">
          <span>Press Enter to send</span>
          <span>•</span>
          <span>Shift+Enter for new line</span>
          <span>•</span>
          <span>Drag files to attach</span>
        </div>
      </form>
    </div>
  )
}
