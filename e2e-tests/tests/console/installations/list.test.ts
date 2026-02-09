import { consoleTest as test, expect } from '../../../lib/fixtures'

test.describe('Console > Installations > List', () => {
  test.use({ storageState: { cookies: [], origins: [] } })

  test('page heading and description visible', async ({ page }) => {
    await expect(page.getByRole('heading', { name: 'Installations' })).toBeVisible()
    await expect(page.getByText('Manage and monitor your cloud deployments')).toBeVisible()
  })

  test('new installation button present and links correctly', async ({ page }) => {
    const newButton = page.getByRole('link', { name: 'New Installation' })
    await expect(newButton).toBeVisible()
    await expect(newButton).toHaveAttribute('href', '/installations/new')
  })

  test('filter tabs are visible', async ({ page }) => {
    await expect(page.getByRole('button', { name: 'All', exact: true })).toBeVisible()
    await expect(page.getByRole('button', { name: 'Pending', exact: true })).toBeVisible()
    await expect(page.getByRole('button', { name: 'Installed', exact: true })).toBeVisible()
  })

  test('search input is present', async ({ page }) => {
    await expect(page.getByPlaceholder('Search installations...')).toBeVisible()
  })

  test('filter tabs can be clicked', async ({ page }) => {
    await page.getByRole('button', { name: 'Pending', exact: true }).click()
    await expect(page.getByRole('button', { name: 'Pending', exact: true })).toBeVisible()

    await page.getByRole('button', { name: 'Installed', exact: true }).click()
    await expect(page.getByRole('button', { name: 'Installed', exact: true })).toBeVisible()

    await page.getByRole('button', { name: 'All', exact: true }).click()
    await expect(page.getByRole('button', { name: 'All', exact: true })).toBeVisible()
  })

  test('empty state or table is shown', async ({ page }) => {
    // Either there are installations (table visible) or the empty state
    const table = page.getByRole('table')
    const emptyState = page.getByText('No installations')

    const hasTable = await table.isVisible().catch(() => false)
    const hasEmpty = await emptyState.isVisible().catch(() => false)

    expect(hasTable || hasEmpty).toBeTruthy()
  })

  test('table has correct columns when installations exist', async ({ page }) => {
    const table = page.getByRole('table')

    if (await table.isVisible().catch(() => false)) {
      await expect(table.getByText('Name')).toBeVisible()
      await expect(table.getByText('Status')).toBeVisible()
      await expect(table.getByText('Actions')).toBeVisible()
    }
  })

  test('header has user dropdown with account settings', async ({ page }) => {
    await page.getByRole('button', { name: /Karthik/ }).click()

    await expect(page.getByRole('menuitem', { name: 'Account Settings' })).toBeVisible()
    await expect(page.getByRole('menuitem', { name: 'Sign Out' })).toBeVisible()
  })
})
