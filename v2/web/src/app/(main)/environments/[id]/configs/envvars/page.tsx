import { EnvVarsList } from '@/components/envvars-list'

interface PageProps {
  params: {
    id: string
  }
}

export default function EnvVarsPage({ params }: PageProps) {
  return <EnvVarsList environmentId={params.id} />
}
