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

// Branch colors for different lanes
const BRANCH_COLORS = [
  'bg-blue-500',
  'bg-green-500',
  'bg-purple-500',
  'bg-orange-500',
  'bg-pink-500',
  'bg-cyan-500',
]

interface SnapshotWithLane {
  snapshot: Snapshot
  lane: number
  isCurrent: boolean
  connectsTo: number | null // lane it connects to (for merge visualization)
}

function buildLaneAssignments(snapshots: Snapshot[], currentSnapshotName?: string): SnapshotWithLane[] {
  if (snapshots.length === 0) return []

  // Sort by creation time (newest first)
  const sorted = [...snapshots].sort((a, b) => {
    const aTime = new Date(a.status.createdAt || a.metadata.creationTimestamp).getTime()
    const bTime = new Date(b.status.createdAt || b.metadata.creationTimestamp).getTime()
    return bTime - aTime
  })

  // Build parent -> children map
  const childrenMap = new Map<string, string[]>()
  const snapshotMap = new Map<string, Snapshot>()

  sorted.forEach(s => {
    snapshotMap.set(s.metadata.name, s)
    const parentName = s.spec.parentSnapshotRef?.name
    if (parentName) {
      const existing = childrenMap.get(parentName) || []
      existing.push(s.metadata.name)
      childrenMap.set(parentName, existing)
    }
  })

  // Find snapshots that are branch points (have multiple children)
  const branchPoints = new Set<string>()
  childrenMap.forEach((children, parent) => {
    if (children.length > 1) {
      branchPoints.add(parent)
    }
  })

  // Assign lanes - newer snapshots get processed first
  const laneAssignments = new Map<string, number>()
  const activeLanes = new Set<number>()
  let nextLane = 0

  // Process in reverse chronological order (newest first)
  sorted.forEach(snapshot => {
    const name = snapshot.metadata.name
    const parentName = snapshot.spec.parentSnapshotRef?.name

    // Check if we already have a lane assignment (from being a parent of processed snapshot)
    if (laneAssignments.has(name)) {
      return
    }

    // If parent exists and has a lane, check if we need a new lane
    if (parentName && laneAssignments.has(parentName)) {
      const parentLane = laneAssignments.get(parentName)!
      const siblings = childrenMap.get(parentName) || []

      // If this is not the first child of the parent, create a new lane
      const isFirstChild = siblings[0] === name
      if (isFirstChild) {
        laneAssignments.set(name, parentLane)
      } else {
        // New branch - assign new lane
        laneAssignments.set(name, nextLane)
        activeLanes.add(nextLane)
        nextLane++
      }
    } else if (parentName && snapshotMap.has(parentName)) {
      // Parent exists but not yet assigned - assign parent's lane first
      const parentLane = nextLane
      laneAssignments.set(parentName, parentLane)
      activeLanes.add(parentLane)
      nextLane++

      const siblings = childrenMap.get(parentName) || []
      const isFirstChild = siblings[0] === name
      if (isFirstChild) {
        laneAssignments.set(name, parentLane)
      } else {
        laneAssignments.set(name, nextLane)
        activeLanes.add(nextLane)
        nextLane++
      }
    } else {
      // Root snapshot or parent not in list - new lane
      laneAssignments.set(name, nextLane)
      activeLanes.add(nextLane)
      nextLane++
    }
  })

  // Build result with lane info
  return sorted.map(snapshot => {
    const lane = laneAssignments.get(snapshot.metadata.name) || 0
    const parentName = snapshot.spec.parentSnapshotRef?.name
    let connectsTo: number | null = null

    if (parentName && laneAssignments.has(parentName)) {
      const parentLane = laneAssignments.get(parentName)!
      if (parentLane !== lane) {
        connectsTo = parentLane
      }
    }

    return {
      snapshot,
      lane,
      isCurrent: snapshot.metadata.name === currentSnapshotName,
      connectsTo,
    }
  })
}

interface SnapshotItemProps {
  item: SnapshotWithLane
  totalLanes: number
  onRestore: (snapshot: Snapshot) => void
  onDelete: (snapshot: Snapshot) => void
  disabled?: boolean
  isLast: boolean
  activeLanesBelow: Set<number>
}

function SnapshotItem({ item, totalLanes, onRestore, onDelete, disabled, isLast, activeLanesBelow }: SnapshotItemProps) {
  const { snapshot, lane, isCurrent, connectsTo } = item
  const shortHash = getShortHash(snapshot.metadata.name)
  const laneWidth = 16 // pixels per lane

  return (
    <div className="flex">
      {/* Lane visualization */}
      <div className="flex-shrink-0 relative" style={{ width: Math.max(totalLanes, 1) * laneWidth + 8 }}>
        {/* Vertical lines for active lanes */}
        {Array.from({ length: totalLanes }).map((_, idx) => {
          const isActive = activeLanesBelow.has(idx) || idx === lane
          if (!isActive) return null

          return (
            <div
              key={idx}
              className={cn(
                "absolute w-0.5 top-0",
                isLast && idx === lane ? "h-4" : "h-full",
                BRANCH_COLORS[idx % BRANCH_COLORS.length]
              )}
              style={{ left: idx * laneWidth + 6 }}
            />
          )
        })}

        {/* Dot for this snapshot */}
        <div
          className={cn(
            "absolute w-2.5 h-2.5 rounded-full top-4 -translate-y-1/2",
            isCurrent && "ring-2 ring-offset-1 ring-offset-background",
            BRANCH_COLORS[lane % BRANCH_COLORS.length],
            isCurrent && "ring-blue-300 dark:ring-blue-700"
          )}
          style={{ left: lane * laneWidth + 4 }}
        />

        {/* Connection line to parent lane if different */}
        {connectsTo !== null && (
          <div
            className={cn(
              "absolute h-0.5 top-4 -translate-y-1/2",
              BRANCH_COLORS[lane % BRANCH_COLORS.length]
            )}
            style={{
              left: Math.min(lane, connectsTo) * laneWidth + 8,
              width: Math.abs(connectsTo - lane) * laneWidth - 2,
            }}
          />
        )}
      </div>

      {/* Content */}
      <div
        className={cn(
          "group flex-1 py-2 px-3 rounded-lg border transition-colors mb-1.5",
          isCurrent
            ? "bg-blue-50 border-blue-200 dark:bg-blue-950/30 dark:border-blue-800"
            : "border-border hover:bg-muted/50"
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
          </div>

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

        {snapshot.spec.description && (
          <p className="text-sm text-foreground mt-1">
            {snapshot.spec.description}
          </p>
        )}

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
  const items = useMemo(() => buildLaneAssignments(snapshots, currentSnapshotName), [snapshots, currentSnapshotName])

  const totalLanes = useMemo(() => {
    if (items.length === 0) return 0
    return Math.max(...items.map(i => i.lane)) + 1
  }, [items])

  // Calculate which lanes are active below each item
  const activeLanesBelowList = useMemo(() => {
    const result: Set<number>[] = []
    const activeLanes = new Set<number>()

    // Process from bottom to top
    for (let i = items.length - 1; i >= 0; i--) {
      result[i] = new Set(activeLanes)
      activeLanes.add(items[i].lane)
    }

    return result
  }, [items])

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
        {items.map((item, idx) => (
          <SnapshotItem
            key={item.snapshot.metadata.name}
            item={item}
            totalLanes={totalLanes}
            onRestore={onRestore}
            onDelete={onDelete}
            disabled={disabled}
            isLast={idx === items.length - 1}
            activeLanesBelow={activeLanesBelowList[idx]}
          />
        ))}
      </div>
    </div>
  )
}
