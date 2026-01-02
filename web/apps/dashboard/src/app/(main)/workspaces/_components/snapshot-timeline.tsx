'use client'

import { useMemo, type ReactElement } from 'react'
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

interface TimelineNode {
  snapshot: Snapshot
  column: number
  parentColumn: number | null
  isCurrent: boolean
  isBranchPoint: boolean
  hasSiblings: boolean
}

const BRANCH_COLORS = [
  { line: '#3b82f6', dot: 'bg-blue-500', ring: 'ring-blue-200 dark:ring-blue-800' },
  { line: '#a855f7', dot: 'bg-purple-500', ring: 'ring-purple-200 dark:ring-purple-800' },
  { line: '#22c55e', dot: 'bg-green-500', ring: 'ring-green-200 dark:ring-green-800' },
  { line: '#f97316', dot: 'bg-orange-500', ring: 'ring-orange-200 dark:ring-orange-800' },
  { line: '#ec4899', dot: 'bg-pink-500', ring: 'ring-pink-200 dark:ring-pink-800' },
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

  const snapshotMap = new Map<string, Snapshot>()
  const childrenMap = new Map<string, string[]>()

  snapshots.forEach(s => snapshotMap.set(s.metadata.name, s))

  snapshots.forEach(snapshot => {
    const parentName = snapshot.spec.parentSnapshotRef?.name
    if (parentName && snapshotMap.has(parentName)) {
      const children = childrenMap.get(parentName) || []
      children.push(snapshot.metadata.name)
      childrenMap.set(parentName, children)
    }
  })

  const sortedByTime = [...snapshots].sort((a, b) =>
    new Date(b.status.createdAt || b.metadata.creationTimestamp).getTime() -
    new Date(a.status.createdAt || a.metadata.creationTimestamp).getTime()
  )

  const mostRecentName = sortedByTime[0]?.metadata.name

  const columnAssignments = new Map<string, number>()
  let nextColumn = 0

  const oldestFirst = [...sortedByTime].reverse()

  oldestFirst.forEach(snapshot => {
    const name = snapshot.metadata.name
    const parentName = snapshot.spec.parentSnapshotRef?.name

    if (!parentName || !snapshotMap.has(parentName)) {
      columnAssignments.set(name, nextColumn++)
    } else {
      const siblings = childrenMap.get(parentName) || []
      const parentColumn = columnAssignments.get(parentName) ?? 0

      if (siblings.length === 1) {
        columnAssignments.set(name, parentColumn)
      } else {
        const siblingIndex = siblings.indexOf(name)
        if (siblingIndex === 0) {
          columnAssignments.set(name, parentColumn)
        } else {
          columnAssignments.set(name, nextColumn++)
        }
      }
    }
  })

  const nodes: TimelineNode[] = sortedByTime.map(snapshot => {
    const name = snapshot.metadata.name
    const parentName = snapshot.spec.parentSnapshotRef?.name
    const children = childrenMap.get(name) || []
    const parentColumn = parentName && snapshotMap.has(parentName)
      ? columnAssignments.get(parentName) ?? null
      : null

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

// SVG-based graph for reliable line rendering
function TimelineGraph({ nodes, maxColumn }: { nodes: TimelineNode[], maxColumn: number }) {
  const rowHeight = 120
  const colWidth = 24
  const graphWidth = (maxColumn + 1) * colWidth
  const graphHeight = nodes.length * rowHeight

  const lines: ReactElement[] = []
  const dots: ReactElement[] = []

  nodes.forEach((node, index) => {
    const { column, parentColumn, isCurrent, isBranchPoint } = node
    const colors = BRANCH_COLORS[column % BRANCH_COLORS.length]
    const cx = column * colWidth + colWidth / 2
    const cy = index * rowHeight + rowHeight / 2

    // Find parent index
    const parentIdx = parentColumn !== null
      ? nodes.findIndex(n => n.snapshot.metadata.name === node.snapshot.spec.parentSnapshotRef?.name)
      : -1
    const parentCy = parentIdx !== -1 ? parentIdx * rowHeight + rowHeight / 2 : -1

    if (parentIdx !== -1) {
      if (parentColumn === column) {
        // Same column - draw vertical line from this node to parent
        lines.push(
          <line
            key={`vline-${index}`}
            x1={cx}
            y1={cy}
            x2={cx}
            y2={parentCy}
            stroke={colors.line}
            strokeWidth={2}
          />
        )
      } else if (parentColumn !== null) {
        // Different column - this is a branch
        const parentCx = parentColumn * colWidth + colWidth / 2

        // Vertical line on THIS column from this node down to parent's row
        lines.push(
          <line
            key={`vline-branch-${index}`}
            x1={cx}
            y1={cy}
            x2={cx}
            y2={parentCy}
            stroke={colors.line}
            strokeWidth={2}
          />
        )

        // Horizontal connector at parent's row from this column to parent column
        lines.push(
          <line
            key={`hline-${index}`}
            x1={cx}
            y1={parentCy}
            x2={parentCx}
            y2={parentCy}
            stroke={colors.line}
            strokeWidth={2}
          />
        )
      }
    }

    // Draw dot
    const dotRadius = isCurrent ? 10 : isBranchPoint ? 8 : 5
    dots.push(
      <circle
        key={`dot-${index}`}
        cx={cx}
        cy={cy}
        r={dotRadius}
        fill={colors.line}
        stroke={isCurrent || isBranchPoint ? 'white' : 'none'}
        strokeWidth={isCurrent ? 3 : 2}
      />
    )

    if (isCurrent) {
      dots.push(
        <circle
          key={`dot-inner-${index}`}
          cx={cx}
          cy={cy}
          r={3}
          fill="white"
        />
      )
    }
  })

  return (
    <svg
      width={graphWidth}
      height={graphHeight}
      className="absolute left-0 top-0 pointer-events-none"
      style={{ minWidth: graphWidth }}
    >
      {lines}
      {dots}
    </svg>
  )
}

interface TimelineItemProps {
  node: TimelineNode
  maxColumn: number
  onRestore: (snapshot: Snapshot) => void
  onDelete: (snapshot: Snapshot) => void
  disabled?: boolean
}

function TimelineItem({ node, maxColumn, onRestore, onDelete, disabled }: TimelineItemProps) {
  const { snapshot, isCurrent, isBranchPoint, hasSiblings } = node
  const graphWidth = (maxColumn + 1) * 24

  return (
    <div className="flex" style={{ minHeight: 120 }}>
      {/* Space for graph */}
      <div style={{ width: graphWidth, minWidth: graphWidth }} className="flex-shrink-0" />

      {/* Content */}
      <div className="flex-1 py-2 pl-4 min-w-0">
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
                {hasSiblings && (
                  <span className="inline-flex items-center gap-1 rounded-md bg-orange-100 px-2 py-0.5 text-xs font-medium text-orange-700 dark:bg-orange-900/30 dark:text-orange-400">
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

              <p className="font-mono text-sm text-foreground truncate">
                {snapshot.metadata.name}
              </p>

              {snapshot.spec.description && (
                <p className="text-sm text-muted-foreground mt-1 italic">
                  &quot;{snapshot.spec.description}&quot;
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
                  <span className="flex items-center gap-1 text-blue-600 dark:text-blue-400">
                    <GitBranch className="h-3 w-3" />
                    from {snapshot.spec.parentSnapshotRef.name.split('-').slice(-2).join('-')}
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
        <TimelineGraph nodes={nodes} maxColumn={maxColumn} />
        {nodes.map((node) => (
          <TimelineItem
            key={node.snapshot.metadata.name}
            node={node}
            maxColumn={maxColumn}
            onRestore={onRestore}
            onDelete={onDelete}
            disabled={disabled}
          />
        ))}
      </div>
    </div>
  )
}
