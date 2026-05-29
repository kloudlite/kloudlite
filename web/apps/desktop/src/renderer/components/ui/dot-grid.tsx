import { cn } from '@/lib/utils'

interface DotGridProps {
  value: number
  total?: number
  size?: number
  className?: string
}

const tiers = [
  { max: 60, className: 'bg-emerald-400' },
  { max: 80, className: 'bg-amber-400' },
  { max: 100, className: 'bg-red-400' },
]

function getColor(value: number): string {
  for (const tier of tiers) {
    if (value <= tier.max) return tier.className
  }
  return tiers[tiers.length - 1].className
}

export function DotGrid({ value, total = 10, size = 10, className }: DotGridProps) {
  const filled = Math.round((value / 100) * total)
  const colorClass = getColor(value)

  return (
    <div className={cn('flex items-center gap-[3px]', className)}>
      {Array.from({ length: total }, (_, i) => (
        <div
          key={i}
          className={cn(
            'rounded-sm transition-all duration-300',
            i < filled
              ? colorClass
              : 'bg-sidebar-foreground/[0.08]'
          )}
          style={{ width: size, height: size }}
        />
      ))}
    </div>
  )
}
