import { ReactNode } from 'react'
import { cn } from '@/lib/utils'
import { LAYOUT } from '@/lib/constants/layout'

interface SectionProps {
  children: ReactNode
  className?: string
}

interface SectionHeaderProps {
  title: string
  icon?: ReactNode
  actions?: ReactNode
  className?: string
}

interface SectionContentProps {
  children: ReactNode
  className?: string
  spacing?: 'tight' | 'compact' | 'normal' | 'loose'
}

// Main section container
export function Section({ children, className }: SectionProps) {
  return (
    <div className={cn(LAYOUT.BACKGROUND.SECTION, className)}>
      {children}
    </div>
  )
}

// Section header with title and optional icon/actions
export function SectionHeader({ title, icon, actions, className }: SectionHeaderProps) {
  return (
    <div className={cn("border-b", LAYOUT.PADDING.SECTION, className)}>
      <div className="flex items-center justify-between">
        <h2 className={cn("font-semibold flex items-center", LAYOUT.GAP.SM)}>
          {icon}
          {title}
        </h2>
        {actions && (
          <div className={cn("flex items-center", LAYOUT.GAP.SM)}>
            {actions}
          </div>
        )}
      </div>
    </div>
  )
}

// Section content with configurable spacing
export function SectionContent({ children, className, spacing = 'normal' }: SectionContentProps) {
  const spacingClasses = {
    tight: LAYOUT.SPACING.TIGHT,
    compact: LAYOUT.SPACING.COMPACT,
    normal: LAYOUT.SPACING.ITEMS,
    loose: LAYOUT.SPACING.LOOSE,
  }

  return (
    <div className={cn(LAYOUT.PADDING.SECTION, spacingClasses[spacing], className)}>
      {children}
    </div>
  )
}

// Complete section with header and content
interface CompleteSectionProps {
  title: string
  icon?: ReactNode
  actions?: ReactNode
  children: ReactNode
  headerClassName?: string
  contentClassName?: string
  spacing?: 'tight' | 'compact' | 'normal' | 'loose'
}

export function CompleteSection({ 
  title, 
  icon, 
  actions, 
  children, 
  headerClassName, 
  contentClassName, 
  spacing = 'normal' 
}: CompleteSectionProps) {
  return (
    <Section>
      <SectionHeader title={title} icon={icon} actions={actions} className={headerClassName} />
      <SectionContent spacing={spacing} className={contentClassName}>
        {children}
      </SectionContent>
    </Section>
  )
}