export type LogLevel = 'error' | 'warn' | 'info' | 'debug' | 'trace' | 'unknown'

export interface LogEntry {
  id: number
  raw: string
  timestamp?: Date
  level: LogLevel
  message: string
  isJson: boolean
  parsed?: Record<string, unknown>
}

// Parse a log line and extract structured data
export function parseLogLine(raw: string, id: number): LogEntry {
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
export function formatTimestamp(date: Date): string {
  return date.toLocaleTimeString('en-US', {
    hour12: false,
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
  })
}

// Level badge colors
export const levelColors: Record<LogLevel, { bg: string; text: string; border: string }> = {
  error: { bg: 'bg-red-500/20', text: 'text-red-400', border: 'border-red-500/30' },
  warn: { bg: 'bg-yellow-500/20', text: 'text-yellow-400', border: 'border-yellow-500/30' },
  info: { bg: 'bg-blue-500/20', text: 'text-blue-400', border: 'border-blue-500/30' },
  debug: { bg: 'bg-zinc-500/20', text: 'text-zinc-400', border: 'border-zinc-500/30' },
  trace: { bg: 'bg-purple-500/20', text: 'text-purple-400', border: 'border-purple-500/30' },
  unknown: { bg: 'bg-zinc-500/10', text: 'text-zinc-500', border: 'border-zinc-500/20' },
}
