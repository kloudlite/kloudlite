import { test, expect, Page } from '@playwright/test'

async function devLogin(page: Page) {
  await page.goto('/api/dev-login')
  await page.waitForURL('**/installations', { timeout: 15_000 })
}

function escapeRegex(str: string) {
  return str.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')
}

test.describe('Installation Detail', () => {
  test.use({ storageState: { cookies: [], origins: [] } })

  test.beforeEach(async ({ page }) => {
    await devLogin(page)
  })

  test('clicking an installation navigates to detail page', async ({ page }) => {
    const table = page.getByRole('table')

    if (await table.isVisible().catch(() => false)) {
      // Get the first installation link in the table body
      const firstLink = table.locator('tbody').getByRole('link').first()
      if (await firstLink.isVisible().catch(() => false)) {
        const installationName = (await firstLink.textContent())?.trim() || ''
        await firstLink.click()

        await expect(page).toHaveURL(/\/installations\/[^/]+/, { timeout: 15_000 })
        await expect(
          page.getByRole('heading', { name: new RegExp(`Installation: ${escapeRegex(installationName)}`) }),
        ).toBeVisible()
      }
    }
  })

  test('detail page has overview and team tabs', async ({ page }) => {
    const table = page.getByRole('table')

    if (await table.isVisible().catch(() => false)) {
      const firstLink = table.locator('tbody').getByRole('link').first()
      if (await firstLink.isVisible().catch(() => false)) {
        await firstLink.click()
        await expect(page).toHaveURL(/\/installations\/[^/]+/, { timeout: 15_000 })

        await expect(page.getByRole('link', { name: 'Overview' })).toBeVisible()
        await expect(page.getByRole('link', { name: 'Team' })).toBeVisible()
      }
    }
  })

  test('detail page has back to installations link', async ({ page }) => {
    const table = page.getByRole('table')

    if (await table.isVisible().catch(() => false)) {
      const firstLink = table.locator('tbody').getByRole('link').first()
      if (await firstLink.isVisible().catch(() => false)) {
        await firstLink.click()
        await expect(page).toHaveURL(/\/installations\/[^/]+/, { timeout: 15_000 })

        await expect(page.getByRole('link', { name: 'Back to Installations' })).toBeVisible()
      }
    }
  })

  test('team tab shows team members heading', async ({ page }) => {
    const table = page.getByRole('table')

    if (await table.isVisible().catch(() => false)) {
      const firstLink = table.locator('tbody').getByRole('link').first()
      if (await firstLink.isVisible().catch(() => false)) {
        await firstLink.click()
        await expect(page).toHaveURL(/\/installations\/[^/]+/, { timeout: 15_000 })

        await page.getByRole('link', { name: 'Team' }).click()
        await expect(page).toHaveURL(/\/installations\/[^/]+\/team/)

        await expect(page.getByText('Team Members')).toBeVisible()
        await expect(page.getByText('Manage who has access to this installation')).toBeVisible()
      }
    }
  })

  test('team tab shows invite member button for owner', async ({ page }) => {
    const table = page.getByRole('table')

    if (await table.isVisible().catch(() => false)) {
      const firstLink = table.locator('tbody').getByRole('link').first()
      if (await firstLink.isVisible().catch(() => false)) {
        await firstLink.click()
        await expect(page).toHaveURL(/\/installations\/[^/]+/, { timeout: 15_000 })

        await page.getByRole('link', { name: 'Team' }).click()
        await expect(page).toHaveURL(/\/installations\/[^/]+\/team/)

        // Owner should see the invite button
        await expect(page.getByRole('button', { name: 'Invite Member' })).toBeVisible()
      }
    }
  })
})
