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

// Extract short hash from snapshot name
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

// Build tree structure from snapshots based on parent references
function buildTree(snapshots: Snapshot[], currentSnapshotName?: string): TreeNode[] {
  if (snapshots.length === 0) return []

  const snapshotMap = new Map<string, Snapshot>()
  const childrenMap = new Map<string, string[]>() // parentName -> childNames[]

  snapshots.forEach(s => {
    snapshotMap.set(s.metadata.name, s)
    const parentName = s.spec.parentSnapshotRef?.name
    if (parentName) {
      const existing = childrenMap.get(parentName) || []
      existing.push(s.metadata.name)
      childrenMap.set(parentName, existing)
    }
  })

  // Find root nodes (no parent or parent not in our list)
  const rootSnapshots = snapshots.filter(s => {
    const parentName = s.spec.parentSnapshotRef?.name
    return !parentName || !snapshotMap.has(parentName)
  })

  // Sort roots by creation time (oldest first - roots at bottom)
  rootSnapshots.sort((a, b) =>
    new Date(a.status.createdAt || a.metadata.creationTimestamp).getTime() -
    new Date(b.status.createdAt || b.metadata.creationTimestamp).getTime()
  )

  // Recursively build tree
  function buildNode(snapshot: Snapshot): TreeNode {
    const childNames = childrenMap.get(snapshot.metadata.name) || []
    const childSnapshots = childNames
      .map(name => snapshotMap.get(name)!)
      .filter(Boolean)
      // Sort children by creation time (newest first)
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

// Flatten tree to render order (DFS, children before parent for bottom-up view)
interface FlatNode {
  snapshot: Snapshot
  isCurrent: boolean
  depth: number
  isLastChild: boolean
  ancestorIsLast: boolean[] // Track which ancestors are last children (for line drawing)
}

function flattenTree(roots: TreeNode[]): FlatNode[] {
  const result: FlatNode[] = []

  function traverse(node: TreeNode, depth: number, isLastChild: boolean, ancestorIsLast: boolean[]) {
    // First render children (newest first, so they appear at top)
    node.children.forEach((child, idx) => {
      traverse(child, depth + 1, idx === node.children.length - 1, [...ancestorIsLast, isLastChild])
    })

    // Then render self
    result.push({
      snapshot: node.snapshot,
      isCurrent: node.isCurrent,
      depth,
      isLastChild,
      ancestorIsLast,
    })
  }

  // Process roots (oldest at bottom)
  roots.forEach((root, idx) => {
    traverse(root, 0, idx === roots.length - 1, [])
  })

  return result
}

interface TreeNodeItemProps {
  node: FlatNode
  onRestore: (snapshot: Snapshot) => void
  onDelete: (snapshot: Snapshot) => void
  disabled?: boolean
  isFirst: boolean
  isLast: boolean
  totalNodes: number
  nodeIndex: number
}

function TreeNodeItem({ node, onRestore, onDelete, disabled, isFirst, isLast }: TreeNodeItemProps) {
  const { snapshot, isCurrent, depth } = node
  const shortHash = getShortHash(snapshot.metadata.name)
  const indentWidth = depth * 24 // 24px per level

  return (
    <div className="relative flex">
      {/* Indent spacer */}
      {depth > 0 && (
        <div style={{ width: indentWidth }} className="flex-shrink-0 relative">
          {/* Vertical lines for each ancestor level */}
          {node.ancestorIsLast.map((isLast, idx) => (
            !isLast && (
              <div
                key={idx}
                className="absolute top-0 bottom-0 w-px bg-border"
                style={{ left: idx * 24 + 10 }}
              />
            )
          ))}
          {/* Horizontal connector to this node */}
          <div
            className="absolute w-3 h-px bg-border"
            style={{
              left: (depth - 1) * 24 + 10,
              top: '50%',
            }}
          />
          {/* Vertical line segment */}
          <div
            className={cn(
              "absolute w-px bg-border",
              node.isLastChild ? "top-0 h-1/2" : "top-0 bottom-0"
            )}
            style={{ left: (depth - 1) * 24 + 10 }}
          />
        </div>
      )}

      {/* Timeline dot column */}
      <div className="relative flex flex-col items-center w-5 flex-shrink-0">
        {/* Line above */}
        {(!isFirst && depth === 0) && <div className="w-px flex-1 bg-border" />}
        {(isFirst || depth > 0) && <div className="flex-1" />}

        {/* Dot */}
        <div
          className={cn(
            "relative z-10 rounded-full flex-shrink-0 transition-all",
            isCurrent
              ? "w-3 h-3 bg-blue-500 ring-[3px] ring-blue-500/20"
              : "w-2.5 h-2.5 bg-gray-300 dark:bg-gray-600"
          )}
        />

        {/* Line below */}
        {(!isLast && depth === 0) && <div className="w-px flex-1 bg-border" />}
        {(isLast || depth > 0) && <div className="flex-1" />}
      </div>

      {/* Content */}
      <div className="flex-1 pb-3 min-w-0 ml-2">
        <div
          className={cn(
            "group rounded-lg border p-3 transition-all",
            isCurrent
              ? "bg-blue-50/50 border-blue-200 dark:bg-blue-950/20 dark:border-blue-900"
              : "bg-card hover:bg-muted/30 hover:border-muted-foreground/20"
          )}
        >
          {/* Header row */}
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

            {/* Actions */}
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

          {/* Description */}
          {snapshot.spec.description && (
            <p className="text-sm text-foreground mb-1.5 line-clamp-2">
              {snapshot.spec.description}
            </p>
          )}

          {/* Meta row */}
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

          {/* Error message */}
          {snapshot.status.state === 'Failed' && snapshot.status.message && (
            <p className="mt-2 text-xs text-red-600 dark:text-red-400 bg-red-50 dark:bg-red-950/30 rounded px-2 py-1.5">
              {snapshot.status.message}
            </p>
          )}
        </div>
      </div>
    </div>
  )
}

export function SnapshotTimeline({ snapshots, onRestore, onDelete, disabled, currentSnapshotName }: SnapshotTimelineProps) {
  const flatNodes = useMemo(() => {
    const tree = buildTree(snapshots, currentSnapshotName)
    return flattenTree(tree)
  }, [snapshots, currentSnapshotName])

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

      <div>
        {flatNodes.map((node, idx) => (
          <TreeNodeItem
            key={node.snapshot.metadata.name}
            node={node}
            isFirst={idx === 0}
            isLast={idx === flatNodes.length - 1}
            onRestore={onRestore}
            onDelete={onDelete}
            disabled={disabled}
            totalNodes={flatNodes.length}
            nodeIndex={idx}
          />
        ))}
      </div>
    </div>
  )
}
