import { KloudliteLogo } from "@/components/kloudlite-logo"
import { ThemeToggle } from "@/components/theme-toggle"

export default function TeamsLayout({
  children,
}: {
  children: React.ReactNode
}) {
  return (
    <div className="bg-muted flex min-h-svh flex-col items-center justify-center gap-6 p-6 md:p-10">
      <div className="absolute top-4 right-4">
        <ThemeToggle />
      </div>
      <div className="flex w-full max-w-lg flex-col gap-6">
        <div className="flex items-center gap-2 self-center">
          <KloudliteLogo className="h-6 w-auto" />
        </div>
        {children}
      </div>
    </div>
  )
}