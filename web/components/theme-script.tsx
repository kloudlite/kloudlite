export function ThemeScript() {
  const script = `
    (function() {
      try {
        function getTheme() {
          const cookie = document.cookie
          const match = cookie.match(/theme=(light|dark|system)/)
          const theme = match ? match[1] : 'system'
          
          if (theme === 'system') {
            return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light'
          }
          return theme
        }
        
        const theme = getTheme()
        const root = document.documentElement
        
        // Only update if not already set correctly
        if (!root.classList.contains(theme)) {
          root.classList.remove('light', 'dark')
          root.classList.add(theme)
        }
      } catch (e) {}
    })()
  `

  return <script dangerouslySetInnerHTML={{ __html: script }} />
}