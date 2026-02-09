import type { TestType, Response } from '@playwright/test'
import { expect } from '@playwright/test'
import { ChildProcess } from 'child_process'
import { mkdtempSync } from 'fs'
import { tmpdir } from 'os'
import { join } from 'path'
import { devLogin, runScript } from './helpers'
import { TIMEOUTS } from './constants'

export interface ProviderConfig {
  name: string
  tabName: string
  prerequisiteText: string
  regionLabel: string
  installUrlPath: string
  uninstallUrlPath: string
  subdomainPrefix: string
  testNamePrefix: string
  tmpDirPrefix: string
}

export function runProviderInstallationTest(
  test: TestType<any, any>,
  config: ProviderConfig,
) {
  test(`complete ${config.name} installation creation, deployment, and verification flow`, async ({ page }) => {
    test.setTimeout(TIMEOUTS.providerTest)

    const testId = Date.now().toString(36)
    const testName = `${config.testNamePrefix} ${testId}`
    const testSubdomain = `${config.subdomainPrefix}${testId}`
    let installationId: string | null = null
    let installationKey: string | null = null
    let installScriptProcess: ChildProcess | null = null
    let installScriptCompleted = false
    let uninstallCompleted = false

    const scriptWorkDir = mkdtempSync(join(tmpdir(), config.tmpDirPrefix))

    try {
      // ========== Step 1: Login and navigate to new installation ==========
      await devLogin(page)

      await page.getByRole('link', { name: 'New Installation' }).click()
      await page.waitForURL('**/installations/new')

      await expect(page.getByRole('heading', { name: 'Create Installation' })).toBeVisible()

      // ========== Step 2: Fill the create installation form ==========
      await page.getByPlaceholder('e.g., Production').fill(testName)
      await page.getByPlaceholder('Production deployment for our platform').fill(
        `Automated E2E ${config.name} test installation`,
      )

      await page.getByPlaceholder('your-company').fill(testSubdomain)
      await expect(page.getByText('Domain is available')).toBeVisible({ timeout: TIMEOUTS.action })

      await expect(page.getByText(new RegExp(`${testSubdomain}\\.`))).toBeVisible()

      const submitButton = page.getByRole('button', { name: 'Continue to Installation' })
      await expect(submitButton).toBeEnabled()

      // ========== Step 3: Submit form and capture installation ID & key ==========
      const createResponsePromise = page.waitForResponse(
        (resp: Response) =>
          resp.url().includes('/api/installations/create-installation') &&
          resp.status() === 200,
      )

      await submitButton.click()

      const createResponse = await createResponsePromise
      const createData = await createResponse.json()
      installationId = createData.installationId
      installationKey = createData.installationKey
      expect(createData.success).toBe(true)
      expect(installationKey).toBeTruthy()
      expect(installationId).toBeTruthy()

      // ========== Step 4: Verify install page and switch to provider tab ==========
      await page.waitForURL('**/installations/new/install', { timeout: TIMEOUTS.navigation })
      await expect(
        page.getByRole('heading', { name: 'Install Kloudlite in Your Cloud' }),
      ).toBeVisible({ timeout: TIMEOUTS.navigation })
      await expect(
        page.getByText('Run the installation command on your cloud provider'),
      ).toBeVisible()

      // Cloud provider tabs visible
      await expect(page.getByRole('tab', { name: 'AWS' })).toBeVisible()
      await expect(page.getByRole('tab', { name: 'GCP' })).toBeVisible()
      await expect(page.getByRole('tab', { name: 'Azure' })).toBeVisible()

      // Switch to the correct provider tab
      await page.getByRole('tab', { name: config.tabName }).click()

      // Verify provider-specific content
      await expect(page.getByText('Prerequisites:')).toBeVisible()
      await expect(page.getByText('Run this command:')).toBeVisible()
      await expect(page.getByText(config.prerequisiteText)).toBeVisible()
      await expect(page.getByText(config.regionLabel)).toBeVisible()

      // Verify the installation command
      const commandText = await page.locator('code').first().textContent()
      expect(commandText).toContain(`curl -fsSL https://get.khost.dev/${config.installUrlPath}`)
      expect(commandText).toContain(`--key ${installationKey}`)

      await expect(page.getByRole('button', { name: 'Copy' })).toBeVisible()
      await expect(page.getByText('Waiting for deployment...')).toBeVisible()

      // ========== Step 5: Execute the installation script ==========
      const installCommand = commandText!
      console.log(`\n[install-script] Working directory: ${scriptWorkDir}`)
      console.log(`[install-script] Executing: ${installCommand}\n`)

      const installScript = runScript(installCommand, scriptWorkDir, 'install-script')
      installScriptProcess = installScript.process

      // ========== Step 6: Wait for the UI to detect the deployment ==========
      await expect(
        page.getByText('Deployment verified. Configuring DNS...'),
      ).toBeVisible({ timeout: TIMEOUTS.deployment })

      console.log('\n[test] Deployment verified by UI — script has registered with console')

      // ========== Step 7: Wait for DNS configuration and redirect ==========
      await expect(
        page.getByText('Installation complete! Redirecting...'),
      ).toBeVisible({ timeout: TIMEOUTS.dns })

      console.log('[test] DNS configured — redirecting to complete page')

      // ========== Step 8: Verify complete page ==========
      await page.waitForURL('**/installations/new/complete', { timeout: 30_000 })

      await expect(
        page.getByRole('heading', { name: /Setting Up Your Installation|Installation Complete!/ }),
      ).toBeVisible({ timeout: 30_000 })

      console.log('[test] Complete page loaded — waiting for install script to finish...')

      const exitCode = await installScript.done
      installScriptCompleted = true
      console.log(`[install-script] Exited with code ${exitCode}`)
      expect(exitCode).toBe(0)

      // Debug: check current ping status
      const pingStatus = await page.evaluate(async (id: string) => {
        const resp = await fetch(`/api/installations/${id}/ping`)
        return resp.json()
      }, installationId)
      console.log(`[test] Ping status after script completed: ${JSON.stringify(pingStatus)}`)

      // ========== Step 9: Wait for installation to become active ==========
      await expect(
        page.getByRole('heading', { name: 'Installation Complete!' }),
      ).toBeVisible({ timeout: TIMEOUTS.deployment })

      console.log('[test] Installation is active!')

      await expect(
        page.getByText('Your Kloudlite installation is ready to use'),
      ).toBeVisible()

      await expect(page.getByText(new RegExp(`${testSubdomain}\\.`))).toBeVisible()

      await expect(
        page.getByRole('button', { name: 'Open Installation Settings' }),
      ).toBeVisible()
      await expect(
        page.getByRole('button', { name: 'View All Installations' }),
      ).toBeVisible()

      await expect(page.getByText('Installation is active and ready!')).toBeVisible()

      // ========== Step 10: Verify installation URL is reachable ==========
      const installationUrl = `https://${testSubdomain}.khost.dev`
      console.log(`[test] Verifying installation URL: ${installationUrl}`)

      const dashboardPage = await page.context().newPage()
      await dashboardPage.goto(installationUrl, {
        timeout: 60_000,
        waitUntil: 'domcontentloaded',
      })

      await expect(dashboardPage.locator('body')).not.toHaveText('', { timeout: 30_000 })
      const dashboardTitle = await dashboardPage.title()
      console.log(`[test] Dashboard page title: "${dashboardTitle}"`)
      expect(dashboardTitle).toBeTruthy()

      await dashboardPage.close()

      // ========== Step 11: Navigate to installation settings ==========
      await page.getByRole('button', { name: 'Open Installation Settings' }).click()
      await page.waitForURL(new RegExp(`/installations/${installationId}`), {
        timeout: TIMEOUTS.navigation,
      })

      await expect(
        page.getByRole('heading', { name: 'Installation Details' }),
      ).toBeVisible({ timeout: TIMEOUTS.action })

      await expect(page.getByText('ACTIVE', { exact: true })).toBeVisible()

      await expect(
        page.getByRole('link', { name: `${testSubdomain}.khost.dev` }).first(),
      ).toBeVisible()

      await expect(page.getByText(installationKey!).first()).toBeVisible()

      // ========== Step 12: Generate and verify superadmin login URL ==========
      await expect(
        page.getByRole('heading', { name: 'Super Admin Access' }),
      ).toBeVisible()

      await page.getByRole('button', { name: 'Generate Login URL' }).click()
      await expect(page.getByText(/Expires in \d+:\d+/)).toBeVisible({ timeout: TIMEOUTS.action })

      const loginUrlCode = page.locator('code').filter({ hasText: 'superadmin-login' })
      await expect(loginUrlCode).toBeVisible()
      const superadminLoginUrl = await loginUrlCode.textContent()
      expect(superadminLoginUrl).toContain(`${testSubdomain}.khost.dev/superadmin-login?token=`)
      console.log(`[test] Generated superadmin login URL: ${superadminLoginUrl}`)

      await expect(page.getByRole('link', { name: 'Open in New Tab' })).toBeVisible()
      await expect(
        page.getByRole('button', { name: 'Generate New URL' }),
      ).toBeVisible()

      const superadminPage = await page.context().newPage()
      const superadminResponse = await superadminPage.goto(superadminLoginUrl!, {
        timeout: 60_000,
        waitUntil: 'domcontentloaded',
      })

      const statusCode = superadminResponse?.status() ?? 0
      console.log(`[test] Superadmin login page status: ${statusCode}, URL: ${superadminPage.url()}`)
      expect(statusCode).toBeGreaterThanOrEqual(200)
      expect(statusCode).toBeLessThan(500)

      await superadminPage.close()

      // ========== Step 13: Verify Danger Zone and uninstall script ==========
      await expect(
        page.getByRole('heading', { name: 'Danger Zone' }),
      ).toBeVisible()

      await expect(page.getByText('Uninstall Script', { exact: true })).toBeVisible()
      await expect(
        page.getByText('Run this command in your terminal to uninstall Kloudlite'),
      ).toBeVisible()

      const uninstallCode = page
        .locator('code')
        .filter({ hasText: 'get.khost.dev/uninstall/' })
      await expect(uninstallCode).toBeVisible()
      const uninstallCommand = await uninstallCode.textContent()
      expect(uninstallCommand).toContain('curl -fsSL https://get.khost.dev/uninstall/')
      expect(uninstallCommand).toContain(`--key ${installationKey}`)
      console.log(`[test] Uninstall command: ${uninstallCommand}`)

      // ========== Step 14: Verify in installations list ==========
      await page.goto('/installations')
      await expect(page.getByText(testName)).toBeVisible({ timeout: TIMEOUTS.action })

      // ========== Step 15: Run uninstall script ==========
      console.log(`\n[uninstall-script] Working directory: ${scriptWorkDir}`)
      console.log(`[uninstall-script] Executing: ${uninstallCommand}\n`)

      const uninstallScript = runScript(uninstallCommand!, scriptWorkDir, 'uninstall-script')
      const uninstallExitCode = await uninstallScript.done

      console.log(`[uninstall-script] Exited with code ${uninstallExitCode}`)
      expect(uninstallExitCode).toBe(0)
      uninstallCompleted = true

      // ========== Step 16: Delete installation record ==========
      await page.evaluate(async (id: string) => {
        await fetch(`/api/installations/${id}/delete`, { method: 'DELETE' })
      }, installationId)

      console.log(`[cleanup] Deleted installation record ${installationId}`)

      await page.reload()
      await expect(page.getByText(testName)).not.toBeVisible({ timeout: TIMEOUTS.action })

      console.log(`[test] Full ${config.name} installation lifecycle complete!`)

      installationId = null
    } finally {
      // Kill install script if still running
      if (installScriptProcess && installScriptProcess.exitCode === null) {
        installScriptProcess.kill()
        console.log('[cleanup] Killed lingering install script process')
      }

      // Run uninstall if install completed but uninstall didn't
      if (installScriptCompleted && !uninstallCompleted && installationKey) {
        console.log('[cleanup] Install completed but uninstall did not — running uninstall...')
        try {
          const cleanupUninstallCmd = `curl -fsSL https://get.khost.dev/${config.uninstallUrlPath} | bash -s -- --key ${installationKey}`
          const cleanupScript = runScript(cleanupUninstallCmd, scriptWorkDir, 'cleanup-uninstall')
          const cleanupCode = await cleanupScript.done
          console.log(`[cleanup] Uninstall exited with code ${cleanupCode}`)
        } catch {
          console.log(`[cleanup] Uninstall script failed — ${config.name} resources may need manual cleanup`)
        }
      }

      // Delete installation record if test failed before cleanup
      if (installationId) {
        try {
          await page.evaluate(async (id: string) => {
            await fetch(`/api/installations/${id}/delete`, { method: 'DELETE' })
          }, installationId)
          console.log(`[cleanup] Deleted installation ${installationId}`)
        } catch {
          console.log(`[cleanup] Failed to delete installation ${installationId}`)
        }
      }
    }
  })
}
