import { cn } from '@/lib/utils'
import { Link } from '@/components/ui/link'
import { ArrowRight } from 'lucide-react'

interface FeatureCardProps {
  title: string
  description: string
  href: string
  linkText?: string
  className?: string
}

export function FeatureCard({ 
  title,
  description,
  href,
  linkText = "Read more",
  className 
}: FeatureCardProps) {
  return (
    <div className={cn(
      "border border-border rounded-lg p-6",
      "hover:border-primary/50 transition-colors",
      className
    )}>
      <h3 className="font-semibold mb-3">{title}</h3>
      <p className="text-sm text-muted-foreground mb-4">
        {description}
      </p>
      <Link 
        href={href} 
        className="text-primary hover:text-primary-hover text-sm font-medium inline-flex items-center gap-1"
      >
        {linkText}
        <ArrowRight className="h-3 w-3" />
      </Link>
    </div>
  )
}