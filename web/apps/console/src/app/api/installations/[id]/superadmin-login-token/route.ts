import { NextRequest, NextResponse } from 'next/server'
import { createHmac, randomBytes } from 'crypto'
import { requireInstallationAccess } from '@/lib/console/authorization'
import { getInstallationById } from '@/lib/console/supabase-storage-service'

// Super admin login token validity (5 minutes)
const TOKEN_VALIDITY_MS = 5 * 60 * 1000

interface SuperAdminLoginTokenPayload {
  type: 'superadmin-login'
  installationId: string
  installationKey: string
  timestamp: number
  nonce: string
  expiresAt: number
}

/**
 * Generate super admin login token
 *
 * This creates a time-limited token that allows one-click super admin login
 * to the installation dashboard. Token is valid for 5 minutes.
 * Any team member can access super admin login.
 */
export async function POST(
  _request: NextRequest,
  { params }: { params: Promise<{ id: string }> }
) {
  try {
    const { id } = await params

    // Verify user has access to this installation (any role)
    await requireInstallationAccess(id)

    const installationId = id

    // Fetch installation for secret key
    const installation = await getInstallationById(installationId)

    if (!installation) {
      return NextResponse.json(
        { error: 'Installation not found' },
        { status: 404 }
      )
    }

    // Use installation secret as signing key
    const installationSecret = installation.secretKey
    if (!installationSecret) {
      return NextResponse.json(
        { error: 'Installation secret not available - complete installation first' },
        { status: 400 }
      )
    }

    // Generate token payload
    const now = Date.now()
    const expiresAt = now + TOKEN_VALIDITY_MS
    const nonce = randomBytes(16).toString('hex')

    const payload: SuperAdminLoginTokenPayload = {
      type: 'superadmin-login',
      installationId: installation.id,
      installationKey: installation.installationKey,
      timestamp: now,
      nonce,
      expiresAt,
    }

    // Create HMAC signature using installation secret
    const payloadString = JSON.stringify(payload)
    const signature = createHmac('sha256', installationSecret)
      .update(payloadString)
      .digest('base64url')

    // Combine payload and signature into token
    const token = `${Buffer.from(payloadString).toString('base64url')}.${signature}`

    // Construct login URL pointing to installation dashboard
    if (!installation.subdomain) {
      return NextResponse.json(
        { error: 'Installation subdomain not configured' },
        { status: 400 }
      )
    }

    const domain = process.env.CLOUDFLARE_DNS_DOMAIN || 'khost.dev'
    const loginUrl = `https://${installation.subdomain}.${domain}/superadmin-login?token=${token}`

    return NextResponse.json({
      token,
      loginUrl,
      expiresAt: new Date(expiresAt).toISOString(),
      validForSeconds: Math.floor(TOKEN_VALIDITY_MS / 1000),
    })
  } catch (error) {
    console.error('Error generating superadmin login token:', error)
    const message = error instanceof Error ? error.message : 'Internal server error'
    const status = message.includes('Unauthorized') ? 401 : message.includes('Forbidden') ? 403 : 500
    return NextResponse.json({ error: message }, { status })
  }
}
