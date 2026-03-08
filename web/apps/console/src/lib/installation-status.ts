import type { Installation } from '@/lib/console/storage'

export interface InstallationStatus {
  /** Short label like 'ACTIVE', 'INSTALLING', etc. */
  status: string
  /** Tailwind classes for the status badge */
  statusColor: string
  /** Human-readable description of the current state */
  description: string
  /** Whether the installation is still pending setup */
  isPending: boolean
  /** Whether there is an active deploy job running */
  isActiveJob: boolean
  /** Step progress string, e.g. "Step 2/5" */
  stepInfo: string | undefined
}

/** Check if an installation has an active job (running or pending) */
export function hasActiveJob(installation: Installation): boolean {
  return (
    (installation.deployJobStatus === 'running' || installation.deployJobStatus === 'pending') &&
    (installation.deployJobOperation === 'install' || installation.deployJobOperation === 'uninstall')
  )
}

/** Derive the full installation status from an Installation record */
export function getInstallationStatus(installation: Installation): InstallationStatus {
  const buildStepInfo = (): string | undefined =>
    installation.deployJobCurrentStep && installation.deployJobTotalSteps
      ? `Step ${installation.deployJobCurrentStep}/${installation.deployJobTotalSteps}`
      : undefined

  // Uninstall operations
  if (installation.deployJobOperation === 'uninstall') {
    if (installation.deployJobStatus === 'failed') {
      return {
        status: 'UNINSTALL FAILED',
        statusColor: 'bg-red-500/10 text-red-700 dark:text-red-400 border border-red-500/20',
        description: 'Uninstall job failed. You can retry from the danger zone below.',
        isPending: false,
        isActiveJob: false,
        stepInfo: undefined,
      }
    }
    return {
      status: 'UNINSTALLING',
      statusColor: 'bg-red-500/10 text-red-700 dark:text-red-400 border border-red-500/20',
      description: `Infrastructure is being torn down${buildStepInfo() ? ` (${buildStepInfo()})` : ''}`,
      isPending: false,
      isActiveJob: true,
      stepInfo: buildStepInfo(),
    }
  }

  // Active install jobs
  if (hasActiveJob(installation) && installation.deployJobOperation === 'install') {
    return {
      status: 'INSTALLING',
      statusColor: 'bg-blue-500/10 text-blue-700 dark:text-blue-400 border border-blue-500/20',
      description: `Installation is in progress${buildStepInfo() ? ` (${buildStepInfo()})` : ''}`,
      isPending: false,
      isActiveJob: true,
      stepInfo: buildStepInfo(),
    }
  }

  // Failed jobs (installation not ready and job failed)
  if (installation.deployJobStatus === 'failed' && !installation.deploymentReady) {
    return {
      status: 'ERROR',
      statusColor: 'bg-red-500/10 text-red-700 dark:text-red-400 border border-red-500/20',
      description: 'Installation job failed',
      isPending: true,
      isActiveJob: false,
      stepInfo: undefined,
    }
  }

  if (!installation.secretKey) {
    return {
      status: 'NOT INSTALLED',
      statusColor: 'bg-foreground/[0.06] text-foreground border border-foreground/10',
      description: 'Installation has not been deployed yet',
      isPending: true,
      isActiveJob: false,
      stepInfo: undefined,
    }
  }

  if (!installation.subdomain) {
    return {
      status: 'PENDING',
      statusColor: 'bg-yellow-500/10 text-yellow-700 dark:text-yellow-400 border border-yellow-500/20',
      description: 'Domain configuration is pending',
      isPending: true,
      isActiveJob: false,
      stepInfo: undefined,
    }
  }

  if (!installation.deploymentReady) {
    return {
      status: 'CONFIGURING',
      statusColor: 'bg-blue-500/10 text-blue-700 dark:text-blue-400 border border-blue-500/20',
      description: 'Installation is being configured',
      isPending: true,
      isActiveJob: false,
      stepInfo: undefined,
    }
  }

  return {
    status: 'ACTIVE',
    statusColor: 'bg-green-500/10 text-green-700 dark:text-green-400 border border-green-500/20',
    description: 'Installation is active and running',
    isPending: false,
    isActiveJob: false,
    stepInfo: undefined,
  }
}
