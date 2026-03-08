export default function Loading() {
  return (
    <div className="bg-background h-screen flex flex-col">
      <div className="mx-auto max-w-7xl w-full px-6 lg:px-12 py-8 space-y-6 animate-pulse">
        <div className="flex items-center justify-between">
          <div className="h-8 w-48 rounded bg-muted" />
          <div className="h-10 w-32 rounded bg-muted" />
        </div>
        <div className="space-y-3">
          {Array.from({ length: 3 }).map((_, i) => (
            <div key={i} className="h-20 rounded-lg border border-foreground/10 bg-muted/30" />
          ))}
        </div>
      </div>
    </div>
  )
}
