'use client'

import { useMemo } from 'react'
import {
  Clock,
  HardDrive,
  AlertCircle,
  Loader2,
  RotateCcw,
  Trash2,
  History,
} from 'lucide-react'
import { Button } from '@kloudlite/ui'
import { cn } from '@/lib/utils'
import type { Snapshot } from '@/lib/services/snapshot.service'

interface SnapshotTimelineProps {
  snapshots: Snapshot[]
  onRestore: (snapshot: Snapshot) => void
  onDelete: (snapshot: Snapshot) => void
  disabled?: boolean
  currentSnapshotName?: string
}

interface TimelineNode {
  snapshot: Snapshot
  isCurrent: boolean
}

function formatTimeAgo(dateString: string): string {
  const date = new Date(dateString)
  const now = new Date()
  const diffInSeconds = Math.floor((now.getTime() - date.getTime()) / 1000)

  if (diffInSeconds < 60) return 'just now'
  if (diffInSeconds < 3600) return `${Math.floor(diffInSeconds / 60)}m ago`
  if (diffInSeconds < 86400) return `${Math.floor(diffInSeconds / 3600)}h ago`
  if (diffInSeconds < 604800) return `${Math.floor(diffInSeconds / 86400)}d ago`
  return date.toLocaleDateString()
}

// Extract short hash from snapshot name
function getShortHash(name: string): string {
  const parts = name.split('-')
  if (parts.length >= 2) {
    return parts.slice(-1)[0]
  }
  return name.slice(-6)
}

function getStateBadge(state: Snapshot['status']['state']) {
  switch (state) {
    case 'Ready':
      return (
        <span className="inline-flex items-center rounded px-1.5 py-0.5 text-[10px] font-medium bg-emerald-100 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-400">
          Ready
        </span>
      )
    case 'Creating':
      return (
        <span className="inline-flex items-center gap-1 rounded px-1.5 py-0.5 text-[10px] font-medium bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-400">
          <Loader2 className="h-2.5 w-2.5 animate-spin" />
          Creating
        </span>
      )
    case 'Restoring':
      return (
        <span className="inline-flex items-center gap-1 rounded px-1.5 py-0.5 text-[10px] font-medium bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-400">
          <Loader2 className="h-2.5 w-2.5 animate-spin" />
          Restoring
        </span>
      )
    case 'Deleting':
      return (
        <span className="inline-flex items-center gap-1 rounded px-1.5 py-0.5 text-[10px] font-medium bg-orange-100 text-orange-700 dark:bg-orange-900/30 dark:text-orange-400">
          <Loader2 className="h-2.5 w-2.5 animate-spin" />
          Deleting
        </span>
      )
    case 'Failed':
      return (
        <span className="inline-flex items-center gap-1 rounded px-1.5 py-0.5 text-[10px] font-medium bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400">
          <AlertCircle className="h-2.5 w-2.5" />
          Failed
        </span>
      )
    case 'Pending':
    default:
      return (
        <span className="inline-flex items-center rounded px-1.5 py-0.5 text-[10px] font-medium bg-gray-100 text-gray-600 dark:bg-gray-800 dark:text-gray-400">
          Pending
        </span>
      )
  }
}

function buildTimeline(snapshots: Snapshot[], currentSnapshotName?: string): TimelineNode[] {
  if (snapshots.length === 0) return []

  const snapshotMap = new Map<string, Snapshot>()
  snapshots.forEach(s => snapshotMap.set(s.metadata.name, s))

  const sortedByTime = [...snapshots].sort((a, b) =>
    new Date(b.status.createdAt || b.metadata.creationTimestamp).getTime() -
    new Date(a.status.createdAt || a.metadata.creationTimestamp).getTime()
  )

  const headSnapshotName = currentSnapshotName && snapshotMap.has(currentSnapshotName)
    ? currentSnapshotName
    : sortedByTime[0]?.metadata.name

  return sortedByTime.map((snapshot) => ({
    snapshot,
    isCurrent: snapshot.metadata.name === headSnapshotName,
  }))
}

interface TimelineItemProps {
  node: TimelineNode
  isFirst: boolean
  isLast: boolean
  onRestore: (snapshot: Snapshot) => void
  onDelete: (snapshot: Snapshot) => void
  disabled?: boolean
}

function TimelineItem({ node, isFirst, isLast, onRestore, onDelete, disabled }: TimelineItemProps) {
  const { snapshot, isCurrent } = node
  const shortHash = getShortHash(snapshot.metadata.name)
  const parentHash = snapshot.spec.parentSnapshotRef
    ? getShortHash(snapshot.spec.parentSnapshotRef.name)
    : null

  return (
    <div className="relative flex gap-3">
      {/* Timeline track */}
      <div className="relative flex flex-col items-center w-5 flex-shrink-0">
        {/* Line above */}
        {!isFirst && <div className="w-px flex-1 bg-border" />}
        {isFirst && <div className="flex-1" />}

        {/* Dot */}
        <div
          className={cn(
            "relative z-10 rounded-full flex-shrink-0 transition-all",
            isCurrent
              ? "w-3 h-3 bg-blue-500 ring-[3px] ring-blue-500/20"
              : "w-2 h-2 bg-gray-300 dark:bg-gray-600"
          )}
        />

        {/* Line below */}
        {!isLast && <div className="w-px flex-1 bg-border" />}
        {isLast && <div className="flex-1" />}
      </div>

      {/* Content */}
      <div className="flex-1 pb-4 min-w-0">
        <div
          className={cn(
            "group rounded-lg border p-3 transition-all",
            isCurrent
              ? "bg-blue-50/50 border-blue-200 dark:bg-blue-950/20 dark:border-blue-900"
              : "bg-card hover:bg-muted/30 hover:border-muted-foreground/20"
          )}
        >
          {/* Header row */}
          <div className="flex items-center justify-between gap-2 mb-1.5">
            <div className="flex items-center gap-2 min-w-0">
              {isCurrent && (
                <span className="inline-flex items-center rounded bg-blue-500 px-1.5 py-0.5 text-[10px] font-semibold text-white uppercase tracking-wide">
                  Current
                </span>
              )}
              <code className="text-xs font-mono text-muted-foreground truncate">
                {shortHash}
              </code>
              {getStateBadge(snapshot.status.state)}
            </div>

            {/* Actions */}
            <div className="flex items-center gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
              {snapshot.status.state === 'Ready' && !isCurrent && (
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() => onRestore(snapshot)}
                  disabled={disabled}
                  className="h-7 px-2 text-xs"
                >
                  <RotateCcw className="h-3 w-3 mr-1" />
                  Restore
                </Button>
              )}
              {(snapshot.status.state === 'Ready' || snapshot.status.state === 'Failed') && (
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() => onDelete(snapshot)}
                  className="h-7 w-7 p-0 text-muted-foreground hover:text-destructive"
                >
                  <Trash2 className="h-3.5 w-3.5" />
                </Button>
              )}
            </div>
          </div>

          {/* Description */}
          {snapshot.spec.description && (
            <p className="text-sm text-foreground mb-1.5 line-clamp-2">
              {snapshot.spec.description}
            </p>
          )}

          {/* Meta row */}
          <div className="flex items-center gap-3 text-[11px] text-muted-foreground">
            <span className="flex items-center gap-1">
              <Clock className="h-3 w-3" />
              {formatTimeAgo(snapshot.status.createdAt || snapshot.metadata.creationTimestamp)}
            </span>
            {snapshot.status.sizeHuman && snapshot.status.sizeHuman !== '0 B' && (
              <span className="flex items-center gap-1">
                <HardDrive className="h-3 w-3" />
                {snapshot.status.sizeHuman}
              </span>
            )}
            {parentHash && (
              <span className="text-muted-foreground/70">
                from {parentHash}
              </span>
            )}
          </div>

          {/* Error message */}
          {snapshot.status.state === 'Failed' && snapshot.status.message && (
            <p className="mt-2 text-xs text-red-600 dark:text-red-400 bg-red-50 dark:bg-red-950/30 rounded px-2 py-1.5">
              {snapshot.status.message}
            </p>
          )}
        </div>
      </div>
    </div>
  )
}

export function SnapshotTimeline({ snapshots, onRestore, onDelete, disabled, currentSnapshotName }: SnapshotTimelineProps) {
  const nodes = useMemo(() => buildTimeline(snapshots, currentSnapshotName), [snapshots, currentSnapshotName])

  if (snapshots.length === 0) {
    return null
  }

  return (
    <div>
      <div className="flex items-center gap-2 mb-3">
        <History className="h-4 w-4 text-muted-foreground" />
        <h4 className="text-sm font-medium">History</h4>
        <span className="text-xs text-muted-foreground">({snapshots.length})</span>
      </div>

      <div>
        {nodes.map((node, idx) => (
          <TimelineItem
            key={node.snapshot.metadata.name}
            node={node}
            isFirst={idx === 0}
            isLast={idx === nodes.length - 1}
            onRestore={onRestore}
            onDelete={onDelete}
            disabled={disabled}
          />
        ))}
      </div>
    </div>
  )
}
