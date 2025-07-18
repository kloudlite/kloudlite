'use client'

import { useState, useMemo } from 'react'
import { sortData, searchData, getNextSortDirection, type SortDirection } from '@/lib/utils/sorting'

interface UseTableStateOptions<T> {
  data: T[]
  searchFields?: (keyof T)[]
  initialSort?: {
    field: keyof T
    direction: SortDirection
  }
}

export function useTableState<T extends Record<string, any>>({
  data,
  searchFields = [],
  initialSort
}: UseTableStateOptions<T>) {
  const [searchQuery, setSearchQuery] = useState('')
  const [sortField, setSortField] = useState<keyof T | undefined>(initialSort?.field)
  const [sortDirection, setSortDirection] = useState<SortDirection>(initialSort?.direction || 'asc')

  // Filter and sort data
  const processedData = useMemo(() => {
    let filtered = data

    // Apply search filter
    if (searchQuery.trim() && searchFields.length > 0) {
      filtered = searchData(data, searchQuery, searchFields)
    }

    // Apply sorting
    if (sortField) {
      filtered = sortData(filtered, sortField, sortDirection)
    }

    return filtered
  }, [data, searchQuery, searchFields, sortField, sortDirection])

  const handleSort = (field: keyof T) => {
    const newDirection = getNextSortDirection(
      String(sortField),
      String(field),
      sortDirection
    )
    setSortField(field)
    setSortDirection(newDirection)
  }

  const handleSearch = (query: string) => {
    setSearchQuery(query)
  }

  const clearSearch = () => {
    setSearchQuery('')
  }

  const resetSort = () => {
    setSortField(undefined)
    setSortDirection('asc')
  }

  return {
    // Data
    processedData,
    
    // Search state
    searchQuery,
    handleSearch,
    clearSearch,
    
    // Sort state
    sortField,
    sortDirection,
    handleSort,
    resetSort,
    
    // Utility
    hasFilters: !!searchQuery.trim() || !!sortField,
    totalCount: data.length,
    filteredCount: processedData.length,
  }
}