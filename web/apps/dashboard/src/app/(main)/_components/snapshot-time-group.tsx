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

      {/* Table */}
      <div className="bg-card overflow-hidden rounded-lg border">
        <table className="min-w-full">
          <thead className="bg-muted/50 border-b">
            <tr>
              <th className="text-muted-foreground px-6 py-3 text-left text-xs font-medium tracking-wider uppercase">
                Name
              </th>
              <th className="text-muted-foreground px-6 py-3 text-left text-xs font-medium tracking-wider uppercase">
                Description
              </th>
              <th className="text-muted-foreground px-6 py-3 text-left text-xs font-medium tracking-wider uppercase">
                Created
              </th>
              <th className="text-muted-foreground px-6 py-3 text-left text-xs font-medium tracking-wider uppercase">
                Size
              </th>
              <th className="text-muted-foreground px-6 py-3 text-left text-xs font-medium tracking-wider uppercase">
                Status
              </th>
              <th className="text-muted-foreground px-6 py-3 text-right text-xs font-medium tracking-wider uppercase">
                Actions
              </th>
            </tr>
          </thead>
          <tbody className="divide-y">
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
                <tr
                  key={snapshot.name}
                  className={isCurrent ? 'bg-blue-50/50 dark:bg-blue-950/20' : 'hover:bg-muted/50'}
                >
                  {/* Name/Hash */}
                  <td className="px-6 py-4 whitespace-nowrap">
                    <div className="flex items-center gap-2">
                      <code className="text-sm font-mono font-medium">
                        {shortHash}
                      </code>
                      {isCurrent && (
                        <Badge variant="default" className="gap-1 text-[10px] px-1.5 py-0">
                          <Star className="h-3 w-3 fill-current" />
                          Current
                        </Badge>
                      )}
                      {isPushed && (
                        <Badge variant="secondary" className="gap-1 text-[10px] px-1.5 py-0">
                          <Cloud className="h-3 w-3" />
                          {snapshot.registry?.tag || 'pushed'}
                        </Badge>
                      )}
                    </div>
                  </td>

                  {/* Description */}
                  <td className="px-6 py-4 max-w-xs">
                    {snapshot.description ? (
                      <span className="text-sm text-foreground truncate block" title={snapshot.description}>
                        {snapshot.description}
                      </span>
                    ) : (
                      <span className="text-muted-foreground text-sm">-</span>
                    )}
                  </td>

                  {/* Created */}
                  <td className="px-6 py-4 whitespace-nowrap">
                    {snapshot.createdAt ? (
                      <div className="flex items-center gap-1 text-sm text-muted-foreground">
                        <Clock className="h-3 w-3" />
                        <span>{formatTimeAgo(snapshot.createdAt)}</span>
                      </div>
                    ) : (
                      <span className="text-muted-foreground text-sm">-</span>
                    )}
                  </td>

                  {/* Size */}
                  <td className="px-6 py-4 whitespace-nowrap">
                    {snapshot.sizeHuman && snapshot.sizeHuman !== '0 B' ? (
                      <div className="flex items-center gap-1 text-sm text-muted-foreground">
                        <HardDrive className="h-3 w-3" />
                        <span>{snapshot.sizeHuman}</span>
                      </div>
                    ) : (
                      <span className="text-muted-foreground text-sm">-</span>
                    )}
                  </td>

                  {/* Status */}
                  <td className="px-6 py-4 whitespace-nowrap">
                    {isInProgress ? (
                      <Badge variant="default" className="gap-1">
                        <Loader2 className="h-3 w-3 animate-spin" />
                        {snapshot.state}
                      </Badge>
                    ) : isFailed ? (
                      <Badge variant="destructive" className="gap-1">
                        <AlertCircle className="h-3 w-3" />
                        Failed
                      </Badge>
                    ) : (
                      <Badge variant="secondary">Ready</Badge>
                    )}
                  </td>

                  {/* Actions */}
                  <td className="px-6 py-4 text-right text-sm whitespace-nowrap">
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
                  </td>
                </tr>
              )
            })}
          </tbody>
        </table>
      </div>
    </div>
  )
}
