import { describe, expect, test } from 'bun:test'
import { mapEnvironmentResource, useEnvironmentStore } from './environments'

describe('environment store helpers', () => {
  test('maps Kubernetes deletionTimestamp to deleting status', () => {
    const env = mapEnvironmentResource(
      {
        metadata: {
          name: 'main',
          namespace: 'wm-karthik-dev',
          deletionTimestamp: '2026-05-31T04:56:43Z',
        },
        spec: {
          activated: true,
          ownedBy: 'karthik',
        },
      },
      'wm-karthik-dev'
    )

    expect(env.status).toBe('deleting')
  })

  test('delete refreshes silently so the environment list does not flicker', async () => {
    let resolveList!: (value: { items: any[] }) => void
    let listStarted!: () => void
    const listStartedPromise = new Promise<void>((resolve) => {
      listStarted = resolve
    })

    ;(globalThis as any).window = {
      electronAPI: {
        deleteEnvironment: async () => ({}),
        listEnvironments: () => {
          listStarted()
          return new Promise((resolve) => {
            resolveList = resolve
          })
        },
      },
    }

    useEnvironmentStore.setState({
      environments: [mapEnvironmentResource({ metadata: { name: 'main' }, spec: { activated: true } }, 'wm-karthik-dev')],
      loading: false,
      refreshing: false,
      error: null,
      deletingEnvs: new Set(),
    })

    const deletePromise = useEnvironmentStore.getState().deleteEnvironment('wm-karthik-dev', 'main')
    await listStartedPromise

    expect(useEnvironmentStore.getState().loading).toBe(false)
    expect(useEnvironmentStore.getState().refreshing).toBe(true)

    resolveList({
      items: [
        {
          metadata: { name: 'main', deletionTimestamp: '2026-05-31T04:56:43Z' },
          spec: { activated: true },
        },
      ],
    })
    await deletePromise
  })

  test('create refreshes silently so the existing environment list does not flicker', async () => {
    let resolveList!: (value: { items: any[] }) => void
    let listStarted!: () => void
    const listStartedPromise = new Promise<void>((resolve) => {
      listStarted = resolve
    })

    ;(globalThis as any).window = {
      electronAPI: {
        createEnvironment: async () => ({}),
        listEnvironments: () => {
          listStarted()
          return new Promise((resolve) => {
            resolveList = resolve
          })
        },
      },
    }

    useEnvironmentStore.setState({
      environments: [mapEnvironmentResource({ metadata: { name: 'main' }, spec: { activated: true } }, 'wm-karthik-dev')],
      loading: false,
      refreshing: false,
      error: null,
      deletingEnvs: new Set(),
    })

    const createPromise = useEnvironmentStore.getState().createEnvironment('wm-karthik-dev', 'preview', {})
    await listStartedPromise

    expect(useEnvironmentStore.getState().loading).toBe(false)
    expect(useEnvironmentStore.getState().refreshing).toBe(true)

    resolveList({
      items: [
        { metadata: { name: 'main' }, spec: { activated: true } },
        { metadata: { name: 'preview' }, spec: { activated: true } },
      ],
    })
    await createPromise
  })
})
