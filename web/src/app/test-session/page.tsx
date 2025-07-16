import { getSession } from '@/actions/auth/session'
import { cookies } from 'next/headers'

export default async function TestSessionPage() {
  const session = await getSession()
  const cookieStore = await cookies()
  const sessionCookie = cookieStore.get('kloudlite-session')
  
  return (
    <div className="container mx-auto py-10">
      <h1 className="text-2xl font-bold mb-4">Session Debug Info</h1>
      
      <div className="space-y-4">
        <div>
          <h2 className="font-semibold">Cookie Value:</h2>
          <pre className="bg-gray-100 p-2 rounded">
            {sessionCookie ? JSON.stringify(sessionCookie, null, 2) : 'No session cookie found'}
          </pre>
        </div>
        
        <div>
          <h2 className="font-semibold">Session Data:</h2>
          <pre className="bg-gray-100 p-2 rounded">
            {session ? JSON.stringify(session, null, 2) : 'No session found'}
          </pre>
        </div>
        
        <div className="flex gap-4">
          <a href="/auth/login" className="text-blue-600 underline">Go to Login</a>
          <a href="/dashboard" className="text-blue-600 underline">Go to Dashboard</a>
          <a href="/teams" className="text-blue-600 underline">Go to Teams</a>
        </div>
      </div>
    </div>
  )
}