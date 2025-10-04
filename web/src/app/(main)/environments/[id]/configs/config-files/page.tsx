import { FilesList } from '../_components/files-list'

interface FilesPageProps {
  params: {
    id: string
  }
}

export default function FilesPage({ params }: FilesPageProps) {
  return <FilesList environmentId={params.id} />
}
