import { cn } from '@kloudlite/lib'

const steps = [
  { label: 'Setup', color: 'gray' },
  { label: 'Code', color: 'blue' },
  { label: 'Build', color: 'gray' },
  { label: 'Deploy', color: 'gray' },
  { label: 'Test', color: 'green' },
]

export function WorkflowVisualization() {
  return (
    <div className="mt-16">
      <div className="flex flex-wrap items-center justify-center gap-3 sm:gap-4">
        {steps.map((step, i) => (
          <div key={step.label} className="flex items-center gap-3 sm:gap-4">
            <div
              className={cn(
                'px-6 py-3 text-sm font-semibold uppercase tracking-wide transition-all',
                step.color === 'blue' && 'border border-primary/20 bg-primary/5 text-primary',
                step.color === 'green' && 'border border-success/20 bg-success/5 text-success',
                step.color === 'gray' &&
                  'border border-foreground/10 bg-foreground/[0.02] text-muted-foreground line-through opacity-50',
              )}
            >
              {step.label}
            </div>
            {i < steps.length - 1 && <div className="h-px w-6 bg-foreground/20 sm:w-8" />}
          </div>
        ))}
      </div>

      <p className="text-muted-foreground mt-12 text-center text-lg font-medium">
        Designed to reduce development loop
      </p>
    </div>
  )
}
