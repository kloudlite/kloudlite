export type SortDirection = 'asc' | 'desc'

// Generic sorting function for arrays of objects
export function sortData<T extends Record<string, any>>(
  data: T[],
  field: keyof T,
  direction: SortDirection = 'asc'
): T[] {
  return [...data].sort((a, b) => {
    let aValue = a[field]
    let bValue = b[field]

    // Handle null/undefined values
    if (aValue == null && bValue == null) return 0
    if (aValue == null) return direction === 'asc' ? 1 : -1
    if (bValue == null) return direction === 'asc' ? -1 : 1

    // Handle Date objects
    if (aValue instanceof Date && bValue instanceof Date) {
      aValue = aValue.getTime()
      bValue = bValue.getTime()
    }

    // Handle strings (case-insensitive)
    if (typeof aValue === 'string' && typeof bValue === 'string') {
      aValue = aValue.toLowerCase()
      bValue = bValue.toLowerCase()
    }

    // Compare values
    if (aValue < bValue) return direction === 'asc' ? -1 : 1
    if (aValue > bValue) return direction === 'asc' ? 1 : -1
    return 0
  })
}

// Generic search function for arrays of objects
export function searchData<T extends Record<string, any>>(
  data: T[],
  searchQuery: string,
  searchFields: (keyof T)[]
): T[] {
  if (!searchQuery.trim()) return data

  const query = searchQuery.toLowerCase()
  
  return data.filter(item =>
    searchFields.some(field => {
      const value = item[field]
      if (value == null) return false
      return String(value).toLowerCase().includes(query)
    })
  )
}

// Utility to get sort direction toggle
export function getNextSortDirection(
  currentField: string,
  newField: string,
  currentDirection: SortDirection
): SortDirection {
  if (currentField === newField) {
    return currentDirection === 'asc' ? 'desc' : 'asc'
  }
  return 'asc'
}