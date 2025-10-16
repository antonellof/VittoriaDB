// This hook is deprecated - WebSocket notifications handle all stats updates automatically
// Keeping file for backward compatibility, but all functions are no-ops

export function useAutoRefresh() {
  // No-op function - WebSocket handles all updates
  const triggerStatsRefresh = () => {
    console.log('📡 Stats are now updated automatically via WebSocket - no manual refresh needed')
  }

  return { triggerStatsRefresh }
}

// Deprecated global functions - no-ops for backward compatibility
export const setGlobalStatsRefresh = (refreshFn: () => void) => {
  // No-op - WebSocket handles updates
}

export const triggerGlobalStatsRefresh = () => {
  // No-op - WebSocket handles updates
}