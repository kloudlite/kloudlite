'use client'

import { useState } from 'react'
import { Users, User, Activity, Grid } from 'lucide-react'

interface EnvironmentFilterProps {
  onFilterChange: (scope: 'all' | 'mine', status: 'all' | 'active') => void
  counts: {
    all: number
    mine: number
    active: number
    mineActive: number
  }
}

export function EnvironmentFilter({ onFilterChange, counts }: EnvironmentFilterProps) {
  const [scopeFilter, setScopeFilter] = useState<'all' | 'mine'>('all')
  const [statusFilter, setStatusFilter] = useState<'all' | 'active'>('all')

  const handleScopeChange = (scope: 'all' | 'mine') => {
    setScopeFilter(scope)
    onFilterChange(scope, statusFilter)
  }

  const handleStatusChange = (status: 'all' | 'active') => {
    setStatusFilter(status)
    onFilterChange(scopeFilter, status)
  }

  // Calculate current count based on both filters
  const getCurrentCount = () => {
    if (scopeFilter === 'mine' && statusFilter === 'active') {
      return counts.mineActive
    } else if (scopeFilter === 'mine') {
      return counts.mine
    } else if (statusFilter === 'active') {
      return counts.active
    }
    return counts.all
  }

  return (
    <div className="flex items-center gap-3">
      {/* Scope Filter */}
      <div className="bg-muted inline-flex items-center rounded-lg p-1">
        <button
          onClick={() => handleScopeChange('all')}
          className={`flex items-center gap-2 rounded-md px-4 py-2 text-sm font-medium transition-all ${
            scopeFilter === 'all'
              ? 'bg-background shadow-sm'
              : 'text-muted-foreground hover:text-foreground'
          } `}
        >
          <Users className="h-4 w-4" />
          <span>All</span>
        </button>

        <button
          onClick={() => handleScopeChange('mine')}
          className={`flex items-center gap-2 rounded-md px-4 py-2 text-sm font-medium transition-all ${
            scopeFilter === 'mine'
              ? 'bg-background shadow-sm'
              : 'text-muted-foreground hover:text-foreground'
          } `}
        >
          <User className="h-4 w-4" />
          <span>Mine</span>
        </button>
      </div>

      {/* Status Filter */}
      <div className="bg-muted inline-flex items-center rounded-lg p-1">
        <button
          onClick={() => handleStatusChange('all')}
          className={`flex items-center gap-2 rounded-md px-4 py-2 text-sm font-medium transition-all ${
            statusFilter === 'all'
              ? 'bg-background shadow-sm'
              : 'text-muted-foreground hover:text-foreground'
          } `}
        >
          <Grid className="h-4 w-4" />
          <span>All Status</span>
        </button>

        <button
          onClick={() => handleStatusChange('active')}
          className={`flex items-center gap-2 rounded-md px-4 py-2 text-sm font-medium transition-all ${
            statusFilter === 'active'
              ? 'bg-background shadow-sm'
              : 'text-muted-foreground hover:text-foreground'
          } `}
        >
          <Activity className="h-4 w-4" />
          <span>Active</span>
        </button>
      </div>

      {/* Result Count */}
      <div className="text-muted-foreground flex items-center gap-2 px-3 py-2 text-sm">
        <span>Showing</span>
        <span className="font-semibold">{getCurrentCount()}</span>
        <span>environments</span>
      </div>
    </div>
  )
}
