// Krustron Dashboard - Pipelines Page
// Author: Anubhav Gain <anubhavg@infopercept.com>

import { useState, useEffect } from 'react'
import { Routes, Route, useNavigate } from 'react-router-dom'
import { motion, AnimatePresence } from 'framer-motion'
import {
  GitBranch,
  Plus,
  Search,
  Play,
  Pause,
  MoreVertical,
  Trash2,
  Settings,
  CheckCircle,
  XCircle,
  Clock,
  Loader2,
  ChevronRight,
  Terminal,
} from 'lucide-react'
import { clsx } from 'clsx'
import { useWebSocketContext } from '@/hooks/useWebSocket'
import { usePipelinesStore } from '@/store/useStore'
import { Pipeline, PipelineRun, WebSocketMessage } from '@/types'
import { showSuccessToast, showErrorToast, showInfoToast } from '@/hooks/useNotificationToasts'
import { pipelinesApi, ApiClientError } from '@/api'

// Stage Status Icon
function StageStatusIcon({ status }: { status: string }) {
  const icons = {
    Pending: Clock,
    Running: Loader2,
    Succeeded: CheckCircle,
    Failed: XCircle,
    Cancelled: XCircle,
  }
  const Icon = icons[status as keyof typeof icons] || Clock

  return (
    <Icon className={clsx(
      'w-4 h-4',
      status === 'Running' && 'animate-spin text-status-progressing',
      status === 'Succeeded' && 'text-status-healthy',
      status === 'Failed' && 'text-status-error',
      status === 'Pending' && 'text-gray-400',
      status === 'Cancelled' && 'text-gray-400'
    )} />
  )
}

// Pipeline Stages Visualization
function PipelineStages({ stages, run }: { stages: Pipeline['stages']; run?: PipelineRun }) {
  return (
    <div className="flex items-center gap-1 overflow-x-auto py-2">
      {stages.map((stage, index) => {
        const stageRun = run?.stages.find(s => s.name === stage.name)
        const status = stageRun?.status || 'Pending'

        return (
          <div key={stage.name} className="flex items-center">
            <div className={clsx(
              'flex items-center gap-2 px-3 py-1.5 rounded-lg text-xs font-medium whitespace-nowrap',
              status === 'Running' && 'bg-status-progressing/20 text-status-progressing',
              status === 'Succeeded' && 'bg-status-healthy/20 text-status-healthy',
              status === 'Failed' && 'bg-status-error/20 text-status-error',
              status === 'Pending' && 'bg-glass-light text-gray-400',
              status === 'Cancelled' && 'bg-glass-light text-gray-400 line-through'
            )}>
              <StageStatusIcon status={status} />
              {stage.name}
            </div>
            {index < stages.length - 1 && (
              <ChevronRight className="w-4 h-4 text-gray-600 mx-1" />
            )}
          </div>
        )
      })}
    </div>
  )
}

// Pipeline Card Component
function PipelineCard({ pipeline }: { pipeline: Pipeline }) {
  const navigate = useNavigate()
  const [menuOpen, setMenuOpen] = useState(false)
  const { subscribe, send, isConnected } = useWebSocketContext()

  // Subscribe to pipeline-specific updates
  useEffect(() => {
    if (!isConnected) return

    const unsubscribe = subscribe(`pipeline:${pipeline.id}`, (message: WebSocketMessage) => {
      console.log('Pipeline update:', message)
    })

    return unsubscribe
  }, [pipeline.id, subscribe, isConnected])

  const lastRunStatus = pipeline.lastRun?.status || 'Never Run'
  const statusStyles = {
    Pending: 'status-info',
    Running: 'status-progressing',
    Succeeded: 'status-healthy',
    Failed: 'status-error',
    Cancelled: 'status-unknown',
    'Never Run': 'status-unknown',
  }[lastRunStatus]

  const handleTrigger = (e: React.MouseEvent) => {
    e.stopPropagation()
    send({
      type: 'pipeline.trigger',
      channel: `pipeline:${pipeline.id}`,
      payload: { pipelineId: pipeline.id },
    })
    showInfoToast('Pipeline triggered', `${pipeline.name} is starting...`)
  }

  return (
    <motion.div
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      className="glass-card-hover p-6 cursor-pointer"
      onClick={() => navigate(`/pipelines/${pipeline.id}`)}
    >
      <div className="flex items-start justify-between">
        <div className="flex items-center gap-4">
          <div className="w-12 h-12 rounded-xl bg-status-info/20 flex items-center justify-center">
            <GitBranch className="w-6 h-6 text-status-info" />
          </div>
          <div>
            <h3 className="text-lg font-semibold text-white">{pipeline.name}</h3>
            <p className="text-sm text-gray-400">
              {pipeline.applicationName || 'No linked application'}
            </p>
          </div>
        </div>

        <div className="flex items-center gap-2">
          <button
            onClick={handleTrigger}
            className="p-2 rounded-lg bg-primary-500/20 hover:bg-primary-500/30 transition-colors"
          >
            <Play className="w-4 h-4 text-primary-400" />
          </button>

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
                    <Terminal className="w-4 h-4" />
                    View Logs
                  </button>
                  <button className="dropdown-item flex items-center gap-2 w-full">
                    <Settings className="w-4 h-4" />
                    Settings
                  </button>
                  <button className="dropdown-item flex items-center gap-2 w-full">
                    <Pause className="w-4 h-4" />
                    Disable
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

      {/* Stages */}
      <div className="mt-4">
        <PipelineStages stages={pipeline.stages} run={pipeline.lastRun} />
      </div>

      {/* Stats */}
      <div className="flex items-center justify-between mt-4 pt-4 border-t border-glass-border">
        <div className="flex items-center gap-4">
          <span className={clsx('status-badge', statusStyles)}>
            <StageStatusIcon status={lastRunStatus} />
            {lastRunStatus}
          </span>
        </div>
        <div className="flex items-center gap-4 text-sm text-gray-400">
          <span className="flex items-center gap-1">
            <CheckCircle className="w-3 h-3 text-status-healthy" />
            {pipeline.successCount}
          </span>
          <span className="flex items-center gap-1">
            <XCircle className="w-3 h-3 text-status-error" />
            {pipeline.failureCount}
          </span>
          <span>Total: {pipeline.runCount}</span>
        </div>
      </div>

      {/* Last Run Time */}
      {pipeline.lastRun && (
        <div className="mt-3 text-xs text-gray-500">
          Last run: {new Date(pipeline.lastRun.startedAt).toLocaleString()}
        </div>
      )}
    </motion.div>
  )
}

// Pipelines List
function PipelinesList() {
  const navigate = useNavigate()
  const { pipelines, loading } = usePipelinesStore()
  const [searchQuery, setSearchQuery] = useState('')

  const filteredPipelines = pipelines.filter(p =>
    p.name.toLowerCase().includes(searchQuery.toLowerCase())
  )

  const handleCreatePipeline = () => {
    navigate('/pipelines/new')
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold text-white">Pipelines</h1>
          <p className="text-gray-400 mt-1">CI/CD pipeline management</p>
        </div>
        <button onClick={handleCreatePipeline} className="glass-btn-primary flex items-center gap-2">
          <Plus className="w-4 h-4" />
          Create Pipeline
        </button>
      </div>

      {/* Search */}
      <div className="relative">
        <Search className="absolute left-4 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400" />
        <input
          type="text"
          placeholder="Search pipelines..."
          value={searchQuery}
          onChange={(e) => setSearchQuery(e.target.value)}
          className="glass-input pl-11"
        />
      </div>

      {/* Summary */}
      <div className="grid grid-cols-2 sm:grid-cols-4 gap-4">
        <div className="glass-card p-4">
          <div className="text-2xl font-bold text-white">{pipelines.length}</div>
          <div className="text-sm text-gray-400">Total Pipelines</div>
        </div>
        <div className="glass-card p-4">
          <div className="text-2xl font-bold text-status-progressing">
            {pipelines.filter(p => p.lastRun?.status === 'Running').length}
          </div>
          <div className="text-sm text-gray-400">Running</div>
        </div>
        <div className="glass-card p-4">
          <div className="text-2xl font-bold text-status-healthy">
            {pipelines.reduce((acc, p) => acc + p.successCount, 0)}
          </div>
          <div className="text-sm text-gray-400">Successful Runs</div>
        </div>
        <div className="glass-card p-4">
          <div className="text-2xl font-bold text-status-error">
            {pipelines.reduce((acc, p) => acc + p.failureCount, 0)}
          </div>
          <div className="text-sm text-gray-400">Failed Runs</div>
        </div>
      </div>

      {/* Pipelines Grid */}
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
      ) : filteredPipelines.length === 0 ? (
        <div className="glass-card p-12 text-center">
          <GitBranch className="w-12 h-12 text-gray-500 mx-auto mb-4" />
          <h3 className="text-lg font-semibold text-white mb-2">No pipelines found</h3>
          <p className="text-gray-400 mb-4">
            Create your first CI/CD pipeline to automate deployments
          </p>
          <button onClick={handleCreatePipeline} className="glass-btn-primary">
            <Plus className="w-4 h-4 mr-2" />
            Create Pipeline
          </button>
        </div>
      ) : (
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
          {filteredPipelines.map((pipeline) => (
            <PipelineCard key={pipeline.id} pipeline={pipeline} />
          ))}
        </div>
      )}
    </div>
  )
}

// Pipeline Detail Page
function PipelineDetail() {
  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-bold text-white">Pipeline Details</h1>
      <p className="text-gray-400">Pipeline detail view with run history and logs</p>
    </div>
  )
}

// Pipeline Create Page
function PipelineCreate() {
  const navigate = useNavigate()
  const [name, setName] = useState('')
  const [description, setDescription] = useState('')
  const [gitRepo, setGitRepo] = useState('')
  const [branch, setBranch] = useState('main')
  const [isSubmitting, setIsSubmitting] = useState(false)

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setIsSubmitting(true)

    try {
      await pipelinesApi.create({
        name,
        description,
        git_repository: gitRepo,
        branch,
        stages: [
          { name: 'Build', type: 'build' },
          { name: 'Test', type: 'test' },
          { name: 'Deploy', type: 'deploy' },
        ],
      })
      showSuccessToast('Pipeline created', `${name} is ready to run`)
      navigate('/pipelines')
    } catch (error) {
      const message = error instanceof ApiClientError ? error.message : 'Network error. Please try again.'
      showErrorToast('Failed to create pipeline', message)
    } finally {
      setIsSubmitting(false)
    }
  }

  return (
    <div className="space-y-6 max-w-2xl">
      <div>
        <h1 className="text-2xl font-bold text-white">Create Pipeline</h1>
        <p className="text-gray-400 mt-1">Configure a new CI/CD pipeline</p>
      </div>

      <form onSubmit={handleSubmit} className="glass-card p-6 space-y-6">
        <div>
          <label className="block text-sm font-medium text-gray-300 mb-2">
            Pipeline Name
          </label>
          <input
            type="text"
            value={name}
            onChange={(e) => setName(e.target.value)}
            placeholder="e.g., my-app-pipeline"
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
            placeholder="Pipeline description..."
            className="glass-input min-h-[80px]"
            rows={3}
          />
        </div>

        <div>
          <label className="block text-sm font-medium text-gray-300 mb-2">
            Git Repository URL
          </label>
          <input
            type="text"
            value={gitRepo}
            onChange={(e) => setGitRepo(e.target.value)}
            placeholder="https://github.com/org/repo"
            className="glass-input"
            required
          />
        </div>

        <div>
          <label className="block text-sm font-medium text-gray-300 mb-2">
            Branch
          </label>
          <input
            type="text"
            value={branch}
            onChange={(e) => setBranch(e.target.value)}
            placeholder="main"
            className="glass-input"
          />
        </div>

        <div className="flex gap-4 pt-4">
          <button
            type="button"
            onClick={() => navigate('/pipelines')}
            className="glass-btn flex-1"
          >
            Cancel
          </button>
          <button
            type="submit"
            disabled={isSubmitting || !name || !gitRepo}
            className="glass-btn-primary flex-1 flex items-center justify-center gap-2"
          >
            {isSubmitting ? (
              <>
                <Loader2 className="w-4 h-4 animate-spin" />
                Creating...
              </>
            ) : (
              <>
                <Plus className="w-4 h-4" />
                Create Pipeline
              </>
            )}
          </button>
        </div>
      </form>
    </div>
  )
}

export default function Pipelines() {
  return (
    <Routes>
      <Route index element={<PipelinesList />} />
      <Route path="new" element={<PipelineCreate />} />
      <Route path=":id/*" element={<PipelineDetail />} />
    </Routes>
  )
}
