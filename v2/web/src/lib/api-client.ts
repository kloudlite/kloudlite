import { env } from '@/lib/env'

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

// Export singleton instance
export const apiClient = new ApiClient()