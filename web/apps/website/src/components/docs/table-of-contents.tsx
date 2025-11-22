'use client'

import { useState, useEffect } from 'react'

interface TocItem {
  id: string
  title: string
  level?: number
}

interface TableOfContentsProps {
  items: TocItem[]
}

export function TableOfContents({ items }: TableOfContentsProps) {
  const [activeSection, setActiveSection] = useState(items[0]?.id || '')

  useEffect(() => {
    const observer = new IntersectionObserver(
      (entries) => {
        entries.forEach((entry) => {
          if (entry.isIntersecting) {
            setActiveSection(entry.target.id)
          }
        })
      },
      { rootMargin: '-100px 0px -80% 0px' }
    )

    const sections = document.querySelectorAll('section[id], div[id]')
    sections.forEach((section) => observer.observe(section))

    return () => observer.disconnect()
  }, [])

  if (!items.length) return null

  return (
    <aside className="hidden xl:block w-64 flex-shrink-0">
      <div className="sticky top-24">
        <h3 className="text-foreground text-sm font-semibold mb-4">On this page</h3>
        <nav className="space-y-2">
          {items.map((item) => (
            <a
              key={item.id}
              href={`#${item.id}`}
              className={`block text-sm py-1 border-l-2 transition-colors ${
                item.level === 2 ? 'pl-4 text-xs' : 'pl-3'
              } ${
                activeSection === item.id
                  ? 'text-primary border-primary font-medium'
                  : 'text-muted-foreground border-transparent hover:text-foreground hover:border-primary'
              }`}
            >
              {item.title}
            </a>
          ))}
        </nav>
      </div>
    </aside>
  )
}
