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

// Visibility enum
const visibilitySchema = z.enum(['private', 'shared', 'open']).optional()

// Resource quotas schema
const resourceQuotasSchema = z
  .object({
    limitsCPU: z.string().optional(),
    limitsMemory: z.string().optional(),
    requestsCPU: z.string().optional(),
    requestsMemory: z.string().optional(),
    persistentVolumeClaims: z.string().optional(),
  })
  .optional()

// Network policy port schema
const networkPolicyPortSchema = z.object({
  port: z.number().int().min(1).max(65535),
  protocol: z.enum(['TCP', 'UDP']).optional(),
})

// Label selector schema
const labelSelectorSchema = z
  .object({
    matchLabels: z.record(z.string()).optional(),
  })
  .optional()

// Network policy peer schema
const networkPolicyPeerSchema = z.object({
  namespaceSelector: labelSelectorSchema,
  podSelector: labelSelectorSchema,
})

// Ingress rule schema
const ingressRuleSchema = z.object({
  from: z.array(networkPolicyPeerSchema).optional(),
  ports: z.array(networkPolicyPortSchema).optional(),
})

// Network policies schema
const networkPoliciesSchema = z
  .object({
    allowedNamespaces: z.array(z.string()).optional(),
    ingressRules: z.array(ingressRuleSchema).optional(),
  })
  .optional()

// Environment spec schema
export const environmentSpecSchema = z.object({
  targetNamespace: z.string().optional(),
  name: z.string().optional(),
  ownedBy: z.string().email('Owner must be a valid email'),
  visibility: visibilitySchema,
  sharedWith: z.array(z.string().email()).optional(),
  activated: z.boolean(),
  labels: z.record(z.string()).optional(),
  annotations: z.record(z.string()).optional(),
  resourceQuotas: resourceQuotasSchema,
  networkPolicies: networkPoliciesSchema,
  cloneFrom: z.string().optional(),
})

// Environment create request schema
export const environmentCreateSchema = z.object({
  name: kubernetesNameSchema,
  spec: environmentSpecSchema,
})

// Environment update request schema
export const environmentUpdateSchema = z.object({
  spec: environmentSpecSchema,
})

// Clone environment schema
export const cloneEnvironmentSchema = z.object({
  sourceName: kubernetesNameSchema,
  targetName: kubernetesNameSchema,
  targetNamespace: kubernetesNameSchema,
  cloneEnvVars: z.boolean(),
  cloneFiles: z.boolean(),
  currentUser: z.string().email(),
})

// Environment variable schema
export const envVarSchema = z.object({
  key: z
    .string()
    .min(1, 'Key is required')
    .max(256, 'Key too long')
    .regex(/^[a-zA-Z_][a-zA-Z0-9_]*$/, 'Key must be a valid environment variable name'),
  value: z.string().max(1048576, 'Value too large'), // 1MB limit
  type: z.enum(['config', 'secret']),
})

// File schema
export const fileSchema = z.object({
  filename: z
    .string()
    .min(1, 'Filename is required')
    .max(253, 'Filename too long')
    .regex(/^[a-zA-Z0-9._-]+$/, 'Filename contains invalid characters'),
  content: z.string().max(10485760, 'File content too large'), // 10MB limit
})

// Import environment config schema
export const importEnvironmentConfigSchema = z.object({
  newEnvName: kubernetesNameSchema,
  targetNamespace: kubernetesNameSchema,
  currentUser: z.string().email(),
  exportData: z.object({
    configs: z.record(z.string()).optional(),
    secrets: z.record(z.string()).optional(),
    files: z
      .array(
        z.object({
          name: z.string(),
          content: z.string(),
        })
      )
      .optional(),
    compositions: z
      .array(
        z.object({
          name: z.string(),
          spec: z.unknown(),
        })
      )
      .optional(),
  }),
})

// Simple parameter validations
export const environmentNameSchema = kubernetesNameSchema

// Export types
export type EnvironmentCreateInput = z.infer<typeof environmentCreateSchema>
export type EnvironmentUpdateInput = z.infer<typeof environmentUpdateSchema>
export type CloneEnvironmentInput = z.infer<typeof cloneEnvironmentSchema>
export type EnvVarInput = z.infer<typeof envVarSchema>
export type FileInput = z.infer<typeof fileSchema>
export type ImportEnvironmentConfigInput = z.infer<typeof importEnvironmentConfigSchema>
