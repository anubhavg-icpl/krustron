// Krustron Dashboard - Audit Logs
// Author: Anubhav Gain <anubhavg@infopercept.com>

import { useEffect, useState } from 'react'
import { motion } from 'framer-motion'
import { ScrollText, RefreshCw } from 'lucide-react'
import { auditApi, ApiClientError } from '@/api'
import type { AuditLog } from '@/api'
import { showErrorToast } from '@/hooks/useNotificationToasts'

export default function Audit() {
  const [logs, setLogs] = useState<AuditLog[]>([])
  const [loading, setLoading] = useState(true)
  const [page, setPage] = useState(1)

  const refresh = async (p = page) => {
    setLoading(true)
    try {
      const data = await auditApi.listLogs({ page: p, limit: 50 })
      setLogs(data)
    } catch (e) {
      // Audit is admin-only; a 403 just means the viewer lacks access.
      const msg = e instanceof ApiClientError ? e.message : 'Network error'
      if (!(e instanceof ApiClientError && e.status === 403)) {
        showErrorToast('Failed to load audit logs', msg)
      }
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    refresh()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [page])

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-white">Audit Logs</h1>
          <p className="text-gray-400 mt-1">All administrative and control-plane actions</p>
        </div>
        <button onClick={() => refresh()} className="glass-btn flex items-center gap-2">
          <RefreshCw className={loading ? 'w-4 h-4 animate-spin' : 'w-4 h-4'} />
          Refresh
        </button>
      </div>

      <div className="glass-card p-6">
        {logs.length === 0 ? (
          <div className="flex flex-col items-center justify-center h-40 text-gray-400">
            <ScrollText className="w-8 h-8 mb-2 opacity-50" />
            <p className="text-sm">{loading ? 'Loading…' : 'No audit entries (admin role required).'}</p>
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="text-left text-gray-400 border-b border-glass-border">
                  <th className="py-2 pr-4">Time</th>
                  <th className="py-2 pr-4">User</th>
                  <th className="py-2 pr-4">Action</th>
                  <th className="py-2 pr-4">Resource</th>
                  <th className="py-2 pr-4">Status</th>
                  <th className="py-2 pr-4">IP</th>
                </tr>
              </thead>
              <tbody>
                {logs.map((log) => (
                  <motion.tr key={log.id} initial={{ opacity: 0 }} animate={{ opacity: 1 }} className="border-b border-glass-border/50">
                    <td className="py-2 pr-4 text-gray-400 whitespace-nowrap">{new Date(log.created_at).toLocaleString()}</td>
                    <td className="py-2 pr-4 text-gray-300">{log.user_email}</td>
                    <td className="py-2 pr-4 font-mono text-white">{log.action}</td>
                    <td className="py-2 pr-4 text-gray-300">
                      {log.resource_type}
                      {log.resource_name ? ` / ${log.resource_name}` : ''}
                    </td>
                    <td className="py-2 pr-4">
                      <span className="status-badge">{log.status}</span>
                    </td>
                    <td className="py-2 pr-4 text-gray-400 font-mono">{log.ip_address}</td>
                  </motion.tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>

      <div className="flex items-center justify-between">
        <button
          onClick={() => setPage((p) => Math.max(1, p - 1))}
          disabled={page <= 1}
          className="glass-btn disabled:opacity-50"
        >
          Previous
        </button>
        <span className="text-sm text-gray-400">Page {page}</span>
        <button
          onClick={() => setPage((p) => p + 1)}
          disabled={logs.length < 50}
          className="glass-btn disabled:opacity-50"
        >
          Next
        </button>
      </div>
    </div>
  )
}
