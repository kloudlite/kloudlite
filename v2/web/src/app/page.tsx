import { auth } from '@/lib/auth'
import { AuthButton } from '@/components/auth-button'

export default async function Home() {
  const session = await auth()

  return (
    <main className="min-h-screen p-8">
      <div className="flex justify-between items-center mb-8">
        <h1 className="text-2xl font-semibold">Kloudlite</h1>
        <AuthButton session={session} />
      </div>
      {session && (
        <div className="mt-8 p-4 border rounded-lg">
          <h2 className="text-lg font-medium mb-2">Session Info</h2>
          <pre className="text-sm">{JSON.stringify(session.user, null, 2)}</pre>
        </div>
      )}
    </main>
  )
}