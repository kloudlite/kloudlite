import { cn } from '@/lib/utils'
import { LAYOUT } from '@/lib/constants/layout'

interface PageHeaderProps {
  title?: string
  description?: string
  actions?: React.ReactNode
  className?: string
}

export function PageHeader({ title, description, actions, className }: PageHeaderProps) {
  return (
    <div className={cn('bg-background border-b', className)}>
      <div className={cn(LAYOUT.CONTAINER, LAYOUT.PADDING.HEADER)}>
        <div className="flex items-center justify-between">
          {(title || description) && (
            <div>
              {title && <h1 className="text-2xl font-bold">{title}</h1>}
              {description && (
                <p className="text-sm text-muted-foreground mt-1">
                  {description}
                </p>
              )}
            </div>
          )}
          {actions && (
            <div className={cn("flex items-center gap-3", !title && !description && "w-full justify-end")}>
              {actions}
            </div>
          )}
        </div>
      </div>
    </div>
  )
}

interface PageContainerProps {
  children: React.ReactNode
  className?: string
}

export function PageContainer({ children, className }: PageContainerProps) {
  return (
    <div className={cn(LAYOUT.BACKGROUND.PAGE, className)}>
      {children}
    </div>
  )
}

interface PageContentProps {
  children: React.ReactNode
  className?: string
}

export function PageContent({ children, className }: PageContentProps) {
  return (
    <div className="flex-1">
      <div className={cn(LAYOUT.CONTAINER, LAYOUT.PADDING.PAGE, LAYOUT.SPACING.SECTION, className)}>
        {children}
      </div>
    </div>
  )
}

interface PageSectionProps {
  children: React.ReactNode
  title?: string
  description?: string
  className?: string
}

export function PageSection({ children, title, description, className }: PageSectionProps) {
  return (
    <div>
      {(title || description) && (
        <div className="mb-6">
          {title && <h2 className="text-xl font-semibold mb-2">{title}</h2>}
          {description && (
            <p className="text-sm text-muted-foreground">{description}</p>
          )}
        </div>
      )}
      <div className={cn(LAYOUT.BACKGROUND.SECTION, className)}>
        {children}
      </div>
    </div>
  )
}