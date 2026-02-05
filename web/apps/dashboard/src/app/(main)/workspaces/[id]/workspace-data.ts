import { cache } from 'react'
import { getWorkspaceByHash } from '@/app/actions/workspace.actions'

/**
 * Request-scoped cached version of getWorkspaceByHash.
 * React's cache() deduplicates calls with the same arguments within a single
 * server request, so layout.tsx + page.tsx share one fetch instead of calling twice.
 */
export const getWorkspaceData = cache(getWorkspaceByHash)
