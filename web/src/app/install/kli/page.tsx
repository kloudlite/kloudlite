import { Metadata } from 'next';

export const metadata: Metadata = {
  title: 'Install kli - Kloudlite Installer CLI',
  description:
    'Download and install the Kloudlite Installer CLI (kli) for your platform',
};

export default function InstallKliPage() {
  const baseUrl =
    process.env.NEXT_PUBLIC_APP_URL || 'https://console.kloudlite.io';

  return (
    <div className="min-h-screen bg-gray-50 py-12 px-4 sm:px-6 lg:px-8">
      <div className="max-w-4xl mx-auto">
        <div className="text-center mb-12">
          <h1 className="text-4xl font-bold text-gray-900 mb-4">
            Install kli
          </h1>
          <p className="text-xl text-gray-600">
            Kloudlite Installer CLI - Multi-cloud Kloudlite installation tool
          </p>
        </div>

        {/* Quick Install */}
        <div className="bg-white shadow rounded-lg p-6 mb-8">
          <h2 className="text-2xl font-semibold text-gray-900 mb-4">
            Quick Install
          </h2>

          {/* Linux */}
          <div className="mb-6">
            <h3 className="text-lg font-medium text-gray-900 mb-2">
              Linux (AMD64)
            </h3>
            <pre className="bg-gray-900 text-gray-100 p-4 rounded overflow-x-auto">
              <code>{`curl -fsSL ${baseUrl}/api/download/kli/linux-amd64 -o kli
chmod +x kli
sudo mv kli /usr/local/bin/kli`}</code>
            </pre>
          </div>

          <div className="mb-6">
            <h3 className="text-lg font-medium text-gray-900 mb-2">
              Linux (ARM64)
            </h3>
            <pre className="bg-gray-900 text-gray-100 p-4 rounded overflow-x-auto">
              <code>{`curl -fsSL ${baseUrl}/api/download/kli/linux-arm64 -o kli
chmod +x kli
sudo mv kli /usr/local/bin/kli`}</code>
            </pre>
          </div>

          {/* macOS */}
          <div className="mb-6">
            <h3 className="text-lg font-medium text-gray-900 mb-2">
              macOS (Intel)
            </h3>
            <pre className="bg-gray-900 text-gray-100 p-4 rounded overflow-x-auto">
              <code>{`curl -fsSL ${baseUrl}/api/download/kli/darwin-amd64 -o kli
chmod +x kli
sudo mv kli /usr/local/bin/kli`}</code>
            </pre>
          </div>

          <div className="mb-6">
            <h3 className="text-lg font-medium text-gray-900 mb-2">
              macOS (Apple Silicon)
            </h3>
            <pre className="bg-gray-900 text-gray-100 p-4 rounded overflow-x-auto">
              <code>{`curl -fsSL ${baseUrl}/api/download/kli/darwin-arm64 -o kli
chmod +x kli
sudo mv kli /usr/local/bin/kli`}</code>
            </pre>
          </div>

          {/* Windows */}
          <div className="mb-6">
            <h3 className="text-lg font-medium text-gray-900 mb-2">
              Windows (PowerShell)
            </h3>
            <pre className="bg-gray-900 text-gray-100 p-4 rounded overflow-x-auto">
              <code>{`# AMD64
Invoke-WebRequest -Uri "${baseUrl}/api/download/kli/windows-amd64" -OutFile "kli.exe"

# ARM64
Invoke-WebRequest -Uri "${baseUrl}/api/download/kli/windows-arm64" -OutFile "kli.exe"`}</code>
            </pre>
            <p className="text-sm text-gray-600 mt-2">
              Add the directory containing kli.exe to your PATH
            </p>
          </div>
        </div>

        {/* Specific Version */}
        <div className="bg-white shadow rounded-lg p-6 mb-8">
          <h2 className="text-2xl font-semibold text-gray-900 mb-4">
            Install Specific Version
          </h2>
          <p className="text-gray-600 mb-4">
            To install a specific version, add the version parameter:
          </p>
          <pre className="bg-gray-900 text-gray-100 p-4 rounded overflow-x-auto">
            <code>{`curl -fsSL ${baseUrl}/api/download/kli/linux-amd64?version=0.1.0 -o kli`}</code>
          </pre>
        </div>

        {/* Direct Downloads */}
        <div className="bg-white shadow rounded-lg p-6 mb-8">
          <h2 className="text-2xl font-semibold text-gray-900 mb-4">
            Direct Downloads
          </h2>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <a
              href={`${baseUrl}/api/download/kli/linux-amd64`}
              className="block p-4 border border-gray-300 rounded hover:bg-gray-50 transition"
            >
              <div className="font-medium text-gray-900">Linux AMD64</div>
              <div className="text-sm text-gray-600">Latest version</div>
            </a>
            <a
              href={`${baseUrl}/api/download/kli/linux-arm64`}
              className="block p-4 border border-gray-300 rounded hover:bg-gray-50 transition"
            >
              <div className="font-medium text-gray-900">Linux ARM64</div>
              <div className="text-sm text-gray-600">Latest version</div>
            </a>
            <a
              href={`${baseUrl}/api/download/kli/darwin-amd64`}
              className="block p-4 border border-gray-300 rounded hover:bg-gray-50 transition"
            >
              <div className="font-medium text-gray-900">macOS Intel</div>
              <div className="text-sm text-gray-600">Latest version</div>
            </a>
            <a
              href={`${baseUrl}/api/download/kli/darwin-arm64`}
              className="block p-4 border border-gray-300 rounded hover:bg-gray-50 transition"
            >
              <div className="font-medium text-gray-900">
                macOS Apple Silicon
              </div>
              <div className="text-sm text-gray-600">Latest version</div>
            </a>
            <a
              href={`${baseUrl}/api/download/kli/windows-amd64`}
              className="block p-4 border border-gray-300 rounded hover:bg-gray-50 transition"
            >
              <div className="font-medium text-gray-900">Windows AMD64</div>
              <div className="text-sm text-gray-600">Latest version</div>
            </a>
            <a
              href={`${baseUrl}/api/download/kli/windows-arm64`}
              className="block p-4 border border-gray-300 rounded hover:bg-gray-50 transition"
            >
              <div className="font-medium text-gray-900">Windows ARM64</div>
              <div className="text-sm text-gray-600">Latest version</div>
            </a>
          </div>
        </div>

        {/* Usage */}
        <div className="bg-white shadow rounded-lg p-6 mb-8">
          <h2 className="text-2xl font-semibold text-gray-900 mb-4">
            Quick Start
          </h2>
          <pre className="bg-gray-900 text-gray-100 p-4 rounded overflow-x-auto mb-4">
            <code>{`# Check version
kli version

# Check AWS prerequisites
kli aws doctor

# Install Kloudlite on AWS
kli aws install --installation-key myenv

# Uninstall
kli aws uninstall --installation-key myenv`}</code>
          </pre>
          <p className="text-gray-600">
            For full documentation, visit{' '}
            <a
              href="https://github.com/kloudlite/kloudlite/tree/development/api/cmd/kli"
              className="text-blue-600 hover:underline"
              target="_blank"
              rel="noopener noreferrer"
            >
              GitHub
            </a>
          </p>
        </div>

        {/* All Releases */}
        <div className="bg-white shadow rounded-lg p-6">
          <h2 className="text-2xl font-semibold text-gray-900 mb-4">
            All Releases
          </h2>
          <p className="text-gray-600 mb-4">
            View all releases and changelog on GitHub:
          </p>
          <a
            href="https://github.com/kloudlite/kloudlite/releases?q=kli-v&expanded=true"
            className="inline-flex items-center px-4 py-2 border border-transparent text-base font-medium rounded-md text-white bg-blue-600 hover:bg-blue-700"
            target="_blank"
            rel="noopener noreferrer"
          >
            View Releases on GitHub
          </a>
        </div>
      </div>
    </div>
  );
}
