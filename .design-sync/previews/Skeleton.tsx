import * as React from 'react'
import { Skeleton } from '@kloudlite/ui'

export const CardLoading = () => (
  <div className="w-80 space-y-4 border p-4">
    <div className="flex items-center gap-3">
      <Skeleton className="size-10 rounded-full" />
      <div className="space-y-2">
        <Skeleton className="h-4 w-40" />
        <Skeleton className="h-3 w-24" />
      </div>
    </div>
    <Skeleton className="h-3 w-full" />
    <Skeleton className="h-3 w-full" />
    <Skeleton className="h-3 w-32" />
  </div>
)

export const TableRows = () => (
  <div className="w-96 space-y-3">
    <div className="flex items-center gap-4">
      <Skeleton className="h-3 w-32" />
      <Skeleton className="h-3 w-20" />
      <Skeleton className="h-3 w-16" />
    </div>
    <div className="flex items-center gap-4">
      <Skeleton className="h-4 w-32" />
      <Skeleton className="h-4 w-20" />
      <Skeleton className="h-4 w-16" />
    </div>
    <div className="flex items-center gap-4">
      <Skeleton className="h-4 w-32" />
      <Skeleton className="h-4 w-20" />
      <Skeleton className="h-4 w-16" />
    </div>
    <div className="flex items-center gap-4">
      <Skeleton className="h-4 w-32" />
      <Skeleton className="h-4 w-20" />
      <Skeleton className="h-4 w-16" />
    </div>
  </div>
)

export const Shapes = () => (
  <div className="flex items-center gap-6">
    <Skeleton className="size-12 rounded-full" />
    <Skeleton className="h-12 w-12" />
    <Skeleton className="h-24 w-40" />
  </div>
)

export const TextBlock = () => (
  <div className="w-80 space-y-2">
    <Skeleton className="h-6 w-64" />
    <Skeleton className="h-3 w-full" />
    <Skeleton className="h-3 w-full" />
    <Skeleton className="h-3 w-80" />
    <Skeleton className="h-3 w-40" />
  </div>
)
