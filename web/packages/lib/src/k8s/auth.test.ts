import { describe, expect, test, vi } from 'vitest'
import { getK8sApiUrl } from './auth'

describe('getK8sApiUrl', () => {
  test('uses HTTPS for the in-cluster Kubernetes API service', () => {
    vi.stubEnv('KUBERNETES_SERVICE_HOST', '10.43.0.1')
    vi.stubEnv('KUBERNETES_SERVICE_PORT', '443')

    expect(getK8sApiUrl(() => true)).toBe('https://10.43.0.1:443')

    vi.unstubAllEnvs()
  })
})
