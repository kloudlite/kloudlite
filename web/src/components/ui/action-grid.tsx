import { Link } from '@/components/ui/link'
import { GridLayout, Card } from '@/components/layout/grid-layout'
import { cn } from '@/lib/utils'
import { LucideIcon } from 'lucide-react'

interface ActionItem {
  href: string
  label: string
  icon: LucideIcon
  description: string
  metric?: string
  status?: string
  disabled?: boolean
}

interface ActionGridProps {
  title?: string
  actions: ActionItem[]
  columns?: 2 | 3 | 4
  className?: string
}

export function ActionGrid({ title, actions, columns = 3, className }: ActionGridProps) {
  return (
    <div className={className}>
      {title && (
        <h3 className="text-lg font-semibold mb-4">{title}</h3>
      )}
      <GridLayout columns={columns}>
        {actions.map((action) => (
          <ActionCard key={action.href} action={action} />
        ))}
      </GridLayout>
    </div>
  )
}

interface ActionCardProps {
  action: ActionItem
  className?: string
}

export function ActionCard({ action, className }: ActionCardProps) {
  const { href, label, icon: Icon, description, metric, status, disabled } = action

  const content = (
    <div className="flex items-start gap-3">
      <div className="p-2 bg-muted rounded-lg group-hover:bg-primary/10 dark:group-hover:bg-primary/20 transition-all duration-200">
        <Icon className="h-5 w-5 text-muted-foreground group-hover:text-primary transition-colors duration-200" />
      </div>
      <div className="flex-1">
        <h4 className="font-medium group-hover:text-primary transition-colors duration-200">
          {label}
        </h4>
        <p className="text-sm text-muted-foreground mt-0.5">
          {description}
        </p>
        {(metric || status) && (
          <div className="mt-2 pt-2 border-t border-border/50">
            <div className="flex items-center justify-between text-xs">
              {metric && (
                <span className="font-medium text-foreground">{metric}</span>
              )}
              {status && (
                <span className="text-muted-foreground">{status}</span>
              )}
            </div>
          </div>
        )}
      </div>
    </div>
  )

  if (disabled) {
    return (
      <Card className={cn(
        'opacity-60 cursor-not-allowed',
        className
      )}>
        {content}
      </Card>
    )
  }

  return (
    <Link
      href={href}
      className={cn(
        'group block',
        className
      )}
    >
      <Card className="hover:border-primary/40 hover:shadow-dashboard-card-shadow transition-all duration-200 hover:scale-[1.02]">
        {content}
      </Card>
    </Link>
  )
}

interface MetricCardProps {
  title: string
  value: string | number
  subtitle?: string
  icon?: LucideIcon
  trend?: 'up' | 'down' | 'neutral'
  className?: string
}

export function MetricCard({ title, value, subtitle, icon: Icon, trend, className }: MetricCardProps) {
  const trendColor = {
    up: 'text-green-600',
    down: 'text-red-600', 
    neutral: 'text-muted-foreground'
  }[trend || 'neutral']

  return (
    <Card className={className} padding="large">
      <div className="flex items-center justify-between">
        <div>
          <p className="text-sm font-medium text-muted-foreground">{title}</p>
          <p className="text-2xl font-bold mt-1">{value}</p>
          {subtitle && (
            <p className={cn('text-sm mt-1', trendColor)}>{subtitle}</p>
          )}
        </div>
        {Icon && (
          <div className="p-2 bg-muted rounded-lg">
            <Icon className="h-5 w-5 text-muted-foreground" />
          </div>
        )}
      </div>
    </Card>
  )
}