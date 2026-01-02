'use client'

import { useMemo } from 'react'
import {
  Clock,
  HardDrive,
  AlertCircle,
  Loader2,
  RotateCcw,
  Trash2,
  GitBranch,
  GitCommit,
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
  if (diffInSeconds < 3600) return `${Math.floor(diffInSeconds / 60)} min ago`
  if (diffInSeconds < 86400) return `${Math.floor(diffInSeconds / 3600)} hours ago`
  if (diffInSeconds < 604800) return `${Math.floor(diffInSeconds / 86400)} days ago`
  return date.toLocaleDateString()
}

// Extract short hash from snapshot name (last segment after the last dash, typically timestamp)
function getShortHash(name: string): string {
  const parts = name.split('-')
  if (parts.length >= 2) {
    // Return last two parts joined (e.g., "20260102-084537" -> "084537")
    return parts.slice(-1)[0]
  }
  return name.slice(-8)
}

function getStateBadge(state: Snapshot['status']['state']) {
  const baseClasses = 'inline-flex items-center gap-1 rounded-full px-2 py-0.5 text-xs font-medium'

  switch (state) {
    case 'Ready':
      return (
        <span className={`${baseClasses} bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400`}>
          Ready
        </span>
      )
    case 'Creating':
      return (
        <span className={`${baseClasses} bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-400`}>
          <Loader2 className="h-3 w-3 animate-spin" />
          Creating
        </span>
      )
    case 'Restoring':
      return (
        <span className={`${baseClasses} bg-yellow-100 text-yellow-700 dark:bg-yellow-900/30 dark:text-yellow-400`}>
          <Loader2 className="h-3 w-3 animate-spin" />
          Restoring
        </span>
      )
    case 'Deleting':
      return (
        <span className={`${baseClasses} bg-orange-100 text-orange-700 dark:bg-orange-900/30 dark:text-orange-400`}>
          <Loader2 className="h-3 w-3 animate-spin" />
          Deleting
        </span>
      )
    case 'Failed':
      return (
        <span className={`${baseClasses} bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400`}>
          <AlertCircle className="h-3 w-3" />
          Failed
        </span>
      )
    case 'Pending':
    default:
      return (
        <span className={`${baseClasses} bg-secondary text-secondary-foreground`}>
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

  // HEAD is determined by backend via lastRestoredSnapshot
  // If not set, default to newest snapshot
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

  return (
    <div className="relative flex gap-4">
      {/* Timeline track */}
      <div className="relative flex flex-col items-center" style={{ width: 20 }}>
        {/* Line above dot */}
        {!isFirst && (
          <div className="w-0.5 flex-1 bg-border" />
        )}
        {isFirst && <div className="flex-1" />}

        {/* Dot */}
        <div
          className={cn(
            "relative z-10 rounded-full flex-shrink-0",
            isCurrent
              ? "w-4 h-4 bg-blue-500 ring-4 ring-blue-100 dark:ring-blue-900/50"
              : "w-2.5 h-2.5 bg-gray-400 dark:bg-gray-500"
          )}
        />

        {/* Line below dot */}
        {!isLast && (
          <div className="w-0.5 flex-1 bg-border" />
        )}
        {isLast && <div className="flex-1" />}
      </div>

      {/* Content */}
      <div className="flex-1 pb-4 min-w-0">
        <div className={cn(
          "rounded-lg border p-4 transition-all",
          isCurrent
            ? "bg-blue-50 border-blue-200 shadow-sm dark:bg-blue-950/30 dark:border-blue-800"
            : "bg-card hover:bg-muted/50 hover:shadow-sm"
        )}>
          <div className="flex items-start justify-between gap-3">
            <div className="min-w-0 flex-1">
              <div className="flex items-center gap-2 flex-wrap mb-2">
                {isCurrent && (
                  <span className="inline-flex items-center gap-1 rounded-md bg-blue-500 px-2 py-0.5 text-xs font-medium text-white">
                    HEAD
                  </span>
                )}
                <span className="inline-flex items-center gap-1 font-mono text-xs text-muted-foreground bg-muted px-1.5 py-0.5 rounded">
                  <GitCommit className="h-3 w-3" />
                  {shortHash}
                </span>
                {getStateBadge(snapshot.status.state)}
              </div>

              {snapshot.spec.description && (
                <p className="text-sm text-foreground">
                  {snapshot.spec.description}
                </p>
              )}

              <div className="flex items-center gap-4 mt-2 text-xs text-muted-foreground flex-wrap">
                <span className="flex items-center gap-1">
                  <Clock className="h-3 w-3" />
                  {formatTimeAgo(snapshot.status.createdAt || snapshot.metadata.creationTimestamp)}
                </span>
                {snapshot.status.sizeHuman && (
                  <span className="flex items-center gap-1">
                    <HardDrive className="h-3 w-3" />
                    {snapshot.status.sizeHuman}
                  </span>
                )}
                {snapshot.spec.parentSnapshotRef && (
                  <span className="flex items-center gap-1 text-muted-foreground">
                    <GitBranch className="h-3 w-3" />
                    parent: {getShortHash(snapshot.spec.parentSnapshotRef.name)}
                  </span>
                )}
              </div>

              {snapshot.status.state === 'Failed' && snapshot.status.message && (
                <p className="mt-2 text-xs text-red-600 dark:text-red-400 bg-red-50 dark:bg-red-950/30 rounded px-2 py-1">
                  {snapshot.status.message}
                </p>
              )}
            </div>

            <div className="flex items-center gap-2 flex-shrink-0">
              {snapshot.status.state === 'Ready' && (
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => onRestore(snapshot)}
                  disabled={disabled}
                  className="h-8"
                >
                  <RotateCcw className="h-3 w-3 mr-1.5" />
                  Restore
                </Button>
              )}
              {(snapshot.status.state === 'Ready' || snapshot.status.state === 'Failed') && (
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() => onDelete(snapshot)}
                  className="h-8 w-8 p-0 text-muted-foreground hover:text-destructive"
                >
                  <Trash2 className="h-4 w-4" />
                </Button>
              )}
            </div>
          </div>
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
      <h4 className="text-sm font-medium flex items-center gap-2 mb-4">
        <GitBranch className="h-4 w-4" />
        Snapshot History
        <span className="text-muted-foreground font-normal">({snapshots.length})</span>
      </h4>

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
