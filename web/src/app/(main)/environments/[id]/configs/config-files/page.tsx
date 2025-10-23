import { FilesList } from '../../../_components/files-list'
import { listFiles } from '@/app/actions/environment-config'
import { AlertCircle } from 'lucide-react'

interface FilesPageProps {
  params: Promise<{
    id: string
  }>
}

// Error component
function FilesError({ error }: { error: string }) {
  return (
    <div className="rounded-lg bg-red-50 border border-red-200 p-4">
      <div className="flex items-center gap-2 text-red-800">
        <AlertCircle className="h-5 w-5" />
        <span className="font-medium">Error loading files</span>
      </div>
      <p className="mt-2 text-sm text-red-700">{error}</p>
    </div>
  )
}

export default async function FilesPage({ params }: FilesPageProps) {
  const { id } = await params
  try {
    const result = await listFiles(id)
    const files = result.files || []

    return <FilesList environmentId={id} files={files} />
  } catch (error) {
    return <FilesError error={error instanceof Error ? error.message : 'Failed to load files'} />
  }
}
