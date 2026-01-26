'use client'

import { useState, useEffect } from 'react'
import { cn } from '@/lib/utils'

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
    <div className="hidden xl:block w-56 flex-shrink-0">
      <div className="sticky top-0 py-8">
        <div className="border-l border-foreground/10 pl-6">
          <h3 className="text-[11px] font-bold uppercase tracking-widest text-muted-foreground/80 mb-6">On this page</h3>
          <nav className="space-y-2">
            {items.map((item) => (
              <a
                key={item.id}
                href={`#${item.id}`}
                className={cn(
                  'group relative block text-[13px] font-medium transition-colors leading-relaxed py-1.5',
                  item.level === 2 && 'pl-4',
                  activeSection === item.id
                    ? 'text-primary'
                    : 'text-muted-foreground hover:text-foreground'
                )}
              >
                {/* Active indicator bar */}
                {activeSection === item.id && (
                  <span className="absolute left-0 top-0 bottom-0 w-[2px] bg-primary -ml-6" />
                )}

                {/* Hover indicator bar (only for non-active items) */}
                {activeSection !== item.id && (
                  <span className="absolute left-0 top-0 bottom-0 w-[2px] bg-foreground/20 scale-y-0 group-hover:scale-y-100 transition-transform duration-300 origin-center -ml-6" />
                )}

                <span className="relative z-10">{item.title}</span>
              </a>
            ))}
          </nav>
        </div>
      </div>
    </div>
  )
}
