'use client'

import { useMemo } from 'react'
import {
  Clock,
  HardDrive,
  AlertCircle,
  Loader2,
  RotateCcw,
  Trash2,
  History,
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
      return (
        <Badge variant="outline" className="bg-emerald-50 text-emerald-700 border-emerald-200 dark:bg-emerald-900/30 dark:text-emerald-400 dark:border-emerald-800">
          Ready
        </Badge>
      )
    case 'Creating':
      return (
        <Badge variant="outline" className="bg-blue-50 text-blue-700 border-blue-200 dark:bg-blue-900/30 dark:text-blue-400 dark:border-blue-800">
          <Loader2 className="h-3 w-3 animate-spin mr-1" />
          Creating
        </Badge>
      )
    case 'Restoring':
      return (
        <Badge variant="outline" className="bg-amber-50 text-amber-700 border-amber-200 dark:bg-amber-900/30 dark:text-amber-400 dark:border-amber-800">
          <Loader2 className="h-3 w-3 animate-spin mr-1" />
          Restoring
        </Badge>
      )
    case 'Deleting':
      return (
        <Badge variant="outline" className="bg-orange-50 text-orange-700 border-orange-200 dark:bg-orange-900/30 dark:text-orange-400 dark:border-orange-800">
          <Loader2 className="h-3 w-3 animate-spin mr-1" />
          Deleting
        </Badge>
      )
    case 'Failed':
      return (
        <Badge variant="destructive">
          <AlertCircle className="h-3 w-3 mr-1" />
          Failed
        </Badge>
      )
    case 'Pending':
    default:
      return (
        <Badge variant="secondary">
          Pending
        </Badge>
      )
  }
}

// Build tree structure from snapshots
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

  // Find root nodes
  const rootSnapshots = snapshots.filter(s => {
    const parentName = s.spec.parentSnapshotRef?.name
    return !parentName || !snapshotMap.has(parentName)
  })

  // Sort roots by creation time (newest first)
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

interface SnapshotCardProps {
  snapshot: Snapshot
  isCurrent: boolean
  onRestore: (snapshot: Snapshot) => void
  onDelete: (snapshot: Snapshot) => void
  disabled?: boolean
}

function SnapshotCard({ snapshot, isCurrent, onRestore, onDelete, disabled }: SnapshotCardProps) {
  const shortHash = getShortHash(snapshot.metadata.name)

  return (
    <div
      className={cn(
        "group rounded-lg border p-3 transition-all",
        isCurrent
          ? "bg-blue-50/50 border-blue-200 dark:bg-blue-950/20 dark:border-blue-900"
          : "bg-card hover:bg-muted/30 hover:border-muted-foreground/20"
      )}
    >
      <div className="flex items-center justify-between gap-2 mb-1.5">
        <div className="flex items-center gap-2 min-w-0 flex-wrap">
          {isCurrent && (
            <Badge className="bg-blue-500 hover:bg-blue-500 text-white">
              Current
            </Badge>
          )}
          <code className="text-xs font-mono text-muted-foreground">
            {shortHash}
          </code>
          {getStateBadge(snapshot.status.state)}
        </div>

        <div className="flex items-center gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
          {snapshot.status.state === 'Ready' && !isCurrent && (
            <Button
              variant="ghost"
              size="sm"
              onClick={() => onRestore(snapshot)}
              disabled={disabled}
              className="h-7 px-2 text-xs"
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
              className="h-7 w-7 p-0 text-muted-foreground hover:text-destructive"
            >
              <Trash2 className="h-3.5 w-3.5" />
            </Button>
          )}
        </div>
      </div>

      {snapshot.spec.description && (
        <p className="text-sm text-foreground mb-1.5 line-clamp-2">
          {snapshot.spec.description}
        </p>
      )}

      <div className="flex items-center gap-3 text-xs text-muted-foreground">
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
        <p className="mt-2 text-xs text-red-600 dark:text-red-400 bg-red-50 dark:bg-red-950/30 rounded px-2 py-1.5">
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
}

function TreeNodeRenderer({ node, onRestore, onDelete, disabled, depth, isLast }: TreeNodeRendererProps) {
  const hasChildren = node.children.length > 0

  return (
    <div className="relative">
      {/* The snapshot card with connector */}
      <div className="flex items-stretch">
        {/* Left connector area */}
        {depth > 0 && (
          <div className="w-6 flex-shrink-0 relative">
            {/* Vertical line from above */}
            <div className={cn(
              "absolute left-2 w-px bg-border",
              isLast ? "top-0 h-5" : "top-0 bottom-0"
            )} />
            {/* Horizontal connector */}
            <div className="absolute left-2 top-5 w-4 h-px bg-border" />
          </div>
        )}

        {/* Card */}
        <div className="flex-1 pb-2">
          <SnapshotCard
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
        <div className={cn("relative", depth > 0 ? "ml-6" : "ml-0")}>
          {/* Vertical line connecting children */}
          {depth === 0 && (
            <div className="absolute left-2 top-0 bottom-2 w-px bg-border" />
          )}

          {node.children.map((child, idx) => (
            <TreeNodeRenderer
              key={child.snapshot.metadata.name}
              node={child}
              onRestore={onRestore}
              onDelete={onDelete}
              disabled={disabled}
              depth={depth + 1}
              isLast={idx === node.children.length - 1}
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
      <div className="flex items-center gap-2 mb-3">
        <History className="h-4 w-4 text-muted-foreground" />
        <h4 className="text-sm font-medium">History</h4>
        <span className="text-xs text-muted-foreground">({snapshots.length})</span>
      </div>

      <div className="space-y-2">
        {trees.map((tree, idx) => (
          <TreeNodeRenderer
            key={tree.snapshot.metadata.name}
            node={tree}
            onRestore={onRestore}
            onDelete={onDelete}
            disabled={disabled}
            depth={0}
            isLast={idx === trees.length - 1}
          />
        ))}
      </div>
    </div>
  )
}
