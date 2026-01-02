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
  CircleDot,
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

interface GraphNode {
  snapshot: Snapshot
  column: number
  hasParent: boolean
  isBranchStart: boolean
  isCurrent: boolean
  connectors: {
    fromColumn: number
    toColumn: number
  }[]
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

const BRANCH_COLORS = [
  'bg-blue-500',
  'bg-green-500',
  'bg-purple-500',
  'bg-orange-500',
  'bg-pink-500',
]

/**
 * Build graph with column assignments for branching visualization
 */
function buildGraph(snapshots: Snapshot[], currentSnapshotName?: string): { nodes: GraphNode[], maxColumn: number } {
  if (snapshots.length === 0) return { nodes: [], maxColumn: 0 }

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

  // Find roots and sort by creation time (oldest first)
  const roots = snapshots
    .filter(s => !hasParentInList.has(s.metadata.name))
    .sort((a, b) =>
      new Date(a.status.createdAt || a.metadata.creationTimestamp).getTime() -
      new Date(b.status.createdAt || b.metadata.creationTimestamp).getTime()
    )

  // Find the most recent snapshot to mark as current
  const sortedByTime = [...snapshots].sort((a, b) =>
    new Date(b.status.createdAt || b.metadata.creationTimestamp).getTime() -
    new Date(a.status.createdAt || a.metadata.creationTimestamp).getTime()
  )
  const mostRecentName = currentSnapshotName || sortedByTime[0]?.metadata.name

  // Build graph with column assignments
  const result: GraphNode[] = []
  let maxColumn = 0
  const columnStack: number[] = [0] // Available columns

  function getNextColumn(): number {
    if (columnStack.length > 0) {
      return columnStack.shift()!
    }
    return ++maxColumn
  }

  function releaseColumn(col: number) {
    if (!columnStack.includes(col)) {
      columnStack.push(col)
      columnStack.sort((a, b) => a - b)
    }
  }

  // Track active columns for each snapshot
  const snapshotColumns = new Map<string, number>()

  function traverse(snapshot: Snapshot, column: number, parentColumn: number | null) {
    snapshotColumns.set(snapshot.metadata.name, column)

    const children = childrenMap.get(snapshot.metadata.name) || []
    children.sort((a, b) =>
      new Date(a.status.createdAt || a.metadata.creationTimestamp).getTime() -
      new Date(b.status.createdAt || b.metadata.creationTimestamp).getTime()
    )

    const isBranchStart = parentColumn !== null && parentColumn !== column
    const connectors: GraphNode['connectors'] = []

    if (parentColumn !== null && parentColumn !== column) {
      connectors.push({ fromColumn: parentColumn, toColumn: column })
    }

    result.push({
      snapshot,
      column,
      hasParent: hasParentInList.has(snapshot.metadata.name),
      isBranchStart,
      isCurrent: snapshot.metadata.name === mostRecentName,
      connectors,
    })

    // Process children - first child continues on same column, others branch
    children.forEach((child, idx) => {
      if (idx === 0) {
        traverse(child, column, column)
      } else {
        const newCol = getNextColumn()
        if (newCol > maxColumn) maxColumn = newCol
        traverse(child, newCol, column)
      }
    })

    // If no children, release this column
    if (children.length === 0) {
      releaseColumn(column)
    }
  }

  roots.forEach((root, idx) => {
    const col = idx === 0 ? 0 : getNextColumn()
    if (col > maxColumn) maxColumn = col
    traverse(root, col, null)
  })

  // Reverse to show newest at top
  return { nodes: result.reverse(), maxColumn }
}

interface TimelineItemProps {
  node: GraphNode
  maxColumn: number
  nextNode: GraphNode | null
  onRestore: (snapshot: Snapshot) => void
  onDelete: (snapshot: Snapshot) => void
  disabled?: boolean
}

function TimelineItem({ node, maxColumn, nextNode, onRestore, onDelete, disabled }: TimelineItemProps) {
  const { snapshot, column, hasParent, isCurrent, connectors } = node
  const columnWidth = 24
  const graphWidth = (maxColumn + 1) * columnWidth + 8

  return (
    <div className="flex">
      {/* Graph area */}
      <div
        className="relative flex-shrink-0"
        style={{ width: graphWidth, minHeight: '80px' }}
      >
        {/* Vertical lines for all columns */}
        {Array.from({ length: maxColumn + 1 }).map((_, col) => {
          // Show line if this column has activity
          const showLine = col === column ||
            (nextNode && nextNode.column === col) ||
            connectors.some(c => c.fromColumn === col || c.toColumn === col)

          if (!showLine) return null

          const isMainLine = col === column

          return (
            <div
              key={col}
              className={cn(
                "absolute top-0 bottom-0 w-0.5",
                isMainLine ? BRANCH_COLORS[col % BRANCH_COLORS.length] : "bg-gray-300 dark:bg-gray-600"
              )}
              style={{ left: col * columnWidth + 7 }}
            />
          )
        })}

        {/* Branch connector (horizontal line) */}
        {connectors.map((conn, idx) => (
          <div
            key={idx}
            className="absolute h-0.5 bg-gray-400 dark:bg-gray-500"
            style={{
              left: Math.min(conn.fromColumn, conn.toColumn) * columnWidth + 8,
              width: Math.abs(conn.toColumn - conn.fromColumn) * columnWidth,
              top: '40px',
            }}
          />
        ))}

        {/* Node dot */}
        <div
          className={cn(
            "absolute rounded-full flex items-center justify-center",
            isCurrent ? "w-5 h-5 ring-2 ring-yellow-400 ring-offset-2 ring-offset-background" : "w-3 h-3",
            BRANCH_COLORS[column % BRANCH_COLORS.length]
          )}
          style={{
            left: column * columnWidth + (isCurrent ? 4 : 6),
            top: isCurrent ? '34px' : '38px'
          }}
        >
          {isCurrent && (
            <div className="w-2 h-2 rounded-full bg-yellow-300" />
          )}
        </div>
      </div>

      {/* Card content */}
      <div className="flex-1 pb-3">
        <div className={cn(
          "rounded-lg border p-4 transition-colors",
          isCurrent
            ? "bg-yellow-50 border-yellow-300 dark:bg-yellow-900/20 dark:border-yellow-700"
            : "bg-card hover:bg-muted/30"
        )}>
          <div className="flex items-start justify-between gap-4">
            <div className="min-w-0 flex-1">
              <div className="flex items-center gap-2 flex-wrap">
                {isCurrent && (
                  <span className="inline-flex items-center gap-1 rounded-full bg-yellow-200 px-2 py-0.5 text-xs font-medium text-yellow-800 dark:bg-yellow-800 dark:text-yellow-200">
                    <CircleDot className="h-3 w-3" />
                    Current
                  </span>
                )}
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

export function SnapshotTimeline({ snapshots, onRestore, onDelete, disabled, currentSnapshotName }: SnapshotTimelineProps) {
  const { nodes, maxColumn } = useMemo(() => buildGraph(snapshots, currentSnapshotName), [snapshots, currentSnapshotName])

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
            maxColumn={maxColumn}
            nextNode={idx < nodes.length - 1 ? nodes[idx + 1] : null}
            onRestore={onRestore}
            onDelete={onDelete}
            disabled={disabled}
          />
        ))}
      </div>
    </div>
  )
}
