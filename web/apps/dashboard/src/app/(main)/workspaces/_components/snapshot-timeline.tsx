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
  children: SnapshotNode[]
  hasParent: boolean
  isLastInChain: boolean
}

interface LineageChain {
  nodes: SnapshotNode[]
  rootName: string | null
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
 * Build lineage chains from flat snapshot list
 * Groups snapshots by their lineage (parent-child relationships)
 */
function buildLineageChains(snapshots: Snapshot[]): LineageChain[] {
  if (snapshots.length === 0) return []

  // Create a map for quick lookup
  const snapshotMap = new Map<string, Snapshot>()
  snapshots.forEach(s => snapshotMap.set(s.metadata.name, s))

  // Build parent -> children map
  const childrenMap = new Map<string, Snapshot[]>()
  const hasParentInList = new Set<string>()

  snapshots.forEach(snapshot => {
    const parentName = snapshot.spec.parentSnapshotRef?.name
    if (parentName && snapshotMap.has(parentName)) {
      hasParentInList.add(snapshot.metadata.name)
      const children = childrenMap.get(parentName) || []
      children.push(snapshot)
      childrenMap.set(parentName, children)
    }
  })

  // Find root snapshots (no parent or parent not in list)
  const rootSnapshots = snapshots.filter(s => !hasParentInList.has(s.metadata.name))

  // Build chains starting from each root
  const chains: LineageChain[] = []

  // Helper to build node tree
  function buildNodeTree(snapshot: Snapshot, isLastInChain: boolean): SnapshotNode {
    const children = childrenMap.get(snapshot.metadata.name) || []
    // Sort children by creation time (newest first)
    children.sort((a, b) =>
      new Date(b.status.createdAt || b.metadata.creationTimestamp).getTime() -
      new Date(a.status.createdAt || a.metadata.creationTimestamp).getTime()
    )

    return {
      snapshot,
      hasParent: hasParentInList.has(snapshot.metadata.name),
      isLastInChain: isLastInChain && children.length === 0,
      children: children.map((child, idx) =>
        buildNodeTree(child, idx === children.length - 1)
      ),
    }
  }

  // Helper to flatten tree to array (for rendering)
  function flattenTree(node: SnapshotNode, result: SnapshotNode[]): void {
    result.push(node)
    node.children.forEach(child => flattenTree(child, result))
  }

  // Sort roots by creation time (newest first)
  rootSnapshots.sort((a, b) =>
    new Date(b.status.createdAt || b.metadata.creationTimestamp).getTime() -
    new Date(a.status.createdAt || a.metadata.creationTimestamp).getTime()
  )

  rootSnapshots.forEach(root => {
    const tree = buildNodeTree(root, true)
    const nodes: SnapshotNode[] = []
    flattenTree(tree, nodes)

    // Reverse to show oldest first within a chain (parent before children)
    // But chains themselves are sorted newest-root first
    chains.push({
      nodes: nodes.reverse(),
      rootName: root.metadata.name,
    })
  })

  return chains
}

interface TimelineItemProps {
  node: SnapshotNode
  isFirst: boolean
  isLast: boolean
  onRestore: (snapshot: Snapshot) => void
  onDelete: (snapshot: Snapshot) => void
  disabled?: boolean
}

function TimelineItem({ node, isFirst, isLast, onRestore, onDelete, disabled }: TimelineItemProps) {
  const { snapshot, hasParent } = node

  return (
    <div className="relative pl-8">
      {/* Continuous vertical line */}
      {!isLast && (
        <div
          className="absolute left-[7px] top-4 bottom-0 w-0.5 bg-border"
        />
      )}
      {!isFirst && (
        <div
          className="absolute left-[7px] top-0 h-4 w-0.5 bg-border"
        />
      )}

      {/* Node dot - positioned absolutely on the line */}
      <div
        className={cn(
          "absolute left-0 top-4 h-4 w-4 rounded-full border-2 flex items-center justify-center bg-background",
          hasParent
            ? "bg-primary border-primary"
            : "border-muted-foreground"
        )}
      >
        {!hasParent && (
          <div className="h-1.5 w-1.5 rounded-full bg-muted-foreground" />
        )}
      </div>

      {/* Snapshot card */}
      <div className="pb-4">
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
                  <span className="flex items-center gap-1 text-primary">
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
  const chains = useMemo(() => buildLineageChains(snapshots), [snapshots])

  if (snapshots.length === 0) {
    return null
  }

  return (
    <div className="space-y-2">
      <h4 className="text-sm font-medium flex items-center gap-2">
        <GitBranch className="h-4 w-4" />
        Snapshot History ({snapshots.length})
      </h4>

      <div className="space-y-6">
        {chains.map((chain, chainIdx) => (
          <div key={chain.rootName || chainIdx}>
            {/* Chain separator (only between chains) */}
            {chainIdx > 0 && (
              <div className="mb-4 flex items-center gap-2">
                <div className="h-px flex-1 bg-border" />
                <span className="text-xs text-muted-foreground">different lineage</span>
                <div className="h-px flex-1 bg-border" />
              </div>
            )}

            {/* Timeline items in this chain */}
            <div>
              {chain.nodes.map((node, nodeIdx) => (
                <TimelineItem
                  key={node.snapshot.metadata.name}
                  node={node}
                  isFirst={nodeIdx === 0}
                  isLast={nodeIdx === chain.nodes.length - 1}
                  onRestore={onRestore}
                  onDelete={onDelete}
                  disabled={disabled}
                />
              ))}
            </div>
          </div>
        ))}
      </div>
    </div>
  )
}
