// Krustron Dashboard - Cost Management Page
// Author: Anubhav Gain <anubhavg@infopercept.com>

import { useState } from 'react'
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
  Box,
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

// Mock data
const costTrend = [
  { date: 'Jan', cost: 4500, budget: 5000 },
  { date: 'Feb', cost: 4800, budget: 5000 },
  { date: 'Mar', cost: 5200, budget: 5000 },
  { date: 'Apr', cost: 4900, budget: 5500 },
  { date: 'May', cost: 5100, budget: 5500 },
  { date: 'Jun', cost: 5400, budget: 5500 },
]

const costByCluster = [
  { name: 'production', value: 3200, color: '#2563eb' },
  { name: 'staging', value: 1500, color: '#f97316' },
  { name: 'development', value: 800, color: '#22c55e' },
  { name: 'testing', value: 400, color: '#eab308' },
]

const costByResource = [
  { name: 'Compute', cost: 3500 },
  { name: 'Storage', cost: 1200 },
  { name: 'Network', cost: 800 },
  { name: 'Load Balancer', cost: 400 },
]

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

  const totalCost = costByCluster.reduce((acc, item) => acc + item.value, 0)

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
          change="+12% vs last month"
          changeType="up"
          icon={DollarSign}
        />
        <CostMetricCard
          title="Forecasted"
          value="$6,200"
          change="Based on current usage"
          changeType="neutral"
          icon={TrendingUp}
        />
        <CostMetricCard
          title="Budget Remaining"
          value="$1,600"
          change="29% of budget"
          changeType="neutral"
          icon={AlertTriangle}
        />
        <CostMetricCard
          title="Potential Savings"
          value="$450"
          change="With recommendations"
          changeType="down"
          icon={ArrowUpRight}
        />
      </div>

      {/* Charts Row */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Cost Trend */}
        <div className="lg:col-span-2 glass-card p-6">
          <h3 className="text-lg font-semibold text-white mb-4">Cost vs Budget</h3>
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
        </div>

        {/* Cost by Cluster */}
        <div className="glass-card p-6">
          <h3 className="text-lg font-semibold text-white mb-4">Cost by Cluster</h3>
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
        </div>
      </div>

      {/* Cost by Resource */}
      <div className="glass-card p-6">
        <h3 className="text-lg font-semibold text-white mb-4">Cost by Resource Type</h3>
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
      </div>

      {/* Recommendations */}
      <div className="glass-card p-6">
        <h3 className="text-lg font-semibold text-white mb-4">Cost Optimization Recommendations</h3>
        <div className="space-y-4">
          <div className="flex items-start gap-4 p-4 bg-glass-light rounded-xl">
            <div className="p-2 rounded-lg bg-status-healthy/20">
              <Cpu className="w-5 h-5 text-status-healthy" />
            </div>
            <div className="flex-1">
              <h4 className="font-medium text-white">Right-size idle pods</h4>
              <p className="text-sm text-gray-400 mt-1">
                5 pods in production cluster have less than 10% CPU utilization.
                Consider reducing resource requests.
              </p>
              <p className="text-sm text-status-healthy mt-2">Potential savings: $120/month</p>
            </div>
          </div>

          <div className="flex items-start gap-4 p-4 bg-glass-light rounded-xl">
            <div className="p-2 rounded-lg bg-status-warning/20">
              <Server className="w-5 h-5 text-status-warning" />
            </div>
            <div className="flex-1">
              <h4 className="font-medium text-white">Use spot instances for non-production</h4>
              <p className="text-sm text-gray-400 mt-1">
                Development and testing clusters can use spot instances for up to 70% cost reduction.
              </p>
              <p className="text-sm text-status-healthy mt-2">Potential savings: $280/month</p>
            </div>
          </div>

          <div className="flex items-start gap-4 p-4 bg-glass-light rounded-xl">
            <div className="p-2 rounded-lg bg-status-info/20">
              <Box className="w-5 h-5 text-status-info" />
            </div>
            <div className="flex-1">
              <h4 className="font-medium text-white">Clean up unused PVCs</h4>
              <p className="text-sm text-gray-400 mt-1">
                3 persistent volume claims are not attached to any pods. Consider deleting unused storage.
              </p>
              <p className="text-sm text-status-healthy mt-2">Potential savings: $50/month</p>
            </div>
          </div>
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
