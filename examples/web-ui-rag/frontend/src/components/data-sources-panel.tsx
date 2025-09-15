'use client'

import { useState, useCallback } from 'react'
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
  const { setCollectionUpdating } = useStatsStore()

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

  // GitHub repository indexing
  const handleGithubIndex = async () => {
    if (!githubUrl.trim()) {
      toast.error('Please enter a GitHub repository URL')
      return
    }

    setIsGithubIndexing(true)
    setCollectionUpdating('github_code', true)

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
      toast.success(`✅ Repository indexed: ${result.files_indexed} files processed`)
      setGithubUrl('')

    } catch (error) {
      console.error('GitHub indexing failed:', error)
      toast.error(`❌ Failed to index repository: ${error}`)
    } finally {
      setIsGithubIndexing(false)
      setCollectionUpdating('github_code', false)
    }
  }

  // Web research
  const handleWebResearch = async () => {
    if (!webSearchQuery.trim()) {
      toast.error('Please enter a search query')
      return
    }

    setIsWebSearching(true)
    setCollectionUpdating('web_research', true)

    try {
      const response = await fetch('http://localhost:8501/research', {
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

      const result = await response.json()
      toast.success(`✅ Web research complete: ${result.stored_count} results stored`)
      setWebSearchQuery('')


    } catch (error) {
      console.error('Web research failed:', error)
      toast.error(`❌ Web research failed: ${error}`)
    } finally {
      setIsWebSearching(false)
      setCollectionUpdating('web_research', false)
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
            </div>
          </TabsContent>
        </Tabs>
      </div>
    </div>
  )
}
