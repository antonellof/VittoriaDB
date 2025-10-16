'use client'

import { useState, useRef, useCallback } from 'react'
import { useDropzone } from 'react-dropzone'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Badge } from '@/components/ui/badge'
import { Switch } from '@/components/ui/switch'
import { Label } from '@/components/ui/label'
import { 
  Send, 
  Paperclip, 
  Search, 
  X, 
  Upload,
  Globe,
  FileText,
  Loader2,
  CheckCircle
} from 'lucide-react'
import { apiClient } from '@/lib/api'
import { cn, getFileIcon, formatFileSize } from '@/lib/utils'
import toast from 'react-hot-toast'

interface ChatInputAreaProps {
  input: string
  handleInputChange: (e: React.ChangeEvent<HTMLInputElement>) => void
  handleSubmit: (e: React.FormEvent<HTMLFormElement>) => void
  isLoading: boolean
  inputRef: React.RefObject<HTMLInputElement>
}

interface UploadedFile {
  file: File
  status: 'uploading' | 'success' | 'error'
  result?: any
}

export function ChatInputArea({
  input,
  handleInputChange,
  handleSubmit,
  isLoading,
  inputRef
}: ChatInputAreaProps) {
  const [webSearchEnabled, setWebSearchEnabled] = useState(false)
  const [uploadedFiles, setUploadedFiles] = useState<UploadedFile[]>([])
  const [isUploading, setIsUploading] = useState(false)
  const [showFileArea, setShowFileArea] = useState(false)

  const onDrop = useCallback(async (acceptedFiles: File[]) => {
    if (acceptedFiles.length === 0) return

    setIsUploading(true)
    setShowFileArea(true)
    
    // Initialize upload tracking
    const newFiles: UploadedFile[] = acceptedFiles.map(file => ({
      file,
      status: 'uploading' as const
    }))
    
    setUploadedFiles(prev => [...prev, ...newFiles])

    try {
      // Upload files one by one
      for (let i = 0; i < acceptedFiles.length; i++) {
        const file = acceptedFiles[i]
        
        try {
          const result = await apiClient.uploadFile(file)
          
          setUploadedFiles(prev => 
            prev.map(f => 
              f.file === file 
                ? { ...f, status: 'success', result }
                : f
            )
          )
          
          toast.success(`✅ ${file.name} uploaded and indexed`)
        } catch (error: any) {
          setUploadedFiles(prev => 
            prev.map(f => 
              f.file === file 
                ? { ...f, status: 'error' }
                : f
            )
          )
          
          toast.error(`❌ Failed to upload ${file.name}`)
        }
      }
    } finally {
      setIsUploading(false)
    }
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
    disabled: isUploading,
    noClick: true, // We'll handle clicks manually
    noKeyboard: true
  })

  const removeFile = (index: number) => {
    setUploadedFiles(prev => prev.filter((_, i) => i !== index))
    if (uploadedFiles.length === 1) {
      setShowFileArea(false)
    }
  }

  const clearAllFiles = () => {
    setUploadedFiles([])
    setShowFileArea(false)
  }

  const triggerFileUpload = () => {
    const input = document.createElement('input')
    input.type = 'file'
    input.multiple = true
    input.accept = '.pdf,.docx,.doc,.txt,.md,.html,.htm'
    input.onchange = (e) => {
      const files = Array.from((e.target as HTMLInputElement).files || [])
      onDrop(files)
    }
    input.click()
  }

  const onSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault()
    if (!input.trim() || isLoading) return

    // Check if web search is enabled and should be triggered
    if (webSearchEnabled) {
      try {
        toast.loading('🔍 Researching on the web...', { duration: 2000 })
        
        // Trigger web research
        await apiClient.webResearch({
          query: input.trim(),
          search_engine: 'duckduckgo',
          max_results: 5
        })
        
        toast.success('✅ Web research completed and stored')
      } catch (error) {
        toast.error('❌ Web research failed')
        console.error('Web research error:', error)
      }
    }

    // Proceed with normal chat submission
    handleSubmit(e)
  }

  const placeholder = webSearchEnabled 
    ? "Ask a question - I'll research it on the web and answer..."
    : "Ask me anything about your documents..."

  return (
    <div className="border-t bg-card/50 backdrop-blur-sm">
      <div 
        {...getRootProps()}
        className={cn(
          "max-w-4xl mx-auto transition-all duration-200",
          isDragActive && "bg-primary/5 border-primary"
        )}
      >
        <input {...getInputProps()} />
        
        {/* File Upload Area (when dragging) */}
        {isDragActive && (
          <div className="p-4 border-2 border-dashed border-primary bg-primary/5 text-center">
            <Upload className="h-8 w-8 mx-auto mb-2 text-primary" />
            <p className="text-sm font-medium text-primary">Drop files here to upload</p>
            <p className="text-xs text-muted-foreground">Supports: PDF, DOCX, TXT, MD, HTML</p>
          </div>
        )}

        {/* Uploaded Files Display */}
        {showFileArea && uploadedFiles.length > 0 && (
          <div className="p-4 border-b bg-muted/20">
            <div className="flex items-center justify-between mb-3">
              <div className="flex items-center gap-2">
                <FileText className="h-4 w-4" />
                <span className="text-sm font-medium">Uploaded Files</span>
                <Badge variant="secondary">{uploadedFiles.length}</Badge>
              </div>
              <Button
                variant="ghost"
                size="sm"
                onClick={clearAllFiles}
                disabled={isUploading}
              >
                <X className="h-4 w-4" />
              </Button>
            </div>
            
            <div className="grid grid-cols-1 sm:grid-cols-2 gap-2">
              {uploadedFiles.map((uploadedFile, index) => (
                <div
                  key={index}
                  className="flex items-center gap-2 p-2 rounded-lg bg-background border"
                >
                  <div className="text-sm">
                    {getFileIcon(uploadedFile.file.name)}
                  </div>
                  
                  <div className="flex-1 min-w-0">
                    <div className="text-xs font-medium truncate">
                      {uploadedFile.file.name}
                    </div>
                    <div className="text-xs text-muted-foreground">
                      {formatFileSize(uploadedFile.file.size)}
                      {uploadedFile.status === 'success' && uploadedFile.result && (
                        <span className="text-green-600 ml-1">
                          • {uploadedFile.result.chunks_created} chunks
                        </span>
                      )}
                    </div>
                  </div>

                  <div className="flex items-center gap-1">
                    {uploadedFile.status === 'uploading' && (
                      <Loader2 className="h-3 w-3 animate-spin text-blue-600" />
                    )}
                    {uploadedFile.status === 'success' && (
                      <CheckCircle className="h-3 w-3 text-green-600" />
                    )}
                    {uploadedFile.status === 'error' && (
                      <X className="h-3 w-3 text-red-600" />
                    )}
                    
                    <Button
                      variant="ghost"
                      size="sm"
                      className="h-6 w-6 p-0"
                      onClick={() => removeFile(index)}
                      disabled={isUploading}
                    >
                      <X className="h-3 w-3" />
                    </Button>
                  </div>
                </div>
              ))}
            </div>
          </div>
        )}

        {/* Main Input Area */}
        <div className="p-4">
          {/* Options Bar */}
          <div className="flex items-center justify-between mb-3">
            <div className="flex items-center gap-4">
              {/* Web Search Toggle */}
              <div className="flex items-center gap-2">
                <Switch
                  id="web-search"
                  checked={webSearchEnabled}
                  onCheckedChange={setWebSearchEnabled}
                  disabled={isLoading}
                />
                <Label htmlFor="web-search" className="text-xs flex items-center gap-1">
                  <Search className="h-3 w-3" />
                  Web Research
                </Label>
              </div>
              
              {/* File Upload Button */}
              <Button
                variant="ghost"
                size="sm"
                onClick={triggerFileUpload}
                disabled={isUploading}
                className="text-xs"
              >
                <Paperclip className="h-3 w-3 mr-1" />
                Attach Files
              </Button>
            </div>

            {/* Status Indicators */}
            <div className="flex items-center gap-2">
              {webSearchEnabled && (
                <Badge variant="secondary" className="text-xs">
                  <Globe className="h-3 w-3 mr-1" />
                  Web Search On
                </Badge>
              )}
              
              {uploadedFiles.length > 0 && (
                <Badge variant="secondary" className="text-xs">
                  <FileText className="h-3 w-3 mr-1" />
                  {uploadedFiles.length} files
                </Badge>
              )}
            </div>
          </div>

          {/* Input Form */}
          <form onSubmit={onSubmit} className="flex gap-2">
            <div className="flex-1 relative">
              <Input
                ref={inputRef}
                value={input}
                onChange={handleInputChange}
                placeholder={placeholder}
                className="pr-12 min-h-[48px]"
                disabled={isLoading}
                autoFocus
              />
              <Button
                type="submit"
                size="sm"
                className="absolute right-1 top-1 h-10 w-10 p-0"
                disabled={!input.trim() || isLoading}
              >
                <Send className="h-4 w-4" />
              </Button>
            </div>
          </form>
          
          <div className="text-xs text-muted-foreground text-center mt-2 flex items-center justify-center gap-4">
            <span>Press Enter to send</span>
            <span>•</span>
            <span>Drag files to upload</span>
            <span>•</span>
            <span>Toggle web search for research</span>
          </div>
        </div>
      </div>
