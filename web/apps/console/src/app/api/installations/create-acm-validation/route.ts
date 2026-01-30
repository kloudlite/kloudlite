import { NextRequest, NextResponse } from 'next/server'
import { getInstallationByKey } from '@/lib/console/storage'
import { createCnameRecord } from '@/lib/console/cloudflare-dns'

// Use Node.js runtime for Supabase
export const runtime = 'nodejs'

/**
 * Create DNS validation CNAME records in Cloudflare for ACM certificate
 * This is called by kli after requesting an ACM certificate
 *
 * POST /api/installations/create-acm-validation
 * Authorization: Bearer {secretKey}
 * Body: {
 *   installationKey: string,
 *   validationRecords: Array<{
 *     name: string,    // "_abc123.subdomain.khost.dev"
 *     value: string    // "_xyz789.acm-validations.aws."
 *   }>
 * }
 *
 * Response: {
 *   success: boolean,
 *   recordIds: string[],
 *   message?: string
 * }
 */
export async function POST(request: NextRequest) {
  try {
    // Validate authorization
    const authHeader = request.headers.get('authorization')
    if (!authHeader || !authHeader.startsWith('Bearer ')) {
      return NextResponse.json(
        { success: false, error: 'Missing or invalid authorization header' },
        { status: 401 }
      )
    }

    const secretKey = authHeader.substring(7)
    const body = await request.json()
    const { installationKey, validationRecords } = body

    // Validate required fields
    if (!installationKey) {
      return NextResponse.json(
        { success: false, error: 'installationKey is required' },
        { status: 400 }
      )
    }

    if (!validationRecords || !Array.isArray(validationRecords) || validationRecords.length === 0) {
      return NextResponse.json(
        { success: false, error: 'validationRecords array is required and must not be empty' },
        { status: 400 }
      )
    }

    // Validate each record has name and value
    for (const record of validationRecords) {
      if (!record.name || !record.value) {
        return NextResponse.json(
          { success: false, error: 'Each validation record must have name and value' },
          { status: 400 }
        )
      }
    }

    // Get installation
    const installation = await getInstallationByKey(installationKey)
    if (!installation) {
      return NextResponse.json(
        { success: false, error: 'Invalid installation key' },
        { status: 404 }
      )
    }

    // Verify secret key
    if (installation.secretKey !== secretKey) {
      return NextResponse.json(
        { success: false, error: 'Invalid secret key' },
        { status: 403 }
      )
    }

    // Verify installation has a subdomain
    if (!installation.subdomain) {
      return NextResponse.json(
        { success: false, error: 'Installation must have a subdomain assigned first' },
        { status: 400 }
      )
    }

    // Create CNAME records for ACM validation
    const recordIds: string[] = []
    const errors: string[] = []

    console.log(`Creating ${validationRecords.length} ACM validation records for installation: ${installationKey}`)

    for (const record of validationRecords) {
      try {
        // ACM validation records should not be proxied
        const recordId = await createCnameRecord(record.name, record.value, false)
        if (recordId) {
          recordIds.push(recordId)
          console.log(`Created ACM validation record: ${record.name} -> ${record.value} (ID: ${recordId})`)
        } else {
          errors.push(`Failed to create record for ${record.name}`)
          console.error(`Failed to create ACM validation record: ${record.name}`)
        }
      } catch (error) {
        const errorMessage = error instanceof Error ? error.message : 'Unknown error'
        errors.push(`Error creating record for ${record.name}: ${errorMessage}`)
        console.error(`Error creating ACM validation record ${record.name}:`, error)
      }
    }

    // Determine success based on whether we created any records
    const success = recordIds.length > 0 && recordIds.length === validationRecords.length

    const response = NextResponse.json({
      success,
      recordIds,
      created: recordIds.length,
      total: validationRecords.length,
      errors: errors.length > 0 ? errors : undefined,
      message: success
        ? 'All validation records created successfully'
        : errors.length > 0
          ? 'Some validation records failed to create'
          : 'No validation records created',
    })

    // Disable all caching
    response.headers.set('Cache-Control', 'no-store, no-cache, must-revalidate, max-age=0')
    response.headers.set('Pragma', 'no-cache')
    response.headers.set('Expires', '0')

    return response
  } catch (error) {
    console.error('Create ACM validation error:', error)
    return NextResponse.json(
      { success: false, error: 'Internal server error' },
      { status: 500 }
    )
  }
}
