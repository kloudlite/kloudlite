'use client'

import { signIn } from 'next-auth/react'

export default function TestSingleOAuth() {
  return (
    <div className="p-8">
      <h1 className="text-2xl font-bold mb-4">Test Single OAuth Provider</h1>
      <div className="space-y-4">
        <button 
          onClick={() => signIn('google')}
          className="bg-blue-500 text-white px-4 py-2 rounded"
        >
          Test Google OAuth
        </button>
        
        <div className="text-sm text-gray-600">
          <p>Click above to test Google OAuth specifically.</p>
          <p>Check the browser console for any errors.</p>
        </div>
      </div>
    </div>
  )
}