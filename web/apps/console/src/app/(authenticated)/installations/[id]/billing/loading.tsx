export default function Loading() {
  return (
    <div className="space-y-6">
      {/* Heading */}
      <div className="space-y-2">
        <div className="h-5 w-20 rounded bg-muted/50 animate-pulse" />
        <div className="h-4 w-56 rounded bg-muted/30 animate-pulse" />
      </div>

      {/* Balance card */}
      <div className="rounded-lg border border-foreground/10 bg-background p-6 space-y-4">
        <div className="h-4 w-28 rounded bg-muted/40 animate-pulse" />
        <div className="h-8 w-32 rounded bg-muted/50 animate-pulse" />
        <div className="h-4 w-48 rounded bg-muted/30 animate-pulse" />
      </div>

      {/* Usage table */}
      <div className="rounded-lg border border-foreground/10 bg-background p-6 space-y-4">
        <div className="h-5 w-32 rounded bg-muted/50 animate-pulse" />
        <div className="space-y-3">
          {[1, 2, 3].map((i) => (
            <div key={i} className="flex items-center justify-between">
              <div className="h-4 w-40 rounded bg-muted/30 animate-pulse" />
              <div className="h-4 w-20 rounded bg-muted/30 animate-pulse" />
            </div>
          ))}
        </div>
      </div>
    </div>
  )
}
