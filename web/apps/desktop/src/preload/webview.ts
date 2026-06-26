import { ipcRenderer } from 'electron'

// Right-click context menu — detect link anchor
function closestLink(target: EventTarget | null): HTMLAnchorElement | null {
  let el = target as HTMLElement | null
  while (el) {
    if (el.tagName === 'A') return el as HTMLAnchorElement
    el = el.parentElement
  }
  return null
}

document.addEventListener('contextmenu', (e) => {
  e.preventDefault()
  const link = closestLink(e.target)
  const href = link?.href || link?.getAttribute('href') || ''
  ipcRenderer.sendToHost('context-menu', e.screenX, e.screenY, href)
})

function isEditable(target: EventTarget | null) {
  const el = target as HTMLElement | null
  if (!el) return false
  return el.isContentEditable || ['INPUT', 'TEXTAREA', 'SELECT'].includes(el.tagName)
}

// Middle-click on links — open in new tab
window.addEventListener('auxclick', (e) => {
  if (e.button !== 1) return
  const link = closestLink(e.target)
  if (link?.href) {
    e.preventDefault()
    ipcRenderer.sendToHost('open-in-new-tab', link.href)
  }
})

window.addEventListener('keydown', (e) => {
  if (e.ctrlKey && e.key === 'Tab') {
    e.preventDefault()
    ipcRenderer.sendToHost('shortcut', e.shiftKey ? 'prev-tab' : 'next-tab')
    return
  }

  if ((e.metaKey || e.ctrlKey) && e.key.toLowerCase() === 'r') {
    e.preventDefault()
    ipcRenderer.sendToHost('shortcut', 'reload')
    return
  }

  if (!e.altKey || e.ctrlKey || e.metaKey || e.shiftKey || isEditable(e.target)) return
  if (e.key === 'ArrowLeft' || e.key === 'Left') {
    e.preventDefault()
    ipcRenderer.sendToHost('shortcut', 'go-back')
  } else if (e.key === 'ArrowRight' || e.key === 'Right') {
    e.preventDefault()
    ipcRenderer.sendToHost('shortcut', 'go-forward')
  }
}, true)

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
