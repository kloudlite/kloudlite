import { cn } from '@/lib/utils'
import { Link } from '@/components/ui/link'
import { LucideIcon } from 'lucide-react'

interface ResourceItem {
  icon: string | LucideIcon
  label: string
  href: string
}

interface ResourceListProps {
  items: ResourceItem[]
  className?: string
}

export function ResourceList({ items, className }: ResourceListProps) {
  return (
    <div className={cn("space-y-3", className)}>
      {items.map((item, index) => (
        <div key={index} className="flex items-center gap-3">
          <span className="text-primary">
            {typeof item.icon === 'string' ? item.icon : <item.icon className="h-4 w-4" />}
          </span>
          <Link 
            href={item.href} 
            className="text-primary hover:text-primary-hover font-medium"
          >
            {item.label}
          </Link>
        </div>
      ))}
    </div>
  )
}