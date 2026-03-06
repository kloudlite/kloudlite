'use client'

import { useState, useEffect, useRef, useCallback, useMemo } from 'react'
import { Terminal, X, Loader2, Download, Trash2, Pause, Play, Search, Filter, RefreshCw, ArrowDownToLine, ChevronsDown } from 'lucide-react'
import { Button } from '@kloudlite/ui'
import { cn } from '@kloudlite/lib'
import type { LogEntry, LogLevel } from './log-parsing'
import { parseLogLine, levelColors } from './log-parsing'
import { LogEntryRow } from './log-entry-row'

interface ServiceLogsViewerProps {
  serviceName: string
  namespace: string
  isOpen: boolean
  onClose: () => void
}

export function ServiceLogsViewer({
  serviceName,
  namespace,
  isOpen,
  onClose,
}: ServiceLogsViewerProps) {
  const [logs, setLogs] = useState<LogEntry[]>([])
  const [isPaused, setIsPaused] = useState(false)
  const [searchTerm, setSearchTerm] = useState('')
  const [levelFilter, setLevelFilter] = useState<LogLevel | 'all'>('all')
  const [showFilters, setShowFilters] = useState(false)
  const [stickyScroll, setStickyScroll] = useState(true)
  const [isAtBottom, setIsAtBottom] = useState(true)
  const [pausedCount, setPausedCount] = useState(0)
  const logsContainerRef = useRef<HTMLDivElement>(null)
  const pausedLogsRef = useRef<LogEntry[]>([])
  const idCounterRef = useRef(0)
  const userScrolledRef = useRef(false)

  const [isConnected, setIsConnected] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [, setReconnectAttempt] = useState(0)
  const [isReconnecting, setIsReconnecting] = useState(false)
  const eventSourceRef = useRef<EventSource | null>(null)
  const isPausedRef = useRef(isPaused)
  const stickyScrollRef = useRef(stickyScroll)
  const reconnectTimeoutRef = useRef<NodeJS.Timeout | null>(null)
  const createConnectionRef = useRef<(clearLogs?: boolean) => void>(() => {})
  const maxReconnectAttempts = 10
  const baseReconnectDelay = 1000 // 1 second

  // Keep refs in sync with state
  useEffect(() => {
    isPausedRef.current = isPaused
  }, [isPaused])

  useEffect(() => {
    stickyScrollRef.current = stickyScroll
  }, [stickyScroll])

  const scrollToBottom = useCallback((force: boolean = false) => {
    if (logsContainerRef.current && (force || (stickyScrollRef.current && !isPausedRef.current))) {
      logsContainerRef.current.scrollTop = logsContainerRef.current.scrollHeight
      setIsAtBottom(true)
    }
  }, [])

  // Check if scrolled to bottom (with small threshold for rounding errors)
  const checkIfAtBottom = useCallback(() => {
    if (!logsContainerRef.current) return true
    const { scrollTop, scrollHeight, clientHeight } = logsContainerRef.current
    return scrollHeight - scrollTop - clientHeight < 20
  }, [])

  // Handle scroll events to detect manual scrolling
  const handleScroll = useCallback(() => {
    if (!logsContainerRef.current) return

    const atBottom = checkIfAtBottom()
    setIsAtBottom(atBottom)

    // If user scrolled up manually, disable sticky scroll
    if (!atBottom && !userScrolledRef.current) {
      userScrolledRef.current = true
      setStickyScroll(false)
    }

    // If user scrolled back to bottom manually, re-enable sticky scroll
    if (atBottom && userScrolledRef.current) {
      userScrolledRef.current = false
      setStickyScroll(true)
    }
  }, [checkIfAtBottom])

  // Jump to bottom and re-enable sticky scroll
  const jumpToBottom = useCallback(() => {
    setStickyScroll(true)
    userScrolledRef.current = false
    scrollToBottom(true)
  }, [scrollToBottom])

  const handleLog = useCallback(
    (data: string) => {
      const entry = parseLogLine(data, idCounterRef.current++)

      if (isPausedRef.current) {
        pausedLogsRef.current.push(entry)
        setPausedCount(pausedLogsRef.current.length)
      } else {
        setLogs((prev) => [...prev, entry])
        setTimeout(scrollToBottom, 0)
      }
    },
    [scrollToBottom]
  )

  // Function to create and manage SSE connection
  const createConnection = useCallback((clearLogs: boolean = false) => {
    if (!serviceName || !namespace) return

    // Close existing connection
    if (eventSourceRef.current) {
      eventSourceRef.current.close()
      eventSourceRef.current = null
    }

    // Clear reconnect timeout
    if (reconnectTimeoutRef.current) {
      clearTimeout(reconnectTimeoutRef.current)
      reconnectTimeoutRef.current = null
    }

    if (clearLogs) {
      setLogs([])
      pausedLogsRef.current = []
      idCounterRef.current = 0
    }
    setError(null)
    setIsReconnecting(false)

    const url = `/api/v1/namespaces/${encodeURIComponent(namespace)}/services/${encodeURIComponent(serviceName)}/logs?tailLines=200&follow=true`
    const eventSource = new EventSource(url)
    eventSourceRef.current = eventSource

    eventSource.onopen = () => {
      setIsConnected(true)
      setError(null)
      setReconnectAttempt(0)
      setIsReconnecting(false)
    }

    eventSource.onmessage = (event) => {
      if (event.data) {
        handleLog(event.data)
      }
    }

    eventSource.onerror = () => {
      setIsConnected(false)

      // Check if connection is closed (not just temporarily errored)
      if (eventSource.readyState === EventSource.CLOSED) {
        // Attempt reconnect with exponential backoff
        setReconnectAttempt((prev) => {
          const newAttempt = prev + 1

          if (newAttempt <= maxReconnectAttempts) {
            setIsReconnecting(true)
            setError(`Reconnecting... (attempt ${newAttempt}/${maxReconnectAttempts})`)

            // Exponential backoff: 1s, 2s, 4s, 8s, max 30s
            const delay = Math.min(baseReconnectDelay * Math.pow(2, newAttempt - 1), 30000)

            reconnectTimeoutRef.current = setTimeout(() => {
              createConnectionRef.current(false) // Don't clear logs on reconnect
            }, delay)
          } else {
            setIsReconnecting(false)
            setError('Connection lost. Click reconnect to try again.')
          }

          return newAttempt
        })
      }
    }
  }, [serviceName, namespace, handleLog])

  useEffect(() => {
    createConnectionRef.current = createConnection
  }, [createConnection])

  // Manual reconnect function
  const handleReconnect = useCallback(() => {
    setReconnectAttempt(0)
    createConnection(false)
  }, [createConnection])

  // SSE connection effect
  useEffect(() => {
    if (!isOpen || !serviceName || !namespace) {
      if (eventSourceRef.current) {
        eventSourceRef.current.close()
        eventSourceRef.current = null
      }
      if (reconnectTimeoutRef.current) {
        clearTimeout(reconnectTimeoutRef.current)
        reconnectTimeoutRef.current = null
      }
      const frame = requestAnimationFrame(() => {
        setIsConnected(false)
        setIsReconnecting(false)
        setReconnectAttempt(0)
      })
      return () => cancelAnimationFrame(frame)
    }

    const frame = requestAnimationFrame(() => {
      // Create initial connection and clear logs
      createConnection(true)
    })

    return () => {
      cancelAnimationFrame(frame)
      if (eventSourceRef.current) {
        eventSourceRef.current.close()
        eventSourceRef.current = null
      }
      if (reconnectTimeoutRef.current) {
        clearTimeout(reconnectTimeoutRef.current)
        reconnectTimeoutRef.current = null
      }
      setIsConnected(false)
      setIsReconnecting(false)
    }
  }, [isOpen, serviceName, namespace, createConnection])

  // Handle pause/resume - flush paused logs when resuming
  useEffect(() => {
    if (!isPaused && pausedLogsRef.current.length > 0) {
      setLogs((prev) => [...prev, ...pausedLogsRef.current])
      pausedLogsRef.current = []
      setPausedCount(0)
      setTimeout(scrollToBottom, 0)
    }
  }, [isPaused, scrollToBottom])

  // Filter logs
  const filteredLogs = useMemo(() => {
    return logs.filter((entry) => {
      // Level filter
      if (levelFilter !== 'all' && entry.level !== levelFilter) {
        return false
      }

      // Search filter
      if (searchTerm) {
        const searchLower = searchTerm.toLowerCase()
        return (
          entry.message.toLowerCase().includes(searchLower) ||
          entry.raw.toLowerCase().includes(searchLower)
        )
      }

      return true
    })
  }, [logs, levelFilter, searchTerm])

  // Count logs by level for filter badges
  const levelCounts = useMemo(() => {
    const counts: Record<LogLevel, number> = {
      error: 0,
      warn: 0,
      info: 0,
      debug: 0,
      trace: 0,
      unknown: 0,
    }
    logs.forEach((entry) => {
      counts[entry.level]++
    })
    return counts
  }, [logs])

  const handleClear = () => {
    setLogs([])
    pausedLogsRef.current = []
    setPausedCount(0)
  }

  const handleDownload = () => {
    const content = logs.map((e) => e.raw).join('\n')
    const blob = new Blob([content], { type: 'text/plain' })
    const downloadUrl = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = downloadUrl
    a.download = `${serviceName}-logs-${new Date().toISOString().slice(0, 19).replace(/:/g, '-')}.txt`
    document.body.appendChild(a)
    a.click()
    document.body.removeChild(a)
    URL.revokeObjectURL(downloadUrl)
  }

  const handleClose = () => {
    setLogs([])
    onClose()
  }

  if (!isOpen) return null

  const isConnecting = !isConnected && !error && !isReconnecting

  return (
    <div className="bg-background fixed inset-x-0 bottom-0 z-50 flex h-[40vh] min-h-[300px] flex-col border-t">
      {/* Header */}
      <div className="bg-muted/50 flex items-center justify-between border-b px-4 py-2">
        <div className="flex items-center gap-3">
          <Terminal className="h-4 w-4" />
          <span className="text-sm font-medium">Logs: {serviceName}</span>
          {isConnecting && (
            <span className="text-muted-foreground flex items-center gap-1 text-xs">
              <Loader2 className="h-3 w-3 animate-spin" />
              Connecting...
            </span>
          )}
          {isConnected && (
            <span className="flex items-center gap-1 text-xs text-green-500">
              <span className="h-2 w-2 animate-pulse rounded-full bg-green-500" />
              Live
            </span>
          )}
          {isReconnecting && (
            <span className="flex items-center gap-1 text-xs text-yellow-500">
              <Loader2 className="h-3 w-3 animate-spin" />
              {error || 'Reconnecting...'}
            </span>
          )}
          {error && !isReconnecting && (
            <span className="flex items-center gap-2 text-xs text-red-500">
              {error}
              <button
                onClick={handleReconnect}
                className="flex items-center gap-1 rounded bg-red-500/20 px-2 py-0.5 hover:bg-red-500/30 transition-colors"
              >
                <RefreshCw className="h-3 w-3" />
                Reconnect
              </button>
            </span>
          )}
          {isPaused && (
            <span className="text-muted-foreground text-xs">
              (Paused - {pausedCount} new)
            </span>
          )}
          <span className="text-muted-foreground text-xs">
            {filteredLogs.length}/{logs.length} logs
          </span>
        </div>
        <div className="flex items-center gap-1">
          <Button
            variant="ghost"
            size="sm"
            onClick={() => setShowFilters(!showFilters)}
            title="Toggle filters"
            className={cn(showFilters && 'bg-muted')}
          >
            <Filter className="h-4 w-4" />
          </Button>
          <Button
            variant="ghost"
            size="sm"
            onClick={() => {
              setStickyScroll(!stickyScroll)
              if (!stickyScroll) {
                // If enabling sticky scroll, jump to bottom
                scrollToBottom(true)
              }
            }}
            title={stickyScroll ? 'Disable auto-scroll' : 'Enable auto-scroll'}
            className={cn(stickyScroll && 'bg-muted text-green-500')}
          >
            <ArrowDownToLine className="h-4 w-4" />
          </Button>
          <Button
            variant="ghost"
            size="sm"
            onClick={() => setIsPaused(!isPaused)}
            title={isPaused ? 'Resume' : 'Pause'}
          >
            {isPaused ? <Play className="h-4 w-4" /> : <Pause className="h-4 w-4" />}
          </Button>
          <Button variant="ghost" size="sm" onClick={handleClear} title="Clear logs">
            <Trash2 className="h-4 w-4" />
          </Button>
          <Button
            variant="ghost"
            size="sm"
            onClick={handleDownload}
            disabled={logs.length === 0}
            title="Download logs"
          >
            <Download className="h-4 w-4" />
          </Button>
          <Button variant="ghost" size="sm" onClick={handleClose} title="Close">
            <X className="h-4 w-4" />
          </Button>
        </div>
      </div>

      {/* Filter bar */}
      {showFilters && (
        <div className="bg-muted/30 flex items-center gap-4 border-b px-4 py-2">
          {/* Search */}
          <div className="relative flex-1 max-w-xs">
            <Search className="absolute left-2 top-1/2 -translate-y-1/2 h-3 w-3 text-muted-foreground" />
            <input
              type="text"
              placeholder="Search logs..."
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
              className="w-full h-7 pl-7 pr-2 text-xs bg-background border rounded focus:outline-none focus:ring-1 focus:ring-ring"
            />
          </div>

          {/* Level filter buttons */}
          <div className="flex items-center gap-1">
            <span className="text-xs text-muted-foreground mr-1">Level:</span>
            <button
              onClick={() => setLevelFilter('all')}
              className={cn(
                'px-2 py-0.5 text-xs rounded transition-colors',
                levelFilter === 'all'
                  ? 'bg-zinc-700 text-white'
                  : 'text-muted-foreground hover:text-foreground'
              )}
            >
              All
            </button>
            {(['error', 'warn', 'info', 'debug'] as const).map((level) => {
              const colors = levelColors[level]
              const count = levelCounts[level]
              return (
                <button
                  key={level}
                  onClick={() => setLevelFilter(level)}
                  className={cn(
                    'px-2 py-0.5 text-xs rounded transition-colors flex items-center gap-1',
                    levelFilter === level
                      ? cn(colors.bg, colors.text)
                      : 'text-muted-foreground hover:text-foreground'
                  )}
                >
                  {level}
                  {count > 0 && (
                    <span className={cn('text-[10px]', levelFilter === level ? colors.text : 'text-muted-foreground')}>
                      ({count})
                    </span>
                  )}
                </button>
              )
            })}
          </div>

          {/* Clear filters */}
          {(searchTerm || levelFilter !== 'all') && (
            <button
              onClick={() => {
                setSearchTerm('')
                setLevelFilter('all')
              }}
              className="text-xs text-muted-foreground hover:text-foreground"
            >
              Clear filters
            </button>
          )}
        </div>
      )}

      {/* Logs content */}
      <div className="relative flex-1">
        <div
          ref={logsContainerRef}
          onScroll={handleScroll}
          className="absolute inset-0 overflow-auto bg-zinc-950 font-mono text-xs"
        >
          {filteredLogs.length === 0 ? (
            <div className="text-muted-foreground flex h-full items-center justify-center">
              {logs.length === 0
                ? isConnecting
                  ? 'Connecting to log stream...'
                  : 'Waiting for logs...'
                : 'No logs match the current filters'}
            </div>
          ) : (
            <div>
              {filteredLogs.map((entry) => (
                <LogEntryRow key={entry.id} entry={entry} searchTerm={searchTerm} />
              ))}
            </div>
          )}
        </div>

        {/* Jump to bottom button - shows when not at bottom and logs exist */}
        {!isAtBottom && logs.length > 0 && (
          <button
            onClick={jumpToBottom}
            className="absolute bottom-4 right-4 flex items-center gap-2 rounded-full bg-zinc-800 px-3 py-1.5 text-xs text-zinc-300 shadow-lg transition-all hover:bg-zinc-700 hover:text-white border border-zinc-700"
          >
            <ChevronsDown className="h-3.5 w-3.5" />
            Jump to latest
          </button>
        )}
      </div>
    </div>
  )
}
