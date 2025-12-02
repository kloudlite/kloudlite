import { redirect } from 'next/navigation'

// Redirect to the default artifact type (container images)
export default function ArtifactsPage() {
  redirect('/artifacts/container-images')
}
