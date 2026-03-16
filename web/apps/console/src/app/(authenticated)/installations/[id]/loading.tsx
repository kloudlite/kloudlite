export default function Loading() {
  return (
    <div className="space-y-6">
      {/* Installation key bar */}
      <div className="rounded-lg border border-foreground/10 bg-background px-4 py-3 flex items-center gap-3">
        <div className="h-4 w-4 rounded bg-muted/50 animate-pulse" />
        <div className="h-4 w-24 rounded bg-muted/40 animate-pulse" />
        <div className="h-4 w-64 rounded bg-muted/30 animate-pulse" />
      </div>

      {/* Main action card */}
      <div className="rounded-lg border border-foreground/10 bg-background p-6 space-y-4">
        <div className="flex items-center gap-2">
          <div className="h-5 w-5 rounded bg-muted/50 animate-pulse" />
          <div className="h-5 w-40 rounded bg-muted/50 animate-pulse" />
        </div>
        <div className="h-4 w-72 rounded bg-muted/40 animate-pulse" />
        <div className="h-4 w-96 rounded bg-muted/30 animate-pulse" />
        <div className="h-9 w-40 rounded-md bg-muted/50 animate-pulse" />
      </div>
    </div>
  )
}
