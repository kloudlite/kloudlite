import { NavigationWrapper } from '@/components/navigation-wrapper'
import { ScrollArea } from '@/components/ui/scroll-area'

export default function MainLayout({
  children,
}: {
  children: React.ReactNode
}) {
  return (
    <div className="h-screen flex flex-col bg-gray-50">
      <div className="flex-shrink-0">
        <NavigationWrapper />
      </div>
      <ScrollArea className="flex-1 overflow-auto">
        {children}
      </ScrollArea>
    </div>
  )
}