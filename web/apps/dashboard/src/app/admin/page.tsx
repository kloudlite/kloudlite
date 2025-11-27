import { redirect } from 'next/navigation'
import { isSystemReady } from '@/lib/system-check'

// Force dynamic rendering - this page uses headers() and API calls
export const dynamic = 'force-dynamic'

export default async function AdminPage() {
  const systemReady = await isSystemReady()

  // Redirect to appropriate page
  redirect(systemReady ? '/admin/users' : '/admin/machine-configs')
}
