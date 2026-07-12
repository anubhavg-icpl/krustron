// Krustron Dashboard - Helm Catalog & Releases
// Author: Anubhav Gain <anubhavg@infopercept.com>

import { useEffect, useState } from 'react'
import { motion } from 'framer-motion'
import { Package, RefreshCw, Trash2, ExternalLink, Search } from 'lucide-react'
import { helmApi, ApiClientError } from '@/api'
import type { HelmRepository, HelmRelease, HelmChart } from '@/api'
import { showErrorToast, showSuccessToast } from '@/hooks/useNotificationToasts'

export default function Helm() {
  const [repos, setRepos] = useState<HelmRepository[]>([])
  const [releases, setReleases] = useState<HelmRelease[]>([])
  const [charts, setCharts] = useState<HelmChart[]>([])
  const [query, setQuery] = useState('')
  const [loading, setLoading] = useState(true)

  const refresh = async () => {
    setLoading(true)
    try {
      const [r, rel] = await Promise.all([
        helmApi.listRepositories().catch(() => [] as HelmRepository[]),
        helmApi.listReleases().catch(() => [] as HelmRelease[]),
      ])
      setRepos(r)
      setReleases(rel)
    } catch (e) {
      showErrorToast('Failed to load Helm data', e instanceof ApiClientError ? e.message : 'Network error')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    refresh()
  }, [])

  const searchCharts = async () => {
    try {
      const results = await helmApi.searchCharts(query || undefined)
      setCharts(results)
    } catch (e) {
      showErrorToast('Chart search failed', e instanceof ApiClientError ? e.message : 'Network error')
    }
  }

  const removeRepo = async (name: string) => {
    if (!window.confirm(`Remove repository "${name}"?`)) return
    try {
      await helmApi.removeRepository(name)
      setRepos((prev) => prev.filter((r) => r.name !== name))
      showSuccessToast('Repository removed', name)
    } catch (e) {
      showErrorToast('Failed to remove repository', e instanceof ApiClientError ? e.message : 'Network error')
    }
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-white">Helm</h1>
          <p className="text-gray-400 mt-1">Repositories, charts, and releases</p>
        </div>
        <button onClick={refresh} className="glass-btn flex items-center gap-2">
          <RefreshCw className={loading ? 'w-4 h-4 animate-spin' : 'w-4 h-4'} />
          Refresh
        </button>
      </div>

      {/* Repositories */}
      <div className="glass-card p-6">
        <h3 className="text-lg font-semibold text-white mb-4">Repositories</h3>
        {repos.length === 0 ? (
          <p className="text-sm text-gray-400">No repositories configured.</p>
        ) : (
          <div className="space-y-2">
            {repos.map((r) => (
              <div key={r.name} className="flex items-center justify-between p-3 bg-glass-light rounded-xl">
                <div className="flex items-center gap-3">
                  <Package className="w-5 h-5 text-primary-400" />
                  <div>
                    <p className="text-sm font-medium text-white">{r.name}</p>
                    <p className="text-xs text-gray-400 font-mono">{r.url}</p>
                  </div>
                </div>
                <div className="flex items-center gap-3">
                  <span className="text-xs text-gray-400">{r.chart_count} charts</span>
                  <button onClick={() => removeRepo(r.name)} className="p-2 rounded-lg hover:bg-glass-light text-status-error">
                    <Trash2 className="w-4 h-4" />
                  </button>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>

      {/* Chart search */}
      <div className="glass-card p-6">
        <h3 className="text-lg font-semibold text-white mb-4">Search Charts</h3>
        <div className="flex gap-2 mb-4">
          <div className="relative flex-1">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400" />
            <input
              value={query}
              onChange={(e) => setQuery(e.target.value)}
              onKeyDown={(e) => e.key === 'Enter' && searchCharts()}
              placeholder="Search charts..."
              className="glass-input pl-10"
            />
          </div>
          <button onClick={searchCharts} className="glass-btn-primary">Search</button>
        </div>
        {charts.length > 0 && (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-3">
            {charts.map((c) => (
              <motion.div key={`${c.repository}/${c.name}/${c.version}`} initial={{ opacity: 0 }} animate={{ opacity: 1 }} className="p-4 bg-glass-light rounded-xl">
                <div className="flex items-center justify-between">
                  <p className="text-sm font-medium text-white">{c.name}</p>
                  <span className="text-xs text-gray-400">{c.version}</span>
                </div>
                <p className="text-xs text-gray-400 mt-1 line-clamp-2">{c.description}</p>
                <p className="text-xs text-gray-500 mt-2 font-mono">{c.repository}</p>
              </motion.div>
            ))}
          </div>
        )}
      </div>

      {/* Releases */}
      <div className="glass-card p-6">
        <h3 className="text-lg font-semibold text-white mb-4">Releases</h3>
        {releases.length === 0 ? (
          <p className="text-sm text-gray-400">No releases installed.</p>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="text-left text-gray-400 border-b border-glass-border">
                  <th className="py-2 pr-4">Name</th>
                  <th className="py-2 pr-4">Namespace</th>
                  <th className="py-2 pr-4">Chart</th>
                  <th className="py-2 pr-4">Version</th>
                  <th className="py-2 pr-4">Status</th>
                  <th className="py-2 pr-4">Deployed</th>
                </tr>
              </thead>
              <tbody>
                {releases.map((r) => (
                  <tr key={r.id} className="border-b border-glass-border/50">
                    <td className="py-2 pr-4 font-medium text-white">{r.name}</td>
                    <td className="py-2 pr-4 text-gray-400 font-mono">{r.namespace}</td>
                    <td className="py-2 pr-4 text-gray-300">{r.chart_name}</td>
                    <td className="py-2 pr-4 text-gray-400">{r.chart_version}</td>
                    <td className="py-2 pr-4">
                      <span className="status-badge">{r.status}</span>
                    </td>
                    <td className="py-2 pr-4 text-gray-400">{r.last_deployed ? new Date(r.last_deployed).toLocaleDateString() : '-'}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>

      {/* Footer link to catalog */}
      <div className="text-center">
        <a href="https://artifacthub.io" target="_blank" rel="noreferrer" className="inline-flex items-center gap-1 text-sm text-primary-400 hover:text-primary-300">
          Browse charts on Artifact Hub <ExternalLink className="w-3 h-3" />
        </a>
      </div>
    </div>
  )
}
