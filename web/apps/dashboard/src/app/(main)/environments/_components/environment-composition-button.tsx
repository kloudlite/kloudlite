import { getEnvironmentDetails } from '@/app/actions/environment.actions'
import { CompositionEditorTrigger } from './composition-editor-trigger'

interface EnvironmentCompositionButtonProps {
  environmentName: string
}

export async function EnvironmentCompositionButton({
  environmentName,
}: EnvironmentCompositionButtonProps) {
  // Fetch environment details to get compose content
  const result = await getEnvironmentDetails(environmentName)

  const composeContent = result.success && result.data ? result.data.compose?.composeContent || null : null

  return <CompositionEditorTrigger environmentName={environmentName} composeContent={composeContent} />
}
