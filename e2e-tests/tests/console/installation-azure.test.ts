import { test, expect, Page } from '@playwright/test'
import { spawn, ChildProcess } from 'child_process'
import { mkdtempSync } from 'fs'
import { tmpdir } from 'os'
import { join } from 'path'

async function devLogin(page: Page) {
  await page.goto('/api/dev-login')
  await page.waitForURL('**/installations', { timeout: 15_000 })
}

/** Run a shell command and stream output, returning the exit code */
function runScript(
  command: string,
  cwd: string,
  label: string,
): { process: ChildProcess; done: Promise<number> } {
  const child = spawn('bash', ['-c', command], {
    cwd,
    stdio: 'pipe',
    env: { ...process.env },
  })

  child.stdout?.on('data', (data) => {
    process.stdout.write(`[${label}:stdout] ${data}`)
  })
  child.stderr?.on('data', (data) => {
    process.stderr.write(`[${label}:stderr] ${data}`)
  })

  const done = new Promise<number>((resolve, reject) => {
    child.on('close', (code) => resolve(code ?? 1))
    child.on('error', (err) => reject(err))
  })

  return { process: child, done }
}

test.describe('Azure Installation Flow', () => {
  test.use({ storageState: { cookies: [], origins: [] } })

  test('complete Azure installation creation, deployment, and verification flow', async ({ page }) => {
    // Azure deployments can be slower — use generous timeout, will tune after first run
    test.setTimeout(1_800_000)

    const testId = Date.now().toString(36)
    const testName = `E2E Azure ${testId}`
    const testSubdomain = `e2e-az-${testId}`
    let installationId: string | null = null
    let installationKey: string | null = null
    let installScriptProcess: ChildProcess | null = null
    let installScriptCompleted = false
    let uninstallCompleted = false

    const scriptWorkDir = mkdtempSync(join(tmpdir(), 'kl-e2e-az-'))

    try {
      // ========== Step 1: Login and navigate to new installation ==========
      await devLogin(page)

      await page.getByRole('link', { name: 'New Installation' }).click()
      await page.waitForURL('**/installations/new')

      await expect(page.getByRole('heading', { name: 'Create Installation' })).toBeVisible()

      // ========== Step 2: Fill the create installation form ==========
      await page.getByPlaceholder('e.g., Production').fill(testName)
      await page.getByPlaceholder('Production deployment for our platform').fill(
        'Automated E2E Azure test installation',
      )

      await page.getByPlaceholder('your-company').fill(testSubdomain)
      await expect(page.getByText('Domain is available')).toBeVisible({ timeout: 10_000 })

      const submitButton = page.getByRole('button', { name: 'Continue to Installation' })
      await expect(submitButton).toBeEnabled()

      // ========== Step 3: Submit form and capture installation ID & key ==========
      const createResponsePromise = page.waitForResponse(
        (resp) =>
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

      // ========== Step 4: Verify install page and switch to Azure tab ==========
      await page.waitForURL('**/installations/new/install', { timeout: 15_000 })
      await expect(
        page.getByRole('heading', { name: 'Install Kloudlite in Your Cloud' }),
      ).toBeVisible({ timeout: 15_000 })

      // Switch to Azure tab
      await page.getByRole('tab', { name: 'Azure' }).click()

      // Verify Azure-specific content
      await expect(page.getByText('Azure CLI configured')).toBeVisible()
      await expect(page.getByText('Select Azure Location:')).toBeVisible()

      // Verify the Azure install command
      const azureCommandText = await page.locator('code').first().textContent()
      expect(azureCommandText).toContain('curl -fsSL https://get.khost.dev/install/azure')
      expect(azureCommandText).toContain(`--key ${installationKey}`)
      console.log(`[test] Azure install command: ${azureCommandText}`)

      // ========== Step 5: Execute the Azure installation script ==========
      const installCommand = azureCommandText!
      console.log(`\n[install-script] Working directory: ${scriptWorkDir}`)
      console.log(`[install-script] Executing: ${installCommand}\n`)

      const installScript = runScript(installCommand, scriptWorkDir, 'install-script')
      installScriptProcess = installScript.process

      // ========== Step 6: Wait for the UI to detect the deployment ==========
      await expect(
        page.getByText('Deployment verified. Configuring DNS...'),
      ).toBeVisible({ timeout: 600_000 })

      console.log('\n[test] Deployment verified by UI — script has registered with console')

      // ========== Step 7: Wait for DNS configuration and redirect ==========
      // Azure LB + DNS propagation may take a few minutes
      await expect(
        page.getByText('Installation complete! Redirecting...'),
      ).toBeVisible({ timeout: 900_000 })

      console.log('[test] DNS configured — redirecting to complete page')

      // ========== Step 8: Verify complete page ==========
      await page.waitForURL('**/installations/new/complete', { timeout: 30_000 })

      await expect(
        page.getByRole('heading', { name: /Setting Up Your Installation|Installation Complete!/ }),
      ).toBeVisible({ timeout: 30_000 })

      console.log('[test] Complete page loaded — waiting for install script to finish...')

      // Wait for install script to complete
      const exitCode = await installScript.done
      installScriptCompleted = true
      console.log(`[install-script] Exited with code ${exitCode}`)
      expect(exitCode).toBe(0)

      // Debug: check current ping status
      const pingStatus = await page.evaluate(async (id) => {
        const resp = await fetch(`/api/installations/${id}/ping`)
        return resp.json()
      }, installationId)
      console.log(`[test] Ping status after script completed: ${JSON.stringify(pingStatus)}`)

      // ========== Step 9: Wait for installation to become active ==========
      await expect(
        page.getByRole('heading', { name: 'Installation Complete!' }),
      ).toBeVisible({ timeout: 600_000 })

      console.log('[test] Installation is active!')

      await expect(
        page.getByText('Your Kloudlite installation is ready to use'),
      ).toBeVisible()

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
        timeout: 15_000,
      })

      await expect(
        page.getByRole('heading', { name: 'Installation Details' }),
      ).toBeVisible({ timeout: 10_000 })

      await expect(page.getByText('ACTIVE', { exact: true })).toBeVisible()

      // ========== Step 12: Generate and verify superadmin login URL ==========
      await expect(
        page.getByRole('heading', { name: 'Super Admin Access' }),
      ).toBeVisible()

      await page.getByRole('button', { name: 'Generate Login URL' }).click()
      await expect(page.getByText(/Expires in \d+:\d+/)).toBeVisible({ timeout: 10_000 })

      const loginUrlCode = page.locator('code').filter({ hasText: 'superadmin-login' })
      await expect(loginUrlCode).toBeVisible()
      const superadminLoginUrl = await loginUrlCode.textContent()
      expect(superadminLoginUrl).toContain(`${testSubdomain}.khost.dev/superadmin-login?token=`)
      console.log(`[test] Generated superadmin login URL: ${superadminLoginUrl}`)

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
      await expect(page.getByText(testName)).toBeVisible({ timeout: 10_000 })

      // ========== Step 15: Run uninstall script to tear down Azure resources ==========
      console.log(`\n[uninstall-script] Working directory: ${scriptWorkDir}`)
      console.log(`[uninstall-script] Executing: ${uninstallCommand}\n`)

      const uninstallScript = runScript(uninstallCommand!, scriptWorkDir, 'uninstall-script')
      const uninstallExitCode = await uninstallScript.done

      console.log(`[uninstall-script] Exited with code ${uninstallExitCode}`)
      expect(uninstallExitCode).toBe(0)
      uninstallCompleted = true

      // ========== Step 16: Delete installation record from console ==========
      await page.evaluate(async (id) => {
        await fetch(`/api/installations/${id}/delete`, { method: 'DELETE' })
      }, installationId)

      console.log(`[cleanup] Deleted installation record ${installationId}`)

      await page.reload()
      await expect(page.getByText(testName)).not.toBeVisible({ timeout: 10_000 })

      console.log('[test] Full Azure installation lifecycle complete!')

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
          const cleanupUninstallCmd = `curl -fsSL https://get.khost.dev/uninstall/azure | bash -s -- --key ${installationKey}`
          const cleanupScript = runScript(cleanupUninstallCmd, scriptWorkDir, 'cleanup-uninstall')
          const cleanupCode = await cleanupScript.done
          console.log(`[cleanup] Uninstall exited with code ${cleanupCode}`)
        } catch {
          console.log('[cleanup] Uninstall script failed — Azure resources may need manual cleanup')
        }
      }

      // Delete installation record if test failed before step 16
      if (installationId) {
        try {
          await page.evaluate(async (id) => {
            await fetch(`/api/installations/${id}/delete`, { method: 'DELETE' })
          }, installationId)
          console.log(`[cleanup] Deleted installation ${installationId}`)
        } catch {
          console.log(`[cleanup] Failed to delete installation ${installationId}`)
        }
      }
    }
  })
})
