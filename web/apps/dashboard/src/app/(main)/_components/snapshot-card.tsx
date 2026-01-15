'use client'

import { useState } from 'react'
import {
  Clock,
  HardDrive,
  AlertCircle,
  Loader2,
  RotateCcw,
  Trash2,
  Cloud,
  CloudUpload,
  Star,
  ChevronDown,
  ChevronUp,
} from 'lucide-react'
import { Button, Badge } from '@kloudlite/ui'
import { cn } from '@/lib/utils'
import type { Snapshot } from '@/lib/services/snapshot.service'

interface SnapshotCardProps {
  snapshot: Snapshot
  isCurrent: boolean
  onRestore: () => void
  onDelete: () => void
  onPush?: () => void
  disabled?: boolean
  timelineData?: {
    lane: number
    activeLanes: Set<number>
    branchFrom?: { fromLane: number; toLane: number } | null
  }
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

// Muted color palette as per design
const BRANCH_COLORS = [
  { stroke: 'rgb(59, 130, 246)', fill: 'rgb(147, 197, 253)' }, // Blue
  { stroke: 'rgb(34, 197, 94)', fill: 'rgb(134, 239, 172)' }, // Green
  { stroke: 'rgb(168, 85, 247)', fill: 'rgb(216, 180, 254)' }, // Purple
  { stroke: 'rgb(249, 115, 22)', fill: 'rgb(253, 186, 116)' }, // Orange
  { stroke: 'rgb(236, 72, 153)', fill: 'rgb(244, 114, 182)' }, // Pink
  { stroke: 'rgb(14, 165, 233)', fill: 'rgb(125, 211, 252)' }, // Cyan
]

const LANE_WIDTH = 8 // Increased from 6
const DOT_RADIUS = 4 // Increased from 3
const GRAPH_LEFT_PADDING = 8

export function SnapshotCard({
  snapshot,
  isCurrent,
  onRestore,
  onDelete,
  onPush,
  disabled = false,
  timelineData,
}: SnapshotCardProps) {
  const [expanded, setExpanded] = useState(false)
  const [isDeleting, setIsDeleting] = useState(false)
  const [isRestoring, setIsRestoring] = useState(false)
  const [isPushing, setIsPushing] = useState(false)

  const shortHash = getShortHash(snapshot.name)
  const isPushed = !!snapshot.registry?.tag
  const isInProgress = [
    'Creating',
    'Uploading',
    'Restoring',
    'Pushing',
    'Pulling',
    'Deleting',
  ].includes(snapshot.state)
  const isFailed = snapshot.state === 'Failed'
  const isActionable =
    !isInProgress && !isFailed && snapshot.state !== 'Pending'
  const canDelete = snapshot.state !== 'Deleting' && !isCurrent

  // Calculate progress for in-progress operations (mock for now)
  const getProgress = () => {
    if (snapshot.state === 'Creating') return 45
    if (snapshot.state === 'Uploading') return 67
    if (snapshot.state === 'Restoring') return 80
    if (snapshot.state === 'Pushing') return 55
    return 0
  }

  const progress = getProgress()

  const handleRestore = async () => {
    setIsRestoring(true)
    try {
      await onRestore()
    } finally {
      setIsRestoring(false)
    }
  }

  const handleDelete = async () => {
    setIsDeleting(true)
    try {
      await onDelete()
    } finally {
      setIsDeleting(false)
    }
  }

  const handlePush = async () => {
    if (!onPush) return
    setIsPushing(true)
    try {
      await onPush()
    } finally {
      setIsPushing(false)
    }
  }

  // Calculate timeline graph dimensions
  const totalLanes = timelineData
    ? Math.max(...Array.from(timelineData.activeLanes)) + 1
    : 1
  const graphWidth = totalLanes * LANE_WIDTH + GRAPH_LEFT_PADDING * 2

  const laneColor = timelineData
    ? BRANCH_COLORS[timelineData.lane % BRANCH_COLORS.length]
    : BRANCH_COLORS[0]

  return (
    <div
      className={cn(
        'group relative flex rounded-lg border transition-all duration-200',
        // Base styling
        'bg-card shadow-sm',
        // Current snapshot styling
        isCurrent && 'ring-2 ring-blue-500/20 shadow-lg',
        // Failed snapshot styling
        isFailed && 'border-destructive/50 bg-destructive/5',
        // Hover effects (not for current or failed)
        !isCurrent &&
          !isFailed &&
          !isInProgress &&
          'hover:shadow-md hover:scale-[1.01]',
        // Active state
        'active:scale-[0.99]'
      )}
    >
      {/* Gradient border for current snapshot */}
      {isCurrent && (
        <div className="absolute inset-0 rounded-lg bg-gradient-to-br from-blue-500 to-cyan-500 opacity-20 pointer-events-none" />
      )}

      {/* Timeline Graph */}
      {timelineData && (
        <div
          className="relative flex-shrink-0 self-stretch"
          style={{ width: graphWidth }}
        >
          {/* Vertical lines for active lanes */}
          {Array.from(timelineData.activeLanes).map((laneIdx) => {
            const color = BRANCH_COLORS[laneIdx % BRANCH_COLORS.length]
            return (
              <div
                key={laneIdx}
                className="absolute top-0 bottom-0"
                style={{
                  left: GRAPH_LEFT_PADDING + laneIdx * LANE_WIDTH + 'px',
                  width: '2px',
                  backgroundColor: color.stroke,
                  opacity: 0.3,
                }}
              />
            )
          })}

          {/* Branch curve (if applicable) */}
          {timelineData.branchFrom && (
            <svg
              className="absolute top-0 left-0 w-full h-full pointer-events-none"
              style={{ overflow: 'visible' }}
            >
              <path
                d={`M ${
                  GRAPH_LEFT_PADDING +
                  timelineData.branchFrom.fromLane * LANE_WIDTH
                } 0 Q ${
                  GRAPH_LEFT_PADDING +
                  timelineData.branchFrom.fromLane * LANE_WIDTH
                } 20, ${
                  GRAPH_LEFT_PADDING +
                  timelineData.branchFrom.toLane * LANE_WIDTH
                } 40`}
                stroke={laneColor.stroke}
                strokeWidth="2"
                fill="none"
                opacity="0.5"
              />
            </svg>
          )}

          {/* Snapshot dot */}
          <div
            className="absolute top-1/2 -translate-y-1/2 rounded-full ring-2 ring-background transition-all"
            style={{
              left:
                GRAPH_LEFT_PADDING +
                (timelineData.lane * LANE_WIDTH - DOT_RADIUS) +
                'px',
              width: DOT_RADIUS * 2 + 'px',
              height: DOT_RADIUS * 2 + 'px',
              backgroundColor: laneColor.fill,
              boxShadow: isCurrent
                ? `0 0 0 3px ${laneColor.stroke}33`
                : undefined,
            }}
          />
        </div>
      )}

      {/* Card Content */}
      <div className="flex-1 p-4 space-y-3 min-w-0">
        {/* Header Row */}
        <div className="flex items-start justify-between gap-3">
          <div className="flex-1 min-w-0 space-y-1">
            {/* Title Line: Hash + Badges */}
            <div className="flex items-center gap-2 flex-wrap">
              <code className="text-sm font-mono font-medium">
                {shortHash}
              </code>

              {isCurrent && (
                <Badge variant="default" className="gap-1">
                  <Star className="h-3 w-3 fill-current" />
                  Current
                </Badge>
              )}

              {isPushed && (
                <Badge variant="secondary" className="gap-1">
                  <Cloud className="h-3 w-3" />
                  {snapshot.registry?.tag || 'pushed'}
                </Badge>
              )}

              {/* State Badge */}
              {isInProgress && (
                <Badge variant="default" className="gap-1">
                  <Loader2 className="h-3 w-3 animate-spin" />
                  {snapshot.state}
                </Badge>
              )}

              {isFailed && (
                <Badge variant="destructive" className="gap-1">
                  <AlertCircle className="h-3 w-3" />
                  Failed
                </Badge>
              )}
            </div>

            {/* Description */}
            {snapshot.description && (
              <p className="text-sm text-foreground line-clamp-2">
                {snapshot.description}
              </p>
            )}

            {/* Metadata Row */}
            <div className="flex items-center gap-4 text-xs text-muted-foreground">
              {snapshot.createdAt && (
                <span className="flex items-center gap-1">
                  <Clock className="h-3 w-3" />
                  {formatTimeAgo(snapshot.createdAt)}
                </span>
              )}

              {snapshot.sizeHuman && (
                <span className="flex items-center gap-1">
                  <HardDrive className="h-3 w-3" />
                  {snapshot.sizeHuman}
                </span>
              )}
            </div>
          </div>

          {/* Actions */}
          <div className="flex items-center gap-1">
            {isActionable && !isPushed && onPush && (
              <Button
                variant="ghost"
                size="sm"
                onClick={handlePush}
                disabled={disabled || isPushing}
                className="h-8 w-8 p-0"
              >
                {isPushing ? (
                  <Loader2 className="h-4 w-4 animate-spin" />
                ) : (
                  <CloudUpload className="h-4 w-4" />
                )}
              </Button>
            )}

            {isActionable && !isCurrent && (
              <Button
                variant="ghost"
                size="sm"
                onClick={handleRestore}
                disabled={disabled || isRestoring}
                className="h-8 w-8 p-0"
              >
                {isRestoring ? (
                  <Loader2 className="h-4 w-4 animate-spin" />
                ) : (
                  <RotateCcw className="h-4 w-4" />
                )}
              </Button>
            )}

            {canDelete && (
              <Button
                variant="ghost"
                size="sm"
                onClick={handleDelete}
                disabled={disabled || isDeleting}
                className="h-8 w-8 p-0 text-destructive hover:text-destructive hover:bg-destructive/10"
              >
                {isDeleting ? (
                  <Loader2 className="h-4 w-4 animate-spin" />
                ) : (
                  <Trash2 className="h-4 w-4" />
                )}
              </Button>
            )}

            {/* Expand button */}
            <Button
              variant="ghost"
              size="sm"
              onClick={() => setExpanded(!expanded)}
              className="h-8 w-8 p-0"
            >
              {expanded ? (
                <ChevronUp className="h-4 w-4" />
              ) : (
                <ChevronDown className="h-4 w-4" />
              )}
            </Button>
          </div>
        </div>

        {/* Progress Bar (for in-progress operations) */}
        {isInProgress && progress > 0 && (
          <div className="relative h-1.5 w-full overflow-hidden rounded-full bg-blue-100 dark:bg-blue-950">
            <div
              className="h-full bg-gradient-to-r from-blue-500 to-cyan-500 transition-all duration-300"
              style={{ width: `${progress}%` }}
            />
          </div>
        )}

        {/* Expanded Details */}
        {expanded && (
          <div className="pt-3 border-t space-y-2 text-xs text-muted-foreground animate-in fade-in-0 slide-in-from-top-1 duration-200">
            <div className="grid grid-cols-2 gap-2">
              <div>
                <span className="font-medium">Full Name:</span>
                <code className="ml-2 font-mono text-[10px]">
                  {snapshot.name}
                </code>
              </div>

              {snapshot.parent && (
                <div>
                  <span className="font-medium">Parent:</span>
                  <code className="ml-2 font-mono text-[10px]">
                    {getShortHash(snapshot.parent)}
                  </code>
                </div>
              )}

              {snapshot.sizeBytes && (
                <div>
                  <span className="font-medium">Size (bytes):</span>
                  <span className="ml-2">
                    {snapshot.sizeBytes.toLocaleString()}
                  </span>
                </div>
              )}

              {snapshot.registry?.digest && (
                <div className="col-span-2">
                  <span className="font-medium">Digest:</span>
                  <code className="ml-2 font-mono text-[10px] break-all">
                    {snapshot.registry.digest}
                  </code>
                </div>
              )}
            </div>

            {snapshot.message && isFailed && (
              <div className="p-2 rounded-md bg-destructive/10 border border-destructive/20">
                <span className="font-medium text-destructive">Error:</span>
                <p className="mt-1 text-destructive/80">{snapshot.message}</p>
              </div>
            )}
          </div>
        )}
      </div>
    </div>
  )
}
