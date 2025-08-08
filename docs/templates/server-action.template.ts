'use server'

import { getServerSession } from 'next-auth'
import { getAuthOptions } from '@/lib/auth/get-auth-options'
import { get{{SERVICE}}Client, getAuthMetadata } from '@/lib/grpc/{{SERVICE_LOWER}}-client'
import { revalidatePath, revalidateTag } from 'next/cache'
import { z } from 'zod'

// ==================== Input Validation ====================

const {{ACTION_NAME}}Schema = z.object({
  // Add your validation schema here
  name: z.string().min(1).max(100),
  description: z.string().max(500).optional(),
  // Example for slug validation
  // slug: z.string().min(3).max(63).regex(/^[a-z0-9-]+$/),
})

// Type-safe input type
type {{ACTION_NAME_PASCAL}}Input = z.infer<typeof {{ACTION_NAME}}Schema>

// Response types
type {{ACTION_NAME_PASCAL}}Response = 
  | { success: true; data: { id: string; [key: string]: any } }
  | { success: false; error: string; details?: any }

// ==================== Main Action ====================

export async function {{ACTION_NAME}}(input: {{ACTION_NAME_PASCAL}}Input): Promise<{{ACTION_NAME_PASCAL}}Response> {
  try {
    // 1. Validate session
    const authOpts = await getAuthOptions()
    const session = await getServerSession(authOpts)
    
    if (!session?.user) {
      return { success: false, error: 'Unauthorized' }
    }

    // 2. Validate input
    const validated = {{ACTION_NAME}}Schema.parse(input)

    // 3. Additional authorization checks if needed
    // const hasPermission = await checkUserPermission(session.user.id, 'resource:action')
    // if (!hasPermission) {
    //   return { success: false, error: 'Insufficient permissions' }
    // }

    // 4. Call backend service
    const client = get{{SERVICE}}Client()
    const metadata = await getAuthMetadata()

    const response = await new Promise<any>((resolve, reject) => {
      client.{{GRPC_METHOD}}(
        {
          ...validated,
          // Map fields as needed for gRPC
        },
        metadata,
        (error, response) => {
          if (error) {
            reject(error)
          } else {
            resolve(response)
          }
        }
      )
    })

    // 5. Revalidate caches
    revalidatePath('/{{RESOURCE_PATH}}')
    revalidateTag('{{RESOURCE_TAG}}')
    
    // Add more specific paths/tags as needed
    // revalidatePath(`/{{RESOURCE_PATH}}/${response.id}`)

    // 6. Return success response
    return {
      success: true,
      data: {
        id: response.{{RESOURCE_LOWER}}.id,
        // Include other important fields
        name: response.{{RESOURCE_LOWER}}.name,
      }
    }
  } catch (error) {
    // 7. Error handling
    if (error instanceof z.ZodError) {
      return { 
        success: false, 
        error: 'Validation failed', 
        details: error.errors 
      }
    }

    if (error instanceof Error) {
      // Check for specific error types
      if (error.message.includes('already exists')) {
        return { success: false, error: 'Resource already exists' }
      }
      
      if (error.message.includes('not found')) {
        return { success: false, error: 'Resource not found' }
      }

      // Log unexpected errors
      console.error('{{ACTION_NAME}} error:', error)
      
      return { 
        success: false, 
        error: error.message || 'An error occurred' 
      }
    }

    return { 
      success: false, 
      error: 'An unexpected error occurred' 
    }
  }
}

// ==================== Related Actions ====================

// Example: List action
export async function list{{RESOURCE_PLURAL}}(
  options?: {
    page?: number
    pageSize?: number
    search?: string
    filters?: Record<string, any>
  }
): Promise<{ success: true; data: any[] } | { success: false; error: string }> {
  try {
    const authOpts = await getAuthOptions()
    const session = await getServerSession(authOpts)
    
    if (!session?.user) {
      return { success: false, error: 'Unauthorized' }
    }

    const client = get{{SERVICE}}Client()
    const metadata = await getAuthMetadata()

    const response = await new Promise<any>((resolve, reject) => {
      client.list{{RESOURCE_PLURAL}}(
        {
          page: options?.page || 1,
          pageSize: options?.pageSize || 20,
          search: options?.search || '',
          ...options?.filters,
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
      data: response.{{RESOURCE_LOWER_PLURAL}} || []
    }
  } catch (error) {
    console.error('list{{RESOURCE_PLURAL}} error:', error)
    return {
      success: false,
      error: error instanceof Error ? error.message : 'Failed to fetch {{RESOURCE_LOWER_PLURAL}}'
    }
  }
}

// Example: Get single resource
export async function get{{RESOURCE}}(id: string) {
  try {
    const authOpts = await getAuthOptions()
    const session = await getServerSession(authOpts)
    
    if (!session?.user) {
      return { success: false, error: 'Unauthorized' }
    }

    const client = get{{SERVICE}}Client()
    const metadata = await getAuthMetadata()

    const response = await new Promise<any>((resolve, reject) => {
      client.get{{RESOURCE}}(
        { id },
        metadata,
        (error, response) => {
          if (error) reject(error)
          else resolve(response)
        }
      )
    })

    return {
      success: true,
      data: response.{{RESOURCE_LOWER}}
    }
  } catch (error) {
    if (error instanceof Error && error.message.includes('not found')) {
      return { success: false, error: 'Resource not found' }
    }
    
    return {
      success: false,
      error: 'Failed to fetch resource'
    }
  }
}