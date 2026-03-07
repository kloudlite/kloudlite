import Stripe from 'stripe'

let instance: Stripe | null = null

export function getStripe(): Stripe {
  if (!instance) {
    const secretKey = process.env.STRIPE_SECRET_KEY
    if (!secretKey) {
      throw new Error('STRIPE_SECRET_KEY must be set')
    }
    instance = new Stripe(secretKey)
  }
  return instance
}
