'use client'

import { useState } from 'react'
import { Button } from '@/components/ui/button'
import { ScrollArea } from '@/components/ui/scroll-area'
import { Separator } from '@/components/ui/separator'
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogTrigger } from '@/components/ui/dialog'
import {
  X,
  Database,
  Activity,
  Settings,
  Info,
  Sparkles,
  RefreshCw,
  Plus,
  MessageSquare,
  Save,
  ChevronDown,
  ChevronRight
} from 'lucide-react'
import { useStatsStore } from '@/store'
import { SettingsPanel } from '@/components/settings-panel'
import { cn } from '@/lib/utils'

interface SidebarProps {
  isOpen: boolean
  onToggle: () => void
  onNewChat?: () => void
  currentSessionId?: string | null
  autoSaveEnabled?: boolean
  onSaveChatHistory?: () => void
}

export function Sidebar({ isOpen, onToggle, onNewChat, currentSessionId, autoSaveEnabled, onSaveChatHistory }: SidebarProps) {
  const [settingsOpen, setSettingsOpen] = useState(false)
  const [collectionsExpanded, setCollectionsExpanded] = useState(true)
  const [systemStatusExpanded, setSystemStatusExpanded] = useState(false)
  const { stats, health, isLoading, updatingCollections } = useStatsStore()

  // No manual refresh needed - WebSocket handles all updates automatically

  if (!isOpen) return null

  return (
    <div className="w-80 bg-card border-r border-border flex flex-col h-full">
      {/* Header */}
      <div className="p-4 border-b">
        <div className="flex items-center justify-between">
          <h2 className="font-semibold">Knowledge Base</h2>
          <Button
            variant="ghost"
            size="sm"
            onClick={onToggle}
            className="lg:hidden"
          >
            <X className="h-4 w-4" />
          </Button>
        </div>
        
        {/* Status Indicator */}
        <div className="flex items-center gap-2 mt-2 text-sm">
          <div className={cn(
            "w-2 h-2 rounded-full",
            health?.vittoriadb_connected ? "bg-green-500" : "bg-red-500"
          )} />
          <span className="text-muted-foreground">
            {health?.vittoriadb_connected ? "Connected" : "Disconnected"}
          </span>
        </div>
      </div>

      {/* Chat Management Section */}
      <div className="p-4 border-b">
        <div className="space-y-3">
          <div className="flex items-center justify-between">
            <h3 className="font-medium flex items-center gap-2">
              <MessageSquare className="h-4 w-4" />
              Chat
            </h3>
            {currentSessionId && (
              <div className="flex items-center gap-1 text-xs text-muted-foreground">
                <div className="w-2 h-2 rounded-full bg-green-500" />
                <span>Active</span>
              </div>
            )}
          </div>
          
          {/* New Chat Button */}
          <Button
            onClick={onNewChat}
            className="w-full flex items-center gap-2"
            variant="outline"
            size="sm"
          >
            <Plus className="h-4 w-4" />
            New Chat
          </Button>
          
          {/* Session Info */}
          {currentSessionId && (
            <div className="bg-muted/50 rounded-lg p-3">
              <div className="flex items-center justify-between mb-2">
                <span className="text-sm font-medium">Current Session</span>
                {autoSaveEnabled && (
                  <div className="flex items-center gap-1 text-xs text-green-600 dark:text-green-400">
                    <Save className="h-3 w-3" />
                    <span>Auto-save</span>
                  </div>
                )}
              </div>
              <div className="text-xs text-muted-foreground">
                ID: {currentSessionId.slice(0, 8)}...
              </div>
              {onSaveChatHistory && (
                <Button
                  onClick={onSaveChatHistory}
                  variant="ghost"
                  size="sm"
                  className="w-full mt-2 h-7 text-xs"
                >
                  <Save className="h-3 w-3 mr-1" />
                  Save Now
                </Button>
              )}
            </div>
          )}
        </div>
      </div>

      {/* Quick Info */}
      <div className="p-4">
        <div className="bg-gradient-to-r from-blue-500/10 to-purple-600/10 rounded-lg p-3 border">
          <div className="flex items-center gap-2 mb-2">
            <Sparkles className="h-4 w-4 text-blue-600" />
            <span className="text-sm font-medium">Quick Tips</span>
          </div>
          <div className="text-xs text-muted-foreground space-y-1">
            <p>💬 <strong>Chat:</strong> Ask questions about your documents</p>
            <p>📁 <strong>Upload:</strong> Drag files into the chat area</p>
            <p>🌐 <strong>Research:</strong> Toggle web search in chat input</p>
            <p>⚙️ <strong>Settings:</strong> Configure API keys below</p>
          </div>
        </div>
      </div>

      <Separator />

      {/* Collection Stats */}
      <ScrollArea className="flex-1">
        <div className="p-4">
          <div className="space-y-4">
            <Button
              variant="ghost"
              size="sm"
              onClick={() => setCollectionsExpanded(!collectionsExpanded)}
              className="w-full justify-start p-2 h-auto"
            >
              <div className="flex items-center gap-2 w-full">
                {collectionsExpanded ? (
                  <ChevronDown className="h-3 w-3" />
                ) : (
                  <ChevronRight className="h-3 w-3" />
                )}
                <Database className="h-4 w-4" />
                <span className="text-sm font-medium">Collections</span>
                <div className="flex items-center gap-1 text-xs text-muted-foreground ml-auto">
                  {isLoading && <RefreshCw className="h-3 w-3 animate-spin" />}
                  <span>Auto-sync</span>
                </div>
              </div>
            </Button>
            
            {collectionsExpanded && stats?.collections ? (
              <div className="space-y-3">
                {Object.entries(stats.collections).map(([name, collection]) => (
                  <div key={name} className="bg-muted/50 rounded-lg p-3">
                    <div className="flex items-center justify-between mb-1">
                      <span className="text-sm font-medium capitalize">
                        {name.replace('_', ' ')}
                      </span>
                      <span className="text-xs text-muted-foreground">
                        {collection.vector_count} docs
                      </span>
                    </div>
              <div className="text-xs text-muted-foreground">
                {collection.dimensions}D • {collection.index_type?.toUpperCase() || 'FLAT'} • {collection.metric}
                {(isLoading || updatingCollections.has(name)) && (
                  <span className="ml-2 text-blue-600">• Updating...</span>
                )}
              </div>
                  </div>
                ))}
              </div>
            ) : (
              <div className="text-sm text-muted-foreground">
                No collections available
              </div>
            )}
          </div>

          <Separator className="my-4" />

          {/* System Info */}
          <div className="space-y-3">
            <Button
              variant="ghost"
              size="sm"
              onClick={() => setSystemStatusExpanded(!systemStatusExpanded)}
              className="w-full justify-start p-2 h-auto"
            >
              <div className="flex items-center gap-2">
                {systemStatusExpanded ? (
                  <ChevronDown className="h-3 w-3" />
                ) : (
                  <ChevronRight className="h-3 w-3" />
                )}
                <Activity className="h-4 w-4" />
                <span className="text-sm font-medium">System Status</span>
              </div>
            </Button>
            
            {systemStatusExpanded && (
              <div className="space-y-2 text-sm pl-2">
              <div className="flex justify-between">
                <span className="text-muted-foreground">Total Vectors:</span>
                <span>{stats?.total_vectors || 0}</span>
              </div>
              
              <div className="flex justify-between">
                <span className="text-muted-foreground">OpenAI:</span>
                <span className={cn(
                  health?.openai_configured ? "text-green-600" : "text-red-600"
                )}>
                  {health?.openai_configured ? "Configured" : "Not configured"}
                </span>
              </div>
              </div>
            )}
          </div>
        </div>
      </ScrollArea>

      {/* Footer */}
      <div className="p-4 border-t">
        <Dialog open={settingsOpen} onOpenChange={setSettingsOpen}>
          <DialogTrigger asChild>
            <Button 
              variant="outline"
              className="w-full justify-start" 
              size="sm"
            >
              <Settings className="h-4 w-4 mr-2" />
              Settings
            </Button>
          </DialogTrigger>
          <DialogContent className="max-w-4xl max-h-[80vh] overflow-y-auto">
            <DialogHeader>
              <DialogTitle>VittoriaDB RAG Settings</DialogTitle>
            </DialogHeader>
            <SettingsPanel />
          </DialogContent>
        </Dialog>
      </div>
    </div>
  )
}
