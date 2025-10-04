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
      <div className="inline-flex items-center p-1 bg-gray-100 rounded-lg">
        <button
          onClick={() => handleScopeChange('all')}
          className={`
            flex items-center gap-2 px-4 py-2 rounded-md text-sm font-medium transition-all
            ${scopeFilter === 'all'
              ? 'bg-white text-gray-900 shadow-sm'
              : 'text-gray-600 hover:text-gray-900'
            }
          `}
        >
          <Users className="h-4 w-4" />
          <span>All</span>
        </button>

        <button
          onClick={() => handleScopeChange('mine')}
          className={`
            flex items-center gap-2 px-4 py-2 rounded-md text-sm font-medium transition-all
            ${scopeFilter === 'mine'
              ? 'bg-white text-gray-900 shadow-sm'
              : 'text-gray-600 hover:text-gray-900'
            }
          `}
        >
          <User className="h-4 w-4" />
          <span>Mine</span>
        </button>
      </div>

      {/* Status Filter */}
      <div className="inline-flex items-center p-1 bg-gray-100 rounded-lg">
        <button
          onClick={() => handleStatusChange('all')}
          className={`
            flex items-center gap-2 px-4 py-2 rounded-md text-sm font-medium transition-all
            ${statusFilter === 'all'
              ? 'bg-white text-gray-900 shadow-sm'
              : 'text-gray-600 hover:text-gray-900'
            }
          `}
        >
          <Grid className="h-4 w-4" />
          <span>All Status</span>
        </button>

        <button
          onClick={() => handleStatusChange('active')}
          className={`
            flex items-center gap-2 px-4 py-2 rounded-md text-sm font-medium transition-all
            ${statusFilter === 'active'
              ? 'bg-white text-gray-900 shadow-sm'
              : 'text-gray-600 hover:text-gray-900'
            }
          `}
        >
          <Activity className="h-4 w-4" />
          <span>Active</span>
        </button>
      </div>

      {/* Result Count */}
      <div className="flex items-center gap-2 px-3 py-2 text-sm text-gray-600">
        <span>Showing</span>
        <span className="font-semibold text-gray-900">{getCurrentCount()}</span>
        <span>environments</span>
      </div>
    </div>
  )
}