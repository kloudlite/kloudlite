import { redirect } from 'next/navigation'

export default function TeamManagementPage() {
  // Team management has moved to organization settings
  redirect('/installations')
}
