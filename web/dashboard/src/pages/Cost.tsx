// Krustron Dashboard - Cost Management Page
// Author: Anubhav Gain <anubhavg@infopercept.com>

import { useState, useEffect } from 'react'
import { Routes, Route } from 'react-router-dom'
import { motion } from 'framer-motion'
import {
  DollarSign,
  TrendingUp,
  TrendingDown,
  ArrowUpRight,
  Download,
  AlertTriangle,
  Server,
  Cpu,
} from 'lucide-react'
import { clsx } from 'clsx'
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
  BarChart,
  Bar,
} from 'recharts'
import { costApi } from '@/api'

// Types
interface CostData {
  date: string
  cost: number
  budget: number
}

interface ClusterCost {
  name: string
  value: number
  color: string
}

interface ResourceCost {
  name: string
  cost: number
}

// Metric Card
function CostMetricCard({
  title,
  value,
  change,
  changeType,
  icon: Icon,
}: {
  title: string
  value: string
  change: string
  changeType: 'up' | 'down' | 'neutral'
  icon: React.ElementType
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
          <div className={clsx(
            'flex items-center gap-1 mt-2 text-sm',
            changeType === 'up' && 'text-status-error',
            changeType === 'down' && 'text-status-healthy',
            changeType === 'neutral' && 'text-gray-400'
          )}>
            {changeType === 'up' && <TrendingUp className="w-4 h-4" />}
            {changeType === 'down' && <TrendingDown className="w-4 h-4" />}
            <span>{change}</span>
          </div>
        </div>
        <div className="p-3 rounded-xl bg-primary-500/20">
          <Icon className="w-6 h-6 text-primary-400" />
        </div>
      </div>
    </motion.div>
  )
}

// Cost Overview
function CostOverview() {
  const [dateRange, setDateRange] = useState('30d')
  const [costTrend] = useState<CostData[]>([])
  const [costByCluster, setCostByCluster] = useState<ClusterCost[]>([])
  const [costByResource, setCostByResource] = useState<ResourceCost[]>([])
  const [totalCost, setTotalCost] = useState(0)
  const [changePercent, setChangePercent] = useState(0)

  // Pull the real cost summary + allocations. The page was previously an empty
  // shell (state initialized to [] and never fetched).
  useEffect(() => {
    let cancelled = false
    ;(async () => {
      try {
        const [summary, allocations] = await Promise.all([
          costApi.getSummary(),
          costApi.listAllocations({ limit: 200 }),
        ])
        if (cancelled) return
        setTotalCost(summary.current_month_cost || 0)
        setChangePercent(summary.change_percent || 0)

        // Group allocations by cluster for the breakdown chart.
        const byCluster = new Map<string, number>()
        const byResource: Record<string, number> = { CPU: 0, Memory: 0, Storage: 0, Network: 0 }
        const palette = ['#6366f1', '#ec4899', '#f59e0b', '#10b981', '#06b6d4', '#8b5cf6']
        for (const a of allocations) {
          const key = a.cluster_name || a.cluster_id || 'unknown'
          byCluster.set(key, (byCluster.get(key) ?? 0) + (a.total_cost || 0))
          byResource.CPU += a.cpu_cost || 0
          byResource.Memory += a.memory_cost || 0
          byResource.Storage += a.storage_cost || 0
          byResource.Network += a.network_cost || 0
        }
        setCostByCluster(
          Array.from(byCluster.entries()).map(([name, value], i) => ({
            name,
            value: Math.round(value * 100) / 100,
            color: palette[i % palette.length],
          }))
        )
        // Only keep resource buckets with non-zero cost for the chart.
        setCostByResource(
          Object.entries(byResource)
            .filter(([, c]) => c > 0)
            .map(([name, cost]) => ({ name, cost: Math.round(cost * 100) / 100 }))
        )
      } catch {
        // Leave zeros — the empty-state UI already renders gracefully.
      }
    })()
    return () => {
      cancelled = true
    }
  }, [dateRange])

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold text-white">Cost Management</h1>
          <p className="text-gray-400 mt-1">Cloud cost allocation and optimization</p>
        </div>
        <div className="flex gap-2">
          <select
            value={dateRange}
            onChange={(e) => setDateRange(e.target.value)}
            className="glass-select"
          >
            <option value="7d">Last 7 days</option>
            <option value="30d">Last 30 days</option>
            <option value="90d">Last 90 days</option>
            <option value="1y">Last year</option>
          </select>
          <button className="glass-btn flex items-center gap-2">
            <Download className="w-4 h-4" />
            Export
          </button>
        </div>
      </div>

      {/* Metric Cards */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        <CostMetricCard
          title="Current Month"
          value={`$${totalCost.toLocaleString()}`}
          change={
            totalCost > 0
              ? `${changePercent >= 0 ? '+' : ''}${changePercent.toFixed(1)}% vs last month`
              : 'No data yet'
          }
          changeType={changePercent > 0 ? 'up' : changePercent < 0 ? 'down' : 'neutral'}
          icon={DollarSign}
        />
        <CostMetricCard
          title="Forecasted"
          value={totalCost > 0 ? `$${Math.round(totalCost * 1.1).toLocaleString()}` : "$0"}
          change="Based on current usage"
          changeType="neutral"
          icon={TrendingUp}
        />
        <CostMetricCard
          title="Budget Remaining"
          value="$0"
          change="Set a budget to track"
          changeType="neutral"
          icon={AlertTriangle}
        />
        <CostMetricCard
          title="Potential Savings"
          value="$0"
          change="Run analysis to find savings"
          changeType="neutral"
          icon={ArrowUpRight}
        />
      </div>

      {/* Charts Row */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Cost Trend */}
        <div className="lg:col-span-2 glass-card p-6">
          <h3 className="text-lg font-semibold text-white mb-4">Cost vs Budget</h3>
          {costTrend.length === 0 ? (
            <div className="flex flex-col items-center justify-center h-64 text-gray-400">
              <TrendingUp className="w-12 h-12 mb-2 opacity-50" />
              <p className="text-sm">No cost data available yet</p>
              <p className="text-xs mt-1">Connect cloud providers to track costs</p>
            </div>
          ) : (
            <div className="h-64">
              <ResponsiveContainer width="100%" height="100%">
                <AreaChart data={costTrend}>
                  <defs>
                    <linearGradient id="costGradient" x1="0" y1="0" x2="0" y2="1">
                      <stop offset="5%" stopColor="#2563eb" stopOpacity={0.3} />
                      <stop offset="95%" stopColor="#2563eb" stopOpacity={0} />
                    </linearGradient>
                  </defs>
                  <CartesianGrid strokeDasharray="3 3" stroke="rgba(255,255,255,0.1)" />
                  <XAxis
                    dataKey="date"
                    stroke="rgba(255,255,255,0.3)"
                    tick={{ fill: 'rgba(255,255,255,0.5)', fontSize: 10 }}
                  />
                  <YAxis
                    stroke="rgba(255,255,255,0.3)"
                    tick={{ fill: 'rgba(255,255,255,0.5)', fontSize: 10 }}
                    tickFormatter={(value) => `$${value / 1000}k`}
                  />
                  <Tooltip
                    contentStyle={{
                      backgroundColor: 'rgba(26, 31, 61, 0.95)',
                      border: '1px solid rgba(255,255,255,0.1)',
                      borderRadius: '12px',
                    }}
                    formatter={(value: number) => [`$${value.toLocaleString()}`, '']}
                  />
                  <Area
                    type="monotone"
                    dataKey="budget"
                    stroke="#6b7280"
                    strokeDasharray="5 5"
                    fill="none"
                    name="Budget"
                  />
                  <Area
                    type="monotone"
                    dataKey="cost"
                    stroke="#2563eb"
                    strokeWidth={2}
                    fillOpacity={1}
                    fill="url(#costGradient)"
                    name="Cost"
                  />
                </AreaChart>
              </ResponsiveContainer>
            </div>
          )}
        </div>

        {/* Cost by Cluster */}
        <div className="glass-card p-6">
          <h3 className="text-lg font-semibold text-white mb-4">Cost by Cluster</h3>
          {costByCluster.length === 0 ? (
            <div className="flex flex-col items-center justify-center h-48 text-gray-400">
              <Server className="w-12 h-12 mb-2 opacity-50" />
              <p className="text-sm">No cluster costs</p>
            </div>
          ) : (
            <>
              <div className="h-48">
                <ResponsiveContainer width="100%" height="100%">
                  <PieChart>
                    <Pie
                      data={costByCluster}
                      cx="50%"
                      cy="50%"
                      innerRadius={40}
                      outerRadius={70}
                      paddingAngle={2}
                      dataKey="value"
                    >
                      {costByCluster.map((entry, index) => (
                        <Cell key={`cell-${index}`} fill={entry.color} />
                      ))}
                    </Pie>
                    <Tooltip
                      contentStyle={{
                        backgroundColor: 'rgba(26, 31, 61, 0.95)',
                        border: '1px solid rgba(255,255,255,0.1)',
                        borderRadius: '12px',
                      }}
                      formatter={(value: number) => [`$${value.toLocaleString()}`, '']}
                    />
                  </PieChart>
                </ResponsiveContainer>
              </div>
              <div className="space-y-2 mt-4">
                {costByCluster.map((item) => (
                  <div key={item.name} className="flex items-center justify-between">
                    <div className="flex items-center gap-2">
                      <div
                        className="w-3 h-3 rounded-full"
                        style={{ backgroundColor: item.color }}
                      />
                      <span className="text-sm text-gray-400">{item.name}</span>
                    </div>
                    <span className="text-sm font-medium text-white">
                      ${item.value.toLocaleString()}
                    </span>
                  </div>
                ))}
              </div>
            </>
          )}
        </div>
      </div>

      {/* Cost by Resource */}
      <div className="glass-card p-6">
        <h3 className="text-lg font-semibold text-white mb-4">Cost by Resource Type</h3>
        {costByResource.length === 0 ? (
          <div className="flex flex-col items-center justify-center h-48 text-gray-400">
            <Cpu className="w-12 h-12 mb-2 opacity-50" />
            <p className="text-sm">No resource cost data</p>
            <p className="text-xs mt-1">Enable cost tracking to see breakdown</p>
          </div>
        ) : (
          <div className="h-48">
            <ResponsiveContainer width="100%" height="100%">
              <BarChart data={costByResource} layout="vertical">
                <CartesianGrid strokeDasharray="3 3" stroke="rgba(255,255,255,0.1)" />
                <XAxis
                  type="number"
                  stroke="rgba(255,255,255,0.3)"
                  tick={{ fill: 'rgba(255,255,255,0.5)', fontSize: 10 }}
                  tickFormatter={(value) => `$${value}`}
                />
                <YAxis
                  type="category"
                  dataKey="name"
                  stroke="rgba(255,255,255,0.3)"
                  tick={{ fill: 'rgba(255,255,255,0.5)', fontSize: 10 }}
                  width={100}
                />
                <Tooltip
                  contentStyle={{
                    backgroundColor: 'rgba(26, 31, 61, 0.95)',
                    border: '1px solid rgba(255,255,255,0.1)',
                    borderRadius: '12px',
                  }}
                  formatter={(value: number) => [`$${value.toLocaleString()}`, 'Cost']}
                />
                <Bar dataKey="cost" fill="#2563eb" radius={[0, 4, 4, 0]} />
              </BarChart>
            </ResponsiveContainer>
          </div>
        )}
      </div>

      {/* Recommendations */}
      <div className="glass-card p-6">
        <h3 className="text-lg font-semibold text-white mb-4">Cost Optimization Recommendations</h3>
        <div className="flex flex-col items-center justify-center h-32 text-gray-400">
          <DollarSign className="w-12 h-12 mb-2 opacity-50" />
          <p className="text-sm">No recommendations yet</p>
          <p className="text-xs mt-1">Connect clusters to analyze cost optimization opportunities</p>
        </div>
      </div>
    </div>
  )
}

export default function Cost() {
  return (
    <Routes>
      <Route index element={<CostOverview />} />
    </Routes>
  )
}
