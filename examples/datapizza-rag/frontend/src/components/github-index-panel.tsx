'use client'

import { useState } from 'react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Github, Loader2, Star, CheckCircle, Code } from 'lucide-react'
import { apiClient } from '@/lib/api'
import { useGitHubStore } from '@/store'
import { isValidGitHubUrl, extractRepoFromGitHubUrl, getLanguageIcon } from '@/lib/utils'
import toast from 'react-hot-toast'

export function GitHubIndexPanel() {
  const [repoUrl, setRepoUrl] = useState('')
  const { isIndexing, results, setIndexing, setResults, setError } = useGitHubStore()

  const handleIndex = async () => {
    if (!repoUrl.trim() || isIndexing) return

    if (!isValidGitHubUrl(repoUrl.trim())) {
      toast.error('❌ Please enter a valid GitHub repository URL')
      return
    }

    setIndexing(true)
    setError(null)

    try {
      const result = await apiClient.indexGitHub({
        repository_url: repoUrl.trim(),
        max_files: 500
      })

      setResults(result)
      
      if (result.success) {
        toast.success(`✅ Indexed ${result.files_indexed} files from ${result.repository}`)
      } else {
        toast.error(`❌ Indexing failed: ${result.message}`)
      }
    } catch (error: any) {
      const errorMsg = error.response?.data?.error || error.message || 'GitHub indexing failed'
      setError(errorMsg)
      toast.error(`❌ ${errorMsg}`)
    } finally {
      setIndexing(false)
    }
  }

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    handleIndex()
  }

  const repoInfo = repoUrl.trim() ? extractRepoFromGitHubUrl(repoUrl.trim()) : null

  return (
    <div className="p-4 space-y-4">
      {/* Index Form */}
      <form onSubmit={handleSubmit} className="space-y-3">
        <div>
          <label className="text-sm font-medium mb-2 block">
            GitHub Repository URL
          </label>
          <Input
            value={repoUrl}
            onChange={(e) => setRepoUrl(e.target.value)}
            placeholder="https://github.com/owner/repo"
            disabled={isIndexing}
          />
          
          {repoInfo && (
            <div className="mt-2 text-xs text-muted-foreground">
              Repository: <span className="font-medium">{repoInfo.owner}/{repoInfo.repo}</span>
            </div>
          )}
        </div>

        <Button
          type="submit"
          className="w-full"
          disabled={!repoUrl.trim() || !isValidGitHubUrl(repoUrl.trim()) || isIndexing}
        >
          {isIndexing ? (
            <>
              <Loader2 className="h-4 w-4 mr-2 animate-spin" />
              Indexing Repository...
            </>
          ) : (
            <>
              <Github className="h-4 w-4 mr-2" />
              Index Repository
            </>
          )}
        </Button>
      </form>

      {/* Results */}
      {results && (
        <div className="space-y-3">
          <div className="flex items-center gap-2">
            <CheckCircle className="h-4 w-4 text-green-600" />
            <span className="text-sm font-medium">Indexing Complete</span>
          </div>

          <div className="bg-muted/50 rounded-lg p-3 space-y-2">
            <div className="text-sm">
              <span className="font-medium">Repository:</span> {results.repository}
            </div>
            <div className="text-sm">
              <span className="font-medium">Files:</span> {results.files_indexed} indexed, {results.files_stored} stored
            </div>
            <div className="text-sm flex items-center gap-1">
              <Star className="h-3 w-3" />
              <span className="font-medium">Stars:</span> {results.repository_stars}
            </div>
            <div className="text-sm">
              <span className="font-medium">Time:</span> {results.processing_time.toFixed(2)}s
            </div>
          </div>

          {results.languages && results.languages.length > 0 && (
            <div className="space-y-2">
              <h4 className="text-sm font-medium">Languages Found:</h4>
              <div className="flex flex-wrap gap-1">
                {results.languages.map((language, index) => (
                  <div
                    key={index}
                    className="inline-flex items-center gap-1 bg-muted/30 rounded px-2 py-1 text-xs"
                  >
                    <span>{getLanguageIcon(language)}</span>
                    <span>{language}</span>
                  </div>
                ))}
              </div>
            </div>
          )}
        </div>
      )}

      {/* Instructions */}
      <div className="text-xs text-muted-foreground space-y-1">
        <p>• Repository code is indexed and made searchable</p>
        <p>• Supports most programming languages</p>
        <p>• Ask questions about code structure and implementation</p>
        <p>• Large repositories may take several minutes to process</p>
      </div>
    </div>
  )
}
