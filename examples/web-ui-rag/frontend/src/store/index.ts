// Zustand store for global state management

import { create } from 'zustand'
import { devtools, persist } from 'zustand/middleware'
import {
  ChatMessage,
  SystemStats,
  HealthResponse,
  WebResearchResponse,
  GitHubIndexResponse,
  UIState,
  ChatState,
  UploadState,
  ResearchState,
  GitHubState,
  StatsState,
} from '@/types'
import { apiClient, WebSocketClient } from '@/lib/api'

// UI Store
interface UIStore extends UIState {
  setSidebarOpen: (open: boolean) => void
  toggleSidebar: () => void
  setDarkMode: (dark: boolean) => void
  toggleDarkMode: () => void
  setCurrentModel: (model: string) => void
  setSearchLimit: (limit: number) => void
}

export const useUIStore = create<UIStore>()(
  devtools(
    persist(
      (set) => ({
        sidebarOpen: true,
        darkMode: false,
        currentModel: 'gpt-4',
        searchLimit: 5,
        
        setSidebarOpen: (open) => set({ sidebarOpen: open }),
        toggleSidebar: () => set((state) => ({ sidebarOpen: !state.sidebarOpen })),
        setDarkMode: (dark) => set({ darkMode: dark }),
        toggleDarkMode: () => set((state) => ({ darkMode: !state.darkMode })),
        setCurrentModel: (model) => set({ currentModel: model }),
        setSearchLimit: (limit) => set({ searchLimit: limit }),
      }),
      {
        name: 'ui-store',
        partialize: (state) => ({
          darkMode: state.darkMode,
          currentModel: state.currentModel,
          searchLimit: state.searchLimit,
        }),
      }
    )
  )
)

// Chat Store
interface ChatStore extends ChatState {
  wsClient: WebSocketClient | null
  addMessage: (message: ChatMessage) => void
  updateLastMessage: (content: string) => void
  setLoading: (loading: boolean) => void
  setConnected: (connected: boolean) => void
  setError: (error: string | null) => void
  clearMessages: () => void
  connectWebSocket: () => Promise<void>
  disconnectWebSocket: () => void
  sendMessage: (message: string) => void
}

export const useChatStore = create<ChatStore>()(
  devtools((set, get) => ({
    messages: [],
    isLoading: false,
    isConnected: false,
    error: null,
    wsClient: null,

    addMessage: (message) =>
      set((state) => ({
        messages: [...state.messages, message],
      })),

    updateLastMessage: (content) =>
      set((state) => {
        const messages = [...state.messages]
        const lastMessage = messages[messages.length - 1]
        if (lastMessage && lastMessage.role === 'assistant') {
          lastMessage.content += content
        }
        return { messages }
      }),

    setLoading: (loading) => set({ isLoading: loading }),
    setConnected: (connected) => set({ isConnected: connected }),
    setError: (error) => set({ error }),
    clearMessages: () => set({ messages: [] }),

    connectWebSocket: async () => {
      const wsURL = apiClient.getWebSocketURL()
      const wsClient = new WebSocketClient(wsURL)

      try {
        await wsClient.connect()
        set({ wsClient, isConnected: true, error: null })

        wsClient.onMessage((data) => {
          const { type, content, sources, error, processing_time } = data

          switch (type) {
            case 'typing':
            case 'status':
              // Handle status updates (could show in UI)
              console.log('Status:', content)
              break

            case 'content':
              get().updateLastMessage(content)
              break

            case 'sources':
              set((state) => {
                const messages = [...state.messages]
                const lastMessage = messages[messages.length - 1]
                if (lastMessage && lastMessage.role === 'assistant') {
                  lastMessage.sources = sources
                }
                return { messages }
              })
              break

            case 'done':
              set({ isLoading: false })
              console.log(`Processing completed in ${processing_time}s`)
              break

            case 'error':
              set({ error, isLoading: false })
              break
          }
        })
      } catch (error) {
        console.error('Failed to connect WebSocket:', error)
        set({ error: 'Failed to connect to server', isConnected: false })
      }
    },

    disconnectWebSocket: () => {
      const { wsClient } = get()
      if (wsClient) {
        wsClient.close()
        set({ wsClient: null, isConnected: false })
      }
    },

    sendMessage: (message) => {
      const { wsClient, messages, currentModel, searchLimit } = get()
      const uiStore = useUIStore.getState()

      if (!wsClient || !wsClient.isConnected) {
        set({ error: 'Not connected to server' })
        return
      }

      // Add user message
      const userMessage: ChatMessage = {
        role: 'user',
        content: message,
        timestamp: Date.now() / 1000,
      }

      // Add assistant message placeholder
      const assistantMessage: ChatMessage = {
        role: 'assistant',
        content: '',
        timestamp: Date.now() / 1000,
      }

      set((state) => ({
        messages: [...state.messages, userMessage, assistantMessage],
        isLoading: true,
        error: null,
      }))

      // Send message via WebSocket
      wsClient.send({
        message,
        chat_history: messages,
        model: uiStore.currentModel,
        search_limit: uiStore.searchLimit,
      })
    },
  }))
)

// Upload Store
interface UploadStore extends UploadState {
  setUploading: (uploading: boolean) => void
  setUploadProgress: (progress: number) => void
  addUploadedFile: (filename: string) => void
  setError: (error: string | null) => void
  clearUploadedFiles: () => void
}

export const useUploadStore = create<UploadStore>()(
  devtools((set) => ({
    isUploading: false,
    uploadProgress: 0,
    uploadedFiles: [],
    error: null,

    setUploading: (uploading) => set({ isUploading: uploading }),
    setUploadProgress: (progress) => set({ uploadProgress: progress }),
    addUploadedFile: (filename) =>
      set((state) => ({
        uploadedFiles: [...state.uploadedFiles, filename],
      })),
    setError: (error) => set({ error }),
    clearUploadedFiles: () => set({ uploadedFiles: [] }),
  }))
)

// Research Store
interface ResearchStore extends ResearchState {
  setResearching: (researching: boolean) => void
  setLastQuery: (query: string) => void
  setResults: (results: WebResearchResponse | null) => void
  setError: (error: string | null) => void
}

export const useResearchStore = create<ResearchStore>()(
  devtools((set) => ({
    isResearching: false,
    lastQuery: '',
    results: null,
    error: null,

    setResearching: (researching) => set({ isResearching: researching }),
    setLastQuery: (query) => set({ lastQuery: query }),
    setResults: (results) => set({ results }),
    setError: (error) => set({ error }),
  }))
)

// GitHub Store
interface GitHubStore extends GitHubState {
  setIndexing: (indexing: boolean) => void
  setLastRepo: (repo: string) => void
  setResults: (results: GitHubIndexResponse | null) => void
  setError: (error: string | null) => void
}

export const useGitHubStore = create<GitHubStore>()(
  devtools((set) => ({
    isIndexing: false,
    lastRepo: '',
    results: null,
    error: null,

    setIndexing: (indexing) => set({ isIndexing: indexing }),
    setLastRepo: (repo) => set({ lastRepo: repo }),
    setResults: (results) => set({ results }),
    setError: (error) => set({ error }),
  }))
)

// Stats Store - WebSocket-only (no polling)
interface StatsStore extends StatsState {
  setStats: (stats: SystemStats | null) => void
  setHealth: (health: HealthResponse | null) => void
  setLoading: (loading: boolean) => void
  setError: (error: string | null) => void
  updatingCollections: Set<string>
  setCollectionUpdating: (collection: string, updating: boolean) => void
  // Remove fetchStats and fetchHealth - WebSocket handles all updates
}

export const useStatsStore = create<StatsStore>()(
  devtools((set, get) => ({
    stats: null,
    health: null,
    isLoading: false,
    error: null,
    updatingCollections: new Set(),

    setStats: (stats) => set({ stats, isLoading: false, error: null }),
    setHealth: (health) => set({ health }),
    setLoading: (loading) => set({ isLoading: loading }),
    setError: (error) => set({ error, isLoading: false }),

    setCollectionUpdating: (collection, updating) => {
      const { updatingCollections } = get()
      const newSet = new Set(updatingCollections)
      if (updating) {
        newSet.add(collection)
      } else {
        newSet.delete(collection)
      }
      set({ updatingCollections: newSet })
    },
    
    // Note: Stats are now updated automatically via WebSocket notifications
    // No manual fetching needed - useWebSocketNotifications handles all updates
  }))
)

// Combined store hook for convenience
export const useStore = () => ({
  ui: useUIStore(),
  chat: useChatStore(),
  upload: useUploadStore(),
  research: useResearchStore(),
  github: useGitHubStore(),
  stats: useStatsStore(),
})
