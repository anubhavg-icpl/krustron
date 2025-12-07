// Krustron Dashboard - Main App Component
// Author: Anubhav Gain <anubhavg@infopercept.com>

import { useEffect, useState } from 'react'
import { Routes, Route, Navigate } from 'react-router-dom'
import { WebSocketProvider } from '@/hooks/useWebSocket'
import { useAuthStore } from '@/store/useStore'
import { useInitializeData } from '@/hooks/useDataFetching'
import { ErrorBoundary } from '@/components/ErrorBoundary'
import Layout from '@/components/Layout'
import Dashboard from '@/pages/Dashboard'
import Clusters from '@/pages/Clusters'
import Applications from '@/pages/Applications'
import Pipelines from '@/pages/Pipelines'
import Security from '@/pages/Security'
import Cost from '@/pages/Cost'
import Settings from '@/pages/Settings'
import Login from '@/pages/Login'

// ============================================================================
// Data Initializer Component
// ============================================================================

function DataInitializer({ children }: { children: React.ReactNode }) {
  // This hook fetches initial data on mount
  useInitializeData()
  return <>{children}</>
}

// ============================================================================
// WebSocket Wrapper (handles reconnection on auth change)
// ============================================================================

function WebSocketWrapper({ children }: { children: React.ReactNode }) {
  const { token, isAuthenticated } = useAuthStore()
  const [wsKey, setWsKey] = useState(0)

  // Force WebSocket reconnection when token changes
  useEffect(() => {
    if (isAuthenticated && token) {
      setWsKey((prev) => prev + 1)
    }
  }, [token, isAuthenticated])

  // Don't pass token in URL - use cookie or handle in onOpen
  // For now, we'll use a simpler approach that doesn't expose token
  const wsUrl = isAuthenticated
    ? `${window.location.protocol === 'https:' ? 'wss:' : 'ws:'}//${window.location.host}/ws`
    : undefined

  if (!isAuthenticated) {
    return <>{children}</>
  }

  return (
    <WebSocketProvider
      key={wsKey}
      options={{
        url: wsUrl,
        onOpen: () => {
          console.log('[WebSocket] Connected')
        },
        onClose: () => {
          console.log('[WebSocket] Disconnected')
        },
      }}
    >
      <DataInitializer>{children}</DataInitializer>
    </WebSocketProvider>
  )
}

// ============================================================================
// Protected Route Wrapper
// ============================================================================

function ProtectedRoute({ children }: { children: React.ReactNode }) {
  const { isAuthenticated } = useAuthStore()

  if (!isAuthenticated) {
    return <Navigate to="/login" replace />
  }

  return <>{children}</>
}

// ============================================================================
// App Component
// ============================================================================

function App() {
  return (
    <ErrorBoundary>
      <WebSocketWrapper>
        <Routes>
          {/* Public routes */}
          <Route path="/login" element={<Login />} />

          {/* Protected routes */}
          <Route
            path="/"
            element={
              <ProtectedRoute>
                <Layout />
              </ProtectedRoute>
            }
          >
            <Route index element={<Dashboard />} />
            <Route path="clusters/*" element={<Clusters />} />
            <Route path="applications/*" element={<Applications />} />
            <Route path="pipelines/*" element={<Pipelines />} />
            <Route path="security/*" element={<Security />} />
            <Route path="cost/*" element={<Cost />} />
            <Route path="settings/*" element={<Settings />} />
          </Route>

          {/* Fallback */}
          <Route path="*" element={<Navigate to="/" replace />} />
        </Routes>
      </WebSocketWrapper>
    </ErrorBoundary>
  )
}

export default App
