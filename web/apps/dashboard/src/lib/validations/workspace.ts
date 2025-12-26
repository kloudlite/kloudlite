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

// Visibility enum
const visibilitySchema = z.enum(['private', 'shared', 'open']).optional()

// Git repository schema
const gitRepositorySchema = z
  .object({
    url: z.string().url('Git URL must be a valid URL'),
    branch: z.string().optional(),
  })
  .optional()

// Resource quota schema
const resourceQuotaSchema = z
  .object({
    cpu: z.string().optional(),
    memory: z.string().optional(),
    storage: z.string().optional(),
    gpus: z.number().int().min(0).optional(),
  })
  .optional()

// Workspace settings schema
const workspaceSettingsSchema = z
  .object({
    autoStop: z.boolean().optional(),
    idleTimeout: z.number().int().min(0).optional(),
    maxRuntime: z.number().int().min(0).optional(),
    startupScript: z.string().optional(),
    environmentVariables: z.record(z.string()).optional(),
    vscodeExtensions: z.array(z.string()).optional(),
    gitConfig: z
      .object({
        userName: z.string().optional(),
        userEmail: z.string().email().optional(),
        defaultBranch: z.string().optional(),
      })
      .optional(),
    dotfilesRepo: z.string().optional(),
  })
  .optional()

// Exposed port schema
const exposedPortSchema = z.object({
  port: z.number().int().min(1).max(65535),
  protocol: z.enum(['tcp', 'udp', 'http']),
})

// Object reference schema
const objectReferenceSchema = z
  .object({
    name: z.string().min(1),
    namespace: z.string().min(1),
    kind: z.string().optional(),
    apiVersion: z.string().optional(),
  })
  .optional()

// Workspace spec schema
export const workspaceSpecSchema = z.object({
  displayName: z.string().min(1, 'Display name is required').max(100, 'Display name too long'),
  description: z.string().max(500, 'Description too long').optional(),
  ownedBy: z.string().min(1, 'Owner is required'),
  visibility: visibilitySchema,
  sharedWith: z.array(z.string().email()).optional(),
  workMachineRef: objectReferenceSchema,
  workmachineName: z.string().optional(),
  environmentRef: objectReferenceSchema,
  machineTypeRef: objectReferenceSchema,
  folderName: z.string().optional(),
  resourceQuota: resourceQuotaSchema,
  settings: workspaceSettingsSchema,
  status: z.enum(['active', 'suspended', 'archived']).optional(),
  tags: z.array(z.string()).optional(),
  vscodeVersion: z.string().optional(),
  gitRepository: gitRepositorySchema,
  copyFrom: z.string().optional(),
  expose: z.array(exposedPortSchema).optional(),
})

// Workspace create request schema
export const workspaceCreateSchema = z.object({
  name: kubernetesNameSchema,
  spec: workspaceSpecSchema,
})

// Workspace update request schema
export const workspaceUpdateSchema = z.object({
  spec: workspaceSpecSchema,
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
