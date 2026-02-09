import { consoleTest as test, expect } from '../../../lib/fixtures'
import { TIMEOUTS } from '../../../lib/constants'

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

test.describe('Console > Installations > Continue Button', () => {
  test.use({ storageState: { cookies: [], origins: [] } })

  let testInstallationId: string | null = null

  test.afterEach(async ({ page }) => {
    // Clean up the test installation if it was created
    if (testInstallationId) {
      try {
        await page.evaluate(async (id: string) => {
          await fetch(`/api/installations/${id}/delete`, { method: 'DELETE' })
        }, testInstallationId)
      } catch {
        // Ignore cleanup errors
      }
      testInstallationId = null
    }
  })

  test('continue button appears for pending installations and navigates to install step', async ({ page }) => {
    // Step 1: Create a pending installation via the API
    const testId = Date.now().toString(36)
    const testName = `E2E Continue ${testId}`
    const testSubdomain = `e2e-cont-${testId}`

    const createResponse = await page.evaluate(
      async ({ name, subdomain }: { name: string; subdomain: string }) => {
        const resp = await fetch('/api/installations/create-installation', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ name, description: 'Test for continue button', subdomain }),
        })
        return resp.json()
      },
      { name: testName, subdomain: testSubdomain },
    )

    expect(createResponse.success).toBe(true)
    expect(createResponse.installationId).toBeTruthy()
    testInstallationId = createResponse.installationId

    // Step 2: Reload the installations list to see the new installation
    await page.goto('/installations')
    await expect(page.getByRole('heading', { name: 'Installations' })).toBeVisible()

    // Step 3: Verify the test installation appears in the list
    await expect(page.getByText(testName)).toBeVisible({ timeout: TIMEOUTS.action })

    // Step 4: Verify the Continue button is visible in the same row
    const row = page.getByRole('row').filter({ hasText: testName })
    const continueButton = row.getByRole('button', { name: 'Continue' })
    await expect(continueButton).toBeVisible()

    // Step 5: Verify the installation shows a pending status (NOT INSTALLED since no secretKey)
    await expect(row.getByText('NOT INSTALLED')).toBeVisible()

    // Step 6: Click Continue and verify it navigates to the install step
    await continueButton.click()
    await page.waitForURL('**/installations/new/install', { timeout: TIMEOUTS.navigation })

    // Step 7: Verify we landed on the install page (not redirected back to /installations/new)
    await expect(
      page.getByRole('heading', { name: 'Install Kloudlite in Your Cloud' }),
    ).toBeVisible({ timeout: TIMEOUTS.navigation })
  })

  test('continue button does not appear for active installations', async ({ page }) => {
    // Check if any active installations exist in the list
    const table = page.getByRole('table')
    if (!(await table.isVisible().catch(() => false))) return

    const activeRows = page.getByRole('row').filter({ hasText: 'ACTIVE' })
    const activeCount = await activeRows.count()

    if (activeCount > 0) {
      // Active installations should NOT have a Continue button
      const firstActiveRow = activeRows.first()
      await expect(firstActiveRow.getByRole('button', { name: 'Continue' })).not.toBeVisible()
    }
  })

  test('pending filter shows only installations with continue button', async ({ page }) => {
    // Switch to Pending filter
    await page.getByRole('button', { name: 'Pending', exact: true }).click()

    const table = page.getByRole('table')
    if (!(await table.isVisible().catch(() => false))) return

    // All visible rows (excluding header) should have a Continue button
    const dataRows = page.locator('tbody tr')
    const rowCount = await dataRows.count()

    for (let i = 0; i < rowCount; i++) {
      await expect(dataRows.nth(i).getByRole('button', { name: 'Continue' })).toBeVisible()
    }
  })
})
