'use client'

import { useState } from 'react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Search, Loader2, ExternalLink, CheckCircle, X } from 'lucide-react'
import { apiClient } from '@/lib/api'
import { useResearchStore, useChatStore } from '@/store'
import { cn } from '@/lib/utils'
import toast from 'react-hot-toast'

export function WebResearchPanel() {
  const [query, setQuery] = useState('')
  const { 
    isResearching, 
    results, 
    progress, 
    currentStep, 
    foundResults,
    sendWebResearch, 
    setError,
    clearProgress,
    setResearching 
  } = useResearchStore()
  
  const { isConnected } = useChatStore()

  const handleResearch = () => {
    if (!query.trim()) return
    
    if (isResearching) {
      // Stop the research
      handleStopResearch()
      return
    }

    if (!isConnected) {
      toast.error('❌ Not connected to server. Please wait for connection.')
      return
    }

    clearProgress()
    setError(null)
    sendWebResearch(query.trim())
  }

  const handleStopResearch = () => {
    setResearching(false)
    clearProgress()
    toast.success('🛑 Research stopped')
  }

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    handleResearch()
  }

  return (
    <div className="p-4 space-y-4">
      {/* Search Form */}
      <form onSubmit={handleSubmit} className="space-y-3">
        <div>
          <label className="text-sm font-medium mb-2 block">
            Research Query
          </label>
          <Input
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            placeholder="Enter topic to research..."
            disabled={isResearching}
          />
        </div>

        <Button
          type="submit"
          className={cn(
            "w-full",
            isResearching && "bg-red-600 hover:bg-red-700"
          )}
          disabled={!query.trim() || !isConnected}
        >
          {isResearching ? (
            <>
              <X className="h-4 w-4 mr-2" />
              Stop Research
            </>
          ) : (
            <>
              <Search className="h-4 w-4 mr-2" />
              Start Research
            </>
          )}
        </Button>
        
        {!isConnected && (
          <p className="text-xs text-red-500 text-center">
            ⚠️ Connecting to server...
          </p>
        )}
      </form>

      {/* Live Progress - Similar to Chat Streaming */}
      {isResearching && (
        <div className="space-y-3">
          <div className="flex items-center gap-2">
            <Loader2 className="h-4 w-4 animate-spin text-blue-600" />
            <span className="text-sm font-medium">Web Research in Progress</span>
          </div>

          <div className="bg-muted/50 rounded-lg p-3 space-y-3">
            {/* Current Status */}
            <div className="flex items-center gap-2 text-sm">
              <div className="w-2 h-2 bg-blue-600 rounded-full animate-pulse" />
              <span className="font-medium">Status:</span> 
              <span className="text-blue-600">{currentStep || 'Initializing...'}</span>
            </div>
            
            {/* Progress Bar */}
            {progress > 0 && (
              <div className="space-y-2">
                <div className="flex justify-between text-xs">
                  <span className="font-medium">Progress</span>
                  <span className="font-mono">{progress}%</span>
                </div>
                <div className="w-full bg-muted rounded-full h-2.5 overflow-hidden">
                  <div 
                    className="bg-gradient-to-r from-blue-500 to-blue-600 h-full rounded-full transition-all duration-500 ease-out"
                    style={{ width: `${progress}%` }}
                  />
                </div>
              </div>
            )}

            {/* Found Results Preview */}
            {foundResults.length > 0 && (
              <div className="space-y-2">
                <div className="flex items-center gap-2">
                  <Search className="h-3 w-3 text-green-600" />
                  <h4 className="text-sm font-medium text-green-600">
                    Found {foundResults.length} Results
                  </h4>
                </div>
                <div className="space-y-2 max-h-40 overflow-y-auto">
                  {foundResults.map((result, index) => (
                    <div
                      key={index}
                      className="bg-muted/30 rounded-md p-2 text-xs space-y-1 border-l-2 border-blue-500"
                    >
                      <div className="font-medium truncate text-foreground">
                        {result.title || `Result ${index + 1}`}
                      </div>
                      {result.url && (
                        <div className="flex items-center gap-1 text-blue-600">
                          <ExternalLink className="h-3 w-3" />
                          <span className="truncate text-xs">{result.url}</span>
                        </div>
                      )}
                      {result.status && (
                        <div className="text-xs text-muted-foreground">
                          Status: {result.status}
                        </div>
                      )}
                    </div>
                  ))}
                </div>
              </div>
            )}

            {/* Research Steps (like chat reasoning) */}
            <div className="text-xs text-muted-foreground space-y-1">
              <p>🔍 Searching web for relevant information...</p>
              <p>🕷️ Scraping and processing content...</p>
              <p>💾 Storing results in knowledge base...</p>
            </div>
          </div>
        </div>
      )}

      {/* Results */}
      {!isResearching && results && (
        <div className="space-y-3">
          <div className="flex items-center gap-2">
            <CheckCircle className="h-4 w-4 text-green-600" />
            <span className="text-sm font-medium">Research Complete</span>
          </div>

          <div className="bg-muted/50 rounded-lg p-3 space-y-2">
            <div className="text-sm">
              <span className="font-medium">Query:</span> {results.query}
            </div>
            <div className="text-sm">
              <span className="font-medium">Results:</span> {results.results_count} found, {results.stored_count} stored
            </div>
            <div className="text-sm">
              <span className="font-medium">Time:</span> {results.processing_time.toFixed(2)}s
            </div>
          </div>

          {results.results && results.results.length > 0 && (
            <div className="space-y-2">
              <h4 className="text-sm font-medium">Sources Found:</h4>
              <div className="space-y-2 max-h-40 overflow-y-auto">
                {results.results.map((result, index) => (
                  <div
                    key={index}
                    className="bg-muted/30 rounded p-2 text-xs space-y-1"
                  >
                    <div className="font-medium truncate">
                      {result.title}
                    </div>
                    <div className="text-muted-foreground truncate">
                      {result.snippet}
                    </div>
                    <div className="flex items-center gap-1 text-blue-600">
                      <ExternalLink className="h-3 w-3" />
                      <span className="truncate">{result.url}</span>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          )}
        </div>
      )}

      {/* Instructions */}
      <div className="text-xs text-muted-foreground space-y-1">
        <p>• Research results are automatically stored in your knowledge base</p>
        <p>• Content is scraped and processed for search</p>
        <p>• Ask questions about researched topics in chat</p>
      </div>
    </div>
  )
}
