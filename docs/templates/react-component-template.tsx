/**
 * Template for React components
 * Copy this template when creating new components in /web/components/
 */

// ===== CLIENT COMPONENT TEMPLATE =====

"use client"

import { useState, useCallback, useMemo } from "react"
import { useRouter, useSearchParams } from "next/navigation"
import { cn } from "@/lib/utils"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Loader2, Plus, Search } from "lucide-react"

// ===== TYPE DEFINITIONS =====

interface Entity {
  id: string
  name: string
  description?: string
  status: 'active' | 'inactive' | 'pending'
  createdAt: Date
  updatedAt: Date
}

interface EntityCardProps {
  entity: Entity
  onEdit?: (entity: Entity) => void
  onDelete?: (id: string) => void
  className?: string
}

// ===== CLIENT COMPONENT =====

export function EntityCard({ 
  entity, 
  onEdit, 
  onDelete,
  className 
}: EntityCardProps) {
  const [isDeleting, setIsDeleting] = useState(false)
  
  const handleDelete = useCallback(async () => {
    if (!onDelete) return
    
    if (!confirm('Are you sure you want to delete this entity?')) {
      return
    }
    
    setIsDeleting(true)
    try {
      await onDelete(entity.id)
    } finally {
      setIsDeleting(false)
    }
  }, [entity.id, onDelete])
  
  const statusColor = useMemo(() => {
    switch (entity.status) {
      case 'active': return 'bg-green-500'
      case 'inactive': return 'bg-gray-400'
      case 'pending': return 'bg-yellow-500'
      default: return 'bg-gray-400'
    }
  }, [entity.status])
  
  return (
    <Card className={cn(
      "relative overflow-hidden transition-all hover:shadow-md",
      className
    )}>
      {/* Status indicator */}
      <div className={cn("absolute inset-x-0 top-0 h-1", statusColor)} />
      
      <CardHeader>
        <div className="flex items-start justify-between">
          <div className="space-y-1">
            <CardTitle className="line-clamp-1">{entity.name}</CardTitle>
            {entity.description && (
              <CardDescription className="line-clamp-2">
                {entity.description}
              </CardDescription>
            )}
          </div>
          
          <div className="flex items-center gap-2">
            {onEdit && (
              <Button
                variant="ghost"
                size="sm"
                onClick={() => onEdit(entity)}
              >
                Edit
              </Button>
            )}
            
            {onDelete && (
              <Button
                variant="ghost"
                size="sm"
                onClick={handleDelete}
                disabled={isDeleting}
              >
                {isDeleting ? (
                  <Loader2 className="h-4 w-4 animate-spin" />
                ) : (
                  'Delete'
                )}
              </Button>
            )}
          </div>
        </div>
      </CardHeader>
      
      <CardContent>
        <div className="flex items-center justify-between text-xs text-muted-foreground">
          <span>Created {new Date(entity.createdAt).toLocaleDateString()}</span>
          <span className="capitalize">{entity.status}</span>
        </div>
      </CardContent>
    </Card>
  )
}

// ===== SERVER COMPONENT TEMPLATE =====

import { Suspense } from "react"
import { notFound } from "next/navigation"
import { getEntity, listEntities } from "@/app/actions/entities"
import { Skeleton } from "@/components/ui/skeleton"

interface EntityListProps {
  searchParams?: {
    page?: string
    search?: string
    status?: string
  }
}

// Server component - no "use client" directive
export async function EntityList({ searchParams }: EntityListProps) {
  const page = parseInt(searchParams?.page || '1')
  const search = searchParams?.search || ''
  const status = searchParams?.status as 'active' | 'inactive' | 'pending' | undefined
  
  const result = await listEntities({
    page,
    search,
    status,
  })
  
  if (!result.success) {
    throw new Error(result.error)
  }
  
  const { items: entities, pageInfo } = result.data
  
  return (
    <div className="space-y-6">
      {/* Search and filters */}
      <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <EntitySearch defaultValue={search} />
        <EntityFilters currentStatus={status} />
      </div>
      
      {/* Entity grid */}
      {entities.length === 0 ? (
        <EmptyState />
      ) : (
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {entities.map((entity) => (
            <EntityCard
              key={entity.id}
              entity={entity}
              onEdit={(entity) => console.log('Edit', entity)}
            />
          ))}
        </div>
      )}
      
      {/* Pagination */}
      {pageInfo.totalPages > 1 && (
        <Pagination
          currentPage={pageInfo.currentPage}
          totalPages={pageInfo.totalPages}
          hasNext={pageInfo.hasNext}
          hasPrevious={pageInfo.hasPrevious}
        />
      )}
    </div>
  )
}

// ===== LOADING STATES =====

export function EntityListSkeleton() {
  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <Skeleton className="h-10 w-64" />
        <Skeleton className="h-10 w-32" />
      </div>
      
      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
        {Array.from({ length: 6 }).map((_, i) => (
          <Card key={i}>
            <CardHeader>
              <Skeleton className="h-5 w-3/4" />
              <Skeleton className="h-4 w-full" />
            </CardHeader>
            <CardContent>
              <Skeleton className="h-3 w-1/2" />
            </CardContent>
          </Card>
        ))}
      </div>
    </div>
  )
}

// ===== EMPTY STATES =====

function EmptyState() {
  const router = useRouter()
  
  return (
    <div className="flex min-h-[400px] flex-col items-center justify-center rounded-lg border-2 border-dashed bg-muted/20 p-8 text-center">
      <div className="mx-auto max-w-md space-y-4">
        <h3 className="text-lg font-semibold">No entities found</h3>
        <p className="text-sm text-muted-foreground">
          Get started by creating your first entity.
        </p>
        <Button onClick={() => router.push('/entities/new')}>
          <Plus className="mr-2 h-4 w-4" />
          Create Entity
        </Button>
      </div>
    </div>
  )
}

// ===== SEARCH COMPONENT =====

"use client"

function EntitySearch({ defaultValue = '' }: { defaultValue?: string }) {
  const router = useRouter()
  const searchParams = useSearchParams()
  const [search, setSearch] = useState(defaultValue)
  
  const handleSearch = useCallback((e: React.FormEvent) => {
    e.preventDefault()
    const params = new URLSearchParams(searchParams)
    
    if (search) {
      params.set('search', search)
    } else {
      params.delete('search')
    }
    
    params.set('page', '1') // Reset to first page
    router.push(`?${params.toString()}`)
  }, [search, searchParams, router])
  
  return (
    <form onSubmit={handleSearch} className="w-full sm:w-64">
      <div className="relative">
        <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
        <Input
          type="search"
          placeholder="Search entities..."
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          className="pl-9"
        />
      </div>
    </form>
  )
}

// ===== FILTER COMPONENT =====

"use client"

function EntityFilters({ currentStatus }: { currentStatus?: string }) {
  const router = useRouter()
  const searchParams = useSearchParams()
  
  const handleStatusChange = useCallback((status: string) => {
    const params = new URLSearchParams(searchParams)
    
    if (status === 'all') {
      params.delete('status')
    } else {
      params.set('status', status)
    }
    
    params.set('page', '1')
    router.push(`?${params.toString()}`)
  }, [searchParams, router])
  
  return (
    <div className="flex items-center gap-2">
      <Button
        variant={!currentStatus ? 'default' : 'outline'}
        size="sm"
        onClick={() => handleStatusChange('all')}
      >
        All
      </Button>
      <Button
        variant={currentStatus === 'active' ? 'default' : 'outline'}
        size="sm"
        onClick={() => handleStatusChange('active')}
      >
        Active
      </Button>
      <Button
        variant={currentStatus === 'inactive' ? 'default' : 'outline'}
        size="sm"
        onClick={() => handleStatusChange('inactive')}
      >
        Inactive
      </Button>
    </div>
  )
}

// ===== PAGINATION COMPONENT =====

"use client"

interface PaginationProps {
  currentPage: number
  totalPages: number
  hasNext: boolean
  hasPrevious: boolean
}

function Pagination({ currentPage, totalPages, hasNext, hasPrevious }: PaginationProps) {
  const router = useRouter()
  const searchParams = useSearchParams()
  
  const navigate = useCallback((page: number) => {
    const params = new URLSearchParams(searchParams)
    params.set('page', page.toString())
    router.push(`?${params.toString()}`)
  }, [searchParams, router])
  
  return (
    <div className="flex items-center justify-center gap-2">
      <Button
        variant="outline"
        size="sm"
        onClick={() => navigate(currentPage - 1)}
        disabled={!hasPrevious}
      >
        Previous
      </Button>
      
      <span className="text-sm text-muted-foreground">
        Page {currentPage} of {totalPages}
      </span>
      
      <Button
        variant="outline"
        size="sm"
        onClick={() => navigate(currentPage + 1)}
        disabled={!hasNext}
      >
        Next
      </Button>
    </div>
  )
}