// Krustron Dashboard - Layout Component
// Author: Anubhav Gain <anubhavg@infopercept.com>

import { useState } from 'react'
import { Outlet, NavLink, useLocation } from 'react-router-dom'
import { motion, AnimatePresence } from 'framer-motion'
import {
  LayoutDashboard,
  Server,
  Box,
  GitBranch,
  Shield,
  DollarSign,
  Settings,
  Bell,
  Search,
  ChevronLeft,
  ChevronRight,
  LogOut,
  User,
  Menu,
  X,
  Sparkles,
} from 'lucide-react'
import { clsx } from 'clsx'
import { useAuthStore, useUIStore, useAlertsStore } from '@/store/useStore'
import { useWebSocketContext } from '@/hooks/useWebSocket'

// Navigation items
const navItems = [
  { path: '/', icon: LayoutDashboard, label: 'Dashboard', exact: true },
  { path: '/clusters', icon: Server, label: 'Clusters' },
  { path: '/applications', icon: Box, label: 'Applications' },
  { path: '/pipelines', icon: GitBranch, label: 'Pipelines' },
  { path: '/security', icon: Shield, label: 'Security' },
  { path: '/cost', icon: DollarSign, label: 'Cost' },
  { path: '/settings', icon: Settings, label: 'Settings' },
]

export default function Layout() {
  const location = useLocation()
  const { user, logout } = useAuthStore()
  const { sidebarCollapsed, toggleSidebar, toggleCommandPalette, toggleNotificationPanel } = useUIStore()
  const { unreadCount } = useAlertsStore()
  const { isConnected, reconnectCount } = useWebSocketContext()
  const [mobileMenuOpen, setMobileMenuOpen] = useState(false)
  const [userMenuOpen, setUserMenuOpen] = useState(false)

  return (
    <div className="min-h-screen bg-surface flex">
      {/* Sidebar - Desktop */}
      <motion.aside
        initial={false}
        animate={{ width: sidebarCollapsed ? 80 : 280 }}
        className="hidden lg:flex flex-col fixed left-0 top-0 bottom-0 z-40 bg-surface-100 border-r border-glass-border"
      >
        {/* Logo */}
        <div className="h-16 flex items-center justify-between px-4 border-b border-glass-border">
          <NavLink to="/" className="flex items-center gap-3">
            <div className="w-10 h-10 rounded-xl bg-gradient-to-br from-primary-500 to-accent-500 flex items-center justify-center">
              <Sparkles className="w-6 h-6 text-white" />
            </div>
            <AnimatePresence>
              {!sidebarCollapsed && (
                <motion.span
                  initial={{ opacity: 0, width: 0 }}
                  animate={{ opacity: 1, width: 'auto' }}
                  exit={{ opacity: 0, width: 0 }}
                  className="text-xl font-bold text-gradient overflow-hidden whitespace-nowrap"
                >
                  Krustron
                </motion.span>
              )}
            </AnimatePresence>
          </NavLink>
          <button
            onClick={toggleSidebar}
            className="p-2 rounded-lg hover:bg-glass-light transition-colors"
          >
            {sidebarCollapsed ? (
              <ChevronRight className="w-5 h-5 text-gray-400" />
            ) : (
              <ChevronLeft className="w-5 h-5 text-gray-400" />
            )}
          </button>
        </div>

        {/* Navigation */}
        <nav className="flex-1 py-4 px-3 space-y-1 overflow-y-auto">
          {navItems.map((item) => {
            const isActive = item.exact
              ? location.pathname === item.path
              : location.pathname.startsWith(item.path)

            return (
              <NavLink
                key={item.path}
                to={item.path}
                className={clsx(
                  'flex items-center gap-3 px-3 py-3 rounded-xl transition-all duration-200',
                  isActive
                    ? 'bg-primary-500/20 text-primary-400 border-l-2 border-primary-500'
                    : 'text-gray-400 hover:bg-glass-light hover:text-white'
                )}
              >
                <item.icon className="w-5 h-5 flex-shrink-0" />
                <AnimatePresence>
                  {!sidebarCollapsed && (
                    <motion.span
                      initial={{ opacity: 0, width: 0 }}
                      animate={{ opacity: 1, width: 'auto' }}
                      exit={{ opacity: 0, width: 0 }}
                      className="font-medium overflow-hidden whitespace-nowrap"
                    >
                      {item.label}
                    </motion.span>
                  )}
                </AnimatePresence>
              </NavLink>
            )
          })}
        </nav>

        {/* Connection Status */}
        <div className="p-4 border-t border-glass-border">
          <div className={clsx(
            'flex items-center gap-2 text-sm',
            sidebarCollapsed && 'justify-center'
          )}>
            <div className={clsx(
              'w-2 h-2 rounded-full',
              isConnected ? 'bg-status-healthy' : 'bg-status-error',
              isConnected && 'pulse-dot pulse-dot-healthy'
            )} />
            <AnimatePresence>
              {!sidebarCollapsed && (
                <motion.span
                  initial={{ opacity: 0 }}
                  animate={{ opacity: 1 }}
                  exit={{ opacity: 0 }}
                  className={isConnected ? 'text-gray-400' : 'text-status-error'}
                >
                  {isConnected ? 'Connected' : reconnectCount > 0 ? `Reconnecting (${reconnectCount})` : 'Disconnected'}
                </motion.span>
              )}
            </AnimatePresence>
          </div>
        </div>
      </motion.aside>

      {/* Mobile Sidebar */}
      <AnimatePresence>
        {mobileMenuOpen && (
          <>
            <motion.div
              initial={{ opacity: 0 }}
              animate={{ opacity: 1 }}
              exit={{ opacity: 0 }}
              className="lg:hidden fixed inset-0 bg-black/60 backdrop-blur-sm z-40"
              onClick={() => setMobileMenuOpen(false)}
            />
            <motion.aside
              initial={{ x: -280 }}
              animate={{ x: 0 }}
              exit={{ x: -280 }}
              transition={{ type: 'spring', damping: 25, stiffness: 300 }}
              className="lg:hidden fixed left-0 top-0 bottom-0 w-72 z-50 bg-surface-100 border-r border-glass-border flex flex-col"
            >
              <div className="h-16 flex items-center justify-between px-4 border-b border-glass-border">
                <NavLink to="/" className="flex items-center gap-3">
                  <div className="w-10 h-10 rounded-xl bg-gradient-to-br from-primary-500 to-accent-500 flex items-center justify-center">
                    <Sparkles className="w-6 h-6 text-white" />
                  </div>
                  <span className="text-xl font-bold text-gradient">Krustron</span>
                </NavLink>
                <button
                  onClick={() => setMobileMenuOpen(false)}
                  className="p-2 rounded-lg hover:bg-glass-light transition-colors"
                >
                  <X className="w-5 h-5 text-gray-400" />
                </button>
              </div>

              <nav className="flex-1 py-4 px-3 space-y-1 overflow-y-auto">
                {navItems.map((item) => {
                  const isActive = item.exact
                    ? location.pathname === item.path
                    : location.pathname.startsWith(item.path)

                  return (
                    <NavLink
                      key={item.path}
                      to={item.path}
                      onClick={() => setMobileMenuOpen(false)}
                      className={clsx(
                        'flex items-center gap-3 px-3 py-3 rounded-xl transition-all duration-200',
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
            </motion.aside>
          </>
        )}
      </AnimatePresence>

      {/* Main Content */}
      <div
        className={clsx(
          'flex-1 flex flex-col min-h-screen transition-all duration-300',
          sidebarCollapsed ? 'lg:ml-20' : 'lg:ml-[280px]'
        )}
      >
        {/* Header */}
        <header className="h-16 border-b border-glass-border bg-surface-100/80 backdrop-blur-xl sticky top-0 z-30 flex items-center justify-between px-4 lg:px-6">
          {/* Left section */}
          <div className="flex items-center gap-4">
            <button
              onClick={() => setMobileMenuOpen(true)}
              className="lg:hidden p-2 rounded-lg hover:bg-glass-light transition-colors"
            >
              <Menu className="w-5 h-5 text-gray-400" />
            </button>

            {/* Search */}
            <button
              onClick={toggleCommandPalette}
              className="hidden sm:flex items-center gap-2 px-4 py-2 bg-glass-light rounded-xl text-gray-400 hover:bg-glass-medium transition-colors"
            >
              <Search className="w-4 h-4" />
              <span className="text-sm">Search...</span>
              <kbd className="hidden md:inline-flex items-center gap-1 px-2 py-0.5 bg-surface rounded text-xs">
                <span className="text-[10px]">⌘</span>K
              </kbd>
            </button>
          </div>

          {/* Right section */}
          <div className="flex items-center gap-2">
            {/* Notifications */}
            <button
              onClick={toggleNotificationPanel}
              className="relative p-2 rounded-lg hover:bg-glass-light transition-colors"
            >
              <Bell className="w-5 h-5 text-gray-400" />
              {unreadCount > 0 && (
                <span className="absolute -top-0.5 -right-0.5 w-5 h-5 bg-status-error rounded-full flex items-center justify-center text-[10px] font-bold text-white">
                  {unreadCount > 9 ? '9+' : unreadCount}
                </span>
              )}
            </button>

            {/* User menu */}
            <div className="relative">
              <button
                onClick={() => setUserMenuOpen(!userMenuOpen)}
                className="flex items-center gap-2 p-2 rounded-xl hover:bg-glass-light transition-colors"
              >
                <div className="w-8 h-8 rounded-lg bg-gradient-to-br from-primary-500 to-accent-500 flex items-center justify-center">
                  <User className="w-4 h-4 text-white" />
                </div>
                <span className="hidden md:block text-sm font-medium text-gray-300">
                  {user?.displayName || user?.username || 'User'}
                </span>
              </button>

              <AnimatePresence>
                {userMenuOpen && (
                  <motion.div
                    initial={{ opacity: 0, y: 10 }}
                    animate={{ opacity: 1, y: 0 }}
                    exit={{ opacity: 0, y: 10 }}
                    className="absolute right-0 mt-2 w-56 glass-card py-2 z-50"
                  >
                    <div className="px-4 py-2 border-b border-glass-border">
                      <p className="text-sm font-medium text-white">{user?.displayName || user?.username}</p>
                      <p className="text-xs text-gray-400">{user?.email}</p>
                    </div>
                    <NavLink
                      to="/settings"
                      onClick={() => setUserMenuOpen(false)}
                      className="flex items-center gap-2 px-4 py-2 text-sm text-gray-300 hover:bg-glass-light hover:text-white"
                    >
                      <Settings className="w-4 h-4" />
                      Settings
                    </NavLink>
                    <button
                      onClick={() => {
                        setUserMenuOpen(false)
                        logout()
                      }}
                      className="flex items-center gap-2 px-4 py-2 w-full text-sm text-status-error hover:bg-glass-light"
                    >
                      <LogOut className="w-4 h-4" />
                      Logout
                    </button>
                  </motion.div>
                )}
              </AnimatePresence>
            </div>
          </div>
        </header>

        {/* Page Content */}
        <main className="flex-1 p-4 lg:p-6">
          <Outlet />
        </main>
      </div>
    </div>
  )
}
