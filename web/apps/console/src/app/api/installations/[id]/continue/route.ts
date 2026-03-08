import { NextResponse } from 'next/server'
import { getRegistrationSession } from '@/lib/console-auth'
import { getInstallationById, getStripeCustomer, upsertStripeCustomer, syncSubscriptionItemsFromStripe, getSubscriptionItems } from '@/lib/console/storage'
import { getStripe } from '@/lib/stripe'
import { SignJWT } from 'jose'
import { cookies } from 'next/headers'

function getPublicOrigin(request: Request): string {
  const proto = request.headers.get('x-forwarded-proto') || 'https'
  const host = request.headers.get('x-forwarded-host') || request.headers.get('host') || ''
  return `${proto}://${host}`
}

/**
 * Continue API route - loads installation context and redirects to the appropriate step
 */
export async function GET(request: Request, { params }: { params: Promise<{ id: string }> }) {
  const { id } = await params
  const origin = getPublicOrigin(request)
  const session = await getRegistrationSession()

  if (!session?.user) {
    return NextResponse.redirect(new URL('/login', origin))
  }

  // Fetch the installation
  const installation = await getInstallationById(id)

  if (!installation) {
    console.error('Installation not found:', id)
    return NextResponse.redirect(new URL('/installations', origin))
  }

  // Verify user owns this installation
  if (installation.userId !== session.user.id) {
    return NextResponse.redirect(new URL('/installations', origin))
  }

  // Update session cookie with this installation's key
  const secret = new TextEncoder().encode(process.env.NEXTAUTH_SECRET)
  const token = await new SignJWT({
    provider: session.provider,
    email: session.user.email,
    name: session.user.name,
    image: session.user.image,
    installationKey: installation.installationKey,
    userId: session.user.id,
  })
    .setProtectedHeader({ alg: 'HS256' })
    .setIssuedAt()
    .setExpirationTime('30d')
    .sign(secret)

  const cookieStore = await cookies()
  cookieStore.set('registration_session', token, {
    httpOnly: true,
    secure: process.env.NODE_ENV === 'production',
    sameSite: 'lax',
    maxAge: 30 * 24 * 60 * 60, // 30 days
  })

  // Helper function to validate subdomain
  const isValidSubdomain = (subdomain: string | null | undefined): boolean => {
    if (!subdomain) return false
    if (subdomain === '0.0.0.0') return false
    if (subdomain.includes('0.0.0.0')) return false
    return true
  }

  // Determine next step based on installation status and hosting type
  const isKloudliteCloud = installation.cloudProvider === 'oci'
  let redirectPath: string

  if (isKloudliteCloud) {
    // Kloudlite Cloud — check Stripe subscription before deploy
    let customer = await getStripeCustomer(id)
    let hasActiveSub = customer?.billingStatus === 'active'

    // Handle webhook race condition: if DB still shows incomplete but
    // the customer has a subscription, check Stripe directly
    if (!hasActiveSub && customer?.stripeCustomerId) {
      try {
        const stripe = getStripe()
        const subs = await stripe.subscriptions.list({
          customer: customer.stripeCustomerId,
          status: 'active',
          limit: 1,
        })
        if (subs.data.length > 0) {
          // Stripe confirms active — update local DB so webhook can catch up
          const periodEnd = subs.data[0].items.data[0]?.current_period_end
          await upsertStripeCustomer({
            installationId: id,
            stripeCustomerId: customer.stripeCustomerId,
            stripeSubscriptionId: subs.data[0].id,
            billingStatus: 'active',
            currentPeriodEnd: periodEnd
              ? new Date(periodEnd * 1000).toISOString()
              : null,
          })
          hasActiveSub = true
        }
      } catch (err) {
        console.error('[continue] Failed to verify subscription with Stripe:', err)
      }
    }

    // Sync subscription items if DB is empty (webhook may not have fired yet)
    if (hasActiveSub) {
      const existingItems = await getSubscriptionItems(id)
      if (existingItems.length === 0 && customer?.stripeSubscriptionId) {
        await syncSubscriptionItemsFromStripe(id, customer.stripeSubscriptionId)
      }
    }

    if (!hasActiveSub) {
      // No subscription yet — go back to plan/payment page
      redirectPath = `/installations/new-kl-cloud?installation=${id}`
    } else if (!installation.deploymentReady) {
      // Subscribed — go to deploy page
      redirectPath = '/installations/new/kloudlite-cloud'
    } else {
      redirectPath = '/installations'
    }
  } else {
    // BYOC flow
    if (!installation.secretKey) {
      redirectPath = '/installations/new/install'
    } else if (!isValidSubdomain(installation.subdomain)) {
      redirectPath = '/installations/new/domain'
    } else if (!installation.deploymentReady) {
      redirectPath = '/installations/new/complete'
    } else {
      redirectPath = '/installations'
    }
  }

  return NextResponse.redirect(new URL(redirectPath, origin))
}
