import { env } from '@/lib/env'
import { auth } from '@/lib/auth'

// API client configuration
export class ApiClient {
  private baseUrl: string

  constructor() {
    this.baseUrl = env.apiUrl
  }

  private async request<T>(
    endpoint: string,
    options: RequestInit = {},
  ): Promise<T> {
    const url = `${this.baseUrl}${endpoint}`

    // Get authentication session (only works in server components/actions)
    let authHeaders: Record<string, string> = {}
    try {
      const session = await auth()
      if (session?.user?.backendToken) {
        authHeaders.Authorization = `Bearer ${session.user.backendToken}`
      }
    } catch (error) {
      // In client components or when no session, continue without auth
      console.debug('No session available for API request:', endpoint)
    }

    const config: RequestInit = {
      ...options,
      headers: {
        'Content-Type': 'application/json',
        ...authHeaders,
        ...options.headers,
      },
    }

    const response = await fetch(url, config)

    if (!response.ok) {
      const errorText = await response.text().catch(() => response.statusText)
      throw new Error(`API Error: ${response.status} ${errorText}`)
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

  post<T>(endpoint: string, data?: any, options?: RequestInit): Promise<T> {
    return this.request<T>(endpoint, {
      ...options,
      method: 'POST',
      body: data ? JSON.stringify(data) : undefined,
    })
  }

  put<T>(endpoint: string, data?: any, options?: RequestInit): Promise<T> {
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

  private async request<T>(
    endpoint: string,
    options: RequestInit = {},
  ): Promise<T> {
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
      throw new Error(`API Error: ${response.status} ${errorText}`)
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

  post<T>(endpoint: string, data?: any, options?: RequestInit): Promise<T> {
    return this.request<T>(endpoint, {
      ...options,
      method: 'POST',
      body: data ? JSON.stringify(data) : undefined,
    })
  }

  put<T>(endpoint: string, data?: any, options?: RequestInit): Promise<T> {
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