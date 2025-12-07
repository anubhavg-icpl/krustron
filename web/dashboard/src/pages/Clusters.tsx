// Krustron Dashboard - Clusters Page
// Author: Anubhav Gain <anubhavg@infopercept.com>

import { useEffect, useState } from 'react'
import { Routes, Route, NavLink, useNavigate } from 'react-router-dom'
import { motion, AnimatePresence } from 'framer-motion'
import {
  Server,
  Plus,
  Search,
  Filter,
  MoreVertical,
  RefreshCw,
  Trash2,
  Settings,
  ExternalLink,
  Cpu,
  HardDrive,
  Box,
  CheckCircle,
  XCircle,
  AlertCircle,
  Activity,
} from 'lucide-react'
import { clsx } from 'clsx'
import { useWebSocketContext, useClusterEvents } from '@/hooks/useWebSocket'
import { useClustersStore } from '@/store/useStore'
import { Cluster, ClusterMetrics, WebSocketMessage } from '@/types'
import {
  AreaChart,
  Area,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
} from 'recharts'

// Cluster Card Component
function ClusterCard({ cluster }: { cluster: Cluster }) {
  const navigate = useNavigate()
  const [menuOpen, setMenuOpen] = useState(false)
  const { subscribe, isConnected } = useWebSocketContext()
  const [metrics, setMetrics] = useState<ClusterMetrics | null>(null)

  // Subscribe to cluster-specific metrics
  useEffect(() => {
    if (!isConnected) return

    const unsubscribe = subscribe(`cluster:${cluster.id}`, (message: WebSocketMessage) => {
      if (message.type === 'cluster.metrics') {
        setMetrics(message.payload as ClusterMetrics)
      }
    })

    return unsubscribe
  }, [cluster.id, subscribe, isConnected])

  const statusColor = {
    connected: 'status-healthy',
    disconnected: 'status-error',
    unknown: 'status-unknown',
  }[cluster.status]

  return (
    <motion.div
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      className="glass-card-hover p-6 cursor-pointer"
      onClick={() => navigate(`/clusters/${cluster.id}`)}
    >
      <div className="flex items-start justify-between">
        <div className="flex items-center gap-4">
          <div className="w-12 h-12 rounded-xl bg-primary-500/20 flex items-center justify-center">
            <Server className="w-6 h-6 text-primary-400" />
          </div>
          <div>
            <h3 className="text-lg font-semibold text-white">{cluster.name}</h3>
            <p className="text-sm text-gray-400">{cluster.server}</p>
          </div>
        </div>

        <div className="flex items-center gap-2">
          <span className={clsx('status-badge', statusColor)}>
            <span className={clsx(
              'w-2 h-2 rounded-full',
              cluster.status === 'connected' && 'bg-status-healthy pulse-dot pulse-dot-healthy',
              cluster.status === 'disconnected' && 'bg-status-error',
              cluster.status === 'unknown' && 'bg-status-unknown'
            )} />
            {cluster.status}
          </span>

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
                    <RefreshCw className="w-4 h-4" />
                    Refresh
                  </button>
                  <button className="dropdown-item flex items-center gap-2 w-full">
                    <Settings className="w-4 h-4" />
                    Settings
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

      {/* Metrics */}
      <div className="grid grid-cols-4 gap-4 mt-6">
        <div>
          <div className="flex items-center gap-2 text-gray-400 text-xs mb-1">
            <Cpu className="w-3 h-3" />
            CPU
          </div>
          <div className="text-lg font-semibold text-white">
            {metrics?.cpu.percentage.toFixed(1) ?? '—'}%
          </div>
        </div>
        <div>
          <div className="flex items-center gap-2 text-gray-400 text-xs mb-1">
            <HardDrive className="w-3 h-3" />
            Memory
          </div>
          <div className="text-lg font-semibold text-white">
            {metrics?.memory.percentage.toFixed(1) ?? '—'}%
          </div>
        </div>
        <div>
          <div className="flex items-center gap-2 text-gray-400 text-xs mb-1">
            <Box className="w-3 h-3" />
            Pods
          </div>
          <div className="text-lg font-semibold text-white">
            {metrics?.pods.running ?? '—'}/{metrics?.pods.total ?? '—'}
          </div>
        </div>
        <div>
          <div className="flex items-center gap-2 text-gray-400 text-xs mb-1">
            <Server className="w-3 h-3" />
            Nodes
          </div>
          <div className="text-lg font-semibold text-white">
            {cluster.nodeCount}
          </div>
        </div>
      </div>

      {/* Labels */}
      {Object.keys(cluster.labels).length > 0 && (
        <div className="flex flex-wrap gap-2 mt-4">
          {Object.entries(cluster.labels).slice(0, 3).map(([key, value]) => (
            <span
              key={key}
              className="px-2 py-1 bg-glass-light rounded-lg text-xs text-gray-400"
            >
              {key}: {value}
            </span>
          ))}
          {Object.keys(cluster.labels).length > 3 && (
            <span className="px-2 py-1 bg-glass-light rounded-lg text-xs text-gray-400">
              +{Object.keys(cluster.labels).length - 3} more
            </span>
          )}
        </div>
      )}
    </motion.div>
  )
}

// Clusters List
function ClustersList() {
  const { clusters, loading, filter, setFilter } = useClustersStore()
  const [searchQuery, setSearchQuery] = useState('')

  // Filter clusters
  const filteredClusters = clusters.filter(cluster => {
    if (searchQuery && !cluster.name.toLowerCase().includes(searchQuery.toLowerCase())) {
      return false
    }
    if (filter.status?.length && !filter.status.includes(cluster.status)) {
      return false
    }
    return true
  })

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold text-white">Clusters</h1>
          <p className="text-gray-400 mt-1">Manage your Kubernetes clusters</p>
        </div>
        <button className="glass-btn-primary flex items-center gap-2">
          <Plus className="w-4 h-4" />
          Add Cluster
        </button>
      </div>

      {/* Filters */}
      <div className="flex flex-col sm:flex-row gap-4">
        <div className="relative flex-1">
          <Search className="absolute left-4 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400" />
          <input
            type="text"
            placeholder="Search clusters..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="glass-input pl-11"
          />
        </div>
        <button className="glass-btn flex items-center gap-2">
          <Filter className="w-4 h-4" />
          Filters
        </button>
      </div>

      {/* Clusters Grid */}
      {loading ? (
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
          {[1, 2, 3, 4].map((i) => (
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
      ) : filteredClusters.length === 0 ? (
        <div className="glass-card p-12 text-center">
          <Server className="w-12 h-12 text-gray-500 mx-auto mb-4" />
          <h3 className="text-lg font-semibold text-white mb-2">No clusters found</h3>
          <p className="text-gray-400 mb-4">
            {searchQuery
              ? 'No clusters match your search criteria'
              : 'Get started by adding your first cluster'}
          </p>
          <button className="glass-btn-primary">
            <Plus className="w-4 h-4 mr-2" />
            Add Cluster
          </button>
        </div>
      ) : (
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
          {filteredClusters.map((cluster) => (
            <ClusterCard key={cluster.id} cluster={cluster} />
          ))}
        </div>
      )}
    </div>
  )
}

// Cluster Detail Page
function ClusterDetail() {
  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-bold text-white">Cluster Details</h1>
      <p className="text-gray-400">Cluster detail view with real-time metrics</p>
      {/* Add cluster detail components here */}
    </div>
  )
}

// Main Clusters Page with Routes
export default function Clusters() {
  return (
    <Routes>
      <Route index element={<ClustersList />} />
      <Route path=":id/*" element={<ClusterDetail />} />
    </Routes>
  )
}
