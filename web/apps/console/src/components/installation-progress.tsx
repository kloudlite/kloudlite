import { Check } from 'lucide-react'
import { cn } from '@/lib/utils'

interface Step {
  number: number
  title: string
  description: string
}

const steps: Step[] = [
  {
    number: 1,
    title: 'Create',
    description: 'Name your installation',
  },
  {
    number: 2,
    title: 'Install',
    description: 'Run installation in your cloud',
  },
  {
    number: 3,
    title: 'Configure Domain',
    description: 'Set your installation domain',
  },
  {
    number: 4,
    title: 'Complete',
    description: 'Access your installation',
  },
]

interface InstallationProgressProps {
  currentStep: number
}

export function InstallationProgress({ currentStep }: InstallationProgressProps) {
  return (
    <div className="mb-8">
      <div className="flex items-start justify-center gap-4">
        {steps.map((step, index) => (
          <div key={step.number} className="flex items-center gap-4">
            {/* Step with Label */}
            <div className="flex flex-col items-center">
              <div
                className={cn(
                  'flex size-10 items-center justify-center rounded-full border-2 font-medium transition-colors',
                  currentStep > step.number
                    ? 'border-primary bg-primary text-primary-foreground'
                    : currentStep === step.number
                      ? 'border-primary bg-background text-primary'
                      : 'border-muted-foreground/30 bg-background text-muted-foreground',
                )}
              >
                {currentStep > step.number ? (
                  <Check className="size-5" />
                ) : (
                  <span className="text-sm">{step.number}</span>
                )}
              </div>
              <div className="mt-3 text-center">
                <p
                  className={cn(
                    'text-sm font-semibold',
                    currentStep >= step.number ? 'text-foreground' : 'text-muted-foreground',
                  )}
                >
                  {step.title}
                </p>
              </div>
            </div>

            {/* Connector Line */}
            {index < steps.length - 1 && (
              <div
                className={cn(
                  'mb-auto h-0.5 w-24 translate-y-5 transition-colors',
                  currentStep > step.number ? 'bg-primary' : 'bg-muted-foreground/30',
                )}
              />
            )}
          </div>
        ))}
      </div>
    </div>
  )
}
