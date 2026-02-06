import { FilesList } from '../../../_components/files-list'
import { listFiles } from '@/app/actions/environment-config'
import { getEnvironmentData } from '../../environment-data'
import { AlertCircle } from 'lucide-react'
import { Alert, AlertTitle, AlertDescription } from '@kloudlite/ui'

interface FilesPageProps {
  params: Promise<{
    id: string
  }>
}

// Error component
function FilesError({ error }: { error: string }) {
  return (
    <Alert variant="destructive">
      <AlertCircle className="h-5 w-5" />
      <AlertTitle>Error loading files</AlertTitle>
      <AlertDescription>{error}</AlertDescription>
    </Alert>
  )
}

export default async function FilesPage({ params }: FilesPageProps) {
  // id is now the environment hash
  const { id: hash } = await params
  try {
    // First get the environment name from the hash
    const envResult = await getEnvironmentData(hash)
    if (!envResult.success || !envResult.data) {
      return <FilesError error="Environment not found" />
    }
    const environmentName = envResult.data.environment.metadata?.name || ''

    const result = await listFiles(environmentName)
    const files = result.files || []

    return <FilesList environmentId={environmentName} files={files} />
  } catch (error) {
    return <FilesError error={error instanceof Error ? error.message : 'Failed to load files'} />
  }
}
