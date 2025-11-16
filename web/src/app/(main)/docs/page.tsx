import { redirect } from 'next/navigation'
import { getAppMode } from '@/lib/app-mode'

// Documentation root page - redirects to installation page
export default function Page() {
  // Only show docs page in website mode
  const mode = getAppMode()
  if (mode === 'website') {
    redirect('/docs/introduction/installation')
  }

  // Redirect to home for other modes
  return (
    <div className="flex min-h-screen items-center justify-center">
      <p className="text-muted-foreground">Documentation page is only available in website mode.</p>
    </div>
  )
}
