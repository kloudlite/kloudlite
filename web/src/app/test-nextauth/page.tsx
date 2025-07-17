'use client'

import { useSession, signIn, signOut } from 'next-auth/react'

export default function TestNextAuth() {
  const { data: session, status } = useSession()

  if (status === 'loading') return <p>Loading...</p>

  if (session) {
    return (
      <div className="p-8">
        <h1 className="text-2xl font-bold mb-4">NextAuth Test - Signed In</h1>
        <pre className="bg-gray-100 p-4 rounded mb-4">
          {JSON.stringify(session, null, 2)}
        </pre>
        <button 
          onClick={() => signOut()}
          className="bg-red-500 text-white px-4 py-2 rounded"
        >
          Sign Out
        </button>
      </div>
    )
  }

  return (
    <div className="p-8">
      <h1 className="text-2xl font-bold mb-4">NextAuth Test - Not Signed In</h1>
      <div className="space-y-4">
        <button 
          onClick={() => signIn('google')}
          className="bg-blue-500 text-white px-4 py-2 rounded block"
        >
          Sign in with Google
        </button>
        <button 
          onClick={() => signIn('github')}
          className="bg-gray-800 text-white px-4 py-2 rounded block"
        >
          Sign in with GitHub
        </button>
        <button 
          onClick={() => signIn('azure-ad')}
          className="bg-blue-600 text-white px-4 py-2 rounded block"
        >
          Sign in with Microsoft
        </button>
      </div>
    </div>
  )
}