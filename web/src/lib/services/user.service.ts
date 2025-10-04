import { apiClient } from '@/lib/api-client'

export interface User {
  metadata: {
    name: string
    namespace?: string
    uid?: string
    resourceVersion?: string
    generation?: number
    creationTimestamp?: string
  }
  spec: {
    email: string
    displayName?: string
    roles: string[]
    isActive?: boolean
    providers?: Array<{
      provider: string
      providerId: string
      email: string
      name?: string | null
      image?: string | null
      connectedAt: string
    }>
    metadata?: Record<string, string>
  }
  status?: {
    isReady?: boolean
    message?: string
    lastLogin?: string
  }
}

export interface CreateUserRequest {
  email: string
  displayName?: string
  roles: string[]
  isActive?: boolean
}

export interface UpdateUserRequest {
  email?: string
  displayName?: string
  roles?: string[]
  isActive?: boolean
}

interface UserList {
  metadata: {
    resourceVersion: string
  }
  items: User[]
}

export interface UserService {
  listUsers(): Promise<User[]>
  getUserByEmail(email: string): Promise<User>
  getUserByName(name: string): Promise<User>
  createUser(data: CreateUserRequest): Promise<User>
  updateUser(name: string, data: UpdateUserRequest): Promise<User>
  deleteUser(name: string): Promise<void>
  activateUser(name: string): Promise<User>
  deactivateUser(name: string): Promise<User>
}

class UserServiceImpl implements UserService {
  async listUsers(): Promise<User[]> {
    const response = await apiClient.get<UserList>('/api/v1/users')
    return response.items || []
  }

  async getUserByEmail(email: string): Promise<User> {
    const response = await apiClient.get<User>(`/api/v1/users/by-email?email=${encodeURIComponent(email)}`)
    return response
  }

  async getUserByName(name: string): Promise<User> {
    const response = await apiClient.get<User>(`/api/v1/users/${name}`)
    return response
  }

  async createUser(data: CreateUserRequest): Promise<User> {
    const response = await apiClient.post<User>('/api/v1/users', data)
    return response
  }

  async updateUser(name: string, data: UpdateUserRequest): Promise<User> {
    const response = await apiClient.put<User>(`/api/v1/users/${name}`, data)
    return response
  }

  async deleteUser(name: string): Promise<void> {
    await apiClient.delete(`/api/v1/users/${name}`)
  }

  async activateUser(name: string): Promise<User> {
    const response = await apiClient.post<User>(`/api/v1/users/${name}/activate`)
    return response
  }

  async deactivateUser(name: string): Promise<User> {
    const response = await apiClient.post<User>(`/api/v1/users/${name}/deactivate`)
    return response
  }
}

export const userService = new UserServiceImpl()