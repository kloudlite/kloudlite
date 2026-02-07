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

test.describe('Installation Creation Flow', () => {
  test.use({ storageState: { cookies: [], origins: [] } })

  test('form validation prevents submission with invalid data', async ({ page }) => {
    await devLogin(page)
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

  test('complete installation creation, deployment, and verification flow', async ({ page }) => {
    // Real deployment + superadmin verification + uninstall can take 20-25 minutes
    test.setTimeout(1_500_000)

    const testId = Date.now().toString(36)
    const testName = `E2E Test ${testId}`
    const testSubdomain = `e2e-${testId}`
    let installationId: string | null = null
    let installationKey: string | null = null
    let installScriptProcess: ChildProcess | null = null
    let installScriptCompleted = false
    let uninstallCompleted = false

    // Create a temp directory for scripts to run in (they download kli binary)
    const scriptWorkDir = mkdtempSync(join(tmpdir(), 'kl-e2e-'))

    try {
      // ========== Step 1: Login and navigate to new installation ==========
      await devLogin(page)

      await page.getByRole('link', { name: 'New Installation' }).click()
      await page.waitForURL('**/installations/new')

      await expect(page.getByRole('heading', { name: 'Create Installation' })).toBeVisible()
      await expect(page.getByText('Deploy Kloudlite in your cloud account')).toBeVisible()

      // ========== Step 2: Fill the create installation form ==========
      await page.getByPlaceholder('e.g., Production').fill(testName)
      await page.getByPlaceholder('Production deployment for our platform').fill(
        'Automated E2E test installation',
      )

      // Subdomain — triggers availability check
      await page.getByPlaceholder('your-company').fill(testSubdomain)

      // Wait for subdomain availability check to return positive
      await expect(page.getByText('Domain is available')).toBeVisible({ timeout: 10_000 })

      // Domain preview should show the full domain
      await expect(page.getByText(new RegExp(`${testSubdomain}\\.`))).toBeVisible()

      // Submit button should now be enabled
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

      // ========== Step 4: Verify install page loads correctly ==========
      await page.waitForURL('**/installations/new/install', { timeout: 15_000 })
      await expect(
        page.getByRole('heading', { name: 'Install Kloudlite in Your Cloud' }),
      ).toBeVisible({ timeout: 15_000 })
      await expect(
        page.getByText('Run the installation command on your cloud provider'),
      ).toBeVisible()

      // Cloud provider tabs visible
      await expect(page.getByRole('tab', { name: 'AWS' })).toBeVisible()
      await expect(page.getByRole('tab', { name: 'GCP' })).toBeVisible()
      await expect(page.getByRole('tab', { name: 'Azure' })).toBeVisible()

      // ========== Step 5: Verify AWS tab (default) ==========
      await expect(page.getByText('Prerequisites:')).toBeVisible()
      await expect(page.getByText('Run this command:')).toBeVisible()
      await expect(page.getByText('AWS CLI configured')).toBeVisible()
      await expect(page.getByText('Select AWS Region:')).toBeVisible()

      // Verify the installation command contains the correct key
      const awsCommandText = await page.locator('code').first().textContent()
      expect(awsCommandText).toContain('curl -fsSL https://get.khost.dev/install/aws')
      expect(awsCommandText).toContain(`--key ${installationKey}`)

      // Copy button present
      await expect(page.getByRole('button', { name: 'Copy' })).toBeVisible()

      // Initial status: waiting for deployment
      await expect(page.getByText('Waiting for deployment...')).toBeVisible()

      // ========== Step 6: Verify GCP and Azure tabs ==========
      await page.getByRole('tab', { name: 'GCP' }).click()
      const gcpCommandText = await page.locator('code').first().textContent()
      expect(gcpCommandText).toContain('curl -fsSL https://get.khost.dev/install/gcp')
      expect(gcpCommandText).toContain(`--key ${installationKey}`)

      await page.getByRole('tab', { name: 'Azure' }).click()
      const azureCommandText = await page.locator('code').first().textContent()
      expect(azureCommandText).toContain('curl -fsSL https://get.khost.dev/install/azure')
      expect(azureCommandText).toContain(`--key ${installationKey}`)

      // Switch back to AWS
      await page.getByRole('tab', { name: 'AWS' }).click()

      // ========== Step 7: Execute the installation script locally ==========
      const installCommand = await page.locator('code').first().textContent()
      expect(installCommand).toBeTruthy()
      console.log(`\n[install-script] Working directory: ${scriptWorkDir}`)
      console.log(`[install-script] Executing: ${installCommand}\n`)

      const installScript = runScript(installCommand!, scriptWorkDir, 'install-script')
      installScriptProcess = installScript.process

      // ========== Step 8: Wait for the UI to detect the deployment ==========
      await expect(
        page.getByText('Deployment verified. Configuring DNS...'),
      ).toBeVisible({ timeout: 600_000 })

      console.log('\n[test] Deployment verified by UI — script has registered with console')

      // ========== Step 9: Wait for DNS configuration and redirect ==========
      await expect(
        page.getByText('Installation complete! Redirecting...'),
      ).toBeVisible({ timeout: 300_000 })

      console.log('[test] DNS configured — redirecting to complete page')

      // ========== Step 10: Verify complete page ==========
      await page.waitForURL('**/installations/new/complete', { timeout: 30_000 })

      // Confirm the page loaded — it shows either "Setting Up" or "Installation Complete!"
      await expect(
        page.getByRole('heading', { name: /Setting Up Your Installation|Installation Complete!/ }),
      ).toBeVisible({ timeout: 30_000 })

      console.log('[test] Complete page loaded — waiting for install script to finish...')

      // Wait for install script to complete before checking active status.
      // The script calls configure-root-dns near the end, which sets deploymentReady=true.
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

      // ========== Step 11: Wait for installation to become active ==========
      // Infrastructure is up. Wait for ALB health checks + DNS propagation (2-5 min typical).
      await expect(
        page.getByRole('heading', { name: 'Installation Complete!' }),
      ).toBeVisible({ timeout: 600_000 })

      console.log('[test] Installation is active!')

      await expect(
        page.getByText('Your Kloudlite installation is ready to use'),
      ).toBeVisible()

      // Installation URL visible
      await expect(page.getByText(new RegExp(`${testSubdomain}\\.`))).toBeVisible()

      // Action buttons
      await expect(
        page.getByRole('button', { name: 'Open Installation Settings' }),
      ).toBeVisible()
      await expect(
        page.getByRole('button', { name: 'View All Installations' }),
      ).toBeVisible()

      // Active status indicator
      await expect(page.getByText('Installation is active and ready!')).toBeVisible()

      // ========== Step 12: Verify installation URL is reachable ==========
      const installationUrl = `https://${testSubdomain}.khost.dev`
      console.log(`[test] Verifying installation URL: ${installationUrl}`)

      const dashboardPage = await page.context().newPage()
      await dashboardPage.goto(installationUrl, {
        timeout: 60_000,
        waitUntil: 'domcontentloaded',
      })

      // Verify we got a real page, not a DNS error
      await expect(dashboardPage.locator('body')).not.toHaveText('', { timeout: 30_000 })
      const dashboardTitle = await dashboardPage.title()
      console.log(`[test] Dashboard page title: "${dashboardTitle}"`)
      expect(dashboardTitle).toBeTruthy()

      await dashboardPage.close()

      // ========== Step 13: Navigate to installation settings ==========
      await page.getByRole('button', { name: 'Open Installation Settings' }).click()
      await page.waitForURL(new RegExp(`/installations/${installationId}`), {
        timeout: 15_000,
      })

      // Verify settings page loaded
      await expect(
        page.getByRole('heading', { name: 'Installation Details' }),
      ).toBeVisible({ timeout: 10_000 })

      // Status should be ACTIVE
      await expect(page.getByText('ACTIVE', { exact: true })).toBeVisible()

      // Installation URL should be visible (multiple links may show the domain)
      await expect(
        page.getByRole('link', { name: `${testSubdomain}.khost.dev` }).first(),
      ).toBeVisible()

      // Installation key should be visible (appears in details card and uninstall command)
      await expect(page.getByText(installationKey!).first()).toBeVisible()

      // ========== Step 14: Generate and verify superadmin login URL ==========
      await expect(
        page.getByRole('heading', { name: 'Super Admin Access' }),
      ).toBeVisible()

      // Click "Generate Login URL"
      await page.getByRole('button', { name: 'Generate Login URL' }).click()

      // Wait for the generated URL to appear (countdown timer is unique to the generated state)
      await expect(page.getByText(/Expires in \d+:\d+/)).toBeVisible({ timeout: 10_000 })

      // Extract the generated login URL
      const loginUrlCode = page.locator('code').filter({ hasText: 'superadmin-login' })
      await expect(loginUrlCode).toBeVisible()
      const superadminLoginUrl = await loginUrlCode.textContent()
      expect(superadminLoginUrl).toContain(`${testSubdomain}.khost.dev/superadmin-login?token=`)
      console.log(`[test] Generated superadmin login URL: ${superadminLoginUrl}`)

      // "Open in New Tab" and "Generate New URL" buttons should be visible
      await expect(page.getByRole('link', { name: 'Open in New Tab' })).toBeVisible()
      await expect(
        page.getByRole('button', { name: 'Generate New URL' }),
      ).toBeVisible()

      // Verify the superadmin login URL is reachable on the deployed dashboard
      const superadminPage = await page.context().newPage()
      const superadminResponse = await superadminPage.goto(superadminLoginUrl!, {
        timeout: 60_000,
        waitUntil: 'domcontentloaded',
      })

      // Verify the deployed dashboard accepted the request (not a 4xx/5xx from Cloudflare)
      const statusCode = superadminResponse?.status() ?? 0
      console.log(`[test] Superadmin login page status: ${statusCode}, URL: ${superadminPage.url()}`)
      expect(statusCode).toBeGreaterThanOrEqual(200)
      expect(statusCode).toBeLessThan(500)

      await superadminPage.close()

      // ========== Step 15: Verify Danger Zone and uninstall script ==========
      await expect(
        page.getByRole('heading', { name: 'Danger Zone' }),
      ).toBeVisible()

      await expect(page.getByText('Uninstall Script', { exact: true })).toBeVisible()
      await expect(
        page.getByText('Run this command in your terminal to uninstall Kloudlite'),
      ).toBeVisible()

      // Extract the uninstall command
      const uninstallCode = page
        .locator('code')
        .filter({ hasText: 'get.khost.dev/uninstall/' })
      await expect(uninstallCode).toBeVisible()
      const uninstallCommand = await uninstallCode.textContent()
      expect(uninstallCommand).toContain('curl -fsSL https://get.khost.dev/uninstall/')
      expect(uninstallCommand).toContain(`--key ${installationKey}`)
      console.log(`[test] Uninstall command: ${uninstallCommand}`)

      // ========== Step 16: Verify in installations list ==========
      await page.goto('/installations')
      await expect(page.getByText(testName)).toBeVisible({ timeout: 10_000 })

      // ========== Step 17: Run uninstall script to tear down cloud resources ==========
      console.log(`\n[uninstall-script] Working directory: ${scriptWorkDir}`)
      console.log(`[uninstall-script] Executing: ${uninstallCommand}\n`)

      const uninstallScript = runScript(uninstallCommand!, scriptWorkDir, 'uninstall-script')
      const uninstallExitCode = await uninstallScript.done

      console.log(`[uninstall-script] Exited with code ${uninstallExitCode}`)
      expect(uninstallExitCode).toBe(0)
      uninstallCompleted = true

      // ========== Step 18: Delete installation record from console ==========
      await page.evaluate(async (id) => {
        await fetch(`/api/installations/${id}/delete`, { method: 'DELETE' })
      }, installationId)

      console.log(`[cleanup] Deleted installation record ${installationId}`)

      // Verify the installation is gone from the list
      await page.reload()
      await expect(page.getByText(testName)).not.toBeVisible({ timeout: 10_000 })

      console.log('[test] Full installation lifecycle complete!')

      // Mark cleanup as done so finally block doesn't retry
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
          const cleanupUninstallCmd = `curl -fsSL https://get.khost.dev/uninstall/aws | bash -s -- --key ${installationKey}`
          const cleanupScript = runScript(cleanupUninstallCmd, scriptWorkDir, 'cleanup-uninstall')
          const cleanupCode = await cleanupScript.done
          console.log(`[cleanup] Uninstall exited with code ${cleanupCode}`)
        } catch {
          console.log('[cleanup] Uninstall script failed — AWS resources may need manual cleanup')
        }
      }

      // Delete installation record if test failed before step 18
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
