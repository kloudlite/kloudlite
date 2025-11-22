import { ScrollArea } from '@kloudlite/ui'

// Website layout - simple layout for marketing and documentation
export default async function MainLayout({ children }: { children: React.ReactNode}) {
  return (
    <ScrollArea className="h-screen">
      <div className="min-h-screen">{children}</div>
    </ScrollArea>
  )
}
