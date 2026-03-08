export const DEV_TOKEN = 'dev-superadmin'

export const TIMEOUTS = {
  action: 10_000,
  navigation: 15_000,
  auth: 15_000,
  stripeCheckout: 60_000,
  deployment: 600_000,
  dns: 900_000,
  providerTest: 1_800_000,
}

export const STRIPE_TEST_CARD = {
  number: '4242424242424242',
  expiry: '12/30',
  cvc: '123',
  name: 'Test User',
}
