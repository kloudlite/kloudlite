import { KloudliteLogo } from '@/components/kloudlite-logo'
import { OAuthButtons } from '@/components/registration/oauth-buttons'

export default function LoginPage() {
  return (
    <div className="grid min-h-screen lg:grid-cols-2">
      {/* Left side - Branding */}
      <div className="hidden flex-col justify-between overflow-hidden bg-gray-900 p-12 text-white lg:flex">
        <div className="py-4">
          <KloudliteLogo variant="white" className="origin-left scale-150" />
        </div>
        <div>
          <h2 className="mb-4 text-2xl font-light">Cloud Development Environments</h2>
          <p className="max-w-md text-sm leading-relaxed text-muted-foreground">
            Designed to reduce the development loop
          </p>
        </div>
      </div>

      {/* Right side - Form */}
      <div className="bg-background flex items-center justify-center p-8">
        <div className="w-full max-w-sm">
          {/* Mobile logo */}
          <div className="mb-8 lg:hidden">
            <KloudliteLogo />
          </div>

          <div className="mb-8">
            <h1 className="text-foreground text-lg font-medium">Reserve your domain</h1>
            <p className="text-muted-foreground mt-1 text-sm">
              Secure your workspace domain and get started with Kloudlite
            </p>
          </div>

          <div className="space-y-6">
            <div className="space-y-4">
              <p className="text-muted-foreground text-center text-sm">
                Sign in with your preferred provider
              </p>

              <OAuthButtons />
            </div>

            <p className="text-muted-foreground text-center text-sm">
              Already have an account?{' '}
              <a
                href="https://app.kloudlite.io/signin"
                className="text-primary font-medium hover:underline"
              >
                Sign in
              </a>
            </p>
          </div>
        </div>
      </div>
    </div>
  )
}
