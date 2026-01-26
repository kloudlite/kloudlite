import { ScrollArea } from '@kloudlite/ui'
import { WebsiteHeader } from '@/components/website-header'
import { WebsiteFooter } from '@/components/website-footer'
import { PageHeroTitle } from '@/components/page-hero-title'
import { FeaturedBlogCard } from '@/components/blog/featured-blog-card'
import { BlogCard } from '@/components/blog/blog-card'
import { GridContainer } from '@/components/blog/grid-container'
import { blogPostsData } from '@/data/blog-posts'

export default function BlogPage() {
  const featuredPost = blogPostsData.find(post => post.featured)
  const regularPosts = blogPostsData.filter(post => !post.featured)

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
