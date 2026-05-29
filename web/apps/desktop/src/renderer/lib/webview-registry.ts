interface WebviewHandle {
  navigate: (url: string) => Promise<void>
  goBack: () => void
  goForward: () => void
  reload: () => void
  executeJavaScript: (code: string) => Promise<unknown>
  capturePage: () => Promise<string>
  getURL: () => string
  getTitle: () => string
  canGoBack: () => boolean
  canGoForward: () => boolean
  getWebContentsId: () => number
}

const registry = new Map<string, WebviewHandle>()

export const webviewRegistry = {
  register(id: string, handle: WebviewHandle) {
    registry.set(id, handle)
  },
  unregister(id: string) {
    registry.delete(id)
  },
  get(id: string) {
    return registry.get(id)
  },
  getAll() {
    return Array.from(registry.entries()).map(([id, handle]) => ({
      id,
      url: handle.getURL(),
      title: handle.getTitle(),
      canGoBack: handle.canGoBack(),
      canGoForward: handle.canGoForward()
    }))
  },
  getFirst() {
    return registry.values().next().value ?? null
  }
}
