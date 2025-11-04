import { NextRequest, NextResponse } from 'next/server'
import { createHmac, randomBytes } from 'crypto'
import { getInstallationSession } from '@/lib/console-auth'
import { createClient } from '@/lib/console/supabase-server'

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
 */
export async function POST(
  request: NextRequest,
  { params }: { params: { id: string } }
) {
  try {
    // Verify user is logged in and has access to this installation
    const session = await getInstallationSession()
    if (!session || !session.userId) {
      return NextResponse.json(
        { error: 'Unauthorized - please log in' },
        { status: 401 }
      )
    }

    const installationId = params.id

    // Verify installation exists and user has access
    const supabase = createClient()
    const { data: installation, error: installationError } = await supabase
      .from('installations')
      .select('id, installation_key, user_id, secret_key, subdomain')
      .eq('id', installationId)
      .single()

    if (installationError || !installation) {
      return NextResponse.json(
        { error: 'Installation not found' },
        { status: 404 }
      )
    }

    // Verify user owns this installation
    if (installation.user_id !== session.userId) {
      return NextResponse.json(
        { error: 'Forbidden - you do not own this installation' },
        { status: 403 }
      )
    }

    // Use installation secret as signing key
    const installationSecret = installation.secret_key
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
      installationKey: installation.installation_key,
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
    return NextResponse.json(
      { error: 'Internal server error' },
      { status: 500 }
    )
  }
}
