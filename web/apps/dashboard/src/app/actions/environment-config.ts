'use server'

import { environmentRepository, configMapRepository, secretRepository } from '@kloudlite/lib/k8s'
import { environmentService } from '@/lib/services/environment.service'
import { environmentNameSchema, envVarSchema, fileSchema } from '@/lib/validations'
import { getSession } from '@/lib/get-session'
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
  EnvVar,
} from '@kloudlite/types'

/**
 * Get the environment's target namespace where configs/secrets are stored
 */
async function getEnvironmentNamespace(environmentName: string): Promise<string> {
  const session = await getSession()
  if (!session?.user?.username) {
    throw new Error('Not authenticated')
  }

  // Get the work machine namespace
  const namespace = `wm-${session.user.username}`

  // Get the environment to find its target namespace
  const environment = await environmentRepository.get(namespace, environmentName)

  return environment.spec?.targetNamespace || namespace
}

// Config actions
export async function getConfig(environmentName: string): Promise<GetConfigResponse> {
  return environmentService.getConfig(environmentName)
}

export async function setConfig(
  environmentName: string,
  data: ConfigData,
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
  data: SecretData,
): Promise<SetSecretResponse> {
  return environmentService.setSecret(environmentName, data)
}

export async function deleteSecret(environmentName: string): Promise<DeleteSecretResponse> {
  return environmentService.deleteSecret(environmentName)
}

// EnvVars actions (unified configs + secrets)
export async function getEnvVars(environmentName: string): Promise<GetEnvVarsResponse> {
  try {
    const targetNamespace = await getEnvironmentNamespace(environmentName)

    // Fetch ConfigMaps and Secrets in parallel from the target namespace
    const [configMaps, secrets] = await Promise.all([
      configMapRepository.list(targetNamespace).catch(() => []),
      secretRepository.list(targetNamespace).catch(() => []),
    ])

    // Convert ConfigMaps to EnvVars with type 'config'
    const configEnvVars: EnvVar[] = []
    for (const cm of configMaps) {
      const data = cm.data || {}
      for (const [key, value] of Object.entries(data)) {
        configEnvVars.push({
          key,
          value: value as string,
          type: 'config',
        })
      }
    }

    // Convert Secrets to EnvVars with type 'secret'
    const secretEnvVars: EnvVar[] = []
    for (const secret of secrets) {
      const data = secret.data || {}
      for (const [key, value] of Object.entries(data)) {
        // Secrets are base64 encoded in Kubernetes
        const decodedValue = Buffer.from(value as string, 'base64').toString('utf-8')
        secretEnvVars.push({
          key,
          value: decodedValue,
          type: 'secret',
        })
      }
    }

    // Combine and return
    const allEnvVars = [...configEnvVars, ...secretEnvVars]

    return {
      envVars: allEnvVars,
      count: allEnvVars.length,
    }
  } catch (error) {
    console.error('Get env vars error:', error)
    throw error
  }
}

export async function createEnvVar(
  environmentName: string,
  key: string,
  value: string,
  type: 'config' | 'secret',
): Promise<SetEnvVarResponse> {
  // Validate environment name
  const nameValidation = environmentNameSchema.safeParse(environmentName)
  if (!nameValidation.success) {
    throw new Error('Invalid environment name')
  }

  // Validate env var data
  const validated = envVarSchema.safeParse({ key, value, type })
  if (!validated.success) {
    throw new Error(validated.error.errors.map((e) => e.message).join(', '))
  }

  return environmentService.createEnvVar(environmentName, validated.data.key, validated.data.value, validated.data.type)
}

export async function setEnvVar(
  environmentName: string,
  key: string,
  value: string,
  type: 'config' | 'secret',
): Promise<SetEnvVarResponse> {
  // Validate environment name
  const nameValidation = environmentNameSchema.safeParse(environmentName)
  if (!nameValidation.success) {
    throw new Error('Invalid environment name')
  }

  // Validate env var data
  const validated = envVarSchema.safeParse({ key, value, type })
  if (!validated.success) {
    throw new Error(validated.error.errors.map((e) => e.message).join(', '))
  }

  return environmentService.setEnvVar(environmentName, validated.data.key, validated.data.value, validated.data.type)
}

export async function deleteEnvVar(
  environmentName: string,
  key: string,
): Promise<DeleteEnvVarResponse> {
  return environmentService.deleteEnvVar(environmentName, key)
}

// File actions
export async function listFiles(environmentName: string): Promise<ListFilesResponse> {
  return environmentService.listFiles(environmentName)
}

export async function getFile(environmentName: string, filename: string): Promise<GetFileResponse> {
  return environmentService.getFile(environmentName, filename)
}

export async function setFile(
  environmentName: string,
  filename: string,
  content: string,
): Promise<SetFileResponse> {
  // Validate environment name
  const nameValidation = environmentNameSchema.safeParse(environmentName)
  if (!nameValidation.success) {
    throw new Error('Invalid environment name')
  }

  // Validate file data
  const validated = fileSchema.safeParse({ filename, content })
  if (!validated.success) {
    throw new Error(validated.error.errors.map((e) => e.message).join(', '))
  }

  return environmentService.setFile(environmentName, validated.data.filename, validated.data.content)
}

export async function deleteFile(
  environmentName: string,
  filename: string,
): Promise<DeleteFileResponse> {
  return environmentService.deleteFile(environmentName, filename)
}
