import { test, expect } from '@playwright/test'

const DEV_TOKEN = 'dev-superadmin'

test.describe('Machine Configurations', () => {
  test.use({ storageState: { cookies: [], origins: [] } })

  test.beforeEach(async ({ page }) => {
    await page.goto(`/superadmin-login?token=${DEV_TOKEN}`)
    await page.waitForURL('**/admin/**', { timeout: 15_000 })
    await page.getByRole('link', { name: 'Machine Configs' }).click()
    await expect(page.getByRole('heading', { level: 1, name: 'Machine Configurations' })).toBeVisible()
  })

  // --- Form Validation & Behavior ---

  test('machine type ID has correct validation pattern', async ({ page }) => {
    await page.getByRole('button', { name: 'Add Configuration' }).first().click()

    const idInput = page.getByLabel('Machine Type ID')
    await expect(idInput).toHaveAttribute('pattern', '^[a-z0-9-]+$')
    await expect(idInput).toHaveAttribute('required', '')
  })

  test('cpu and memory are required, gpu is optional', async ({ page }) => {
    await page.getByRole('button', { name: 'Add Configuration' }).first().click()

    await expect(page.getByLabel('CPU (vCPU)')).toHaveAttribute('required', '')
    await expect(page.getByLabel('Memory (GB)')).toHaveAttribute('required', '')
    await expect(page.getByLabel('GPU (optional)')).not.toHaveAttribute('required', '')
  })

  test('category dropdown has all options', async ({ page }) => {
    await page.getByRole('button', { name: 'Add Configuration' }).first().click()

    await page.getByRole('combobox').click()

    await expect(page.getByRole('option', { name: 'General Purpose' })).toBeVisible()
    await expect(page.getByRole('option', { name: 'Compute Optimized' })).toBeVisible()
    await expect(page.getByRole('option', { name: 'Memory Optimized' })).toBeVisible()
    await expect(page.getByRole('option', { name: 'GPU Accelerated' })).toBeVisible()
    await expect(page.getByRole('option', { name: 'Development' })).toBeVisible()
  })

  test('active toggle defaults to checked for new config', async ({ page }) => {
    await page.getByRole('button', { name: 'Add Configuration' }).first().click()

    await expect(page.getByLabel('Active (available for use)')).toBeChecked()
  })

  test('active toggle can be unchecked', async ({ page }) => {
    await page.getByRole('button', { name: 'Add Configuration' }).first().click()

    const toggle = page.getByLabel('Active (available for use)')
    await expect(toggle).toBeChecked()
    await toggle.click()
    await expect(toggle).not.toBeChecked()
  })

  test('category can be changed', async ({ page }) => {
    await page.getByRole('button', { name: 'Add Configuration' }).first().click()

    await page.getByRole('combobox').click()
    await page.getByRole('option', { name: 'Compute Optimized' }).click()

    // Combobox should now show the selected value
    await expect(page.getByRole('combobox')).toHaveText('Compute Optimized')
  })

  test('form fields accept input values', async ({ page }) => {
    await page.getByRole('button', { name: 'Add Configuration' }).first().click()

    await page.getByLabel('Machine Type ID').fill('test-machine')
    await page.getByLabel('Display Name').fill('Test Machine')
    await page.getByLabel('Description').fill('A test description')
    await page.getByLabel('CPU (vCPU)').fill('4')
    await page.getByLabel('Memory (GB)').fill('8')
    await page.getByLabel('GPU (optional)').fill('1')

    await expect(page.getByLabel('Machine Type ID')).toHaveValue('test-machine')
    await expect(page.getByLabel('Display Name')).toHaveValue('Test Machine')
    await expect(page.getByLabel('Description')).toHaveValue('A test description')
    await expect(page.getByLabel('CPU (vCPU)')).toHaveValue('4')
    await expect(page.getByLabel('Memory (GB)')).toHaveValue('8')
    await expect(page.getByLabel('GPU (optional)')).toHaveValue('1')
  })
})
