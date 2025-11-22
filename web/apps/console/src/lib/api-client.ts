import { env } from '@/lib/env'
import { getSession } from '@/lib/get-session'

// API client configuration
export class ApiClient {
  private baseUrl: string

  constructor() {
    this.baseUrl = env.apiUrl
  }

  private async request<T>(endpoint: string, options: RequestInit = {}): Promise<T> {
    const url = `${this.baseUrl}${endpoint}`

    // Get authentication session (only works in server components/actions)
    const authHeaders: Record<string, string> = {}
    try {
      const session = await getSession()
      if (session?.user?.backendToken) {
        authHeaders.Authorization = `Bearer ${session.user.backendToken}`
      }
    } catch (error) {
      console.error('[API Client] Error getting session for endpoint:', endpoint, error)
    }

    // Explicitly construct headers to ensure they're passed correctly
    const headers: Record<string, string> = {
      'Content-Type': 'application/json',
      ...authHeaders,
    }

    // Merge any additional headers from options
    if (options.headers) {
      const optionHeaders = options.headers as Record<string, string>
      Object.assign(headers, optionHeaders)
    }

    const config: RequestInit = {
      ...options,
      cache: 'no-store', // Prevent Next.js from caching and potentially stripping headers
      headers,
    }

    const response = await fetch(url, config)

    if (!response.ok) {
      const errorText = await response.text().catch(() => response.statusText)

      // Try to parse JSON error response
      try {
        const errorJson = JSON.parse(errorText)
        // Extract the most relevant error message
        const message = errorJson.error || errorJson.message || errorText
        throw new Error(message)
      } catch (_parseError) {
        // If not JSON, use the raw error text
        throw new Error(errorText || `Request failed with status ${response.status}`)
      }
    }

    // Handle empty responses (like 204 No Content)
    if (response.status === 204) {
      return {} as T
    }

    const text = await response.text()
    if (!text) {
      return {} as T
    }

    try {
      return JSON.parse(text)
    } catch {
      return text as unknown as T
    }
  }

  // HTTP methods
  get<T>(endpoint: string, options?: RequestInit): Promise<T> {
    return this.request<T>(endpoint, { ...options, method: 'GET' })
  }

  post<T>(endpoint: string, data?: unknown, options?: RequestInit): Promise<T> {
    return this.request<T>(endpoint, {
      ...options,
      method: 'POST',
      body: data ? JSON.stringify(data) : undefined,
    })
  }

  put<T>(endpoint: string, data?: unknown, options?: RequestInit): Promise<T> {
    return this.request<T>(endpoint, {
      ...options,
      method: 'PUT',
      body: data ? JSON.stringify(data) : undefined,
    })
  }

  delete<T>(endpoint: string, options?: RequestInit): Promise<T> {
    return this.request<T>(endpoint, { ...options, method: 'DELETE' })
  }
}

// Unauthenticated API client for auth endpoints
export class UnauthenticatedApiClient {
  private baseUrl: string

  constructor() {
    this.baseUrl = env.apiUrl
  }

  private async request<T>(endpoint: string, options: RequestInit = {}): Promise<T> {
    const url = `${this.baseUrl}${endpoint}`

    const config: RequestInit = {
      ...options,
      headers: {
        'Content-Type': 'application/json',
        ...options.headers,
      },
    }

    const response = await fetch(url, config)

    if (!response.ok) {
      const errorText = await response.text().catch(() => response.statusText)

      // Try to parse JSON error response
      try {
        const errorJson = JSON.parse(errorText)
        // Extract the most relevant error message
        const message = errorJson.error || errorJson.message || errorText
        throw new Error(message)
      } catch (_parseError) {
        // If not JSON, use the raw error text
        throw new Error(errorText || `Request failed with status ${response.status}`)
      }
    }

    // Handle empty responses (like 204 No Content)
    if (response.status === 204) {
      return {} as T
    }

    const text = await response.text()
    if (!text) {
      return {} as T
    }

    try {
      return JSON.parse(text)
    } catch {
      return text as unknown as T
    }
  }

  // HTTP methods
  get<T>(endpoint: string, options?: RequestInit): Promise<T> {
    return this.request<T>(endpoint, { ...options, method: 'GET' })
  }

  post<T>(endpoint: string, data?: unknown, options?: RequestInit): Promise<T> {
    return this.request<T>(endpoint, {
      ...options,
      method: 'POST',
      body: data ? JSON.stringify(data) : undefined,
    })
  }

  put<T>(endpoint: string, data?: unknown, options?: RequestInit): Promise<T> {
    return this.request<T>(endpoint, {
      ...options,
      method: 'PUT',
      body: data ? JSON.stringify(data) : undefined,
    })
  }

  delete<T>(endpoint: string, options?: RequestInit): Promise<T> {
    return this.request<T>(endpoint, { ...options, method: 'DELETE' })
  }
}

// Export singleton instances
export const apiClient = new ApiClient()
export const unauthenticatedApiClient = new UnauthenticatedApiClient()
