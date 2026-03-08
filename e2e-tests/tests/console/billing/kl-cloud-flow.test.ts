import { test, expect } from '@playwright/test'
import { devLogin } from '../../../lib/helpers'
import { TIMEOUTS, STRIPE_TEST_CARD } from '../../../lib/constants'

test.use({ storageState: { cookies: [], origins: [] } })

/**
 * End-to-end test for the KL Cloud pay-as-you-go credits billing flow:
 * Login → Add Credits (top-up via Stripe Invoice) → Create Installation → Verify billing state
 *
 * This replaces the old subscription-based flow. Credits are added via Stripe hosted invoices
 * (not Checkout Sessions), and installations are gated by credit balance rather than
 * active subscriptions.
 */
test.describe.serial('KL Cloud Billing Flow (Pay-as-you-go Credits)', () => {
  const testId = Date.now().toString(36)
  const installationName = `E2E Test ${testId}`
  const subdomain = `e2e-${testId}`

  test('login and verify dashboard', async ({ page }) => {
    test.setTimeout(30_000)

    await devLogin(page)

    await expect(page).toHaveURL(/\/installations$/)
    await expect(
      page.getByRole('heading', { name: 'Installations' }),
    ).toBeVisible()
    console.log('[test] Logged in successfully')
  })

  test('navigate to billing settings and verify initial state', async ({
    page,
  }) => {
    test.setTimeout(30_000)

    await devLogin(page)
    await page.goto('/installations/settings/billing')
    await page.waitForLoadState('networkidle')

    // Billing page should show Credit Balance card
    await expect(page.getByText('Credit Balance')).toBeVisible({
      timeout: TIMEOUTS.action,
    })
    console.log('[test] Credit Balance card visible')

    // "Add Credits" button should exist
    await expect(
      page.getByRole('button', { name: /add credits/i }),
    ).toBeVisible()
    console.log('[test] Add Credits button visible')

    // Transaction History section should exist (may show "No transactions yet")
    await expect(page.getByText('Transaction History')).toBeVisible()
    console.log('[test] Transaction History section visible')

    // Active Usage section should exist
    await expect(page.getByText('Active Usage')).toBeVisible()
    console.log('[test] Billing settings initial state verified')
  })

  test('add credits via top-up through Stripe invoice', async ({ page }) => {
    test.setTimeout(120_000)

    await devLogin(page)
    await page.goto('/installations/settings/billing')
    await page.waitForLoadState('networkidle')

    // Wait for the credit balance to load
    await expect(page.getByText('Credit Balance')).toBeVisible({
      timeout: TIMEOUTS.action,
    })

    // ==================== Step 1: Open Add Credits dialog ====================
    await page.getByRole('button', { name: /add credits/i }).click()

    // Wait for dialog to appear
    await expect(
      page.getByRole('heading', { name: /add credits/i }),
    ).toBeVisible()
    console.log('[test] Add Credits dialog opened')

    // ==================== Step 2: Enter top-up amount ====================
    const amountInput = page.locator('#topup-amount')
    await amountInput.clear()
    await amountInput.fill('100')
    console.log('[test] Entered top-up amount: $100')

    // ==================== Step 3: Submit — redirects to Stripe hosted invoice ====================
    // Click the "Add $100.00" button in the dialog
    const addButton = page
      .getByRole('button', { name: /add \$100/i })
    await expect(addButton).toBeEnabled()
    await addButton.click()
    console.log('[test] Clicked Add Credits, waiting for Stripe redirect...')

    // ==================== Step 4: Handle Stripe hosted invoice page ====================
    await page.waitForURL(/invoice\.stripe\.com/, {
      timeout: TIMEOUTS.stripeCheckout,
    })
    console.log(`[test] On Stripe Invoice page: ${page.url()}`)

    // Wait for the invoice page to render
    await page.waitForLoadState('domcontentloaded')

    // Stripe hosted invoice page has a "Pay" button to expand the card form.
    // Wait for either a "Pay" button or the card form to appear directly.
    const payButton = page.getByRole('button', { name: /pay/i })
    const cardNumberField = page.getByPlaceholder('1234 1234 1234 1234')

    await expect(payButton.or(cardNumberField).first()).toBeVisible({
      timeout: 30_000,
    })

    // If there's a "Pay" button that needs to be clicked to reveal card form, click it
    if (await payButton.isVisible().catch(() => false)) {
      if (!(await cardNumberField.isVisible().catch(() => false))) {
        await payButton.click()
        console.log('[test] Clicked Pay button to expand card form')
      }
    }

    // Wait for card form fields to appear
    await expect(cardNumberField).toBeVisible({ timeout: 15_000 })
    console.log('[test] Card form visible')

    // ==================== Step 5: Fill card details ====================
    // Stripe invoice page has card fields directly on the page (similar to Checkout).
    await cardNumberField.click()
    await cardNumberField.pressSequentially(STRIPE_TEST_CARD.number, {
      delay: 50,
    })
    console.log('[test] Filled card number')

    const expiryField = page.getByPlaceholder('MM / YY')
    await expiryField.click()
    await expiryField.pressSequentially(STRIPE_TEST_CARD.expiry, { delay: 50 })
    console.log('[test] Filled expiry')

    const cvcField = page
      .getByPlaceholder('CVC')
      .or(page.getByRole('textbox', { name: 'CVC' }))
      .first()
    await cvcField.click()
    await cvcField.pressSequentially(STRIPE_TEST_CARD.cvc, { delay: 50 })
    console.log('[test] Filled CVC')

    // Cardholder name (if present)
    const nameField = page.getByPlaceholder('Full name on card')
    if (await nameField.isVisible({ timeout: 2_000 }).catch(() => false)) {
      await nameField.fill(STRIPE_TEST_CARD.name)
      console.log('[test] Filled cardholder name')
    }

    // ==================== Step 6: Submit payment ====================
    // On Stripe invoice page, the submit button says "Pay $X.XX"
    const submitPayButton = page.getByRole('button', { name: /pay \$/i })
    await expect(submitPayButton).toBeVisible({ timeout: 10_000 })
    await submitPayButton.click()
    console.log('[test] Clicked Pay, waiting for confirmation...')

    // ==================== Step 7: Wait for redirect back to app ====================
    // After payment, Stripe redirects back to our app
    await page.waitForURL(
      (url) => !url.href.includes('invoice.stripe.com'),
      { timeout: TIMEOUTS.stripeCheckout },
    )

    const returnUrl = page.url()
    console.log(`[test] Returned to app: ${returnUrl}`)

    // Should be back on the billing settings page or the app
    // Wait for the page to settle and show updated balance
    await page.waitForLoadState('networkidle')

    // Verify credit balance has updated (should be ~$100)
    // The balance text is inside a div with the formatted currency
    await expect(page.getByText(/\$\d+\.\d{2}/)).toBeVisible({
      timeout: TIMEOUTS.action,
    })
    console.log('[test] Credit balance visible after top-up')
  })

  test('create new installation with sufficient balance', async ({ page }) => {
    test.setTimeout(120_000)

    await devLogin(page)

    // ==================== Step 1: Navigate to KL Cloud form ====================
    await page.getByRole('button', { name: 'New Installation' }).click()
    await page.getByText('Kloudlite Cloud').click()
    await page.waitForURL('**/installations/new-kl-cloud')

    await expect(
      page.getByRole('heading', { name: 'Create Installation' }),
    ).toBeVisible()
    console.log('[test] On KL Cloud form')

    // ==================== Step 2: Fill installation details ====================
    await page.getByPlaceholder('e.g., Production').fill(installationName)
    await page.getByRole('textbox', { name: 'your-company' }).fill(subdomain)

    // Wait for subdomain availability check
    await expect(page.getByText('Domain is available')).toBeVisible({
      timeout: TIMEOUTS.action,
    })
    console.log('[test] Subdomain available')

    // ==================== Step 3: Verify WorkMachine Configuration ====================
    // The form should show the WorkMachine Configuration section with radio options
    await expect(page.getByText('WorkMachine Configuration')).toBeVisible()

    // Select the first compute tier (should already be selected by default)
    const radioGroup = page.locator('[role="radiogroup"]')
    await expect(radioGroup).toBeVisible()
    console.log('[test] WorkMachine configuration visible')

    // ==================== Step 4: Verify cost summary ====================
    await expect(page.getByText('Estimated Cost')).toBeVisible()
    await expect(page.getByText('Estimated total')).toBeVisible()
    console.log('[test] Cost summary visible')

    // ==================== Step 5: Verify balance and submit ====================
    // The form shows "Current balance:" with the amount
    await expect(page.getByText('Current balance:')).toBeVisible()

    // The "Create Installation" button should be enabled (sufficient balance)
    const createButton = page.getByRole('button', {
      name: /create installation/i,
    })
    await expect(createButton).toBeVisible({ timeout: TIMEOUTS.action })
    await expect(createButton).toBeEnabled()
    console.log('[test] Create Installation button is active (sufficient balance)')

    // Intercept API call to capture installation ID
    const createResponsePromise = page.waitForResponse(
      (resp) =>
        resp.url().includes('/api/installations/create-installation') &&
        resp.status() === 200,
    )

    await createButton.click()
    console.log('[test] Clicked Create Installation')

    // Capture installation ID from creation response
    const createResponse = await createResponsePromise
    const createData = await createResponse.json()
    expect(createData.installationId).toBeTruthy()
    console.log(`[test] Installation created: ${createData.installationId}`)

    // ==================== Step 6: Verify redirect ====================
    // After creation, the form redirects through /api/installations/{id}/continue
    // which routes to the deploy page
    await page.waitForURL(
      (url) =>
        url.href.includes('/installations/new/kloudlite-cloud') ||
        url.href.includes('/installations'),
      { timeout: TIMEOUTS.navigation },
    )

    const finalUrl = page.url()
    console.log(`[test] Post-creation redirect: ${finalUrl}`)

    // Should NOT land on billing settings (no subscription required)
    expect(finalUrl).not.toContain('/installations/settings/billing')
    console.log('[test] Installation created successfully, correct redirect')
  })

  test('verify billing settings after installation creation', async ({
    page,
  }) => {
    test.setTimeout(30_000)

    await devLogin(page)
    await page.goto('/installations/settings/billing')
    await page.waitForLoadState('networkidle')

    // ==================== Step 1: Verify credit balance unchanged ====================
    // Credits are only deducted when resources actually run, not on installation creation
    await expect(page.getByText('Credit Balance')).toBeVisible({
      timeout: TIMEOUTS.action,
    })

    // Balance should still show a positive amount (the top-up amount)
    // We can't assert the exact amount but we can verify it's visible and positive
    const balanceElement = page.locator(
      '.text-green-600, .text-green-400',
    ).first()
    await expect(balanceElement).toBeVisible({ timeout: TIMEOUTS.action })
    console.log('[test] Credit balance still shows positive (not deducted on creation)')

    // ==================== Step 2: Verify transaction history ====================
    await expect(page.getByText('Transaction History')).toBeVisible()

    // Should show at least one top-up transaction from our earlier test
    const topupBadge = page.getByText('Top-up', { exact: true }).first()
    await expect(topupBadge).toBeVisible({ timeout: TIMEOUTS.action })
    console.log('[test] Top-up transaction visible in history')

    // ==================== Step 3: Verify billing management options ====================
    await expect(page.getByText('Billing Management')).toBeVisible()
    await expect(
      page.getByRole('button', { name: /manage payment methods/i }),
    ).toBeVisible()
    console.log('[test] Billing settings verified after installation creation')
  })
})
