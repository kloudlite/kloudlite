import { describe, expect, test } from 'bun:test'
import { mapWorkspaceResource, useWorkspaceStore } from './workspaces'

describe('workspace store helpers', () => {
  test('maps metadata name to workspace identity', () => {
    const workspace = mapWorkspaceResource(
      {
        metadata: { name: 'api-dev', namespace: 'wm-karthik-dev' },
        spec: { ownedBy: 'karthik', activated: true },
      },
      'wm-karthik-dev'
    )

    expect(workspace.id).toBe('api-dev')
    expect(workspace.name).toBe('api-dev')
    expect(workspace.slug).toBe('api-dev')
    expect(workspace.namespace).toBe('wm-karthik-dev')
  })

  test('maps Kubernetes deletionTimestamp to deleting status', () => {
    const workspace = mapWorkspaceResource(
      {
        metadata: {
          name: 'api-dev',
          namespace: 'wm-karthik-dev',
          deletionTimestamp: '2026-05-31T06:00:00Z',
        },
        spec: { activated: true },
      },
      'wm-karthik-dev'
    )

    expect(workspace.status).toBe('deleting')
  })

  test('maps active workspace with connected environment from API fields', () => {
    const workspace = mapWorkspaceResource(
      {
        metadata: { name: 'main', namespace: 'wm-karthik-dev' },
        spec: {
          status: 'active',
          ownedBy: 'karthik',
          environmentConnection: {
            environmentRef: { name: 'dev-env' },
          },
        },
        status: { phase: 'Running' },
      },
      'wm-karthik-dev'
    )

    expect(workspace.status).toBe('running')
    expect(workspace.environmentName).toBe('dev-env')
  })

  test('create refreshes silently so the existing workspace list does not flicker', async () => {
    let createdSpec: Record<string, unknown> | undefined
    let resolveList!: (value: { items: any[] }) => void
    let listStarted!: () => void
    const listStartedPromise = new Promise<void>((resolve) => {
      listStarted = resolve
    })

    ;(globalThis as any).window = {
      electronAPI: {
        createWorkspace: async (_namespace: string, _name: string, spec: Record<string, unknown>) => {
          createdSpec = spec
          return {}
        },
        listWorkspaces: () => {
          listStarted()
          return new Promise((resolve) => {
            resolveList = resolve
          })
        },
      },
    }

    useWorkspaceStore.setState({
      workspaces: [mapWorkspaceResource({ metadata: { name: 'api-dev' }, spec: { activated: true } }, 'wm-karthik-dev')],
      loading: false,
      refreshing: false,
      error: null,
      deletingWorkspaces: new Set(),
    })

    const createPromise = useWorkspaceStore.getState().createWorkspace('wm-karthik-dev', 'frontend-dev', {
      visibility: 'private',
      gitRepository: { url: 'github.com/kloudlite/api' },
    })
    await listStartedPromise

    expect(useWorkspaceStore.getState().loading).toBe(false)
    expect(useWorkspaceStore.getState().refreshing).toBe(true)
    expect(createdSpec).toEqual({
      displayName: 'frontend-dev',
      ownedBy: 'karthik',
      workmachine: 'karthik-dev',
      status: 'active',
      visibility: 'private',
      gitRepository: { url: 'github.com/kloudlite/api' },
    })

    resolveList({
      items: [
        { metadata: { name: 'api-dev' }, spec: { activated: true } },
        { metadata: { name: 'frontend-dev' }, spec: { activated: true } },
      ],
    })
    await createPromise
  })

  test('delete refreshes silently and tracks deleting workspaces', async () => {
    let resolveList!: (value: { items: any[] }) => void
    let listStarted!: () => void
    const listStartedPromise = new Promise<void>((resolve) => {
      listStarted = resolve
    })

    ;(globalThis as any).window = {
      electronAPI: {
        deleteWorkspace: async () => ({}),
        listWorkspaces: () => {
          listStarted()
          return new Promise((resolve) => {
            resolveList = resolve
          })
        },
      },
    }

    useWorkspaceStore.setState({
      workspaces: [mapWorkspaceResource({ metadata: { name: 'api-dev' }, spec: { activated: true } }, 'wm-karthik-dev')],
      loading: false,
      refreshing: false,
      error: null,
      deletingWorkspaces: new Set(),
    })

    const deletePromise = useWorkspaceStore.getState().deleteWorkspace('wm-karthik-dev', 'api-dev')
    await listStartedPromise

    expect(useWorkspaceStore.getState().loading).toBe(false)
    expect(useWorkspaceStore.getState().refreshing).toBe(true)
    expect(useWorkspaceStore.getState().deletingWorkspaces.has('api-dev')).toBe(true)

    resolveList({
      items: [
        {
          metadata: { name: 'api-dev', deletionTimestamp: '2026-05-31T06:00:00Z' },
          spec: { activated: true },
        },
      ],
    })
    await deletePromise

    expect(useWorkspaceStore.getState().deletingWorkspaces.has('api-dev')).toBe(false)
    expect(useWorkspaceStore.getState().workspaces[0].status).toBe('deleting')
  })
})
