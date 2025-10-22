import { KloudliteLogo } from '@/components/kloudlite-logo'
import { OAuthButtons } from '@/components/registration/oauth-buttons'

export default function LoginPage() {
  return (
    <div className="min-h-screen grid lg:grid-cols-2">
      {/* Left side - Branding */}
      <div className="hidden lg:flex bg-gray-900 text-white p-12 flex-col justify-between overflow-hidden">
        <div className="py-4">
          <KloudliteLogo variant="white" className="scale-150 origin-left" />
        </div>
        <div>
          <h2 className="text-2xl font-light mb-4">Cloud Development Environments</h2>
          <p className="text-sm text-gray-400 leading-relaxed max-w-md">
            Designed to reduce the development loop
          </p>
        </div>
      </div>

      {/* Right side - Form */}
      <div className="flex items-center justify-center p-8 bg-background">
        <div className="w-full max-w-sm">
          {/* Mobile logo */}
          <div className="lg:hidden mb-8">
            <KloudliteLogo />
          </div>

          <div className="mb-8">
            <h1 className="text-lg font-medium text-foreground">
              Reserve your domain
            </h1>
            <p className="mt-1 text-sm text-muted-foreground">
              Secure your workspace domain and get started with Kloudlite
            </p>
          </div>

          <div className="space-y-6">
            <div className="space-y-4">
              <p className="text-sm text-muted-foreground text-center">
                Sign in with your preferred provider
              </p>

              <OAuthButtons />
            </div>

            <p className="text-sm text-center text-muted-foreground">
              Already have an account?{' '}
              <a href="https://app.kloudlite.io/signin" className="text-primary hover:underline font-medium">
                Sign in
              </a>
            </p>
          </div>
        </div>
      </div>
    </div>
  )
}
