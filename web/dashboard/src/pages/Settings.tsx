// Krustron Dashboard - Settings Page
// Author: Anubhav Gain <anubhavg@infopercept.com>

import { useState, useEffect } from 'react'
import { Routes, Route, NavLink, useLocation } from 'react-router-dom'
import { motion } from 'framer-motion'
import {
  User,
  Bell,
  Shield,
  Palette,
  Key,
  Users,
  Webhook,
  Server,
  Save,
  Eye,
  EyeOff,
  Plus,
  Trash2,
  Edit2,
  Loader2,
  CheckCircle,
  XCircle,
  RefreshCw,
} from 'lucide-react'
import { clsx } from 'clsx'
import { useAuthStore, useUIStore } from '@/store/useStore'
import { showSuccessToast, showErrorToast } from '@/hooks/useNotificationToasts'

// Settings Navigation
const settingsNav = [
  { path: '/settings', icon: User, label: 'Profile', exact: true },
  { path: '/settings/notifications', icon: Bell, label: 'Notifications' },
  { path: '/settings/security', icon: Shield, label: 'Security' },
  { path: '/settings/appearance', icon: Palette, label: 'Appearance' },
  { path: '/settings/api-keys', icon: Key, label: 'API Keys' },
  { path: '/settings/teams', icon: Users, label: 'Teams' },
  { path: '/settings/webhooks', icon: Webhook, label: 'Webhooks' },
]

// Settings Layout
function SettingsLayout({ children }: { children: React.ReactNode }) {
  const location = useLocation()

  return (
    <div className="flex flex-col lg:flex-row gap-6">
      {/* Sidebar */}
      <div className="lg:w-64 flex-shrink-0">
        <nav className="glass-card p-2 space-y-1">
          {settingsNav.map((item) => {
            const isActive = item.exact
              ? location.pathname === item.path
              : location.pathname.startsWith(item.path)

            return (
              <NavLink
                key={item.path}
                to={item.path}
                className={clsx(
                  'flex items-center gap-3 px-4 py-3 rounded-xl transition-all duration-200',
                  isActive
                    ? 'bg-primary-500/20 text-primary-400'
                    : 'text-gray-400 hover:bg-glass-light hover:text-white'
                )}
              >
                <item.icon className="w-5 h-5" />
                <span className="font-medium">{item.label}</span>
              </NavLink>
            )
          })}
        </nav>
      </div>

      {/* Content */}
      <div className="flex-1">{children}</div>
    </div>
  )
}

// Profile Settings
function ProfileSettings() {
  const { user, updateUser } = useAuthStore()
  const [formData, setFormData] = useState({
    displayName: user?.displayName || '',
    email: user?.email || '',
    username: user?.username || '',
  })

  const handleSave = () => {
    updateUser(formData)
  }

  return (
    <SettingsLayout>
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        className="glass-card p-6"
      >
        <h2 className="text-xl font-bold text-white mb-6">Profile Settings</h2>

        <div className="space-y-6">
          {/* Avatar */}
          <div className="flex items-center gap-6">
            <div className="w-20 h-20 rounded-2xl bg-gradient-to-br from-primary-500 to-accent-500 flex items-center justify-center">
              <User className="w-10 h-10 text-white" />
            </div>
            <div>
              <button className="glass-btn">Change Avatar</button>
              <p className="text-sm text-gray-400 mt-2">JPG, PNG or GIF. Max 2MB.</p>
            </div>
          </div>

          {/* Form Fields */}
          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            <div>
              <label className="block text-sm font-medium text-gray-400 mb-2">
                Display Name
              </label>
              <input
                type="text"
                value={formData.displayName}
                onChange={(e) => setFormData({ ...formData, displayName: e.target.value })}
                className="glass-input"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-400 mb-2">
                Username
              </label>
              <input
                type="text"
                value={formData.username}
                onChange={(e) => setFormData({ ...formData, username: e.target.value })}
                className="glass-input"
              />
            </div>
            <div className="md:col-span-2">
              <label className="block text-sm font-medium text-gray-400 mb-2">
                Email
              </label>
              <input
                type="email"
                value={formData.email}
                onChange={(e) => setFormData({ ...formData, email: e.target.value })}
                className="glass-input"
              />
            </div>
          </div>

          <div className="flex justify-end">
            <button onClick={handleSave} className="glass-btn-primary flex items-center gap-2">
              <Save className="w-4 h-4" />
              Save Changes
            </button>
          </div>
        </div>
      </motion.div>
    </SettingsLayout>
  )
}

// Appearance Settings
function AppearanceSettings() {
  const { theme, setTheme, sidebarCollapsed, setSidebarCollapsed } = useUIStore()

  return (
    <SettingsLayout>
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        className="glass-card p-6"
      >
        <h2 className="text-xl font-bold text-white mb-6">Appearance</h2>

        <div className="space-y-6">
          {/* Theme */}
          <div>
            <label className="block text-sm font-medium text-gray-400 mb-4">Theme</label>
            <div className="grid grid-cols-2 gap-4">
              <button
                onClick={() => setTheme('dark')}
                className={clsx(
                  'p-4 rounded-xl border-2 transition-all',
                  theme === 'dark'
                    ? 'border-primary-500 bg-primary-500/10'
                    : 'border-glass-border hover:border-primary-500/50'
                )}
              >
                <div className="w-full h-20 bg-surface rounded-lg mb-2" />
                <span className="text-sm text-white">Dark</span>
              </button>
              <button
                onClick={() => setTheme('light')}
                className={clsx(
                  'p-4 rounded-xl border-2 transition-all',
                  theme === 'light'
                    ? 'border-primary-500 bg-primary-500/10'
                    : 'border-glass-border hover:border-primary-500/50'
                )}
              >
                <div className="w-full h-20 bg-gray-100 rounded-lg mb-2" />
                <span className="text-sm text-white">Light</span>
              </button>
            </div>
          </div>

          {/* Sidebar */}
          <div className="flex items-center justify-between p-4 bg-glass-light rounded-xl">
            <div>
              <h4 className="font-medium text-white">Collapsed Sidebar</h4>
              <p className="text-sm text-gray-400">Show only icons in the sidebar</p>
            </div>
            <button
              onClick={() => setSidebarCollapsed(!sidebarCollapsed)}
              className={clsx(
                'relative w-12 h-6 rounded-full transition-colors',
                sidebarCollapsed ? 'bg-primary-500' : 'bg-glass-heavy'
              )}
            >
              <span
                className={clsx(
                  'absolute top-1 w-4 h-4 rounded-full bg-white transition-all',
                  sidebarCollapsed ? 'left-7' : 'left-1'
                )}
              />
            </button>
          </div>
        </div>
      </motion.div>
    </SettingsLayout>
  )
}

// Security Settings
function SecuritySettings() {
  const [showPassword, setShowPassword] = useState(false)

  return (
    <SettingsLayout>
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        className="space-y-6"
      >
        {/* Change Password */}
        <div className="glass-card p-6">
          <h2 className="text-xl font-bold text-white mb-6">Change Password</h2>

          <div className="space-y-4 max-w-md">
            <div>
              <label className="block text-sm font-medium text-gray-400 mb-2">
                Current Password
              </label>
              <div className="relative">
                <input
                  type={showPassword ? 'text' : 'password'}
                  className="glass-input pr-10"
                />
                <button
                  onClick={() => setShowPassword(!showPassword)}
                  className="absolute right-3 top-1/2 -translate-y-1/2"
                >
                  {showPassword ? (
                    <EyeOff className="w-4 h-4 text-gray-400" />
                  ) : (
                    <Eye className="w-4 h-4 text-gray-400" />
                  )}
                </button>
              </div>
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-400 mb-2">
                New Password
              </label>
              <input type="password" className="glass-input" />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-400 mb-2">
                Confirm New Password
              </label>
              <input type="password" className="glass-input" />
            </div>
            <button className="glass-btn-primary">Update Password</button>
          </div>
        </div>

        {/* Two-Factor Auth */}
        <div className="glass-card p-6">
          <div className="flex items-center justify-between">
            <div>
              <h3 className="text-lg font-semibold text-white">Two-Factor Authentication</h3>
              <p className="text-sm text-gray-400 mt-1">
                Add an extra layer of security to your account
              </p>
            </div>
            <button className="glass-btn-primary">Enable 2FA</button>
          </div>
        </div>

        {/* Active Sessions */}
        <div className="glass-card p-6">
          <h3 className="text-lg font-semibold text-white mb-4">Active Sessions</h3>
          <div className="space-y-3">
            <div className="flex items-center justify-between p-4 bg-glass-light rounded-xl">
              <div className="flex items-center gap-3">
                <Server className="w-5 h-5 text-gray-400" />
                <div>
                  <p className="text-sm font-medium text-white">Current Session</p>
                  <p className="text-xs text-gray-400">macOS - Chrome - San Francisco, CA</p>
                </div>
              </div>
              <span className="status-badge status-healthy">Active</span>
            </div>
          </div>
        </div>
      </motion.div>
    </SettingsLayout>
  )
}

// Notifications Settings
function NotificationSettings() {
  const [settings, setSettings] = useState({
    emailAlerts: true,
    slackNotifications: true,
    pipelineUpdates: true,
    securityAlerts: true,
    costAlerts: false,
  })

  const toggleSetting = (key: keyof typeof settings) => {
    setSettings({ ...settings, [key]: !settings[key] })
  }

  return (
    <SettingsLayout>
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        className="glass-card p-6"
      >
        <h2 className="text-xl font-bold text-white mb-6">Notification Preferences</h2>

        <div className="space-y-4">
          {Object.entries(settings).map(([key, value]) => (
            <div
              key={key}
              className="flex items-center justify-between p-4 bg-glass-light rounded-xl"
            >
              <div>
                <h4 className="font-medium text-white capitalize">
                  {key.replace(/([A-Z])/g, ' $1').trim()}
                </h4>
                <p className="text-sm text-gray-400">
                  Receive notifications for {key.replace(/([A-Z])/g, ' $1').toLowerCase()}
                </p>
              </div>
              <button
                onClick={() => toggleSetting(key as keyof typeof settings)}
                className={clsx(
                  'relative w-12 h-6 rounded-full transition-colors',
                  value ? 'bg-primary-500' : 'bg-glass-heavy'
                )}
              >
                <span
                  className={clsx(
                    'absolute top-1 w-4 h-4 rounded-full bg-white transition-all',
                    value ? 'left-7' : 'left-1'
                  )}
                />
              </button>
            </div>
          ))}
        </div>
      </motion.div>
    </SettingsLayout>
  )
}

// API Keys Settings
function APIKeysSettings() {
  return (
    <SettingsLayout>
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        className="glass-card p-6"
      >
        <div className="flex items-center justify-between mb-6">
          <h2 className="text-xl font-bold text-white">API Keys</h2>
          <button className="glass-btn-primary flex items-center gap-2">
            <Key className="w-4 h-4" />
            Generate New Key
          </button>
        </div>

        <div className="text-center py-12 text-gray-400">
          <Key className="w-12 h-12 mx-auto mb-4 opacity-50" />
          <p>No API keys generated yet</p>
          <p className="text-sm mt-2">Generate your first API key to get started</p>
        </div>
      </motion.div>
    </SettingsLayout>
  )
}

// Teams Settings
interface TeamMember {
  id: string
  email: string
  name: string
  role: string
  avatar_url?: string
  is_active: boolean
  last_login_at?: string
  created_at: string
}

interface Role {
  id: string
  name: string
  display_name: string
  description: string
  permissions: string[]
  is_system: boolean
}

function TeamsSettings() {
  const { token } = useAuthStore()
  const [members, setMembers] = useState<TeamMember[]>([])
  const [roles, setRoles] = useState<Role[]>([])
  const [loading, setLoading] = useState(true)
  const [showInviteModal, setShowInviteModal] = useState(false)
  const [inviteForm, setInviteForm] = useState({ email: '', name: '', password: '', role: 'user' })
  const [isSubmitting, setIsSubmitting] = useState(false)

  useEffect(() => {
    fetchTeamData()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [token])

  const fetchTeamData = async () => {
    setLoading(true)
    try {
      const [usersRes, rolesRes] = await Promise.all([
        fetch('/api/v1/users', {
          headers: { 'Authorization': `Bearer ${token}` },
        }),
        fetch('/api/v1/rbac/roles', {
          headers: { 'Authorization': `Bearer ${token}` },
        }),
      ])

      if (usersRes.ok) {
        const usersData = await usersRes.json()
        setMembers(usersData.data || [])
      }

      if (rolesRes.ok) {
        const rolesData = await rolesRes.json()
        setRoles(rolesData.data || [])
      }
    } catch (error) {
      console.error('Failed to fetch team data:', error)
    } finally {
      setLoading(false)
    }
  }

  const handleInvite = async (e: React.FormEvent) => {
    e.preventDefault()
    setIsSubmitting(true)

    try {
      const response = await fetch('/api/v1/users', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}`,
        },
        body: JSON.stringify(inviteForm),
      })

      if (response.ok) {
        showSuccessToast('User invited', `${inviteForm.email} has been added to the team`)
        setShowInviteModal(false)
        setInviteForm({ email: '', name: '', password: '', role: 'user' })
        fetchTeamData()
      } else {
        const data = await response.json().catch(() => ({}))
        showErrorToast('Failed to invite user', data.message || 'Please try again')
      }
    } catch (error) {
      showErrorToast('Failed to invite user', 'Network error')
    } finally {
      setIsSubmitting(false)
    }
  }

  const handleDeleteUser = async (id: string, email: string) => {
    if (!confirm(`Are you sure you want to remove ${email}?`)) return

    try {
      const response = await fetch(`/api/v1/users/${id}`, {
        method: 'DELETE',
        headers: { 'Authorization': `Bearer ${token}` },
      })

      if (response.ok) {
        showSuccessToast('User removed', `${email} has been removed from the team`)
        fetchTeamData()
      } else {
        showErrorToast('Failed to remove user', 'Please try again')
      }
    } catch (error) {
      showErrorToast('Failed to remove user', 'Network error')
    }
  }

  return (
    <SettingsLayout>
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        className="space-y-6"
      >
        {/* Team Members */}
        <div className="glass-card p-6">
          <div className="flex items-center justify-between mb-6">
            <h2 className="text-xl font-bold text-white">Team Members</h2>
            <button
              onClick={() => setShowInviteModal(true)}
              className="glass-btn-primary flex items-center gap-2"
            >
              <Plus className="w-4 h-4" />
              Invite User
            </button>
          </div>

          {loading ? (
            <div className="flex items-center justify-center py-12">
              <Loader2 className="w-8 h-8 animate-spin text-primary-500" />
            </div>
          ) : members.length === 0 ? (
            <div className="text-center py-12 text-gray-400">
              <Users className="w-12 h-12 mx-auto mb-4 opacity-50" />
              <p>No team members found</p>
            </div>
          ) : (
            <div className="space-y-3">
              {members.map((member) => (
                <div
                  key={member.id}
                  className="flex items-center justify-between p-4 bg-glass-light rounded-xl"
                >
                  <div className="flex items-center gap-4">
                    <div className="w-10 h-10 rounded-lg bg-gradient-to-br from-primary-500 to-accent-500 flex items-center justify-center">
                      <User className="w-5 h-5 text-white" />
                    </div>
                    <div>
                      <p className="font-medium text-white">{member.name || member.email}</p>
                      <p className="text-sm text-gray-400">{member.email}</p>
                    </div>
                  </div>
                  <div className="flex items-center gap-4">
                    <span className={clsx(
                      'status-badge',
                      member.is_active ? 'status-healthy' : 'status-error'
                    )}>
                      {member.is_active ? 'Active' : 'Inactive'}
                    </span>
                    <span className="px-3 py-1 bg-primary-500/20 text-primary-400 rounded-lg text-sm capitalize">
                      {member.role}
                    </span>
                    <button
                      onClick={() => handleDeleteUser(member.id, member.email)}
                      className="p-2 hover:bg-glass-medium rounded-lg transition-colors text-gray-400 hover:text-status-error"
                    >
                      <Trash2 className="w-4 h-4" />
                    </button>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>

        {/* Roles */}
        <div className="glass-card p-6">
          <div className="flex items-center justify-between mb-6">
            <h2 className="text-xl font-bold text-white">Roles</h2>
            <button className="glass-btn flex items-center gap-2">
              <Plus className="w-4 h-4" />
              Create Role
            </button>
          </div>

          {roles.length === 0 ? (
            <div className="text-center py-12 text-gray-400">
              <Shield className="w-12 h-12 mx-auto mb-4 opacity-50" />
              <p>No roles defined</p>
            </div>
          ) : (
            <div className="space-y-3">
              {roles.map((role) => (
                <div
                  key={role.id}
                  className="flex items-center justify-between p-4 bg-glass-light rounded-xl"
                >
                  <div>
                    <p className="font-medium text-white">{role.display_name || role.name}</p>
                    <p className="text-sm text-gray-400">{role.description}</p>
                    <div className="flex flex-wrap gap-1 mt-2">
                      {role.permissions.slice(0, 5).map((perm) => (
                        <span key={perm} className="px-2 py-0.5 bg-glass-medium rounded text-xs text-gray-400">
                          {perm}
                        </span>
                      ))}
                      {role.permissions.length > 5 && (
                        <span className="px-2 py-0.5 bg-glass-medium rounded text-xs text-gray-400">
                          +{role.permissions.length - 5} more
                        </span>
                      )}
                    </div>
                  </div>
                  <div className="flex items-center gap-2">
                    {role.is_system && (
                      <span className="px-2 py-1 bg-gray-600/50 text-gray-400 rounded text-xs">
                        System
                      </span>
                    )}
                    {!role.is_system && (
                      <button className="p-2 hover:bg-glass-medium rounded-lg transition-colors text-gray-400 hover:text-white">
                        <Edit2 className="w-4 h-4" />
                      </button>
                    )}
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      </motion.div>

      {/* Invite Modal */}
      {showInviteModal && (
        <div className="fixed inset-0 bg-black/60 backdrop-blur-sm z-50 flex items-center justify-center p-4">
          <motion.div
            initial={{ opacity: 0, scale: 0.95 }}
            animate={{ opacity: 1, scale: 1 }}
            className="glass-card p-6 w-full max-w-md"
          >
            <h3 className="text-lg font-bold text-white mb-4">Invite Team Member</h3>
            <form onSubmit={handleInvite} className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-400 mb-2">Email</label>
                <input
                  type="email"
                  value={inviteForm.email}
                  onChange={(e) => setInviteForm({ ...inviteForm, email: e.target.value })}
                  className="glass-input"
                  required
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-400 mb-2">Name</label>
                <input
                  type="text"
                  value={inviteForm.name}
                  onChange={(e) => setInviteForm({ ...inviteForm, name: e.target.value })}
                  className="glass-input"
                  required
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-400 mb-2">Password</label>
                <input
                  type="password"
                  value={inviteForm.password}
                  onChange={(e) => setInviteForm({ ...inviteForm, password: e.target.value })}
                  className="glass-input"
                  required
                  minLength={8}
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-400 mb-2">Role</label>
                <select
                  value={inviteForm.role}
                  onChange={(e) => setInviteForm({ ...inviteForm, role: e.target.value })}
                  className="glass-input"
                >
                  <option value="user">User</option>
                  <option value="admin">Admin</option>
                  {roles.map((role) => (
                    <option key={role.id} value={role.name}>{role.display_name || role.name}</option>
                  ))}
                </select>
              </div>
              <div className="flex gap-3 pt-2">
                <button
                  type="button"
                  onClick={() => setShowInviteModal(false)}
                  className="glass-btn flex-1"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  disabled={isSubmitting}
                  className="glass-btn-primary flex-1 flex items-center justify-center gap-2"
                >
                  {isSubmitting ? (
                    <>
                      <Loader2 className="w-4 h-4 animate-spin" />
                      Inviting...
                    </>
                  ) : (
                    'Invite'
                  )}
                </button>
              </div>
            </form>
          </motion.div>
        </div>
      )}
    </SettingsLayout>
  )
}

// Webhooks Settings
interface WebhookConfig {
  id: string
  name: string
  url: string
  events: string[]
  secret?: string
  active: boolean
  created_at: string
}

function WebhooksSettings() {
  const { token } = useAuthStore()
  const [webhooks, setWebhooks] = useState<WebhookConfig[]>([])
  const [loading, setLoading] = useState(true)
  const [showCreateModal, setShowCreateModal] = useState(false)
  const [webhookForm, setWebhookForm] = useState({
    name: '',
    url: '',
    events: ['pipeline.success', 'pipeline.failure'],
    secret: '',
  })
  const [isSubmitting, setIsSubmitting] = useState(false)

  const availableEvents = [
    { value: 'pipeline.success', label: 'Pipeline Success' },
    { value: 'pipeline.failure', label: 'Pipeline Failure' },
    { value: 'pipeline.started', label: 'Pipeline Started' },
    { value: 'deployment.success', label: 'Deployment Success' },
    { value: 'deployment.failure', label: 'Deployment Failure' },
    { value: 'cluster.connected', label: 'Cluster Connected' },
    { value: 'cluster.disconnected', label: 'Cluster Disconnected' },
    { value: 'alert.critical', label: 'Critical Alert' },
    { value: 'alert.high', label: 'High Alert' },
    { value: 'security.vulnerability', label: 'Security Vulnerability' },
  ]

  useEffect(() => {
    fetchWebhooks()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [token])

  const fetchWebhooks = async () => {
    setLoading(true)
    try {
      // Get notification settings which includes webhook config
      const response = await fetch('/api/v1/settings/notifications', {
        headers: { 'Authorization': `Bearer ${token}` },
      })

      if (response.ok) {
        const data = await response.json()
        // Parse webhooks from settings if available
        const webhookData = data.data?.webhooks || []
        setWebhooks(Array.isArray(webhookData) ? webhookData : [])
      }
    } catch (error) {
      console.error('Failed to fetch webhooks:', error)
    } finally {
      setLoading(false)
    }
  }

  const handleCreateWebhook = async (e: React.FormEvent) => {
    e.preventDefault()
    setIsSubmitting(true)

    try {
      // Create a new webhook entry
      const newWebhook: WebhookConfig = {
        id: `wh-${Date.now()}`,
        name: webhookForm.name,
        url: webhookForm.url,
        events: webhookForm.events,
        secret: webhookForm.secret || undefined,
        active: true,
        created_at: new Date().toISOString(),
      }

      // Update settings with new webhook
      const response = await fetch('/api/v1/settings/notifications', {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}`,
        },
        body: JSON.stringify({
          webhook: {
            enabled: true,
            webhooks: [...webhooks, newWebhook],
          },
        }),
      })

      if (response.ok) {
        showSuccessToast('Webhook created', webhookForm.name)
        setShowCreateModal(false)
        setWebhookForm({ name: '', url: '', events: ['pipeline.success', 'pipeline.failure'], secret: '' })
        setWebhooks([...webhooks, newWebhook])
      } else {
        showErrorToast('Failed to create webhook', 'Please try again')
      }
    } catch (error) {
      showErrorToast('Failed to create webhook', 'Network error')
    } finally {
      setIsSubmitting(false)
    }
  }

  const toggleWebhook = async (webhook: WebhookConfig) => {
    const updatedWebhooks = webhooks.map((w) =>
      w.id === webhook.id ? { ...w, active: !w.active } : w
    )
    setWebhooks(updatedWebhooks)
    showSuccessToast(
      webhook.active ? 'Webhook disabled' : 'Webhook enabled',
      webhook.name
    )
  }

  const deleteWebhook = async (webhook: WebhookConfig) => {
    if (!confirm(`Are you sure you want to delete "${webhook.name}"?`)) return

    const updatedWebhooks = webhooks.filter((w) => w.id !== webhook.id)
    setWebhooks(updatedWebhooks)
    showSuccessToast('Webhook deleted', webhook.name)
  }

  const generateSecret = () => {
    const array = new Uint8Array(32)
    crypto.getRandomValues(array)
    const secret = Array.from(array, (byte) => byte.toString(16).padStart(2, '0')).join('')
    setWebhookForm({ ...webhookForm, secret })
  }

  const toggleEvent = (event: string) => {
    if (webhookForm.events.includes(event)) {
      setWebhookForm({
        ...webhookForm,
        events: webhookForm.events.filter((e) => e !== event),
      })
    } else {
      setWebhookForm({
        ...webhookForm,
        events: [...webhookForm.events, event],
      })
    }
  }

  return (
    <SettingsLayout>
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        className="space-y-6"
      >
        {/* Webhooks List */}
        <div className="glass-card p-6">
          <div className="flex items-center justify-between mb-6">
            <div>
              <h2 className="text-xl font-bold text-white">Webhooks</h2>
              <p className="text-sm text-gray-400 mt-1">
                Receive real-time notifications for events
              </p>
            </div>
            <button
              onClick={() => setShowCreateModal(true)}
              className="glass-btn-primary flex items-center gap-2"
            >
              <Plus className="w-4 h-4" />
              Add Webhook
            </button>
          </div>

          {loading ? (
            <div className="flex items-center justify-center py-12">
              <Loader2 className="w-8 h-8 animate-spin text-primary-500" />
            </div>
          ) : webhooks.length === 0 ? (
            <div className="text-center py-12 text-gray-400">
              <Webhook className="w-12 h-12 mx-auto mb-4 opacity-50" />
              <p>No webhooks configured</p>
              <p className="text-sm mt-2">Add a webhook to receive event notifications</p>
            </div>
          ) : (
            <div className="space-y-3">
              {webhooks.map((webhook) => (
                <div
                  key={webhook.id}
                  className="p-4 bg-glass-light rounded-xl"
                >
                  <div className="flex items-start justify-between">
                    <div className="flex-1">
                      <div className="flex items-center gap-3">
                        <h4 className="font-medium text-white">{webhook.name}</h4>
                        <span className={clsx(
                          'status-badge',
                          webhook.active ? 'status-healthy' : 'status-unknown'
                        )}>
                          {webhook.active ? 'Active' : 'Inactive'}
                        </span>
                      </div>
                      <p className="text-sm text-gray-400 mt-1 font-mono truncate">
                        {webhook.url}
                      </p>
                      <div className="flex flex-wrap gap-1 mt-2">
                        {webhook.events.map((event) => (
                          <span
                            key={event}
                            className="px-2 py-0.5 bg-primary-500/20 text-primary-400 rounded text-xs"
                          >
                            {event}
                          </span>
                        ))}
                      </div>
                    </div>
                    <div className="flex items-center gap-2">
                      <button
                        onClick={() => toggleWebhook(webhook)}
                        className={clsx(
                          'p-2 rounded-lg transition-colors',
                          webhook.active
                            ? 'hover:bg-glass-medium text-status-healthy'
                            : 'hover:bg-glass-medium text-gray-400'
                        )}
                        title={webhook.active ? 'Disable' : 'Enable'}
                      >
                        {webhook.active ? (
                          <CheckCircle className="w-4 h-4" />
                        ) : (
                          <XCircle className="w-4 h-4" />
                        )}
                      </button>
                      <button
                        onClick={() => deleteWebhook(webhook)}
                        className="p-2 hover:bg-glass-medium rounded-lg transition-colors text-gray-400 hover:text-status-error"
                      >
                        <Trash2 className="w-4 h-4" />
                      </button>
                    </div>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>

        {/* Integration Guides */}
        <div className="glass-card p-6">
          <h3 className="text-lg font-semibold text-white mb-4">Integration Guides</h3>
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            <div className="p-4 bg-glass-light rounded-xl">
              <div className="w-10 h-10 rounded-lg bg-[#4A154B]/20 flex items-center justify-center mb-3">
                <svg className="w-6 h-6 text-[#E01E5A]" viewBox="0 0 24 24" fill="currentColor">
                  <path d="M5.042 15.165a2.528 2.528 0 0 1-2.52 2.523A2.528 2.528 0 0 1 0 15.165a2.527 2.527 0 0 1 2.522-2.52h2.52v2.52zm1.271 0a2.527 2.527 0 0 1 2.521-2.52 2.527 2.527 0 0 1 2.521 2.52v6.313A2.528 2.528 0 0 1 8.834 24a2.528 2.528 0 0 1-2.521-2.522v-6.313zM8.834 5.042a2.528 2.528 0 0 1-2.521-2.52A2.528 2.528 0 0 1 8.834 0a2.528 2.528 0 0 1 2.521 2.522v2.52H8.834zm0 1.271a2.528 2.528 0 0 1 2.521 2.521 2.528 2.528 0 0 1-2.521 2.521H2.522A2.528 2.528 0 0 1 0 8.834a2.528 2.528 0 0 1 2.522-2.521h6.312zm10.124 2.521a2.528 2.528 0 0 1 2.522-2.521A2.528 2.528 0 0 1 24 8.834a2.528 2.528 0 0 1-2.52 2.521h-2.522V8.834zm-1.271 0a2.528 2.528 0 0 1-2.521 2.521 2.528 2.528 0 0 1-2.521-2.521V2.522A2.528 2.528 0 0 1 15.166 0a2.528 2.528 0 0 1 2.521 2.522v6.312zm-2.521 10.124a2.528 2.528 0 0 1 2.521 2.522A2.528 2.528 0 0 1 15.166 24a2.528 2.528 0 0 1-2.521-2.52v-2.522h2.521zm0-1.271a2.528 2.528 0 0 1-2.521-2.521 2.528 2.528 0 0 1 2.521-2.521h6.312A2.528 2.528 0 0 1 24 15.166a2.528 2.528 0 0 1-2.52 2.521h-6.312z"/>
                </svg>
              </div>
              <h4 className="font-medium text-white">Slack</h4>
              <p className="text-sm text-gray-400 mt-1">Send notifications to Slack channels</p>
            </div>
            <div className="p-4 bg-glass-light rounded-xl">
              <div className="w-10 h-10 rounded-lg bg-[#5865F2]/20 flex items-center justify-center mb-3">
                <svg className="w-6 h-6 text-[#5865F2]" viewBox="0 0 24 24" fill="currentColor">
                  <path d="M20.317 4.37a19.791 19.791 0 0 0-4.885-1.515.074.074 0 0 0-.079.037c-.21.375-.444.864-.608 1.25a18.27 18.27 0 0 0-5.487 0 12.64 12.64 0 0 0-.617-1.25.077.077 0 0 0-.079-.037A19.736 19.736 0 0 0 3.677 4.37a.07.07 0 0 0-.032.027C.533 9.046-.32 13.58.099 18.057a.082.082 0 0 0 .031.057 19.9 19.9 0 0 0 5.993 3.03.078.078 0 0 0 .084-.028 14.09 14.09 0 0 0 1.226-1.994.076.076 0 0 0-.041-.106 13.107 13.107 0 0 1-1.872-.892.077.077 0 0 1-.008-.128 10.2 10.2 0 0 0 .372-.292.074.074 0 0 1 .077-.01c3.928 1.793 8.18 1.793 12.062 0a.074.074 0 0 1 .078.01c.12.098.246.198.373.292a.077.077 0 0 1-.006.127 12.299 12.299 0 0 1-1.873.892.077.077 0 0 0-.041.107c.36.698.772 1.362 1.225 1.993a.076.076 0 0 0 .084.028 19.839 19.839 0 0 0 6.002-3.03.077.077 0 0 0 .032-.054c.5-5.177-.838-9.674-3.549-13.66a.061.061 0 0 0-.031-.03z"/>
                </svg>
              </div>
              <h4 className="font-medium text-white">Discord</h4>
              <p className="text-sm text-gray-400 mt-1">Post updates to Discord servers</p>
            </div>
            <div className="p-4 bg-glass-light rounded-xl">
              <div className="w-10 h-10 rounded-lg bg-gray-500/20 flex items-center justify-center mb-3">
                <Webhook className="w-6 h-6 text-gray-400" />
              </div>
              <h4 className="font-medium text-white">Custom</h4>
              <p className="text-sm text-gray-400 mt-1">Send to any HTTP endpoint</p>
            </div>
          </div>
        </div>
      </motion.div>

      {/* Create Webhook Modal */}
      {showCreateModal && (
        <div className="fixed inset-0 bg-black/60 backdrop-blur-sm z-50 flex items-center justify-center p-4">
          <motion.div
            initial={{ opacity: 0, scale: 0.95 }}
            animate={{ opacity: 1, scale: 1 }}
            className="glass-card p-6 w-full max-w-lg max-h-[90vh] overflow-y-auto"
          >
            <h3 className="text-lg font-bold text-white mb-4">Add Webhook</h3>
            <form onSubmit={handleCreateWebhook} className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-400 mb-2">Name</label>
                <input
                  type="text"
                  value={webhookForm.name}
                  onChange={(e) => setWebhookForm({ ...webhookForm, name: e.target.value })}
                  placeholder="e.g., Slack Notifications"
                  className="glass-input"
                  required
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-400 mb-2">Webhook URL</label>
                <input
                  type="url"
                  value={webhookForm.url}
                  onChange={(e) => setWebhookForm({ ...webhookForm, url: e.target.value })}
                  placeholder="https://hooks.slack.com/services/..."
                  className="glass-input"
                  required
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-400 mb-2">
                  Secret (Optional)
                </label>
                <div className="flex gap-2">
                  <input
                    type="text"
                    value={webhookForm.secret}
                    onChange={(e) => setWebhookForm({ ...webhookForm, secret: e.target.value })}
                    placeholder="Webhook signing secret"
                    className="glass-input flex-1"
                  />
                  <button
                    type="button"
                    onClick={generateSecret}
                    className="glass-btn flex items-center gap-2"
                  >
                    <RefreshCw className="w-4 h-4" />
                    Generate
                  </button>
                </div>
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-400 mb-2">Events</label>
                <div className="grid grid-cols-2 gap-2">
                  {availableEvents.map((event) => (
                    <label
                      key={event.value}
                      className={clsx(
                        'flex items-center gap-2 p-2 rounded-lg cursor-pointer transition-colors',
                        webhookForm.events.includes(event.value)
                          ? 'bg-primary-500/20 border border-primary-500/50'
                          : 'bg-glass-light border border-transparent hover:border-glass-border'
                      )}
                    >
                      <input
                        type="checkbox"
                        checked={webhookForm.events.includes(event.value)}
                        onChange={() => toggleEvent(event.value)}
                        className="sr-only"
                      />
                      <div className={clsx(
                        'w-4 h-4 rounded border flex items-center justify-center',
                        webhookForm.events.includes(event.value)
                          ? 'bg-primary-500 border-primary-500'
                          : 'border-gray-500'
                      )}>
                        {webhookForm.events.includes(event.value) && (
                          <CheckCircle className="w-3 h-3 text-white" />
                        )}
                      </div>
                      <span className="text-sm text-white">{event.label}</span>
                    </label>
                  ))}
                </div>
              </div>
              <div className="flex gap-3 pt-2">
                <button
                  type="button"
                  onClick={() => setShowCreateModal(false)}
                  className="glass-btn flex-1"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  disabled={isSubmitting || webhookForm.events.length === 0}
                  className="glass-btn-primary flex-1 flex items-center justify-center gap-2"
                >
                  {isSubmitting ? (
                    <>
                      <Loader2 className="w-4 h-4 animate-spin" />
                      Creating...
                    </>
                  ) : (
                    'Create Webhook'
                  )}
                </button>
              </div>
            </form>
          </motion.div>
        </div>
      )}
    </SettingsLayout>
  )
}

export default function Settings() {
  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-white">Settings</h1>
        <p className="text-gray-400 mt-1">Manage your account and preferences</p>
      </div>

      <Routes>
        <Route index element={<ProfileSettings />} />
        <Route path="notifications" element={<NotificationSettings />} />
        <Route path="security" element={<SecuritySettings />} />
        <Route path="appearance" element={<AppearanceSettings />} />
        <Route path="api-keys" element={<APIKeysSettings />} />
        <Route path="teams" element={<TeamsSettings />} />
        <Route path="webhooks" element={<WebhooksSettings />} />
      </Routes>
    </div>
  )
}
