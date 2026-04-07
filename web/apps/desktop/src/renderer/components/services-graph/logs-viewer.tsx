import { useState, useEffect, useRef } from 'react'
import { X, Pause, Play, Trash2, Download } from 'lucide-react'
import { cn } from '@/lib/utils'
import { Sheet } from '@/components/ui'

interface LogsViewerProps {
  serviceName: string
  onClose: () => void
}

const LOG_LEVELS = ['INFO', 'DEBUG', 'WARN', 'ERROR'] as const
type LogLevel = typeof LOG_LEVELS[number]

interface LogLine {
  ts: string
  level: LogLevel
  message: string
}

const SAMPLE_LOGS: LogLine[] = [
  { ts: '14:32:15.234', level: 'INFO', message: 'Server starting on port 8080' },
  { ts: '14:32:15.456', level: 'INFO', message: 'Connected to postgres at postgres:5432' },
  { ts: '14:32:15.478', level: 'INFO', message: 'Connected to redis at redis:6379' },
  { ts: '14:32:15.512', level: 'DEBUG', message: 'Loaded 12 environment variables' },
  { ts: '14:32:15.621', level: 'INFO', message: 'gRPC server listening on :50051' },
  { ts: '14:32:16.001', level: 'INFO', message: 'Health check endpoint registered at /healthz' },
  { ts: '14:32:18.234', level: 'INFO', message: 'Request: GET /api/users 200 12ms' },
  { ts: '14:32:18.567', level: 'INFO', message: 'Request: POST /api/auth/login 200 45ms' },
  { ts: '14:32:19.123', level: 'WARN', message: 'Slow query detected: SELECT * FROM users took 234ms' },
  { ts: '14:32:20.456', level: 'INFO', message: 'Request: GET /api/sessions 200 8ms' },
  { ts: '14:32:21.789', level: 'DEBUG', message: 'Cache hit: user:42 (ttl 280s)' },
  { ts: '14:32:22.012', level: 'ERROR', message: 'Failed to connect to upstream: connection refused' },
  { ts: '14:32:22.234', level: 'INFO', message: 'Retrying upstream connection (attempt 1/3)' },
  { ts: '14:32:23.456', level: 'INFO', message: 'Upstream reconnected successfully' },
  { ts: '14:32:24.678', level: 'INFO', message: 'Request: GET /api/users/42 200 15ms' },
  { ts: '14:32:25.901', level: 'DEBUG', message: 'Trace: span_id=a1b2c3 duration=45ms' },
  { ts: '14:32:26.234', level: 'INFO', message: 'Request: PUT /api/users/42 200 89ms' },
  { ts: '14:32:27.567', level: 'WARN', message: 'Rate limit approaching for IP 10.0.0.5: 95/100 req/min' },
  { ts: '14:32:28.890', level: 'INFO', message: 'Background job: cleanup_sessions completed (47 removed)' },
  { ts: '14:32:30.123', level: 'INFO', message: 'Metrics flushed: cpu=42% mem=512MB requests=1247' },
]

const levelStyles: Record<LogLevel, string> = {
  INFO: 'text-blue-400',
  DEBUG: 'text-muted-foreground/60',
  WARN: 'text-amber-400',
  ERROR: 'text-red-400',
}

export function LogsViewer({ serviceName, onClose }: LogsViewerProps) {
  const [paused, setPaused] = useState(false)
  const [logs, setLogs] = useState<LogLine[]>(SAMPLE_LOGS)
  const [filter, setFilter] = useState<LogLevel | 'ALL'>('ALL')
  const [search, setSearch] = useState('')
  const scrollRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    if (paused) return
    const interval = setInterval(() => {
      const samples = [
        { level: 'INFO' as LogLevel, message: `Request: GET /api/items/${Math.floor(Math.random() * 1000)} 200 ${Math.floor(Math.random() * 50)}ms` },
        { level: 'DEBUG' as LogLevel, message: `Cache lookup: key=session:${Math.random().toString(36).slice(2, 10)}` },
        { level: 'INFO' as LogLevel, message: `Heartbeat: ok` },
      ]
      const sample = samples[Math.floor(Math.random() * samples.length)]
      const now = new Date()
      const ts = `${String(now.getHours()).padStart(2, '0')}:${String(now.getMinutes()).padStart(2, '0')}:${String(now.getSeconds()).padStart(2, '0')}.${String(now.getMilliseconds()).padStart(3, '0')}`
      setLogs((prev) => [...prev.slice(-200), { ts, ...sample }])
    }, 1500)
    return () => clearInterval(interval)
  }, [paused])

  useEffect(() => {
    if (!paused && scrollRef.current) {
      scrollRef.current.scrollTop = scrollRef.current.scrollHeight
    }
  }, [logs, paused])

  const filtered = logs.filter((l) => {
    if (filter !== 'ALL' && l.level !== filter) return false
    if (search && !l.message.toLowerCase().includes(search.toLowerCase())) return false
    return true
  })

  return (
    <Sheet side="right" width="640px" onClose={onClose}>
      {(close) => (
        <div className="flex h-full flex-col">
          {/* Header */}
          <div className="flex items-center gap-3 border-b border-border/30 px-5 py-3">
            <div>
              <h2 className="text-[15px] font-semibold text-foreground">{serviceName}</h2>
              <p className="text-[11px] text-muted-foreground">Live logs · {filtered.length} lines</p>
            </div>
            <div className="ml-auto flex items-center gap-1.5">
              <button
                className="flex h-7 w-7 items-center justify-center rounded-md text-muted-foreground transition-colors hover:bg-accent hover:text-foreground"
                onClick={() => setPaused(!paused)}
                title={paused ? 'Resume' : 'Pause'}
              >
                {paused ? <Play className="h-3.5 w-3.5" /> : <Pause className="h-3.5 w-3.5" />}
              </button>
              <button
                className="flex h-7 w-7 items-center justify-center rounded-md text-muted-foreground transition-colors hover:bg-accent hover:text-foreground"
                onClick={() => setLogs([])}
                title="Clear"
              >
                <Trash2 className="h-3.5 w-3.5" />
              </button>
              <button
                className="flex h-7 w-7 items-center justify-center rounded-md text-muted-foreground transition-colors hover:bg-accent hover:text-foreground"
                title="Download"
              >
                <Download className="h-3.5 w-3.5" />
              </button>
              <div className="mx-1 h-5 w-px bg-border" />
              <button
                className="flex h-7 w-7 items-center justify-center rounded-md text-muted-foreground transition-colors hover:bg-accent hover:text-foreground"
                onClick={close}
              >
                <X className="h-4 w-4" />
              </button>
            </div>
          </div>

          {/* Toolbar */}
          <div className="flex items-center gap-2 border-b border-border/30 px-5 py-2">
            <input
              type="text"
              className="flex-1 rounded-md border border-border bg-background px-2.5 py-1 text-[12px] outline-none focus:border-primary"
              placeholder="Search logs..."
              value={search}
              onChange={(e) => setSearch(e.target.value)}
            />
            <div className="flex items-center gap-0.5 rounded-md bg-accent/50 p-0.5">
              <button
                className={cn(
                  'rounded px-2 py-0.5 text-[11px] font-medium transition-colors',
                  filter === 'ALL' ? 'bg-background text-foreground shadow-sm' : 'text-muted-foreground hover:text-foreground'
                )}
                onClick={() => setFilter('ALL')}
              >
                All
              </button>
              {LOG_LEVELS.map((lvl) => (
                <button
                  key={lvl}
                  className={cn(
                    'rounded px-2 py-0.5 text-[11px] font-medium transition-colors',
                    filter === lvl ? 'bg-background text-foreground shadow-sm' : 'text-muted-foreground hover:text-foreground'
                  )}
                  onClick={() => setFilter(lvl)}
                >
                  {lvl}
                </button>
              ))}
            </div>
          </div>

          {/* Log content */}
          <div
            ref={scrollRef}
            className="flex-1 overflow-auto bg-[#1e1e2e] px-4 py-3 font-mono text-[11px] leading-[18px]"
            style={{ fontFamily: 'SF Mono, Fira Code, Menlo, Consolas, monospace' }}
          >
            {filtered.map((log, i) => (
              <div key={i} className="flex gap-2 hover:bg-white/5">
                <span className="shrink-0 text-[#6c7086]">{log.ts}</span>
                <span className={cn('shrink-0 font-semibold', levelStyles[log.level])}>{log.level.padEnd(5)}</span>
                <span className="text-[#cdd6f4]">{log.message}</span>
              </div>
            ))}
            {filtered.length === 0 && (
              <p className="text-center text-[#6c7086]">No logs match the filter</p>
            )}
          </div>
        </div>
      )}
    </Sheet>
  )
}
