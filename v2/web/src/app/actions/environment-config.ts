'use server'

import { environmentService } from '@/lib/services/environment.service'
import type {
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
} from '@/types/environment'

// Config actions
export async function getConfig(environmentName: string): Promise<GetConfigResponse> {
  return environmentService.getConfig(environmentName)
}

export async function setConfig(
  environmentName: string,
  data: ConfigData
): Promise<SetConfigResponse> {
  return environmentService.setConfig(environmentName, data)
}

export async function deleteConfig(environmentName: string): Promise<DeleteConfigResponse> {
  return environmentService.deleteConfig(environmentName)
}

// Secret actions
export async function getSecret(environmentName: string): Promise<GetSecretResponse> {
  return environmentService.getSecret(environmentName)
}

export async function setSecret(
  environmentName: string,
  data: SecretData
): Promise<SetSecretResponse> {
  return environmentService.setSecret(environmentName, data)
}

export async function deleteSecret(environmentName: string): Promise<DeleteSecretResponse> {
  return environmentService.deleteSecret(environmentName)
}

// EnvVars actions (unified configs + secrets)
export async function getEnvVars(environmentName: string): Promise<GetEnvVarsResponse> {
  return environmentService.getEnvVars(environmentName)
}

export async function createEnvVar(
  environmentName: string,
  key: string,
  value: string,
  type: 'config' | 'secret'
): Promise<SetEnvVarResponse> {
  return environmentService.createEnvVar(environmentName, key, value, type)
}

export async function setEnvVar(
  environmentName: string,
  key: string,
  value: string,
  type: 'config' | 'secret'
): Promise<SetEnvVarResponse> {
  return environmentService.setEnvVar(environmentName, key, value, type)
}

export async function deleteEnvVar(
  environmentName: string,
  key: string
): Promise<DeleteEnvVarResponse> {
  return environmentService.deleteEnvVar(environmentName, key)
}

// File actions
export async function listFiles(environmentName: string): Promise<ListFilesResponse> {
  return environmentService.listFiles(environmentName)
}

export async function getFile(
  environmentName: string,
  filename: string
): Promise<GetFileResponse> {
  return environmentService.getFile(environmentName, filename)
}

export async function setFile(
  environmentName: string,
  filename: string,
  content: string
): Promise<SetFileResponse> {
  return environmentService.setFile(environmentName, filename, content)
}

export async function deleteFile(
  environmentName: string,
  filename: string
): Promise<DeleteFileResponse> {
  return environmentService.deleteFile(environmentName, filename)
}
