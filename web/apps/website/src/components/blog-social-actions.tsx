'use client'

import { useState, useEffect } from 'react'
import { Button } from '@kloudlite/ui'
import { Share2, Bookmark, Twitter, Linkedin, Facebook } from 'lucide-react'

interface SocialButtonsProps {
  slug: string
  title: string
  excerpt: string
  type: 'header' | 'share'
}

export function SocialButtons({ slug, title, excerpt, type }: SocialButtonsProps) {
  const [isBookmarked, setIsBookmarked] = useState(false)

  useEffect(() => {
    // Check if article is bookmarked
    const bookmarks = JSON.parse(localStorage.getItem('blog-bookmarks') || '[]')
    setIsBookmarked(bookmarks.includes(slug))
  }, [slug])

  const toggleBookmark = () => {
    const bookmarks = JSON.parse(localStorage.getItem('blog-bookmarks') || '[]')

    if (isBookmarked) {
      // Remove bookmark
      const updated = bookmarks.filter((b: string) => b !== slug)
      localStorage.setItem('blog-bookmarks', JSON.stringify(updated))
      setIsBookmarked(false)
    } else {
      // Add bookmark
      const updated = [...bookmarks, slug]
      localStorage.setItem('blog-bookmarks', JSON.stringify(updated))
      setIsBookmarked(true)
    }
  }

  const shareUrl = typeof window !== 'undefined' ? window.location.href : ''

  const shareOnTwitter = () => {
    const text = `${title} - ${excerpt}`
    const url = `https://twitter.com/intent/tweet?text=${encodeURIComponent(text)}&url=${encodeURIComponent(shareUrl)}`
    window.open(url, '_blank', 'width=600,height=400')
  }

  const shareOnLinkedIn = () => {
    const url = `https://www.linkedin.com/sharing/share-offsite/?url=${encodeURIComponent(shareUrl)}`
    window.open(url, '_blank', 'width=600,height=400')
  }

  const shareOnFacebook = () => {
    const url = `https://www.facebook.com/sharer/sharer.php?u=${encodeURIComponent(shareUrl)}`
    window.open(url, '_blank', 'width=600,height=400')
  }

  const handleShare = async () => {
    if (navigator.share) {
      try {
        await navigator.share({
          title,
          text: excerpt,
          url: shareUrl,
        })
      } catch (err) {
        console.log('Share cancelled')
      }
    }
  }

  if (type === 'header') {
    return (
      <div className="flex items-center gap-2">
        <Button
          variant="ghost"
          size="sm"
          onClick={toggleBookmark}
          className="p-2 rounded-none hover:bg-foreground/5 transition-colors text-muted-foreground hover:text-foreground"
          aria-label={isBookmarked ? 'Remove bookmark' : 'Bookmark article'}
        >
          <Bookmark className={`h-5 w-5 ${isBookmarked ? 'fill-current text-primary' : ''}`} />
        </Button>
        <Button
          variant="ghost"
          size="sm"
          onClick={handleShare}
          className="p-2 rounded-none hover:bg-foreground/5 transition-colors text-muted-foreground hover:text-foreground"
          aria-label="Share article"
        >
          <Share2 className="h-5 w-5" />
        </Button>
      </div>
    )
  }

  return (
    <div className="flex items-center gap-2">
      <Button
        variant="ghost"
        size="sm"
        onClick={shareOnTwitter}
        className="p-2.5 rounded-none hover:bg-foreground/5 transition-colors text-muted-foreground hover:text-foreground"
        aria-label="Share on Twitter"
      >
        <Twitter className="h-5 w-5" />
      </Button>
      <Button
        variant="ghost"
        size="sm"
        onClick={shareOnLinkedIn}
        className="p-2.5 rounded-none hover:bg-foreground/5 transition-colors text-muted-foreground hover:text-foreground"
        aria-label="Share on LinkedIn"
      >
        <Linkedin className="h-5 w-5" />
      </Button>
      <Button
        variant="ghost"
        size="sm"
        onClick={shareOnFacebook}
        className="p-2.5 rounded-none hover:bg-foreground/5 transition-colors text-muted-foreground hover:text-foreground"
        aria-label="Share on Facebook"
      >
        <Facebook className="h-5 w-5" />
      </Button>
    </div>
  )
}
