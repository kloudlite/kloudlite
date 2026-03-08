export default function Loading() {
  return (
    <div className="space-y-6 animate-pulse p-8">
      <div className="h-8 w-64 rounded bg-muted" />
      <div className="h-4 w-96 rounded bg-muted" />
      <div className="grid grid-cols-3 gap-4 mt-6">
        {Array.from({ length: 3 }).map((_, i) => (
          <div key={i} className="h-24 rounded-lg border border-foreground/10 bg-muted/30" />
        ))}
      </div>
    </div>
  )
}
