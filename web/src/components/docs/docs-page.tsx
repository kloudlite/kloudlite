'use client'

import { TableOfContents } from '@/components/docs/table-of-contents'
import { Breadcrumb } from '@/components/docs/breadcrumb'
import { Button } from '@/components/ui/button'
import { Link } from '@/components/ui/link'
import { ThemeToggleClient } from '@/components/theme-toggle-client'
import { Edit, Calendar, Clock, Star, BookOpen, Github, MessageSquare, Home, ExternalLink } from 'lucide-react'
import { cn } from '@/lib/utils'

interface DocsPageProps {
  children: React.ReactNode
  title?: string
  description?: string
  lastUpdated?: string
  editUrl?: string
  className?: string
  tags?: string[]
  difficulty?: 'beginner' | 'intermediate' | 'advanced'
  estimatedTime?: number
}

function DifficultyBadge({ level }: { level: 'beginner' | 'intermediate' | 'advanced' }) {
  const colors = {
    beginner: 'bg-green-100 text-green-800 border-green-200',
    intermediate: 'bg-yellow-100 text-yellow-800 border-yellow-200',
    advanced: 'bg-red-100 text-red-800 border-red-200'
  }
  
  return (
    <span className={cn(
      'inline-flex items-center gap-1 px-2 py-1 rounded-none text-xs font-medium border',
      colors[level]
    )}>
      <Star className="h-3 w-3" />
      {level.charAt(0).toUpperCase() + level.slice(1)}
    </span>
  )
}

export function DocsPage({ 
  children, 
  title, 
  description, 
  lastUpdated, 
  editUrl,
  className,
  tags = [],
  difficulty,
  estimatedTime
}: DocsPageProps) {
  return (
    <div className={cn('relative', className)}>
      <div className="max-w-7xl mx-auto px-6 py-8">
        <div className="flex gap-8">
          {/* Main content */}
          <div className="flex-1 min-w-0">
            {/* Breadcrumb */}
            <Breadcrumb className="mb-6" />
            
            {/* Page header */}
            {(title || description) && (
              <div className="mb-8">
                {title && (
                  <h1 className="text-3xl font-semibold mb-4">
                    {title}
                  </h1>
                )}
                {description && (
                  <p className="text-muted-foreground mb-6 max-w-2xl">
                    {description}
                  </p>
                )}
                
                {/* Meta information */}
                {(difficulty || estimatedTime || tags.length > 0) && (
                  <div className="flex flex-wrap items-center gap-4 mb-6">
                    {difficulty && <DifficultyBadge level={difficulty} />}
                    {estimatedTime && (
                      <span className="inline-flex items-center gap-1 px-2 py-1 rounded-none text-xs font-medium bg-muted text-muted-foreground">
                        <Clock className="h-3 w-3" />
                        {estimatedTime} min read
                      </span>
                    )}
                    {tags.map((tag) => (
                      <span 
                        key={tag} 
                        className="inline-flex items-center px-2 py-1 rounded-none text-xs font-medium bg-primary/10 text-primary"
                      >
                        #{tag}
                      </span>
                    ))}
                  </div>
                )}
              </div>
            )}
            
            {/* Content */}
            <div className="bg-card border border-border rounded-none">
              <div className="p-8 lg:p-12">
                <div className="docs-content">
                  {children}
                </div>
              </div>
            </div>
            
            {/* Article Footer */}
            <div className="mt-8 pt-6 border-t border-border">
              <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4">
                <div className="flex flex-wrap items-center gap-4 text-sm text-muted-foreground">
                  {lastUpdated && (
                    <div className="flex items-center gap-1.5">
                      <Calendar className="h-3 w-3" />
                      <span>Updated {lastUpdated}</span>
                    </div>
                  )}
                  <div className="flex items-center gap-1.5">
                    <BookOpen className="h-3 w-3" />
                    <span>Documentation</span>
                  </div>
                </div>
                
                <div className="flex items-center gap-2">
                  <Button asChild variant="outline" size="sm" className="rounded-none">
                    <Link href="#">
                      <MessageSquare className="h-3 w-3 mr-1.5" />
                      Feedback
                    </Link>
                  </Button>
                  {editUrl && (
                    <Button asChild variant="outline" size="sm" className="rounded-none">
                      <Link href={editUrl}>
                        <Edit className="h-3 w-3 mr-1.5" />
                        Edit page
                      </Link>
                    </Button>
                  )}
                </div>
              </div>
            </div>
          </div>
          
          {/* Table of Contents Sidebar */}
          <aside className="hidden xl:block w-64 shrink-0">
            <div className="sticky top-24">
              <div className="bg-card border border-border rounded-none">
                <div className="p-4 pl-6">
                  <TableOfContents />
                </div>
                
                {/* Footer content under TOC */}
                <div className="border-t border-border px-4 py-3 bg-muted/30">
                  <div className="flex items-center justify-between mb-3">
                    <span className="text-xs text-muted-foreground font-medium">v1.0.0</span>
                    <div className="w-1.5 h-1.5 bg-green-500 rounded-full" />
                  </div>
                  <div className="flex items-center justify-between">
                    <div className="flex items-center gap-4">
                      <Link href="/" className="text-xs text-muted-foreground hover:text-foreground flex items-center gap-1.5 transition-colors">
                        <Home className="h-3 w-3" />
                        <span>Home</span>
                      </Link>
                      <Link href="https://github.com/kloudlite/kloudlite" className="text-xs text-muted-foreground hover:text-foreground flex items-center gap-1.5 transition-colors">
                        <ExternalLink className="h-3 w-3" />
                        <span>GitHub</span>
                      </Link>
                    </div>
                    <ThemeToggleClient />
                  </div>
                </div>
              </div>
            </div>
          </aside>
        </div>
      </div>
    </div>
  )
}