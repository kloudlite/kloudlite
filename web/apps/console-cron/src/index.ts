import { processDueJobs } from './renewal'

const POLL_INTERVAL = 5 * 60 * 1000 // 5 minutes

async function runCycle(): Promise<void> {
  const start = Date.now()
  console.log(`[cron] Cycle started at ${new Date().toISOString()}`)

  try {
    await processDueJobs()
  } catch (err) {
    console.error('[cron] Cycle failed:', err)
  }

  const elapsed = Date.now() - start
  console.log(`[cron] Cycle completed in ${elapsed}ms`)
}

// Run immediately on startup, then every 5 minutes
runCycle()
setInterval(runCycle, POLL_INTERVAL)

console.log('[cron] console-cron started — polling every 5 minutes for scheduled jobs')
