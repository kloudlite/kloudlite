import { test, expect } from '@playwright/test'
import { devLogin } from '../../../lib/helpers'
import { TIMEOUTS, STRIPE_TEST_CARD } from '../../../lib/constants'

test.use({ storageState: { cookies: [], origins: [] } })

/**
 * End-to-end test for the KL Cloud billing flow:
 * Login → Create Installation → Stripe Checkout → Post-payment redirect → Deploy page → Billing Settings
 *
 * This test validates the critical post-checkout redirect fix where the success_url
 * now routes through /api/installations/{id}/continue instead of going to billing settings.
 */
test.describe.serial('KL Cloud Billing Flow', () => {
  const testId = Date.now().toString(36)
  const installationName = `E2E Test ${testId}`
  const subdomain = `e2e-${testId}`
  let installationId: string | null = null

  test('create KL Cloud installation and complete Stripe checkout', async ({
    page,
  }) => {
    test.setTimeout(120_000)

    // ==================== Step 1: Login ====================
    await devLogin(page)
    await expect(
      page.getByRole('heading', { name: 'Installations' }),
    ).toBeVisible()
    console.log('[test] Logged in successfully')

    // ==================== Step 2: Navigate to KL Cloud form ====================
    await page.getByRole('button', { name: 'New Installation' }).click()
    await page.getByText('Kloudlite Cloud').click()
    await page.waitForURL('**/installations/new-kl-cloud')

    await expect(
      page.getByRole('heading', { name: 'Create Installation' }),
    ).toBeVisible()
    console.log('[test] On KL Cloud form')

    // ==================== Step 3: Fill installation details ====================
    await page.getByPlaceholder('e.g., Production').fill(installationName)
    await page.getByRole('textbox', { name: 'your-company' }).fill(subdomain)

    // Wait for subdomain availability check
    await expect(page.getByText('Domain is available')).toBeVisible({
      timeout: TIMEOUTS.action,
    })
    console.log('[test] Subdomain available')

    // ==================== Step 4: Add 1 user (Tier 1) ====================
    await page
      .getByRole('button', { name: /increase users for tier 1/i })
      .click()

    // Verify summary updates (monthly total shows user count)
    await expect(page.getByText(/Monthly total \(1 user\)/)).toBeVisible()
    console.log('[test] Added 1 Tier 1 user')

    // ==================== Step 5: Submit form ====================
    const submitButton = page.getByRole('button', {
      name: /create & subscribe/i,
    })
    await expect(submitButton).toBeEnabled()

    // Intercept API calls
    const createResponsePromise = page.waitForResponse(
      (resp) =>
        resp.url().includes('/api/installations/create-installation') &&
        resp.status() === 200,
    )

    await submitButton.click()

    // Capture installation ID from creation response
    const createResponse = await createResponsePromise
    const createData = await createResponse.json()
    installationId = createData.installationId
    expect(installationId).toBeTruthy()
    console.log(`[test] Installation created: ${installationId}`)

    // ==================== Step 6: Stripe Checkout ====================
    await page.waitForURL(/checkout\.stripe\.com/, {
      timeout: TIMEOUTS.stripeCheckout,
    })
    console.log(`[test] On Stripe Checkout: ${page.url()}`)

    // Wait for Stripe checkout to render (don't use networkidle — Stripe keeps connections open)
    await page.waitForLoadState('domcontentloaded')

    // Stripe Checkout shows different UIs depending on customer state:
    // - New customer: may show "Pay without Link" button
    // - Returning customer: shows Link/Amazon buttons at top, then "OR" separator,
    //   then payment method radio buttons (Card, Cash App Pay, Bank)
    //
    // We need to get to the Card form. Wait for the page to stabilize.
    await page.waitForLoadState('domcontentloaded')

    // Wait for either "Pay without Link" button or the payment method section
    const payWithoutLink = page.getByRole('button', {
      name: 'Pay without Link',
    })
    const paymentMethodHeading = page.getByText('Payment method')

    await expect(
      payWithoutLink.or(paymentMethodHeading).first(),
    ).toBeVisible({ timeout: 20_000 })

    if (await payWithoutLink.isVisible().catch(() => false)) {
      await payWithoutLink.click()
      console.log('[test] Clicked "Pay without Link"')
      // Wait for payment methods to appear
      await expect(paymentMethodHeading).toBeVisible({ timeout: 15_000 })
    }
    console.log('[test] Payment method section visible')

    // Now we need to select Card and expand its form.
    // Stripe uses radio buttons or accordion items for payment methods.
    const cardNumberField = page.getByPlaceholder('1234 1234 1234 1234')
    if (
      !(await cardNumberField.isVisible({ timeout: 3_000 }).catch(() => false))
    ) {
      // Click the Card radio/accordion item using force: true to bypass overlay interception
      const cardAccordionButton = page.locator(
        'button[data-testid="card-accordion-item-button"]',
      )
      if (
        await cardAccordionButton
          .isVisible({ timeout: 2_000 })
          .catch(() => false)
      ) {
        await cardAccordionButton.click({ force: true })
        console.log('[test] Clicked card accordion button (force)')
      } else {
        // Fallback: click the radio input for card
        const cardRadio = page.locator(
          'input[type="radio"][value="card"], input[name="paymentMethod"][value="card"]',
        )
        if (
          await cardRadio.isVisible({ timeout: 1_000 }).catch(() => false)
        ) {
          await cardRadio.click({ force: true })
          console.log('[test] Clicked card radio input')
        } else {
          // Last resort: click any element containing "Card" text near payment method
          await page.getByText('Card', { exact: true }).first().click({
            force: true,
          })
          console.log('[test] Clicked Card text (force)')
        }
      }
    }

    // Wait for card number field to appear
    await expect(cardNumberField).toBeVisible({ timeout: 15_000 })
    console.log('[test] Card form visible')

    // ==================== Step 7: Fill card details ====================
    // Stripe Checkout has card fields directly on the page (not in iframes).
    // Use pressSequentially for card number — Stripe fields may need keystrokes.
    await cardNumberField.click()
    await cardNumberField.pressSequentially(STRIPE_TEST_CARD.number, {
      delay: 50,
    })
    console.log('[test] Filled card number')

    const expiryField = page.getByPlaceholder('MM / YY')
    await expiryField.click()
    await expiryField.pressSequentially(STRIPE_TEST_CARD.expiry, { delay: 50 })
    console.log('[test] Filled expiry')

    const cvcField = page.getByPlaceholder('CVC').or(page.getByRole('textbox', { name: 'CVC' })).first()
    await cvcField.click()
    await cvcField.pressSequentially(STRIPE_TEST_CARD.cvc, { delay: 50 })
    console.log('[test] Filled CVC')

    // Cardholder name
    const nameField = page.getByPlaceholder('Full name on card')
    if (await nameField.isVisible({ timeout: 2_000 }).catch(() => false)) {
      await nameField.fill(STRIPE_TEST_CARD.name)
      console.log('[test] Filled cardholder name')
    }

    // Country/region (should default to something, leave as-is)

    // ==================== Step 8: Submit payment ====================
    const subscribeButton = page.getByRole('button', { name: 'Subscribe' })
    await expect(subscribeButton).toBeVisible()
    await subscribeButton.click()
    console.log('[test] Clicked Subscribe, waiting for redirect...')

    // ==================== Step 9: Verify post-checkout redirect ====================
    // After payment, Stripe redirects to our success_url which is:
    // /api/installations/{id}/continue
    // That route checks subscription status and redirects to the deploy page.
    await page.waitForURL(
      (url) => !url.href.includes('checkout.stripe.com'),
      { timeout: TIMEOUTS.stripeCheckout },
    )

    const finalUrl = page.url()
    console.log(`[test] Post-checkout redirect: ${finalUrl}`)

    // CRITICAL ASSERTION: Should NOT land on billing settings
    expect(finalUrl).not.toContain('/installations/settings/billing')

    // Should land on the deploy page or the installation form (if subscription needs
    // to be verified) or installations list. The /continue route handles state-based routing.
    const validDestinations = [
      '/installations/new/kloudlite-cloud', // Deploy page (subscription confirmed)
      '/installations/new-kl-cloud', // Form page (if subscription not yet synced)
      '/installations', // List (if deployment already done)
    ]
    const landedOnValid = validDestinations.some((dest) =>
      finalUrl.includes(dest),
    )
    expect(landedOnValid).toBe(true)
    console.log('[test] Post-checkout redirect PASSED — correct destination')
  })

  test('billing settings shows active subscription with line items', async ({
    page,
  }) => {
    test.setTimeout(30_000)

    await devLogin(page)
    await page.goto('/installations/settings/billing')
    await page.waitForLoadState('networkidle')

    // Billing page heading
    await expect(page.getByRole('heading', { name: /billing/i })).toBeVisible()

    // Subscription should be active (synced from Stripe directly or via webhook)
    await expect(page.getByText('Active', { exact: true })).toBeVisible({
      timeout: TIMEOUTS.action,
    })
    console.log('[test] Subscription shows Active status')

    // Should show the Control Plane line item
    await expect(page.getByText(/control plane/i)).toBeVisible({
      timeout: TIMEOUTS.action,
    })
    console.log('[test] Control Plane item visible')

    // Should show "Manage Billing" button (for opening Stripe portal)
    await expect(page.getByText(/manage billing/i)).toBeVisible()
    console.log('[test] Billing settings page verified')
  })

  test('installation appears in list with Continue button', async ({
    page,
  }) => {
    test.setTimeout(30_000)

    await devLogin(page)
    await page.goto('/installations')

    // Installation should be visible — scope to the table row
    const installationRow = page
      .locator('tr')
      .filter({ hasText: installationName })
    await expect(installationRow).toBeVisible({
      timeout: TIMEOUTS.action,
    })

    // Should have a Continue button within the row (not yet deployed)
    const continueButton = installationRow.getByRole('button', {
      name: 'Continue',
    })
    await expect(continueButton).toBeVisible()
    console.log('[test] Installation in list with Continue button')
  })

  test('continue button navigates to deploy page', async ({ page }) => {
    test.setTimeout(30_000)

    await devLogin(page)
    await page.goto('/installations')

    // Scope to the row for our test installation
    const installationRow = page
      .locator('tr')
      .filter({ hasText: installationName })
    await expect(installationRow).toBeVisible({ timeout: TIMEOUTS.action })

    // Click Continue — should go through /api/installations/{id}/continue
    // which verifies the active subscription and routes to deploy page
    const continueButton = installationRow.getByRole('button', {
      name: 'Continue',
    })
    await continueButton.click()

    // Should land on the deploy page (subscription is active)
    await page.waitForURL(/\/installations\/new\/kloudlite-cloud/, {
      timeout: TIMEOUTS.navigation,
    })

    await expect(
      page.getByRole('heading', { name: /deploying kloudlite cloud/i }),
    ).toBeVisible()
    console.log('[test] Continue button routes to deploy page correctly')
  })

  test('cleanup: delete test installation and cancel subscription', async ({
    page,
  }) => {
    test.setTimeout(30_000)

    if (!installationId) {
      console.log('[cleanup] No installation to clean up')
      return
    }

    await devLogin(page)

    // Delete the test installation
    const deleteResult = await page.evaluate(async (id: string) => {
      const resp = await fetch(`/api/installations/${id}/delete`, {
        method: 'DELETE',
      })
      return { status: resp.status, ok: resp.ok }
    }, installationId)
    console.log(
      `[cleanup] Delete installation: ${deleteResult.status} (${deleteResult.ok ? 'ok' : 'failed'})`,
    )

    // Verify it's gone
    await page.goto('/installations')
    await expect(page.getByText(installationName)).not.toBeVisible({
      timeout: TIMEOUTS.action,
    })
    console.log('[cleanup] Installation removed from list')

    // Note: Stripe subscription is NOT cancelled here on purpose.
    // The org still has an active subscription which is correct — billing is org-level.
    // The subscription would be cancelled via the billing settings "Cancel Subscription" button
    // in a real workflow.
  })
})
