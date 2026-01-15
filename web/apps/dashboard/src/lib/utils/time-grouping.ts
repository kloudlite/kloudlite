export function groupSnapshotsByTime<T extends { createdAt?: string }>(
  snapshots: T[]
): Record<string, T[]> {
  const now = new Date()
  const groups: Record<string, T[]> = {
    Today: [],
    Yesterday: [],
    'This Week': [],
    'This Month': [],
    Older: [],
  }

  snapshots.forEach((snapshot) => {
    const created = new Date(snapshot.createdAt || '')
    const diffMs = now.getTime() - created.getTime()
    const diffDays = Math.floor(diffMs / (1000 * 60 * 60 * 24))

    if (diffDays === 0) {
      groups.Today.push(snapshot)
    } else if (diffDays === 1) {
      groups.Yesterday.push(snapshot)
    } else if (diffDays <= 7) {
      groups['This Week'].push(snapshot)
    } else if (diffDays <= 30) {
      groups['This Month'].push(snapshot)
    } else {
      groups.Older.push(snapshot)
    }
  })

  return groups
}
