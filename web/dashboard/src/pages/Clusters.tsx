// Krustron Dashboard - Clusters Page
// Author: Anubhav Gain <anubhavg@infopercept.com>

import { useEffect, useState } from 'react'
import { Routes, Route, useNavigate } from 'react-router-dom'
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
  Loader2,
  Upload,
  FileText,
} from 'lucide-react'
import { clsx } from 'clsx'
import { useWebSocketContext } from '@/hooks/useWebSocket'
import { useClustersStore } from '@/store/useStore'
import { Cluster, ClusterMetrics, WebSocketMessage } from '@/types'

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
  const navigate = useNavigate()
  const { clusters, loading, filter } = useClustersStore()
  const [searchQuery, setSearchQuery] = useState('')

  const handleAddCluster = () => {
    navigate('/clusters/new')
  }

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
        <button onClick={handleAddCluster} className="glass-btn-primary flex items-center gap-2">
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
          <button onClick={handleAddCluster} className="glass-btn-primary">
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

// Cluster Create Page
function ClusterCreate() {
  const navigate = useNavigate()
  const [name, setName] = useState('')
  const [server, setServer] = useState('')
  const [description, setDescription] = useState('')
  const [kubeconfig, setKubeconfig] = useState('')
  const [connectionType, setConnectionType] = useState<'kubeconfig' | 'serviceaccount'>('kubeconfig')
  const [isSubmitting, setIsSubmitting] = useState(false)

  const handleFileUpload = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    if (file) {
      const reader = new FileReader()
      reader.onload = (event) => {
        setKubeconfig(event.target?.result as string)
      }
      reader.readAsText(file)
    }
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setIsSubmitting(true)

    try {
      const response = await fetch('/api/v1/clusters', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${localStorage.getItem('token')}`,
        },
        body: JSON.stringify({
          name,
          server,
          description,
          kubeconfig: connectionType === 'kubeconfig' ? kubeconfig : undefined,
        }),
      })

      if (response.ok) {
        navigate('/clusters')
      }
    } catch (error) {
      console.error('Failed to create cluster:', error)
    } finally {
      setIsSubmitting(false)
    }
  }

  return (
    <div className="space-y-6 max-w-2xl">
      <div>
        <h1 className="text-2xl font-bold text-white">Add Cluster</h1>
        <p className="text-gray-400 mt-1">Connect a new Kubernetes cluster</p>
      </div>

      <form onSubmit={handleSubmit} className="glass-card p-6 space-y-6">
        <div>
          <label className="block text-sm font-medium text-gray-300 mb-2">
            Cluster Name
          </label>
          <input
            type="text"
            value={name}
            onChange={(e) => setName(e.target.value)}
            placeholder="e.g., production-cluster"
            className="glass-input"
            required
          />
        </div>

        <div>
          <label className="block text-sm font-medium text-gray-300 mb-2">
            API Server URL
          </label>
          <input
            type="text"
            value={server}
            onChange={(e) => setServer(e.target.value)}
            placeholder="https://kubernetes.example.com:6443"
            className="glass-input"
            required
          />
        </div>

        <div>
          <label className="block text-sm font-medium text-gray-300 mb-2">
            Description
          </label>
          <textarea
            value={description}
            onChange={(e) => setDescription(e.target.value)}
            placeholder="Cluster description..."
            className="glass-input min-h-[80px]"
            rows={2}
          />
        </div>

        <div>
          <label className="block text-sm font-medium text-gray-300 mb-2">
            Connection Method
          </label>
          <div className="flex gap-4">
            <button
              type="button"
              onClick={() => setConnectionType('kubeconfig')}
              className={clsx(
                'flex-1 p-4 rounded-xl border transition-all',
                connectionType === 'kubeconfig'
                  ? 'border-primary-500 bg-primary-500/10'
                  : 'border-glass-border bg-glass-light hover:border-gray-600'
              )}
            >
              <FileText className={clsx(
                'w-6 h-6 mx-auto mb-2',
                connectionType === 'kubeconfig' ? 'text-primary-400' : 'text-gray-400'
              )} />
              <div className={clsx(
                'text-sm font-medium',
                connectionType === 'kubeconfig' ? 'text-white' : 'text-gray-400'
              )}>
                Kubeconfig
              </div>
            </button>
            <button
              type="button"
              onClick={() => setConnectionType('serviceaccount')}
              className={clsx(
                'flex-1 p-4 rounded-xl border transition-all',
                connectionType === 'serviceaccount'
                  ? 'border-primary-500 bg-primary-500/10'
                  : 'border-glass-border bg-glass-light hover:border-gray-600'
              )}
            >
              <Server className={clsx(
                'w-6 h-6 mx-auto mb-2',
                connectionType === 'serviceaccount' ? 'text-primary-400' : 'text-gray-400'
              )} />
              <div className={clsx(
                'text-sm font-medium',
                connectionType === 'serviceaccount' ? 'text-white' : 'text-gray-400'
              )}>
                Service Account
              </div>
            </button>
          </div>
        </div>

        {connectionType === 'kubeconfig' && (
          <div>
            <label className="block text-sm font-medium text-gray-300 mb-2">
              Kubeconfig File
            </label>
            <div className="border-2 border-dashed border-glass-border rounded-xl p-6 text-center hover:border-gray-600 transition-colors">
              <input
                type="file"
                accept=".yaml,.yml,.conf"
                onChange={handleFileUpload}
                className="hidden"
                id="kubeconfig-upload"
              />
              <label htmlFor="kubeconfig-upload" className="cursor-pointer">
                <Upload className="w-8 h-8 text-gray-400 mx-auto mb-2" />
                <p className="text-sm text-gray-400">
                  {kubeconfig ? 'File uploaded' : 'Click to upload kubeconfig'}
                </p>
                <p className="text-xs text-gray-500 mt-1">
                  Supports .yaml, .yml, .conf files
                </p>
              </label>
            </div>
            {kubeconfig && (
              <div className="mt-2 p-2 bg-glass-light rounded-lg">
                <p className="text-xs text-status-healthy">Kubeconfig loaded successfully</p>
              </div>
            )}
          </div>
        )}

        {connectionType === 'serviceaccount' && (
          <div className="p-4 bg-glass-light rounded-xl">
            <p className="text-sm text-gray-400">
              To connect using a Service Account, install the Krustron agent in your cluster:
            </p>
            <pre className="mt-2 p-3 bg-glass-dark rounded-lg text-xs text-gray-300 overflow-x-auto">
              kubectl apply -f https://krustron.io/agent/install.yaml
            </pre>
          </div>
        )}

        <div className="flex gap-4 pt-4">
          <button
            type="button"
            onClick={() => navigate('/clusters')}
            className="glass-btn flex-1"
          >
            Cancel
          </button>
          <button
            type="submit"
            disabled={isSubmitting || !name || !server}
            className="glass-btn-primary flex-1 flex items-center justify-center gap-2"
          >
            {isSubmitting ? (
              <>
                <Loader2 className="w-4 h-4 animate-spin" />
                Adding...
              </>
            ) : (
              <>
                <Plus className="w-4 h-4" />
                Add Cluster
              </>
            )}
          </button>
        </div>
      </form>
    </div>
  )
}

// Main Clusters Page with Routes
export default function Clusters() {
  return (
    <Routes>
      <Route index element={<ClustersList />} />
      <Route path="new" element={<ClusterCreate />} />
      <Route path=":id/*" element={<ClusterDetail />} />
    </Routes>
  )
}
