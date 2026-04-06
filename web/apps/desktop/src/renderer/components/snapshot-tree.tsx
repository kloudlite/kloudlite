import { cn } from '@/lib/utils'
import { Plus, RotateCcw, Trash2 } from 'lucide-react'

// Snapshots — tree structure (restore + new snapshot = branch)
export interface Snapshot {
  id: string
  name: string
  description: string
  author: string
  date: string
  size: string
  parentId: string | null
  isHead?: boolean
}

interface TreeNode {
  snapshot: Snapshot
  children: TreeNode[]
  depth: number
  isOnHeadPath: boolean
}

interface FlatItem {
  node: TreeNode
  col: number
  showFork: boolean
  forkFromCol: number
}

function buildTree(snapshots: Snapshot[]): TreeNode | null {
  const map = new Map<string, TreeNode>()
  const headId = snapshots.find((s) => s.isHead)?.id

  const headPath = new Set<string>()
  if (headId) {
    let current = headId
    while (current) {
      headPath.add(current)
      const snap = snapshots.find((s) => s.id === current)
      current = snap?.parentId ?? ''
    }
  }

  for (const snap of snapshots) {
    map.set(snap.id, { snapshot: snap, children: [], depth: 0, isOnHeadPath: headPath.has(snap.id) })
  }

  let root: TreeNode | null = null
  for (const snap of snapshots) {
    const node = map.get(snap.id)!
    if (snap.parentId && map.has(snap.parentId)) {
      const parent = map.get(snap.parentId)!
      parent.children.push(node)
      node.depth = parent.depth + 1
    } else {
      root = node
    }
  }
  return root
}

function flattenTree(root: TreeNode): FlatItem[] {
  const result: FlatItem[] = []
  let nextCol = 1

  function walk(n: TreeNode, col: number, forkFromCol: number, showFork: boolean) {
    result.push({ node: n, col, showFork, forkFromCol })

    const sorted = [...n.children].sort((a, b) => {
      if (a.isOnHeadPath && !b.isOnHeadPath) return -1
      if (!a.isOnHeadPath && b.isOnHeadPath) return 1
      return 0
    })

    sorted.forEach((child, i) => {
      if (i === 0) {
        walk(child, col, col, false)
      } else {
        const forkCol = nextCol++
        walk(child, forkCol, col, true)
      }
    })
  }

  walk(root, 0, 0, false)
  result.reverse()
  for (const item of result) {
    if (item.showFork) {
      const parentIdx = result.findIndex((f) => f.node.snapshot.id === item.node.snapshot.parentId)
      if (parentIdx >= 0) {
        item.forkFromCol = result[parentIdx].col
      }
    }
  }
  return result
}

// Demo data generator
export function generateSnapshots(seed: string): Snapshot[] {
  const hash = (s: string) => {
    let h = 0
    for (let i = 0; i < s.length; i++) h = ((h << 5) - h + s.charCodeAt(i)) | 0
    return Math.abs(h)
  }
  const r = hash(seed)

  const mainSteps = [
    { name: 'Created', desc: 'Empty', size: '0.2 MB' },
    { name: 'Initial setup', desc: '2 packages', size: '8.1 MB' },
    { name: 'Add dependencies', desc: '5 packages', size: '10.2 MB' },
    { name: 'Configure services', desc: '3 services', size: '11.0 MB' },
    { name: 'Production ready', desc: '4 services, 2 configs', size: '12.4 MB' },
  ]

  const branchSteps = [
    ['Try alternative setup', 'Alternative tuning'],
    ['Different strategy', 'Strategy validation'],
    ['Experiment'],
  ]

  const authors = ['karthik', 'sohail']
  const times = ['2 weeks ago', '10 days ago', '1 week ago', '5 days ago', '3 days ago', '1 day ago', '2 hours ago']

  const count = 3 + (r % 3)
  const snapshots: Snapshot[] = []

  for (let i = 0; i < count && i < mainSteps.length; i++) {
    snapshots.push({
      id: `s${i + 1}`,
      name: mainSteps[i].name,
      description: mainSteps[i].desc,
      author: authors[i % 2],
      date: times[i] || `${i} days ago`,
      size: mainSteps[i].size,
      parentId: i === 0 ? null : `s${i}`,
      isHead: i === count - 1,
    })
  }

  const branchCount = (r % 3)
  for (let b = 0; b < branchCount && b < branchSteps.length; b++) {
    const forkPoint = 1 + ((r + b * 7) % Math.max(count - 2, 1))
    const steps = branchSteps[b]
    for (let j = 0; j < steps.length; j++) {
      snapshots.push({
        id: `b${b + 1}-${j + 1}`,
        name: steps[j],
        description: `${2 + j} packages`,
        author: authors[(b + j) % 2],
        date: `${3 + b * 2 + j} days ago`,
        size: `${(8 + b + j * 0.8).toFixed(1)} MB`,
        parentId: j === 0 ? `s${forkPoint + 1}` : `b${b + 1}-${j}`,
      })
    }
  }

  return snapshots
}

const ROW_H = 80
const COL_W = 28
const DOT_R = 6

interface SnapshotTreeProps {
  snapshots: Snapshot[]
  title: string
  subtitle: string
}

export function SnapshotTree({ snapshots, title, subtitle }: SnapshotTreeProps) {
  const tree = buildTree(snapshots)
  const flat = tree ? flattenTree(tree) : []
  const maxCol = Math.max(...flat.map((f) => f.col), 0)
  const graphWidth = (maxCol + 1) * COL_W + 8
  const totalHeight = flat.length * ROW_H

  const lines: { x1: number; y1: number; x2: number; y2: number; head: boolean }[] = []
  for (let i = 0; i < flat.length; i++) {
    const item = flat[i]
    const parentId = item.node.snapshot.parentId
    if (!parentId) continue

    const parentIdx = flat.findIndex((f) => f.node.snapshot.id === parentId)
    if (parentIdx < 0) continue

    const parentItem = flat[parentIdx]
    const nx = item.col * COL_W + DOT_R
    const ny = i * ROW_H + ROW_H / 2
    const px = parentItem.col * COL_W + DOT_R
    const py = parentIdx * ROW_H + ROW_H / 2

    if (item.col === parentItem.col) {
      lines.push({ x1: nx, y1: ny, x2: px, y2: py, head: item.node.isOnHeadPath })
    } else {
      lines.push({ x1: nx, y1: ny, x2: px, y2: ny, head: false })
      lines.push({ x1: px, y1: ny, x2: px, y2: py, head: false })
    }
  }

  return (
    <div className="p-6">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-[16px] font-semibold text-foreground">{title}</h2>
          <p className="mt-1 text-[13px] text-muted-foreground">{subtitle}</p>
        </div>
        <button className="flex items-center gap-1.5 rounded-lg bg-primary px-3 py-1.5 text-[12px] font-medium text-primary-foreground transition-colors hover:bg-primary/90">
          <Plus className="h-3.5 w-3.5" />
          Take Snapshot
        </button>
      </div>

      <div className="relative mt-6 flex">
        <div className="relative shrink-0" style={{ width: graphWidth }}>
          <svg width={graphWidth} height={totalHeight} className="absolute left-0 top-0">
            {lines.map((line, i) => (
              <line
                key={i}
                x1={line.x1} y1={line.y1} x2={line.x2} y2={line.y2}
                stroke={line.head ? 'var(--primary)' : 'var(--border)'}
                strokeWidth={2}
                opacity={line.head ? 0.5 : 0.4}
              />
            ))}
            {flat.map((item, i) => {
              const cx = item.col * COL_W + DOT_R
              const cy = i * ROW_H + ROW_H / 2
              const snap = item.node.snapshot
              if (snap.isHead) {
                return (
                  <g key={snap.id}>
                    <circle cx={cx} cy={cy} r={DOT_R + 3} fill="var(--primary)" opacity={0.15} />
                    <circle cx={cx} cy={cy} r={DOT_R} fill="var(--primary)" />
                  </g>
                )
              }
              if (item.node.isOnHeadPath) {
                return <circle key={snap.id} cx={cx} cy={cy} r={DOT_R - 1} fill="var(--primary)" opacity={0.6} />
              }
              return (
                <g key={snap.id}>
                  <circle cx={cx} cy={cy} r={DOT_R - 1} fill="var(--background)" stroke="var(--border)" strokeWidth={2} />
                </g>
              )
            })}
          </svg>
        </div>

        <div className="flex-1">
          {flat.map((item) => {
            const snap = item.node.snapshot
            const n = item.node
            return (
              <div key={snap.id} className="flex items-center" style={{ height: ROW_H }}>
                <div className={cn(
                  'flex-1 rounded-xl border px-4 py-2.5 transition-colors',
                  snap.isHead
                    ? 'border-primary/30 bg-primary/[0.03]'
                    : n.isOnHeadPath
                      ? 'border-primary/15 hover:bg-accent/20'
                      : 'border-border/40 hover:bg-accent/20'
                )}>
                  <div className="flex items-start justify-between gap-3">
                    <div className="min-w-0 flex-1">
                      <div className="flex items-center gap-2 flex-wrap">
                        <p className="text-[13px] font-medium text-foreground">{snap.name}</p>
                        {snap.isHead && (
                          <span className="shrink-0 rounded-full bg-primary/10 px-2 py-0.5 text-[10px] font-semibold text-primary">HEAD</span>
                        )}
                        {!n.isOnHeadPath && (
                          <span className="shrink-0 rounded-full bg-accent px-2 py-0.5 text-[10px] font-medium text-muted-foreground">branch</span>
                        )}
                      </div>
                      <div className="mt-1 flex items-center gap-2 text-[11px] text-muted-foreground">
                        <span>{snap.author}</span>
                        <span>·</span>
                        <span>{snap.date}</span>
                        <span>·</span>
                        <span>{snap.size}</span>
                      </div>
                    </div>
                    {!snap.isHead && (
                      <div className="flex shrink-0 items-center gap-1.5">
                        <button className="flex items-center gap-1 rounded-lg border border-border px-2.5 py-1 text-[11px] font-medium text-muted-foreground transition-colors hover:bg-accent hover:text-foreground">
                          <RotateCcw className="h-3 w-3" />
                          Restore
                        </button>
                        <button className="flex items-center rounded-lg border border-border p-1 text-muted-foreground/40 transition-colors hover:border-red-500/30 hover:bg-red-500/10 hover:text-red-500">
                          <Trash2 className="h-3.5 w-3.5" />
                        </button>
                      </div>
                    )}
                  </div>
                </div>
              </div>
            )
          })}
        </div>
      </div>
    </div>
  )
}
