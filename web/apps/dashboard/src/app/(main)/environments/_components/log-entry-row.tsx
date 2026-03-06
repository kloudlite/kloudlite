'use client'

import { useState } from 'react'
import { ChevronDown, ChevronRight } from 'lucide-react'
import { cn } from '@kloudlite/lib'
import type { LogEntry } from './log-parsing'
import { formatTimestamp, levelColors } from './log-parsing'

interface LogEntryRowProps {
  entry: LogEntry
  searchTerm: string
}

// Log entry component with expandable JSON
export function LogEntryRow({ entry, searchTerm }: LogEntryRowProps) {
  const [expanded, setExpanded] = useState(false)
  const colors = levelColors[entry.level]

  // Highlight search term in text
  const highlightText = (text: string) => {
    if (!searchTerm) return text
    const regex = new RegExp(`(${searchTerm.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')})`, 'gi')
    const parts = text.split(regex)
    return parts.map((part, i) =>
      regex.test(part) ? (
        <mark key={`${text}-${i}`} className="bg-yellow-500/40 text-yellow-200 rounded px-0.5">{part}</mark>
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
