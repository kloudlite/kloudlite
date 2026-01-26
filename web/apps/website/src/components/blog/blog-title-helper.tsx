import { HighlightedWord } from './highlighted-word'

interface BlogPost {
  slug: string
  title: string
}

// Get title with highlighted word based on post slug
export function getBlogTitle(post: BlogPost, variant: 'hover' | 'static' = 'hover') {
  switch (post.slug) {
    case 'introducing-kloudlite':
      return <>Introducing Kloudlite: Cloud Development <HighlightedWord variant={variant}>Environments</HighlightedWord></>
    case 'getting-started-with-workspaces':
      return <>Getting Started with Kloudlite <HighlightedWord variant={variant}>Workspaces</HighlightedWord></>
    case 'environment-forking-explained':
      return <>Environment <HighlightedWord variant={variant}>Forking</HighlightedWord>: Test Changes Without Breaking Production</>
    case 'service-intercepts-deep-dive':
      return <>Service <HighlightedWord variant={variant}>Intercepts</HighlightedWord>: Route Production Traffic to Your Workspace</>
    case 'nix-package-management':
      return <>Why We Chose <HighlightedWord variant={variant}>Nix</HighlightedWord> for Package Management</>
    case 'kubernetes-native-development':
      return <><HighlightedWord variant={variant}>Kubernetes-Native</HighlightedWord> Development Made Simple</>
    default:
      return post.title
  }
}
