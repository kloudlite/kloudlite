import { test as setup } from '@playwright/test'

// Warm up: hit key pages to trigger compilation, then log in via API route.
setup('warm up dev server', async ({ page }) => {
  await page.goto('/auth/signin')
  await page.goto('/api/superadmin-login?token=dev-superadmin')
  await page.waitForURL('**/admin/**', { timeout: 15_000 })
})
