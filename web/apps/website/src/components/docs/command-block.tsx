import { ReactNode } from 'react'
import { LucideIcon } from 'lucide-react'

interface CommandBlockProps {
  icon: LucideIcon
  title: string
  alias?: string
  description: string
  children: ReactNode
}

export function CommandBlock({ icon: Icon, title, alias, description, children }: CommandBlockProps) {
  return (
    <div className="group relative border border-foreground/10 rounded-sm overflow-hidden bg-background">
      {/* Animated gradient accent line */}
      <div className="absolute top-0 left-0 right-0 h-[2px] bg-gradient-to-r from-transparent via-primary to-transparent scale-x-0 group-hover:scale-x-100 transition-transform duration-500 origin-center" />

      <div className="bg-foreground/[0.02] group-hover:bg-foreground/[0.03] border-b border-foreground/10 px-6 py-4 transition-colors duration-300 relative z-10">
        <div className="flex items-center gap-3 mb-2">
          <div className="w-8 h-8 rounded-sm bg-primary/10 border border-primary/20 flex items-center justify-center group-hover:scale-110 transition-transform duration-300">
            <Icon className="h-4 w-4 text-primary" />
          </div>
          <code className="text-foreground group-hover:text-primary font-mono text-base font-semibold transition-colors duration-300">{title}</code>
          {alias && (
            <span className="text-muted-foreground group-hover:text-foreground text-xs font-mono ml-auto transition-colors duration-300">
              {alias}
            </span>
          )}
        </div>
        <p className="text-muted-foreground group-hover:text-foreground text-[14px] leading-relaxed transition-colors duration-300">{description}</p>
      </div>
      <div className="p-6 group-hover:bg-foreground/[0.015] transition-colors duration-300 relative z-10">
        {children}
      </div>
    </div>
  )
}

interface CodeExampleProps {
  title?: string
  children: ReactNode
}

export function CodeExample({ title, children }: CodeExampleProps) {
  return (
    <div className="group">
      {title && (
        <p className="text-foreground text-sm font-medium mb-3">{title}</p>
      )}
      <div className="relative overflow-hidden bg-foreground/[0.04] hover:bg-foreground/[0.05] border border-foreground/10 hover:border-foreground/20 rounded-sm p-4 font-mono text-[13px] overflow-x-auto transition-[background-color,border-color] duration-300">
        {/* Top accent line */}
        <div className="absolute top-0 left-0 right-0 h-[2px] bg-gradient-to-r from-transparent via-primary to-transparent scale-x-0 group-hover:scale-x-100 transition-transform duration-500 origin-center" />

        <div className="space-y-2 text-foreground/90 group-hover:text-foreground transition-colors duration-300 relative z-10">
          {children}
        </div>
      </div>
    </div>
  )
}

export function CodeLine({ children }: { children: ReactNode }) {
  return <pre className="m-0">{children}</pre>
}
