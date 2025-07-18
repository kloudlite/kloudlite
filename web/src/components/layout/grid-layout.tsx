import { cn } from '@/lib/utils'
import { LAYOUT } from '@/lib/constants/layout'

interface GridLayoutProps {
  children: React.ReactNode
  columns?: 2 | 3 | 4
  className?: string
}

export function GridLayout({ children, columns = 3, className }: GridLayoutProps) {
  const gridClass = {
    2: LAYOUT.GRID.COLS_2,
    3: LAYOUT.GRID.COLS_3,
    4: LAYOUT.GRID.COLS_4,
  }[columns]

  return (
    <div className={cn(gridClass, className)}>
      {children}
    </div>
  )
}

interface CardProps {
  children: React.ReactNode
  className?: string
  padding?: 'default' | 'large'
}

export function Card({ children, className, padding = 'default' }: CardProps) {
  const paddingClass = padding === 'large' ? LAYOUT.PADDING.CARD_LG : LAYOUT.PADDING.CARD
  
  return (
    <div className={cn(LAYOUT.BACKGROUND.CARD, paddingClass, className)}>
      {children}
    </div>
  )
}

interface CardHeaderProps {
  title: string
  description?: string
  actions?: React.ReactNode
  className?: string
}

export function CardHeader({ title, description, actions, className }: CardHeaderProps) {
  return (
    <div className={cn('flex items-center justify-between mb-4', className)}>
      <div>
        <h3 className="text-lg font-semibold">{title}</h3>
        {description && (
          <p className="text-sm text-muted-foreground mt-1">{description}</p>
        )}
      </div>
      {actions}
    </div>
  )
}

interface FlexLayoutProps {
  children: React.ReactNode
  direction?: 'row' | 'col'
  gap?: 'sm' | 'md' | 'lg'
  className?: string
}

export function FlexLayout({ children, direction = 'row', gap = 'md', className }: FlexLayoutProps) {
  const gapClass = {
    sm: LAYOUT.GAP.SM,
    md: LAYOUT.GAP.MD,
    lg: LAYOUT.GAP.LG,
  }[gap]

  const flexClass = direction === 'col' ? 'flex flex-col' : 'flex items-center'

  return (
    <div className={cn(flexClass, gapClass, className)}>
      {children}
    </div>
  )
}