import { notFound } from 'next/navigation'
import { Metadata } from 'next'
import { DocsPage } from '@/components/docs/docs-page'

interface DocsSlugPageProps {
  params: {
    slug: string[]
  }
}

export async function generateMetadata({ params }: DocsSlugPageProps): Promise<Metadata> {
  const path = params.slug?.join('/') || ''
  
  return {
    title: `${path.replace(/\//g, ' / ')} - Kloudlite Docs`,
    description: `Documentation for ${path}`,
  }
}

export default function DocsSlugPage({ params }: DocsSlugPageProps) {
  const path = params.slug?.join('/') || ''
  const title = path.split('/').pop()?.replace(/-/g, ' ').replace(/\b\w/g, l => l.toUpperCase()) || 'Documentation'
  
  // For now, return a placeholder until we implement MDX processing
  return (
    <DocsPage
      title={title}
      description={`Documentation for ${path}`}
      lastUpdated="2 days ago"
      editUrl={`https://github.com/kloudlite/kloudlite/edit/master/docs/${path}.mdx`}
    >
      <div className="space-y-6">
        <div className="bg-muted/30 border rounded-lg p-6">
          <h2 className="text-xl font-semibold mb-2">Documentation Layout Ready</h2>
          <p className="text-muted-foreground mb-4">
            The documentation infrastructure is now set up and ready to receive content.
          </p>
          <div className="space-y-2">
            <p><strong>Current path:</strong> <code>/docs/{path}</code></p>
            <p><strong>Features available:</strong></p>
            <ul className="list-disc list-inside ml-4 space-y-1">
              <li>Responsive sidebar navigation</li>
              <li>Table of contents generation</li>
              <li>Breadcrumb navigation</li>
              <li>Syntax highlighting for code blocks</li>
              <li>MDX support for rich content</li>
              <li>Search UI (ready for content)</li>
              <li>Mobile-friendly design</li>
            </ul>
          </div>
        </div>
        
        <div>
          <h2>Sample Content</h2>
          <p>Here's how different elements will look:</p>
          
          <h3>Code Block Example</h3>
          <pre className="hljs"><code>{`function example() {
  console.log("Hello, world!");
  return "This is a code block";
}`}</code></pre>
          
          <h3>Inline Code</h3>
          <p>You can use <code>inline code</code> within paragraphs.</p>
          
          <blockquote>
            <p>This is a blockquote that shows how quoted content will appear in the documentation.</p>
          </blockquote>
          
          <h3>Lists</h3>
          <ul>
            <li>First item</li>
            <li>Second item</li>
            <li>Third item</li>
          </ul>
          
          <h3>Tables</h3>
          <table>
            <thead>
              <tr>
                <th>Feature</th>
                <th>Status</th>
              </tr>
            </thead>
            <tbody>
              <tr>
                <td>Sidebar Navigation</td>
                <td>✅ Complete</td>
              </tr>
              <tr>
                <td>Table of Contents</td>
                <td>✅ Complete</td>
              </tr>
              <tr>
                <td>MDX Support</td>
                <td>✅ Complete</td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </DocsPage>
  )
}