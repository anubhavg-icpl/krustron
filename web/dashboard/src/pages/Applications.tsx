// Krustron Dashboard - Applications Page
// Author: Anubhav Gain <anubhavg@infopercept.com>

import { useState, useEffect } from 'react'
import { Routes, Route, useNavigate } from 'react-router-dom'
import { motion, AnimatePresence } from 'framer-motion'
import {
  Box,
  Plus,
  Search,
  Filter,
  MoreVertical,
  RefreshCw,
  Trash2,
  Play,
  GitBranch,
  ExternalLink,
  CheckCircle,
  XCircle,
  AlertCircle,
  Clock,
  Loader2,
} from 'lucide-react'
import { clsx } from 'clsx'
import { useWebSocketContext } from '@/hooks/useWebSocket'
import { useApplicationsStore } from '@/store/useStore'
import { Application, WebSocketMessage } from '@/types'

// Application Card Component
function ApplicationCard({ app }: { app: Application }) {
  const navigate = useNavigate()
  const [menuOpen, setMenuOpen] = useState(false)
  const { subscribe, isConnected } = useWebSocketContext()

  // Subscribe to application-specific updates
  useEffect(() => {
    if (!isConnected) return

    const unsubscribe = subscribe(`app:${app.id}`, (message: WebSocketMessage) => {
      // Handle real-time updates
      console.log('App update:', message)
    })

    return unsubscribe
  }, [app.id, subscribe, isConnected])

  const syncStatusStyles = {
    Synced: 'status-synced',
    OutOfSync: 'status-out-of-sync',
    Unknown: 'status-unknown',
  }[app.syncStatus]

  const healthStatusStyles = {
    Healthy: 'status-healthy',
    Progressing: 'status-progressing',
    Degraded: 'status-error',
    Suspended: 'status-warning',
    Missing: 'status-error',
    Unknown: 'status-unknown',
  }[app.healthStatus]

  const healthIcon = {
    Healthy: CheckCircle,
    Progressing: Loader2,
    Degraded: XCircle,
    Suspended: Clock,
    Missing: AlertCircle,
    Unknown: AlertCircle,
  }[app.healthStatus]

  const HealthIcon = healthIcon

  return (
    <motion.div
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      className="glass-card-hover p-6 cursor-pointer"
      onClick={() => navigate(`/applications/${app.id}`)}
    >
      <div className="flex items-start justify-between">
        <div className="flex items-center gap-4">
          <div className="w-12 h-12 rounded-xl bg-accent-500/20 flex items-center justify-center">
            <Box className="w-6 h-6 text-accent-400" />
          </div>
          <div>
            <h3 className="text-lg font-semibold text-white">{app.name}</h3>
            <p className="text-sm text-gray-400">{app.namespace} / {app.clusterName}</p>
          </div>
        </div>

        <div className="flex items-center gap-2">
          <div className="relative">
            <button
              onClick={(e) => {
                e.stopPropagation()
                setMenuOpen(!menuOpen)
              }}
              className="p-2 rounded-lg hover:bg-glass-light transition-colors"
            >
              <MoreVertical className="w-4 h-4 text-gray-400" />
            </button>

            <AnimatePresence>
              {menuOpen && (
                <motion.div
                  initial={{ opacity: 0, y: 10 }}
                  animate={{ opacity: 1, y: 0 }}
                  exit={{ opacity: 0, y: 10 }}
                  className="dropdown-menu"
                  onClick={(e) => e.stopPropagation()}
                >
                  <button className="dropdown-item flex items-center gap-2 w-full">
                    <Play className="w-4 h-4" />
                    Sync
                  </button>
                  <button className="dropdown-item flex items-center gap-2 w-full">
                    <RefreshCw className="w-4 h-4" />
                    Hard Refresh
                  </button>
                  <button className="dropdown-item flex items-center gap-2 w-full">
                    <ExternalLink className="w-4 h-4" />
                    Open in browser
                  </button>
                  <hr className="border-glass-border my-1" />
                  <button className="dropdown-item flex items-center gap-2 w-full text-status-error">
                    <Trash2 className="w-4 h-4" />
                    Delete
                  </button>
                </motion.div>
              )}
            </AnimatePresence>
          </div>
        </div>
      </div>

      {/* Status Badges */}
      <div className="flex items-center gap-3 mt-4">
        <span className={clsx('status-badge', syncStatusStyles)}>
          <RefreshCw className={clsx(
            'w-3 h-3',
            app.syncStatus === 'OutOfSync' && 'animate-spin'
          )} />
          {app.syncStatus}
        </span>
        <span className={clsx('status-badge', healthStatusStyles)}>
          <HealthIcon className={clsx(
            'w-3 h-3',
            app.healthStatus === 'Progressing' && 'animate-spin'
          )} />
          {app.healthStatus}
        </span>
      </div>

      {/* Source Info */}
      <div className="mt-4 p-3 bg-glass-light rounded-xl">
        <div className="flex items-center gap-2 text-sm">
          <GitBranch className="w-4 h-4 text-gray-400" />
          <span className="text-gray-300 truncate">{app.source.repoUrl}</span>
        </div>
        <div className="flex items-center justify-between mt-2 text-xs text-gray-400">
          <span>{app.source.path || '/'}</span>
          <span>{app.source.targetRevision}</span>
        </div>
      </div>

      {/* Resources Summary */}
      <div className="flex items-center justify-between mt-4 text-sm">
        <span className="text-gray-400">
          {app.resources.length} resources
        </span>
        <div className="flex items-center gap-2">
          {app.resources.filter(r => r.health?.status === 'Healthy').length > 0 && (
            <span className="flex items-center gap-1 text-status-healthy">
              <CheckCircle className="w-3 h-3" />
              {app.resources.filter(r => r.health?.status === 'Healthy').length}
            </span>
          )}
          {app.resources.filter(r => r.health?.status === 'Degraded').length > 0 && (
            <span className="flex items-center gap-1 text-status-error">
              <XCircle className="w-3 h-3" />
              {app.resources.filter(r => r.health?.status === 'Degraded').length}
            </span>
          )}
          {app.resources.filter(r => r.health?.status === 'Progressing').length > 0 && (
            <span className="flex items-center gap-1 text-status-progressing">
              <Loader2 className="w-3 h-3 animate-spin" />
              {app.resources.filter(r => r.health?.status === 'Progressing').length}
            </span>
          )}
        </div>
      </div>
    </motion.div>
  )
}

// Applications List
function ApplicationsList() {
  const { applications, loading, filter, setFilter } = useApplicationsStore()
  const [searchQuery, setSearchQuery] = useState('')
  const [statusFilter, setStatusFilter] = useState<string[]>([])

  // Filter applications
  const filteredApps = applications.filter(app => {
    if (searchQuery) {
      const query = searchQuery.toLowerCase()
      if (!app.name.toLowerCase().includes(query) &&
          !app.namespace.toLowerCase().includes(query)) {
        return false
      }
    }
    if (statusFilter.length > 0 && !statusFilter.includes(app.syncStatus)) {
      return false
    }
    return true
  })

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold text-white">Applications</h1>
          <p className="text-gray-400 mt-1">GitOps application deployments</p>
        </div>
        <button className="glass-btn-primary flex items-center gap-2">
          <Plus className="w-4 h-4" />
          New Application
        </button>
      </div>

      {/* Filters */}
      <div className="flex flex-col sm:flex-row gap-4">
        <div className="relative flex-1">
          <Search className="absolute left-4 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400" />
          <input
            type="text"
            placeholder="Search applications..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="glass-input pl-11"
          />
        </div>
        <div className="flex gap-2">
          <button
            onClick={() => setStatusFilter([])}
            className={clsx(
              'px-3 py-2 rounded-xl text-sm transition-colors',
              statusFilter.length === 0 ? 'bg-primary-500/20 text-primary-400' : 'text-gray-400 hover:bg-glass-light'
            )}
          >
            All
          </button>
          <button
            onClick={() => setStatusFilter(['Synced'])}
            className={clsx(
              'px-3 py-2 rounded-xl text-sm transition-colors',
              statusFilter.includes('Synced') ? 'bg-status-synced/20 text-status-synced' : 'text-gray-400 hover:bg-glass-light'
            )}
          >
            Synced
          </button>
          <button
            onClick={() => setStatusFilter(['OutOfSync'])}
            className={clsx(
              'px-3 py-2 rounded-xl text-sm transition-colors',
              statusFilter.includes('OutOfSync') ? 'bg-status-outOfSync/20 text-status-outOfSync' : 'text-gray-400 hover:bg-glass-light'
            )}
          >
            OutOfSync
          </button>
        </div>
      </div>

      {/* Summary Stats */}
      <div className="grid grid-cols-2 sm:grid-cols-4 gap-4">
        <div className="glass-card p-4">
          <div className="text-2xl font-bold text-white">{applications.length}</div>
          <div className="text-sm text-gray-400">Total Apps</div>
        </div>
        <div className="glass-card p-4">
          <div className="text-2xl font-bold text-status-synced">
            {applications.filter(a => a.syncStatus === 'Synced').length}
          </div>
          <div className="text-sm text-gray-400">Synced</div>
        </div>
        <div className="glass-card p-4">
          <div className="text-2xl font-bold text-status-outOfSync">
            {applications.filter(a => a.syncStatus === 'OutOfSync').length}
          </div>
          <div className="text-sm text-gray-400">Out of Sync</div>
        </div>
        <div className="glass-card p-4">
          <div className="text-2xl font-bold text-status-healthy">
            {applications.filter(a => a.healthStatus === 'Healthy').length}
          </div>
          <div className="text-sm text-gray-400">Healthy</div>
        </div>
      </div>

      {/* Applications Grid */}
      {loading ? (
        <div className="grid grid-cols-1 lg:grid-cols-2 xl:grid-cols-3 gap-6">
          {[1, 2, 3, 4, 5, 6].map((i) => (
            <div key={i} className="glass-card p-6 animate-pulse">
              <div className="flex items-center gap-4">
                <div className="w-12 h-12 rounded-xl skeleton" />
                <div className="flex-1">
                  <div className="h-5 w-32 skeleton mb-2" />
                  <div className="h-4 w-48 skeleton" />
                </div>
              </div>
            </div>
          ))}
        </div>
      ) : filteredApps.length === 0 ? (
        <div className="glass-card p-12 text-center">
          <Box className="w-12 h-12 text-gray-500 mx-auto mb-4" />
          <h3 className="text-lg font-semibold text-white mb-2">No applications found</h3>
          <p className="text-gray-400 mb-4">
            {searchQuery
              ? 'No applications match your search criteria'
              : 'Get started by creating your first application'}
          </p>
          <button className="glass-btn-primary">
            <Plus className="w-4 h-4 mr-2" />
            New Application
          </button>
        </div>
      ) : (
        <div className="grid grid-cols-1 lg:grid-cols-2 xl:grid-cols-3 gap-6">
          {filteredApps.map((app) => (
            <ApplicationCard key={app.id} app={app} />
          ))}
        </div>
      )}
    </div>
  )
}

// Application Detail Page
function ApplicationDetail() {
  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-bold text-white">Application Details</h1>
      <p className="text-gray-400">Application detail view with sync status and resources</p>
    </div>
  )
}

// Main Applications Page with Routes
export default function Applications() {
  return (
    <Routes>
      <Route index element={<ApplicationsList />} />
      <Route path=":id/*" element={<ApplicationDetail />} />
    </Routes>
  )
}
