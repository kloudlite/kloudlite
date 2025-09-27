export default function Loading() {
  return (
    <div className="flex min-h-screen items-center justify-center bg-background">
      <div className="space-x-1">
        <span className="inline-block h-2 w-2 animate-pulse rounded-full bg-primary/60" />
        <span className="inline-block h-2 w-2 animate-pulse rounded-full bg-primary/60 [animation-delay:0.2s]" />
        <span className="inline-block h-2 w-2 animate-pulse rounded-full bg-primary/60 [animation-delay:0.4s]" />
      </div>
    </div>
  )
}