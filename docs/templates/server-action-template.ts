/**
 * Template for Next.js server actions
 * Copy this template when creating new server actions in /web/app/actions/
 */

'use server'

import { getServerSession } from 'next-auth'
import { getAuthOptions } from '@/lib/auth/get-auth-options'
import { revalidatePath, revalidateTag } from 'next/cache'
import { z } from 'zod'
import { getGrpcClient, getAuthMetadata } from '@/lib/grpc/client'

// ===== INPUT VALIDATION SCHEMAS =====

const createEntitySchema = z.object({
  name: z.string().min(1).max(100),
  description: z.string().max(500).optional(),
  metadata: z.record(z.string()).optional(),
})

const updateEntitySchema = z.object({
  id: z.string(),
  name: z.string().min(1).max(100).optional(),
  description: z.string().max(500).optional(),
  status: z.enum(['active', 'inactive', 'pending']).optional(),
})

const listOptionsSchema = z.object({
  page: z.number().min(1).default(1),
  pageSize: z.number().min(1).max(100).default(20),
  sortBy: z.string().default('createdAt'),
  sortOrder: z.enum(['asc', 'desc']).default('desc'),
  status: z.enum(['active', 'inactive', 'pending']).optional(),
  search: z.string().optional(),
})

// ===== TYPE DEFINITIONS =====

type CreateEntityInput = z.infer<typeof createEntitySchema>
type UpdateEntityInput = z.infer<typeof updateEntitySchema>
type ListOptions = z.infer<typeof listOptionsSchema>

// Standard response types
type ActionResponse<T> = 
  | { success: true; data: T }
  | { success: false; error: string; details?: any }

type Entity = {
  id: string
  name: string
  description?: string
  status: 'active' | 'inactive' | 'pending'
  metadata?: Record<string, string>
  createdAt: Date
  updatedAt: Date
}

type PaginatedResponse<T> = {
  items: T[]
  pageInfo: {
    totalItems: number
    totalPages: number
    currentPage: number
    pageSize: number
    hasNext: boolean
    hasPrevious: boolean
  }
}

// ===== HELPER FUNCTIONS =====

async function requireAuth() {
  const authOpts = await getAuthOptions()
  const session = await getServerSession(authOpts)
  
  if (!session?.user) {
    throw new Error('Unauthorized')
  }
  
  return session
}

async function requirePermission(permission: string) {
  const session = await requireAuth()
  
  // Check if user has required permission
  // This is a simplified example - implement based on your auth system
  if (!session.user.permissions?.includes(permission)) {
    throw new Error('Insufficient permissions')
  }
  
  return session
}

// ===== SERVER ACTIONS =====

/**
 * Creates a new entity
 * @param input - Entity creation parameters
 * @returns Created entity or error
 */
export async function createEntity(input: CreateEntityInput): Promise<ActionResponse<Entity>> {
  try {
    // 1. Authenticate and authorize
    const session = await requireAuth()
    
    // 2. Validate input
    const validated = createEntitySchema.parse(input)
    
    // 3. Call backend service
    const client = getGrpcClient()
    const metadata = await getAuthMetadata()
    
    const response = await new Promise<any>((resolve, reject) => {
      client.createEntity(
        {
          name: validated.name,
          description: validated.description,
          metadata: validated.metadata,
        },
        metadata,
        (error, response) => {
          if (error) reject(error)
          else resolve(response)
        }
      )
    })
    
    // 4. Revalidate caches
    revalidatePath('/entities')
    revalidateTag('entities')
    
    // 5. Return success response
    return {
      success: true,
      data: {
        id: response.entity.id,
        name: response.entity.name,
        description: response.entity.description,
        status: response.entity.status,
        metadata: response.entity.metadata,
        createdAt: new Date(response.entity.createdAt),
        updatedAt: new Date(response.entity.updatedAt),
      }
    }
  } catch (error) {
    // Handle different error types
    if (error instanceof z.ZodError) {
      return {
        success: false,
        error: 'Validation failed',
        details: error.errors,
      }
    }
    
    if (error instanceof Error) {
      // Check for specific gRPC errors
      if (error.message.includes('already exists')) {
        return {
          success: false,
          error: 'An entity with this name already exists',
        }
      }
      
      return {
        success: false,
        error: error.message,
      }
    }
    
    return {
      success: false,
      error: 'An unexpected error occurred',
    }
  }
}

/**
 * Updates an existing entity
 */
export async function updateEntity(input: UpdateEntityInput): Promise<ActionResponse<Entity>> {
  try {
    const session = await requireAuth()
    const validated = updateEntitySchema.parse(input)
    
    const client = getGrpcClient()
    const metadata = await getAuthMetadata()
    
    const response = await new Promise<any>((resolve, reject) => {
      client.updateEntity(validated, metadata, (error, response) => {
        if (error) reject(error)
        else resolve(response)
      })
    })
    
    // Revalidate specific entity page and list
    revalidatePath(`/entities/${validated.id}`)
    revalidatePath('/entities')
    
    return {
      success: true,
      data: mapEntityFromGrpc(response.entity)
    }
  } catch (error) {
    return handleActionError(error)
  }
}

/**
 * Deletes an entity
 */
export async function deleteEntity(id: string): Promise<ActionResponse<{ deleted: boolean }>> {
  try {
    const session = await requirePermission('entities:delete')
    
    const client = getGrpcClient()
    const metadata = await getAuthMetadata()
    
    await new Promise<void>((resolve, reject) => {
      client.deleteEntity({ id }, metadata, (error) => {
        if (error) reject(error)
        else resolve()
      })
    })
    
    // Revalidate caches
    revalidatePath('/entities')
    revalidateTag('entities')
    
    return {
      success: true,
      data: { deleted: true }
    }
  } catch (error) {
    return handleActionError(error)
  }
}

/**
 * Lists entities with pagination and filtering
 */
export async function listEntities(options?: Partial<ListOptions>): Promise<ActionResponse<PaginatedResponse<Entity>>> {
  try {
    const session = await requireAuth()
    const validated = listOptionsSchema.parse(options || {})
    
    const client = getGrpcClient()
    const metadata = await getAuthMetadata()
    
    const response = await new Promise<any>((resolve, reject) => {
      client.listEntities(
        {
          page: validated.page,
          pageSize: validated.pageSize,
          sortBy: validated.sortBy,
          sortOrder: validated.sortOrder.toUpperCase(),
          status: validated.status,
          searchQuery: validated.search,
        },
        metadata,
        (error, response) => {
          if (error) reject(error)
          else resolve(response)
        }
      )
    })
    
    return {
      success: true,
      data: {
        items: response.entities.map(mapEntityFromGrpc),
        pageInfo: {
          totalItems: response.pageInfo.totalItems,
          totalPages: response.pageInfo.totalPages,
          currentPage: response.pageInfo.currentPage,
          pageSize: response.pageInfo.pageSize,
          hasNext: response.pageInfo.hasNext,
          hasPrevious: response.pageInfo.hasPrevious,
        }
      }
    }
  } catch (error) {
    return handleActionError(error)
  }
}

/**
 * Gets a single entity by ID
 */
export async function getEntity(id: string): Promise<ActionResponse<Entity>> {
  try {
    const session = await requireAuth()
    
    const client = getGrpcClient()
    const metadata = await getAuthMetadata()
    
    const response = await new Promise<any>((resolve, reject) => {
      client.getEntity({ identifier: id }, metadata, (error, response) => {
        if (error) reject(error)
        else resolve(response)
      })
    })
    
    return {
      success: true,
      data: mapEntityFromGrpc(response.entity)
    }
  } catch (error) {
    return handleActionError(error)
  }
}

// ===== UTILITY FUNCTIONS =====

function mapEntityFromGrpc(grpcEntity: any): Entity {
  return {
    id: grpcEntity.id,
    name: grpcEntity.name,
    description: grpcEntity.description || undefined,
    status: grpcEntity.status.toLowerCase().replace('entity_status_', ''),
    metadata: grpcEntity.metadata || undefined,
    createdAt: new Date(grpcEntity.createdAt),
    updatedAt: new Date(grpcEntity.updatedAt),
  }
}

function handleActionError(error: unknown): ActionResponse<any> {
  if (error instanceof z.ZodError) {
    return {
      success: false,
      error: 'Validation failed',
      details: error.errors,
    }
  }
  
  if (error instanceof Error) {
    // Map common gRPC errors
    if (error.message.includes('not found')) {
      return { success: false, error: 'Entity not found' }
    }
    if (error.message.includes('permission denied')) {
      return { success: false, error: 'You do not have permission to perform this action' }
    }
    if (error.message.includes('already exists')) {
      return { success: false, error: 'An entity with this identifier already exists' }
    }
    
    return { success: false, error: error.message }
  }
  
  return { success: false, error: 'An unexpected error occurred' }
}