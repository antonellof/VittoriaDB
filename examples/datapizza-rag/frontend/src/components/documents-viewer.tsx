'use client'

import { useState, useEffect } from 'react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { 
  Dialog, 
  DialogContent, 
  DialogHeader, 
  DialogTitle,
  DialogTrigger 
} from '@/components/ui/dialog'
import { 
  FileText, 
  Globe, 
  Code, 
  MessageSquare, 
  Search, 
  ExternalLink,
  Calendar,
  Hash,
  Loader2,
  X,
  Trash2
} from 'lucide-react'
import { apiClient } from '@/lib/api'
import { cn } from '@/lib/utils'
import toast from 'react-hot-toast'

interface DocumentChunk {
  id: string
  filename?: string
  title?: string
  url?: string
  content_preview: string
  metadata: {
    source?: string
    source_collection?: string
    timestamp?: number
    content_hash?: string
    [key: string]: any
  }
  score?: number
}

interface GroupedDocument {
  id: string
  filename?: string
  title?: string
  url?: string
  chunks: DocumentChunk[]
  totalChunks: number
  bestScore?: number
  latestTimestamp?: number
  metadata: {
    source?: string
    source_collection?: string
    [key: string]: any
  }
}

interface DocumentsViewerProps {
  collectionName: string
  collectionTitle: string
  icon?: React.ComponentType<{ className?: string }>
  trigger?: React.ReactNode
}

export function DocumentsViewer({ 
  collectionName, 
  collectionTitle, 
  icon: Icon = FileText,
  trigger 
}: DocumentsViewerProps) {
  const [isOpen, setIsOpen] = useState(false)
  const [documents, setDocuments] = useState<GroupedDocument[]>([])
  const [loading, setLoading] = useState(false)
  const [searchQuery, setSearchQuery] = useState('')
  const [filteredDocs, setFilteredDocs] = useState<GroupedDocument[]>([])
  const [deletingDocs, setDeletingDocs] = useState<Set<string>>(new Set())

  const fetchDocuments = async () => {
    setLoading(true)
    try {
      // Use the new original documents endpoint that groups by source file
      const response = await apiClient.getOriginalDocuments(collectionName)

      console.log('📋 Original documents response:', response)

      // Convert original documents to GroupedDocument format
      const groupedDocs: GroupedDocument[] = response.documents.map(doc => {
        // Convert chunks to DocumentChunk format
        const chunks: DocumentChunk[] = doc.chunks.map(chunk => ({
          id: chunk.chunk_id,
          filename: doc.filename,
          title: doc.title,
          url: doc.chunks.find(c => c.metadata.url)?.metadata.url, // Get URL from any chunk that has it
          content_preview: chunk.content_preview,
          metadata: chunk.metadata,
          score: chunk.score
        }))

        return {
          id: doc.document_id,
          filename: doc.filename,
          title: doc.title,
          url: doc.chunks.find(c => c.metadata.url)?.metadata.url,
          chunks: chunks,
          totalChunks: doc.total_chunks,
          bestScore: Math.max(...doc.chunks.map(c => c.score)),
          latestTimestamp: doc.upload_timestamp,
          metadata: {
            document_id: doc.document_id,
            filename: doc.filename,
            title: doc.title,
            file_type: doc.file_type,
            upload_timestamp: doc.upload_timestamp,
            content_hash: doc.content_hash,
            total_size: doc.total_size
          }
        }
      })

      // Sort by upload time (newest first)
      groupedDocs.sort((a, b) => (b.latestTimestamp || 0) - (a.latestTimestamp || 0))

      console.log('📋 Processed documents:', groupedDocs)

      setDocuments(groupedDocs)
      setFilteredDocs(groupedDocs)
    } catch (error: any) {
      console.error('Failed to fetch documents:', error)
      toast.error(`Failed to load ${collectionTitle.toLowerCase()}: ${error.message}`)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    if (isOpen) {
      fetchDocuments()
    }
  }, [isOpen, collectionName])

  useEffect(() => {
    if (!searchQuery) {
      setFilteredDocs(documents)
    } else {
      const filtered = documents.filter(doc => 
        (doc.filename?.toLowerCase().includes(searchQuery.toLowerCase())) ||
        (doc.title?.toLowerCase().includes(searchQuery.toLowerCase())) ||
        (doc.url?.toLowerCase().includes(searchQuery.toLowerCase())) ||
        doc.chunks.some(chunk => chunk.content_preview.toLowerCase().includes(searchQuery.toLowerCase()))
      )
      setFilteredDocs(filtered)
    }
  }, [searchQuery, documents])

  const getDocumentIcon = (doc: GroupedDocument) => {
    if (doc.url) return Globe
    if (doc.metadata.source_collection === 'github_code') return Code
    if (doc.metadata.source_collection === 'chat_history') return MessageSquare
    return FileText
  }

  const getDocumentTitle = (doc: GroupedDocument) => {
    return doc.title || doc.filename || doc.url || `Document ${doc.id.substring(0, 8)}`
  }


  const handleDeleteDocument = async (doc: GroupedDocument) => {
    if (!confirm(`Are you sure you want to delete "${getDocumentTitle(doc)}"? This will remove all ${doc.totalChunks} chunks permanently.`)) {
      return
    }

    setDeletingDocs(prev => new Set(prev).add(doc.id))

    try {
      // Delete all chunks belonging to this document
      let deletedCount = 0
      const errors: string[] = []

      for (const chunk of doc.chunks) {
        try {
          console.log(`🗑️ Deleting chunk: ${chunk.id}`)
          const result = await apiClient.deleteDocumentById(collectionName, chunk.id)
          
          if (result.success) {
            deletedCount++
          } else {
            errors.push(`Failed to delete chunk ${chunk.id}`)
          }
        } catch (error: any) {
          console.error(`Failed to delete chunk ${chunk.id}:`, error)
          errors.push(`Error deleting chunk ${chunk.id}: ${error.message}`)
        }
      }

      if (deletedCount > 0) {
        toast.success(`✅ Deleted ${deletedCount} chunks from "${getDocumentTitle(doc)}"`)
        
        // Refresh the documents list
        await fetchDocuments()
      }

      if (errors.length > 0) {
        console.error('Some deletions failed:', errors)
        toast.error(`⚠️ ${errors.length} chunks failed to delete`)
      }

    } catch (error: any) {
      console.error('Delete failed:', error)
      toast.error(`❌ Delete failed: ${error.message}`)
    } finally {
      setDeletingDocs(prev => {
        const newSet = new Set(prev)
        newSet.delete(doc.id)
        return newSet
      })
    }
  }

  const formatTimestamp = (timestamp?: number) => {
    if (!timestamp) return 'Unknown'
    return new Date(timestamp * 1000).toLocaleString()
  }

  return (
    <Dialog open={isOpen} onOpenChange={setIsOpen}>
      <DialogTrigger asChild>
        {trigger || (
          <Button variant="ghost" size="sm" className="justify-start">
            <Icon className="h-4 w-4 mr-2" />
            View {collectionTitle}
          </Button>
        )}
      </DialogTrigger>
      <DialogContent className="max-w-4xl max-h-[80vh] overflow-hidden flex flex-col">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <Icon className="h-5 w-5" />
            {collectionTitle} Collection
            {!loading && (
              <span className="text-sm font-normal text-muted-foreground">
                ({filteredDocs.length} {filteredDocs.length === 1 ? 'document' : 'documents'})
              </span>
            )}
          </DialogTitle>
        </DialogHeader>

        {/* Search */}
        <div className="flex gap-2 mb-4">
          <div className="relative flex-1">
            <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-muted-foreground" />
            <Input
              placeholder={`Search ${collectionTitle.toLowerCase()}...`}
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="pl-10"
            />
          </div>
          <Button
            onClick={fetchDocuments}
            variant="outline"
            size="sm"
            disabled={loading}
          >
            {loading ? (
              <Loader2 className="h-4 w-4 animate-spin" />
            ) : (
              'Refresh'
            )}
          </Button>
        </div>

        {/* Documents List */}
        <div className="flex-1 overflow-y-auto space-y-3">
          {loading ? (
            <div className="flex items-center justify-center py-8">
              <Loader2 className="h-6 w-6 animate-spin mr-2" />
              <span>Loading {collectionTitle.toLowerCase()}...</span>
            </div>
          ) : filteredDocs.length === 0 ? (
            <div className="text-center py-8 text-muted-foreground">
              {documents.length === 0 ? (
                <>
                  <Icon className="h-12 w-12 mx-auto mb-4 opacity-50" />
                  <p>No documents found in {collectionTitle.toLowerCase()}</p>
                </>
              ) : (
                <>
                  <Search className="h-12 w-12 mx-auto mb-4 opacity-50" />
                  <p>No documents match your search</p>
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={() => setSearchQuery('')}
                    className="mt-2"
                  >
                    <X className="h-4 w-4 mr-1" />
                    Clear search
                  </Button>
                </>
              )}
            </div>
          ) : (
            filteredDocs.map((doc) => {
              const DocIcon = getDocumentIcon(doc)
              
              return (
                <div
                  key={doc.id}
                  className="border rounded-lg p-4 hover:bg-muted/50 transition-colors"
                >
                  <div className="flex items-start gap-3">
                    <DocIcon className="h-5 w-5 mt-0.5 text-muted-foreground flex-shrink-0" />
                    
                    <div className="flex-1 min-w-0">
                      <div className="flex items-start justify-between gap-2">
                        <div className="flex-1 min-w-0">
                          <div className="flex items-center gap-2">
                            <h3 className="font-medium truncate">
                              {getDocumentTitle(doc)}
                            </h3>
                            
                            <span className="text-xs bg-blue-100 text-blue-700 px-2 py-0.5 rounded-full">
                              {doc.totalChunks} {doc.totalChunks === 1 ? 'part' : 'parts'}
                            </span>
                          </div>
                          
                          {doc.url && (
                            <a
                              href={doc.url}
                              target="_blank"
                              rel="noopener noreferrer"
                              className="text-sm text-blue-600 hover:underline flex items-center gap-1 mt-1"
                            >
                              <ExternalLink className="h-3 w-3" />
                              {doc.url}
                            </a>
                          )}
                        </div>
                        
                        <div className="flex items-center gap-2">
                          <Button
                            variant="ghost"
                            size="sm"
                            className="h-8 w-8 p-0 text-red-600 hover:text-red-700 hover:bg-red-50"
                            onClick={(e) => {
                              e.stopPropagation()
                              handleDeleteDocument(doc)
                            }}
                            disabled={deletingDocs.has(doc.id)}
                            title={`Delete ${getDocumentTitle(doc)}`}
                          >
                            {deletingDocs.has(doc.id) ? (
                              <Loader2 className="h-4 w-4 animate-spin" />
                            ) : (
                              <Trash2 className="h-4 w-4" />
                            )}
                          </Button>
                        </div>
                      </div>

                      {/* Show preview from the first chunk */}
                      <p className="text-sm text-muted-foreground mt-2 line-clamp-2">
                        {doc.chunks[0]?.content_preview || 'No content preview available'}
                      </p>

                      {/* Metadata */}
                      <div className="flex flex-wrap gap-3 mt-3 text-xs text-muted-foreground">
                        {doc.latestTimestamp && (
                          <div className="flex items-center gap-1">
                            <Calendar className="h-3 w-3" />
                            {formatTimestamp(doc.latestTimestamp)}
                          </div>
                        )}
                        
                        {doc.metadata.content_hash && (
                          <div className="flex items-center gap-1">
                            <Hash className="h-3 w-3" />
                            {doc.metadata.content_hash.substring(0, 8)}...
                          </div>
                        )}
                        
                        {doc.metadata.source && (
                          <div className="bg-muted px-2 py-0.5 rounded">
                            {doc.metadata.source}
                          </div>
                        )}
                      </div>

                    </div>
                  </div>
                </div>
              )
            })
          )}
        </div>
      </DialogContent>
    </Dialog>
  )
}
