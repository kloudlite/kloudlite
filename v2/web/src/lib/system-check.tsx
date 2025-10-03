import { listMachineTypes } from '@/app/actions/machine-type.actions'
import { signOutAction } from '@/app/actions/auth'
import { Button } from '@/components/ui/button'

/**
 * Check if the system is ready (has machine types configured)
 */
export async function isSystemReady(): Promise<boolean> {
  const machineTypesResult = await listMachineTypes()
  return machineTypesResult.success &&
    machineTypesResult.data?.items.some(mt => mt.spec.active !== false) || false
}

/**
 * Render the system setup in progress page
 */
export function SystemSetupPage() {
  return (
    <div className="fixed inset-0 flex items-center justify-center bg-background">
      <div className="text-center space-y-6">
        <div className="space-y-4">
          <h1 className="text-3xl font-bold">System Setup in Progress</h1>
          <p className="text-muted-foreground">
            The system is being configured. Please check back soon.
          </p>
        </div>
        <form action={signOutAction}>
          <Button type="submit" variant="outline">
            Sign Out
          </Button>
        </form>
      </div>
    </div>
  )
}
