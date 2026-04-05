import { useCallback, useEffect, useRef } from 'react'
import { Sidebar } from '@/components/sidebar'
import { AddressBar } from '@/components/address-bar'
import { WebviewArea, type WebviewAreaHandle } from '@/components/webview-area'
import { useTabStore } from '@/store/tabs'

export function App() {
  const handleRef = useRef<WebviewAreaHandle | null>(null)
  const { addTab, closeTab, activeTabId, tabs, setActiveTab } = useTabStore()

  const setHandle = useCallback((handle: WebviewAreaHandle) => {
    handleRef.current = handle
  }, [])

  useEffect(() => {
    function onKeyDown(e: KeyboardEvent) {
      const mod = e.metaKey || e.ctrlKey

      if (mod && e.key === 't') {
        e.preventDefault()
        addTab()
      } else if (mod && e.key === 'w') {
        e.preventDefault()
        if (activeTabId) closeTab(activeTabId)
      } else if (mod && e.key === 'l') {
        e.preventDefault()
        const input = document.querySelector<HTMLInputElement>('input[type="text"]')
        input?.focus()
        input?.select()
      } else if (mod && e.key === 'r') {
        e.preventDefault()
        handleRef.current?.reload()
      } else if (mod && e.key === '[') {
        e.preventDefault()
        handleRef.current?.goBack()
      } else if (mod && e.key === ']') {
        e.preventDefault()
        handleRef.current?.goForward()
      } else if (mod && e.key >= '1' && e.key <= '9') {
        e.preventDefault()
        const index = parseInt(e.key) - 1
        const { tabs: currentTabs } = useTabStore.getState()
        if (index < currentTabs.length) {
          setActiveTab(currentTabs[index].id)
        }
      }
    }

    window.addEventListener('keydown', onKeyDown)
    return () => window.removeEventListener('keydown', onKeyDown)
  }, [addTab, closeTab, activeTabId, setActiveTab])

  useEffect(() => {
    if (tabs.length === 0) {
      addTab()
    }
  }, [])

  return (
    <div className="flex h-full bg-background text-foreground">
      <Sidebar />
      <div className="flex flex-1 flex-col">
        <div className="drag-region flex h-12 items-center" />
        <AddressBar
          onNavigate={(url) => handleRef.current?.navigate(url)}
          onGoBack={() => handleRef.current?.goBack()}
          onGoForward={() => handleRef.current?.goForward()}
          onReload={() => handleRef.current?.reload()}
        />
        <WebviewArea onHandle={setHandle} />
      </div>
    </div>
  )
}
