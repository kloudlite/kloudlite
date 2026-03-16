export default function Loading() {
  return (
    <main className="mx-auto max-w-6xl w-full px-6 lg:px-12 py-10 space-y-6">
        {/* Title row */}
        <div className="flex items-center justify-between">
          <div className="h-7 w-40 rounded-md bg-muted/50 animate-pulse" />
          <div className="flex items-center gap-3">
            <div className="h-9 w-52 rounded-md bg-muted/50 animate-pulse" />
            <div className="h-9 w-36 rounded-md bg-muted/50 animate-pulse" />
          </div>
        </div>

        {/* Tabs */}
        <div className="flex gap-4 border-b border-foreground/10 pb-2">
          <div className="h-4 w-12 rounded bg-muted/50 animate-pulse" />
          <div className="h-4 w-16 rounded bg-muted/40 animate-pulse" />
          <div className="h-4 w-18 rounded bg-muted/40 animate-pulse" />
        </div>

        {/* Table */}
        <div className="rounded-lg border border-foreground/10 overflow-hidden">
          <div className="bg-muted/20 px-6 py-3 flex gap-6">
            <div className="h-3 w-16 rounded bg-muted/50 animate-pulse" />
            <div className="h-3 w-16 rounded bg-muted/50 animate-pulse" />
            <div className="h-3 w-16 rounded bg-muted/50 animate-pulse" />
            <div className="h-3 w-16 rounded bg-muted/50 animate-pulse" />
          </div>
          {[1, 2, 3].map((i) => (
            <div key={i} className="px-6 py-4 border-t border-foreground/5 flex items-center gap-6">
              <div className="h-4 w-32 rounded bg-muted/40 animate-pulse" />
              <div className="h-5 w-12 rounded-md bg-muted/30 animate-pulse" />
              <div className="h-4 w-36 rounded bg-muted/40 animate-pulse" />
              <div className="h-5 w-16 rounded-md bg-muted/30 animate-pulse" />
            </div>
          ))}
        </div>
    </main>
  )
}
