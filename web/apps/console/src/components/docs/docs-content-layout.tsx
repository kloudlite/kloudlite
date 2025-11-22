import { TableOfContents } from './table-of-contents'

interface TocItem {
  id: string
  title: string
  level?: number
}

interface DocsContentLayoutProps {
  children: React.ReactNode
  tocItems?: TocItem[]
}

export function DocsContentLayout({ children, tocItems = [] }: DocsContentLayoutProps) {
  return (
    <div className="relative flex gap-8 max-w-[1400px] mx-auto px-4 sm:px-6 lg:px-8">
      {/* Main content */}
      <div className="prose prose-slate dark:prose-invert flex-1 min-w-0 py-8 px-0 xl:px-8">
        {children}
      </div>

      {/* Table of Contents */}
      {tocItems.length > 0 && <TableOfContents items={tocItems} />}
    </div>
  )
}
