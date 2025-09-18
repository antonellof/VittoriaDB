'use client'

import { useState, useCallback, useEffect } from 'react'
import { useDropzone } from 'react-dropzone'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { 
  Upload, 
  Github, 
  Globe, 
  FileText, 
  Code, 
  Loader2, 
  CheckCircle, 
  X,
  AlertCircle,
  Database,
  Plus
} from 'lucide-react'
import { cn, getFileIcon, formatFileSize } from '@/lib/utils'
import { useStatsStore } from '@/store'
import { useWebSocketNotifications } from '@/hooks/useWebSocketNotifications'
import toast from 'react-hot-toast'

interface DataSourcesPanelProps {
  onClose?: () => void
}

interface AttachedFile {
  file: File
  id: string
  result?: any
  status?: 'uploading' | 'success' | 'error'
}

interface GitHubRepo {
  url: string
  status?: 'indexing' | 'success' | 'error'
  result?: any
}

export function DataSourcesPanel({ onClose }: DataSourcesPanelProps) {
  const [attachedFiles, setAttachedFiles] = useState<AttachedFile[]>([])
  const [githubUrl, setGithubUrl] = useState('')
  const [isGithubIndexing, setIsGithubIndexing] = useState(false)
  const [webSearchQuery, setWebSearchQuery] = useState('')
  const [isWebSearching, setIsWebSearching] = useState(false)
  const [streamingResults, setStreamingResults] = useState<Array<{
    index: number
    title: string
    url: string
    content_preview?: string
    status: 'found' | 'scraped' | 'storing' | 'stored' | 'error'
    doc_id?: string
    error_message?: string
  }>>([])
  const [streamingStats, setStreamingStats] = useState({
    query: '',
    urls_found: 0,
    urls_scraped: 0,
    results_stored: 0,
    processing_time: 0,
    message: ''
  })
  const { setCollectionUpdating } = useStatsStore()
  const { webResearch } = useWebSocketNotifications()

  // Update web searching state based on notifications
  useEffect(() => {
    const activeResearch = webResearch.find(research => research.status === 'researching')
    setIsWebSearching(!!activeResearch)
  }, [webResearch])

  // File upload handling
  const onDrop = useCallback(async (acceptedFiles: File[]) => {
    // Filter out image files
    const documentFiles = acceptedFiles.filter(file => {
      const isImage = file.type.startsWith('image/')
      if (isImage) {
        toast.error(`Skipped ${file.name} - images are not supported`)
      }
      return !isImage
    })

    if (documentFiles.length === 0) return

    // Add files with uploading status
    const newFiles: AttachedFile[] = documentFiles.map(file => ({
      file,
      id: `${file.name}-${Date.now()}`,
      status: 'uploading'
    }))
    
    setAttachedFiles(prev => [...prev, ...newFiles])

    // Upload files one by one
    for (const attachedFile of newFiles) {
      try {
        setCollectionUpdating('documents', true)
        
        const formData = new FormData()
        formData.append('file', attachedFile.file)
        
        const response = await fetch('http://localhost:8501/upload', {
          method: 'POST',
          body: formData,
        })
        
        if (!response.ok) {
          throw new Error(`Upload failed: ${response.statusText}`)
        }
        
        const result = await response.json()
        
        // Update file status
        setAttachedFiles(prev => prev.map(f => 
          f.id === attachedFile.id 
            ? { ...f, status: 'success', result }
            : f
        ))
        
        toast.success(`✅ ${attachedFile.file.name} uploaded successfully`)
        
      } catch (error) {
        console.error('Upload failed:', error)
        setAttachedFiles(prev => prev.map(f => 
          f.id === attachedFile.id 
            ? { ...f, status: 'error' }
            : f
        ))
        toast.error(`❌ Failed to upload ${attachedFile.file.name}`)
      } finally {
        setCollectionUpdating('documents', false)
      }
    }

  }, [setCollectionUpdating])

  const { getRootProps, getInputProps, isDragActive } = useDropzone({
    onDrop,
    accept: {
      'text/plain': ['.txt'],
      'text/markdown': ['.md'],
      'text/html': ['.html'],
      'application/pdf': ['.pdf'],
      'application/vnd.openxmlformats-officedocument.wordprocessingml.document': ['.docx'],
      'application/msword': ['.doc'],
      'text/csv': ['.csv'],
      'application/json': ['.json'],
    },
    multiple: true
  })

  // GitHub repository indexing (now background)
  const handleGithubIndex = async () => {
    if (!githubUrl.trim()) {
      toast.error('Please enter a GitHub repository URL')
      return
    }

    setIsGithubIndexing(true)

    try {
      const response = await fetch('http://localhost:8501/github/index', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          repository_url: githubUrl.trim()
        }),
      })

      if (!response.ok) {
        throw new Error(`GitHub indexing failed: ${response.statusText}`)
      }

      const result = await response.json()
      toast.success(`🚀 ${result.message}`)
      setGithubUrl('')

    } catch (error) {
      console.error('GitHub indexing failed:', error)
      toast.error(`❌ Failed to start GitHub indexing: ${error}`)
    } finally {
      setIsGithubIndexing(false)
    }
  }

  // Web research with streaming
  const handleWebResearch = async () => {
    if (!webSearchQuery.trim()) {
      toast.error('Please enter a search query')
      return
    }

    setIsWebSearching(true)
    setStreamingResults([])
    setStreamingStats({
      query: webSearchQuery.trim(),
      urls_found: 0,
      urls_scraped: 0,
      results_stored: 0,
      processing_time: 0,
      message: 'Starting web research...'
    })

    try {
      const response = await fetch('http://localhost:8501/research/stream', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          query: webSearchQuery.trim(),
          search_engine: 'duckduckgo'
        }),
      })

      if (!response.ok) {
        throw new Error(`Web research failed: ${response.statusText}`)
      }

      const reader = response.body?.getReader()
      if (!reader) {
        throw new Error('No response body reader')
      }

      const decoder = new TextDecoder()

      try {
        while (true) {
          const { done, value } = await reader.read()
          if (done) break

          const chunk = decoder.decode(value)
          const lines = chunk.split('\n')

          for (const line of lines) {
            if (line.startsWith('data: ')) {
              try {
                const data = JSON.parse(line.slice(6))

                switch (data.type) {
                  case 'start':
                    setStreamingStats(prev => ({
                      ...prev,
                      message: data.message
                    }))
                    break

                  case 'results_found':
                    setStreamingStats(prev => ({
                      ...prev,
                      urls_found: data.count,
                      message: data.message
                    }))
                    break

                  case 'result_detail':
                    setStreamingResults(prev => {
                      const newResults = [...prev]
                      const existingIndex = newResults.findIndex(r => r.index === data.index)

                      if (existingIndex >= 0) {
                        newResults[existingIndex] = {
                          ...newResults[existingIndex],
                          ...data,
                          status: data.status
                        }
                      } else {
                        newResults.push({
                          index: data.index,
                          title: data.title,
                          url: data.url,
                          content_preview: data.content_preview,
                          status: data.status
                        })
                      }

                      return newResults.sort((a, b) => a.index - b.index)
                    })
                    break

                  case 'storing':
                    setStreamingResults(prev => 
                      prev.map(result => 
                        result.index === data.index 
                          ? { ...result, status: 'storing' }
                          : result
                      )
                    )
                    setStreamingStats(prev => ({
                      ...prev,
                      message: data.message
                    }))
                    break

                  case 'result_stored':
                    setStreamingResults(prev => 
                      prev.map(result => 
                        result.index === data.index 
                          ? { ...result, status: 'stored', doc_id: data.doc_id }
                          : result
                      )
                    )
                    setStreamingStats(prev => ({
                      ...prev,
                      results_stored: prev.results_stored + 1
                    }))
                    break

                  case 'storage_error':
                    setStreamingResults(prev => 
                      prev.map(result => 
                        result.index === data.index 
                          ? { ...result, status: 'error', error_message: data.message }
                          : result
                      )
                    )
                    break

                  case 'complete':
                    setStreamingStats(prev => ({
                      ...prev,
                      urls_found: data.urls_found,
                      urls_scraped: data.urls_scraped,
                      results_stored: data.results_stored,
                      processing_time: data.processing_time,
                      message: data.message
                    }))
                    break

                  case 'error':
                    setStreamingStats(prev => ({
                      ...prev,
                      message: `Error: ${data.message}`,
                      processing_time: data.processing_time || 0
                    }))
                    break
                }
              } catch (parseError) {
                console.error('Failed to parse streaming data:', parseError)
              }
            }
          }
        }
      } finally {
        reader.releaseLock()
      }

      console.log('✅ Web research streaming completed')
      setWebSearchQuery('')

    } catch (error) {
      console.error('Web research failed:', error)
      toast.error(`❌ Web research failed: ${error}`)
      setStreamingStats(prev => ({
        ...prev,
        message: `Error: ${error instanceof Error ? error.message : 'Unknown error'}`
      }))
    } finally {
      setIsWebSearching(false)
    }
  }

  const removeFile = (fileId: string) => {
    setAttachedFiles(prev => prev.filter(f => f.id !== fileId))
  }

  return (
    <div className="w-full max-w-4xl mx-auto bg-card rounded-lg">
      <div className="p-6 pb-0">
        <div className="flex items-center justify-between">
          <div>
            <h3 className="text-lg font-semibold flex items-center gap-2">
              <Database className="w-5 h-5" />
              Data Sources
            </h3>
            <p className="text-sm text-muted-foreground mt-1">
              Add documents, code repositories, and web research to your knowledge base
            </p>
          </div>

        </div>
      </div>
      <div className="p-6 pt-4">
        <Tabs defaultValue="files" className="w-full">
          <TabsList className="grid w-full grid-cols-3">
            <TabsTrigger value="files" className="flex items-center gap-2">
              <FileText className="w-4 h-4" />
              Documents
            </TabsTrigger>
            <TabsTrigger value="github" className="flex items-center gap-2">
              <Github className="w-4 h-4" />
              GitHub
            </TabsTrigger>
            <TabsTrigger value="web" className="flex items-center gap-2">
              <Globe className="w-4 h-4" />
              Web Research
            </TabsTrigger>
          </TabsList>

          <TabsContent value="files" className="space-y-4">
            <div
              {...getRootProps()}
              className={cn(
                "border-2 border-dashed rounded-lg p-8 text-center cursor-pointer transition-colors",
                isDragActive 
                  ? "border-primary bg-primary/5" 
                  : "border-muted-foreground/25 hover:border-primary/50"
              )}
            >
              <input {...getInputProps()} />
              <Upload className="w-12 h-12 mx-auto mb-4 text-muted-foreground" />
              <p className="text-lg font-medium mb-2">
                {isDragActive ? "Drop files here..." : "Upload Documents"}
              </p>
              <p className="text-sm text-muted-foreground mb-4">
                Drag & drop files here, or click to select
              </p>
              <p className="text-xs text-muted-foreground">
                Supported: PDF, DOCX, TXT, MD, HTML, CSV, JSON
              </p>
            </div>

            {attachedFiles.length > 0 && (
              <div className="space-y-2">
                <Label className="text-sm font-medium">Uploaded Files</Label>
                {attachedFiles.map((attachedFile) => (
                  <div
                    key={attachedFile.id}
                    className="flex items-center gap-3 p-3 border rounded-lg"
                  >
                    {getFileIcon(attachedFile.file.name)}
                    <div className="flex-1 min-w-0">
                      <p className="text-sm font-medium truncate">
                        {attachedFile.file.name}
                      </p>
                      <p className="text-xs text-muted-foreground">
                        {formatFileSize(attachedFile.file.size)}
                      </p>
                    </div>
                    <div className="flex items-center gap-2">
                      {attachedFile.status === 'uploading' && (
                        <Loader2 className="w-4 h-4 animate-spin text-blue-500" />
                      )}
                      {attachedFile.status === 'success' && (
                        <CheckCircle className="w-4 h-4 text-green-500" />
                      )}
                      {attachedFile.status === 'error' && (
                        <AlertCircle className="w-4 h-4 text-red-500" />
                      )}
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => removeFile(attachedFile.id)}
                      >
                        <X className="w-4 h-4" />
                      </Button>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </TabsContent>

          <TabsContent value="github" className="space-y-4">
            <div className="space-y-4">
              <div>
                <Label htmlFor="github-url">Repository URL</Label>
                <div className="flex gap-2 mt-1">
                  <Input
                    id="github-url"
                    placeholder="https://github.com/owner/repository"
                    value={githubUrl}
                    onChange={(e) => setGithubUrl(e.target.value)}
                    disabled={isGithubIndexing}
                  />
                  <Button 
                    onClick={handleGithubIndex}
                    disabled={isGithubIndexing || !githubUrl.trim()}
                  >
                    {isGithubIndexing ? (
                      <Loader2 className="w-4 h-4 animate-spin" />
                    ) : (
                      <Plus className="w-4 h-4" />
                    )}
                    {isGithubIndexing ? 'Indexing...' : 'Index'}
                  </Button>
                </div>
              </div>
              <div className="p-4 bg-muted/50 rounded-lg">
                <div className="flex items-start gap-3">
                  <Code className="w-5 h-5 text-blue-500 mt-0.5" />
                  <div>
                    <h4 className="font-medium mb-1">GitHub Repository Indexing</h4>
                    <p className="text-sm text-muted-foreground mb-2">
                      Index code repositories to search through source code, documentation, and README files.
                    </p>
                    <ul className="text-xs text-muted-foreground space-y-1">
                      <li>• Supports public repositories</li>
                      <li>• Indexes code files, documentation, and README</li>
                      <li>• Enables semantic code search</li>
                    </ul>
                  </div>
                </div>
              </div>
            </div>
          </TabsContent>

          <TabsContent value="web" className="space-y-4">
            <div className="space-y-4">
              <div>
                <Label htmlFor="web-query">Search Query</Label>
                <div className="flex gap-2 mt-1">
                  <Input
                    id="web-query"
                    placeholder="Enter search terms..."
                    value={webSearchQuery}
                    onChange={(e) => setWebSearchQuery(e.target.value)}
                    disabled={isWebSearching}
                  />
                  <Button 
                    onClick={handleWebResearch}
                    disabled={isWebSearching || !webSearchQuery.trim()}
                  >
                    {isWebSearching ? (
                      <Loader2 className="w-4 h-4 animate-spin" />
                    ) : (
                      <Plus className="w-4 h-4" />
                    )}
                    {isWebSearching ? 'Searching...' : 'Research'}
                  </Button>
                </div>
              </div>
              <div className="p-4 bg-muted/50 rounded-lg">
                <div className="flex items-start gap-3">
                  <Globe className="w-5 h-5 text-green-500 mt-0.5" />
                  <div>
                    <h4 className="font-medium mb-1">Web Research</h4>
                    <p className="text-sm text-muted-foreground mb-2">
                      Search the web for current information and store results in your knowledge base.
                    </p>
                    <ul className="text-xs text-muted-foreground space-y-1">
                      <li>• Searches current web content</li>
                      <li>• Stores articles and web pages</li>
                      <li>• Enables up-to-date information retrieval</li>
                    </ul>
                  </div>
                </div>
              </div>

              {/* Web Research Results */}
              {webResearch.length > 0 && (
                <div className="space-y-3">
                  <h4 className="font-medium text-sm">Research Progress</h4>
                  {webResearch.map((research) => (
                    <div key={research.research_id} className="border rounded-lg p-3 space-y-2">
                      <div className="flex items-center justify-between">
                        <div className="flex items-center gap-2">
                          <div className={cn(
                            "w-2 h-2 rounded-full",
                            research.status === 'researching' ? "bg-blue-500 animate-pulse" :
                            research.status === 'completed' ? "bg-green-500" :
                            "bg-red-500"
                          )} />
                          <span className="font-medium text-sm truncate max-w-[200px]">
                            {research.query}
                          </span>
                        </div>
                        <span className="text-xs text-muted-foreground">
                          {research.status === 'researching' ? `${research.progress || 0}%` : research.status}
                        </span>
                      </div>
                      
                      {research.message && (
                        <p className="text-xs text-muted-foreground">{research.message}</p>
                      )}
                      
                      {(research.urls_found || research.urls_scraped || research.results_stored) && (
                        <div className="flex gap-4 text-xs text-muted-foreground">
                          {research.urls_found && <span>Found: {research.urls_found}</span>}
                          {research.urls_scraped && <span>Scraped: {research.urls_scraped}</span>}
                          {research.results_stored && <span>Stored: {research.results_stored}</span>}
                        </div>
                      )}
                      
                      {research.results && research.results.length > 0 && (
                        <div className="space-y-2">
                          <p className="text-xs font-medium">Search Results:</p>
                          <div className="space-y-2 max-h-48 overflow-y-auto">
                            {research.results.map((result, index) => (
                              <div key={index} className="border rounded p-2 space-y-1 bg-muted/30">
                                <div className="flex items-start gap-2">
                                  <div className={cn(
                                    "w-2 h-2 rounded-full mt-1 flex-shrink-0",
                                    result.status === 'found' ? "bg-yellow-500" :
                                    result.status === 'scraping' ? "bg-blue-500 animate-pulse" :
                                    result.status === 'scraped' ? "bg-blue-600" :
                                    result.status === 'stored' ? "bg-green-500" :
                                    "bg-red-500"
                                  )} />
                                  <div className="flex-1 min-w-0">
                                    <div className="flex items-center gap-2 mb-1">
                                      <span className="font-medium text-xs truncate" title={result.title}>
                                        {result.title}
                                      </span>
                                      <span className={cn(
                                        "text-xs px-1.5 py-0.5 rounded text-white flex-shrink-0",
                                        result.status === 'found' ? "bg-yellow-600" :
                                        result.status === 'scraping' ? "bg-blue-600" :
                                        result.status === 'scraped' ? "bg-blue-700" :
                                        result.status === 'stored' ? "bg-green-600" :
                                        "bg-red-600"
                                      )}>
                                        {result.status}
                                      </span>
                                    </div>
                                    <a 
                                      href={result.url} 
                                      target="_blank" 
                                      rel="noopener noreferrer"
                                      className="text-xs text-blue-600 hover:text-blue-800 underline truncate block"
                                      title={result.url}
                                    >
                                      {result.url}
                                    </a>
                                    {result.content_preview && (
                                      <p className="text-xs text-muted-foreground mt-1 line-clamp-2">
                                        {result.content_preview}
                                      </p>
                                    )}
                                  </div>
                                </div>
                              </div>
                            ))}
                          </div>
                        </div>
                      )}
                    </div>
                  ))}
                </div>
              )}
            </div>
          </TabsContent>
        </Tabs>
      </div>
    </div>
  )
}
