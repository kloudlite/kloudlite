'use client'

import { usePathname } from 'next/navigation'
import { ChevronRight, Home } from 'lucide-react'
import { Link } from '@/components/ui/link'
import { cn } from '@/lib/utils'

interface BreadcrumbItem {
  title: string
  href: string
}

interface BreadcrumbProps {
  className?: string
}

export function Breadcrumb({ className }: BreadcrumbProps) {
  const pathname = usePathname()
  
  // Generate breadcrumb items from pathname
  const generateBreadcrumbs = (): BreadcrumbItem[] => {
    const segments = pathname.split('/').filter(Boolean)
    const breadcrumbs: BreadcrumbItem[] = []
    
    // Add home
    breadcrumbs.push({
      title: 'Home',
      href: '/'
    })
    
    // Add docs home if we're in docs
    if (segments[0] === 'docs') {
      breadcrumbs.push({
        title: 'Documentation',
        href: '/docs'
      })
      
      // Add subsequent segments
      let currentPath = '/docs'
      for (let i = 1; i < segments.length; i++) {
        currentPath += `/${segments[i]}`
        
        // Convert slug to title (e.g., "getting-started" -> "Getting Started")
        const title = segments[i]
          .split('-')
          .map(word => word.charAt(0).toUpperCase() + word.slice(1))
          .join(' ')
        
        breadcrumbs.push({
          title,
          href: currentPath
        })
      }
    }
    
    return breadcrumbs
  }
  
  const breadcrumbs = generateBreadcrumbs()
  
  if (breadcrumbs.length <= 1) {
    return null
  }
  
  return (
    <nav className={cn('flex items-center space-x-2 text-sm', className)}>
      {breadcrumbs.map((item, index) => (
        <div key={item.href} className="flex items-center">
          {index > 0 && (
            <ChevronRight className="h-3 w-3 text-muted-foreground mx-2" />
          )}
          
          {index === 0 && (
            <Home className="h-3 w-3 text-muted-foreground mr-1" />
          )}
          
          {index === breadcrumbs.length - 1 ? (
            <span className="text-foreground font-medium">{item.title}</span>
          ) : (
            <Link
              href={item.href}
              className="text-muted-foreground hover:text-foreground transition-colors"
            >
              {item.title}
            </Link>
          )}
        </div>
      ))}
    </nav>
  )
}