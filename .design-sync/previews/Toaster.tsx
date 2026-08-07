import * as React from 'react'
import { Button, Toaster } from '@kloudlite/ui'
import { toast } from 'sonner'

// transform on the wrapper makes the fixed-position Toaster resolve against
// this box instead of the viewport, so the toast paints inside the cell.
const Stage = ({ children }: { children: React.ReactNode }) => (
  <div
    className="relative w-96 h-64 border p-4"
    style={{ transform: 'translateZ(0)', overflow: 'hidden' }}
  >
    {children}
  </div>
)

const useToast = (fire: () => void) => {
  React.useEffect(() => {
    const id = setTimeout(fire, 0)
    return () => clearTimeout(id)
  }, [])
}

export const Basic = () => {
  useToast(() =>
    toast('Deployment updated', {
      description: 'api-gateway rolled out to 3/3 replicas.',
      duration: Infinity,
    })
  )
  return (
    <Stage>
      <Button variant="outline">Show toast</Button>
      <Toaster position="bottom-right" expand visibleToasts={1} />
    </Stage>
  )
}

export const Success = () => {
  useToast(() =>
    toast.success('Environment created', {
      description: 'staging is now reconciling on ap-south-1.',
      duration: Infinity,
    })
  )
  return (
    <Stage>
      <Button>Create environment</Button>
      <Toaster position="bottom-right" expand visibleToasts={1} />
    </Stage>
  )
}

export const ErrorWithAction = () => {
  useToast(() =>
    toast.error('Install failed', {
      description: 'Could not reach the cluster API on acme-eu.',
      duration: Infinity,
      action: { label: 'Retry', onClick: () => {} },
    })
  )
  return (
    <Stage>
      <Button variant="destructive">Retry install</Button>
      <Toaster position="bottom-right" expand visibleToasts={1} />
    </Stage>
  )
}
