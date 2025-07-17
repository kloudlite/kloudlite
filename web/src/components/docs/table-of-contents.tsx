'use client'

import { useState, useEffect } from 'react'
import { Link } from '@/components/ui/link'
import { cn } from '@/lib/utils'

interface TocItem {
  id: string
  title: string
  level: number
}

interface TableOfContentsProps {
  className?: string
}

export function TableOfContents({ className }: TableOfContentsProps) {
  const [toc, setToc] = useState<TocItem[]>([])
  const [activeId, setActiveId] = useState<string>('')

  useEffect(() => {
    // Generate table of contents from headings
    const headings = document.querySelectorAll('h2, h3, h4, h5, h6')
    const tocItems: TocItem[] = []

    headings.forEach((heading) => {
      const id = heading.id || heading.textContent?.toLowerCase().replace(/\s+/g, '-') || ''
      if (id && !heading.id) {
        heading.id = id
      }
      
      if (id) {
        tocItems.push({
          id,
          title: heading.textContent || '',
          level: parseInt(heading.tagName.charAt(1))
        })
      }
    })

    setToc(tocItems)
  }, [])

  useEffect(() => {
    // Track active heading
    const observer = new IntersectionObserver(
      (entries) => {
        entries.forEach((entry) => {
          if (entry.isIntersecting) {
            setActiveId(entry.target.id)
          }
        })
      },
      {
        rootMargin: '-80px 0px -80% 0px',
        threshold: 0.1
      }
    )

    const headings = document.querySelectorAll('h2, h3, h4, h5, h6')
    headings.forEach((heading) => {
      if (heading.id) {
        observer.observe(heading)
      }
    })

    return () => observer.disconnect()
  }, [toc])

  if (toc.length === 0) {
    return null
  }

  const groupedToc = toc.reduce((acc, item) => {
    if (item.level === 2) {
      acc.push({ ...item, children: [] })
    } else if (acc.length > 0) {
      acc[acc.length - 1].children.push(item)
    }
    return acc
  }, [] as Array<TocItem & { children: TocItem[] }>)

  return (
    <div className={cn('', className)}>
      <h5 className="font-semibold text-sm text-foreground mb-4">On this page</h5>
      <nav>
        <ul className="space-y-3">
          {groupedToc.map((section) => (
            <li key={section.id}>
              <Link
                href={`#${section.id}`}
                className={cn(
                  'relative block text-sm font-medium transition-colors py-1',
                  activeId === section.id 
                    ? 'text-primary' 
                    : 'text-foreground hover:text-primary'
                )}
              >
                {activeId === section.id && (
                  <span className="absolute -left-4 top-0 w-0.5 h-full bg-primary" />
                )}
                <span className="block truncate">
                  {section.title}
                </span>
              </Link>
              
              {section.children.length > 0 && (
                <ul className="mt-2 space-y-1">
                  {section.children.map((child) => (
                    <li key={child.id}>
                      <Link
                        href={`#${child.id}`}
                        className={cn(
                          'relative block text-xs transition-colors py-1 pl-4',
                          activeId === child.id
                            ? 'text-primary font-medium'
                            : 'text-muted-foreground hover:text-foreground'
                        )}
                      >
                        {activeId === child.id && (
                          <span className="absolute -left-4 top-0 w-0.5 h-full bg-primary" />
                        )}
                        <span className="block truncate">
                          {child.title}
                        </span>
                      </Link>
                    </li>
                  ))}
                </ul>
              )}
            </li>
          ))}
        </ul>
      </nav>
    </div>
  )
}