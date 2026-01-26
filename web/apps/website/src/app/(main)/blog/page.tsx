import { ScrollArea } from '@kloudlite/ui'
import { WebsiteHeader } from '@/components/website-header'
import { WebsiteFooter } from '@/components/website-footer'
import { PageHeroTitle } from '@/components/page-hero-title'
import { FeaturedBlogCard } from '@/components/blog/featured-blog-card'
import { BlogCard } from '@/components/blog/blog-card'
import { GridContainer } from '@/components/blog/grid-container'

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
                  <PageHeroTitle accentedWord="Insights.">
                    Engineering
                  </PageHeroTitle>
                  <p className="text-muted-foreground mx-auto mt-8 max-w-lg text-lg leading-relaxed">
                    Guides, updates, and best practices from the Kloudlite team.
                  </p>
                </div>
              </div>

              {/* Featured Post */}
              {featuredPost && (
                <div className="border-t border-foreground/10 -mx-6 lg:-mx-12">
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
