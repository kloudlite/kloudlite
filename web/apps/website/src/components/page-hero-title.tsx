interface PageHeroTitleProps {
  children: React.ReactNode
  accentedWord: string
}

export function PageHeroTitle({ children, accentedWord }: PageHeroTitleProps) {
  return (
    <h1 className="text-[2.5rem] font-semibold leading-[1.1] tracking-[-0.02em] sm:text-4xl md:text-5xl lg:text-5xl xl:text-6xl text-foreground">
      {children} <span className="relative inline-block">
        <span className="relative z-10">{accentedWord}</span>
        <span className="absolute bottom-0 left-0 right-0 h-1 bg-primary"></span>
      </span>
    </h1>
  )
}
