import { redirect } from 'next/navigation'
import { APP_MODE } from '@/lib/app-mode'

// Documentation root page - redirects to installation page
function DocsPage() {
  redirect('/docs/introduction/installation')
}

export default function Page() {
  // Only show docs page in website mode
  if (APP_MODE === 'website') {
    return <DocsPage />
  }

  // Redirect to home for other modes
  return (
    <div className="flex min-h-screen items-center justify-center">
      <p className="text-muted-foreground">Documentation page is only available in website mode.</p>
    </div>
  )
}
