import { ReactNode } from 'react'
import { cn } from '@/lib/utils'
import { LAYOUT } from '@/lib/constants/layout'

interface OverviewCardProps {
  icon: ReactNode
  title: string
  value: string | number
  className?: string
}

export function OverviewCard({ icon, title, value, className }: OverviewCardProps) {
  return (
    <div className={cn(LAYOUT.BACKGROUND.SECTION, className)}>
      <div className={LAYOUT.PADDING.CARD_LG}>
        <div className={cn("flex items-center", LAYOUT.GAP.MD)}>
          {icon}
          <div className="min-w-0">
            <p className="text-sm text-muted-foreground">{title}</p>
            <p className="text-2xl font-bold">{value}</p>
          </div>
        </div>
      </div>
    </div>
  )
}

interface OverviewGridProps {
  children: ReactNode
  columns?: 2 | 3 | 4
  className?: string
}

export function OverviewGrid({ children, columns = 3, className }: OverviewGridProps) {
  const gridClasses = {
    2: LAYOUT.GRID.RESPONSIVE_COLS_2,
    3: LAYOUT.GRID.RESPONSIVE_COLS_3,
    4: LAYOUT.GRID.RESPONSIVE_COLS_4,
  }

  return (
    <div className={cn(gridClasses[columns], "min-w-0", className)}>
      {children}
    </div>
  )
}