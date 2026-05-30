export const ENV_CONFIG_NAME = 'env-config'
export const ENV_SECRET_NAME = 'env-secret'
export const filenamePattern = /^[a-zA-Z0-9](?:[a-zA-Z0-9._-]{0,61}[a-zA-Z0-9])?$/

const envConfigLabels = {
  'kloudlite.io/config-type': 'envvars',
  'kloudlite.io/resource-type': 'config',
}

const envSecretLabels = {
  'kloudlite.io/config-type': 'envvars',
  'kloudlite.io/resource-type': 'secret',
}

export function sanitizeConfigMapName(filename: string): string {
  const hash = hashString(filename)
  const normalized = filename
    .toLowerCase()
    .replace(/[^a-z0-9-]+/g, '-')
    .replace(/^-+|-+$/g, '')
    .slice(0, 236)

  return `file-${normalized || 'config'}-${hash}`
}

function hashString(value: string): string {
  let hash = 5381
  for (let i = 0; i < value.length; i += 1) {
    hash = ((hash << 5) + hash + value.charCodeAt(i)) >>> 0
  }
  return hash.toString(36).slice(0, 8)
}

export function buildConfigMap(namespace: string, data: Record<string, string>) {
  return {
    apiVersion: 'v1',
    kind: 'ConfigMap',
    metadata: {
      name: ENV_CONFIG_NAME,
      namespace,
      labels: envConfigLabels,
    },
    data,
  }
}

export function buildEnvSecret(namespace: string, data: Record<string, string>) {
  return {
    apiVersion: 'v1',
    kind: 'Secret',
    type: 'Opaque',
    metadata: {
      name: ENV_SECRET_NAME,
      namespace,
      labels: envSecretLabels,
    },
    stringData: data,
  }
}

export function buildFileConfigMap(namespace: string, filename: string, content: string) {
  if (!filenamePattern.test(filename)) {
    throw new Error('Filename must be 1-63 label-safe characters, start/end with a letter or number, and contain only letters, numbers, dots, underscores, or dashes')
  }

  return {
    apiVersion: 'v1',
    kind: 'ConfigMap',
    metadata: {
      name: sanitizeConfigMapName(filename),
      namespace,
      labels: {
        'kloudlite.io/resource-type': 'file',
        'kloudlite.io/filename': filename,
      },
    },
    data: { content },
  }
}
