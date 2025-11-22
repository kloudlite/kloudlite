import { apiClient } from '@/lib/api-client'

export interface ConnectionToken {
  metadata: {
    name: string
    namespace?: string
    uid?: string
    resourceVersion?: string
    generation?: number
    creationTimestamp?: string
  }
  spec: {
    displayName: string
    userId: string
    sshJumpHost: string
    sshPort: number
    apiUrl: string
    expiresAt?: string
  }
  status?: {
    isReady?: boolean
    message?: string
    lastUsed?: string
    token?: string
  }
}

export interface CreateConnectionTokenRequest {
  displayName: string
  webUrl?: string
}

export interface ConnectionTokenResponse {
  token: ConnectionToken
  jwt: string
}

interface ConnectionTokenList {
  metadata: {
    resourceVersion: string
  }
  items: ConnectionToken[]
}

export interface ConnectionTokenService {
  listTokens(): Promise<ConnectionToken[]>
  createToken(data: CreateConnectionTokenRequest): Promise<ConnectionTokenResponse>
  deleteToken(name: string): Promise<void>
}

class ConnectionTokenServiceImpl implements ConnectionTokenService {
  async listTokens(): Promise<ConnectionToken[]> {
    const response = await apiClient.get<ConnectionTokenList>('/api/v1/connection-tokens')
    return response.items || []
  }

  async createToken(data: CreateConnectionTokenRequest): Promise<ConnectionTokenResponse> {
    const response = await apiClient.post<ConnectionTokenResponse>(
      '/api/v1/connection-tokens',
      data,
    )
    return response
  }

  async deleteToken(name: string): Promise<void> {
    await apiClient.delete(`/api/v1/connection-tokens/${name}`)
  }
}

export const connectionTokenService = new ConnectionTokenServiceImpl()
