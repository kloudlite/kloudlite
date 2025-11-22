import { redirect } from 'next/navigation'

// Documentation root page - redirects to installation page
export default function Page() {
  redirect('/docs/introduction/installation')
}
