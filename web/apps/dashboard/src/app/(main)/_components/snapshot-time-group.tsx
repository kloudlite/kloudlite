'use client'

import type { Snapshot } from '@/lib/services/snapshot.service'
import { SnapshotCard } from './snapshot-card'

interface SnapshotTimeGroupProps {
  label: string
  snapshots: Snapshot[]
  currentSnapshotName?: string
  onRestore: (snapshot: Snapshot) => void
  onDelete: (snapshot: Snapshot) => void
  onPush?: (snapshot: Snapshot) => void
  disabled?: boolean
  showTimeline?: boolean
}

// Build graph data for timeline visualization
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

function buildGraph(
  snapshots: Snapshot[],
  currentSnapshotName?: string
): GraphRow[] {
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

  sorted.forEach((s) => {
    snapshotMap.set(s.name, s)
    const parentName = s.parent
    if (parentName) {
      const children = childrenMap.get(parentName) || []
      children.push(s.name)
      childrenMap.set(parentName, children)
    }
  })

  // Assign lanes (max 3 lanes as per design)
  const laneMap = new Map<string, number>()
  let maxLane = 0
  const MAX_LANES = 2 // 0, 1, 2 = 3 lanes total

  const chronological = [...sorted].reverse()

  chronological.forEach((snapshot) => {
    const name = snapshot.name
    if (laneMap.has(name)) return

    const parentName = snapshot.parent

    if (parentName && laneMap.has(parentName)) {
      const parentLane = laneMap.get(parentName)!
      const siblings = childrenMap.get(parentName) || []
      const siblingIndex = siblings.indexOf(name)

      if (siblingIndex === 0) {
        // First child stays in parent's lane
        laneMap.set(name, parentLane)
      } else if (maxLane < MAX_LANES) {
        // Branch to new lane
        maxLane++
        laneMap.set(name, maxLane)
      } else {
        // Reuse existing lane
        laneMap.set(name, (maxLane % (MAX_LANES + 1)))
      }
    } else {
      // Root snapshot or orphan
      if (!laneMap.has(name)) {
        const usedLanes = new Set(laneMap.values())
        if (!usedLanes.has(0)) {
          laneMap.set(name, 0)
        } else if (maxLane < MAX_LANES) {
          maxLane++
          laneMap.set(name, maxLane)
        } else {
          laneMap.set(name, 0)
        }
      }
    }
  })

  // Build graph rows
  const rows: GraphRow[] = []
  const activeLanes = new Set<number>()

  // Collect all active lanes
  for (let i = sorted.length - 1; i >= 0; i--) {
    const snapshot = sorted[i]
    const lane = laneMap.get(snapshot.name) || 0
    activeLanes.add(lane)
  }

  const currentActiveLanes = new Set(activeLanes)

  sorted.forEach((snapshot) => {
    const name = snapshot.name
    const lane = laneMap.get(name) || 0
    const isCurrent = name === currentSnapshotName

    const parentName = snapshot.parent
    const parentLane = parentName ? laneMap.get(parentName) || null : null

    let branchFrom: { fromLane: number; toLane: number } | null = null
    if (parentLane !== null && parentLane !== lane) {
      branchFrom = { fromLane: parentLane, toLane: lane }
    }

    rows.push({
      item: { snapshot, lane, isCurrent, parentLane },
      activeLanes: new Set(currentActiveLanes),
      branchFrom,
    })
  })

  return rows
}

export function SnapshotTimeGroup({
  label,
  snapshots,
  currentSnapshotName,
  onRestore,
  onDelete,
  onPush,
  disabled = false,
  showTimeline = true,
}: SnapshotTimeGroupProps) {
  const graphRows = showTimeline ? buildGraph(snapshots, currentSnapshotName) : []

  return (
    <div className="space-y-3">
      {/* Time Group Label */}
      <h3 className="text-xs font-medium uppercase tracking-wider text-muted-foreground pl-1">
        {label}
      </h3>

      {/* Snapshots */}
      <div className="space-y-3">
        {showTimeline ? (
          // With timeline graph
          graphRows.map(({ item, activeLanes, branchFrom }) => (
            <SnapshotCard
              key={item.snapshot.name}
              snapshot={item.snapshot}
              isCurrent={item.isCurrent}
              onRestore={() => onRestore(item.snapshot)}
              onDelete={() => onDelete(item.snapshot)}
              onPush={onPush ? () => onPush(item.snapshot) : undefined}
              disabled={disabled}
              timelineData={{
                lane: item.lane,
                activeLanes,
                branchFrom,
              }}
            />
          ))
        ) : (
          // Without timeline graph
          snapshots.map((snapshot) => (
            <SnapshotCard
              key={snapshot.name}
              snapshot={snapshot}
              isCurrent={snapshot.name === currentSnapshotName}
              onRestore={() => onRestore(snapshot)}
              onDelete={() => onDelete(snapshot)}
              onPush={onPush ? () => onPush(snapshot) : undefined}
              disabled={disabled}
            />
          ))
        )}
      </div>
    </div>
  )
}
