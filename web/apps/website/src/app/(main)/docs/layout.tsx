import { DocsSidebar } from './_components/docs-sidebar'
import { WebsiteHeader } from '@/components/website-header'
import { ScrollArea } from '@kloudlite/ui'

export default function DocsLayout({ children }: { children: React.ReactNode }) {
  return (
    <div className="bg-background h-screen flex flex-col">
      {/* Fixed Header */}
      <div className="sticky top-0 z-50">
        <WebsiteHeader currentPage="docs" alwaysShowBorder />
      </div>

      {/* Scrollable Content Area */}
      <div className="flex-1 overflow-hidden">
        <ScrollArea className="h-full">
          <div className="mx-auto w-full max-w-[90rem] px-4 sm:px-6 lg:px-8">
            <div className="flex gap-6 lg:gap-8 xl:gap-12">
              <DocsSidebar />

              {/* Main Content */}
              <div className="flex-1 min-w-0 py-8 lg:py-12">
                {children}
              </div>
            </div>
          </div>
        </ScrollArea>
      </div>
    </div>
  )
}
