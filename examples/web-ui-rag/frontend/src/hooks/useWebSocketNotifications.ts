'use client'

import { useEffect, useRef, useState, useCallback } from 'react'
import { useStatsStore } from '@/store'
import toast from 'react-hot-toast'

interface NotificationData {
  type: string
  data: any
  timestamp: number
  id?: string
}

interface ProcessingStatus {
  document_id: string
  filename: string
  status: 'processing' | 'completed' | 'error'
  progress?: number
  message?: string
  chunks_created?: number
  processing_time?: number
  error?: string
}

interface GitHubIndexingStatus {
  indexing_id: string
  repo_url: string
  status: 'indexing' | 'completed' | 'error'
  progress?: number
  message?: string
  files_indexed?: number
  repository?: string
  processing_time?: number
  error?: string
}

export function useWebSocketNotifications() {
  const [isConnected, setIsConnected] = useState(false)
  const [processingFiles, setProcessingFiles] = useState<Map<string, ProcessingStatus>>(new Map())
  const [githubIndexing, setGithubIndexing] = useState<Map<string, GitHubIndexingStatus>>(new Map())
  const wsRef = useRef<WebSocket | null>(null)
  const reconnectTimeoutRef = useRef<NodeJS.Timeout | null>(null)
  const { setStats, setHealth, setCollectionUpdating } = useStatsStore()

  const connect = useCallback(() => {
    if (wsRef.current?.readyState === WebSocket.OPEN) {
      return
    }

    try {
      const wsUrl = process.env.NEXT_PUBLIC_WS_URL || 'ws://localhost:8501'
      wsRef.current = new WebSocket(`${wsUrl}/ws/notifications`)

      wsRef.current.onopen = () => {
        console.log('🔗 WebSocket connected')
        setIsConnected(true)
        
        // Clear any reconnection timeout
        if (reconnectTimeoutRef.current) {
          clearTimeout(reconnectTimeoutRef.current)
          reconnectTimeoutRef.current = null
        }

        // Request initial stats immediately upon connection
        if (wsRef.current?.readyState === WebSocket.OPEN) {
          wsRef.current.send(JSON.stringify({ type: 'get_stats' }))
          wsRef.current.send(JSON.stringify({ type: 'get_health' }))
        }

        // Send ping to keep connection alive
        const pingInterval = setInterval(() => {
          if (wsRef.current?.readyState === WebSocket.OPEN) {
            wsRef.current.send(JSON.stringify({ type: 'ping' }))
          } else {
            clearInterval(pingInterval)
          }
        }, 30000) // Ping every 30 seconds
      }

      wsRef.current.onmessage = (event) => {
        try {
          const notification: NotificationData = JSON.parse(event.data)
          handleNotification(notification)
        } catch (error) {
          console.error('Failed to parse WebSocket message:', error)
        }
      }

      wsRef.current.onclose = (event) => {
        console.log('🔌 WebSocket disconnected:', event.code, event.reason)
        setIsConnected(false)
        
        // Attempt to reconnect after a delay
        if (!reconnectTimeoutRef.current) {
          reconnectTimeoutRef.current = setTimeout(() => {
            console.log('🔄 Attempting to reconnect WebSocket...')
            connect()
          }, 3000)
        }
      }

      wsRef.current.onerror = (error) => {
        console.error('❌ WebSocket error:', error)
        setIsConnected(false)
      }

    } catch (error) {
      console.error('Failed to create WebSocket connection:', error)
      setIsConnected(false)
    }
  }, [])

  const handleNotification = useCallback((notification: NotificationData) => {
    console.log('📨 Received notification:', notification.type, notification.data)

    switch (notification.type) {
      case 'stats_update':
        setStats(notification.data)
        break

      case 'processing_start':
        const startData = notification.data as ProcessingStatus
        setProcessingFiles(prev => new Map(prev.set(startData.document_id, startData)))
        toast.loading(`📤 Processing ${startData.filename}...`, { 
          id: startData.document_id,
          duration: Infinity 
        })
        setCollectionUpdating('documents', true)
        break

      case 'processing_progress':
        const progressData = notification.data
        setProcessingFiles(prev => {
          const newMap = new Map(prev)
          const existing = newMap.get(progressData.document_id)
          if (existing) {
            newMap.set(progressData.document_id, {
              ...existing,
              progress: progressData.progress,
              message: progressData.message
            })
          }
          return newMap
        })
        
        // Update toast with progress
        toast.loading(`⚙️ ${progressData.message} (${progressData.progress}%)`, {
          id: progressData.document_id,
          duration: Infinity
        })
        break

      case 'processing_complete':
        const completeData = notification.data
        setProcessingFiles(prev => {
          const newMap = new Map(prev)
          newMap.delete(completeData.document_id)
          return newMap
        })
        
        toast.success(`✅ Processing complete: ${completeData.chunks_created} chunks created`, {
          id: completeData.document_id,
          duration: 4000
        })
        setCollectionUpdating('documents', false)
        break

      case 'processing_error':
        const errorData = notification.data
        setProcessingFiles(prev => {
          const newMap = new Map(prev)
          newMap.delete(errorData.document_id)
          return newMap
        })
        
        toast.error(`❌ Processing failed: ${errorData.error}`, {
          id: errorData.document_id,
          duration: 6000
        })
        setCollectionUpdating('documents', false)
        break

      case 'collection_update':
        const collectionData = notification.data
        toast.success(`📊 Collection '${collectionData.collection_name}' updated`)
        break

      case 'health_update':
        const healthData = notification.data
        setHealth(healthData)
        if (healthData.status === 'unhealthy') {
          toast.error('⚠️ System health warning')
        }
        break

      case 'system_status':
        // Connection status - no toast needed
        console.log('WebSocket status:', notification.data.status)
        break

      case 'github_indexing_start':
        const githubStartData = notification.data
        const githubStatus: GitHubIndexingStatus = {
          indexing_id: githubStartData.indexing_id,
          repo_url: githubStartData.repo_url,
          status: 'indexing',
          progress: 0,
          message: githubStartData.message
        }
        setGithubIndexing(prev => new Map(prev.set(githubStartData.indexing_id, githubStatus)))
        toast.loading(`🔗 ${githubStartData.message}`, { 
          id: githubStartData.indexing_id,
          duration: Infinity 
        })
        setCollectionUpdating('github_code', true)
        break

      case 'github_indexing_progress':
        const githubProgressData = notification.data
        setGithubIndexing(prev => {
          const newMap = new Map(prev)
          const existing = newMap.get(githubProgressData.indexing_id)
          if (existing) {
            newMap.set(githubProgressData.indexing_id, {
              ...existing,
              progress: githubProgressData.progress,
              message: githubProgressData.message
            })
          }
          return newMap
        })
        
        // Update toast with progress
        toast.loading(`⚙️ ${githubProgressData.message} (${githubProgressData.progress}%)`, {
          id: githubProgressData.indexing_id,
          duration: Infinity
        })
        break

      case 'github_indexing_complete':
        const githubCompleteData = notification.data
        setGithubIndexing(prev => {
          const newMap = new Map(prev)
          newMap.delete(githubCompleteData.indexing_id)
          return newMap
        })
        
        toast.success(`✅ ${githubCompleteData.message}`, {
          id: githubCompleteData.indexing_id,
          duration: 6000
        })
        setCollectionUpdating('github_code', false)
        break

      case 'github_indexing_error':
        const githubErrorData = notification.data
        setGithubIndexing(prev => {
          const newMap = new Map(prev)
          newMap.delete(githubErrorData.indexing_id)
          return newMap
        })
        
        toast.error(`❌ ${githubErrorData.message}`, {
          id: githubErrorData.indexing_id,
          duration: 8000
        })
        setCollectionUpdating('github_code', false)
        break

      default:
        console.log('Unknown notification type:', notification.type)
    }
  }, [setStats, setHealth, setCollectionUpdating])

  const disconnect = useCallback(() => {
    if (reconnectTimeoutRef.current) {
      clearTimeout(reconnectTimeoutRef.current)
      reconnectTimeoutRef.current = null
    }
    
    if (wsRef.current) {
      wsRef.current.close()
      wsRef.current = null
    }
    
    setIsConnected(false)
  }, [])

  const requestProcessingStatus = useCallback(() => {
    if (wsRef.current?.readyState === WebSocket.OPEN) {
      wsRef.current.send(JSON.stringify({ type: 'get_processing_status' }))
    }
  }, [])

  // Auto-connect on mount
  useEffect(() => {
    connect()
    
    return () => {
      disconnect()
    }
  }, [connect, disconnect])

  return {
    isConnected,
    processingFiles: Array.from(processingFiles.values()),
    githubIndexing: Array.from(githubIndexing.values()),
    connect,
    disconnect,
    requestProcessingStatus
  }
}
