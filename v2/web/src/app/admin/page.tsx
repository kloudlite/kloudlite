import { redirect } from 'next/navigation'
import { auth } from '@/lib/auth'
import { isSystemReady } from '@/lib/system-check'

export default async function AdminPage() {
  const systemReady = await isSystemReady()

  // Redirect to appropriate page
  redirect(systemReady ? '/admin/users' : '/admin/machine-configs')
}