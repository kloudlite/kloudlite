export default function Loading() {
  return (
    <main className="mx-auto max-w-6xl w-full px-6 lg:px-12 py-10 space-y-8">
        {/* Title */}
        <div className="space-y-2">
          <div className="h-7 w-24 rounded bg-muted/50 animate-pulse" />
          <div className="h-4 w-52 rounded bg-muted/30 animate-pulse" />
        </div>

        {/* Profile card */}
        <div className="rounded-lg border border-foreground/10 bg-background p-6 space-y-5">
          <div className="space-y-1.5">
            <div className="h-5 w-36 rounded bg-muted/50 animate-pulse" />
            <div className="h-3 w-52 rounded bg-muted/30 animate-pulse" />
          </div>

          {/* Avatar */}
          <div className="flex items-start gap-6">
            <div className="h-24 w-24 rounded-full bg-muted/40 animate-pulse" />
            <div className="space-y-2 pt-2">
              <div className="h-4 w-24 rounded bg-muted/50 animate-pulse" />
              <div className="h-3 w-48 rounded bg-muted/30 animate-pulse" />
            </div>
          </div>

          <div className="h-px bg-foreground/10" />

          {/* Fields */}
          {[1, 2, 3].map((i) => (
            <div key={i} className="space-y-2">
              <div className="h-4 w-24 rounded bg-muted/40 animate-pulse" />
              <div className="h-10 w-full rounded border border-foreground/10 bg-muted/20 animate-pulse" />
              {i < 3 && <div className="h-px bg-foreground/10 mt-3" />}
            </div>
          ))}
        </div>
    </main>
  )
}
