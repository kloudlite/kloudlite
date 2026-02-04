import { SettingsSectionNav } from '../../_components/settings-section-nav'

interface LayoutProps {
  children: React.ReactNode
  params: Promise<{
    id: string
  }>
}

export default async function SettingsLayout({ children, params }: LayoutProps) {
  const { id } = await params
  return (
    <div className="flex gap-8">
      <SettingsSectionNav environmentId={id} />
      <div className="flex-1 min-w-0">{children}</div>
    </div>
  )
}
