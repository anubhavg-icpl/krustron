// Krustron Dashboard - Security Page
// Author: Anubhav Gain <anubhavg@infopercept.com>

import { useState } from 'react'
import { Routes, Route } from 'react-router-dom'
import { motion } from 'framer-motion'
import {
  Shield,
  Search,
  Eye,
  Download,
  RefreshCw,
  Filter,
  ArrowDownRight,
} from 'lucide-react'
import { clsx } from 'clsx'
import {
  BarChart,
  Bar,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  PieChart,
  Pie,
  Cell,
} from 'recharts'

// Mock data for visualization
const severityData = [
  { name: 'Critical', value: 3, color: '#ef4444' },
  { name: 'High', value: 12, color: '#f97316' },
  { name: 'Medium', value: 28, color: '#eab308' },
  { name: 'Low', value: 45, color: '#3b82f6' },
]

const vulnerabilityTrend = [
  { date: 'Week 1', critical: 5, high: 15, medium: 30, low: 50 },
  { date: 'Week 2', critical: 4, high: 12, medium: 28, low: 48 },
  { date: 'Week 3', critical: 3, high: 14, medium: 25, low: 45 },
  { date: 'Week 4', critical: 3, high: 12, medium: 28, low: 45 },
]

// Vulnerability Card
function VulnerabilityCard({
  cve,
  severity,
  title,
  packageName,
  installedVersion,
  fixedVersion,
  affectedImages,
}: {
  cve: string
  severity: 'critical' | 'high' | 'medium' | 'low'
  title: string
  packageName: string
  installedVersion: string
  fixedVersion?: string
  affectedImages: number
}) {
  const severityStyles = {
    critical: 'bg-status-error/20 text-status-error border-status-error/30',
    high: 'bg-accent-500/20 text-accent-400 border-accent-500/30',
    medium: 'bg-status-warning/20 text-status-warning border-status-warning/30',
    low: 'bg-status-info/20 text-status-info border-status-info/30',
  }

  return (
    <motion.div
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      className="glass-card p-4"
    >
      <div className="flex items-start justify-between">
        <div className="flex items-center gap-3">
          <span className={clsx('status-badge border', severityStyles[severity])}>
            {severity.toUpperCase()}
          </span>
          <span className="text-sm font-mono text-gray-400">{cve}</span>
        </div>
        <button className="p-2 rounded-lg hover:bg-glass-light transition-colors">
          <Eye className="w-4 h-4 text-gray-400" />
        </button>
      </div>

      <h4 className="text-white font-medium mt-3 line-clamp-2">{title}</h4>

      <div className="mt-3 space-y-2 text-sm">
        <div className="flex items-center justify-between">
          <span className="text-gray-400">Package</span>
          <span className="text-gray-300 font-mono">{packageName}</span>
        </div>
        <div className="flex items-center justify-between">
          <span className="text-gray-400">Installed</span>
          <span className="text-gray-300 font-mono">{installedVersion}</span>
        </div>
        {fixedVersion && (
          <div className="flex items-center justify-between">
            <span className="text-gray-400">Fixed in</span>
            <span className="text-status-healthy font-mono">{fixedVersion}</span>
          </div>
        )}
        <div className="flex items-center justify-between">
          <span className="text-gray-400">Affected Images</span>
          <span className="text-gray-300">{affectedImages}</span>
        </div>
      </div>
    </motion.div>
  )
}

// Security Overview
function SecurityOverview() {
  const [searchQuery, setSearchQuery] = useState('')

  const totalVulnerabilities = severityData.reduce((acc, item) => acc + item.value, 0)

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold text-white">Security</h1>
          <p className="text-gray-400 mt-1">Vulnerability scanning and policy compliance</p>
        </div>
        <div className="flex gap-2">
          <button className="glass-btn flex items-center gap-2">
            <Download className="w-4 h-4" />
            Export Report
          </button>
          <button className="glass-btn-primary flex items-center gap-2">
            <RefreshCw className="w-4 h-4" />
            Run Scan
          </button>
        </div>
      </div>

      {/* Summary Cards */}
      <div className="grid grid-cols-2 lg:grid-cols-5 gap-4">
        <div className="glass-card p-4">
          <div className="flex items-center justify-between">
            <Shield className="w-5 h-5 text-gray-400" />
            <span className="text-xs text-status-healthy flex items-center gap-1">
              <ArrowDownRight className="w-3 h-3" />
              12%
            </span>
          </div>
          <div className="text-2xl font-bold text-white mt-2">{totalVulnerabilities}</div>
          <div className="text-sm text-gray-400">Total Issues</div>
        </div>

        {severityData.map((item) => (
          <div key={item.name} className="glass-card p-4">
            <div className="flex items-center justify-between">
              <div
                className="w-3 h-3 rounded-full"
                style={{ backgroundColor: item.color }}
              />
            </div>
            <div className="text-2xl font-bold text-white mt-2">{item.value}</div>
            <div className="text-sm text-gray-400">{item.name}</div>
          </div>
        ))}
      </div>

      {/* Charts Row */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Severity Distribution */}
        <div className="glass-card p-6">
          <h3 className="text-lg font-semibold text-white mb-4">Severity Distribution</h3>
          <div className="h-48">
            <ResponsiveContainer width="100%" height="100%">
              <PieChart>
                <Pie
                  data={severityData}
                  cx="50%"
                  cy="50%"
                  innerRadius={40}
                  outerRadius={70}
                  paddingAngle={2}
                  dataKey="value"
                >
                  {severityData.map((entry, index) => (
                    <Cell key={`cell-${index}`} fill={entry.color} />
                  ))}
                </Pie>
                <Tooltip
                  contentStyle={{
                    backgroundColor: 'rgba(26, 31, 61, 0.95)',
                    border: '1px solid rgba(255,255,255,0.1)',
                    borderRadius: '12px',
                  }}
                />
              </PieChart>
            </ResponsiveContainer>
          </div>
          <div className="flex flex-wrap justify-center gap-4 mt-4">
            {severityData.map((item) => (
              <div key={item.name} className="flex items-center gap-2">
                <div
                  className="w-3 h-3 rounded-full"
                  style={{ backgroundColor: item.color }}
                />
                <span className="text-xs text-gray-400">{item.name}</span>
              </div>
            ))}
          </div>
        </div>

        {/* Trend Chart */}
        <div className="lg:col-span-2 glass-card p-6">
          <h3 className="text-lg font-semibold text-white mb-4">Vulnerability Trend</h3>
          <div className="h-48">
            <ResponsiveContainer width="100%" height="100%">
              <BarChart data={vulnerabilityTrend}>
                <CartesianGrid strokeDasharray="3 3" stroke="rgba(255,255,255,0.1)" />
                <XAxis
                  dataKey="date"
                  stroke="rgba(255,255,255,0.3)"
                  tick={{ fill: 'rgba(255,255,255,0.5)', fontSize: 10 }}
                />
                <YAxis
                  stroke="rgba(255,255,255,0.3)"
                  tick={{ fill: 'rgba(255,255,255,0.5)', fontSize: 10 }}
                />
                <Tooltip
                  contentStyle={{
                    backgroundColor: 'rgba(26, 31, 61, 0.95)',
                    border: '1px solid rgba(255,255,255,0.1)',
                    borderRadius: '12px',
                  }}
                />
                <Bar dataKey="critical" stackId="a" fill="#ef4444" />
                <Bar dataKey="high" stackId="a" fill="#f97316" />
                <Bar dataKey="medium" stackId="a" fill="#eab308" />
                <Bar dataKey="low" stackId="a" fill="#3b82f6" />
              </BarChart>
            </ResponsiveContainer>
          </div>
        </div>
      </div>

      {/* Filters */}
      <div className="flex flex-col sm:flex-row gap-4">
        <div className="relative flex-1">
          <Search className="absolute left-4 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400" />
          <input
            type="text"
            placeholder="Search vulnerabilities..."
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

      {/* Vulnerabilities List */}
      <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-4">
        <VulnerabilityCard
          cve="CVE-2024-1234"
          severity="critical"
          title="Remote Code Execution in libcurl"
          packageName="libcurl"
          installedVersion="7.68.0"
          fixedVersion="7.88.0"
          affectedImages={5}
        />
        <VulnerabilityCard
          cve="CVE-2024-5678"
          severity="high"
          title="Authentication Bypass in OpenSSL"
          packageName="openssl"
          installedVersion="1.1.1f"
          fixedVersion="1.1.1t"
          affectedImages={12}
        />
        <VulnerabilityCard
          cve="CVE-2024-9012"
          severity="medium"
          title="Cross-Site Scripting in nginx"
          packageName="nginx"
          installedVersion="1.18.0"
          fixedVersion="1.24.0"
          affectedImages={3}
        />
        <VulnerabilityCard
          cve="CVE-2024-3456"
          severity="low"
          title="Information Disclosure in glibc"
          packageName="glibc"
          installedVersion="2.31"
          affectedImages={8}
        />
      </div>
    </div>
  )
}

export default function Security() {
  return (
    <Routes>
      <Route index element={<SecurityOverview />} />
    </Routes>
  )
}
