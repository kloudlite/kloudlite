import { redirect } from 'next/navigation'

// Redirect to the default artifact type (container repos)
export default function ArtifactsPage() {
  redirect('/artifacts/container-repos')
}
