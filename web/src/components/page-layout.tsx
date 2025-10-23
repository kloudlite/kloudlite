import { ReactNode } from 'react'
import { NavigationWrapper } from '@/components/navigation-wrapper'
import { ScrollArea } from '@/components/ui/scroll-area'

interface PageLayoutProps {
  children: ReactNode
  stickyNav?: boolean
  customHeader?: ReactNode
  noScrollArea?: boolean
}

export function PageLayout({
  children,
  stickyNav = false,
  customHeader,
  noScrollArea = false,
}: PageLayoutProps) {
  const navContent = (
    <>
      <NavigationWrapper />
      {customHeader}
    </>
  )

  if (noScrollArea) {
    return (
      <div className="flex min-h-screen flex-col bg-gray-50">
        {stickyNav ? <div className="sticky top-0 z-20">{navContent}</div> : navContent}
        {children}
      </div>
    )
  }

  return (
    <div className="flex min-h-screen flex-col">
      {stickyNav ? <div className="sticky top-0 z-20">{navContent}</div> : navContent}
      <ScrollArea className="flex-1 bg-gray-50">{children}</ScrollArea>
    </div>
  )
}
