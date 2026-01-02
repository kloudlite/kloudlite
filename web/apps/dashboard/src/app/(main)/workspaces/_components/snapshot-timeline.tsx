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

interface TreeNode {
  snapshot: Snapshot
  children: TreeNode[]
  isCurrent: boolean
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

function buildTree(snapshots: Snapshot[], currentSnapshotName?: string): TreeNode[] {
  if (snapshots.length === 0) return []

  const snapshotMap = new Map<string, Snapshot>()
  const childrenMap = new Map<string, string[]>()

  snapshots.forEach(s => {
    snapshotMap.set(s.metadata.name, s)
    const parentName = s.spec.parentSnapshotRef?.name
    if (parentName) {
      const existing = childrenMap.get(parentName) || []
      existing.push(s.metadata.name)
      childrenMap.set(parentName, existing)
    }
  })

  const rootSnapshots = snapshots.filter(s => {
    const parentName = s.spec.parentSnapshotRef?.name
    return !parentName || !snapshotMap.has(parentName)
  })

  rootSnapshots.sort((a, b) =>
    new Date(b.status.createdAt || b.metadata.creationTimestamp).getTime() -
    new Date(a.status.createdAt || a.metadata.creationTimestamp).getTime()
  )

  function buildNode(snapshot: Snapshot): TreeNode {
    const childNames = childrenMap.get(snapshot.metadata.name) || []
    const childSnapshots = childNames
      .map(name => snapshotMap.get(name)!)
      .filter(Boolean)
      .sort((a, b) =>
        new Date(b.status.createdAt || b.metadata.creationTimestamp).getTime() -
        new Date(a.status.createdAt || a.metadata.creationTimestamp).getTime()
      )

    return {
      snapshot,
      children: childSnapshots.map(buildNode),
      isCurrent: snapshot.metadata.name === currentSnapshotName,
    }
  }

  return rootSnapshots.map(buildNode)
}

interface SnapshotItemProps {
  snapshot: Snapshot
  isCurrent: boolean
  onRestore: (snapshot: Snapshot) => void
  onDelete: (snapshot: Snapshot) => void
  disabled?: boolean
}

function SnapshotItem({ snapshot, isCurrent, onRestore, onDelete, disabled }: SnapshotItemProps) {
  const shortHash = getShortHash(snapshot.metadata.name)

  return (
    <div
      className={cn(
        "group py-2.5 px-3 rounded-lg border transition-colors",
        isCurrent
          ? "bg-blue-50 border-blue-200 dark:bg-blue-950/30 dark:border-blue-800"
          : "border-border hover:bg-muted/50 hover:border-muted-foreground/30"
      )}
    >
      <div className="flex items-center justify-between gap-3">
        <div className="flex items-center gap-3 min-w-0">
          {/* Hash */}
          <code className={cn(
            "text-sm font-mono",
            isCurrent ? "text-blue-600 dark:text-blue-400 font-semibold" : "text-muted-foreground"
          )}>
            {shortHash}
          </code>

          {/* Current badge */}
          {isCurrent && (
            <Badge variant="secondary" className="text-[10px] px-1.5 py-0 h-4 bg-blue-100 text-blue-700 dark:bg-blue-900 dark:text-blue-300">
              HEAD
            </Badge>
          )}

          {/* State badge */}
          {getStateBadge(snapshot.status.state)}
        </div>

        {/* Actions */}
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

      {/* Description */}
      {snapshot.spec.description && (
        <p className="text-sm text-foreground mt-1 ml-0">
          {snapshot.spec.description}
        </p>
      )}

      {/* Meta */}
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

      {/* Error */}
      {snapshot.status.state === 'Failed' && snapshot.status.message && (
        <p className="mt-1.5 text-xs text-red-600 dark:text-red-400">
          {snapshot.status.message}
        </p>
      )}
    </div>
  )
}

interface TreeNodeRendererProps {
  node: TreeNode
  onRestore: (snapshot: Snapshot) => void
  onDelete: (snapshot: Snapshot) => void
  disabled?: boolean
  depth: number
  isLast: boolean
  parentHasMore: boolean[]
}

function TreeNodeRenderer({ node, onRestore, onDelete, disabled, depth, isLast, parentHasMore }: TreeNodeRendererProps) {
  const hasChildren = node.children.length > 0

  return (
    <div className="relative">
      <div className="flex">
        {/* Tree lines */}
        {depth > 0 && (
          <div className="flex-shrink-0 relative" style={{ width: depth * 20 }}>
            {/* Vertical lines for ancestors */}
            {parentHasMore.map((hasMore, idx) => (
              hasMore && (
                <div
                  key={idx}
                  className="absolute top-0 bottom-0 w-0.5 bg-gray-300 dark:bg-gray-600"
                  style={{ left: idx * 20 + 8 }}
                />
              )
            ))}
            {/* Elbow connector */}
            <div
              className="absolute w-0.5 bg-gray-300 dark:bg-gray-600"
              style={{
                left: (depth - 1) * 20 + 8,
                top: 0,
                height: isLast ? 20 : '100%',
              }}
            />
            <div
              className="absolute h-0.5 bg-gray-300 dark:bg-gray-600"
              style={{
                left: (depth - 1) * 20 + 8,
                top: 20,
                width: 12,
              }}
            />
          </div>
        )}

        {/* Snapshot item */}
        <div className="flex-1 min-w-0 mb-2">
          <SnapshotItem
            snapshot={node.snapshot}
            isCurrent={node.isCurrent}
            onRestore={onRestore}
            onDelete={onDelete}
            disabled={disabled}
          />
        </div>
      </div>

      {/* Children */}
      {hasChildren && (
        <div className="relative">
          {node.children.map((child, idx) => (
            <TreeNodeRenderer
              key={child.snapshot.metadata.name}
              node={child}
              onRestore={onRestore}
              onDelete={onDelete}
              disabled={disabled}
              depth={depth + 1}
              isLast={idx === node.children.length - 1}
              parentHasMore={[...parentHasMore, !isLast]}
            />
          ))}
        </div>
      )}
    </div>
  )
}

export function SnapshotTimeline({ snapshots, onRestore, onDelete, disabled, currentSnapshotName }: SnapshotTimelineProps) {
  const trees = useMemo(() => buildTree(snapshots, currentSnapshotName), [snapshots, currentSnapshotName])

  if (snapshots.length === 0) {
    return null
  }

  return (
    <div>
      <div className="flex items-center justify-between mb-2">
        <span className="text-xs font-medium text-muted-foreground uppercase tracking-wider">
          Snapshots
        </span>
        <span className="text-xs text-muted-foreground">{snapshots.length}</span>
      </div>

      <div className="space-y-0">
        {trees.map((tree, idx) => (
          <TreeNodeRenderer
            key={tree.snapshot.metadata.name}
            node={tree}
            onRestore={onRestore}
            onDelete={onDelete}
            disabled={disabled}
            depth={0}
            isLast={idx === trees.length - 1}
            parentHasMore={[]}
          />
        ))}
      </div>
    </div>
  )
}
