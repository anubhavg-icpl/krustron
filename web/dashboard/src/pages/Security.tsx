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
} from 'lucide-react'
import { clsx } from 'clsx'
import {
  Tooltip,
  ResponsiveContainer,
  PieChart,
  Pie,
  Cell,
} from 'recharts'
import { BarChart2 } from 'lucide-react'

// Types
interface Vulnerability {
  cve: string
  severity: 'critical' | 'high' | 'medium' | 'low'
  title: string
  packageName: string
  installedVersion: string
  fixedVersion?: string
  affectedImages: number
}

// Vulnerability Card
function VulnerabilityCard({
  cve,
  severity,
  title,
  packageName,
  installedVersion,
  fixedVersion,
  affectedImages,
}: Vulnerability) {
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
  const [vulnerabilities] = useState<Vulnerability[]>([])

  // Calculate severity counts from real data
  const severityData = [
    { name: 'Critical', value: vulnerabilities.filter(v => v.severity === 'critical').length, color: '#ef4444' },
    { name: 'High', value: vulnerabilities.filter(v => v.severity === 'high').length, color: '#f97316' },
    { name: 'Medium', value: vulnerabilities.filter(v => v.severity === 'medium').length, color: '#eab308' },
    { name: 'Low', value: vulnerabilities.filter(v => v.severity === 'low').length, color: '#3b82f6' },
  ]

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
          {totalVulnerabilities === 0 ? (
            <div className="flex flex-col items-center justify-center h-48 text-gray-400">
              <Shield className="w-12 h-12 mb-2 opacity-50" />
              <p className="text-sm">No vulnerabilities found</p>
            </div>
          ) : (
            <>
              <div className="h-48">
                <ResponsiveContainer width="100%" height="100%">
                  <PieChart>
                    <Pie
                      data={severityData.filter(d => d.value > 0)}
                      cx="50%"
                      cy="50%"
                      innerRadius={40}
                      outerRadius={70}
                      paddingAngle={2}
                      dataKey="value"
                    >
                      {severityData.filter(d => d.value > 0).map((entry, index) => (
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
            </>
          )}
        </div>

        {/* Trend Chart */}
        <div className="lg:col-span-2 glass-card p-6">
          <h3 className="text-lg font-semibold text-white mb-4">Vulnerability Trend</h3>
          <div className="flex flex-col items-center justify-center h-48 text-gray-400">
            <BarChart2 className="w-12 h-12 mb-2 opacity-50" />
            <p className="text-sm">Run a scan to see vulnerability trends</p>
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
      {vulnerabilities.length === 0 ? (
        <div className="glass-card p-12 text-center">
          <Shield className="w-16 h-16 mx-auto mb-4 text-gray-600" />
          <h3 className="text-xl font-semibold text-white mb-2">No Vulnerabilities</h3>
          <p className="text-gray-400 mb-4">Run a security scan to detect vulnerabilities in your images</p>
          <button className="glass-btn-primary">
            <RefreshCw className="w-4 h-4 mr-2 inline" />
            Run Scan
          </button>
        </div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-4">
          {vulnerabilities
            .filter(v => v.title.toLowerCase().includes(searchQuery.toLowerCase()) ||
                        v.cve.toLowerCase().includes(searchQuery.toLowerCase()))
            .map((vuln) => (
              <VulnerabilityCard key={vuln.cve} {...vuln} />
            ))}
        </div>
      )}
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
