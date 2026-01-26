export default function InstallingPage() {
  return (
    <div className="bg-background min-h-screen flex items-center justify-center">
      <div className="max-w-2xl mx-auto p-8 text-center">
        {/* Logo/Brand */}
        <div className="mb-12">
          <h1 className="text-6xl font-bold text-primary">
            Kloudlite
          </h1>
          <p className="text-muted-foreground text-xl mt-2">
            Development Environments Platform
          </p>
        </div>

        {/* Installation Progress */}
        <div className="bg-card border p-10">
          {/* Animated Spinner */}
          <div className="flex justify-center mb-6">
            <div className="relative">
              <div className="w-16 h-16 border-4 border-muted"></div>
              <div className="w-16 h-16 border-4 border-primary border-t-transparent animate-spin absolute top-0 left-0"></div>
            </div>
          </div>

          {/* Status Text */}
          <h2 className="text-foreground text-2xl font-semibold mb-3">
            Installation in Progress
          </h2>
          <p className="text-muted-foreground mb-8 text-base">
            We&apos;re setting up your Kloudlite instance. This typically takes 5-10 minutes.
          </p>

          {/* Progress Steps */}
          <div className="space-y-4 text-left max-w-md mx-auto">
            <div className="flex items-center space-x-3">
              <div className="w-2 h-2 bg-success animate-pulse"></div>
              <span className="text-foreground text-base">
                Installing Kubernetes cluster (K3s)
              </span>
            </div>
            <div className="flex items-center space-x-3">
              <div className="w-2 h-2 bg-success animate-pulse"></div>
              <span className="text-foreground text-base">
                Deploying API server
              </span>
            </div>
            <div className="flex items-center space-x-3">
              <div className="w-2 h-2 bg-success animate-pulse"></div>
              <span className="text-foreground text-base">
                Configuring networking
              </span>
            </div>
            <div className="flex items-center space-x-3">
              <div className="w-2 h-2 bg-muted-foreground"></div>
              <span className="text-muted-foreground text-base">
                Setting up web console
              </span>
            </div>
          </div>

          {/* Additional Info */}
          <div className="mt-8 p-4 bg-primary/10 border border-primary">
            <p className="text-foreground text-base">
              <strong>Note:</strong> This page will automatically redirect once installation is complete.
            </p>
          </div>
        </div>

        {/* Footer */}
        <div className="mt-8 text-base text-muted-foreground">
          <p>
            Need help?{' '}
            <a
              href="https://docs.kloudlite.io"
              className="text-primary hover:underline"
              target="_blank"
              rel="noopener noreferrer"
            >
              Visit our documentation
            </a>
          </p>
        </div>
      </div>
    </div>
  )
}
