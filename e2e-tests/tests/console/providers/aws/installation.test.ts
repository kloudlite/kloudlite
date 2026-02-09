import { consoleTest, expect } from '../../../../lib/fixtures'
import { runProviderInstallationTest } from '../../../../lib/provider-test'

consoleTest.describe('Console > Providers > AWS > Installation', () => {
  consoleTest.use({ storageState: { cookies: [], origins: [] } })

  consoleTest('form validation prevents submission with invalid data', async ({ page }) => {
    await page.goto('/installations/new')

    await expect(page.getByRole('heading', { name: 'Create Installation' })).toBeVisible()

    const submitButton = page.getByRole('button', { name: 'Continue to Installation' })

    // Submit button disabled when form is empty (subdomain not verified)
    await expect(submitButton).toBeDisabled()

    // Fill only name — button still disabled (no valid subdomain)
    await page.getByPlaceholder('e.g., Production').fill('Test Installation')
    await expect(submitButton).toBeDisabled()

    // Type invalid subdomain (uppercase not allowed)
    await page.getByPlaceholder('your-company').fill('INVALID')
    await expect(submitButton).toBeDisabled()

    // Clear and type valid but short subdomain
    await page.getByPlaceholder('your-company').fill('ab')
    await expect(submitButton).toBeDisabled()
  })

  runProviderInstallationTest(consoleTest, {
    name: 'AWS',
    tabName: 'AWS',
    prerequisiteText: 'AWS CLI configured',
    regionLabel: 'Select AWS Region:',
    installUrlPath: 'install/aws',
    uninstallUrlPath: 'uninstall/aws',
    subdomainPrefix: 'e2e-',
    testNamePrefix: 'E2E Test',
    tmpDirPrefix: 'kl-e2e-',
  })
})
