import { ReactNode } from 'react'

interface PageTitleProps {
  children: ReactNode
}

export function PageTitle({ children }: PageTitleProps) {
  return (
    <div className="relative mb-8">
      <h1 className="text-foreground font-bold tracking-tight text-3xl sm:text-[2.5rem] lg:text-5xl leading-tight sm:leading-[1.1] mt-0 relative inline-block pb-2">
        {children}
        {/* Permanent accent underline */}
        <span className="absolute -bottom-0 left-0 right-0 h-1 bg-primary" />
      </h1>
    </div>
  )
}
