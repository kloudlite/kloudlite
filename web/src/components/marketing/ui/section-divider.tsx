import { cn } from '@/lib/utils'

interface SectionDividerProps {
  className?: string
}

export function SectionDivider({ className }: SectionDividerProps) {
  return (
    <div className={cn("my-32 flex justify-center", className)}>
      <div className="w-full max-w-2xl h-px bg-gradient-to-r from-transparent via-border to-transparent" />
    </div>
  )
}