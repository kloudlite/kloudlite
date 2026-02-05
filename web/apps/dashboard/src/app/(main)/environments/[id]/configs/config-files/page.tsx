import { FilesList } from '../../../_components/files-list'
import { listFiles } from '@/app/actions/environment-config'
import { getEnvironmentData } from '../../environment-data'
import { AlertCircle } from 'lucide-react'

interface FilesPageProps {
  params: Promise<{
    id: string
  }>
}

// Error component
function FilesError({ error }: { error: string }) {
  return (
    <div className="rounded-lg border border-red-200 bg-red-50 p-4">
      <div className="flex items-center gap-2 text-red-800">
        <AlertCircle className="h-5 w-5" />
        <span className="font-medium">Error loading files</span>
      </div>
      <p className="mt-2 text-sm text-red-700">{error}</p>
    </div>
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
