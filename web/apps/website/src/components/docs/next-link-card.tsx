import Link from 'next/link'
import { ArrowRight } from 'lucide-react'

interface NextLinkCardProps {
  href: string
  title: string
  description: string
}

export function NextLinkCard({ href, title, description }: NextLinkCardProps) {
  return (
    <Link
      href={href}
      className="group relative block overflow-hidden bg-foreground/[0.015] hover:bg-foreground/[0.03] border border-foreground/10 hover:border-primary/50 transition-[background-color,border-color] duration-300 p-5 rounded-sm no-underline"
    >
      {/* Animated vertical accent bar */}
      <div className="absolute left-0 top-0 w-[3px] h-full bg-primary scale-y-0 group-hover:scale-y-100 transition-transform duration-300 origin-top" />

      <div className="flex items-center justify-between gap-4 relative z-10">
        <div className="flex-1">
          <p className="text-foreground font-semibold text-[15px] mb-1 group-hover:text-primary transition-colors duration-300">
            {title}
          </p>
          <p className="text-muted-foreground group-hover:text-foreground text-[13px] leading-relaxed m-0 transition-colors duration-300">
            {description}
          </p>
        </div>
        <div className="flex-shrink-0 w-8 h-8 rounded-sm bg-foreground/[0.04] border border-foreground/10 group-hover:bg-primary/10 group-hover:border-primary/20 flex items-center justify-center transition-[background-color,border-color,transform] duration-300 group-hover:scale-110 group-hover:translate-x-1">
          <ArrowRight className="h-4 w-4 text-muted-foreground group-hover:text-primary transition-colors duration-300" />
        </div>
      </div>
    </Link>
  )
}
