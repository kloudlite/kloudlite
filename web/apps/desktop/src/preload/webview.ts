import { ipcRenderer } from 'electron'

document.addEventListener('DOMContentLoaded', () => {
  const titleObserver = new MutationObserver(() => {
    ipcRenderer.sendToHost('page-title-updated', document.title)
  })

  const titleEl = document.querySelector('title')
  if (titleEl) {
    titleObserver.observe(titleEl, { childList: true, characterData: true, subtree: true })
  }

  ipcRenderer.sendToHost('page-title-updated', document.title)
})
