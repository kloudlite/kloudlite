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

    const secretInput = page.getByLabel('Client Secret')
    await expect(secretInput).toHaveAttribute('type', 'password')

    // Click the eye toggle button (the button adjacent to the secret input)
    const eyeButton = page.getByRole('dialog').locator('#client-secret + button')
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
