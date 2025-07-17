import type { MDXComponents } from 'mdx/types'
import { Code, Copy, Check, ExternalLink } from 'lucide-react'
import { useState } from 'react'
import { Button } from '@/components/ui/button'
import { Link } from '@/components/ui/link'
import { cn } from '@/lib/utils'

// Custom code block component with copy functionality
function CodeBlock({ 
  className, 
  children, 
  ...props 
}: React.HTMLAttributes<HTMLPreElement>) {
  const [copied, setCopied] = useState(false)
  
  const copyToClipboard = async () => {
    if (typeof children === 'string') {
      await navigator.clipboard.writeText(children)
      setCopied(true)
      setTimeout(() => setCopied(false), 2000)
    }
  }

  return (
    <div className="relative group">
      <pre className={cn('overflow-x-auto', className)} {...props}>
        {children}
      </pre>
      <Button
        variant="ghost"
        size="sm"
        className="absolute top-2 right-2 opacity-0 group-hover:opacity-100 transition-opacity"
        onClick={copyToClipboard}
      >
        {copied ? <Check className="h-4 w-4" /> : <Copy className="h-4 w-4" />}
      </Button>
    </div>
  )
}

// Custom inline code component
function InlineCode({ 
  className, 
  children, 
  ...props 
}: React.HTMLAttributes<HTMLElement>) {
  return (
    <code 
      className={cn(
        'relative rounded bg-muted px-[0.3rem] py-[0.2rem] text-sm font-mono',
        className
      )}
      {...props}
    >
      {children}
    </code>
  )
}

// Custom callout component
function Callout({ 
  type = 'info', 
  children,
  ...props 
}: React.HTMLAttributes<HTMLDivElement> & { type?: 'info' | 'warning' | 'error' | 'success' }) {
  const styles = {
    info: 'border-primary/20 bg-primary/10 text-primary',
    warning: 'border-warning/20 bg-warning/10 text-warning',
    error: 'border-destructive/20 bg-destructive/10 text-destructive',
    success: 'border-success/20 bg-success/10 text-success'
  }

  return (
    <div 
      className={cn(
        'my-6 border-l-4 p-4 rounded-r-lg',
        styles[type]
      )}
      {...props}
    >
      {children}
    </div>
  )
}

// Enhanced link component
function CustomLink({
  href,
  children,
  ...props
}: React.AnchorHTMLAttributes<HTMLAnchorElement>) {
  const isExternal = href?.startsWith('http')
  
  return (
    <Link 
      href={href || ''} 
      className={cn(
        'inline-flex items-center gap-1 text-primary hover:underline',
        isExternal && 'after:content-["↗"] after:text-xs after:ml-1'
      )}
      {...props}
    >
      {children}
      {isExternal && <ExternalLink className="h-3 w-3" />}
    </Link>
  )
}

export function useMDXComponents(components: MDXComponents): MDXComponents {
  return {
    // Typography
    h1: ({ className, ...props }) => (
      <h1 
        className={cn('scroll-m-20 text-3xl font-bold tracking-tight', className)} 
        {...props} 
      />
    ),
    h2: ({ className, ...props }) => (
      <h2 
        className={cn('scroll-m-20 border-b pb-2 text-2xl font-semibold tracking-tight first:mt-0', className)} 
        {...props} 
      />
    ),
    h3: ({ className, ...props }) => (
      <h3 
        className={cn('scroll-m-20 text-xl font-semibold tracking-tight', className)} 
        {...props} 
      />
    ),
    h4: ({ className, ...props }) => (
      <h4 
        className={cn('scroll-m-20 text-lg font-semibold tracking-tight', className)} 
        {...props} 
      />
    ),
    h5: ({ className, ...props }) => (
      <h5 
        className={cn('scroll-m-20 text-base font-semibold tracking-tight', className)} 
        {...props} 
      />
    ),
    h6: ({ className, ...props }) => (
      <h6 
        className={cn('scroll-m-20 text-sm font-semibold tracking-tight', className)} 
        {...props} 
      />
    ),
    p: ({ className, ...props }) => (
      <p 
        className={cn('leading-7 [&:not(:first-child)]:mt-6', className)} 
        {...props} 
      />
    ),
    ul: ({ className, ...props }) => (
      <ul 
        className={cn('my-6 ml-6 list-disc', className)} 
        {...props} 
      />
    ),
    ol: ({ className, ...props }) => (
      <ol 
        className={cn('my-6 ml-6 list-decimal', className)} 
        {...props} 
      />
    ),
    li: ({ className, ...props }) => (
      <li 
        className={cn('mt-2', className)} 
        {...props} 
      />
    ),
    blockquote: ({ className, ...props }) => (
      <blockquote 
        className={cn('mt-6 border-l-2 pl-6 italic', className)} 
        {...props} 
      />
    ),
    
    // Code
    pre: CodeBlock,
    code: InlineCode,
    
    // Links
    a: CustomLink,
    
    // Tables
    table: ({ className, ...props }) => (
      <div className="my-6 w-full overflow-y-auto">
        <table className={cn('w-full', className)} {...props} />
      </div>
    ),
    tr: ({ className, ...props }) => (
      <tr 
        className={cn('m-0 border-t p-0 even:bg-muted', className)} 
        {...props} 
      />
    ),
    th: ({ className, ...props }) => (
      <th 
        className={cn('border px-4 py-2 text-left font-bold [&[align=center]]:text-center [&[align=right]]:text-right', className)} 
        {...props} 
      />
    ),
    td: ({ className, ...props }) => (
      <td 
        className={cn('border px-4 py-2 text-left [&[align=center]]:text-center [&[align=right]]:text-right', className)} 
        {...props} 
      />
    ),
    
    // Custom components
    Callout,
    
    ...components,
  }
}