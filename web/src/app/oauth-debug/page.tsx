export default function OAuthDebugPage() {
  const baseUrl = process.env.NEXTAUTH_URL || 'http://localhost:3000'
  
  return (
    <div className="p-8 max-w-4xl mx-auto">
      <h1 className="text-2xl font-bold mb-6">OAuth Configuration Debug</h1>
      
      <div className="space-y-6">
        <div className="bg-gray-50 p-4 rounded-lg">
          <h2 className="text-lg font-semibold mb-3">Required Redirect URIs</h2>
          <p className="text-sm text-gray-600 mb-4">
            Add these exact URLs to your OAuth provider settings:
          </p>
          
          <div className="space-y-4">
            <div>
              <h3 className="font-medium text-blue-600">Google OAuth:</h3>
              <code className="bg-white p-2 rounded border block">
                {baseUrl}/api/auth/callback/google
              </code>
            </div>
            
            <div>
              <h3 className="font-medium text-gray-800">GitHub OAuth:</h3>
              <code className="bg-white p-2 rounded border block">
                {baseUrl}/api/auth/callback/github
              </code>
            </div>
            
            <div>
              <h3 className="font-medium text-blue-500">Microsoft/Azure AD:</h3>
              <code className="bg-white p-2 rounded border block">
                {baseUrl}/api/auth/callback/azure-ad
              </code>
            </div>
          </div>
        </div>

        <div className="bg-yellow-50 p-4 rounded-lg">
          <h2 className="text-lg font-semibold mb-3">Common Issues</h2>
          <ul className="list-disc list-inside space-y-2 text-sm">
            <li>Make sure the redirect URI matches exactly (no trailing slashes)</li>
            <li>Check that your OAuth app is enabled/published</li>
            <li>Verify client ID and secret are correct</li>
            <li>For Google: Enable Google+ API and configure OAuth consent screen</li>
            <li>For Microsoft: Check that the app registration is in the correct tenant</li>
          </ul>
        </div>

        <div className="bg-blue-50 p-4 rounded-lg">
          <h2 className="text-lg font-semibold mb-3">Environment Variables</h2>
          <div className="space-y-2 text-sm">
            <div>NEXTAUTH_URL: <code>{baseUrl}</code></div>
            <div>NEXTAUTH_SECRET: {process.env.NEXTAUTH_SECRET ? '✓ Set' : '✗ Not set'}</div>
            <div>GOOGLE_CLIENT_ID: {process.env.GOOGLE_CLIENT_ID ? '✓ Set' : '✗ Not set'}</div>
            <div>GITHUB_CLIENT_ID: {process.env.GITHUB_CLIENT_ID ? '✓ Set' : '✗ Not set'}</div>
            <div>MICROSOFT_CLIENT_ID: {process.env.MICROSOFT_CLIENT_ID ? '✓ Set' : '✗ Not set'}</div>
          </div>
        </div>
      </div>
    </div>
  )
}