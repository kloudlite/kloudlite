import { test as setup } from '@playwright/test'

// Warm up: hit login page then dev-login to trigger compilation
setup('warm up console dev server', async ({ page }) => {
  await page.goto('/login')
  await page.goto('/api/dev-login')
  await page.waitForURL('**/installations', { timeout: 15_000 })
})
