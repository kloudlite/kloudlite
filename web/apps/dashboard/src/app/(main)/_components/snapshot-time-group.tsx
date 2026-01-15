'use client'

import {
  Clock,
  HardDrive,
  Loader2,
  MoreHorizontal,
  RotateCcw,
  Trash2,
  CloudUpload,
  Cloud,
  AlertCircle,
  Star,
} from 'lucide-react'
import { Button, Badge } from '@kloudlite/ui'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
  DropdownMenuSeparator,
} from '@kloudlite/ui'
import type { Snapshot } from '@/lib/services/snapshot.service'

interface SnapshotTimeGroupProps {
  label: string
  snapshots: Snapshot[]
  currentSnapshotName?: string
  onRestore: (snapshot: Snapshot) => void
  onDelete: (snapshot: Snapshot) => void
  onPush?: (snapshot: Snapshot) => void
  disabled?: boolean
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

export function SnapshotTimeGroup({
  label,
  snapshots,
  currentSnapshotName,
  onRestore,
  onDelete,
  onPush,
  disabled = false,
}: SnapshotTimeGroupProps) {
  // Sort by creation time (newest first)
  const sorted = [...snapshots].sort((a, b) => {
    const aTime = new Date(a.createdAt || '').getTime()
    const bTime = new Date(b.createdAt || '').getTime()
    return bTime - aTime
  })

  return (
    <div className="space-y-3">
      {/* Time Group Label */}
      <h3 className="text-xs font-medium uppercase tracking-wider text-muted-foreground">
        {label}
      </h3>

      {/* Cards */}
      <div className="space-y-2">
        {sorted.map((snapshot) => {
          const shortHash = getShortHash(snapshot.name)
          const isPushed = !!snapshot.registry?.digest
          const isCurrent = snapshot.name === currentSnapshotName
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

          return (
            <div
              key={snapshot.name}
              className={`bg-card border rounded-lg p-4 ${
                isCurrent ? 'border-blue-500' : ''
              } ${!isCurrent && !isInProgress ? 'hover:bg-muted/50' : ''}`}
            >
              <div className="flex items-start justify-between gap-4">
                {/* Left: Info */}
                <div className="flex-1 min-w-0 space-y-2">
                  {/* Header: Hash + Badges */}
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
                    {!isInProgress && !isFailed && !isCurrent && (
                      <Badge variant="secondary">Ready</Badge>
                    )}
                  </div>

                  {/* Description */}
                  {snapshot.description && (
                    <p className="text-sm text-foreground">
                      {snapshot.description}
                    </p>
                  )}

                  {/* Metadata */}
                  <div className="flex items-center gap-4 text-xs text-muted-foreground">
                    {snapshot.createdAt && (
                      <div className="flex items-center gap-1">
                        <Clock className="h-3 w-3" />
                        <span>{formatTimeAgo(snapshot.createdAt)}</span>
                      </div>
                    )}
                    {snapshot.sizeHuman && snapshot.sizeHuman !== '0 B' && (
                      <div className="flex items-center gap-1">
                        <HardDrive className="h-3 w-3" />
                        <span>{snapshot.sizeHuman}</span>
                      </div>
                    )}
                  </div>
                </div>

                {/* Right: Actions */}
                <div>
                  {isInProgress || isFailed ? (
                    <Button variant="ghost" size="sm" className="h-8 w-8 p-0" disabled>
                      <MoreHorizontal className="h-4 w-4 opacity-30" />
                    </Button>
                  ) : (
                    <DropdownMenu>
                      <DropdownMenuTrigger asChild>
                        <Button variant="ghost" size="sm" className="h-8 w-8 p-0" disabled={disabled}>
                          <MoreHorizontal className="h-4 w-4" />
                        </Button>
                      </DropdownMenuTrigger>
                      <DropdownMenuContent align="end">
                        {isActionable && !isCurrent && (
                          <DropdownMenuItem onClick={() => onRestore(snapshot)}>
                            <RotateCcw className="mr-2 h-4 w-4" />
                            Restore
                          </DropdownMenuItem>
                        )}
                        {isActionable && !isPushed && onPush && (
                          <DropdownMenuItem onClick={() => onPush(snapshot)}>
                            <CloudUpload className="mr-2 h-4 w-4" />
                            Push to Registry
                          </DropdownMenuItem>
                        )}
                        {(isActionable || isFailed) && !isCurrent && (
                          <>
                            <DropdownMenuSeparator />
                            <DropdownMenuItem
                              className="text-red-600"
                              onClick={() => onDelete(snapshot)}
                            >
                              <Trash2 className="mr-2 h-4 w-4" />
                              Delete
                            </DropdownMenuItem>
                          </>
                        )}
                      </DropdownMenuContent>
                    </DropdownMenu>
                  )}
                </div>
              </div>
            </div>
          )
        })}
      </div>
    </div>
  )
}
