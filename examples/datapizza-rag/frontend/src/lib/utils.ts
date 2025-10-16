import { type ClassValue, clsx } from "clsx"
import { twMerge } from "tailwind-merge"

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}

export function formatTimestamp(timestamp: number): string {
  const date = new Date(timestamp * 1000)
  const now = new Date()
  const diff = now.getTime() - date.getTime()
  
  // Less than 1 minute
  if (diff < 60000) {
    return 'Just now'
  }
  
  // Less than 1 hour
  if (diff < 3600000) {
    const minutes = Math.floor(diff / 60000)
    return `${minutes}m ago`
  }
  
  // Less than 1 day
  if (diff < 86400000) {
    const hours = Math.floor(diff / 3600000)
    return `${hours}h ago`
  }
  
  // More than 1 day
  return date.toLocaleDateString()
}

export function formatFileSize(bytes: number): string {
  if (bytes === 0) return '0 Bytes'
  
  const k = 1024
  const sizes = ['Bytes', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

export function formatDuration(seconds: number): string {
  if (seconds < 1) {
    return `${Math.round(seconds * 1000)}ms`
  }
  
  if (seconds < 60) {
    return `${seconds.toFixed(1)}s`
  }
  
  const minutes = Math.floor(seconds / 60)
  const remainingSeconds = Math.floor(seconds % 60)
  
  return `${minutes}m ${remainingSeconds}s`
}

export function truncateText(text: string, maxLength: number): string {
  if (text.length <= maxLength) return text
  return text.substring(0, maxLength) + '...'
}

export function extractDomain(url: string): string {
  try {
    return new URL(url).hostname
  } catch {
    return 'unknown'
  }
}

export function isValidUrl(string: string): boolean {
  try {
    new URL(string)
    return true
  } catch {
    return false
  }
}

export function isValidGitHubUrl(url: string): boolean {
  const githubPattern = /^https?:\/\/github\.com\/[^\/]+\/[^\/]+\/?$/
  return githubPattern.test(url)
}

export function extractRepoFromGitHubUrl(url: string): { owner: string; repo: string } | null {
  const match = url.match(/github\.com\/([^\/]+)\/([^\/]+)/)
  if (match) {
    return {
      owner: match[1],
      repo: match[2].replace('.git', '')
    }
  }
  return null
}

export function highlightSearchTerms(text: string, searchTerm: string): string {
  if (!searchTerm.trim()) return text
  
  const regex = new RegExp(`(${searchTerm.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')})`, 'gi')
  return text.replace(regex, '<mark>$1</mark>')
}

export function debounce<T extends (...args: any[]) => any>(
  func: T,
  wait: number
): (...args: Parameters<T>) => void {
  let timeout: NodeJS.Timeout | null = null
  
  return (...args: Parameters<T>) => {
    if (timeout) clearTimeout(timeout)
    timeout = setTimeout(() => func(...args), wait)
  }
}

export function throttle<T extends (...args: any[]) => any>(
  func: T,
  limit: number
): (...args: Parameters<T>) => void {
  let inThrottle: boolean
  
  return (...args: Parameters<T>) => {
    if (!inThrottle) {
      func(...args)
      inThrottle = true
      setTimeout(() => (inThrottle = false), limit)
    }
  }
}

export function generateId(): string {
  return Math.random().toString(36).substring(2) + Date.now().toString(36)
}

export function copyToClipboard(text: string): Promise<void> {
  if (navigator.clipboard && window.isSecureContext) {
    return navigator.clipboard.writeText(text)
  } else {
    // Fallback for older browsers
    const textArea = document.createElement('textarea')
    textArea.value = text
    textArea.style.position = 'absolute'
    textArea.style.left = '-999999px'
    document.body.prepend(textArea)
    textArea.select()
    
    try {
      document.execCommand('copy')
    } catch (error) {
      console.error('Failed to copy text: ', error)
    } finally {
      textArea.remove()
    }
    
    return Promise.resolve()
  }
}

export function scrollToBottom(element: HTMLElement, smooth: boolean = true): void {
  element.scrollTo({
    top: element.scrollHeight,
    behavior: smooth ? 'smooth' : 'auto'
  })
}

export function getFileExtension(filename: string): string {
  return filename.slice((filename.lastIndexOf('.') - 1 >>> 0) + 2).toLowerCase()
}

export function getFileIcon(filename: string): string {
  const ext = getFileExtension(filename)
  
  const iconMap: Record<string, string> = {
    pdf: '📄',
    doc: '📝',
    docx: '📝',
    txt: '📄',
    md: '📝',
    html: '🌐',
    htm: '🌐',
    js: '📜',
    ts: '📜',
    py: '🐍',
    java: '☕',
    cpp: '⚙️',
    c: '⚙️',
    go: '🐹',
    rs: '🦀',
    php: '🐘',
    rb: '💎',
    json: '📋',
    xml: '📋',
    yaml: '📋',
    yml: '📋',
    css: '🎨',
    scss: '🎨',
    less: '🎨',
  }
  
  return iconMap[ext] || '📄'
}

export function formatNumber(num: number): string {
  if (num >= 1000000) {
    return (num / 1000000).toFixed(1) + 'M'
  }
  if (num >= 1000) {
    return (num / 1000).toFixed(1) + 'K'
  }
  return num.toString()
}

export function getSourceIcon(source: string): string {
  const iconMap: Record<string, string> = {
    documents: '📄',
    web_research: '🌐',
    github_code: '👨‍💻',
    chat_history: '💬',
  }
  
  return iconMap[source] || '📄'
}

export function getLanguageIcon(language: string): string {
  const iconMap: Record<string, string> = {
    javascript: '📜',
    typescript: '📜',
    python: '🐍',
    java: '☕',
    cpp: '⚙️',
    c: '⚙️',
    go: '🐹',
    rust: '🦀',
    php: '🐘',
    ruby: '💎',
    swift: '🦉',
    kotlin: '🎯',
    scala: '⚡',
    r: '📊',
    sql: '🗃️',
    bash: '💻',
    html: '🌐',
    css: '🎨',
    scss: '🎨',
    less: '🎨',
    json: '📋',
    xml: '📋',
    yaml: '📋',
    markdown: '📝',
  }
  
  return iconMap[language.toLowerCase()] || '📄'
}
