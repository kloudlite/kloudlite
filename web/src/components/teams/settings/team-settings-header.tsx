'use client'

import Link from 'next/link'
import { usePathname } from 'next/navigation'
import { cn } from '@/lib/utils'
import { LAYOUT } from '@/lib/constants/layout'
import { Settings, Users, Server, ChevronRight } from 'lucide-react'
import { 
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbLink,
  BreadcrumbList,
  BreadcrumbPage,
  BreadcrumbSeparator,
} from '@/components/ui/breadcrumb'

interface TeamSettingsHeaderProps {
  teamname: string
  teamDisplayName?: string
}

const settingsNavigation = [
  {
    name: 'General',
    href: '/settings/general',
    icon: Settings,
    description: 'Basic team information and settings'
  },
  {
    name: 'User Management',
    href: '/settings/users',
    icon: Users,
    description: 'Manage team members and permissions'
  },
  {
    name: 'Infrastructure',
    href: '/settings/infrastructure',
    icon: Server,
    description: 'Infrastructure resources and policies'
  },
]

export function TeamSettingsHeader({ 
  teamname, 
  teamDisplayName
}: TeamSettingsHeaderProps) {
  const pathname = usePathname()
  
  // Get current tab info
  const currentTab = settingsNavigation.find(tab => 
    pathname === `/${teamname}${tab.href}`
  )

  return (
    <>
      {/* Header Content - Non-sticky */}
      <header className="bg-background">
        <div className={cn(LAYOUT.CONTAINER, LAYOUT.PADDING.HEADER)}>
          {/* Breadcrumbs */}
          <Breadcrumb className="mb-2 sm:mb-3 md:mb-4">
            <BreadcrumbList>
              <BreadcrumbItem>
                <BreadcrumbLink href={`/${teamname}`}>
                  {teamDisplayName || 'Team'}
                </BreadcrumbLink>
              </BreadcrumbItem>
              <BreadcrumbSeparator>
                <ChevronRight className="h-4 w-4" />
              </BreadcrumbSeparator>
              <BreadcrumbItem>
                <BreadcrumbPage>Settings</BreadcrumbPage>
              </BreadcrumbItem>
            </BreadcrumbList>
          </Breadcrumb>

          {/* Header */}
          <div className="mb-3 sm:mb-4 md:mb-5 lg:mb-6">
            <h1 className="text-2xl font-bold">
              {currentTab ? currentTab.name : 'Settings'}
            </h1>
            <p className="text-muted-foreground mt-1 text-sm">
              {currentTab ? currentTab.description : 'Manage your team settings, members, and infrastructure'}
            </p>
          </div>
        </div>
      </header>

      {/* Navigation Tabs - Always sticky */}
      <div className="sticky top-0 z-10 bg-background border-b border-border">
        <div className={cn(LAYOUT.CONTAINER, LAYOUT.PADDING.HEADER_X)}>
          <nav>
            {/* Responsive tabs - scrollable on mobile and tablet */}
            <div className={cn("flex overflow-x-auto lg:overflow-x-visible [&::-webkit-scrollbar]:hidden [-ms-overflow-style:none] [scrollbar-width:none]", LAYOUT.GAP.RESPONSIVE_MD)}>
              {settingsNavigation.map((item) => {
                const href = `/${teamname}${item.href}`
                const isActive = pathname === href
                const Icon = item.icon

                return (
                  <Link
                    key={item.name}
                    href={href}
                    className={cn(
                      'flex items-center border-b-2 font-medium transition-colors relative group whitespace-nowrap',
                      LAYOUT.GAP.RESPONSIVE,
                      LAYOUT.PADDING.TAB,
                      isActive
                        ? 'border-primary text-primary'
                        : 'border-transparent text-muted-foreground hover:text-foreground hover:border-border'
                    )}
                  >
                    <Icon className="h-4 w-4 flex-shrink-0" />
                    <span className="hidden sm:inline">{item.name}</span>
                    <span className="sm:hidden text-sm">{item.name.split(' ')[0]}</span>
                    
                    {/* Tooltip on hover - hidden on mobile */}
                    <div className="hidden sm:block absolute top-full left-1/2 transform -translate-x-1/2 mt-2 px-3 py-2 bg-popover text-popover-foreground text-xs rounded-md border shadow-md opacity-0 group-hover:opacity-100 transition-opacity duration-200 pointer-events-none whitespace-nowrap z-50">
                      {item.description}
                      <div className="absolute -top-1 left-1/2 transform -translate-x-1/2 w-2 h-2 bg-popover border-l border-t rotate-45"></div>
                    </div>
                  </Link>
                )
              })}
            </div>
          </nav>
        </div>
      </div>
    </>
  )
}