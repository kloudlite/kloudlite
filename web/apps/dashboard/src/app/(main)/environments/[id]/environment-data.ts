import { cache } from 'react'
import { getEnvironmentByHash } from '@/app/actions/environment.actions'

/**
 * Request-scoped cached version of getEnvironmentByHash.
 * React's cache() deduplicates calls with the same arguments within a single
 * server request, so layout.tsx + page.tsx share one fetch instead of calling twice.
 */
export const getEnvironmentData = cache(getEnvironmentByHash)
