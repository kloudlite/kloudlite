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
  column: number
  parentColumn: number | null
  isCurrent: boolean
  isBranchPoint: boolean // Has multiple children
  hasSiblings: boolean // Has siblings (shares parent with others)
}

// Colors for different columns/branches
const BRANCH_COLORS = [
  { line: 'bg-blue-500', dot: 'bg-blue-500', ring: 'ring-blue-200 dark:ring-blue-900' },
  { line: 'bg-purple-500', dot: 'bg-purple-500', ring: 'ring-purple-200 dark:ring-purple-900' },
  { line: 'bg-green-500', dot: 'bg-green-500', ring: 'ring-green-200 dark:ring-green-900' },
  { line: 'bg-orange-500', dot: 'bg-orange-500', ring: 'ring-orange-200 dark:ring-orange-900' },
  { line: 'bg-pink-500', dot: 'bg-pink-500', ring: 'ring-pink-200 dark:ring-pink-900' },
]

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

function buildTimeline(snapshots: Snapshot[]): { nodes: TimelineNode[], maxColumn: number } {
  if (snapshots.length === 0) return { nodes: [], maxColumn: 0 }

  // Create maps for relationships
  const snapshotMap = new Map<string, Snapshot>()
  const childrenMap = new Map<string, string[]>() // parent -> children names

  snapshots.forEach(s => snapshotMap.set(s.metadata.name, s))

  // Build parent -> children relationships
  snapshots.forEach(snapshot => {
    const parentName = snapshot.spec.parentSnapshotRef?.name
    if (parentName && snapshotMap.has(parentName)) {
      const children = childrenMap.get(parentName) || []
      children.push(snapshot.metadata.name)
      childrenMap.set(parentName, children)
    }
  })

  // Sort by creation time, newest first (this is display order)
  const sortedByTime = [...snapshots].sort((a, b) =>
    new Date(b.status.createdAt || b.metadata.creationTimestamp).getTime() -
    new Date(a.status.createdAt || a.metadata.creationTimestamp).getTime()
  )

  const mostRecentName = sortedByTime[0]?.metadata.name

  // Assign columns: each unique lineage path gets a column
  // We trace from each snapshot to root and assign columns based on branch points
  const columnAssignments = new Map<string, number>()
  let nextColumn = 0

  // Process in reverse chronological order (oldest first for column assignment)
  const oldestFirst = [...sortedByTime].reverse()

  oldestFirst.forEach(snapshot => {
    const name = snapshot.metadata.name
    const parentName = snapshot.spec.parentSnapshotRef?.name

    if (!parentName || !snapshotMap.has(parentName)) {
      // Root snapshot - assign new column
      columnAssignments.set(name, nextColumn++)
    } else {
      // Has parent in list
      const siblings = childrenMap.get(parentName) || []
      const parentColumn = columnAssignments.get(parentName) ?? 0

      if (siblings.length === 1) {
        // Only child - inherit parent's column
        columnAssignments.set(name, parentColumn)
      } else {
        // Multiple children - first child inherits, others get new columns
        const siblingIndex = siblings.indexOf(name)
        if (siblingIndex === 0) {
          columnAssignments.set(name, parentColumn)
        } else {
          columnAssignments.set(name, nextColumn++)
        }
      }
    }
  })

  // Build timeline nodes
  const nodes: TimelineNode[] = sortedByTime.map(snapshot => {
    const name = snapshot.metadata.name
    const parentName = snapshot.spec.parentSnapshotRef?.name
    const children = childrenMap.get(name) || []
    const parentColumn = parentName && snapshotMap.has(parentName)
      ? columnAssignments.get(parentName) ?? null
      : null

    // Check if this snapshot has siblings
    const siblings = parentName ? (childrenMap.get(parentName) || []) : []

    return {
      snapshot,
      column: columnAssignments.get(name) ?? 0,
      parentColumn,
      isCurrent: name === mostRecentName,
      isBranchPoint: children.length > 1,
      hasSiblings: siblings.length > 1,
    }
  })

  return { nodes, maxColumn: Math.max(0, nextColumn - 1) }
}

interface TimelineItemProps {
  node: TimelineNode
  index: number
  totalNodes: number
  maxColumn: number
  allNodes: TimelineNode[]
  onRestore: (snapshot: Snapshot) => void
  onDelete: (snapshot: Snapshot) => void
  disabled?: boolean
}

function TimelineItem({
  node,
  index,
  totalNodes,
  maxColumn,
  allNodes,
  onRestore,
  onDelete,
  disabled
}: TimelineItemProps) {
  const { snapshot, column, parentColumn, isCurrent, isBranchPoint, hasSiblings } = node
  const isFirst = index === 0
  const isLast = index === totalNodes - 1
  const colors = BRANCH_COLORS[column % BRANCH_COLORS.length]

  // Calculate which columns need vertical lines at this row
  // A column needs a line if there's a snapshot below that uses this column
  const activeColumns = new Set<number>()
  for (let i = index; i < totalNodes; i++) {
    activeColumns.add(allNodes[i].column)
    // Also add parent columns for nodes that branch
    if (allNodes[i].parentColumn !== null && allNodes[i].parentColumn !== allNodes[i].column) {
      // Find if parent is below current index
      const parentIdx = allNodes.findIndex(n => n.snapshot.metadata.name === allNodes[i].snapshot.spec.parentSnapshotRef?.name)
      if (parentIdx > index) {
        activeColumns.add(allNodes[i].parentColumn)
      }
    }
  }

  // Width for the graph area (each column is 24px wide)
  const graphWidth = (maxColumn + 1) * 24
  const dotCenterX = column * 24 + 12

  return (
    <div className="relative flex">
      {/* Graph area with columns */}
      <div
        className="relative flex-shrink-0"
        style={{ width: Math.max(graphWidth, 24), minHeight: '100%' }}
      >
        {/* Vertical lines for all active columns */}
        {Array.from(activeColumns).map(col => {
          const colColors = BRANCH_COLORS[col % BRANCH_COLORS.length]
          const colX = col * 24 + 12
          const isCurrentColumn = col === column

          return (
            <div
              key={col}
              className={cn("absolute w-0.5", colColors.line)}
              style={{
                left: colX - 1,
                top: isCurrentColumn && isFirst ? '50%' : 0,
                bottom: isCurrentColumn && isLast ? '50%' : 0,
              }}
            />
          )
        })}

        {/* Horizontal connector line from parent column to this column */}
        {parentColumn !== null && parentColumn !== column && (
          <div
            className={cn("absolute h-0.5", colors.line)}
            style={{
              left: Math.min(parentColumn, column) * 24 + 12,
              width: Math.abs(parentColumn - column) * 24,
              top: '50%',
              transform: 'translateY(-50%)',
            }}
          />
        )}

        {/* Dot for this snapshot */}
        <div
          className={cn(
            "absolute rounded-full flex items-center justify-center transform -translate-x-1/2 -translate-y-1/2",
            isCurrent
              ? `w-6 h-6 ${colors.dot} ring-4 ${colors.ring}`
              : isBranchPoint
                ? `w-5 h-5 ${colors.dot} ring-2 ${colors.ring}`
                : `w-3 h-3 ${colors.dot}`
          )}
          style={{
            left: dotCenterX,
            top: '50%',
          }}
        >
          {isCurrent && (
            <GitCommitHorizontal className="w-3 h-3 text-white" />
          )}
        </div>
      </div>

      {/* Content */}
      <div className="flex-1 py-2 min-w-0">
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
                  <span className={cn(
                    "inline-flex items-center gap-1 rounded-md px-2 py-0.5 text-xs font-medium",
                    "bg-orange-100 text-orange-700 dark:bg-orange-900/30 dark:text-orange-400"
                  )}>
                    <GitBranch className="h-3 w-3" />
                    Branch
                  </span>
                )}
                {isBranchPoint && (
                  <span className="inline-flex items-center gap-1 rounded-md bg-purple-100 px-2 py-0.5 text-xs font-medium text-purple-700 dark:bg-purple-900/30 dark:text-purple-400">
                    <GitBranch className="h-3 w-3" />
                    Fork
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
                {snapshot.spec.parentSnapshotRef && (
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
  const { nodes, maxColumn } = useMemo(() => buildTimeline(snapshots), [snapshots])

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
            index={idx}
            totalNodes={nodes.length}
            maxColumn={maxColumn}
            allNodes={nodes}
            onRestore={onRestore}
            onDelete={onDelete}
            disabled={disabled}
          />
        ))}
      </div>
    </div>
  )
}
