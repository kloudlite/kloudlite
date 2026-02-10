import { redirect } from 'next/navigation'
import { isSystemReady } from '@/lib/system-check'
import { env } from '@/lib/env'

// Force dynamic rendering - this page uses headers() and API calls
export const dynamic = 'force-dynamic'

export default async function AdminPage() {
  // In Kloudlite Cloud mode, always redirect to users (machine configs are admin-managed)
  if (env.isKloudliteCloud) {
    redirect('/admin/users')
  }

  const systemReady = await isSystemReady()

  // Redirect to appropriate page
  redirect(systemReady ? '/admin/users' : '/admin/machine-configs')
}
