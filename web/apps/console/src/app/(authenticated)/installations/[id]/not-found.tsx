import Link from 'next/link'

export default function NotFound() {
  return (
    <div className="flex min-h-[400px] flex-col items-center justify-center gap-4 p-8">
      <h2 className="text-lg font-semibold">Installation not found</h2>
      <p className="text-muted-foreground text-sm">
        The installation you are looking for does not exist or you do not have access.
      </p>
      <Link
        href="/installations"
        className="bg-primary text-primary-foreground rounded-md px-4 py-2 text-sm"
      >
        Back to installations
      </Link>
    </div>
  )
}
