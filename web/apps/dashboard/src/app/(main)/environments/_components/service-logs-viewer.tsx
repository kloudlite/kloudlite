'use client'

import { useState, useEffect, useRef, useCallback, useMemo } from 'react'
import { Terminal, X, Loader2, Download, Trash2, Pause, Play, Search, ChevronDown, ChevronRight, Filter } from 'lucide-react'
import { Button } from '@kloudlite/ui'
import { cn } from '@kloudlite/lib'
import { useWebSocket } from '@/lib/hooks/use-websocket'

interface ServiceLogsViewerProps {
  serviceName: string
  namespace: string
  isOpen: boolean
  onClose: () => void
}

interface LogMessage {
  type: string
  data?: string
  pod?: string
  container?: string
  error?: string
}

type LogLevel = 'error' | 'warn' | 'info' | 'debug' | 'trace' | 'unknown'

interface LogEntry {
  id: number
  raw: string
  timestamp?: Date
  level: LogLevel
  message: string
  isJson: boolean
  parsed?: Record<string, unknown>
}

// Parse a log line and extract structured data
function parseLogLine(raw: string, id: number): LogEntry {
  // Try to parse as JSON first
  try {
    const parsed = JSON.parse(raw)
    if (typeof parsed === 'object' && parsed !== null) {
      // Extract common fields from various JSON log formats
      const level = extractLevel(parsed)
      const message = extractMessage(parsed, raw)
      const timestamp = extractTimestamp(parsed)

      return {
        id,
        raw,
        timestamp,
        level,
        message,
        isJson: true,
        parsed,
      }
    }
  } catch {
    // Not JSON, parse as plain text
  }

  // Plain text log - try to detect level from content
  const level = detectLevelFromText(raw)

  return {
    id,
    raw,
    level,
    message: raw,
    isJson: false,
  }
}

// Extract log level from parsed JSON
function extractLevel(parsed: Record<string, unknown>): LogLevel {
  // Common level field names
  const levelFields = ['level', 'lvl', 'severity', 's', 'log.level', 'loglevel']

  for (const field of levelFields) {
    const value = getNestedValue(parsed, field)
    if (value) {
      const normalized = String(value).toLowerCase()
      if (normalized.includes('error') || normalized.includes('err') || normalized === 'e') return 'error'
      if (normalized.includes('warn') || normalized === 'w') return 'warn'
      if (normalized.includes('info') || normalized === 'i') return 'info'
      if (normalized.includes('debug') || normalized === 'd') return 'debug'
      if (normalized.includes('trace') || normalized === 't') return 'trace'
    }
  }

  return 'unknown'
}

// Extract message from parsed JSON
function extractMessage(parsed: Record<string, unknown>, raw: string): string {
  // For MongoDB logs, check for nested message in attr.message.msg first
  const nestedMsg = getNestedValue(parsed, 'attr.message.msg')
  if (nestedMsg && typeof nestedMsg === 'string') {
    return nestedMsg
  }

  // Common message field names
  const messageFields = ['message', 'msg', 'text', 'log', 'error', 'err']

  for (const field of messageFields) {
    const value = getNestedValue(parsed, field)
    if (value && typeof value === 'string') {
      return value
    }
  }

  // Fallback to raw
  return raw
}

// Extract timestamp from parsed JSON
function extractTimestamp(parsed: Record<string, unknown>): Date | undefined {
  // Common timestamp field names
  const tsFields = ['timestamp', 'time', 'ts', 't', '@timestamp', 'datetime', 'date']

  for (const field of tsFields) {
    const value = getNestedValue(parsed, field)
    if (value) {
      // Handle MongoDB's $date format
      if (typeof value === 'object' && value !== null && '$date' in value) {
        const date = new Date((value as { $date: string }).$date)
        if (!isNaN(date.getTime())) return date
      }
      // Handle string timestamps
      if (typeof value === 'string' || typeof value === 'number') {
        const date = new Date(value)
        if (!isNaN(date.getTime())) return date
      }
    }
  }

  return undefined
}

// Get nested value from object using dot notation
function getNestedValue(obj: Record<string, unknown>, path: string): unknown {
  const parts = path.split('.')
  let current: unknown = obj

  for (const part of parts) {
    if (current && typeof current === 'object' && part in current) {
      current = (current as Record<string, unknown>)[part]
    } else {
      return undefined
    }
  }

  return current
}

// Detect log level from plain text
function detectLevelFromText(text: string): LogLevel {
  const lower = text.toLowerCase()

  // Check for level indicators
  if (/\b(error|err|fatal|critical|panic)\b/i.test(lower)) return 'error'
  if (/\b(warn|warning)\b/i.test(lower)) return 'warn'
  if (/\b(info)\b/i.test(lower)) return 'info'
  if (/\b(debug|dbg)\b/i.test(lower)) return 'debug'
  if (/\b(trace|verbose)\b/i.test(lower)) return 'trace'

  return 'unknown'
}

// Format timestamp for display
function formatTimestamp(date: Date): string {
  return date.toLocaleTimeString('en-US', {
    hour12: false,
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
  })
}

// Level badge colors
const levelColors: Record<LogLevel, { bg: string; text: string; border: string }> = {
  error: { bg: 'bg-red-500/20', text: 'text-red-400', border: 'border-red-500/30' },
  warn: { bg: 'bg-yellow-500/20', text: 'text-yellow-400', border: 'border-yellow-500/30' },
  info: { bg: 'bg-blue-500/20', text: 'text-blue-400', border: 'border-blue-500/30' },
  debug: { bg: 'bg-zinc-500/20', text: 'text-zinc-400', border: 'border-zinc-500/30' },
  trace: { bg: 'bg-purple-500/20', text: 'text-purple-400', border: 'border-purple-500/30' },
  unknown: { bg: 'bg-zinc-500/10', text: 'text-zinc-500', border: 'border-zinc-500/20' },
}

// Log entry component with expandable JSON
function LogEntryRow({ entry, searchTerm }: { entry: LogEntry; searchTerm: string }) {
  const [expanded, setExpanded] = useState(false)
  const colors = levelColors[entry.level]

  // Highlight search term in text
  const highlightText = (text: string) => {
    if (!searchTerm) return text
    const regex = new RegExp(`(${searchTerm.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')})`, 'gi')
    const parts = text.split(regex)
    return parts.map((part, i) =>
      regex.test(part) ? (
        <mark key={i} className="bg-yellow-500/40 text-yellow-200 rounded px-0.5">{part}</mark>
      ) : part
    )
  }

  return (
    <div className={cn('group border-b border-zinc-800/50 hover:bg-zinc-900/50')}>
      <div className="flex items-start gap-2 py-1 px-2">
        {/* Expand button for JSON logs */}
        {entry.isJson ? (
          <button
            onClick={() => setExpanded(!expanded)}
            className="mt-0.5 text-zinc-500 hover:text-zinc-300 flex-shrink-0"
          >
            {expanded ? <ChevronDown className="h-3 w-3" /> : <ChevronRight className="h-3 w-3" />}
          </button>
        ) : (
          <span className="w-3 flex-shrink-0" />
        )}

        {/* Timestamp */}
        {entry.timestamp && (
          <span className="text-zinc-500 flex-shrink-0 w-16">
            {formatTimestamp(entry.timestamp)}
          </span>
        )}

        {/* Level badge */}
        <span className={cn(
          'text-[10px] font-medium uppercase px-1.5 py-0.5 rounded flex-shrink-0 w-12 text-center',
          colors.bg, colors.text
        )}>
          {entry.level === 'unknown' ? '---' : entry.level}
        </span>

        {/* Message (extracted for JSON, raw for plain text) */}
        <span className={cn('flex-1 break-all', entry.level === 'error' ? 'text-red-400' : entry.level === 'warn' ? 'text-yellow-400' : 'text-zinc-300')}>
          {highlightText(entry.message)}
        </span>
      </div>

      {/* Expanded JSON view */}
      {expanded && entry.isJson && entry.parsed && (
        <div className="ml-8 mr-2 mb-2 p-2 bg-zinc-900 rounded border border-zinc-800 overflow-x-auto">
          <pre className="text-[10px] text-zinc-400">
            {JSON.stringify(entry.parsed, null, 2)}
          </pre>
        </div>
      )}
    </div>
  )
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
  const logsContainerRef = useRef<HTMLDivElement>(null)
  const pausedLogsRef = useRef<LogEntry[]>([])
  const idCounterRef = useRef(0)

  const scrollToBottom = useCallback(() => {
    if (logsContainerRef.current && !isPaused) {
      logsContainerRef.current.scrollTop = logsContainerRef.current.scrollHeight
    }
  }, [isPaused])

  const url = useMemo(() => {
    if (!isOpen) return null
    return `/api/v1/namespaces/${encodeURIComponent(namespace)}/services/${encodeURIComponent(serviceName)}/logs-ws?tailLines=200`
  }, [isOpen, namespace, serviceName])

  const handleLog = useCallback(
    (data: string) => {
      const entry = parseLogLine(data, idCounterRef.current++)

      if (isPaused) {
        pausedLogsRef.current.push(entry)
      } else {
        setLogs((prev) => [...prev, entry])
        setTimeout(scrollToBottom, 0)
      }
    },
    [isPaused, scrollToBottom]
  )

  const eventHandlers = useMemo(
    () => ({
      log: (data: string | LogMessage) => {
        if (typeof data === 'string') {
          handleLog(data)
        } else if (data?.data) {
          handleLog(data.data)
        }
      },
      connected: () => {
        // Connection confirmed
      },
      error: (data: string | { error?: string }) => {
        const errorMsg = typeof data === 'string' ? data : data?.error || 'Unknown error'
        console.error('Log stream error:', errorMsg)
      },
    }),
    [handleLog]
  )

  const { isConnected, error } = useWebSocket<LogMessage>(url, {
    enabled: isOpen,
    eventHandlers,
  })

  // Handle pause/resume
  useEffect(() => {
    if (!isPaused && pausedLogsRef.current.length > 0) {
      setLogs((prev) => [...prev, ...pausedLogsRef.current])
      pausedLogsRef.current = []
      setTimeout(scrollToBottom, 0)
    }
  }, [isPaused, scrollToBottom])

  // Clear logs when closed or when service changes
  useEffect(() => {
    if (!isOpen) {
      setLogs([])
      pausedLogsRef.current = []
      idCounterRef.current = 0
    }
  }, [isOpen])

  // Clear logs when service or namespace changes
  useEffect(() => {
    setLogs([])
    pausedLogsRef.current = []
    idCounterRef.current = 0
  }, [serviceName, namespace])

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

  const isConnecting = !isConnected && !error

  return (
    <div className="bg-background fixed inset-x-0 bottom-0 z-50 flex h-[40vh] min-h-[300px] flex-col border-t shadow-lg">
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
          {error && <span className="text-xs text-red-500">{error}</span>}
          {isPaused && (
            <span className="text-muted-foreground text-xs">
              (Paused - {pausedLogsRef.current.length} new)
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
      <div
        ref={logsContainerRef}
        className="flex-1 overflow-auto bg-zinc-950 font-mono text-xs"
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
    </div>
  )
}
