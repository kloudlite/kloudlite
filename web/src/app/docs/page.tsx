import { Metadata } from 'next'
import { Book, Search, FileText, Code, Users, Settings, ArrowRight, Zap, ShieldCheck, ExternalLink } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Link } from '@/components/ui/link'

export const metadata: Metadata = {
  title: 'Documentation - Kloudlite',
  description: 'Kloudlite platform documentation and guides',
}

const documentationSections = [
  {
    title: 'Getting Started',
    description: 'Quick start guide and basic concepts',
    icon: Book,
    href: '/docs/getting-started',
    articles: ['Introduction', 'Quick Start', 'Installation'],
  },
  {
    title: 'Authentication',
    description: 'OAuth setup and session management',
    icon: Users,
    href: '/docs/authentication',
    articles: ['OAuth Setup', 'Providers', 'Session Management'],
  },
  {
    title: 'API Reference',
    description: 'Complete API documentation',
    icon: Code,
    href: '/docs/api',
    articles: ['Authentication API', 'Teams API', 'Resources API'],
  },
  {
    title: 'Architecture',
    description: 'System architecture and components',
    icon: Settings,
    href: '/docs/architecture',
    articles: ['Overview', 'Microservices', 'Infrastructure'],
  }
]

export default function DocsHomePage() {
  return (
    <div className="min-h-screen bg-background">
      <div className="max-w-7xl mx-auto px-6 py-12">
        {/* Hero Section */}
        <div className="text-center mb-16">
          <div className="inline-flex items-center gap-2 px-4 py-2 bg-primary/10 text-primary rounded-none text-sm font-medium mb-6">
            <Zap className="h-4 w-4" />
            Developer Documentation
          </div>
          <h1 className="text-4xl font-bold mb-6">
            Kloudlite Documentation
          </h1>
          <p className="text-muted-foreground mb-8 max-w-2xl mx-auto">
            Everything you need to build, deploy, and scale applications with Kloudlite's cloud-native development platform
          </p>
        </div>

        {/* Documentation Sections */}
        <div className="grid gap-6 md:grid-cols-2 mb-16">
          {documentationSections.map((section) => (
            <div 
              key={section.title} 
              className="group relative border border-border rounded-none p-8 hover:shadow-sm transition-all duration-200 bg-card"
            >
              <div className="flex items-start gap-4">
                <div className="flex-shrink-0">
                  <div className="p-3 bg-muted rounded-none">
                    <section.icon className="h-6 w-6 text-primary" />
                  </div>
                </div>
                <div className="flex-1">
                  <h3 className="font-semibold text-lg mb-3">{section.title}</h3>
                  <p className="text-muted-foreground mb-4">{section.description}</p>
                  <ul className="space-y-2 mb-6">
                    {section.articles.map((article) => (
                      <li key={article} className="flex items-center gap-2 text-sm text-muted-foreground">
                        <div className="w-1.5 h-1.5 bg-primary/60 rounded-full" />
                        {article}
                      </li>
                    ))}
                  </ul>
                  <Button asChild variant="outline" size="sm" className="rounded-none">
                    <Link href={section.href}>
                      View Section
                      <ArrowRight className="ml-2 h-4 w-4" />
                    </Link>
                  </Button>
                </div>
              </div>
            </div>
          ))}
        </div>

        {/* Quick Actions */}
        <div className="grid md:grid-cols-3 gap-6 mb-16">
          <div className="bg-card border border-border rounded-none p-6">
            <Book className="h-6 w-6 text-primary mb-4" />
            <h3 className="font-semibold mb-2">Quick Start</h3>
            <p className="text-sm text-muted-foreground mb-4">Get up and running in minutes</p>
            <Button asChild variant="outline" size="sm" className="rounded-none">
              <Link href="/docs/getting-started">
                Start Building
              </Link>
            </Button>
          </div>
          
          <div className="bg-card border border-border rounded-none p-6">
            <Code className="h-6 w-6 text-primary mb-4" />
            <h3 className="font-semibold mb-2">API Reference</h3>
            <p className="text-sm text-muted-foreground mb-4">Complete API documentation</p>
            <Button asChild variant="outline" size="sm" className="rounded-none">
              <Link href="/docs/api">
                Browse APIs
              </Link>
            </Button>
          </div>
          
          <div className="bg-card border border-border rounded-none p-6">
            <ShieldCheck className="h-6 w-6 text-primary mb-4" />
            <h3 className="font-semibold mb-2">Security</h3>
            <p className="text-sm text-muted-foreground mb-4">Authentication & authorization</p>
            <Button asChild variant="outline" size="sm" className="rounded-none">
              <Link href="/docs/authentication">
                Learn More
              </Link>
            </Button>
          </div>
        </div>

        {/* Footer Links */}
        <div className="border-t border-border pt-8">
          <div className="flex flex-col md:flex-row items-center justify-between gap-4">
            <div className="text-sm text-muted-foreground">
              Need help? Check out our community resources
            </div>
            <div className="flex items-center gap-4">
              <Button asChild variant="ghost" size="sm" className="rounded-none">
                <Link href="https://github.com/kloudlite/kloudlite">
                  <ExternalLink className="h-4 w-4 mr-2" />
                  GitHub
                </Link>
              </Button>
              <Button asChild variant="ghost" size="sm" className="rounded-none">
                <Link href="/">
                  <ArrowRight className="h-4 w-4 mr-2 rotate-180" />
                  Back to Home
                </Link>
              </Button>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}