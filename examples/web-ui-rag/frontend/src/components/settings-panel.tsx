'use client'

import { useState, useEffect } from 'react'
import { useTheme } from 'next-themes'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Separator } from '@/components/ui/separator'
import { Switch } from '@/components/ui/switch'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Badge } from '@/components/ui/badge'
import { 
  Settings, 
  Key, 
  Bot, 
  Database, 
  Globe, 
  Github,
  Save,
  Eye,
  EyeOff,
  TestTube,
  CheckCircle,
  XCircle,
  AlertCircle,
  Moon,
  Sun,
  Monitor
} from 'lucide-react'
import { apiClient } from '@/lib/api'
import { useStatsStore } from '@/store'
import { cn } from '@/lib/utils'
import toast from 'react-hot-toast'

interface SettingsConfig {
  openai_api_key: string
  ollama_url: string
  github_token: string
  vittoriadb_url: string
  default_model: string
  search_limit: number
  chunk_size: number
  chunk_overlap: number
  enable_web_research: boolean
  enable_github_indexing: boolean
}

export function SettingsPanel() {
  const { theme, setTheme } = useTheme()
  
  const [config, setConfig] = useState<SettingsConfig>({
    openai_api_key: '',
    ollama_url: 'http://localhost:11434',
    github_token: '',
    vittoriadb_url: 'http://localhost:8080',
    default_model: 'gpt-4',
    search_limit: 5,
    chunk_size: 1000,
    chunk_overlap: 200,
    enable_web_research: true,
    enable_github_indexing: true
  })

  const [showApiKeys, setShowApiKeys] = useState({
    openai: false,
    github: false
  })

  const [testResults, setTestResults] = useState({
    openai: null as boolean | null,
    ollama: null as boolean | null,
    vittoriadb: null as boolean | null,
    github: null as boolean | null
  })

  const [isTesting, setIsTesting] = useState(false)
  const [isSaving, setIsSaving] = useState(false)

  const { health, fetchHealth } = useStatsStore()

  // Load initial config
  useEffect(() => {
    loadConfig()
  }, [])

  const loadConfig = async () => {
    try {
      const response = await apiClient.getConfig()
      setConfig(prev => ({
        ...prev,
        openai_api_key: response.openai_configured ? '••••••••••••••••' : '',
        github_token: response.github_configured ? '••••••••••••••••' : '',
        default_model: response.current_model,
        search_limit: response.search_limit,
        vittoriadb_url: response.vittoriadb_url
      }))
    } catch (error) {
      console.error('Failed to load config:', error)
    }
  }

  const testConnection = async (service: keyof typeof testResults) => {
    setIsTesting(true)
    
    try {
      switch (service) {
        case 'openai':
          // Test OpenAI API
          const openaiTest = await fetch('/api/chat', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
              messages: [{ role: 'user', content: 'test' }],
              model: config.default_model
            })
          })
          setTestResults(prev => ({ ...prev, openai: openaiTest.ok }))
          break

        case 'ollama':
          // Test Ollama connection
          const ollamaTest = await fetch(`${config.ollama_url}/api/tags`)
          setTestResults(prev => ({ ...prev, ollama: ollamaTest.ok }))
          break

        case 'vittoriadb':
          // Test VittoriaDB connection
          const vittoriaTest = await apiClient.health()
          setTestResults(prev => ({ ...prev, vittoriadb: vittoriaTest.vittoriadb_connected }))
          break

        case 'github':
          // Test GitHub token (if provided)
          if (config.github_token && !config.github_token.includes('•')) {
            const githubTest = await fetch('https://api.github.com/user', {
              headers: { 'Authorization': `token ${config.github_token}` }
            })
            setTestResults(prev => ({ ...prev, github: githubTest.ok }))
          }
          break
      }
    } catch (error) {
      setTestResults(prev => ({ ...prev, [service]: false }))
    } finally {
      setIsTesting(false)
    }
  }

  const testAllConnections = async () => {
    setIsTesting(true)
    
    // Test all services in parallel
    await Promise.all([
      testConnection('openai'),
      testConnection('ollama'),
      testConnection('vittoriadb'),
      testConnection('github')
    ])
    
    setIsTesting(false)
  }

  const saveConfig = async () => {
    setIsSaving(true)
    
    try {
      // Here you would typically save to your backend
      // For now, we'll just save to localStorage and show success
      localStorage.setItem('vittoriadb-config', JSON.stringify(config))
      
      // Refresh health status
      await fetchHealth()
      
      toast.success('✅ Settings saved successfully!')
    } catch (error) {
      toast.error('❌ Failed to save settings')
      console.error('Save config error:', error)
    } finally {
      setIsSaving(false)
    }
  }

  const resetToDefaults = () => {
    setConfig({
      openai_api_key: '',
      ollama_url: 'http://localhost:11434',
      github_token: '',
      vittoriadb_url: 'http://localhost:8080',
      default_model: 'gpt-4',
      search_limit: 5,
      chunk_size: 1000,
      chunk_overlap: 200,
      enable_web_research: true,
      enable_github_indexing: true
    })
    setTestResults({
      openai: null,
      ollama: null,
      vittoriadb: null,
      github: null
    })
  }

  const getStatusIcon = (status: boolean | null) => {
    if (status === null) return <AlertCircle className="h-4 w-4 text-muted-foreground" />
    if (status === true) return <CheckCircle className="h-4 w-4 text-green-600" />
    return <XCircle className="h-4 w-4 text-red-600" />
  }

  const getStatusBadge = (status: boolean | null) => {
    if (status === null) return <Badge variant="secondary">Not Tested</Badge>
    if (status === true) return <Badge variant="default" className="bg-green-600">Connected</Badge>
    return <Badge variant="destructive">Failed</Badge>
  }

  return (
    <div className="p-4 space-y-6">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          <Settings className="h-5 w-5" />
          <h2 className="text-lg font-semibold">Settings</h2>
        </div>
        <div className="flex gap-2">
          <Button
            variant="outline"
            size="sm"
            onClick={testAllConnections}
            disabled={isTesting}
          >
            <TestTube className="h-4 w-4 mr-1" />
            {isTesting ? 'Testing...' : 'Test All'}
          </Button>
          <Button
            size="sm"
            onClick={saveConfig}
            disabled={isSaving}
          >
            <Save className="h-4 w-4 mr-1" />
            {isSaving ? 'Saving...' : 'Save'}
          </Button>
        </div>
      </div>

      <Tabs defaultValue="api-keys" className="w-full">
        <TabsList className="grid w-full grid-cols-5">
          <TabsTrigger value="api-keys">API Keys</TabsTrigger>
          <TabsTrigger value="services">Services</TabsTrigger>
          <TabsTrigger value="processing">Processing</TabsTrigger>
          <TabsTrigger value="features">Features</TabsTrigger>
          <TabsTrigger value="appearance">Appearance</TabsTrigger>
        </TabsList>

        <TabsContent value="api-keys" className="space-y-4">
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <Key className="h-4 w-4" />
                API Keys
              </CardTitle>
              <CardDescription>
                Configure API keys for external services
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              {/* OpenAI API Key */}
              <div className="space-y-2">
                <div className="flex items-center justify-between">
                  <Label htmlFor="openai-key">OpenAI API Key</Label>
                  <div className="flex items-center gap-2">
                    {getStatusIcon(testResults.openai)}
                    {getStatusBadge(testResults.openai)}
                    <Button
                      variant="ghost"
                      size="sm"
                      onClick={() => testConnection('openai')}
                      disabled={isTesting || !config.openai_api_key}
                    >
                      Test
                    </Button>
                  </div>
                </div>
                <div className="relative">
                  <Input
                    id="openai-key"
                    type={showApiKeys.openai ? 'text' : 'password'}
                    value={config.openai_api_key}
                    onChange={(e) => setConfig(prev => ({ ...prev, openai_api_key: e.target.value }))}
                    placeholder="sk-..."
                  />
                  <Button
                    variant="ghost"
                    size="sm"
                    className="absolute right-1 top-1 h-6 w-6 p-0"
                    onClick={() => setShowApiKeys(prev => ({ ...prev, openai: !prev.openai }))}
                  >
                    {showApiKeys.openai ? <EyeOff className="h-3 w-3" /> : <Eye className="h-3 w-3" />}
                  </Button>
                </div>
                <p className="text-xs text-muted-foreground">
                  Required for AI chat responses. Get your key from{' '}
                  <a href="https://platform.openai.com/api-keys" target="_blank" className="underline">
                    OpenAI Platform
                  </a>
                </p>
              </div>

              {/* GitHub Token */}
              <div className="space-y-2">
                <div className="flex items-center justify-between">
                  <Label htmlFor="github-token">GitHub Token (Optional)</Label>
                  <div className="flex items-center gap-2">
                    {getStatusIcon(testResults.github)}
                    {getStatusBadge(testResults.github)}
                    <Button
                      variant="ghost"
                      size="sm"
                      onClick={() => testConnection('github')}
                      disabled={isTesting || !config.github_token || config.github_token.includes('•')}
                    >
                      Test
                    </Button>
                  </div>
                </div>
                <div className="relative">
                  <Input
                    id="github-token"
                    type={showApiKeys.github ? 'text' : 'password'}
                    value={config.github_token}
                    onChange={(e) => setConfig(prev => ({ ...prev, github_token: e.target.value }))}
                    placeholder="ghp_..."
                  />
                  <Button
                    variant="ghost"
                    size="sm"
                    className="absolute right-1 top-1 h-6 w-6 p-0"
                    onClick={() => setShowApiKeys(prev => ({ ...prev, github: !prev.github }))}
                  >
                    {showApiKeys.github ? <EyeOff className="h-3 w-3" /> : <Eye className="h-3 w-3" />}
                  </Button>
                </div>
                <p className="text-xs text-muted-foreground">
                  For private repositories and higher rate limits. Get your token from{' '}
                  <a href="https://github.com/settings/tokens" target="_blank" className="underline">
                    GitHub Settings
                  </a>
                </p>
              </div>
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="services" className="space-y-4">
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <Globe className="h-4 w-4" />
                Service URLs
              </CardTitle>
              <CardDescription>
                Configure service endpoints and connections
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              {/* Ollama URL */}
              <div className="space-y-2">
                <div className="flex items-center justify-between">
                  <Label htmlFor="ollama-url">Ollama URL</Label>
                  <div className="flex items-center gap-2">
                    {getStatusIcon(testResults.ollama)}
                    {getStatusBadge(testResults.ollama)}
                    <Button
                      variant="ghost"
                      size="sm"
                      onClick={() => testConnection('ollama')}
                      disabled={isTesting}
                    >
                      Test
                    </Button>
                  </div>
                </div>
                <Input
                  id="ollama-url"
                  value={config.ollama_url}
                  onChange={(e) => setConfig(prev => ({ ...prev, ollama_url: e.target.value }))}
                  placeholder="http://localhost:11434"
                />
                <p className="text-xs text-muted-foreground">
                  Local Ollama server for embeddings. Install with: curl -fsSL https://ollama.ai/install.sh | sh
                </p>
              </div>

              {/* VittoriaDB URL */}
              <div className="space-y-2">
                <div className="flex items-center justify-between">
                  <Label htmlFor="vittoriadb-url">VittoriaDB URL</Label>
                  <div className="flex items-center gap-2">
                    {getStatusIcon(testResults.vittoriadb)}
                    {getStatusBadge(testResults.vittoriadb)}
                    <Button
                      variant="ghost"
                      size="sm"
                      onClick={() => testConnection('vittoriadb')}
                      disabled={isTesting}
                    >
                      Test
                    </Button>
                  </div>
                </div>
                <Input
                  id="vittoriadb-url"
                  value={config.vittoriadb_url}
                  onChange={(e) => setConfig(prev => ({ ...prev, vittoriadb_url: e.target.value }))}
                  placeholder="http://localhost:8080"
                />
                <p className="text-xs text-muted-foreground">
                  VittoriaDB server endpoint for vector operations
                </p>
              </div>

              {/* Default Model */}
              <div className="space-y-2">
                <Label htmlFor="default-model">Default AI Model</Label>
                <Select
                  value={config.default_model}
                  onValueChange={(value) => setConfig(prev => ({ ...prev, default_model: value }))}
                >
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="gpt-4o">GPT-4o (Recommended)</SelectItem>
                    <SelectItem value="gpt-4o-mini">GPT-4o Mini</SelectItem>
                    <SelectItem value="gpt-4">GPT-4</SelectItem>
                    <SelectItem value="gpt-4-turbo">GPT-4 Turbo</SelectItem>
                  </SelectContent>
                </Select>
                <p className="text-xs text-muted-foreground">
                  Default OpenAI model for chat responses
                </p>
              </div>
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="processing" className="space-y-4">
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <Database className="h-4 w-4" />
                Processing Settings
              </CardTitle>
              <CardDescription>
                Configure document processing and search parameters
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              {/* Search Limit */}
              <div className="space-y-2">
                <Label htmlFor="search-limit">Search Results Limit</Label>
                <Input
                  id="search-limit"
                  type="number"
                  min="1"
                  max="20"
                  value={config.search_limit}
                  onChange={(e) => setConfig(prev => ({ ...prev, search_limit: parseInt(e.target.value) || 5 }))}
                />
                <p className="text-xs text-muted-foreground">
                  Number of search results to retrieve for context (1-20)
                </p>
              </div>

              {/* Chunk Size */}
              <div className="space-y-2">
                <Label htmlFor="chunk-size">Document Chunk Size</Label>
                <Input
                  id="chunk-size"
                  type="number"
                  min="100"
                  max="5000"
                  value={config.chunk_size}
                  onChange={(e) => setConfig(prev => ({ ...prev, chunk_size: parseInt(e.target.value) || 1000 }))}
                />
                <p className="text-xs text-muted-foreground">
                  Size of document chunks in characters (100-5000)
                </p>
              </div>

              {/* Chunk Overlap */}
              <div className="space-y-2">
                <Label htmlFor="chunk-overlap">Chunk Overlap</Label>
                <Input
                  id="chunk-overlap"
                  type="number"
                  min="0"
                  max="1000"
                  value={config.chunk_overlap}
                  onChange={(e) => setConfig(prev => ({ ...prev, chunk_overlap: parseInt(e.target.value) || 200 }))}
                />
                <p className="text-xs text-muted-foreground">
                  Overlap between chunks in characters (0-1000)
                </p>
              </div>
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="features" className="space-y-4">
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <Bot className="h-4 w-4" />
                Feature Toggles
              </CardTitle>
              <CardDescription>
                Enable or disable specific features
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              {/* Web Research */}
              <div className="flex items-center justify-between">
                <div className="space-y-0.5">
                  <Label htmlFor="web-research">Web Research</Label>
                  <p className="text-xs text-muted-foreground">
                    Allow web research and automatic content storage
                  </p>
                </div>
                <Switch
                  id="web-research"
                  checked={config.enable_web_research}
                  onCheckedChange={(checked) => setConfig(prev => ({ ...prev, enable_web_research: checked }))}
                />
              </div>

              {/* GitHub Indexing */}
              <div className="flex items-center justify-between">
                <div className="space-y-0.5">
                  <Label htmlFor="github-indexing">GitHub Repository Indexing</Label>
                  <p className="text-xs text-muted-foreground">
                    Allow indexing of GitHub repositories for code search
                  </p>
                </div>
                <Switch
                  id="github-indexing"
                  checked={config.enable_github_indexing}
                  onCheckedChange={(checked) => setConfig(prev => ({ ...prev, enable_github_indexing: checked }))}
                />
              </div>
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="appearance" className="space-y-4">
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <Monitor className="h-4 w-4" />
                Appearance
              </CardTitle>
              <CardDescription>
                Customize the look and feel of the interface
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              {/* Theme Selection */}
              <div className="space-y-3">
                <Label>Theme</Label>
                <div className="grid grid-cols-3 gap-2">
                  <Button
                    variant={theme === 'light' ? 'default' : 'outline'}
                    size="sm"
                    onClick={() => setTheme('light')}
                    className="flex items-center gap-2"
                  >
                    <Sun className="h-4 w-4" />
                    Light
                  </Button>
                  <Button
                    variant={theme === 'dark' ? 'default' : 'outline'}
                    size="sm"
                    onClick={() => setTheme('dark')}
                    className="flex items-center gap-2"
                  >
                    <Moon className="h-4 w-4" />
                    Dark
                  </Button>
                  <Button
                    variant={theme === 'system' ? 'default' : 'outline'}
                    size="sm"
                    onClick={() => setTheme('system')}
                    className="flex items-center gap-2"
                  >
                    <Monitor className="h-4 w-4" />
                    System
                  </Button>
                </div>
                <p className="text-xs text-muted-foreground">
                  Choose your preferred theme. System will match your device's theme.
                </p>
              </div>

              {/* Theme Preview */}
              <div className="space-y-2">
                <Label>Preview</Label>
                <div className="border rounded-lg p-4 space-y-2">
                  <div className="flex items-center gap-2">
                    <div className="w-6 h-6 rounded-full bg-primary flex items-center justify-center">
                      <span className="text-xs text-primary-foreground font-bold">V</span>
                    </div>
                    <div>
                      <div className="text-sm font-medium">VittoriaDB Assistant</div>
                      <div className="text-xs text-muted-foreground">AI-powered knowledge base</div>
                    </div>
                  </div>
                  <div className="bg-muted/50 rounded p-2 text-xs">
                    This is how messages will appear in {theme || 'system'} mode.
                  </div>
                </div>
              </div>
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>

      <div className="flex justify-between">
        <Button variant="outline" onClick={resetToDefaults}>
          Reset to Defaults
        </Button>
        <div className="flex gap-2">
          <Button variant="outline" onClick={testAllConnections} disabled={isTesting}>
            <TestTube className="h-4 w-4 mr-1" />
            Test All Connections
          </Button>
          <Button onClick={saveConfig} disabled={isSaving}>
            <Save className="h-4 w-4 mr-1" />
            Save Settings
          </Button>
        </div>
      </div>
    </div>
  )
}
