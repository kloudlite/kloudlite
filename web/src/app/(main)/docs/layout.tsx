import { DocsSidebar } from './_components/docs-sidebar'
import { getTheme } from '@/lib/theme-server'
import { WebsiteHeader } from '@/components/website-header'

export default async function DocsLayout({ children }: { children: React.ReactNode }) {
  const theme = await getTheme()
  return (
    <div className="bg-background flex min-h-screen flex-col">
      <WebsiteHeader currentPage="docs" showSearch={true} />

      {/* Main Content Area with Sidebar */}
      <div className="mx-auto w-full max-w-[90rem]">
        <div className="flex">
          <DocsSidebar initialTheme={theme} />

          {/* Main Content */}
          <div className="flex-1 min-w-0">
            {children}
          </div>
        </div>
      </div>
    </div>
  )
}
