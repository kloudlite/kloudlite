'use client'

import { useMemo } from 'react'
import {
  Clock,
  HardDrive,
  AlertCircle,
  Loader2,
  RotateCcw,
  Trash2,
} from 'lucide-react'
import { Button, Badge } from '@kloudlite/ui'
import { cn } from '@/lib/utils'
import type { Snapshot } from '@/lib/services/snapshot.service'

interface SnapshotTimelineProps {
  snapshots: Snapshot[]
  onRestore: (snapshot: Snapshot) => void
  onDelete: (snapshot: Snapshot) => void
  disabled?: boolean
  currentSnapshotName?: string
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
      return null
    case 'Creating':
      return (
        <span className="inline-flex items-center gap-1 text-xs text-blue-600 dark:text-blue-400">
          <Loader2 className="h-3 w-3 animate-spin" />
          Creating
        </span>
      )
    case 'Restoring':
      return (
        <span className="inline-flex items-center gap-1 text-xs text-amber-600 dark:text-amber-400">
          <Loader2 className="h-3 w-3 animate-spin" />
          Restoring
        </span>
      )
    case 'Deleting':
      return (
        <span className="inline-flex items-center gap-1 text-xs text-orange-600 dark:text-orange-400">
          <Loader2 className="h-3 w-3 animate-spin" />
          Deleting
        </span>
      )
    case 'Failed':
      return (
        <span className="inline-flex items-center gap-1 text-xs text-red-600 dark:text-red-400">
          <AlertCircle className="h-3 w-3" />
          Failed
        </span>
      )
    case 'Pending':
    default:
      return (
        <span className="text-xs text-muted-foreground">Pending</span>
      )
  }
}

interface SnapshotItemProps {
  snapshot: Snapshot
  isCurrent: boolean
  onRestore: (snapshot: Snapshot) => void
  onDelete: (snapshot: Snapshot) => void
  disabled?: boolean
  isLast: boolean
}

function SnapshotItem({ snapshot, isCurrent, onRestore, onDelete, disabled, isLast }: SnapshotItemProps) {
  const shortHash = getShortHash(snapshot.metadata.name)

  return (
    <div className="flex">
      {/* Timeline line and dot */}
      <div className="flex flex-col items-center mr-3">
        <div className={cn(
          "w-2 h-2 rounded-full mt-4 flex-shrink-0",
          isCurrent
            ? "bg-blue-500 ring-2 ring-blue-200 dark:ring-blue-800"
            : "bg-gray-400 dark:bg-gray-500"
        )} />
        {!isLast && (
          <div className="w-0.5 flex-1 bg-gray-200 dark:bg-gray-700 mt-1" />
        )}
      </div>

      {/* Content */}
      <div
        className={cn(
          "group flex-1 py-2.5 px-3 rounded-lg border transition-colors mb-2",
          isCurrent
            ? "bg-blue-50 border-blue-200 dark:bg-blue-950/30 dark:border-blue-800"
            : "border-border hover:bg-muted/50"
        )}
      >
        <div className="flex items-center justify-between gap-3">
          <div className="flex items-center gap-2 min-w-0">
            {/* Hash */}
            <code className={cn(
              "text-sm font-mono",
              isCurrent ? "text-blue-600 dark:text-blue-400 font-semibold" : "text-foreground"
            )}>
              {shortHash}
            </code>

            {/* Current badge */}
            {isCurrent && (
              <Badge variant="secondary" className="text-[10px] px-1.5 py-0 h-4 bg-blue-100 text-blue-700 dark:bg-blue-900 dark:text-blue-300">
                HEAD
              </Badge>
            )}

            {/* State badge */}
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
                className="h-6 px-2 text-xs"
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
                className="h-6 w-6 p-0 text-muted-foreground hover:text-destructive"
              >
                <Trash2 className="h-3 w-3" />
              </Button>
            )}
          </div>
        </div>

        {/* Description */}
        {snapshot.spec.description && (
          <p className="text-sm text-foreground mt-1">
            {snapshot.spec.description}
          </p>
        )}

        {/* Meta */}
        <div className="flex items-center gap-3 mt-1 text-xs text-muted-foreground">
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
        </div>

        {/* Error */}
        {snapshot.status.state === 'Failed' && snapshot.status.message && (
          <p className="mt-1.5 text-xs text-red-600 dark:text-red-400">
            {snapshot.status.message}
          </p>
        )}
      </div>
    </div>
  )
}

export function SnapshotTimeline({ snapshots, onRestore, onDelete, disabled, currentSnapshotName }: SnapshotTimelineProps) {
  // Sort snapshots by creation time (newest first) - simple flat list like git log
  const sortedSnapshots = useMemo(() => {
    return [...snapshots].sort((a, b) => {
      const aTime = new Date(a.status.createdAt || a.metadata.creationTimestamp).getTime()
      const bTime = new Date(b.status.createdAt || b.metadata.creationTimestamp).getTime()
      return bTime - aTime
    })
  }, [snapshots])

  if (snapshots.length === 0) {
    return null
  }

  return (
    <div>
      <div className="flex items-center justify-between mb-3">
        <span className="text-xs font-medium text-muted-foreground uppercase tracking-wider">
          Snapshots
        </span>
        <span className="text-xs text-muted-foreground">{snapshots.length}</span>
      </div>

      <div>
        {sortedSnapshots.map((snapshot, idx) => (
          <SnapshotItem
            key={snapshot.metadata.name}
            snapshot={snapshot}
            isCurrent={snapshot.metadata.name === currentSnapshotName}
            onRestore={onRestore}
            onDelete={onDelete}
            disabled={disabled}
            isLast={idx === sortedSnapshots.length - 1}
          />
        ))}
      </div>
    </div>
  )
}
