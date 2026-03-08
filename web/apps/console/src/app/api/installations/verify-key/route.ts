import { NextRequest, NextResponse } from 'next/server'
import { apiError } from '@/lib/api-helpers'
import {
  getInstallationByKey,
  markInstallationComplete,
  updateHealthCheck,
  updateInstallation,
} from '@/lib/console/storage'
import {
  getActiveUsagePeriodsForInstallation,
  openUsagePeriod,
  closeUsagePeriod,
  debitCredits,
  getHourlyRate,
} from '@/lib/console/storage/credits'

interface HeartbeatMachine {
  machine_id: string
  machine_type: string
  started_at: string
}

interface HeartbeatVolume {
  volume_id: string
  volume_type: 'vm' | 'object'
  size_gb: number
  created_at: string
}

async function reconcileUsagePeriods(
  installationId: string,
  orgId: string,
  machines: HeartbeatMachine[],
  volumes: HeartbeatVolume[],
): Promise<void> {
  const activePeriods = await getActiveUsagePeriodsForInstallation(installationId)

  // Build lookup sets
  const reportedMachineIds = new Set(machines.map((m) => m.machine_id))
  const reportedVolumeIds = new Set(volumes.map((v) => v.volume_id))
  const reportedResourceIds = new Set([...reportedMachineIds, ...reportedVolumeIds])

  const activePeriodResourceIds = new Set(activePeriods.map((p) => p.resourceId))

  // Resources running but no open period (missed start event)
  for (const machine of machines) {
    if (!activePeriodResourceIds.has(machine.machine_id)) {
      const hourlyRate = await getHourlyRate(machine.machine_type)
      await openUsagePeriod({
        installationId,
        orgId,
        resourceId: machine.machine_id,
        resourceType: machine.machine_type,
        hourlyRate,
      })
    }
  }

  for (const volume of volumes) {
    if (!activePeriodResourceIds.has(volume.volume_id)) {
      const pricingType = volume.volume_type === 'vm' ? 'storage.vm' : 'storage.object'
      const baseRate = await getHourlyRate(pricingType)
      const hourlyRate = baseRate * volume.size_gb
      await openUsagePeriod({
        installationId,
        orgId,
        resourceId: volume.volume_id,
        resourceType: pricingType,
        hourlyRate,
      })
    }
  }

  // Open period but resource not reported (missed stop event)
  for (const period of activePeriods) {
    // Skip controlplane resources — they don't appear in heartbeat
    if (period.resourceType.startsWith('controlplane.')) {
      continue
    }

    if (!reportedResourceIds.has(period.resourceId)) {
      const closedPeriod = await closeUsagePeriod(period.resourceId, installationId)
      if (closedPeriod && closedPeriod.totalCost > 0) {
        await debitCredits(
          orgId,
          closedPeriod.totalCost,
          `Usage: ${closedPeriod.resourceType} ${closedPeriod.resourceId} (auto-closed by heartbeat)`,
        )
      }
    }
  }
}

// Use Node.js runtime for Supabase (uses Node.js APIs)
export const runtime = 'nodejs'
/**
 * Verify installation key (POST method)
 * Used by installation script to verify the key and get user info
 * Also used by deployment to poll for configuration (every 10 minutes)
 */
export async function POST(request: NextRequest) {
  try {
    const body = await request.json()
    const { installationKey, provider, region, running_machines, volumes } = body

    if (!installationKey) {
      return apiError('Installation key is required', 400)
    }

    // Look up installation by key
    const installation = await getInstallationByKey(installationKey)

    if (!installation) {
      return apiError('Invalid installation key', 404)
    }

    // Generate secret key on first verification (if not exists)
    let updatedInstallation = installation
    if (!installation.secretKey) {
      console.log('First deployment verification for installation:', installation.id)
      console.log('Generating secret key for installation key:', installationKey)

      const secretKey = crypto.randomUUID()

      // Atomically mark installation complete and set secret key
      updatedInstallation = await markInstallationComplete(installation.id, secretKey)

      console.log('Secret key generated and installation marked as complete')
    }

    // Update cloud provider and location if provided (from kli install)
    if (provider || region) {
      const updates: { cloudProvider?: 'aws' | 'gcp' | 'azure'; cloudLocation?: string } = {}
      if (provider && ['aws', 'gcp', 'azure'].includes(provider)) {
        updates.cloudProvider = provider as 'aws' | 'gcp' | 'azure'
      }
      if (region) {
        updates.cloudLocation = region
      }
      if (Object.keys(updates).length > 0) {
        updatedInstallation = await updateInstallation(installation.id, updates)
      }
    }

    // Atomically update last health check timestamp (deployment is polling)
    updatedInstallation = await updateHealthCheck(installation.id)

    // Reconcile usage periods from heartbeat data (if provided)
    if (running_machines || volumes) {
      try {
        await reconcileUsagePeriods(
          installation.id,
          installation.orgId,
          running_machines ?? [],
          volumes ?? [],
        )
      } catch (reconcileError) {
        console.error('Heartbeat reconciliation error:', reconcileError)
        // Don't let reconciliation failures break the verify-key response
      }
    }

    // Return only operational information needed by deployment
    const response = NextResponse.json({
      success: true,
      installationId: updatedInstallation.id,
      secretKey: updatedInstallation.secretKey,
      subdomain: updatedInstallation.subdomain,
      deploymentReady: updatedInstallation.deploymentReady || false,
      dnsConfigurations: updatedInstallation.dnsConfigurations || [],
      cloudProvider: updatedInstallation.cloudProvider || null,
    })

    // Disable all caching
    response.headers.set('Cache-Control', 'no-store, no-cache, must-revalidate, max-age=0')
    response.headers.set('Pragma', 'no-cache')
    response.headers.set('Expires', '0')

    return response
  } catch (error) {
    console.error('Verification error:', error)
    return apiError('Internal server error', 500)
  }
}
