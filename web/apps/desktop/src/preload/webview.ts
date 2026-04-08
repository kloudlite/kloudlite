import { ipcRenderer } from 'electron'

// Right-click context menu
document.addEventListener('contextmenu', (e) => {
  e.preventDefault()
  ipcRenderer.sendToHost('context-menu', e.screenX, e.screenY)
})

// Simple swipe detection — no visual, just back/forward action
let accumulator = 0
let idleTimer: ReturnType<typeof setTimeout> | null = null

window.addEventListener('wheel', (e) => {
  if (Math.abs(e.deltaY) > Math.abs(e.deltaX) * 1.5) return
  if (Math.abs(e.deltaX) < 1) return

  accumulator += e.deltaX

  if (idleTimer) clearTimeout(idleTimer)
  idleTimer = setTimeout(() => {
    if (Math.abs(accumulator) > 80) {
      ipcRenderer.sendToHost('swipe-navigate', accumulator > 0 ? 'forward' : 'back')
    }
    accumulator = 0
  }, 60)
}, { passive: true })

function extractMetadata() {
  function getMeta(names: string[]): string {
    for (const name of names) {
      const el = document.querySelector(`meta[name="${name}"], meta[property="${name}"]`) as HTMLMetaElement | null
      if (el?.content) return el.content
    }
    return ''
  }

  return {
    title: document.title,
    description: getMeta(['description', 'og:description', 'twitter:description']),
    siteName: getMeta(['og:site_name', 'application-name', 'apple-mobile-web-app-title']),
    keywords: getMeta(['keywords']),
    author: getMeta(['author']),
    type: getMeta(['og:type']),
  }
}

document.addEventListener('DOMContentLoaded', () => {
  const titleObserver = new MutationObserver(() => {
    ipcRenderer.sendToHost('page-title-updated', document.title)
  })

  const titleEl = document.querySelector('title')
  if (titleEl) {
    titleObserver.observe(titleEl, { childList: true, characterData: true, subtree: true })
  }

  ipcRenderer.sendToHost('page-title-updated', document.title)

  // Send metadata after a short delay to let the page fully render
  setTimeout(() => {
    ipcRenderer.sendToHost('page-metadata', extractMetadata())
  }, 500)
})

// Also send metadata when the page finishes loading (for SPAs)
window.addEventListener('load', () => {
  setTimeout(() => {
    ipcRenderer.sendToHost('page-metadata', extractMetadata())
  }, 300)
})
