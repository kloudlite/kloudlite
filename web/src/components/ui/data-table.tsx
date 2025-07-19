'use client'

import { useState, useMemo } from 'react'
import { Button } from '@/components/ui/button'
import { ArrowUp, ArrowDown, ArrowUpDown, Users, Plus } from 'lucide-react'
import { cn } from '@/lib/utils'

export type SortDirection = 'asc' | 'desc'

export interface ColumnDef<T> {
  key: keyof T
  header: string
  sortable?: boolean
  className?: string
  hideOnMobile?: boolean
  render?: (value: T[keyof T], item: T) => React.ReactNode
}

interface DataTableProps<T> {
  data: T[]
  columns: ColumnDef<T>[]
  initialDisplayCount?: number
  onSort?: (field: keyof T, direction: SortDirection) => void
  sortField?: keyof T
  sortDirection?: SortDirection
  emptyState?: {
    icon?: React.ComponentType<{ className?: string }>
    title: string
    description: string
    action?: React.ReactNode
  }
  className?: string
  focusedRowIndex?: number
  onKeyDown?: (e: React.KeyboardEvent) => void
}

export function DataTable<T extends Record<string, any>>({
  data,
  columns,
  initialDisplayCount = 10,
  onSort,
  sortField,
  sortDirection = 'asc',
  emptyState,
  className,
  focusedRowIndex = -1,
  onKeyDown
}: DataTableProps<T>) {
  const [displayCount, setDisplayCount] = useState(initialDisplayCount)
  const hasMore = displayCount < data.length
  const displayedData = data.slice(0, displayCount)

  const handleLoadMore = () => {
    setDisplayCount(prev => Math.min(prev + 10, data.length))
  }

  const SortableHeader = ({ 
    column 
  }: { 
    column: ColumnDef<T>
  }) => {
    if (!column.sortable || !onSort) {
      return (
        <th className={cn('text-left font-medium px-6 py-3', column.className)}>
          {column.header}
        </th>
      )
    }

    const isActive = sortField === column.key
    const Icon = isActive 
      ? (sortDirection === 'asc' ? ArrowUp : ArrowDown)
      : ArrowUpDown

    return (
      <th className={cn('text-left font-medium px-6 py-3', column.className)}>
        <Button 
          variant="ghost" 
          size="sm" 
          onClick={() => onSort(column.key, isActive && sortDirection === 'asc' ? 'desc' : 'asc')}
          className="h-auto p-0 hover:bg-transparent font-medium text-muted-foreground hover:text-foreground"
        >
          <span className="flex items-center gap-1">
            {column.header}
            <Icon className={cn('h-3.5 w-3.5', isActive ? 'text-foreground' : 'text-muted-foreground/50')} />
          </span>
        </Button>
      </th>
    )
  }

  const EmptyState = () => {
    const Icon = emptyState?.icon || Users
    
    return (
      <tr>
        <td colSpan={columns.length} className="px-6 py-12 text-center">
          <div className="max-w-sm mx-auto">
            <Icon className="h-10 w-10 mx-auto mb-3 text-muted-foreground/50" />
            <h3 className="font-medium mb-1">
              {emptyState?.title || 'No data found'}
            </h3>
            <p className="text-sm text-muted-foreground mb-4">
              {emptyState?.description || 'No items to display'}
            </p>
            {emptyState?.action}
          </div>
        </td>
      </tr>
    )
  }

  return (
    <>
      <div className={cn('overflow-x-auto', className)}>
        <table className="w-full">
          <thead>
            <tr className="border-b text-sm text-muted-foreground">
              {columns.map((column, index) => (
                <SortableHeader key={String(column.key) + index} column={column} />
              ))}
            </tr>
          </thead>
          <tbody className="divide-y">
            {displayedData.length === 0 ? (
              <EmptyState />
            ) : (
              displayedData.map((item, index) => {
                const isFocused = focusedRowIndex === index
                return (
                  <tr 
                    key={index} 
                    className={cn(
                      'hover:bg-muted/50 transition-colors group focus-within:bg-muted/50',
                      isFocused && 'bg-muted/70 ring-1 ring-ring'
                    )}
                    role="row"
                    tabIndex={isFocused ? 0 : -1}
                  >
                    {columns.map((column, colIndex) => {
                      const value = item[column.key]
                      const cellContent = column.render ? column.render(value, item) : String(value || '')
                      
                      return (
                        <td 
                          key={String(column.key) + colIndex}
                          className={cn('px-6 py-4', column.className)}
                        >
                          {cellContent}
                        </td>
                      )
                    })}
                  </tr>
                )
              })
            )}
          </tbody>
        </table>
      </div>
      
      {hasMore && (
        <div className="border-t px-6 py-4 text-center">
          <Button variant="ghost" onClick={handleLoadMore}>
            Load More ({data.length - displayCount} remaining)
          </Button>
        </div>
      )}
    </>
  )
}