import { DocsPage } from '@/components/docs/docs-page'

export default function InstallationPage() {
  return (
    <DocsPage
      title="Installation"
      description="Learn how to install and set up Kloudlite for your development environment"
      lastUpdated="January 16, 2025"
      difficulty="beginner"
      estimatedTime={15}
      tags={['installation', 'setup', 'getting-started']}
    >
      <div className="space-y-12">
        <section>
          <h2>System Requirements</h2>
          <p>
            Before installing Kloudlite, ensure your system meets these requirements:
          </p>
          
          <div className="mt-6">
            <h3>Minimum Requirements</h3>
            <ul className="space-y-2 ml-6 mt-4">
              <li className="flex items-start gap-2">
                <span className="text-muted-foreground">•</span>
                <span><strong>Operating System</strong>: macOS 10.15+, Windows 10+, or Linux (Ubuntu 18.04+)</span>
              </li>
              <li className="flex items-start gap-2">
                <span className="text-muted-foreground">•</span>
                <span><strong>Memory</strong>: 8GB RAM minimum (16GB recommended)</span>
              </li>
              <li className="flex items-start gap-2">
                <span className="text-muted-foreground">•</span>
                <span><strong>Storage</strong>: 10GB available disk space</span>
              </li>
              <li className="flex items-start gap-2">
                <span className="text-muted-foreground">•</span>
                <span><strong>Internet</strong>: Stable broadband connection</span>
              </li>
            </ul>
          </div>
          
          <div className="mt-6">
            <h3>Required Software</h3>
            <p className="mt-3 mb-4">The following software must be installed on your system:</p>
            <ul className="space-y-2 ml-6">
              <li className="flex items-start gap-2">
                <span className="text-muted-foreground">•</span>
                <span><strong>Git</strong> (version 2.20 or higher)</span>
              </li>
              <li className="flex items-start gap-2">
                <span className="text-muted-foreground">•</span>
                <span><strong>Docker</strong> (version 20.10 or higher)</span>
              </li>
              <li className="flex items-start gap-2">
                <span className="text-muted-foreground">•</span>
                <span><strong>Node.js</strong> (version 18.0 or higher)</span>
              </li>
              <li className="flex items-start gap-2">
                <span className="text-muted-foreground">•</span>
                <span><strong>kubectl</strong> (version 1.24 or higher)</span>
              </li>
            </ul>
          </div>
        </section>

        <section>
          <h2>Installation Methods</h2>
          <p>
            Kloudlite can be installed using several methods. Choose the one that best fits your environment:
          </p>
          
          <div className="mt-8 space-y-8">
            <div>
              <h3>Method 1: Using Homebrew (macOS/Linux)</h3>
              <p className="mt-3 mb-4">
                The easiest way to install Kloudlite on macOS or Linux is through Homebrew:
              </p>
              <pre className="bg-muted rounded-lg p-4 overflow-x-auto">
                <code>{`# Add the Kloudlite tap
brew tap kloudlite/tap

# Install Kloudlite CLI
brew install kloudlite

# Verify installation
kloudlite version`}</code>
              </pre>
            </div>
            
            <div>
              <h3>Method 2: Using npm</h3>
              <p className="mt-3 mb-4">
                If you have Node.js installed, you can install Kloudlite globally using npm:
              </p>
              <pre className="bg-muted rounded-lg p-4 overflow-x-auto">
                <code>{`# Install globally
npm install -g @kloudlite/cli

# Verify installation
kloudlite version`}</code>
              </pre>
            </div>
            
            <div>
              <h3>Method 3: Direct Download</h3>
              <p className="mt-3 mb-4">
                You can download the binary directly from our releases page:
              </p>
              <ol className="space-y-2 ml-6 list-decimal">
                <li>Visit <a href="https://github.com/kloudlite/kloudlite/releases" className="text-primary hover:text-primary-hover underline">github.com/kloudlite/kloudlite/releases</a></li>
                <li>Download the appropriate binary for your platform</li>
                <li>Extract the archive</li>
                <li>Move the binary to your PATH</li>
              </ol>
              <pre className="bg-muted rounded-lg p-4 overflow-x-auto mt-4">
                <code>{`# Example for macOS/Linux
tar -xzf kloudlite-darwin-amd64.tar.gz
sudo mv kloudlite /usr/local/bin/
chmod +x /usr/local/bin/kloudlite`}</code>
              </pre>
            </div>
          </div>
        </section>

        <section>
          <h2>Post-Installation Setup</h2>
          <p>
            After installing Kloudlite, complete the setup process:
          </p>
          
          <div className="mt-8 space-y-8">
            <div>
              <h3>1. Initialize Configuration</h3>
              <p className="mt-3 mb-4">
                Run the initialization command to set up your local configuration:
              </p>
              <pre className="bg-muted rounded-lg p-4 overflow-x-auto">
                <code>kloudlite init</code>
              </pre>
              <p className="mt-3 text-sm text-muted-foreground">
                This will create a configuration file at <code className="bg-muted px-1 py-0.5 rounded">~/.kloudlite/config.yaml</code>
              </p>
            </div>
            
            <div>
              <h3>2. Authenticate</h3>
              <p className="mt-3 mb-4">
                Connect your Kloudlite CLI to your account:
              </p>
              <pre className="bg-muted rounded-lg p-4 overflow-x-auto">
                <code>kloudlite auth login</code>
              </pre>
              <p className="mt-3 text-sm text-muted-foreground">
                This will open your browser for authentication. After successful login, your credentials will be stored securely.
              </p>
            </div>
            
            <div>
              <h3>3. Verify Installation</h3>
              <p className="mt-3 mb-4">
                Confirm everything is working correctly:
              </p>
              <pre className="bg-muted rounded-lg p-4 overflow-x-auto">
                <code>{`# Check CLI version
kloudlite version

# Verify authentication
kloudlite auth status

# List available commands
kloudlite help`}</code>
              </pre>
            </div>
          </div>
        </section>

        <section>
          <h2>Configuration Options</h2>
          <p>
            Kloudlite can be configured through environment variables or the config file:
          </p>
          
          <div className="mt-6">
            <h3>Environment Variables</h3>
            <div className="mt-4 overflow-x-auto">
              <table className="w-full border border-border">
                <thead>
                  <tr className="border-b border-border bg-muted/50">
                    <th className="text-left p-4 font-semibold">Variable</th>
                    <th className="text-left p-4 font-semibold">Description</th>
                    <th className="text-left p-4 font-semibold">Default</th>
                  </tr>
                </thead>
                <tbody>
                  <tr className="border-b border-border/50">
                    <td className="p-4 font-mono text-sm">KLOUDLITE_CONFIG_PATH</td>
                    <td className="p-4">Path to config file</td>
                    <td className="p-4 font-mono text-sm">~/.kloudlite/config.yaml</td>
                  </tr>
                  <tr className="border-b border-border/50">
                    <td className="p-4 font-mono text-sm">KLOUDLITE_API_URL</td>
                    <td className="p-4">API endpoint URL</td>
                    <td className="p-4 font-mono text-sm">https://api.kloudlite.io</td>
                  </tr>
                  <tr className="border-b border-border/50">
                    <td className="p-4 font-mono text-sm">KLOUDLITE_LOG_LEVEL</td>
                    <td className="p-4">Logging level</td>
                    <td className="p-4 font-mono text-sm">info</td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>
          
          <div className="mt-8">
            <h3>Config File</h3>
            <p className="mt-3 mb-4">The default configuration file structure:</p>
            <pre className="bg-muted rounded-lg p-4 overflow-x-auto">
              <code>{`# ~/.kloudlite/config.yaml
api_url: https://api.kloudlite.io
auth:
  token: <your-auth-token>
defaults:
  region: us-west-2
  project: default
logging:
  level: info
  format: json`}</code>
            </pre>
          </div>
        </section>

        <section>
          <h2>Troubleshooting</h2>
          
          <div className="mt-6">
            <h3>Common Issues</h3>
            
            <div className="mt-6 space-y-6">
              <div>
                <h4 className="font-semibold">Issue: Command not found</h4>
                <p className="mt-2 mb-3">
                  If you get a "command not found" error, ensure the binary is in your PATH:
                </p>
                <pre className="bg-muted rounded-lg p-4 overflow-x-auto">
                  <code>{`echo $PATH
which kloudlite`}</code>
                </pre>
              </div>
              
              <div>
                <h4 className="font-semibold">Issue: Permission denied</h4>
                <p className="mt-2 mb-3">
                  On Linux/macOS, you may need to make the binary executable:
                </p>
                <pre className="bg-muted rounded-lg p-4 overflow-x-auto">
                  <code>chmod +x /usr/local/bin/kloudlite</code>
                </pre>
              </div>
              
              <div>
                <h4 className="font-semibold">Issue: Docker not running</h4>
                <p className="mt-2 mb-3">
                  Kloudlite requires Docker to be running. Start Docker Desktop or the Docker daemon:
                </p>
                <pre className="bg-muted rounded-lg p-4 overflow-x-auto">
                  <code>{`# macOS
open -a Docker

# Linux
sudo systemctl start docker`}</code>
                </pre>
              </div>
            </div>
          </div>
          
          <div className="mt-8">
            <h3>Getting Help</h3>
            <p className="mt-3 mb-4">If you encounter issues during installation:</p>
            <ol className="space-y-2 ml-6 list-decimal">
              <li>Check our <a href="/docs/troubleshooting" className="text-primary hover:text-primary-hover underline">troubleshooting guide</a></li>
              <li>Visit our <a href="https://github.com/kloudlite/kloudlite/issues" className="text-primary hover:text-primary-hover underline">GitHub issues</a></li>
              <li>Join our <a href="https://discord.gg/kloudlite" className="text-primary hover:text-primary-hover underline">Discord community</a></li>
            </ol>
          </div>
        </section>

        <section>
          <h2>Next Steps</h2>
          <p>
            Now that you have Kloudlite installed, you're ready to:
          </p>
          
          <ul className="mt-4 space-y-2 ml-6">
            <li className="flex items-start gap-2">
              <span className="text-muted-foreground">•</span>
              <a href="/docs/getting-started/quickstart" className="text-primary hover:text-primary-hover underline">Create your first project</a>
            </li>
            <li className="flex items-start gap-2">
              <span className="text-muted-foreground">•</span>
              <a href="/docs/development/local" className="text-primary hover:text-primary-hover underline">Set up your development environment</a>
            </li>
            <li className="flex items-start gap-2">
              <span className="text-muted-foreground">•</span>
              <a href="/docs/concepts" className="text-primary hover:text-primary-hover underline">Learn about core concepts</a>
            </li>
          </ul>
        </section>
      </div>
    </DocsPage>
  )
}