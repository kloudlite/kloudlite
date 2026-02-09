import { test, expect } from '@playwright/test'

test.describe('Console > Settings > Theme', () => {
  test.use({ storageState: { cookies: [], origins: [] } })

  test.beforeEach(async ({ page }) => {
    await page.goto('/login')
  })

  test('switch to dark mode', async ({ page }) => {
    await page.getByRole('button', { name: 'Toggle theme' }).click()
    await page.getByRole('menuitem', { name: 'Dark' }).click()

    await expect(page.locator('html')).toHaveClass(/dark/)

    const cookies = await page.context().cookies()
    expect(cookies.find(c => c.name === 'theme')?.value).toBe('dark')
  })

  test('switch back to light mode', async ({ page }) => {
    await page.getByRole('button', { name: 'Toggle theme' }).click()
    await page.getByRole('menuitem', { name: 'Dark' }).click()
    await expect(page.locator('html')).toHaveClass(/dark/)

    await page.getByRole('button', { name: 'Toggle theme' }).click()
    await page.getByRole('menuitem', { name: 'Light' }).click()
    await expect(page.locator('html')).toHaveClass(/light/)
    await expect(page.locator('html')).not.toHaveClass(/dark/)
  })

  test('theme persists after reload', async ({ page }) => {
    await page.getByRole('button', { name: 'Toggle theme' }).click()
    await page.getByRole('menuitem', { name: 'Dark' }).click()

    await expect(page.locator('html')).toHaveClass(/dark/)

    await page.reload()

    await expect(page.locator('html')).toHaveClass(/dark/, { timeout: 10_000 })
  })

  test('system theme follows prefers-color-scheme', async ({ page }) => {
    await page.emulateMedia({ colorScheme: 'dark' })

    await page.getByRole('button', { name: 'Toggle theme' }).click()
    await page.getByRole('menuitem', { name: 'System' }).click()

    await expect(page.locator('html')).toHaveClass(/dark/)
  })
})
