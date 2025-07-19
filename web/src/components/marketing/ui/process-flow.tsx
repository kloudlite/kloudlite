import React from 'react'
import { cn } from '@/lib/utils'

interface ProcessStep {
  label: string
  active: boolean
  strikethrough?: boolean
}

interface ProcessFlowProps {
  steps: ProcessStep[]
  className?: string
}

export function ProcessFlow({ steps, className }: ProcessFlowProps) {
  return (
    <div className={cn("flex items-center justify-center gap-2 sm:gap-3", className)}>
      <div className="flex items-center gap-2 sm:gap-3 flex-wrap justify-center">
        {steps.map((step, index) => (
          <React.Fragment key={index}>
            <div className={cn(
              "px-3 py-2 sm:px-4 font-mono text-sm sm:text-base lg:text-lg border rounded",
              step.active 
                ? "bg-success/10 border-success/20 text-success"
                : "bg-muted/50 border-border text-muted-foreground"
            )}>
              {step.strikethrough ? (
                <span className="line-through">{step.label}</span>
              ) : (
                step.label
              )}
            </div>
            {index < steps.length - 1 && (
              <span className="text-muted-foreground text-sm sm:text-base">→</span>
            )}
          </React.Fragment>
        ))}
      </div>
    </div>
  )
}