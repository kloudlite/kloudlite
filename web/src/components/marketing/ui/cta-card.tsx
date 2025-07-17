import { cn } from '@/lib/utils'
import { Button } from '@/components/ui/button'
import { Link } from '@/components/ui/link'
import { ArrowRight } from 'lucide-react'

interface CTACardProps {
  title: string
  subtitle?: string
  description?: string
  buttonText?: string
  buttonHref: string
  className?: string
}

export function CTACard({ 
  title,
  subtitle,
  description,
  buttonText = "Get Started",
  buttonHref,
  className 
}: CTACardProps) {
  return (
    <div className={cn(
      "group relative p-6 border border-primary bg-primary/5",
      "flex flex-col justify-center transition-all duration-300",
      "hover:bg-primary/10 hover:shadow-lg hover:-translate-y-1",
      className
    )}>
      <h3 className="font-semibold mb-2 text-lg">{title}</h3>
      {subtitle && (
        <p className="text-sm text-muted-foreground mb-1">{subtitle}</p>
      )}
      {description && (
        <p className="text-sm text-muted-foreground mb-4">{description}</p>
      )}
      <Button 
        size="sm" 
        className="group-hover:scale-105 transition-transform duration-300" 
        asChild
      >
        <Link href={buttonHref}>
          {buttonText}
          <ArrowRight className="ml-2 h-4 w-4 group-hover:translate-x-1 transition-transform" />
        </Link>
      </Button>
    </div>
  )
}