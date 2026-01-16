import { ScrollArea } from '@kloudlite/ui'
import { WebsiteHeader } from '@/components/website-header'
import { WebsiteFooter } from '@/components/website-footer'
import { SocialButtons } from '@/components/blog-social-actions'
import { cn } from '@kloudlite/lib'
import Link from 'next/link'
import { ArrowLeft } from 'lucide-react'
import { notFound } from 'next/navigation'

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
    <div className={cn('relative mx-auto max-w-7xl', className)}>
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

// Sample blog posts data (in a real app, this would come from a CMS or markdown files)
const blogPosts = {
  'introducing-kloudlite': {
    slug: 'introducing-kloudlite',
    title: 'Introducing Kloudlite: Cloud Development Environments',
    excerpt: 'Learn how Kloudlite is revolutionizing the way developers build and test applications with cloud-native development environments.',
    date: '2024-01-15',
    readTime: '5 min read',
    category: 'Product',
    author: {
      name: 'Kloudlite Team',
      role: 'Engineering',
      avatar: 'KT',
      bio: 'Building the future of cloud development environments.'
    },
    content: `
# Introduction

Kloudlite is transforming how developers build and test applications by providing cloud-native development environments that eliminate setup time and configuration headaches.

## The Problem with Traditional Development

Traditional development workflows require developers to:
- Clone repositories locally
- Install dependencies
- Set up databases and services
- Configure environment variables
- Mock external services
- Deploy to test environments
- Wait for CI/CD pipelines

This process can take hours or even days, especially for complex microservices applications.

## The Kloudlite Solution

With Kloudlite, you can:

1. **Instant Workspaces**: Start coding in seconds with pre-configured cloud environments
2. **Real Service Connections**: Connect to actual databases and APIs, not mocks
3. **Environment Forking**: Create isolated copies of entire environments for testing
4. **Service Intercepts**: Route production traffic to your workspace for debugging

## Key Features

### Workspace Management
Kloudlite workspaces are fully-featured development environments that run in the cloud. Each workspace includes:
- VS Code Server for in-browser development
- Full terminal access via SSH or web
- Nix-based package management for reproducible environments
- Persistent storage for your code and data

### Environment Forking
Fork entire environments with a single command, including all services, databases, and configurations. Test changes in isolation without affecting other developers.

### Service Intercepts
Route live traffic from production services to your workspace. Debug production issues with real data without deploying changes.

## Getting Started

Getting started with Kloudlite is simple:

\`\`\`bash
# Install Kloudlite CLI
curl -fsSL https://get.kloudlite.io | bash

# Create your first workspace
kl workspace create my-workspace

# Connect to an environment
kl environment connect dev
\`\`\`

## What's Next?

In upcoming blog posts, we'll dive deeper into:
- Setting up your first workspace
- Using environment forking effectively
- Debugging with service intercepts
- Best practices for team collaboration

Stay tuned, and happy coding!
    `
  },
  'getting-started-with-workspaces': {
    slug: 'getting-started-with-workspaces',
    title: 'Getting Started with Kloudlite Workspaces',
    excerpt: 'A comprehensive guide to setting up your first workspace and connecting to your development environment.',
    date: '2024-01-10',
    readTime: '8 min read',
    category: 'Tutorial',
    author: {
      name: 'Kloudlite Team',
      role: 'Engineering',
      avatar: 'KT',
      bio: 'Building the future of cloud development environments.'
    },
    content: `
# Getting Started with Kloudlite Workspaces

This guide will walk you through creating and using your first Kloudlite workspace.

## Prerequisites

Before getting started, make sure you have:
- A Kloudlite account (sign up at kloudlite.io)
- The Kloudlite CLI installed
- Access to at least one environment

## Creating Your First Workspace

Creating a workspace is straightforward:

\`\`\`bash
kl workspace create my-first-workspace --image comprehensive
\`\`\`

This creates a new workspace with the comprehensive image, which includes:
- Node.js, Python, Go, and other common languages
- VS Code Server
- Git and common development tools
- Nix package manager

## Accessing Your Workspace

You have multiple ways to access your workspace:

### 1. VS Code in Browser
The easiest way is through the web UI at dashboard.kloudlite.io

### 2. SSH Access
\`\`\`bash
kl workspace connect my-first-workspace
\`\`\`

### 3. VS Code Remote
Add your workspace as an SSH remote in VS Code desktop.

## Installing Packages

Kloudlite uses Nix for package management, ensuring reproducible environments:

\`\`\`bash
# Install packages
kl pkg add nodejs python3 postgresql

# List installed packages
kl pkg list

# Remove packages
kl pkg remove nodejs
\`\`\`

## Connecting to Environments

Connect your workspace to an environment to access services:

\`\`\`bash
kl environment connect development
\`\`\`

Now you can access all services in the development environment from your workspace.

## Next Steps

- Learn about environment forking
- Set up service intercepts
- Explore advanced workspace configurations
    `
  },
  'environment-forking-explained': {
    slug: 'environment-forking-explained',
    title: 'Environment Forking: Test Changes Without Breaking Production',
    excerpt: 'Discover how environment forking enables safe testing and experimentation without impacting your production services.',
    date: '2024-01-05',
    readTime: '6 min read',
    category: 'Feature',
    author: {
      name: 'Kloudlite Team',
      role: 'Engineering',
      avatar: 'KT',
      bio: 'Building the future of cloud development environments.'
    },
    content: `# Environment Forking Explained`
  }
}

interface BlogPostPageProps {
  params: Promise<{
    slug: string
  }>
}

export default async function BlogPostPage({ params }: BlogPostPageProps) {
  const { slug } = await params
  const post = blogPosts[slug as keyof typeof blogPosts]

  if (!post) {
    notFound()
  }

  // Get related posts (all other posts)
  const relatedPosts = Object.values(blogPosts)
    .filter(p => p.slug !== slug)
    .slice(0, 3)

  return (
    <div className="bg-background h-screen">
      <ScrollArea className="h-full">
        <WebsiteHeader currentPage="blog" />
        <main>
          <div className="px-6 pt-8 lg:px-8 lg:pt-12">
            <GridContainer className="px-6 lg:px-12">
              {/* Back to Blog */}
              <div className="py-8 border-b border-foreground/10 -mx-6 lg:-mx-12">
                <div className="px-6 lg:px-12">
                  <Link
                    href="/blog"
                    className="inline-flex items-center gap-2 text-muted-foreground hover:text-foreground transition-colors text-sm"
                  >
                    <ArrowLeft className="h-4 w-4" />
                    Back to Blog
                  </Link>
                </div>
              </div>

              {/* Article Header */}
              <div className="py-12 lg:py-16 border-b border-foreground/10 -mx-6 lg:-mx-12">
                <div className="px-6 lg:px-12 max-w-3xl mx-auto">
                  {/* Category */}
                  <div className="mb-6">
                    <span className="text-primary text-xs font-semibold uppercase tracking-wider">
                      {post.category}
                    </span>
                  </div>

                  {/* Title */}
                  <h1 className="text-[2.5rem] sm:text-[3rem] font-bold text-foreground mb-6 leading-[1.1] tracking-tight">
                    {post.title}
                  </h1>

                  {/* Subtitle/Excerpt */}
                  <p className="text-xl text-muted-foreground leading-relaxed mb-8">
                    {post.excerpt}
                  </p>

                  {/* Author and Meta */}
                  <div className="flex items-center justify-between py-6 border-t border-foreground/10">
                    <div className="flex items-center gap-4">
                      <div className="w-12 h-12 rounded-none bg-primary/10 flex items-center justify-center flex-shrink-0">
                        <span className="text-primary font-semibold text-sm">{post.author.avatar}</span>
                      </div>
                      <div>
                        <div className="font-semibold text-foreground text-sm">{post.author.name}</div>
                        <div className="flex items-center gap-3 text-muted-foreground text-sm">
                          <time dateTime={post.date}>
                            {new Date(post.date).toLocaleDateString('en-US', {
                              month: 'short',
                              day: 'numeric',
                              year: 'numeric'
                            })}
                          </time>
                          <span>·</span>
                          <span>{post.readTime}</span>
                        </div>
                      </div>
                    </div>

                    {/* Social actions */}
                    <div className="hidden sm:flex">
                      <SocialButtons slug={slug} title={post.title} excerpt={post.excerpt} type="header" />
                    </div>
                  </div>
                </div>
              </div>

              {/* Three Column Grid Layout */}
              <div className="grid lg:grid-cols-[1fr_2fr_1fr] -mx-6 lg:-mx-12">
                {/* Left Sidebar - Author */}
                <div className="hidden lg:block border-r border-foreground/10 p-8 lg:p-10">
                  <div className="sticky top-24">
                    <h3 className="text-xs font-semibold text-muted-foreground mb-6 uppercase tracking-wider">Written by</h3>
                    <div className="space-y-4">
                      <div className="w-16 h-16 rounded-none bg-primary/10 flex items-center justify-center">
                        <span className="text-primary font-semibold text-lg">{post.author.avatar}</span>
                      </div>
                      <div>
                        <div className="font-semibold text-foreground mb-1">{post.author.name}</div>
                        <div className="text-muted-foreground text-sm mb-3">{post.author.role}</div>
                        <p className="text-muted-foreground text-sm leading-relaxed">
                          {post.author.bio}
                        </p>
                      </div>
                    </div>
                  </div>
                </div>

                {/* Center Content */}
                <div className="border-r border-foreground/10 p-8 lg:p-10">
                  <div className="max-w-2xl mx-auto">
                  <div className="prose-article">
                    {post.content.split('\n\n').map((block, index) => {
                      const lines = block.trim().split('\n')
                      const firstLine = lines[0]

                      // Heading level 1
                      if (firstLine.startsWith('# ')) {
                        return (
                          <h2 key={index} className="text-[1.75rem] font-bold text-foreground mt-12 mb-6 first:mt-0 leading-[1.3]">
                            {firstLine.slice(2)}
                          </h2>
                        )
                      }

                      // Heading level 2
                      if (firstLine.startsWith('## ')) {
                        return (
                          <h3 key={index} className="text-[1.5rem] font-bold text-foreground mt-10 mb-5 leading-[1.3]">
                            {firstLine.slice(3)}
                          </h3>
                        )
                      }

                      // Heading level 3
                      if (firstLine.startsWith('### ')) {
                        return (
                          <h4 key={index} className="text-[1.25rem] font-semibold text-foreground mt-8 mb-4 leading-[1.4]">
                            {firstLine.slice(4)}
                          </h4>
                        )
                      }

                      // Code block
                      if (firstLine.startsWith('```')) {
                        const code = lines.slice(1, -1).join('\n')
                        return (
                          <div key={index} className="my-8">
                            <pre className="bg-muted p-5 rounded-lg overflow-x-auto border border-foreground/10">
                              <code className="text-[0.9375rem] font-mono text-foreground leading-[1.6]">{code}</code>
                            </pre>
                          </div>
                        )
                      }

                      // Unordered list
                      if (firstLine.startsWith('- ') || firstLine.startsWith('* ')) {
                        return (
                          <ul key={index} className="space-y-2 my-6">
                            {lines.map((line, i) => (
                              <li key={i} className="flex items-start gap-3 text-[1.0625rem] text-foreground leading-[1.7]">
                                <span className="w-1.5 h-1.5 bg-foreground rounded-none mt-2.5 flex-shrink-0" />
                                <span>{line.slice(2)}</span>
                              </li>
                            ))}
                          </ul>
                        )
                      }

                      // Numbered list
                      if (/^\d+\./.test(firstLine)) {
                        return (
                          <ol key={index} className="space-y-2 my-6 pl-6">
                            {lines.map((line, i) => {
                              const content = line.replace(/^\d+\.\s*/, '')
                              const [title, ...rest] = content.split(':')

                              if (rest.length > 0) {
                                return (
                                  <li key={i} className="text-[1.0625rem] leading-[1.7] list-decimal">
                                    <strong className="font-semibold text-foreground">{title}:</strong>
                                    <span className="text-foreground">{rest.join(':')}</span>
                                  </li>
                                )
                              }

                              return (
                                <li key={i} className="text-[1.0625rem] text-foreground leading-[1.7] list-decimal">
                                  {content}
                                </li>
                              )
                            })}
                          </ol>
                        )
                      }

                      // Regular paragraph
                      if (block.trim()) {
                        return (
                          <p key={index} className="text-[1.0625rem] text-foreground leading-[1.7] mb-6">
                            {block.trim()}
                          </p>
                        )
                      }

                      return null
                    })}
                  </div>

                    {/* Share Section */}
                    <div className="mt-16 pt-8 border-t border-foreground/10">
                      <div className="flex items-center justify-between">
                        <span className="text-sm text-muted-foreground">Share this article</span>
                        <SocialButtons slug={slug} title={post.title} excerpt={post.excerpt} type="share" />
                      </div>
                    </div>
                  </div>
                </div>

                {/* Right Sidebar - Related Articles */}
                <div className="hidden lg:block p-8 lg:p-10">
                  <div className="sticky top-24">
                    {relatedPosts.length > 0 && (
                      <div>
                        <h3 className="text-xs font-semibold text-muted-foreground mb-6 uppercase tracking-wider">Related Articles</h3>
                        <div className="space-y-8">
                          {relatedPosts.map((related) => (
                            <Link
                              key={related.slug}
                              href={`/blog/${related.slug}`}
                              className="block group"
                            >
                              <div>
                                <div className="mb-3">
                                  <span className="text-xs font-semibold text-primary uppercase tracking-wider">
                                    {related.category}
                                  </span>
                                </div>
                                <h4 className="text-base font-bold text-foreground mb-3 leading-snug group-hover:text-primary transition-colors">
                                  {related.title}
                                </h4>
                                <p className="text-sm text-muted-foreground leading-relaxed mb-3 line-clamp-3">
                                  {related.excerpt}
                                </p>
                                <div className="flex items-center gap-2 text-muted-foreground text-xs">
                                  <time dateTime={related.date}>
                                    {new Date(related.date).toLocaleDateString('en-US', {
                                      month: 'short',
                                      day: 'numeric'
                                    })}
                                  </time>
                                  <span>·</span>
                                  <span>{related.readTime}</span>
                                </div>
                              </div>
                            </Link>
                          ))}
                        </div>
                      </div>
                    )}
                  </div>
                </div>
              </div>
            </GridContainer>
          </div>

          <WebsiteFooter />
        </main>
      </ScrollArea>
    </div>
  )
}

// Generate static params for known blog posts
export async function generateStaticParams() {
  return Object.keys(blogPosts).map((slug) => ({
    slug,
  }))
}
