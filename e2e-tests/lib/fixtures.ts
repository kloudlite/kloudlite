import { test as base, expect } from '@playwright/test'
import { devLogin, superadminLogin } from './helpers'

export const dashboardTest = base.extend<{}>({
  page: async ({ page }, use) => {
    await superadminLogin(page)
    await use(page)
  },
})

export const consoleTest = base.extend<{}>({
  page: async ({ page }, use) => {
    await devLogin(page)
    await use(page)
  },
})

export { expect }
