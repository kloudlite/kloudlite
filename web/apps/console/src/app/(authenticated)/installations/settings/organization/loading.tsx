export default function Loading() {
  return (
    <div className="space-y-8">
      {/* Org info */}
      <div className="flex items-center gap-3">
        <div className="h-5 w-5 rounded bg-muted/50 animate-pulse" />
        <div className="space-y-1.5">
          <div className="h-5 w-36 rounded bg-muted/50 animate-pulse" />
          <div className="h-3 w-24 rounded bg-muted/30 animate-pulse" />
        </div>
      </div>

      {/* Team Members */}
      <div className="space-y-5">
        <div className="flex items-center justify-between">
          <div className="space-y-1.5">
            <div className="h-5 w-32 rounded bg-muted/50 animate-pulse" />
            <div className="h-3 w-56 rounded bg-muted/30 animate-pulse" />
          </div>
          <div className="h-9 w-28 rounded-md bg-muted/50 animate-pulse" />
        </div>

        {/* Table */}
        <div className="rounded-lg border border-foreground/10 overflow-hidden">
          <div className="bg-muted/20 px-6 py-3 flex gap-8">
            <div className="h-3 w-16 rounded bg-muted/50 animate-pulse" />
            <div className="h-3 w-12 rounded bg-muted/50 animate-pulse" />
            <div className="h-3 w-10 rounded bg-muted/50 animate-pulse" />
            <div className="h-3 w-12 rounded bg-muted/50 animate-pulse" />
          </div>
          {[1, 2].map((i) => (
            <div key={i} className="px-6 py-4 border-t border-foreground/5 flex items-center gap-8">
              <div className="h-4 w-28 rounded bg-muted/40 animate-pulse" />
              <div className="h-4 w-40 rounded bg-muted/30 animate-pulse" />
              <div className="h-5 w-16 rounded-md bg-muted/30 animate-pulse" />
              <div className="h-4 w-20 rounded bg-muted/30 animate-pulse" />
            </div>
          ))}
        </div>
      </div>
    </div>
  )
}
