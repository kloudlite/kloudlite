import { test, expect } from '@playwright/test'

const DEV_TOKEN = 'dev-superadmin'

test.describe('User Management', () => {
  test.use({ storageState: { cookies: [], origins: [] } })

  test.beforeEach(async ({ page }) => {
    await page.goto(`/superadmin-login?token=${DEV_TOKEN}`)
    await page.waitForURL('**/admin/**', { timeout: 15_000 })
    await page.getByRole('link', { name: 'User Management' }).click()
    await expect(page.getByRole('heading', { level: 1, name: 'User Management' })).toBeVisible()
  })

  // --- Add User Form Validation ---

  test('create button disabled when form is empty', async ({ page }) => {
    await page.getByRole('button', { name: 'Add User' }).click()

    await expect(page.getByRole('dialog')).toBeVisible()
    await expect(page.getByRole('button', { name: 'Create' })).toBeDisabled()
  })

  test('create button disabled without role selected', async ({ page }) => {
    await page.getByRole('button', { name: 'Add User' }).click()

    await page.getByPlaceholder('Enter email address').fill('noone@example.com')
    await expect(page.getByRole('button', { name: 'Create' })).toBeDisabled()
  })

  test('username auto-fills from email prefix', async ({ page }) => {
    await page.getByRole('button', { name: 'Add User' }).click()

    await page.getByPlaceholder('Enter email address').fill('john.doe@example.com')
    await expect(page.getByPlaceholder('Enter username (e.g., john-doe)')).toHaveValue('john-doe')
  })

  test('selecting a role enables create button when email is filled', async ({ page }) => {
    await page.getByRole('button', { name: 'Add User' }).click()

    await page.getByPlaceholder('Enter email address').fill('test@example.com')

    const dialog = page.getByRole('dialog')
    await dialog.getByRole('button', { name: 'User' }).click()

    await expect(page.getByRole('button', { name: 'Create' })).toBeEnabled()
  })

  // --- Row-dependent tests (require users in the table) ---

  test('edit user dialog shows disabled email and username when users exist', async ({ page }) => {
    const rows = page.getByRole('table').getByRole('row')
    const rowCount = await rows.count()

    if (rowCount > 1) {
      await rows.nth(1).getByRole('button').click()
      await page.getByRole('menuitem', { name: 'Edit User' }).click()

      const dialog = page.getByRole('dialog', { name: 'Edit User' })
      await expect(dialog).toBeVisible()
      await expect(dialog.getByPlaceholder('Enter email address')).toBeDisabled()
      await expect(dialog.getByPlaceholder('Enter username (e.g., john-doe)')).toBeDisabled()
      await expect(dialog.getByText('Email cannot be changed after user creation')).toBeVisible()
    }
  })

  test('reset password dialog validation when users exist', async ({ page }) => {
    const rows = page.getByRole('table').getByRole('row')
    const rowCount = await rows.count()

    if (rowCount > 1) {
      await rows.nth(1).getByRole('button').click()
      await page.getByRole('menuitem', { name: 'Reset Password' }).click()

      const dialog = page.getByRole('dialog', { name: 'Reset Password' })
      await expect(dialog).toBeVisible()
      await expect(dialog.getByRole('button', { name: 'Reset Password' })).toBeDisabled()

      // Short password shows validation error
      await page.getByPlaceholder('Enter new password (min 8 characters)').fill('short')
      await expect(page.getByText('Password must be at least 8 characters long', { exact: true })).toBeVisible()
      await expect(dialog.getByRole('button', { name: 'Reset Password' })).toBeDisabled()

      // Valid password enables button
      await page.getByPlaceholder('Enter new password (min 8 characters)').fill('validpass123')
      await expect(page.getByText('Password must be at least 8 characters long', { exact: true })).toBeHidden()
      await expect(dialog.getByRole('button', { name: 'Reset Password' })).toBeEnabled()
    }
  })

  test('delete user confirmation dialog when users exist', async ({ page }) => {
    const rows = page.getByRole('table').getByRole('row')
    const rowCount = await rows.count()

    if (rowCount > 1) {
      await rows.nth(1).getByRole('button').click()
      await page.getByRole('menuitem', { name: 'Delete User' }).click()

      const dialog = page.getByRole('dialog', { name: 'Delete User' })
      await expect(dialog).toBeVisible()
      await expect(dialog.getByText('Are you sure you want to delete user')).toBeVisible()
      await expect(dialog.getByText('This action cannot be undone')).toBeVisible()

      // Cancel preserves user
      await dialog.getByRole('button', { name: 'Cancel' }).click()
      await expect(page.getByRole('dialog')).toBeHidden()
    }
  })

  test('user row actions menu has all options when users exist', async ({ page }) => {
    const rows = page.getByRole('table').getByRole('row')
    const rowCount = await rows.count()

    if (rowCount > 1) {
      await rows.nth(1).getByRole('button').click()

      await expect(page.getByRole('menuitem', { name: 'Edit User' })).toBeVisible()
      await expect(page.getByRole('menuitem', { name: 'Reset Password' })).toBeVisible()
      await expect(page.getByRole('menuitem', { name: 'Send Invite' })).toBeVisible()
      await expect(page.getByRole('menuitem', { name: /Disable User|Enable User/ })).toBeVisible()
      await expect(page.getByRole('menuitem', { name: 'Delete User' })).toBeVisible()
    }
  })
})
