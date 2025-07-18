import { DashboardSidebar } from './dashboard-sidebar'
import { Menu, X } from 'lucide-react'

interface DashboardLayoutProps {
  children: React.ReactNode
  teamSlug: string
  teamName: string
}

export async function DashboardLayout({ children, teamSlug, teamName }: DashboardLayoutProps) {
  return (
    <div className="h-screen bg-background">
      {/* Hidden checkbox for CSS-only toggle */}
      <input type="checkbox" id="sidebar-toggle" className="peer hidden" />
      
      {/* Overlay for mobile - visible when checkbox is checked */}
      <label 
        htmlFor="sidebar-toggle" 
        className="fixed inset-0 z-40 bg-background/80 lg:hidden cursor-pointer transition-all duration-300 ease-in-out invisible opacity-0 peer-checked:visible peer-checked:opacity-100"
        aria-label="Close sidebar"
      />
      
      {/* Sidebar - CSS transforms based on checkbox state */}
      <div className="fixed inset-y-0 left-0 z-50 w-80 transform -translate-x-full transition-transform duration-300 ease-in-out peer-checked:translate-x-0 lg:translate-x-0 border-r border-border bg-dashboard-sidebar shadow-dashboard-sidebar">
        <DashboardSidebar 
          teamSlug={teamSlug} 
          teamName={teamName}
        />
      </div>
      
      {/* Mobile close button - positioned outside sidebar */}
      <label
        htmlFor="sidebar-toggle"
        className="fixed top-4 left-[336px] z-50 lg:hidden cursor-pointer p-2 bg-background border border-border rounded-md shadow-sm transition-all duration-300 ease-in-out -translate-x-full opacity-0 peer-checked:translate-x-0 peer-checked:opacity-100"
        aria-label="Close sidebar"
      >
        <X className="h-5 w-5" />
      </label>
      
      {/* Main Content Area - add padding for fixed sidebar on desktop */}
      <div className="flex h-full flex-col lg:pl-80">
        {/* Header with mobile menu button */}
        <header className="h-16 border-b bg-dashboard-header/95 backdrop-blur-xl flex items-center px-4 lg:hidden shadow-dashboard-card-shadow">
          <label
            htmlFor="sidebar-toggle"
            className="p-2 hover:bg-muted rounded-md transition-colors cursor-pointer"
            aria-label="Open sidebar"
          >
            <Menu className="h-5 w-5" />
          </label>
          <span className="ml-3 font-semibold">{teamName}</span>
        </header>
        
        {/* Content */}
        <main className="flex-1 overflow-y-auto bg-dashboard-bg flex items-center justify-center">
          <div className="text-center">
            <div className="text-6xl font-light text-muted-foreground/50 mb-4">🚧</div>
            <h2 className="text-xl font-medium text-muted-foreground mb-2">Content Area</h2>
            <p className="text-sm text-muted-foreground">Page content will appear here</p>
          </div>
        </main>
      </div>
    </div>
  )
}