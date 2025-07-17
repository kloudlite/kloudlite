'use client'

import { useState } from 'react'
import { Menu, X } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { ThemeToggleClient } from '@/components/theme-toggle-client'
import { DocsSidebar } from '@/components/docs/docs-sidebar'
import { cn } from '@/lib/utils'

interface DocsLayoutProps {
  children: React.ReactNode
}

export function DocsLayout({ children }: DocsLayoutProps) {
  const [sidebarOpen, setSidebarOpen] = useState(false)

  return (
    <div className="h-screen bg-background flex overflow-hidden">
      {/* Mobile overlay */}
      {sidebarOpen && (
        <div 
          className="fixed inset-0 bg-black/50 z-40 lg:hidden"
          onClick={() => setSidebarOpen(false)}
        />
      )}

      {/* Sidebar */}
      <DocsSidebar isOpen={sidebarOpen} onClose={() => setSidebarOpen(false)} />
      
      {/* Mobile close button - positioned outside sidebar */}
      {sidebarOpen && (
        <Button
          variant="ghost"
          size="sm"
          onClick={() => setSidebarOpen(false)}
          className="fixed top-4 left-[288px] z-50 lg:hidden bg-background border border-border"
        >
          <X className="h-5 w-5" />
        </Button>
      )}

      {/* Main Content Area - Full Height */}
      <div className="flex-1 flex flex-col h-screen overflow-hidden">
        {/* Mobile Header */}
        <header className="lg:hidden sticky top-0 z-30 flex items-center h-16 px-4 bg-background border-b border-border flex-shrink-0">
          <Button
            variant="ghost"
            size="sm"
            onClick={() => setSidebarOpen(true)}
            className="rounded-none"
          >
            <Menu className="h-5 w-5" />
          </Button>
          <span className="ml-4 font-semibold">Kloudlite Docs</span>
          <div className="ml-auto">
            <ThemeToggleClient />
          </div>
        </header>

        {/* Scrollable Content */}
        <main className="flex-1 overflow-y-auto">
          {children}
        </main>
      </div>
    </div>
  )
}