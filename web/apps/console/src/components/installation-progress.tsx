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
    <div className="flex items-center gap-1">
      {steps.map((step, index) => (
        <div key={step.number} className="flex items-center gap-1">
          <div className="flex items-center gap-2.5">
            <div
              className={cn(
                'flex size-7 items-center justify-center text-xs font-semibold transition-all rounded-full',
                currentStep > step.number
                  ? 'bg-primary text-primary-foreground'
                  : currentStep === step.number
                    ? 'bg-primary text-primary-foreground'
                    : 'bg-muted text-muted-foreground',
              )}
            >
              {currentStep > step.number ? (
                <Check className="size-3.5 stroke-[2.5]" />
              ) : (
                <span>{step.number}</span>
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
            <div className="w-6 h-px bg-foreground/10 mx-2" />
          )}
        </div>
      ))}
    </div>
  )
}
