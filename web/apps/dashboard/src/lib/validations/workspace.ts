import { z } from 'zod'

// Kubernetes name validation (DNS-1123 label)
const kubernetesNameSchema = z
  .string()
  .min(1, 'Name is required')
  .max(63, 'Name must be at most 63 characters')
  .regex(
    /^[a-z0-9][a-z0-9-]*[a-z0-9]$|^[a-z0-9]$/,
    'Name must start and end with alphanumeric characters and contain only lowercase letters, numbers, and hyphens'
  )

// Namespace validation
const namespaceSchema = z
  .string()
  .min(1, 'Namespace is required')
  .max(63, 'Namespace must be at most 63 characters')
  .regex(
    /^[a-z0-9][a-z0-9-]*[a-z0-9]$|^[a-z0-9]$/,
    'Namespace must be a valid Kubernetes namespace name'
  )

// Git URL validation - accepts both HTTPS and SSH formats
const gitUrlSchema = z.string().refine(
  (url) => {
    if (!url) return true
    // Accept HTTPS URLs
    if (url.startsWith('https://') || url.startsWith('http://')) {
      try {
        new URL(url)
        return true
      } catch {
        return false
      }
    }
    // Accept SSH format: git@host:user/repo.git
    if (/^git@[\w.-]+:[\w./-]+\.git$/.test(url)) {
      return true
    }
    // Accept SSH format without .git suffix
    if (/^git@[\w.-]+:[\w./-]+$/.test(url)) {
      return true
    }
    return false
  },
  { message: 'Git URL must be a valid HTTPS or SSH URL' }
)

// Git repository schema
const gitRepositorySchema = z
  .object({
    url: gitUrlSchema,
    branch: z.string().optional(),
  })
  .optional()

// Workspace settings schema - only fields used by controller
const workspaceSettingsSchema = z
  .object({
    idleTimeout: z.number().int().min(0).optional(),
    startupScript: z.string().optional(),
    environmentVariables: z.record(z.string()).optional(),
  })
  .optional()

// Exposed port schema - matches Go API (only port, no protocol)
const exposedPortSchema = z.object({
  port: z.number().int().min(1).max(65535),
})

// Object reference schema (for environmentRef inside environmentConnection)
const objectReferenceSchema = z.object({
  name: z.string().min(1),
  namespace: z.string().min(1),
  kind: z.string().optional(),
  apiVersion: z.string().optional(),
})

// Environment connection schema - matches Go API
const environmentConnectionSchema = z
  .object({
    environmentRef: objectReferenceSchema,
  })
  .optional()

// Visibility enum
const visibilitySchema = z.enum(['private', 'shared', 'open']).optional()

// Workspace spec schema - only fields used by controller
export const workspaceSpecSchema = z.object({
  displayName: z.string().min(1, 'Display name is required').max(100, 'Display name too long'),
  ownedBy: z.string().min(1, 'Owner is required'),
  workmachine: z.string().optional(), // Auto-populated from namespace by webhook
  visibility: visibilitySchema,
  sharedWith: z.array(z.string()).optional(),
  environmentConnection: environmentConnectionSchema,
  gitRepository: gitRepositorySchema,
  settings: workspaceSettingsSchema,
  vscodeVersion: z.string().optional(),
  status: z.enum(['active', 'suspended', 'archived']).optional(),
  copyFrom: z.string().optional(),
  expose: z.array(exposedPortSchema).optional(),
})

// Workspace create request schema
export const workspaceCreateSchema = z.object({
  name: kubernetesNameSchema,
  spec: workspaceSpecSchema,
})

// Workspace update spec schema - all fields optional for partial updates
export const workspaceUpdateSpecSchema = z.object({
  displayName: z.string().min(1).max(100).optional(),
  ownedBy: z.string().optional(),
  workmachine: z.string().optional(),
  visibility: visibilitySchema,
  sharedWith: z.array(z.string()).optional(),
  environmentConnection: environmentConnectionSchema,
  gitRepository: gitRepositorySchema,
  settings: workspaceSettingsSchema,
  vscodeVersion: z.string().optional(),
  status: z.enum(['active', 'suspended', 'archived']).optional(),
  copyFrom: z.string().optional(),
  expose: z.array(exposedPortSchema).optional(),
})

// Workspace update request schema
export const workspaceUpdateSchema = z.object({
  spec: workspaceUpdateSpecSchema,
})

// Package update schema
export const packageUpdateSchema = z.object({
  packages: z.array(
    z.object({
      name: z.string().min(1, 'Package name is required'),
      nixpkgsCommit: z.string().optional(),
    })
  ),
})

// Simple parameter validations
export const workspaceNameSchema = kubernetesNameSchema
export const workspaceNamespaceSchema = namespaceSchema.default('default')

// Export types
export type WorkspaceCreateInput = z.infer<typeof workspaceCreateSchema>
export type WorkspaceUpdateInput = z.infer<typeof workspaceUpdateSchema>
export type PackageUpdateInput = z.infer<typeof packageUpdateSchema>
