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
  GitCommitHorizontal,
} from 'lucide-react'
import { Button } from '@kloudlite/ui'
import { cn } from '@/lib/utils'
import type { Snapshot } from '@/lib/services/snapshot.service'

interface SnapshotTimelineProps {
  snapshots: Snapshot[]
  onRestore: (snapshot: Snapshot) => void
  onDelete: (snapshot: Snapshot) => void
  disabled?: boolean
}

interface TimelineNode {
  snapshot: Snapshot
  hasParent: boolean
  isCurrent: boolean
  siblingCount: number // Number of siblings (other children of same parent)
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

function buildTimeline(snapshots: Snapshot[]): TimelineNode[] {
  if (snapshots.length === 0) return []

  // Create maps
  const snapshotMap = new Map<string, Snapshot>()
  const childrenMap = new Map<string, Snapshot[]>()
  const hasParentInList = new Set<string>()

  snapshots.forEach(s => snapshotMap.set(s.metadata.name, s))

  snapshots.forEach(snapshot => {
    const parentName = snapshot.spec.parentSnapshotRef?.name
    if (parentName && snapshotMap.has(parentName)) {
      hasParentInList.add(snapshot.metadata.name)
      const children = childrenMap.get(parentName) || []
      children.push(snapshot)
      childrenMap.set(parentName, children)
    }
  })

  // Find most recent snapshot
  const sortedByTime = [...snapshots].sort((a, b) =>
    new Date(b.status.createdAt || b.metadata.creationTimestamp).getTime() -
    new Date(a.status.createdAt || a.metadata.creationTimestamp).getTime()
  )
  const mostRecentName = sortedByTime[0]?.metadata.name

  // Build timeline - sort by creation time, newest first
  return sortedByTime.map(snapshot => {
    const parentName = snapshot.spec.parentSnapshotRef?.name
    const siblings = parentName ? (childrenMap.get(parentName) || []) : []

    return {
      snapshot,
      hasParent: hasParentInList.has(snapshot.metadata.name),
      isCurrent: snapshot.metadata.name === mostRecentName,
      siblingCount: siblings.length - 1, // -1 because we don't count self
    }
  })
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
  const { snapshot, hasParent, isCurrent, siblingCount } = node
  const hasSiblings = siblingCount > 0

  return (
    <div className="relative flex gap-4">
      {/* Timeline track */}
      <div className="flex flex-col items-center w-6 flex-shrink-0">
        {/* Line above */}
        <div className={cn(
          "w-0.5 flex-1 min-h-3",
          isFirst ? "bg-transparent" : "bg-gray-200 dark:bg-gray-700"
        )} />

        {/* Dot */}
        <div className={cn(
          "relative flex-shrink-0 rounded-full flex items-center justify-center",
          isCurrent
            ? "w-5 h-5 bg-blue-500 ring-4 ring-blue-100 dark:ring-blue-900"
            : hasSiblings
              ? "w-4 h-4 bg-orange-400 ring-2 ring-orange-100 dark:ring-orange-900"
              : "w-3 h-3 bg-gray-400 dark:bg-gray-500"
        )}>
          {isCurrent && (
            <GitCommitHorizontal className="w-3 h-3 text-white" />
          )}
        </div>

        {/* Line below */}
        <div className={cn(
          "w-0.5 flex-1 min-h-3",
          isLast ? "bg-transparent" : "bg-gray-200 dark:bg-gray-700"
        )} />
      </div>

      {/* Content */}
      <div className="flex-1 pb-4 min-w-0">
        <div className={cn(
          "rounded-lg border p-4 transition-all",
          isCurrent
            ? "bg-blue-50 border-blue-200 shadow-sm dark:bg-blue-950/30 dark:border-blue-800"
            : "bg-card hover:bg-muted/50 hover:shadow-sm"
        )}>
          {/* Header */}
          <div className="flex items-start justify-between gap-3">
            <div className="min-w-0 flex-1">
              {/* Tags row */}
              <div className="flex items-center gap-2 flex-wrap mb-2">
                {isCurrent && (
                  <span className="inline-flex items-center gap-1 rounded-md bg-blue-500 px-2 py-0.5 text-xs font-medium text-white">
                    HEAD
                  </span>
                )}
                {hasSiblings && (
                  <span className="inline-flex items-center gap-1 rounded-md bg-orange-100 px-2 py-0.5 text-xs font-medium text-orange-700 dark:bg-orange-900/30 dark:text-orange-400">
                    <GitBranch className="h-3 w-3" />
                    Branch
                  </span>
                )}
                {getStateBadge(snapshot.status.state)}
              </div>

              {/* Name */}
              <p className="font-mono text-sm text-foreground truncate">
                {snapshot.metadata.name}
              </p>

              {/* Description */}
              {snapshot.spec.description && (
                <p className="text-sm text-muted-foreground mt-1 italic">
                  &quot;{snapshot.spec.description}&quot;
                </p>
              )}

              {/* Metadata row */}
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
                {hasParent && snapshot.spec.parentSnapshotRef && (
                  <span className="flex items-center gap-1 text-blue-600 dark:text-blue-400">
                    <GitBranch className="h-3 w-3" />
                    from {snapshot.spec.parentSnapshotRef.name.split('-').slice(-2).join('-')}
                  </span>
                )}
              </div>

              {/* Error message */}
              {snapshot.status.state === 'Failed' && snapshot.status.message && (
                <p className="mt-2 text-xs text-red-600 dark:text-red-400 bg-red-50 dark:bg-red-950/30 rounded px-2 py-1">
                  {snapshot.status.message}
                </p>
              )}
            </div>

            {/* Actions */}
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

export function SnapshotTimeline({ snapshots, onRestore, onDelete, disabled }: SnapshotTimelineProps) {
  const nodes = useMemo(() => buildTimeline(snapshots), [snapshots])

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

      <div className="relative">
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
