'use client'

import { useState, useCallback } from 'react'
import { useDropzone } from 'react-dropzone'
import { Button } from '@/components/ui/button'
import { Progress } from '@/components/ui/progress'
import { Upload, File, CheckCircle, XCircle, Loader2 } from 'lucide-react'
import { apiClient, uploadFiles } from '@/lib/api'
import { useUploadStore } from '@/store'
import { cn, getFileIcon, formatFileSize } from '@/lib/utils'
import toast from 'react-hot-toast'

interface UploadedFile {
  file: File
  status: 'uploading' | 'success' | 'error'
  result?: any
  error?: string
}

export function FileUploadPanel() {
  const [uploadedFiles, setUploadedFiles] = useState<UploadedFile[]>([])
  const { isUploading, setUploading } = useUploadStore()

  const onDrop = useCallback(async (acceptedFiles: File[]) => {
    if (acceptedFiles.length === 0) return

    setUploading(true)
    
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
          
          toast.success(`✅ ${file.name} uploaded successfully`)
        } catch (error: any) {
          setUploadedFiles(prev => 
            prev.map(f => 
              f.file === file 
                ? { ...f, status: 'error', error: error.message }
                : f
            )
          )
          
          toast.error(`❌ Failed to upload ${file.name}`)
        }
      }
    } finally {
      setUploading(false)
    }
  }, [setUploading])

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
    disabled: isUploading
  })

  const clearFiles = () => {
    setUploadedFiles([])
  }

  return (
    <div className="p-4 space-y-4">
      {/* Drop Zone */}
      <div
        {...getRootProps()}
        className={cn(
          "border-2 border-dashed rounded-lg p-6 text-center cursor-pointer transition-colors",
          isDragActive ? "border-primary bg-primary/5" : "border-muted-foreground/25",
          isUploading && "opacity-50 cursor-not-allowed"
        )}
      >
        <input {...getInputProps()} />
        
        <Upload className="h-8 w-8 mx-auto mb-2 text-muted-foreground" />
        
        {isDragActive ? (
          <p className="text-sm">Drop files here...</p>
        ) : (
          <div>
            <p className="text-sm font-medium mb-1">
              Click or drag files to upload
            </p>
            <p className="text-xs text-muted-foreground">
              Supports: PDF, DOCX, DOC, TXT, MD, HTML
            </p>
          </div>
        )}
      </div>

      {/* Upload Progress */}
      {uploadedFiles.length > 0 && (
        <div className="space-y-3">
          <div className="flex items-center justify-between">
            <h4 className="font-medium text-sm">Uploaded Files</h4>
            <Button
              variant="ghost"
              size="sm"
              onClick={clearFiles}
              disabled={isUploading}
            >
              Clear
            </Button>
          </div>

          <div className="space-y-2 max-h-60 overflow-y-auto">
            {uploadedFiles.map((uploadedFile, index) => (
              <div
                key={index}
                className="flex items-center gap-3 p-2 rounded-lg bg-muted/50"
              >
                <div className="text-lg">
                  {getFileIcon(uploadedFile.file.name)}
                </div>
                
                <div className="flex-1 min-w-0">
                  <div className="text-sm font-medium truncate">
                    {uploadedFile.file.name}
                  </div>
                  <div className="text-xs text-muted-foreground">
                    {formatFileSize(uploadedFile.file.size)}
                  </div>
                  
                  {uploadedFile.status === 'success' && uploadedFile.result && (
                    <div className="text-xs text-green-600">
                      {uploadedFile.result.chunks_created} chunks created
                    </div>
                  )}
                  
                  {uploadedFile.status === 'error' && (
                    <div className="text-xs text-red-600">
                      {uploadedFile.error}
                    </div>
                  )}
                </div>

                <div className="flex-shrink-0">
                  {uploadedFile.status === 'uploading' && (
                    <Loader2 className="h-4 w-4 animate-spin text-blue-600" />
                  )}
                  {uploadedFile.status === 'success' && (
                    <CheckCircle className="h-4 w-4 text-green-600" />
                  )}
                  {uploadedFile.status === 'error' && (
                    <XCircle className="h-4 w-4 text-red-600" />
                  )}
                </div>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Instructions */}
      <div className="text-xs text-muted-foreground space-y-1">
        <p>• Files are automatically processed and indexed</p>
        <p>• Text content is chunked for optimal search</p>
        <p>• Embeddings are generated using Ollama</p>
      </div>
    </div>
  )
}
