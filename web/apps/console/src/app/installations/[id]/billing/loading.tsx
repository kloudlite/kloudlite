export default function Loading() {
  return (
    <div className="space-y-6 animate-pulse p-8">
      <div className="h-8 w-48 rounded bg-muted" />
      <div className="h-4 w-72 rounded bg-muted" />
      <div className="h-64 rounded-lg border border-foreground/10 bg-muted/30 mt-6" />
    </div>
  )
}
