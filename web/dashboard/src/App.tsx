// Krustron Dashboard - Main App Component
// Author: Anubhav Gain <anubhavg@infopercept.com>

import { Routes, Route, Navigate } from 'react-router-dom'
import { WebSocketProvider } from '@/hooks/useWebSocket'
import { useAuthStore } from '@/store/useStore'
import Layout from '@/components/Layout'
import Dashboard from '@/pages/Dashboard'
import Clusters from '@/pages/Clusters'
import Applications from '@/pages/Applications'
import Pipelines from '@/pages/Pipelines'
import Security from '@/pages/Security'
import Cost from '@/pages/Cost'
import Settings from '@/pages/Settings'
import Login from '@/pages/Login'

// Protected Route wrapper
function ProtectedRoute({ children }: { children: React.ReactNode }) {
  const { isAuthenticated } = useAuthStore()

  if (!isAuthenticated) {
    return <Navigate to="/login" replace />
  }

  return <>{children}</>
}

function App() {
  const { isAuthenticated, token } = useAuthStore()

  // WebSocket URL with auth token
  const wsUrl = token
    ? `${window.location.protocol === 'https:' ? 'wss:' : 'ws:'}//${window.location.host}/ws?token=${token}`
    : undefined

  return (
    <WebSocketProvider options={{ url: wsUrl }}>
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
    </WebSocketProvider>
  )
}

export default App
