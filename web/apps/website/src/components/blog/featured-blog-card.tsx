import Link from 'next/link'
import { ArrowRight, Calendar, Clock } from 'lucide-react'
import { HighlightedWord } from './highlighted-word'

interface FeaturedBlogCardProps {
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

export function FeaturedBlogCard({ post }: FeaturedBlogCardProps) {
  return (
    <Link
      href={`/blog/${post.slug}`}
      className="group block hover:bg-foreground/[0.02] transition-[background-color] duration-300"
    >
      <div className="p-12 lg:p-16 space-y-6">
        {/* Category and Featured Badge */}
        <div className="flex items-center gap-4">
          <span className="inline-flex items-center px-3 py-1.5 bg-primary text-primary-foreground text-xs font-bold uppercase tracking-wider">
            Featured
          </span>
          <span className="text-sm font-semibold text-muted-foreground uppercase tracking-wider">
            {post.category}
          </span>
        </div>

        {/* Title */}
        <h2 className="text-3xl sm:text-4xl lg:text-5xl font-bold text-foreground leading-[1.15]">
          Introducing <HighlightedWord>Kloudlite</HighlightedWord>: Cloud Development Environments
        </h2>

        {/* Excerpt */}
        <p className="text-muted-foreground text-lg leading-relaxed max-w-3xl">
          {post.excerpt}
        </p>

        {/* Metadata Footer */}
        <div className="flex flex-wrap items-center gap-6 text-muted-foreground text-sm pt-6">
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
          <div className="flex items-center gap-2 text-primary font-semibold ml-auto">
            <span>Read article</span>
            <ArrowRight className="h-4 w-4 transition-transform duration-300 group-hover:translate-x-1" />
          </div>
        </div>
      </div>
    </Link>
  )
}
