import { apiClient } from '@/lib/api-client'
import type {
  Environment,
  EnvironmentCreateRequest,
  EnvironmentUpdateRequest,
  EnvironmentListResponse,
  EnvironmentResponse,
  EnvironmentDeleteResponse,
  EnvironmentStatusResponse,
  ConfigData,
  SetConfigResponse,
  GetConfigResponse,
  DeleteConfigResponse,
  SecretData,
  SetSecretResponse,
  GetSecretResponse,
  DeleteSecretResponse,
  SetFileResponse,
  GetFileResponse,
  ListFilesResponse,
  DeleteFileResponse,
  GetEnvVarsResponse,
  SetEnvVarResponse,
  DeleteEnvVarResponse,
} from '@kloudlite/types'

export class EnvironmentService {
  private baseUrl = '/api/v1/environments'

  /**
   * List all environments
   */
  async listEnvironments(): Promise<EnvironmentListResponse> {
    return apiClient.get<EnvironmentListResponse>(this.baseUrl)
  }

  /**
   * Get a specific environment by name
   */
  async getEnvironment(name: string): Promise<Environment> {
    const response = await apiClient.get<Environment>(`${this.baseUrl}/${name}`)
    return response
  }

  /**
   * Create a new environment
   */
  async createEnvironment(data: EnvironmentCreateRequest): Promise<EnvironmentResponse> {
    return apiClient.post<EnvironmentResponse>(this.baseUrl, data)
  }

  /**
   * Update an environment
   */
  async updateEnvironment(
    name: string,
    data: EnvironmentUpdateRequest,
  ): Promise<EnvironmentResponse> {
    return apiClient.put<EnvironmentResponse>(`${this.baseUrl}/${name}`, data)
  }

  /**
   * Delete an environment
   */
  async deleteEnvironment(name: string): Promise<EnvironmentDeleteResponse> {
    return apiClient.delete<EnvironmentDeleteResponse>(`${this.baseUrl}/${name}`)
  }

  /**
   * Activate an environment
   */
  async activateEnvironment(name: string): Promise<EnvironmentResponse> {
    return apiClient.post<EnvironmentResponse>(`${this.baseUrl}/${name}/activate`, undefined)
  }

  /**
   * Deactivate an environment
   */
  async deactivateEnvironment(name: string): Promise<EnvironmentResponse> {
    return apiClient.post<EnvironmentResponse>(`${this.baseUrl}/${name}/deactivate`, undefined)
  }

  /**
   * Get environment status
   */
  async getEnvironmentStatus(name: string): Promise<EnvironmentStatusResponse> {
    return apiClient.get<EnvironmentStatusResponse>(`${this.baseUrl}/${name}/status`)
  }

  // Config operations
  async setConfig(name: string, data: ConfigData): Promise<SetConfigResponse> {
    return apiClient.put(`${this.baseUrl}/${name}/config`, { data })
  }

  async getConfig(name: string): Promise<GetConfigResponse> {
    return apiClient.get(`${this.baseUrl}/${name}/config`)
  }

  async deleteConfig(name: string): Promise<DeleteConfigResponse> {
    return apiClient.delete(`${this.baseUrl}/${name}/config`)
  }

  // Secret operations
  async setSecret(name: string, data: SecretData): Promise<SetSecretResponse> {
    return apiClient.put(`${this.baseUrl}/${name}/secret`, { data })
  }

  async getSecret(name: string): Promise<GetSecretResponse> {
    return apiClient.get(`${this.baseUrl}/${name}/secret`)
  }

  async deleteSecret(name: string): Promise<DeleteSecretResponse> {
    return apiClient.delete(`${this.baseUrl}/${name}/secret`)
  }

  // EnvVars operations (unified configs + secrets)
  async getEnvVars(name: string): Promise<GetEnvVarsResponse> {
    return apiClient.get(`${this.baseUrl}/${name}/envvars`)
  }

  async createEnvVar(
    name: string,
    key: string,
    value: string,
    type: 'config' | 'secret',
  ): Promise<SetEnvVarResponse> {
    return apiClient.post(`${this.baseUrl}/${name}/envvars`, { key, value, type })
  }

  async setEnvVar(
    name: string,
    key: string,
    value: string,
    type: 'config' | 'secret',
  ): Promise<SetEnvVarResponse> {
    return apiClient.put(`${this.baseUrl}/${name}/envvars`, { key, value, type })
  }

  async deleteEnvVar(name: string, key: string): Promise<DeleteEnvVarResponse> {
    return apiClient.delete(`${this.baseUrl}/${name}/envvars/${key}`)
  }

  // File operations
  async setFile(name: string, filename: string, content: string): Promise<SetFileResponse> {
    return apiClient.put(`${this.baseUrl}/${name}/files/${filename}`, { content })
  }

  async getFile(name: string, filename: string): Promise<GetFileResponse> {
    return apiClient.get(`${this.baseUrl}/${name}/files/${filename}`)
  }

  async listFiles(name: string): Promise<ListFilesResponse> {
    return apiClient.get(`${this.baseUrl}/${name}/files`)
  }

  async deleteFile(name: string, filename: string): Promise<DeleteFileResponse> {
    return apiClient.delete(`${this.baseUrl}/${name}/files/${filename}`)
  }

  /**
   * Clone an environment by creating a new environment with cloneFrom field
   */
  async cloneEnvironment(
    sourceName: string,
    targetName: string,
    targetNamespace: string,
    _cloneEnvVars: boolean,
    _cloneFiles: boolean,
    currentUser: string,
  ): Promise<EnvironmentResponse> {
    const request: EnvironmentCreateRequest = {
      name: targetName,
      spec: {
        targetNamespace,
        ownedBy: currentUser,
        activated: false,
        cloneFrom: sourceName,
      },
    }
    return this.createEnvironment(request)
  }
}

// Export singleton instance
export const environmentService = new EnvironmentService()
