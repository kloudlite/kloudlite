'use client'

import { Button } from '@/components/ui/button'
import { SimpleThemeToggle } from '@/components/simple-theme-toggle'
import { Home, HelpCircle, Github, BookOpen } from 'lucide-react'

export function SidebarActions() {
  return (
    <div className="flex items-center justify-between">
      {/* Quick Links */}
      <div className="flex items-center gap-1">
        <Button 
          variant="ghost" 
          size="icon" 
          className="size-8 hover:bg-gray-200/50 dark:hover:bg-gray-700/50 transition-all duration-200"
          onClick={() => window.location.href = '/'}
        >
          <Home className="size-4" />
          <span className="sr-only">Home</span>
        </Button>
        <Button 
          variant="ghost" 
          size="icon" 
          className="size-8 hover:bg-gray-200/50 dark:hover:bg-gray-700/50 transition-all duration-200"
          onClick={() => window.location.href = '/docs'}
        >
          <BookOpen className="size-4" />
          <span className="sr-only">Documentation</span>
        </Button>
        <Button 
          variant="ghost" 
          size="icon" 
          className="size-8 hover:bg-gray-200/50 dark:hover:bg-gray-700/50 transition-all duration-200"
          onClick={() => window.location.href = 'mailto:support@kloudlite.io'}
        >
          <HelpCircle className="size-4" />
          <span className="sr-only">Support</span>
        </Button>
        <Button 
          variant="ghost" 
          size="icon" 
          className="size-8 hover:bg-gray-200/50 dark:hover:bg-gray-700/50 transition-all duration-200"
          onClick={() => window.open('https://github.com/kloudlite/kloudlite', '_blank')}
        >
          <Github className="size-4" />
          <span className="sr-only">GitHub</span>
        </Button>
      </div>
      
      {/* Theme Toggle */}
      <SimpleThemeToggle />
    </div>
  )
}