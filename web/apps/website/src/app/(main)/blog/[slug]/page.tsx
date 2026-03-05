import { ScrollArea } from '@kloudlite/ui'
import { WebsiteHeader } from '@/components/website-header'
import { WebsiteFooter } from '@/components/website-footer'
import { SocialButtons } from '@/components/blog-social-actions'
import { getBlogTitle } from '@/components/blog/blog-title-helper'
import { GridContainer } from '@/components/blog/grid-container'
import Link from 'next/link'
import { ArrowLeft, Calendar, Clock } from 'lucide-react'
import { notFound } from 'next/navigation'
import { blogPostsMap, blogPostsData } from '@/data/blog-posts'

interface BlogPostPageProps {
  params: Promise<{
    slug: string
  }>
}

export default async function BlogPostPage({ params }: BlogPostPageProps) {
  const { slug } = await params
  const post = blogPostsMap[slug]

  if (!post) {
    notFound()
  }

  // Get related posts from the post's relatedPosts array
  const relatedPosts = post.relatedPosts
    .map(relatedSlug => blogPostsMap[relatedSlug])
    .filter(Boolean)
    .slice(0, 3)

  return (
    <div className="bg-background h-screen">
      <ScrollArea className="h-full">
        <WebsiteHeader currentPage="blog" />
        <main>
          <div className="px-6 pt-8 lg:px-8 lg:pt-12">
            <GridContainer maxWidth="max-w-7xl" className="px-6 lg:px-12">
              {/* Back to Blog */}
              <div className="py-8 border-b border-foreground/10 bg-foreground/[0.01] -mx-6 lg:-mx-12">
                <div className="px-6 lg:px-12">
                  <Link
                    href="/blog"
                    className="group inline-flex items-center gap-2 text-muted-foreground hover:text-primary transition-colors duration-300 text-sm"
                  >
                    <ArrowLeft className="h-4 w-4 transition-transform duration-300 group-hover:-translate-x-1" />
                    <span className="relative">
                      Back to Blog
                      <span className="absolute -bottom-0.5 left-0 right-0 h-0.5 bg-primary scale-x-0 group-hover:scale-x-100 transition-transform duration-300 origin-left" />
                    </span>
                  </Link>
                </div>
              </div>

              {/* Article Header */}
              <div className="py-12 lg:py-16 border-b border-foreground/10 -mx-6 lg:-mx-12">
                <div className="px-6 lg:px-12 max-w-3xl mx-auto">
                  {/* Category */}
                  <div className="mb-6">
                    <span className="inline-flex items-center px-3 py-1.5 bg-foreground/[0.06] border border-foreground/10 rounded-sm text-[11px] font-semibold uppercase tracking-wider text-primary">
                      {post.category}
                    </span>
                  </div>

                  {/* Title */}
                  <h1 className="text-[2.5rem] sm:text-4xl lg:text-5xl font-bold text-foreground mb-6 leading-[1.1] tracking-[-0.02em]">
                    {getBlogTitle(post, 'static')}
                  </h1>

                  {/* Subtitle/Excerpt */}
                  <p className="text-xl text-muted-foreground leading-relaxed mb-8">
                    {post.excerpt}
                  </p>

                  {/* Author and Meta */}
                  <div className="flex items-center justify-between py-6 border-t border-foreground/10">
                    <div className="group flex items-center gap-4">
                      <div className="w-12 h-12 rounded-sm bg-gradient-to-br from-primary/10 to-primary/5 border border-primary/20 flex items-center justify-center flex-shrink-0 transition-all duration-300 group-hover:scale-110 group-hover:border-primary/30">
                        <span className="text-primary font-semibold text-sm transition-colors duration-300">{post.author.avatar}</span>
                      </div>
                      <div>
                        <div className="font-semibold text-foreground text-sm transition-colors duration-300 group-hover:text-primary">{post.author.name}</div>
                        <div className="flex items-center gap-3 text-muted-foreground text-xs">
                          <div className="flex items-center gap-1.5 transition-colors duration-300 group-hover:text-foreground">
                            <Calendar className="h-3 w-3" />
                            <time dateTime={post.date}>
                              {new Date(post.date).toLocaleDateString('en-US', {
                                month: 'short',
                                day: 'numeric',
                                year: 'numeric'
                              })}
                            </time>
                          </div>
                          <span>·</span>
                          <div className="flex items-center gap-1.5 transition-colors duration-300 group-hover:text-foreground">
                            <Clock className="h-3 w-3" />
                            <span>{post.readTime}</span>
                          </div>
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

              {/* Two Column Grid Layout */}
              <div className="grid lg:grid-cols-[3fr_1fr] -mx-6 lg:-mx-12">
                {/* Content */}
                <div className="border-r border-foreground/10 p-8 lg:p-10">
                  <div className="prose-article">
                    {post.content.split('\n\n').map((block, index) => {
                      const lines = block.trim().split('\n')
                      const firstLine = lines[0]
                      // Create a simple hash from the block content for unique keys
                      const blockKey = `${index}-${block.trim().slice(0, 20)}`

                      // Heading level 1
                      if (firstLine.startsWith('# ')) {
                        return (
                          <h2 key={blockKey} className="group text-[1.75rem] lg:text-3xl font-bold text-foreground mt-12 mb-6 first:mt-0 leading-[1.3] tracking-tight">
                            <span className="relative inline-block">
                              {firstLine.slice(2)}
                              <span className="absolute -bottom-1 left-0 right-0 h-0.5 bg-primary scale-x-0 group-hover:scale-x-100 transition-transform duration-300 origin-left" />
                            </span>
                          </h2>
                        )
                      }

                      // Heading level 2
                      if (firstLine.startsWith('## ')) {
                        return (
                          <h3 key={blockKey} className="text-[1.5rem] lg:text-2xl font-bold text-foreground mt-10 mb-5 leading-[1.3] tracking-tight">
                            {firstLine.slice(3)}
                          </h3>
                        )
                      }

                      // Heading level 3
                      if (firstLine.startsWith('### ')) {
                        return (
                          <h4 key={blockKey} className="text-[1.25rem] lg:text-xl font-semibold text-foreground mt-8 mb-4 leading-[1.4] tracking-tight">
                            {firstLine.slice(4)}
                          </h4>
                        )
                      }

                      // Code block
                      if (firstLine.startsWith('```')) {
                        const code = lines.slice(1, -1).join('\n')
                        return (
                          <div key={blockKey} className="group relative my-8">
                            {/* Top accent line */}
                            <div className="absolute top-0 left-0 right-0 h-[2px] bg-gradient-to-r from-transparent via-primary to-transparent scale-x-0 group-hover:scale-x-100 transition-transform duration-500 origin-center" />

                            <pre className="bg-foreground/[0.04] group-hover:bg-foreground/[0.05] p-5 rounded-sm overflow-x-auto border border-foreground/10 group-hover:border-foreground/20 transition-all duration-300">
                              <code className="text-[0.9375rem] font-mono text-foreground leading-[1.6] transition-colors duration-300">{code}</code>
                            </pre>
                          </div>
                        )
                      }

                      // Unordered list
                      if (firstLine.startsWith('- ') || firstLine.startsWith('* ')) {
                        return (
                          <ul key={blockKey} className="space-y-3 my-6">
                            {lines.map((line, i) => (
                              <li key={`${blockKey}-${i}`} className="group flex items-start gap-3 text-[1.0625rem] text-foreground leading-[1.7]">
                                <span className="w-1.5 h-1.5 bg-foreground rounded-sm mt-2.5 flex-shrink-0 transition-all duration-300 group-hover:scale-125 group-hover:bg-primary" />
                                <span>{line.slice(2)}</span>
                              </li>
                            ))}
                          </ul>
                        )
                      }

                      // Numbered list
                      if (/^\d+\./.test(firstLine)) {
                        return (
                          <ol key={blockKey} className="space-y-2 my-6 pl-6">
                            {lines.map((line, i) => {
                              const content = line.replace(/^\d+\.\s*/, '')
                              const [title, ...rest] = content.split(':')

                              if (rest.length > 0) {
                                return (
                                  <li key={`${blockKey}-${i}`} className="text-[1.0625rem] leading-[1.7] list-decimal">
                                    <strong className="font-semibold text-foreground">{title}:</strong>
                                    <span className="text-foreground">{rest.join(':')}</span>
                                  </li>
                                )
                              }

                              return (
                                <li key={`${blockKey}-${i}`} className="text-[1.0625rem] text-foreground leading-[1.7] list-decimal">
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
                          <p key={blockKey} className="text-[1.0625rem] text-foreground leading-[1.7] mb-6">
                            {block.trim()}
                          </p>
                        )
                      }

                      return null
                    })}
                  </div>

                  {/* Share Section */}
                  <div className="mt-16">
                    <div className="p-6 bg-foreground/[0.015] border border-foreground/10 rounded-sm">
                      <div className="space-y-3">
                        <h3 className="text-base font-semibold text-foreground">Enjoyed this article?</h3>
                        <p className="text-sm text-muted-foreground">Share it with your network</p>
                        <div className="pt-3">
                          <SocialButtons slug={slug} title={post.title} excerpt={post.excerpt} type="share" />
                        </div>
                      </div>
                    </div>
                  </div>
                </div>

                {/* Right Sidebar - Author & Related Articles */}
                <div className="hidden lg:block">
                  <div className="sticky top-24">
                    {/* Written by */}
                    <div className="p-8 lg:p-10 border-b border-foreground/10">
                      <h3 className="text-xs font-semibold text-muted-foreground mb-6 uppercase tracking-wider">Written by</h3>
                      <div className="w-24 h-24 border border-foreground/10 bg-foreground/[0.06] flex items-center justify-center mb-4">
                        <span className="text-primary font-semibold text-xl">{post.author.avatar}</span>
                      </div>
                      <div className="font-semibold text-foreground mb-1">{post.author.name}</div>
                      <div className="text-muted-foreground text-sm mb-3">{post.author.role}</div>
                      <p className="text-muted-foreground text-sm leading-relaxed">
                        {post.author.bio}
                      </p>
                    </div>

                    {/* Related Articles */}
                    {relatedPosts.length > 0 && (
                      <div className="p-8 lg:p-10">
                        <h3 className="text-xs font-semibold text-muted-foreground mb-6 uppercase tracking-wider">Related Articles</h3>
                        <div className="space-y-0">
                          {relatedPosts.map((related) => (
                            <Link
                              key={related.slug}
                              href={`/blog/${related.slug}`}
                              className="group block border-b border-foreground/10 last:border-b-0 py-6 first:pt-0 hover:bg-foreground/[0.015] -mx-8 lg:-mx-10 px-8 lg:px-10 transition-[background-color] duration-300"
                            >
                              <div className="mb-3">
                                <span className="inline-block px-3 py-1 bg-foreground/[0.06] text-muted-foreground text-[11px] font-semibold uppercase tracking-wider">
                                  {related.category}
                                </span>
                              </div>
                              <h4 className="text-base font-bold text-foreground mb-3 leading-snug">
                                {getBlogTitle(related)}
                              </h4>
                              <p className="text-sm text-muted-foreground leading-relaxed mb-4 line-clamp-3">
                                {related.excerpt}
                              </p>
                              <div className="flex items-center gap-3 text-muted-foreground text-xs">
                                <time dateTime={related.date}>
                                  {new Date(related.date).toLocaleDateString('en-US', {
                                    month: 'short',
                                    day: 'numeric'
                                  })}
                                </time>
                                <span>•</span>
                                <span>{related.readTime}</span>
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
  return blogPostsData.map((post) => ({
    slug: post.slug,
  }))
}
