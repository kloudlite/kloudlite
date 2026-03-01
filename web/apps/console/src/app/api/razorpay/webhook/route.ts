import { NextResponse } from 'next/server'
import crypto from 'crypto'
import {
  getSubscriptionByRazorpayId,
  updateSubscriptionStatus,
  upsertInvoice,
} from '@/lib/console/storage'

export const runtime = 'nodejs'

function verifyWebhookSignature(body: string, signature: string, secret: string): boolean {
  const expectedSignature = crypto
    .createHmac('sha256', secret)
    .update(body)
    .digest('hex')
  return crypto.timingSafeEqual(
    Buffer.from(expectedSignature),
    Buffer.from(signature),
  )
}

export async function POST(request: Request) {
  const secret = process.env.RAZORPAY_WEBHOOK_SECRET
  if (!secret) {
    console.error('RAZORPAY_WEBHOOK_SECRET not configured')
    return NextResponse.json({ error: 'Webhook secret not configured' }, { status: 500 })
  }

  const body = await request.text()
  const signature = request.headers.get('x-razorpay-signature')

  if (!signature || !verifyWebhookSignature(body, signature, secret)) {
    return NextResponse.json({ error: 'Invalid signature' }, { status: 401 })
  }

  const event = JSON.parse(body)
  const eventType = event.event as string

  try {
    switch (eventType) {
      case 'subscription.activated': {
        const sub = event.payload.subscription.entity
        await updateSubscriptionStatus(
          sub.id,
          'active',
          sub.current_start ? new Date(sub.current_start * 1000).toISOString() : undefined,
          sub.current_end ? new Date(sub.current_end * 1000).toISOString() : undefined,
        )
        break
      }

      case 'subscription.charged': {
        const sub = event.payload.subscription.entity
        const payment = event.payload.payment?.entity

        await updateSubscriptionStatus(
          sub.id,
          'active',
          sub.current_start ? new Date(sub.current_start * 1000).toISOString() : undefined,
          sub.current_end ? new Date(sub.current_end * 1000).toISOString() : undefined,
        )

        if (payment) {
          const subscription = await getSubscriptionByRazorpayId(sub.id)
          if (subscription) {
            await upsertInvoice({
              subscriptionId: subscription.id,
              installationId: subscription.installationId,
              razorpayInvoiceId: payment.invoice_id || payment.id,
              razorpayPaymentId: payment.id,
              amount: payment.amount,
              currency: payment.currency?.toUpperCase() || 'USD',
              status: 'paid',
              billingStart: sub.current_start
                ? new Date(sub.current_start * 1000).toISOString()
                : undefined,
              billingEnd: sub.current_end
                ? new Date(sub.current_end * 1000).toISOString()
                : undefined,
              paidAt: new Date().toISOString(),
            })
          }
        }
        break
      }

      case 'subscription.paused': {
        const sub = event.payload.subscription.entity
        await updateSubscriptionStatus(sub.id, 'paused')
        break
      }

      case 'subscription.cancelled': {
        const sub = event.payload.subscription.entity
        await updateSubscriptionStatus(sub.id, 'cancelled')
        break
      }

      case 'subscription.completed': {
        const sub = event.payload.subscription.entity
        await updateSubscriptionStatus(sub.id, 'expired')
        break
      }

      case 'payment.failed': {
        const payment = event.payload.payment?.entity
        console.error('Payment failed:', payment?.id, payment?.error_description)
        break
      }

      default:
        console.log('Unhandled webhook event:', eventType)
    }

    return NextResponse.json({ status: 'ok' })
  } catch (error) {
    console.error('Webhook processing error:', error)
    return NextResponse.json({ error: 'Processing failed' }, { status: 500 })
  }
}
