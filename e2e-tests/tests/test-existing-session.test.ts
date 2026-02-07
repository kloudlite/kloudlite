import { test, expect } from '@playwright/test'

const DEV_TOKEN = 'dev-superadmin'

test.describe('Existing session → superadmin login', () => {
  test.use({ storageState: { cookies: [], origins: [] } })

  test('shows warning when user is already logged in', async ({ page }) => {
    // Step 1: Log in as regular user
    await page.goto('/auth/signin')
    await page.getByRole('textbox', { name: 'Email address' }).fill('karthik@kloudlite.io')
    await page.getByRole('textbox', { name: 'Password' }).fill('karthik001')
    await page.getByRole('button', { name: 'Sign in' }).click()

    // Wait for login to complete (credentials login does bcrypt + K8s API calls, needs more time under parallel load)
    await page.waitForURL(url => !url.pathname.includes('/auth/signin'), { timeout: 30_000 })

    // Step 2: Navigate to superadmin login while already logged in
    await page.goto(`/superadmin-login?token=${DEV_TOKEN}`)

    // Step 3: Should see the warning about active session
    await expect(page.getByText('Active Session Detected')).toBeVisible({ timeout: 10_000 })
    await expect(page.getByText('Continuing will end that session')).toBeVisible()
    await expect(page.getByRole('button', { name: 'Continue as Super Admin' })).toBeVisible()
    await expect(page.getByRole('link', { name: 'Cancel' })).toBeVisible()

    // Step 4: Click Continue as Super Admin (form submission → Server Action → signIn)
    await page.getByRole('button', { name: 'Continue as Super Admin' }).click()

    // Should redirect to admin dashboard
    await page.waitForURL('**/admin/**', { timeout: 15_000 })
    await expect(page.getByRole('banner').getByText('Admin', { exact: true })).toBeVisible()
  })
})
