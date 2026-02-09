import type { Page } from '@playwright/test'
import { spawn, ChildProcess } from 'child_process'
import { DEV_TOKEN, TIMEOUTS } from './constants'

export async function devLogin(page: Page) {
  await page.goto('/api/dev-login')
  await page.waitForURL('**/installations', { timeout: TIMEOUTS.auth })
}

export async function superadminLogin(page: Page, token = DEV_TOKEN) {
  await page.goto(`/superadmin-login?token=${token}`)
  await page.waitForURL('**/admin/**', { timeout: TIMEOUTS.auth })
}

export function runScript(
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

export function escapeRegex(str: string) {
  return str.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')
}
