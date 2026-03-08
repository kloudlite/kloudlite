import { NextRequest, NextResponse } from 'next/server'
import { apiError } from '@/lib/api-helpers'
import { getInstallationByKey } from '@/lib/console/storage/installations'
import {
  insertUsageEvent,
  openUsagePeriod,
  closeUsagePeriod,
  debitCredits,
  getHourlyRate,
} from '@/lib/console/storage/credits'

// Use Node.js runtime for Supabase (uses Node.js APIs)
export const runtime = 'nodejs'

const VALID_EVENT_TYPES = [
  'workmachine.started',
  'workmachine.stopped',
  'workmachine.resized',
  'controlplane.started',
  'controlplane.stopped',
  'storage.provisioned',
  'storage.resized',
  'storage.deleted',
] as const

type EventType = (typeof VALID_EVENT_TYPES)[number]

function isValidEventType(value: string): value is EventType {
  return (VALID_EVENT_TYPES as readonly string[]).includes(value)
}

function isStartEvent(eventType: EventType): boolean {
  return eventType.endsWith('.started') || eventType === 'storage.provisioned'
}

function isStopEvent(eventType: EventType): boolean {
  return eventType.endsWith('.stopped') || eventType === 'storage.deleted'
}

function isResizeEvent(eventType: EventType): boolean {
  return eventType.endsWith('.resized')
}

function isStorageEvent(eventType: EventType): boolean {
  return eventType.startsWith('storage.')
}

/**
 * Receive usage events from in-cluster WorkMachine controller.
 * Auth: x-installation-key header (cluster-to-API, not session-based).
 */
export async function POST(request: NextRequest) {
  try {
    // 1. Validate installation key header
    const installationKey = request.headers.get('x-installation-key')
    if (!installationKey) {
      return apiError('Invalid installation key', 401)
    }

    // 2. Look up installation
    const installation = await getInstallationByKey(installationKey)
    if (!installation) {
      return apiError('Invalid installation key', 401)
    }

    // 3. Parse and validate request body
    const body = await request.json()
    const { event_type, resource_id, resource_type, metadata, timestamp } = body

    if (!event_type || !resource_id || !timestamp) {
      return apiError('Missing required fields', 400)
    }

    if (!isValidEventType(event_type)) {
      return apiError('Invalid event type', 400)
    }

    const orgId = installation.orgId

    // 4. Insert usage event (idempotent — duplicates return null)
    const usageEvent = await insertUsageEvent({
      installationId: installation.id,
      eventType: event_type,
      resourceId: resource_id,
      resourceType: resource_type,
      metadata,
      eventTimestamp: timestamp,
    })

    if (!usageEvent) {
      // Duplicate event — return 200 OK for idempotency, skip processing
      const response = NextResponse.json({ success: true })
      response.headers.set('Cache-Control', 'no-store')
      return response
    }

    // 5. Process based on event type
    if (isStartEvent(event_type)) {
      let hourlyRate = await getHourlyRate(resource_type)

      // For storage events, multiply rate by size_gb (rate is per GB-hour)
      if (isStorageEvent(event_type) && metadata?.size_gb) {
        hourlyRate = hourlyRate * Number(metadata.size_gb)
      }

      await openUsagePeriod({
        installationId: installation.id,
        orgId,
        resourceId: resource_id,
        resourceType: resource_type,
        hourlyRate,
      })
    } else if (isStopEvent(event_type)) {
      const closedPeriod = await closeUsagePeriod(resource_id, installation.id)

      if (closedPeriod && closedPeriod.totalCost > 0) {
        const hours = closedPeriod.endedAt
          ? (
              (Date.parse(closedPeriod.endedAt) - Date.parse(closedPeriod.startedAt)) /
              (1000 * 60 * 60)
            ).toFixed(1)
          : '0'
        const description = `Usage: ${closedPeriod.resourceType} ${resource_id} (${hours} hours)`
        await debitCredits(orgId, closedPeriod.totalCost, description, closedPeriod.id)
      }
    } else if (isResizeEvent(event_type)) {
      // Close current period and debit accrued cost
      const closedPeriod = await closeUsagePeriod(resource_id, installation.id)

      if (closedPeriod && closedPeriod.totalCost > 0) {
        const hours = closedPeriod.endedAt
          ? (
              (Date.parse(closedPeriod.endedAt) - Date.parse(closedPeriod.startedAt)) /
              (1000 * 60 * 60)
            ).toFixed(1)
          : '0'
        const description = `Usage: ${closedPeriod.resourceType} ${resource_id} (${hours} hours)`
        await debitCredits(orgId, closedPeriod.totalCost, description, closedPeriod.id)
      }

      // Open new period with new resource type and rate
      const newResourceType = metadata?.new_type ?? resource_type
      let newHourlyRate = await getHourlyRate(newResourceType as string)

      // For storage resize, use new_size_gb for rate calculation
      if (isStorageEvent(event_type) && metadata?.new_size_gb) {
        newHourlyRate = newHourlyRate * Number(metadata.new_size_gb)
      }

      await openUsagePeriod({
        installationId: installation.id,
        orgId,
        resourceId: resource_id,
        resourceType: newResourceType as string,
        hourlyRate: newHourlyRate,
      })
    }

    const response = NextResponse.json({ success: true })
    response.headers.set('Cache-Control', 'no-store')
    return response
  } catch (error) {
    console.error('Usage event error:', error)
    return apiError('Internal server error', 500)
  }
}
