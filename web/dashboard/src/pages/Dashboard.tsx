// Krustron Dashboard - Main Dashboard Page
// Author: Anubhav Gain <anubhavg@infopercept.com>

import { useEffect, useState } from 'react'
import { motion } from 'framer-motion'
import {
  Server,
  Box,
  GitBranch,
  AlertTriangle,
  TrendingUp,
  TrendingDown,
  Activity,
  Cpu,
  HardDrive,
  CheckCircle,
  XCircle,
  AlertCircle,
} from 'lucide-react'
import {
  AreaChart,
  Area,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  PieChart,
  Pie,
  Cell,
} from 'recharts'
import { clsx } from 'clsx'
import { useWebSocketContext } from '@/hooks/useWebSocket'
import { useClustersStore, useApplicationsStore, usePipelinesStore, useAlertsStore } from '@/store/useStore'
import { WebSocketMessage, ClusterMetrics, PipelineRun } from '@/types'

// Metric Card Component
function MetricCard({
  title,
  value,
  change,
  changeType,
  icon: Icon,
  color,
}: {
  title: string
  value: string | number
  change?: string
  changeType?: 'up' | 'down' | 'neutral'
  icon: React.ElementType
  color: string
}) {
  return (
    <motion.div
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      className="glass-card p-6"
    >
      <div className="flex items-start justify-between">
        <div>
          <p className="text-sm text-gray-400">{title}</p>
          <p className="text-3xl font-bold text-white mt-1">{value}</p>
          {change && (
            <div className={clsx(
              'flex items-center gap-1 mt-2 text-sm',
              changeType === 'up' && 'text-status-healthy',
              changeType === 'down' && 'text-status-error',
              changeType === 'neutral' && 'text-gray-400'
            )}>
              {changeType === 'up' && <TrendingUp className="w-4 h-4" />}
              {changeType === 'down' && <TrendingDown className="w-4 h-4" />}
              <span>{change}</span>
            </div>
          )}
        </div>
        <div className={clsx('p-3 rounded-xl', color)}>
          <Icon className="w-6 h-6 text-white" />
        </div>
      </div>
    </motion.div>
  )
}

// Status Donut Chart
function StatusDonut({
  data,
  title,
}: {
  data: { name: string; value: number; color: string }[]
  title: string
}) {
  const total = data.reduce((acc, item) => acc + item.value, 0)

  return (
    <div className="glass-card p-6">
      <h3 className="text-lg font-semibold text-white mb-4">{title}</h3>
      <div className="flex items-center gap-6">
        <div className="w-32 h-32">
          <ResponsiveContainer width="100%" height="100%">
            <PieChart>
              <Pie
                data={data}
                cx="50%"
                cy="50%"
                innerRadius={35}
                outerRadius={50}
                paddingAngle={2}
                dataKey="value"
              >
                {data.map((entry, index) => (
                  <Cell key={`cell-${index}`} fill={entry.color} />
                ))}
              </Pie>
            </PieChart>
          </ResponsiveContainer>
        </div>
        <div className="flex-1 space-y-2">
          {data.map((item, index) => (
            <div key={index} className="flex items-center justify-between">
              <div className="flex items-center gap-2">
                <div
                  className="w-3 h-3 rounded-full"
                  style={{ backgroundColor: item.color }}
                />
                <span className="text-sm text-gray-400">{item.name}</span>
              </div>
              <span className="text-sm font-medium text-white">
                {item.value} ({total > 0 ? Math.round((item.value / total) * 100) : 0}%)
              </span>
            </div>
          ))}
        </div>
      </div>
    </div>
  )
}

// Recent Activity Item
function ActivityItem({
  icon: Icon,
  iconColor,
  title,
  description,
  time,
}: {
  icon: React.ElementType
  iconColor: string
  title: string
  description: string
  time: string
}) {
  return (
    <div className="flex items-start gap-3 p-3 rounded-xl hover:bg-glass-light transition-colors">
      <div className={clsx('p-2 rounded-lg', iconColor)}>
        <Icon className="w-4 h-4 text-white" />
      </div>
      <div className="flex-1 min-w-0">
        <p className="text-sm font-medium text-white truncate">{title}</p>
        <p className="text-xs text-gray-400 truncate">{description}</p>
      </div>
      <span className="text-xs text-gray-500 whitespace-nowrap">{time}</span>
    </div>
  )
}

export default function Dashboard() {
  const { subscribe, isConnected } = useWebSocketContext()
  const { clusters } = useClustersStore()
  const { applications } = useApplicationsStore()
  const { pipelines } = usePipelinesStore()
  const { alerts, unreadCount } = useAlertsStore()

  // Real-time metrics state
  const [, setClusterMetrics] = useState<ClusterMetrics[]>([])
  const [, setRecentPipelineRuns] = useState<PipelineRun[]>([])
  const [cpuHistory, setCpuHistory] = useState<{ time: string; value: number }[]>([])

  // Subscribe to real-time updates
  useEffect(() => {
    if (!isConnected) return

    const unsubscribeCluster = subscribe('cluster.metrics', (message: WebSocketMessage) => {
      const metrics = message.payload as ClusterMetrics
      setClusterMetrics(prev => {
        const existing = prev.findIndex(m => m.clusterId === metrics.clusterId)
        if (existing >= 0) {
          const updated = [...prev]
          updated[existing] = metrics
          return updated
        }
        return [...prev, metrics]
      })

      // Update CPU history
      setCpuHistory(prev => [
        ...prev.slice(-29),
        {
          time: new Date().toLocaleTimeString('en-US', { hour12: false, hour: '2-digit', minute: '2-digit' }),
          value: metrics.cpu.percentage,
        },
      ])
    })

    const unsubscribePipeline = subscribe('pipeline.status', (message: WebSocketMessage) => {
      const run = message.payload as PipelineRun
      setRecentPipelineRuns(prev => [run, ...prev.slice(0, 9)])
    })

    return () => {
      unsubscribeCluster()
      unsubscribePipeline()
    }
  }, [subscribe, isConnected])

  // Mock data for initial render
  const mockCpuHistory = Array.from({ length: 30 }, (_, i) => ({
    time: `${String(i).padStart(2, '0')}:00`,
    value: Math.floor(Math.random() * 40) + 30,
  }))

  const displayCpuHistory = cpuHistory.length > 0 ? cpuHistory : mockCpuHistory

  // Calculate summary stats
  const healthyClusters = clusters.filter(c => c.status === 'connected').length
  const syncedApps = applications.filter(a => a.syncStatus === 'Synced').length
  const successfulPipelines = pipelines.filter(p => p.lastRun?.status === 'Succeeded').length
  const criticalAlerts = alerts.filter(a => a.severity === 'critical' && a.status === 'firing').length

  // Application status distribution
  const appStatusData = [
    { name: 'Synced', value: applications.filter(a => a.syncStatus === 'Synced').length, color: '#22c55e' },
    { name: 'OutOfSync', value: applications.filter(a => a.syncStatus === 'OutOfSync').length, color: '#f97316' },
    { name: 'Unknown', value: applications.filter(a => a.syncStatus === 'Unknown').length, color: '#6b7280' },
  ]

  // Pipeline status distribution
  const pipelineStatusData = [
    { name: 'Succeeded', value: pipelines.filter(p => p.lastRun?.status === 'Succeeded').length, color: '#22c55e' },
    { name: 'Failed', value: pipelines.filter(p => p.lastRun?.status === 'Failed').length, color: '#ef4444' },
    { name: 'Running', value: pipelines.filter(p => p.lastRun?.status === 'Running').length, color: '#3b82f6' },
    { name: 'Pending', value: pipelines.filter(p => !p.lastRun || p.lastRun.status === 'Pending').length, color: '#6b7280' },
  ]

  return (
    <div className="space-y-6">
      {/* Header */}
      <div>
        <h1 className="text-2xl font-bold text-white">Dashboard</h1>
        <p className="text-gray-400 mt-1">Real-time overview of your Kubernetes infrastructure</p>
      </div>

      {/* Metric Cards */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        <MetricCard
          title="Clusters"
          value={clusters.length}
          change={`${healthyClusters} healthy`}
          changeType="up"
          icon={Server}
          color="bg-primary-500"
        />
        <MetricCard
          title="Applications"
          value={applications.length}
          change={`${syncedApps} synced`}
          changeType="up"
          icon={Box}
          color="bg-accent-500"
        />
        <MetricCard
          title="Pipelines"
          value={pipelines.length}
          change={`${successfulPipelines} succeeded`}
          changeType="up"
          icon={GitBranch}
          color="bg-status-info"
        />
        <MetricCard
          title="Active Alerts"
          value={unreadCount}
          change={criticalAlerts > 0 ? `${criticalAlerts} critical` : 'All clear'}
          changeType={criticalAlerts > 0 ? 'down' : 'up'}
          icon={AlertTriangle}
          color={criticalAlerts > 0 ? 'bg-status-error' : 'bg-status-healthy'}
        />
      </div>

      {/* Charts Row */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* CPU Usage Chart */}
        <div className="lg:col-span-2 glass-card p-6">
          <div className="flex items-center justify-between mb-4">
            <h3 className="text-lg font-semibold text-white">Cluster CPU Usage</h3>
            <div className="flex items-center gap-2 text-sm text-gray-400">
              <Activity className="w-4 h-4" />
              <span>Real-time</span>
            </div>
          </div>
          <div className="h-64">
            <ResponsiveContainer width="100%" height="100%">
              <AreaChart data={displayCpuHistory}>
                <defs>
                  <linearGradient id="cpuGradient" x1="0" y1="0" x2="0" y2="1">
                    <stop offset="5%" stopColor="#2563eb" stopOpacity={0.3} />
                    <stop offset="95%" stopColor="#2563eb" stopOpacity={0} />
                  </linearGradient>
                </defs>
                <CartesianGrid strokeDasharray="3 3" stroke="rgba(255,255,255,0.1)" />
                <XAxis
                  dataKey="time"
                  stroke="rgba(255,255,255,0.3)"
                  tick={{ fill: 'rgba(255,255,255,0.5)', fontSize: 10 }}
                />
                <YAxis
                  stroke="rgba(255,255,255,0.3)"
                  tick={{ fill: 'rgba(255,255,255,0.5)', fontSize: 10 }}
                  domain={[0, 100]}
                  tickFormatter={(value) => `${value}%`}
                />
                <Tooltip
                  contentStyle={{
                    backgroundColor: 'rgba(26, 31, 61, 0.95)',
                    border: '1px solid rgba(255,255,255,0.1)',
                    borderRadius: '12px',
                  }}
                  labelStyle={{ color: '#fff' }}
                  itemStyle={{ color: '#2563eb' }}
                  formatter={(value: number) => [`${value.toFixed(1)}%`, 'CPU']}
                />
                <Area
                  type="monotone"
                  dataKey="value"
                  stroke="#2563eb"
                  strokeWidth={2}
                  fillOpacity={1}
                  fill="url(#cpuGradient)"
                />
              </AreaChart>
            </ResponsiveContainer>
          </div>
        </div>

        {/* Resource Usage */}
        <div className="glass-card p-6">
          <h3 className="text-lg font-semibold text-white mb-4">Resource Usage</h3>
          <div className="space-y-6">
            <div>
              <div className="flex items-center justify-between mb-2">
                <div className="flex items-center gap-2">
                  <Cpu className="w-4 h-4 text-primary-400" />
                  <span className="text-sm text-gray-400">CPU</span>
                </div>
                <span className="text-sm font-medium text-white">45%</span>
              </div>
              <div className="progress-bar">
                <div className="progress-fill" style={{ width: '45%' }} />
              </div>
            </div>
            <div>
              <div className="flex items-center justify-between mb-2">
                <div className="flex items-center gap-2">
                  <HardDrive className="w-4 h-4 text-accent-400" />
                  <span className="text-sm text-gray-400">Memory</span>
                </div>
                <span className="text-sm font-medium text-white">62%</span>
              </div>
              <div className="progress-bar">
                <div className="progress-fill" style={{ width: '62%' }} />
              </div>
            </div>
            <div>
              <div className="flex items-center justify-between mb-2">
                <div className="flex items-center gap-2">
                  <Box className="w-4 h-4 text-status-info" />
                  <span className="text-sm text-gray-400">Pods</span>
                </div>
                <span className="text-sm font-medium text-white">78%</span>
              </div>
              <div className="progress-bar">
                <div className="progress-fill" style={{ width: '78%' }} />
              </div>
            </div>
          </div>
        </div>
      </div>

      {/* Status Distributions */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <StatusDonut data={appStatusData} title="Application Status" />
        <StatusDonut data={pipelineStatusData} title="Pipeline Status" />
      </div>

      {/* Recent Activity */}
      <div className="glass-card p-6">
        <h3 className="text-lg font-semibold text-white mb-4">Recent Activity</h3>
        <div className="space-y-1">
          <ActivityItem
            icon={CheckCircle}
            iconColor="bg-status-healthy"
            title="Pipeline 'build-deploy' succeeded"
            description="Application 'my-app' deployed to production"
            time="2 min ago"
          />
          <ActivityItem
            icon={Box}
            iconColor="bg-primary-500"
            title="Application 'api-gateway' synced"
            description="3 resources updated in cluster 'production'"
            time="5 min ago"
          />
          <ActivityItem
            icon={AlertCircle}
            iconColor="bg-status-warning"
            title="High memory usage detected"
            description="Cluster 'staging' memory at 85%"
            time="10 min ago"
          />
          <ActivityItem
            icon={Server}
            iconColor="bg-accent-500"
            title="New cluster connected"
            description="Cluster 'development' added successfully"
            time="15 min ago"
          />
          <ActivityItem
            icon={XCircle}
            iconColor="bg-status-error"
            title="Pipeline 'test-suite' failed"
            description="Test stage failed with exit code 1"
            time="20 min ago"
          />
        </div>
      </div>
    </div>
  )
}
