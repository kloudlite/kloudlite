import { redirect } from 'next/navigation'

interface TeamSettingsPageProps {
  params: Promise<{ teamname: string }>
}

export default async function TeamSettingsPage({ params }: TeamSettingsPageProps) {
  // Await params in Next.js 15
  const { teamname } = await params
  
  // Redirect to general settings as the default view
  redirect(`/${teamname}/settings/general`)
}