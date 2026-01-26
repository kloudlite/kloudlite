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
    <div className="relative flex gap-16 max-w-[1400px]">
      {/* Main content */}
      <div className="flex-1 min-w-0 py-8 pb-24 lg:pb-16">
        <article className="prose prose-neutral dark:prose-invert max-w-none
          prose-headings:font-bold prose-headings:tracking-tight
          prose-h1:text-[2.5rem] prose-h1:lg:text-5xl prose-h1:mb-8 prose-h1:leading-[1.1] prose-h1:mt-0
          prose-h2:text-2xl prose-h2:lg:text-3xl prose-h2:mt-16 prose-h2:mb-6 prose-h2:leading-tight prose-h2:border-t prose-h2:border-foreground/10 prose-h2:pt-8
          prose-h3:text-xl prose-h3:lg:text-2xl prose-h3:mt-12 prose-h3:mb-4
          prose-p:text-[15px] prose-p:lg:text-base prose-p:leading-relaxed prose-p:text-muted-foreground
          prose-a:text-primary prose-a:font-medium prose-a:no-underline hover:prose-a:underline
          prose-strong:text-foreground prose-strong:font-semibold
          prose-code:text-sm prose-code:font-mono prose-code:bg-foreground/[0.05] prose-code:px-1.5 prose-code:py-0.5 prose-code:rounded-sm prose-code:border prose-code:border-foreground/10
          prose-pre:bg-foreground/[0.03] prose-pre:border prose-pre:border-foreground/10 prose-pre:rounded-sm
          prose-img:rounded-sm prose-img:border prose-img:border-foreground/10
          prose-ul:text-muted-foreground prose-ul:leading-relaxed
          prose-ol:text-muted-foreground prose-ol:leading-relaxed
          prose-li:text-[15px] prose-li:lg:text-base prose-li:marker:text-muted-foreground">
          {children}
        </article>
      </div>

      {/* Table of Contents */}
      {tocItems.length > 0 && <TableOfContents items={tocItems} />}
    </div>
  )
}
