export function ThemeScript({ theme }: { theme: string }) {
  const script = `
    (function() {
      document.documentElement.className = '${theme}';
    })();
  `
  
  return (
    <script 
      dangerouslySetInnerHTML={{ __html: script }} 
      suppressHydrationWarning
    />
  )
}