import { test, expect } from '@playwright/test'

const DEV_TOKEN = 'dev-superadmin'

test.describe('OAuth Providers', () => {
  test.use({ storageState: { cookies: [], origins: [] } })

  test.beforeEach(async ({ page }) => {
    await page.goto(`/superadmin-login?token=${DEV_TOKEN}`)
    await page.waitForURL('**/admin/**', { timeout: 15_000 })
    await page.getByRole('link', { name: 'OAuth Providers' }).click()
    await expect(page.getByRole('heading', { name: 'OAuth Provider Configuration' })).toBeVisible()
  })

  // --- Smoke ---

  test('all three providers listed with toggles', async ({ page }) => {
    await expect(page.getByRole('heading', { name: 'Google' })).toBeVisible()
    await expect(page.getByRole('heading', { name: 'GitHub' })).toBeVisible()
    await expect(page.getByRole('heading', { name: 'Microsoft' })).toBeVisible()

    expect(await page.getByRole('switch').count()).toBe(3)
  })

  // --- Provider-specific Business Logic ---

  test('Microsoft dialog shows tenant ID field', async ({ page }) => {
    const msCard = page.locator('.rounded-lg.border').filter({ hasText: 'Microsoft' })
    await msCard.getByRole('button', { name: /Set up credentials|Edit credentials/ }).click()

    const dialog = page.getByRole('dialog')
    await expect(dialog.getByText('Configure Microsoft OAuth')).toBeVisible()
    await expect(dialog.getByLabel('Client ID')).toBeVisible()
    await expect(dialog.getByLabel('Client Secret')).toBeVisible()
    await expect(dialog.getByLabel('Tenant ID')).toBeVisible()
    await expect(dialog.getByPlaceholder("Enter tenant ID (or leave empty for 'common')")).toBeVisible()
  })

  test('client secret field toggles visibility', async ({ page }) => {
    const googleCard = page.locator('.rounded-lg.border').filter({ hasText: 'Google' })
    await googleCard.getByRole('button', { name: /Set up credentials|Edit credentials/ }).click()

    const dialog = page.getByRole('dialog')
    const secretInput = dialog.getByLabel('Client Secret')
    await expect(secretInput).toHaveAttribute('type', 'password')

    // Click the eye toggle button (sibling of the secret input inside its wrapper)
    const eyeButton = secretInput.locator('..').getByRole('button')
    await eyeButton.click()
    await expect(secretInput).toHaveAttribute('type', 'text')

    // Click again to hide
    await eyeButton.click()
    await expect(secretInput).toHaveAttribute('type', 'password')
  })

  // --- Toggle State ---

  test('unconfigured provider toggle is disabled', async ({ page }) => {
    const unconfiguredCard = page.locator('.rounded-lg.border').filter({ hasText: 'Not configured' }).first()

    if (await unconfiguredCard.isVisible()) {
      await expect(unconfiguredCard.getByRole('switch')).toBeDisabled()
    }
  })
})

// --- Login screen integration ---

const OAUTH_PROVIDERS = [
  { name: 'Google', buttonText: 'Continue with Google' },
  { name: 'GitHub', buttonText: 'Continue with GitHub' },
  { name: 'Microsoft', buttonText: 'Continue with Microsoft' },
] as const

test.describe('OAuth providers on login screen', () => {
  test.use({ storageState: { cookies: [], origins: [] } })
  test.describe.configure({ timeout: 60_000 })

  test('all configured OAuth providers are visible on signin page', async ({ page }) => {
    // Login as superadmin and go to OAuth Providers page
    await page.goto(`/superadmin-login?token=${DEV_TOKEN}`)
    await page.waitForURL('**/admin/**', { timeout: 15_000 })
    await page.getByRole('link', { name: 'OAuth Providers' }).click()
    await expect(page.getByRole('heading', { name: 'OAuth Provider Configuration' })).toBeVisible()

    // Record which providers were originally disabled, and enable all configured ones
    const originallyDisabled: string[] = []

    for (const { name } of OAUTH_PROVIDERS) {
      const card = page.locator('.rounded-lg.border').filter({ hasText: name })
      const badge = card.getByText('Configured')
      const toggle = card.getByRole('switch')

      // Only enable providers that have credentials configured
      if (await badge.isVisible()) {
        if (!(await toggle.isChecked())) {
          originallyDisabled.push(name)
          await toggle.click()
          await expect(toggle).toBeChecked({ timeout: 10_000 })
        }
      }
    }

    // Sign out to reach the signin page
    await page.getByRole('button', { name: 'Super Admin (Dev)' }).click()
    await page.getByRole('menuitem', { name: 'Sign out' }).click()
    await page.waitForURL('**/auth/signin**', { timeout: 15_000 })

    // Verify all three OAuth provider buttons are visible
    for (const { buttonText } of OAUTH_PROVIDERS) {
      await expect(
        page.getByRole('button', { name: buttonText }),
      ).toBeVisible()
    }

    // Verify the "Or continue with email" divider and credentials form
    await expect(page.getByText('Or continue with email')).toBeVisible()
    await expect(page.getByRole('textbox', { name: 'Email address' })).toBeVisible()
    await expect(page.getByRole('button', { name: 'Sign in' })).toBeVisible()

    // Restore original state — disable providers that were not originally enabled
    if (originallyDisabled.length > 0) {
      await page.goto(`/superadmin-login?token=${DEV_TOKEN}`)
      await page.waitForURL('**/admin/**', { timeout: 15_000 })
      await page.getByRole('link', { name: 'OAuth Providers' }).click()
      await expect(page.getByRole('heading', { name: 'OAuth Provider Configuration' })).toBeVisible()

      for (const name of originallyDisabled) {
        const card = page.locator('.rounded-lg.border').filter({ hasText: name })
        const toggle = card.getByRole('switch')
        if (await toggle.isChecked()) {
          await toggle.click()
          await expect(toggle).not.toBeChecked({ timeout: 10_000 })
        }
      }
    }
  })
})
