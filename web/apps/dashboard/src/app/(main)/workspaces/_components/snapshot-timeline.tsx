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

// Metro-style branch colors (gitgraph.js inspired)
const BRANCH_COLORS = [
  { bg: 'bg-blue-500', stroke: '#3b82f6', fill: '#3b82f6' },
  { bg: 'bg-emerald-500', stroke: '#10b981', fill: '#10b981' },
  { bg: 'bg-violet-500', stroke: '#8b5cf6', fill: '#8b5cf6' },
  { bg: 'bg-amber-500', stroke: '#f59e0b', fill: '#f59e0b' },
  { bg: 'bg-rose-500', stroke: '#f43f5e', fill: '#f43f5e' },
  { bg: 'bg-cyan-500', stroke: '#06b6d4', fill: '#06b6d4' },
]

interface SnapshotWithLane {
  snapshot: Snapshot
  lane: number
  isCurrent: boolean
  parentLane: number | null
}

interface GraphRow {
  item: SnapshotWithLane
  activeLanes: Set<number>
  branchFrom: { fromLane: number; toLane: number } | null
}

function buildGraph(snapshots: Snapshot[], currentSnapshotName?: string): GraphRow[] {
  if (snapshots.length === 0) return []

  // Sort by creation time (newest first)
  const sorted = [...snapshots].sort((a, b) => {
    const aTime = new Date(a.status.createdAt || a.metadata.creationTimestamp).getTime()
    const bTime = new Date(b.status.createdAt || b.metadata.creationTimestamp).getTime()
    return bTime - aTime
  })

  // Build maps
  const snapshotMap = new Map<string, Snapshot>()
  const childrenMap = new Map<string, string[]>()

  sorted.forEach(s => {
    snapshotMap.set(s.metadata.name, s)
    const parentName = s.spec.parentSnapshotRef?.name
    if (parentName) {
      const children = childrenMap.get(parentName) || []
      children.push(s.metadata.name)
      childrenMap.set(parentName, children)
    }
  })

  // Assign lanes using a simple algorithm
  const laneMap = new Map<string, number>()
  let maxLane = 0

  // Process chronologically (oldest first) for lane assignment
  const chronological = [...sorted].reverse()

  chronological.forEach(snapshot => {
    const name = snapshot.metadata.name
    if (laneMap.has(name)) return

    const parentName = snapshot.spec.parentSnapshotRef?.name
    const children = childrenMap.get(name) || []

    if (parentName && laneMap.has(parentName)) {
      const parentLane = laneMap.get(parentName)!
      const siblings = childrenMap.get(parentName) || []
      const siblingIndex = siblings.indexOf(name)

      if (siblingIndex === 0) {
        // First child stays on parent's lane
        laneMap.set(name, parentLane)
      } else {
        // Additional children get new lanes
        maxLane++
        laneMap.set(name, maxLane)
      }
    } else {
      // Root or orphan - assign to lane 0 if available, otherwise new lane
      if (!laneMap.has(name)) {
        const usedLanes = new Set(laneMap.values())
        if (!usedLanes.has(0)) {
          laneMap.set(name, 0)
        } else {
          maxLane++
          laneMap.set(name, maxLane)
        }
      }
    }
  })

  // Build graph rows (newest first for display)
  const rows: GraphRow[] = []
  const activeLanes = new Set<number>()

  // Process from oldest to newest to track active lanes
  for (let i = sorted.length - 1; i >= 0; i--) {
    const snapshot = sorted[i]
    const lane = laneMap.get(snapshot.metadata.name) || 0
    activeLanes.add(lane)
  }

  // Now build rows from newest to oldest
  const currentActiveLanes = new Set(activeLanes)

  sorted.forEach((snapshot, idx) => {
    const name = snapshot.metadata.name
    const lane = laneMap.get(name) || 0
    const parentName = snapshot.spec.parentSnapshotRef?.name
    const parentLane = parentName && laneMap.has(parentName) ? laneMap.get(parentName)! : null

    // Check if this is where a branch starts (has parent on different lane)
    let branchFrom: { fromLane: number; toLane: number } | null = null
    if (parentLane !== null && parentLane !== lane) {
      branchFrom = { fromLane: parentLane, toLane: lane }
    }

    rows.push({
      item: {
        snapshot,
        lane,
        isCurrent: name === currentSnapshotName,
        parentLane,
      },
      activeLanes: new Set(currentActiveLanes),
      branchFrom,
    })

    // For the next row (older), check if this lane should still be active
    // A lane becomes inactive after its last (oldest) commit on that lane
    const isLastOnLane = !sorted.slice(idx + 1).some(s => laneMap.get(s.metadata.name) === lane)
    if (isLastOnLane) {
      currentActiveLanes.delete(lane)
    }
  })

  return rows
}

const LANE_WIDTH = 20
const DOT_SIZE = 10
const LINE_WIDTH = 2

interface GraphColumnProps {
  row: GraphRow
  totalLanes: number
  isLast: boolean
  rowHeight: number
}

function GraphColumn({ row, totalLanes, isLast, rowHeight }: GraphColumnProps) {
  const { item, activeLanes, branchFrom } = row
  const width = Math.max(totalLanes, 1) * LANE_WIDTH + 12

  return (
    <svg
      width={width}
      height={rowHeight}
      className="flex-shrink-0"
      style={{ minHeight: rowHeight }}
    >
      {/* Draw vertical lines for active lanes */}
      {Array.from(activeLanes).map(laneIdx => {
        const x = laneIdx * LANE_WIDTH + LANE_WIDTH / 2
        const color = BRANCH_COLORS[laneIdx % BRANCH_COLORS.length]
        const isCurrentLane = laneIdx === item.lane

        // Line should go full height, except for the last item on this lane
        const lineEnd = isLast && isCurrentLane ? rowHeight / 2 : rowHeight

        return (
          <line
            key={laneIdx}
            x1={x}
            y1={0}
            x2={x}
            y2={lineEnd}
            stroke={color.stroke}
            strokeWidth={LINE_WIDTH}
          />
        )
      })}

      {/* Draw branch curve if this is a branch point */}
      {branchFrom && (
        <path
          d={`M ${branchFrom.fromLane * LANE_WIDTH + LANE_WIDTH / 2} 0
              Q ${branchFrom.fromLane * LANE_WIDTH + LANE_WIDTH / 2} ${rowHeight / 2},
                ${branchFrom.toLane * LANE_WIDTH + LANE_WIDTH / 2} ${rowHeight / 2}`}
          fill="none"
          stroke={BRANCH_COLORS[branchFrom.toLane % BRANCH_COLORS.length].stroke}
          strokeWidth={LINE_WIDTH}
        />
      )}

      {/* Draw commit dot */}
      <circle
        cx={item.lane * LANE_WIDTH + LANE_WIDTH / 2}
        cy={rowHeight / 2}
        r={DOT_SIZE / 2}
        fill={BRANCH_COLORS[item.lane % BRANCH_COLORS.length].fill}
        stroke={item.isCurrent ? '#fff' : 'none'}
        strokeWidth={item.isCurrent ? 2 : 0}
      />

      {/* Outer ring for current/HEAD */}
      {item.isCurrent && (
        <circle
          cx={item.lane * LANE_WIDTH + LANE_WIDTH / 2}
          cy={rowHeight / 2}
          r={DOT_SIZE / 2 + 3}
          fill="none"
          stroke={BRANCH_COLORS[item.lane % BRANCH_COLORS.length].stroke}
          strokeWidth={2}
          opacity={0.5}
        />
      )}
    </svg>
  )
}

interface SnapshotRowProps {
  row: GraphRow
  totalLanes: number
  onRestore: (snapshot: Snapshot) => void
  onDelete: (snapshot: Snapshot) => void
  disabled?: boolean
  isLast: boolean
}

function SnapshotRow({ row, totalLanes, onRestore, onDelete, disabled, isLast }: SnapshotRowProps) {
  const { item } = row
  const { snapshot, isCurrent } = item
  const shortHash = getShortHash(snapshot.metadata.name)
  const rowHeight = 52

  return (
    <div className="flex items-stretch">
      <GraphColumn row={row} totalLanes={totalLanes} isLast={isLast} rowHeight={rowHeight} />

      <div
        className={cn(
          "group flex-1 py-2 px-3 rounded-md border transition-colors my-0.5",
          isCurrent
            ? "bg-blue-50 border-blue-200 dark:bg-blue-950/30 dark:border-blue-800"
            : "border-transparent hover:bg-muted/50 hover:border-border"
        )}
      >
        <div className="flex items-center justify-between gap-3">
          <div className="flex items-center gap-2 min-w-0">
            <code className={cn(
              "text-sm font-mono",
              isCurrent ? "text-blue-600 dark:text-blue-400 font-semibold" : "text-foreground"
            )}>
              {shortHash}
            </code>

            {isCurrent && (
              <Badge variant="secondary" className="text-[10px] px-1.5 py-0 h-4 bg-blue-100 text-blue-700 dark:bg-blue-900 dark:text-blue-300">
                HEAD
              </Badge>
            )}

            {getStateBadge(snapshot.status.state)}

            {snapshot.spec.description && (
              <span className="text-sm text-muted-foreground truncate">
                {snapshot.spec.description}
              </span>
            )}
          </div>

          <div className="flex items-center gap-2">
            <span className="flex items-center gap-1 text-xs text-muted-foreground">
              <Clock className="h-3 w-3" />
              {formatTimeAgo(snapshot.status.createdAt || snapshot.metadata.creationTimestamp)}
            </span>

            {snapshot.status.sizeHuman && snapshot.status.sizeHuman !== '0 B' && (
              <span className="flex items-center gap-1 text-xs text-muted-foreground">
                <HardDrive className="h-3 w-3" />
                {snapshot.status.sizeHuman}
              </span>
            )}

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
        </div>

        {snapshot.status.state === 'Failed' && snapshot.status.message && (
          <p className="mt-1 text-xs text-red-600 dark:text-red-400">
            {snapshot.status.message}
          </p>
        )}
      </div>
    </div>
  )
}

export function SnapshotTimeline({ snapshots, onRestore, onDelete, disabled, currentSnapshotName }: SnapshotTimelineProps) {
  const rows = useMemo(() => buildGraph(snapshots, currentSnapshotName), [snapshots, currentSnapshotName])

  const totalLanes = useMemo(() => {
    if (rows.length === 0) return 1
    return Math.max(...rows.map(r => r.item.lane)) + 1
  }, [rows])

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

      <div className="space-y-0">
        {rows.map((row, idx) => (
          <SnapshotRow
            key={row.item.snapshot.metadata.name}
            row={row}
            totalLanes={totalLanes}
            onRestore={onRestore}
            onDelete={onDelete}
            disabled={disabled}
            isLast={idx === rows.length - 1}
          />
        ))}
      </div>
    </div>
  )
}
