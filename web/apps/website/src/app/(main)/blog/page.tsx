import { ScrollArea } from '@kloudlite/ui'
import { WebsiteHeader } from '@/components/website-header'
import { WebsiteFooter } from '@/components/website-footer'
import { cn } from '@kloudlite/lib'
import Link from 'next/link'
import { ArrowRight, Calendar, Clock } from 'lucide-react'

function CrossMarker({ className }: { className?: string }) {
  return (
    <div className={cn('absolute', className)}>
      <div className="absolute left-1/2 top-0 -translate-x-1/2 w-px h-5 bg-foreground/20" />
      <div className="absolute top-1/2 left-0 -translate-y-1/2 h-px w-5 bg-foreground/20" />
    </div>
  )
}

function GridContainer({ children, className }: { children: React.ReactNode; className?: string }) {
  return (
    <div className={cn('relative mx-auto max-w-5xl', className)}>
      <div className="absolute inset-0 pointer-events-none overflow-visible">
        <div className="absolute inset-y-0 left-0 w-px bg-foreground/10" />
        <div className="absolute inset-y-0 right-0 w-px bg-foreground/10" />
        <div className="absolute inset-x-0 top-0 h-px bg-foreground/10" />
        <div className="absolute inset-x-0 bottom-0 h-px bg-foreground/10" />
        <CrossMarker className="top-0 left-0 -translate-x-1/2 -translate-y-1/2 w-5 h-5" />
        <CrossMarker className="top-0 right-0 translate-x-1/2 -translate-y-1/2 w-5 h-5" />
        <CrossMarker className="bottom-0 left-0 -translate-x-1/2 translate-y-1/2 w-5 h-5" />
        <CrossMarker className="bottom-0 right-0 translate-x-1/2 translate-y-1/2 w-5 h-5" />
      </div>
      <div className="relative">{children}</div>
    </div>
  )
}

// Sample blog posts data
const blogPosts = [
  {
    slug: 'introducing-kloudlite',
    title: 'Introducing Kloudlite: Cloud Development Environments',
    excerpt: 'Learn how Kloudlite is revolutionizing the way developers build and test applications with cloud-native development environments.',
    date: '2024-01-15',
    readTime: '5 min read',
    category: 'Product',
    featured: true,
  },
  {
    slug: 'getting-started-with-workspaces',
    title: 'Getting Started with Kloudlite Workspaces',
    excerpt: 'A comprehensive guide to setting up your first workspace and connecting to your development environment.',
    date: '2024-01-10',
    readTime: '8 min read',
    category: 'Tutorial',
    featured: false,
  },
  {
    slug: 'environment-forking-explained',
    title: 'Environment Forking: Test Changes Without Breaking Production',
    excerpt: 'Discover how environment forking enables safe testing and experimentation without impacting your production services.',
    date: '2024-01-05',
    readTime: '6 min read',
    category: 'Feature',
    featured: false,
  },
  {
    slug: 'service-intercepts-deep-dive',
    title: 'Service Intercepts: Route Production Traffic to Your Workspace',
    excerpt: 'Learn how service intercepts allow you to debug production issues by routing live traffic to your local development environment.',
    date: '2023-12-28',
    readTime: '7 min read',
    category: 'Technical',
    featured: false,
  },
  {
    slug: 'nix-package-management',
    title: 'Why We Chose Nix for Package Management',
    excerpt: 'Understanding the benefits of Nix-based package management for reproducible development environments.',
    date: '2023-12-20',
    readTime: '10 min read',
    category: 'Architecture',
    featured: false,
  },
  {
    slug: 'kubernetes-native-development',
    title: 'Kubernetes-Native Development Made Simple',
    excerpt: 'How Kloudlite abstracts away Kubernetes complexity while giving you the power of cloud-native infrastructure.',
    date: '2023-12-15',
    readTime: '9 min read',
    category: 'Technical',
    featured: false,
  },
]

function FeaturedBlogCard({ post }: { post: typeof blogPosts[0] }) {
  return (
    <Link
      href={`/blog/${post.slug}`}
      className="group block border-b border-foreground/10"
    >
      <div className="p-12 lg:p-16 space-y-6">
        <div className="flex items-center gap-4">
          <span className="inline-flex items-center px-3 py-1 bg-primary/10 text-primary text-xs font-semibold uppercase tracking-wider rounded-none">
            Featured
          </span>
          <span className="text-muted-foreground text-sm">
            {post.category}
          </span>
        </div>

        <h2 className="text-3xl sm:text-4xl lg:text-5xl font-bold text-foreground leading-tight group-hover:text-primary transition-colors">
          {post.title}
        </h2>

        <p className="text-muted-foreground text-lg leading-relaxed max-w-3xl">
          {post.excerpt}
        </p>

        <div className="flex items-center gap-6 text-muted-foreground text-sm pt-4">
          <div className="flex items-center gap-2">
            <Calendar className="h-4 w-4" />
            <time dateTime={post.date}>
              {new Date(post.date).toLocaleDateString('en-US', {
                month: 'long',
                day: 'numeric',
                year: 'numeric'
              })}
            </time>
          </div>
          <div className="flex items-center gap-2">
            <Clock className="h-4 w-4" />
            <span>{post.readTime}</span>
          </div>
          <div className="flex items-center gap-2 text-primary font-medium ml-auto">
            <span>Read article</span>
            <ArrowRight className="h-4 w-4 transition-transform group-hover:translate-x-1" />
          </div>
        </div>
      </div>
    </Link>
  )
}

function BlogCard({ post }: { post: typeof blogPosts[0] }) {
  return (
    <Link
      href={`/blog/${post.slug}`}
      className="group block border-b border-foreground/10 sm:border-r sm:odd:border-r sm:even:border-r-0 transition-colors hover:bg-foreground/[0.015]"
    >
      <div className="p-8 lg:p-10 h-full flex flex-col">
        <div className="mb-4">
          <span className="inline-block px-2.5 py-1 bg-muted text-muted-foreground text-xs font-medium uppercase tracking-wider rounded">
            {post.category}
          </span>
        </div>

        <h3 className="text-foreground text-xl font-bold mb-3 leading-tight group-hover:text-primary transition-colors">
          {post.title}
        </h3>

        <p className="text-muted-foreground text-base leading-relaxed mb-6 flex-1">
          {post.excerpt}
        </p>

        <div className="flex items-center gap-4 text-muted-foreground text-xs pt-4 border-t border-foreground/10">
          <time dateTime={post.date}>
            {new Date(post.date).toLocaleDateString('en-US', {
              month: 'short',
              day: 'numeric',
              year: 'numeric'
            })}
          </time>
          <span>•</span>
          <span>{post.readTime}</span>
          <ArrowRight className="h-4 w-4 text-primary transition-transform group-hover:translate-x-1 ml-auto" />
        </div>
      </div>
    </Link>
  )
}

export default function BlogPage() {
  const featuredPost = blogPosts.find(post => post.featured)
  const regularPosts = blogPosts.filter(post => !post.featured)

  return (
    <div className="bg-background h-screen">
      <ScrollArea className="h-full">
        <WebsiteHeader currentPage="blog" />
        <main>
          <div className="px-6 pt-8 lg:px-8 lg:pt-12">
            <GridContainer className="px-6 lg:px-12">
              {/* Hero Section */}
              <div className="py-20 lg:py-28">
                <div className="text-center">
                  <h1 className="text-[2.5rem] font-semibold leading-[1.1] tracking-[-0.02em] sm:text-5xl md:text-6xl lg:text-[4rem]">
                    <span className="text-foreground">Engineering Insights</span>
                  </h1>
                  <p className="text-muted-foreground mx-auto mt-6 max-w-lg text-lg leading-relaxed">
                    Guides, updates, and best practices from the Kloudlite team.
                  </p>
                </div>
              </div>

              {/* Featured Post */}
              {featuredPost && (
                <div className="-mx-6 lg:-mx-12">
                  <FeaturedBlogCard post={featuredPost} />
                </div>
              )}

              {/* Regular Posts Grid */}
              <div className="grid sm:grid-cols-2 border-t border-foreground/10 -mx-6 lg:-mx-12">
                {regularPosts.map((post) => (
                  <BlogCard key={post.slug} post={post} />
                ))}
              </div>
            </GridContainer>
          </div>

          <WebsiteFooter />
        </main>
      </ScrollArea>
    </div>
  )
}
