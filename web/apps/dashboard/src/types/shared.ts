/**
 * Shared type definitions used across the dashboard app
 */

/**
 * Common PageProps interface for dynamic route pages
 * Used by pages with [id] parameter
 */
export interface PageProps {
  params: Promise<{ id: string }>
}

/**
 * Extended PageProps with namespace parameter
 */
export interface PagePropsWithNamespace {
  params: Promise<{ id: string; namespace?: string }>
}

/**
 * Layout props for layouts that receive children and params
 */
export interface LayoutProps {
  children: React.ReactNode
  params: Promise<{ id: string }>
}

/**
 * Pinned workspace reference for quick access
 */
export interface PinnedWorkspace {
  id: string
  name: string
  hash: string
  environment: string
  status: 'active' | 'idle'
}

/**
 * Pinned environment reference for quick access
 */
export interface PinnedEnvironment {
  id: string
  name: string
  hash: string
  status: 'active' | 'idle'
}

/**
 * Common status type used across resources
 */
export type ResourceStatus = 'active' | 'idle' | 'suspended' | 'archived' | 'pending' | 'error'
