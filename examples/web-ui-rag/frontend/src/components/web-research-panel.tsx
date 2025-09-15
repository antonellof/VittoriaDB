'use client'

import { useState } from 'react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Search, Loader2, ExternalLink, CheckCircle } from 'lucide-react'
import { apiClient } from '@/lib/api'
import { useResearchStore } from '@/store'
import { cn } from '@/lib/utils'
import toast from 'react-hot-toast'

export function WebResearchPanel() {
  const [query, setQuery] = useState('')
  const { isResearching, results, setResearching, setResults, setError } = useResearchStore()

  const handleResearch = async () => {
    if (!query.trim() || isResearching) return

    setResearching(true)
    setError(null)

    try {
      const result = await apiClient.webResearch({
        query: query.trim(),
        search_engine: 'duckduckgo',
        max_results: 5
      })

      setResults(result)
      
      if (result.success) {
        toast.success(`✅ Found ${result.results_count} results and stored ${result.stored_count} in knowledge base`)
      } else {
        toast.error(`❌ Research failed: ${result.message}`)
      }
    } catch (error: any) {
      const errorMsg = error.response?.data?.error || error.message || 'Research failed'
      setError(errorMsg)
      toast.error(`❌ ${errorMsg}`)
    } finally {
      setResearching(false)
    }
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
          className="w-full"
          disabled={!query.trim() || isResearching}
        >
          {isResearching ? (
            <>
              <Loader2 className="h-4 w-4 mr-2 animate-spin" />
              Researching...
            </>
          ) : (
            <>
              <Search className="h-4 w-4 mr-2" />
              Start Research
            </>
          )}
        </Button>
      </form>

      {/* Results */}
      {results && (
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
