import { test, expect } from '@playwright/test'
import { devLogin } from '../../../lib/helpers'

test.describe('Console > Auth > Login', () => {
  test.use({ storageState: { cookies: [], origins: [] } })

  test.beforeEach(async ({ page }) => {
    await page.goto('/login')
  })

  test('shows welcome heading and subtitle', async ({ page }) => {
    await expect(page.getByRole('heading', { name: 'Welcome back' })).toBeVisible()
    await expect(page.getByText('Sign in to your installation console')).toBeVisible()
  })

  test('OAuth buttons are visible', async ({ page }) => {
    await expect(page.getByRole('button', { name: 'Continue with GitHub' })).toBeVisible()
    await expect(page.getByRole('button', { name: 'Continue with Google' })).toBeVisible()
    await expect(page.getByRole('button', { name: 'Continue with Microsoft' })).toBeVisible()
  })

  test('magic link form has email input and submit button', async ({ page }) => {
    await expect(page.getByText('Or continue with email')).toBeVisible()
    await expect(page.getByPlaceholder('you@example.com')).toBeVisible()
    await expect(page.getByRole('button', { name: 'Send magic link' })).toBeVisible()
  })

  test('dev login link is visible', async ({ page }) => {
    await expect(
      page.getByText('[Dev] Quick login as karthik@kloudlite.io'),
    ).toBeVisible()
  })

  test('feature highlights are visible', async ({ page }) => {
    await expect(page.getByText('No credit card')).toBeVisible()
    await expect(page.getByText('Free forever')).toBeVisible()
    await expect(page.getByText('Deploy in minutes')).toBeVisible()
    await expect(page.getByText('Enterprise security')).toBeVisible()
  })

  test('"New to Kloudlite?" section is visible', async ({ page }) => {
    await expect(page.getByText('New to Kloudlite?')).toBeVisible()
  })

  test('dev login redirects to installations', async ({ page }) => {
    await page.getByText('[Dev] Quick login as karthik@kloudlite.io').click()
    await page.waitForURL('**/installations', { timeout: 15_000 })
    await expect(page.getByRole('heading', { name: 'Installations' })).toBeVisible()
  })
})

test.describe('Console > Auth > Sign Out', () => {
  test.use({ storageState: { cookies: [], origins: [] } })

  test('sign out redirects to login', async ({ page }) => {
    await devLogin(page)

    // Open user dropdown
    await page.getByRole('button', { name: /Karthik/ }).click()

    // Click Sign Out
    await page.getByRole('menuitem', { name: 'Sign Out' }).click()

    // Should redirect to login page
    await page.waitForURL('**/login', { timeout: 15_000 })
    await expect(page.getByRole('heading', { name: 'Welcome back' })).toBeVisible()
  })
})
