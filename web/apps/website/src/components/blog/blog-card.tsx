import Link from 'next/link'
import { ArrowRight } from 'lucide-react'
import { getBlogTitle } from './blog-title-helper'

interface BlogCardProps {
  post: {
    slug: string
    title: string
    excerpt: string
    date: string
    readTime: string
    category: string
    featured: boolean
  }
}

export function BlogCard({ post }: BlogCardProps) {
  return (
    <Link
      href={`/blog/${post.slug}`}
      className="group block border-b border-foreground/10 sm:border-r sm:odd:border-r sm:even:border-r-0 hover:bg-foreground/[0.015] transition-[background-color] duration-300 h-full"
    >
      <div className="p-8 lg:p-10 h-full flex flex-col">
        {/* Category Badge */}
        <div className="mb-5">
          <span className="inline-block px-3 py-1 bg-foreground/[0.06] text-muted-foreground text-[11px] font-semibold uppercase tracking-wider">
            {post.category}
          </span>
        </div>

        {/* Title */}
        <h3 className="text-foreground text-xl lg:text-2xl font-bold mb-4 leading-tight">
          {getBlogTitle(post)}
        </h3>

        {/* Excerpt */}
        <p className="text-muted-foreground text-[15px] leading-relaxed mb-auto">
          {post.excerpt}
        </p>

        {/* Footer */}
        <div className="flex items-center justify-between pt-6 mt-6 border-t border-foreground/10">
          <div className="flex items-center gap-3 text-muted-foreground text-xs">
            <time dateTime={post.date}>
              {new Date(post.date).toLocaleDateString('en-US', {
                month: 'short',
                day: 'numeric',
                year: 'numeric'
              })}
            </time>
            <span>•</span>
            <span>{post.readTime}</span>
          </div>
          <ArrowRight className="h-4 w-4 text-primary transition-transform duration-300 group-hover:translate-x-1" />
        </div>
      </div>
    </Link>
  )
}
