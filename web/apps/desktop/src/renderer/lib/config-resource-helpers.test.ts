import { describe, expect, test } from 'bun:test'
import {
  buildConfigMap,
  buildEnvSecret,
  buildFileConfigMap,
  sanitizeConfigMapName,
} from './config-resource-helpers'

describe('config resource helpers', () => {
  test('builds canonical env configmap', () => {
    expect(buildConfigMap('env-demo', { NODE_ENV: 'dev' })).toEqual({
      apiVersion: 'v1',
      kind: 'ConfigMap',
      metadata: {
        name: 'env-config',
        namespace: 'env-demo',
        labels: {
          'kloudlite.io/config-type': 'envvars',
          'kloudlite.io/resource-type': 'config',
        },
      },
      data: { NODE_ENV: 'dev' },
    })
  })

  test('builds canonical env secret with stringData', () => {
    expect(buildEnvSecret('env-demo', { TOKEN: 'secret' })).toEqual({
      apiVersion: 'v1',
      kind: 'Secret',
      type: 'Opaque',
      metadata: {
        name: 'env-secret',
        namespace: 'env-demo',
        labels: {
          'kloudlite.io/config-type': 'envvars',
          'kloudlite.io/resource-type': 'secret',
        },
      },
      stringData: { TOKEN: 'secret' },
    })
  })

  test('builds stable file configmap names and labels', () => {
    expect(sanitizeConfigMapName('app.json')).toMatch(/^file-app-json-[a-z0-9]+$/)
    expect(sanitizeConfigMapName('a.b')).not.toBe(sanitizeConfigMapName('a/b'))
    expect(buildFileConfigMap('env-demo', 'app.json', '{ }')).toMatchObject({
      metadata: {
        name: sanitizeConfigMapName('app.json'),
        namespace: 'env-demo',
        labels: {
          'kloudlite.io/resource-type': 'file',
          'kloudlite.io/filename': 'app.json',
        },
      },
      data: { content: '{ }' },
    })
  })

  test('rejects filenames that cannot be stored in labels', () => {
    expect(() => buildFileConfigMap('env-demo', 'bad/name.json', '')).toThrow('Filename must')
    expect(() => buildFileConfigMap('env-demo', '.env', '')).toThrow('Filename must')
    expect(() => buildFileConfigMap('env-demo', 'a'.repeat(64), '')).toThrow('Filename must')
  })
})
