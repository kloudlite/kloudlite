import type React from 'react'
import Link from 'next/link'
import { cn } from '@kloudlite/lib'

interface FeatureCardContainerProps {
  children: React.ReactNode
  className?: string
  href?: string
}

export function FeatureCardContainer({ children, className, href }: FeatureCardContainerProps) {
  const content = (
    <>
      <div className="absolute bottom-0 left-0 h-[1px] w-0 bg-primary transition-all duration-700 ease-out group-hover:w-full" />
      <div className="relative flex flex-col">{children}</div>
    </>
  )

  if (href) {
    return (
      <Link
        href={href}
        className={cn(
          'group relative overflow-hidden border-b border-foreground/10 bg-background p-8 transition-all duration-500 hover:bg-foreground/[0.02] lg:p-10',
          className,
        )}
      >
        {content}
      </Link>
    )
  }

  return (
    <div
      className={cn(
        'group relative overflow-hidden border-b border-foreground/10 bg-background p-8 transition-all duration-500 hover:bg-foreground/[0.02] lg:p-10',
        className,
      )}
    >
      {content}
    </div>
  )
}

interface FeatureCardProps {
  icon: React.ReactNode
  title: string
  description: string
}

export function FeatureCard({ icon, title, description }: FeatureCardProps) {
  return (
    <div className="flex h-full flex-col space-y-3">
      <div className="text-muted-foreground transition-colors duration-500 group-hover:text-primary">
        <div className="h-8 w-8 opacity-60 transition-all duration-500 group-hover:scale-110 group-hover:opacity-100">
          {icon}
        </div>
      </div>

      <div className="flex-1 space-y-4">
        <h3 className="text-foreground text-xl leading-tight font-bold tracking-tight transition-colors duration-500 group-hover:text-primary sm:text-2xl">
          {title}
        </h3>
        <p className="text-muted-foreground text-sm leading-relaxed sm:text-base">{description}</p>
      </div>
    </div>
  )
}
