import { ResourceSectionNav } from '@/components/resource-section-nav'

interface LayoutProps {
  children: React.ReactNode
  params: {
    id: string
  }
}

export default function ResourcesLayout({ children, params }: LayoutProps) {
  return (
    <div className="mx-auto max-w-7xl px-6 py-8">
      <div className="flex gap-6">
        <ResourceSectionNav environmentId={params.id} />
        <div className="flex-1">
          {children}
        </div>
      </div>
    </div>
  )
}
