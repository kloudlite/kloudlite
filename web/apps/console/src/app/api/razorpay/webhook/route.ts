import { NextResponse } from 'next/server'
import crypto from 'crypto'
import {
  getSubscriptionByRazorpayId,
  updateSubscriptionStatus,
  upsertInvoice,
} from '@/lib/console/storage'

function verifyWebhookSignature(body: string, signature: string, secret: string): boolean {
  const expected = crypto.createHmac('sha256', secret).update(body).digest('hex')
  return crypto.timingSafeEqual(Buffer.from(expected, 'hex'), Buffer.from(signature, 'hex'))
}

export async function POST(request: Request) {
  const secret = process.env.RAZORPAY_WEBHOOK_SECRET
  if (!secret) {
    console.error('RAZORPAY_WEBHOOK_SECRET not configured')
    return NextResponse.json({ error: 'Webhook secret not configured' }, { status: 500 })
  }

  const body = await request.text()
  const signature = request.headers.get('x-razorpay-signature')

  if (!signature) {
    return NextResponse.json({ error: 'Missing signature' }, { status: 401 })
  }

  // Verify HMAC signature using timing-safe comparison
  try {
    if (!verifyWebhookSignature(body, signature, secret)) {
      return NextResponse.json({ error: 'Invalid signature' }, { status: 401 })
    }
  } catch {
    return NextResponse.json({ error: 'Invalid signature format' }, { status: 401 })
  }

  let event: Record<string, unknown>
  try {
    event = JSON.parse(body)
  } catch {
    return NextResponse.json({ error: 'Invalid JSON body' }, { status: 400 })
  }

  const eventType = event.event as string

  try {
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const payload = event.payload as any

    switch (eventType) {
      case 'subscription.activated': {
        const sub = payload.subscription.entity
        await updateSubscriptionStatus(
          sub.id,
          'active',
          sub.current_start ? new Date(sub.current_start * 1000).toISOString() : undefined,
          sub.current_end ? new Date(sub.current_end * 1000).toISOString() : undefined,
        )
        break
      }
      case 'subscription.authenticated': {
        const sub = payload.subscription.entity
        await updateSubscriptionStatus(sub.id, 'authenticated')
        break
      }
      case 'subscription.charged': {
        const sub = payload.subscription.entity
        await updateSubscriptionStatus(
          sub.id,
          'active',
          sub.current_start ? new Date(sub.current_start * 1000).toISOString() : undefined,
          sub.current_end ? new Date(sub.current_end * 1000).toISOString() : undefined,
        )
        break
      }
      case 'subscription.paused': {
        const sub = payload.subscription.entity
        await updateSubscriptionStatus(sub.id, 'paused')
        break
      }
      case 'subscription.cancelled': {
        const sub = payload.subscription.entity
        await updateSubscriptionStatus(sub.id, 'cancelled')
        break
      }
      case 'subscription.completed': {
        const sub = payload.subscription.entity
        await updateSubscriptionStatus(sub.id, 'expired')
        break
      }
      case 'payment.failed': {
        const payment = payload.payment?.entity
        console.error('Payment failed:', payment?.id, payment?.error_description)
        break
      }
      case 'payment.succeeded': {
        const payment = payload.payment?.entity
        if (!payment) break

        // Look up the subscription from our DB using the order_id from payment notes
        const orderId = payment.order_id
        if (orderId) {
          const subscription = await getSubscriptionByRazorpayId(orderId)
          if (subscription) {
            await upsertInvoice({
              subscriptionId: subscription.id,
              installationId: subscription.installationId,
              razorpayInvoiceId: orderId,
              razorpayPaymentId: payment.id,
              amount: payment.amount,
              currency: (payment.currency as string)?.toUpperCase() || 'INR',
              status: 'paid',
              billingStart: subscription.currentStart ?? new Date().toISOString(),
              billingEnd: subscription.currentEnd ?? new Date().toISOString(),
              paidAt: new Date().toISOString(),
            })
          }
        }
        break
      }
      default:
        console.log('Unhandled webhook event:', eventType)
    }

    return NextResponse.json({ status: 'ok' })
  } catch (error) {
    console.error('Webhook processing error:', { eventType, error })
    return NextResponse.json({ error: 'Processing failed' }, { status: 500 })
  }
}
