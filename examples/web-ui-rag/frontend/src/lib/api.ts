// API client for VittoriaDB RAG Backend

import axios, { AxiosInstance, AxiosResponse } from 'axios'
import {
  ChatRequest,
  ChatResponse,
  FileUploadResponse,
  WebResearchRequest,
  WebResearchResponse,
  GitHubIndexRequest,
  GitHubIndexResponse,
  SearchRequest,
  SearchResponse,
  SystemStats,
  HealthResponse,
  ConfigResponse,
} from '@/types'

class APIClient {
  private client: AxiosInstance
  private baseURL: string

  constructor() {
    this.baseURL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8501'
    
    this.client = axios.create({
      baseURL: this.baseURL,
      timeout: 30000, // 30 seconds
      headers: {
        'Content-Type': 'application/json',
      },
    })

    // Request interceptor
    this.client.interceptors.request.use(
      (config) => {
        console.log(`🔄 API Request: ${config.method?.toUpperCase()} ${config.url}`)
        return config
      },
      (error) => {
        console.error('❌ API Request Error:', error)
        return Promise.reject(error)
      }
    )

    // Response interceptor
    this.client.interceptors.response.use(
      (response) => {
        console.log(`✅ API Response: ${response.status} ${response.config.url}`)
        return response
      },
      (error) => {
        console.error('❌ API Response Error:', error.response?.data || error.message)
        return Promise.reject(error)
      }
    )
  }

  // Health check
  async health(): Promise<HealthResponse> {
    const response: AxiosResponse<HealthResponse> = await this.client.get('/health')
    return response.data
  }

  // System statistics
  async getStats(): Promise<SystemStats> {
    const response: AxiosResponse<SystemStats> = await this.client.get('/stats')
    return response.data
  }

  // Configuration
  async getConfig(): Promise<ConfigResponse> {
    const response: AxiosResponse<ConfigResponse> = await this.client.get('/config')
    return response.data
  }

  // Chat (non-streaming)
  async chat(request: ChatRequest): Promise<ChatResponse> {
    const response: AxiosResponse<ChatResponse> = await this.client.post('/chat', request)
    return response.data
  }

  // File upload
  async uploadFile(file: File): Promise<FileUploadResponse> {
    const formData = new FormData()
    formData.append('file', file)

    const response: AxiosResponse<FileUploadResponse> = await this.client.post(
      '/upload',
      formData,
      {
        headers: {
          'Content-Type': 'multipart/form-data',
        },
        timeout: 60000, // 1 minute for file uploads
      }
    )
    return response.data
  }

  // Web research
  async webResearch(request: WebResearchRequest): Promise<WebResearchResponse> {
    const response: AxiosResponse<WebResearchResponse> = await this.client.post(
      '/research',
      request,
      {
        timeout: 60000, // 1 minute for web research
      }
    )
    return response.data
  }

  // GitHub indexing
  async indexGitHub(request: GitHubIndexRequest): Promise<GitHubIndexResponse> {
    const response: AxiosResponse<GitHubIndexResponse> = await this.client.post(
      '/github/index',
      request,
      {
        timeout: 300000, // 5 minutes for GitHub indexing
      }
    )
    return response.data
  }

  // Search knowledge base
  async search(request: SearchRequest): Promise<SearchResponse> {
    const response: AxiosResponse<SearchResponse> = await this.client.post('/search', request)
    return response.data
  }

  // Get WebSocket URL
  getWebSocketURL(): string {
    const wsProtocol = this.baseURL.startsWith('https') ? 'wss' : 'ws'
    const wsURL = this.baseURL.replace(/^https?/, wsProtocol)
    return `${wsURL}/ws/chat`
  }
}

// WebSocket client for streaming chat
export class WebSocketClient {
  private ws: WebSocket | null = null
  private url: string
  private reconnectAttempts = 0
  private maxReconnectAttempts = 5
  private reconnectDelay = 1000

  constructor(url: string) {
    this.url = url
  }

  connect(): Promise<void> {
    return new Promise((resolve, reject) => {
      try {
        this.ws = new WebSocket(this.url)

        this.ws.onopen = () => {
          console.log('✅ WebSocket connected')
          this.reconnectAttempts = 0
          resolve()
        }

        this.ws.onerror = (error) => {
          console.error('❌ WebSocket error:', error)
          reject(error)
        }

        this.ws.onclose = (event) => {
          console.log('🔌 WebSocket closed:', event.code, event.reason)
          this.handleReconnect()
        }
      } catch (error) {
        reject(error)
      }
    })
  }

  private handleReconnect() {
    if (this.reconnectAttempts < this.maxReconnectAttempts) {
      this.reconnectAttempts++
      console.log(`🔄 Reconnecting WebSocket (attempt ${this.reconnectAttempts})...`)
      
      setTimeout(() => {
        this.connect().catch(console.error)
      }, this.reconnectDelay * this.reconnectAttempts)
    } else {
      console.error('❌ Max reconnection attempts reached')
    }
  }

  send(data: any): void {
    if (this.ws && this.ws.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify(data))
    } else {
      console.error('❌ WebSocket not connected')
    }
  }

  onMessage(callback: (data: any) => void): void {
    if (this.ws) {
      this.ws.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data)
          callback(data)
        } catch (error) {
          console.error('❌ Failed to parse WebSocket message:', error)
        }
      }
    }
  }

  close(): void {
    if (this.ws) {
      this.ws.close()
      this.ws = null
    }
  }

  get isConnected(): boolean {
    return this.ws?.readyState === WebSocket.OPEN
  }
}

// Create singleton API client
export const apiClient = new APIClient()

// Utility functions
export const uploadFiles = async (files: File[]): Promise<FileUploadResponse[]> => {
  const results: FileUploadResponse[] = []
  
  for (const file of files) {
    try {
      const result = await apiClient.uploadFile(file)
      results.push(result)
    } catch (error) {
      console.error(`Failed to upload ${file.name}:`, error)
      // Add error result
      results.push({
        success: false,
        message: `Failed to upload ${file.name}`,
        filename: file.name,
        document_id: '',
        chunks_created: 0,
        processing_time: 0,
        metadata: {},
      })
    }
  }
  
  return results
}

export const isAPIError = (error: any): boolean => {
  return error.response && error.response.data && error.response.data.error
}

export const getAPIErrorMessage = (error: any): string => {
  if (isAPIError(error)) {
    return error.response.data.error
  }
  
  if (error.message) {
    return error.message
  }
  
  return 'An unknown error occurred'
}

export default apiClient
