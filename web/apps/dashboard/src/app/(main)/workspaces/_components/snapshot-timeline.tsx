'use client'

import { useMemo } from 'react'
import {
  Clock,
  HardDrive,
  AlertCircle,
  Loader2,
  RotateCcw,
  Trash2,
  Cloud,
  CloudUpload,
} from 'lucide-react'
import { Button, Badge } from '@kloudlite/ui'
import { cn } from '@/lib/utils'
import type { Snapshot } from '@/lib/services/snapshot.service'

interface SnapshotTimelineProps {
  snapshots: Snapshot[]
  onRestore: (snapshot: Snapshot) => void
  onDelete: (snapshot: Snapshot) => void
  onPush?: (snapshot: Snapshot) => void
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

function getStateBadge(state: Snapshot['state']) {
  switch (state) {
    case 'Ready':
    case 'Completed':
      return null
    case 'Creating':
    case 'Uploading':
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
    case 'Pushing':
      return (
        <span className="inline-flex items-center gap-1 text-xs text-cyan-600 dark:text-cyan-400">
          <Loader2 className="h-3 w-3 animate-spin" />
          Pushing
        </span>
      )
    case 'Pulling':
      return (
        <span className="inline-flex items-center gap-1 text-xs text-cyan-600 dark:text-cyan-400">
          <Loader2 className="h-3 w-3 animate-spin" />
          Pulling
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

// Metro-style branch colors
const BRANCH_COLORS = [
  { stroke: '#3b82f6', fill: '#3b82f6' },
  { stroke: '#10b981', fill: '#10b981' },
  { stroke: '#8b5cf6', fill: '#8b5cf6' },
  { stroke: '#f59e0b', fill: '#f59e0b' },
  { stroke: '#f43f5e', fill: '#f43f5e' },
  { stroke: '#06b6d4', fill: '#06b6d4' },
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
    const aTime = new Date(a.createdAt || '').getTime()
    const bTime = new Date(b.createdAt || '').getTime()
    return bTime - aTime
  })

  // Build maps
  const snapshotMap = new Map<string, Snapshot>()
  const childrenMap = new Map<string, string[]>()

  sorted.forEach(s => {
    snapshotMap.set(s.name, s)
    const parentName = s.parent
    if (parentName) {
      const children = childrenMap.get(parentName) || []
      children.push(s.name)
      childrenMap.set(parentName, children)
    }
  })

  // Assign lanes
  const laneMap = new Map<string, number>()
  let maxLane = 0

  const chronological = [...sorted].reverse()

  chronological.forEach(snapshot => {
    const name = snapshot.name
    if (laneMap.has(name)) return

    const parentName = snapshot.parent

    if (parentName && laneMap.has(parentName)) {
      const parentLane = laneMap.get(parentName)!
      const siblings = childrenMap.get(parentName) || []
      const siblingIndex = siblings.indexOf(name)

      if (siblingIndex === 0) {
        laneMap.set(name, parentLane)
      } else {
        maxLane++
        laneMap.set(name, maxLane)
      }
    } else {
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

  // Build graph rows
  const rows: GraphRow[] = []
  const activeLanes = new Set<number>()

  for (let i = sorted.length - 1; i >= 0; i--) {
    const snapshot = sorted[i]
    const lane = laneMap.get(snapshot.name) || 0
    activeLanes.add(lane)
  }

  const currentActiveLanes = new Set(activeLanes)

  sorted.forEach((snapshot, idx) => {
    const name = snapshot.name
    const lane = laneMap.get(name) || 0
    const parentName = snapshot.parent
    const parentLane = parentName && laneMap.has(parentName) ? laneMap.get(parentName)! : null

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

    const isLastOnLane = !sorted.slice(idx + 1).some(s => laneMap.get(s.name) === lane)
    if (isLastOnLane) {
      currentActiveLanes.delete(lane)
    }
  })

  return rows
}

const LANE_WIDTH = 16
const DOT_SIZE = 8
const LINE_WIDTH = 2

interface SnapshotRowProps {
  row: GraphRow
  totalLanes: number
  onRestore: (snapshot: Snapshot) => void
  onDelete: (snapshot: Snapshot) => void
  onPush?: (snapshot: Snapshot) => void
  disabled?: boolean
  isFirst: boolean
  isLast: boolean
}

function SnapshotRow({ row, totalLanes, onRestore, onDelete, onPush, disabled, isFirst, isLast }: SnapshotRowProps) {
  const { item, activeLanes, branchFrom } = row
  const { snapshot, isCurrent } = item
  const shortHash = getShortHash(snapshot.name)
  const isPushed = !!snapshot.registry?.digest
  const isActionable = snapshot.state === 'Ready' || snapshot.state === 'Completed'
  const canDelete = isActionable || snapshot.state === 'Failed'

  const graphWidth = Math.max(totalLanes, 1) * LANE_WIDTH + 8
  const dotX = item.lane * LANE_WIDTH + LANE_WIDTH / 2

  return (
    <div className="flex items-center">
      {/* Graph column */}
      <div
        className="relative flex-shrink-0 self-stretch"
        style={{ width: graphWidth }}
      >
        {/* Vertical lines for active lanes */}
        {Array.from(activeLanes).map(laneIdx => {
          const x = laneIdx * LANE_WIDTH + LANE_WIDTH / 2
          const color = BRANCH_COLORS[laneIdx % BRANCH_COLORS.length]
          const isCurrentLane = laneIdx === item.lane

          return (
            <div
              key={laneIdx}
              className="absolute"
              style={{
                left: x - LINE_WIDTH / 2,
                top: isFirst && isCurrentLane ? '50%' : 0,
                bottom: isLast && isCurrentLane ? '50%' : 0,
                width: LINE_WIDTH,
                backgroundColor: color.stroke,
              }}
            />
          )
        })}

        {/* Branch curve */}
        {branchFrom && (
          <svg
            className="absolute inset-0 pointer-events-none"
            style={{ width: graphWidth, height: '100%' }}
            viewBox={`0 0 ${graphWidth} 44`}
            preserveAspectRatio="none"
          >
            <path
              d={`M ${branchFrom.fromLane * LANE_WIDTH + LANE_WIDTH / 2} 0
                  C ${branchFrom.fromLane * LANE_WIDTH + LANE_WIDTH / 2} 22,
                    ${branchFrom.toLane * LANE_WIDTH + LANE_WIDTH / 2} 22,
                    ${branchFrom.toLane * LANE_WIDTH + LANE_WIDTH / 2} 22`}
              fill="none"
              stroke={BRANCH_COLORS[branchFrom.toLane % BRANCH_COLORS.length].stroke}
              strokeWidth={LINE_WIDTH}
              vectorEffect="non-scaling-stroke"
            />
          </svg>
        )}

        {/* Commit dot */}
        <div
          className="absolute top-1/2 -translate-y-1/2 rounded-full"
          style={{
            left: dotX - DOT_SIZE / 2,
            width: DOT_SIZE,
            height: DOT_SIZE,
            backgroundColor: BRANCH_COLORS[item.lane % BRANCH_COLORS.length].fill,
            boxShadow: isCurrent ? `0 0 0 3px ${BRANCH_COLORS[item.lane % BRANCH_COLORS.length].stroke}33` : undefined,
          }}
        />
      </div>

      {/* Content - Grid layout for consistent columns */}
      <div
        className={cn(
          "group flex-1 grid items-center gap-x-3 py-2 px-3 rounded-md transition-colors min-h-[44px]",
          isCurrent
            ? "bg-blue-50 dark:bg-blue-950/30"
            : "hover:bg-muted/50"
        )}
        style={{ gridTemplateColumns: '1fr auto auto auto' }}
      >
        {/* Column 1: Name, badges, description */}
        <div className="flex items-center gap-2 min-w-0 overflow-hidden">
          <code className={cn(
            "text-sm font-mono flex-shrink-0",
            isCurrent ? "text-blue-600 dark:text-blue-400 font-semibold" : "text-foreground"
          )}>
            {shortHash}
          </code>

          {isCurrent && (
            <Badge variant="secondary" className="text-[10px] px-1.5 py-0 h-4 flex-shrink-0 bg-blue-100 text-blue-700 dark:bg-blue-900 dark:text-blue-300">
              HEAD
            </Badge>
          )}

          {isPushed && (
            <Badge variant="secondary" className="text-[10px] px-1.5 py-0 h-4 gap-1 flex-shrink-0 bg-cyan-100 text-cyan-700 dark:bg-cyan-900 dark:text-cyan-300">
              <Cloud className="h-3 w-3" />
              {snapshot.registry?.tag || 'pushed'}
            </Badge>
          )}

          {getStateBadge(snapshot.state)}

          {snapshot.description && (
            <span className="text-xs text-muted-foreground truncate" title={snapshot.description}>
              {snapshot.description}
            </span>
          )}
        </div>

        {/* Column 2: Time */}
        <div className="flex items-center gap-1 text-xs text-muted-foreground w-[70px] justify-end flex-shrink-0">
          {snapshot.createdAt && (
            <>
              <Clock className="h-3 w-3 flex-shrink-0" />
              <span className="whitespace-nowrap">{formatTimeAgo(snapshot.createdAt)}</span>
            </>
          )}
        </div>

        {/* Column 3: Size */}
        <div className="flex items-center gap-1 text-xs text-muted-foreground w-[70px] justify-end flex-shrink-0">
          {snapshot.sizeHuman && snapshot.sizeHuman !== '0 B' && (
            <>
              <HardDrive className="h-3 w-3 flex-shrink-0" />
              <span className="whitespace-nowrap">{snapshot.sizeHuman}</span>
            </>
          )}
        </div>

        {/* Column 4: Actions - always visible */}
        <div className="flex items-center gap-1 justify-end w-[110px] flex-shrink-0">
          {isActionable && !isPushed && onPush && (
            <Button
              variant="ghost"
              size="sm"
              onClick={() => onPush(snapshot)}
              disabled={disabled}
              className="h-7 w-7 p-0 text-muted-foreground hover:text-foreground"
              title="Push to registry"
            >
              <CloudUpload className="h-3.5 w-3.5" />
            </Button>
          )}
          {isActionable && !isCurrent && (
            <Button
              variant="ghost"
              size="sm"
              onClick={() => onRestore(snapshot)}
              disabled={disabled}
              className="h-7 w-7 p-0 text-muted-foreground hover:text-foreground"
              title="Restore snapshot"
            >
              <RotateCcw className="h-3.5 w-3.5" />
            </Button>
          )}
          {canDelete && (
            <Button
              variant="ghost"
              size="sm"
              onClick={() => onDelete(snapshot)}
              disabled={disabled}
              className="h-7 w-7 p-0 text-muted-foreground hover:text-destructive"
              title="Delete snapshot"
            >
              <Trash2 className="h-3.5 w-3.5" />
            </Button>
          )}
        </div>
      </div>
    </div>
  )
}

export function SnapshotTimeline({ snapshots, onRestore, onDelete, onPush, disabled, currentSnapshotName }: SnapshotTimelineProps) {
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
            key={row.item.snapshot.name}
            row={row}
            totalLanes={totalLanes}
            onRestore={onRestore}
            onDelete={onDelete}
            onPush={onPush}
            disabled={disabled}
            isFirst={idx === 0}
            isLast={idx === rows.length - 1}
          />
        ))}
      </div>
    </div>
  )
}
