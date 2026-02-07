import { test, expect } from '@playwright/test'

test.describe('CLI Download Page', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/install/kli')
  })

  test('page heading is visible', async ({ page }) => {
    await expect(page.getByRole('heading', { name: 'Install kli' })).toBeVisible()
    await expect(
      page.getByText('Kloudlite Installer CLI - Multi-cloud Kloudlite installation tool'),
    ).toBeVisible()
  })

  test('quick install section is visible', async ({ page }) => {
    await expect(page.getByRole('heading', { name: 'Quick Install' })).toBeVisible()
  })

  test('platform sections are visible', async ({ page }) => {
    await expect(page.getByRole('heading', { name: 'Linux (AMD64)' })).toBeVisible()
    await expect(page.getByRole('heading', { name: 'Linux (ARM64)' })).toBeVisible()
    await expect(page.getByRole('heading', { name: 'macOS (Intel)' })).toBeVisible()
    await expect(page.getByRole('heading', { name: 'macOS (Apple Silicon)' })).toBeVisible()
    await expect(page.getByRole('heading', { name: 'Windows (PowerShell)' })).toBeVisible()
  })

  test('direct downloads section has all platform links', async ({ page }) => {
    await expect(page.getByRole('heading', { name: 'Direct Downloads' })).toBeVisible()

    await expect(page.getByText('Linux AMD64', { exact: true })).toBeVisible()
    await expect(page.getByText('Linux ARM64', { exact: true })).toBeVisible()
    await expect(page.getByText('macOS Intel', { exact: true })).toBeVisible()
    await expect(page.getByText('macOS Apple Silicon', { exact: true })).toBeVisible()
    await expect(page.getByText('Windows AMD64', { exact: true })).toBeVisible()
    await expect(page.getByText('Windows ARM64', { exact: true })).toBeVisible()
  })

  test('quick start section is visible', async ({ page }) => {
    await expect(page.getByRole('heading', { name: 'Quick Start' })).toBeVisible()
    await expect(page.getByText('kli version')).toBeVisible()
  })

  test('specific version section is visible', async ({ page }) => {
    await expect(page.getByRole('heading', { name: 'Install Specific Version' })).toBeVisible()
  })

  test('all releases section links to GitHub', async ({ page }) => {
    await expect(page.getByRole('heading', { name: 'All Releases' })).toBeVisible()
    await expect(page.getByRole('link', { name: 'View Releases on GitHub' })).toBeVisible()
  })
})
