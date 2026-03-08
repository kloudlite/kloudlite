import { redirect } from 'next/navigation'
import { getRegistrationSession } from '@/lib/console-auth'
import { getSelectedOrg } from '@/lib/console/get-selected-org'
import { NewByocContent } from './new-byoc-content'

export default async function NewByocPage() {
  const session = await getRegistrationSession()
  if (!session?.user) redirect('/login')

  const currentOrg = await getSelectedOrg(session.user.id, session.user.name, session.user.email)
  if (!currentOrg) redirect('/installations')

  return <NewByocContent orgId={currentOrg.id} />
}
