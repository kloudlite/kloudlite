interface HighlightedWordProps {
  children: string
  variant?: 'hover' | 'static'
}

export function HighlightedWord({ children, variant = 'hover' }: HighlightedWordProps) {
  if (variant === 'static') {
    return (
      <span className="relative inline-block">
        <span className="relative z-10">{children}</span>
        <span className="absolute bottom-0 left-0 right-0 h-1 bg-primary" />
      </span>
    )
  }

  return (
    <span className="relative inline-block">
      <span className="relative z-10">{children}</span>
      <span className="absolute bottom-0 left-0 right-0 h-0.5 bg-primary scale-x-0 group-hover:scale-x-100 transition-transform duration-300 origin-left" />
    </span>
  )
}
