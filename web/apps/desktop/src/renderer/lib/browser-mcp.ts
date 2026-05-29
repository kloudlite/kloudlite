import { webviewRegistry } from './webview-registry'
import { useTabStore } from '@/store/tabs'

function getHandle(tabId?: string) {
  if (tabId) return webviewRegistry.get(tabId)
  const activeId = useTabStore.getState().activeTabId
  if (activeId) return webviewRegistry.get(activeId)
  return webviewRegistry.getFirst()
}

export async function executeBrowserCommand(command: string, args: Record<string, unknown>): Promise<unknown> {
  switch (command) {
    case 'list_tabs': {
      const tabs = webviewRegistry.getAll()
      const storeTabs = useTabStore.getState()
      return tabs.map((t) => ({
        ...t,
        isActive: t.id === storeTabs.activeTabId
      }))
    }

    case 'new_tab': {
      const url = (args.url as string) || 'https://google.com'
      useTabStore.getState().addTab(url)
      return `Opened new tab at ${url}`
    }

    case 'close_tab': {
      const tabId = args.tabId as string
      if (!tabId) return 'No tabId provided'
      useTabStore.getState().closeTab(tabId)
      return `Closed tab ${tabId}`
    }

    case 'activate_tab': {
      const tabId = args.tabId as string
      if (!tabId) return 'No tabId provided'
      useTabStore.getState().setActiveTab(tabId)
      return `Switched to tab ${tabId}`
    }

    case 'browser_navigate': {
      const url = args.url as string
      if (!url) return 'No URL provided'
      const handle = getHandle(args.tabId as string)
      if (!handle) return 'No browser tab available'
      await handle.navigate(url)
      return `Navigated to ${url}`
    }

    case 'browser_go_back': {
      const handle = getHandle(args.tabId as string)
      if (!handle) return 'No browser tab available'
      handle.goBack()
      return 'Went back'
    }

    case 'browser_go_forward': {
      const handle = getHandle(args.tabId as string)
      if (!handle) return 'No browser tab available'
      handle.goForward()
      return 'Went forward'
    }

    case 'browser_reload': {
      const handle = getHandle(args.tabId as string)
      if (!handle) return 'No browser tab available'
      handle.reload()
      return 'Reloaded'
    }

    case 'browser_screenshot': {
      const handle = getHandle(args.tabId as string)
      if (!handle) return 'No browser tab available'
      const dataUrl = await handle.capturePage()
      return { type: 'image', data: dataUrl, mimeType: 'image/png' }
    }

    case 'browser_click': {
      const selector = args.selector as string
      if (!selector) return 'No selector provided'
      const handle = getHandle(args.tabId as string)
      if (!handle) return 'No browser tab available'
      await handle.executeJavaScript(`
        (() => {
          const el = document.querySelector(${JSON.stringify(selector)})
          if (!el) throw new Error('Element not found: ${selector}')
          el.click()
          return 'Clicked ' + ${JSON.stringify(selector)}
        })()
      `)
      return `Clicked ${selector}`
    }

    case 'browser_fill': {
      const selector = args.selector as string
      const value = args.value as string
      if (!selector || value === undefined) return 'Missing selector or value'
      const handle = getHandle(args.tabId as string)
      if (!handle) return 'No browser tab available'
      await handle.executeJavaScript(`
        (() => {
          const el = document.querySelector(${JSON.stringify(selector)})
          if (!el) throw new Error('Element not found: ${selector}')
          if (el instanceof HTMLInputElement || el instanceof HTMLTextAreaElement) {
            el.value = ${JSON.stringify(value)}
            el.dispatchEvent(new Event('input', { bubbles: true }))
            el.dispatchEvent(new Event('change', { bubbles: true }))
            return 'Filled ' + ${JSON.stringify(selector)}
          }
          throw new Error('Element is not an input: ${selector}')
        })()
      `)
      return `Filled ${selector}`
    }

    case 'browser_get_text': {
      const selector = args.selector as string
      const handle = getHandle(args.tabId as string)
      if (!handle) return 'No browser tab available'
      const result = await handle.executeJavaScript(`
        (() => {
          ${selector
            ? `const el = document.querySelector(${JSON.stringify(selector)})
               if (!el) throw new Error('Element not found: ${selector}')
               return el.innerText || el.textContent || ''`
            : 'return document.body.innerText || document.body.textContent || \'\''
          }
        })()
      `)
      return String(result)
    }

    case 'browser_evaluate': {
      const script = args.script as string
      if (!script) return 'No script provided'
      const handle = getHandle(args.tabId as string)
      if (!handle) return 'No browser tab available'
      const result = await handle.executeJavaScript(script)
      return result !== undefined ? String(result) : 'undefined'
    }

    case 'browser_get_url': {
      const handle = getHandle(args.tabId as string)
      if (!handle) return ''
      return handle.getURL()
    }

    case 'browser_get_title': {
      const handle = getHandle(args.tabId as string)
      if (!handle) return ''
      return handle.getTitle()
    }

    default:
      throw new Error(`Unknown command: ${command}`)
  }
}
