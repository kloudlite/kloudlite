import { Check } from 'lucide-react'
import { cn } from '@kloudlite/lib'

interface Step {
  number: number
  title: string
}

const steps: Step[] = [
  {
    number: 1,
    title: 'Details',
  },
  {
    number: 2,
    title: 'Deploy',
  },
  {
    number: 3,
    title: 'Complete',
  },
]

interface InstallationProgressProps {
  currentStep: number
}

export function InstallationProgress({ currentStep }: InstallationProgressProps) {
  return (
    <div className="flex items-center gap-2">
      {steps.map((step, index) => (
        <div key={step.number} className="flex items-center gap-2">
          <div className="flex items-center gap-2">
            <div
              className={cn(
                'flex size-6 items-center justify-center text-xs font-semibold transition-all duration-300 border rounded-full',
                currentStep > step.number
                  ? 'bg-primary text-primary-foreground border-primary'
                  : currentStep === step.number
                    ? 'bg-primary text-primary-foreground border-primary'
                    : 'bg-background border-border text-muted-foreground',
              )}
            >
              {currentStep > step.number ? (
                <Check className="size-3 stroke-[3]" />
              ) : (
                <span className="text-[11px]">{step.number}</span>
              )}
            </div>
            <span
              className={cn(
                'text-sm font-medium',
                currentStep >= step.number ? 'text-foreground' : 'text-muted-foreground',
              )}
            >
              {step.title}
            </span>
          </div>
          {index < steps.length - 1 && (
            <div className="w-8 h-px bg-border mx-1" />
          )}
        </div>
      ))}
    </div>
  )
}
