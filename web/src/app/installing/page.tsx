export default function InstallingPage() {
  return (
    <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-surface-100 to-surface-200 dark:from-surface-900 dark:to-surface-950">
      <div className="max-w-2xl mx-auto p-8 text-center">
        {/* Logo/Brand */}
        <div className="mb-8">
          <h1 className="text-6xl font-bold bg-gradient-to-r from-primary-500 to-primary-700 bg-clip-text text-transparent">
            Kloudlite
          </h1>
          <p className="text-xl text-surface-600 dark:text-surface-400 mt-2">
            Development Environments Platform
          </p>
        </div>

        {/* Installation Progress */}
        <div className="bg-surface-50 dark:bg-surface-900 rounded-2xl p-8 shadow-xl border border-surface-200 dark:border-surface-800">
          {/* Animated Spinner */}
          <div className="flex justify-center mb-6">
            <div className="relative">
              <div className="w-16 h-16 border-4 border-surface-300 dark:border-surface-700 rounded-full"></div>
              <div className="w-16 h-16 border-4 border-primary-500 border-t-transparent rounded-full animate-spin absolute top-0 left-0"></div>
            </div>
          </div>

          {/* Status Text */}
          <h2 className="text-2xl font-semibold text-surface-900 dark:text-surface-100 mb-3">
            Installation in Progress
          </h2>
          <p className="text-surface-600 dark:text-surface-400 mb-6">
            We're setting up your Kloudlite instance. This typically takes 5-10 minutes.
          </p>

          {/* Progress Steps */}
          <div className="space-y-3 text-left max-w-md mx-auto">
            <div className="flex items-center space-x-3">
              <div className="w-2 h-2 rounded-full bg-green-500 animate-pulse"></div>
              <span className="text-sm text-surface-700 dark:text-surface-300">
                Installing Kubernetes cluster (K3s)
              </span>
            </div>
            <div className="flex items-center space-x-3">
              <div className="w-2 h-2 rounded-full bg-green-500 animate-pulse"></div>
              <span className="text-sm text-surface-700 dark:text-surface-300">
                Deploying API server
              </span>
            </div>
            <div className="flex items-center space-x-3">
              <div className="w-2 h-2 rounded-full bg-green-500 animate-pulse"></div>
              <span className="text-sm text-surface-700 dark:text-surface-300">
                Configuring networking
              </span>
            </div>
            <div className="flex items-center space-x-3">
              <div className="w-2 h-2 rounded-full bg-surface-400 dark:bg-surface-600"></div>
              <span className="text-sm text-surface-500 dark:text-surface-500">
                Setting up web console
              </span>
            </div>
          </div>

          {/* Additional Info */}
          <div className="mt-8 p-4 bg-primary-50 dark:bg-primary-950/30 rounded-lg border border-primary-200 dark:border-primary-800">
            <p className="text-sm text-primary-700 dark:text-primary-300">
              <strong>Note:</strong> This page will automatically redirect once installation is complete.
            </p>
          </div>
        </div>

        {/* Footer */}
        <div className="mt-8 text-sm text-surface-500 dark:text-surface-500">
          <p>
            Need help?{' '}
            <a
              href="https://docs.kloudlite.io"
              className="text-primary-600 dark:text-primary-400 hover:underline"
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
