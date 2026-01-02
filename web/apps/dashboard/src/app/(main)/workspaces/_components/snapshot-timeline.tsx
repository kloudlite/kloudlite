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

interface SnapshotNode {
  snapshot: Snapshot
  hasParent: boolean
  depth: number
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

/**
 * Build a flat list with depth information for rendering
 */
function buildSnapshotList(snapshots: Snapshot[]): SnapshotNode[] {
  if (snapshots.length === 0) return []

  // Create maps for lookup
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

  // Find roots
  const roots = snapshots.filter(s => !hasParentInList.has(s.metadata.name))

  // Sort roots by creation time (oldest first for bottom-up display)
  roots.sort((a, b) =>
    new Date(a.status.createdAt || a.metadata.creationTimestamp).getTime() -
    new Date(b.status.createdAt || b.metadata.creationTimestamp).getTime()
  )

  // Build flat list with DFS
  const result: SnapshotNode[] = []

  function traverse(snapshot: Snapshot, depth: number) {
    const children = childrenMap.get(snapshot.metadata.name) || []
    // Sort children by creation time (oldest first)
    children.sort((a, b) =>
      new Date(a.status.createdAt || a.metadata.creationTimestamp).getTime() -
      new Date(b.status.createdAt || b.metadata.creationTimestamp).getTime()
    )

    result.push({
      snapshot,
      hasParent: hasParentInList.has(snapshot.metadata.name),
      depth,
    })

    children.forEach(child => traverse(child, depth + 1))
  }

  roots.forEach(root => traverse(root, 0))

  // Reverse to show newest at top
  return result.reverse()
}

interface TimelineItemProps {
  node: SnapshotNode
  isLast: boolean
  onRestore: (snapshot: Snapshot) => void
  onDelete: (snapshot: Snapshot) => void
  disabled?: boolean
}

function TimelineItem({ node, isLast, onRestore, onDelete, disabled }: TimelineItemProps) {
  const { snapshot, hasParent } = node

  return (
    <div className="flex gap-3">
      {/* Left side - line and dot */}
      <div className="flex flex-col items-center" style={{ width: '16px' }}>
        {/* Line segment above dot */}
        <div
          className={cn(
            "w-0.5 flex-1",
            hasParent ? "bg-gray-300 dark:bg-gray-600" : "bg-transparent"
          )}
          style={{ minHeight: '8px' }}
        />
        {/* Dot */}
        <div
          className={cn(
            "w-3 h-3 rounded-full flex-shrink-0",
            hasParent
              ? "bg-blue-500"
              : "bg-gray-400 dark:bg-gray-500"
          )}
        />
        {/* Line segment below dot */}
        <div
          className={cn(
            "w-0.5 flex-1",
            !isLast ? "bg-gray-300 dark:bg-gray-600" : "bg-transparent"
          )}
          style={{ minHeight: '8px' }}
        />
      </div>

      {/* Right side - card content */}
      <div className="flex-1 pb-3">
        <div className="bg-card rounded-lg border p-4 hover:bg-muted/30 transition-colors">
          <div className="flex items-start justify-between gap-4">
            <div className="min-w-0 flex-1">
              <div className="flex items-center gap-2 flex-wrap">
                <span className="truncate font-mono text-sm">
                  {snapshot.metadata.name}
                </span>
                {getStateBadge(snapshot.status.state)}
              </div>

              <div className="text-muted-foreground mt-2 flex items-center gap-3 text-xs flex-wrap">
                {snapshot.status.sizeHuman && (
                  <span className="flex items-center gap-1">
                    <HardDrive className="h-3 w-3" />
                    {snapshot.status.sizeHuman}
                  </span>
                )}
                <span className="flex items-center gap-1">
                  <Clock className="h-3 w-3" />
                  {formatTimeAgo(snapshot.status.createdAt || snapshot.metadata.creationTimestamp)}
                </span>
                {hasParent && snapshot.spec.parentSnapshotRef && (
                  <span className="flex items-center gap-1 text-blue-500">
                    <GitBranch className="h-3 w-3" />
                    from {snapshot.spec.parentSnapshotRef.name.split('-').slice(-2).join('-')}
                  </span>
                )}
              </div>

              {snapshot.spec.description && (
                <p className="text-muted-foreground mt-2 text-sm italic">
                  &quot;{snapshot.spec.description}&quot;
                </p>
              )}

              {snapshot.status.state === 'Failed' && snapshot.status.message && (
                <p className="mt-2 text-xs text-red-600 dark:text-red-400">
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
                  title={disabled ? 'Resource must be running to restore' : undefined}
                >
                  <RotateCcw className="mr-1 h-3 w-3" />
                  Restore
                </Button>
              )}
              {(snapshot.status.state === 'Ready' || snapshot.status.state === 'Failed') && (
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() => onDelete(snapshot)}
                  className="text-destructive hover:text-destructive"
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
  const nodes = useMemo(() => buildSnapshotList(snapshots), [snapshots])

  if (snapshots.length === 0) {
    return null
  }

  return (
    <div className="space-y-2">
      <h4 className="text-sm font-medium flex items-center gap-2">
        <GitBranch className="h-4 w-4" />
        Snapshot History ({snapshots.length})
      </h4>

      <div>
        {nodes.map((node, idx) => (
          <TimelineItem
            key={node.snapshot.metadata.name}
            node={node}
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
