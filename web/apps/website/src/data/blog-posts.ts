// Blog posts data structure
// This file contains all blog post metadata and content

export interface BlogPost {
  slug: string
  title: string
  excerpt: string
  date: string
  readTime: string
  category: 'Product' | 'Tutorial' | 'Feature' | 'Technical' | 'Architecture' | 'Platform' | 'Workflow' | 'Philosophy'
  featured: boolean
  author: {
    name: string
    role: string
    avatar: string
    bio: string
  }
  content: string
}

export const defaultAuthor = {
  name: 'Kloudlite Team',
  role: 'Engineering',
  avatar: 'KT',
  bio: 'Building the future of cloud development environments.'
}

export const blogPostsData: BlogPost[] = [
  // Core Features
  {
    slug: 'environment-forking',
    title: 'Environment Forking: Clone Entire Environments with a Single Command',
    excerpt: 'Learn how to fork entire environments with all services, databases, and configurations for isolated testing without affecting other developers.',
    date: '2024-02-01',
    readTime: '7 min read',
    category: 'Feature',
    featured: true,
    author: defaultAuthor,
    content: '# Environment Forking\n\nContent coming soon...'
  },
  {
    slug: 'workspace-forking',
    title: 'Workspace Forking: Parallel Development Made Simple',
    excerpt: 'Fork workspaces instantly for parallel work, run multiple experiments, or spin up AI agents simultaneously without configuration hassle.',
    date: '2024-01-28',
    readTime: '6 min read',
    category: 'Feature',
    featured: false,
    author: defaultAuthor,
    content: '# Workspace Forking\n\nContent coming soon...'
  },
  {
    slug: 'service-intercepts',
    title: 'Service Intercepts: Debug Production with Real Traffic',
    excerpt: 'Route environment service traffic directly to your workspace to debug production issues with real data, no mocks, no redeployment.',
    date: '2024-01-25',
    readTime: '8 min read',
    category: 'Feature',
    featured: false,
    author: defaultAuthor,
    content: '# Service Intercepts\n\nContent coming soon...'
  },
  {
    slug: 'nix-package-management',
    title: 'Nix Package Management: Reproducible Dependencies Everywhere',
    excerpt: 'Discover how Nix-based package management ensures every developer gets the same, reproducible environment with a single command.',
    date: '2024-01-22',
    readTime: '10 min read',
    category: 'Technical',
    featured: false,
    author: defaultAuthor,
    content: '# Nix Package Management\n\nContent coming soon...'
  },
  {
    slug: 'environment-switching',
    title: 'Environment Switching: Seamlessly Move Between Contexts',
    excerpt: 'Switch between dev, staging, and production environments seamlessly with automatic DNS resolution and service routing.',
    date: '2024-01-20',
    readTime: '5 min read',
    category: 'Feature',
    featured: false,
    author: defaultAuthor,
    content: '# Environment Switching\n\nContent coming soon...'
  },
  {
    slug: 'environment-snapshots',
    title: 'Environment Snapshots: Capture and Restore Complete States',
    excerpt: 'Capture complete environment states instantly and restore them later for reproducible testing, debugging, or rollback scenarios.',
    date: '2024-01-18',
    readTime: '6 min read',
    category: 'Feature',
    featured: false,
    author: defaultAuthor,
    content: '# Environment Snapshots\n\nContent coming soon...'
  },
  {
    slug: 'workspace-snapshots',
    title: 'Workspace Snapshots: Share Configurations Effortlessly',
    excerpt: 'Save and share workspace configurations with your team, ensuring everyone starts with the same setup and tools.',
    date: '2024-01-15',
    readTime: '5 min read',
    category: 'Feature',
    featured: false,
    author: defaultAuthor,
    content: '# Workspace Snapshots\n\nContent coming soon...'
  },
  {
    slug: 'ai-ready-workspaces',
    title: 'AI-Ready Workspaces: Built-in Support for AI Coding Tools',
    excerpt: 'Explore built-in support for Claude Code, Gemini CLI, OpenCode, Codex CLI, and MCP - enabling vibecoding sessions out of the box.',
    date: '2024-01-12',
    readTime: '9 min read',
    category: 'Feature',
    featured: true,
    author: defaultAuthor,
    content: '# AI-Ready Workspaces\n\nContent coming soon...'
  },
  {
    slug: 'ide-integration',
    title: 'IDE Integration: Access from Your Favorite Editor',
    excerpt: 'Connect your local VS Code, Cursor, JetBrains, Zed, or any IDE directly to cloud resources with zero latency and full IntelliSense.',
    date: '2024-01-10',
    readTime: '7 min read',
    category: 'Tutorial',
    featured: false,
    author: defaultAuthor,
    content: '# IDE Integration\n\nContent coming soon...'
  },
  {
    slug: 'docker-compose-compatible',
    title: 'Docker Compose Compatible: Zero Migration Required',
    excerpt: 'Use your existing Docker Compose files without modifications. If it runs in Docker, it runs in Kloudlite.',
    date: '2024-01-08',
    readTime: '6 min read',
    category: 'Technical',
    featured: false,
    author: defaultAuthor,
    content: '# Docker Compose Compatible\n\nContent coming soon...'
  },
  {
    slug: 'network-isolation',
    title: 'Network Isolation: Private, Secure Environment Boundaries',
    excerpt: 'Each environment runs in its own network namespace with complete isolation - no cross-contamination between staging and production.',
    date: '2024-01-05',
    readTime: '8 min read',
    category: 'Technical',
    featured: false,
    author: defaultAuthor,
    content: '# Network Isolation\n\nContent coming soon...'
  },
  {
    slug: 'team-collaboration',
    title: 'Team Collaboration: Share Environments and Work Together',
    excerpt: 'Share environments with your team so multiple developers can connect and work with the same services simultaneously.',
    date: '2024-01-03',
    readTime: '6 min read',
    category: 'Workflow',
    featured: false,
    author: defaultAuthor,
    content: '# Team Collaboration\n\nContent coming soon...'
  },

  // Infrastructure & Platform
  {
    slug: 'vpn-gateway',
    title: 'VPN Gateway: Secure Access to Your Workspaces from Anywhere',
    excerpt: 'Securely access your workspaces and environment services from anywhere with encrypted VPN connections - no public endpoints needed.',
    date: '2023-12-30',
    readTime: '7 min read',
    category: 'Platform',
    featured: false,
    author: defaultAuthor,
    content: '# VPN Gateway\n\nContent coming soon...'
  },
  {
    slug: 'compute-storage',
    title: 'Compute & Storage: Dedicated Resources for Your Workloads',
    excerpt: 'Get dedicated CPU, memory, and persistent storage that survives restarts and keeps your data safe across workspace sessions.',
    date: '2023-12-28',
    readTime: '6 min read',
    category: 'Platform',
    featured: false,
    author: defaultAuthor,
    content: '# Compute & Storage\n\nContent coming soon...'
  },
  {
    slug: 'auto-stop',
    title: 'Auto Stop: Save Resources with Intelligent Idle Detection',
    excerpt: 'Idle machines stop automatically to save resources, then resume in seconds when you need them - no cold starts.',
    date: '2023-12-25',
    readTime: '5 min read',
    category: 'Platform',
    featured: false,
    author: defaultAuthor,
    content: '# Auto Stop\n\nContent coming soon...'
  },
  {
    slug: 'flexible-resources',
    title: 'Flexible Resources: Scale from 1 vCPU to 16 vCPU',
    excerpt: 'Choose from multiple machine sizes and scale up as your needs grow - from 1 vCPU to 16 vCPU and up to 64GB RAM.',
    date: '2023-12-22',
    readTime: '6 min read',
    category: 'Platform',
    featured: false,
    author: defaultAuthor,
    content: '# Flexible Resources\n\nContent coming soon...'
  },
  {
    slug: 'gpu-enabled',
    title: 'GPU Enabled: Accelerate AI/ML Workloads',
    excerpt: 'Run GPU-enabled nodes perfect for training models, running inference, and processing data for AI/ML development.',
    date: '2023-12-20',
    readTime: '8 min read',
    category: 'Platform',
    featured: false,
    author: defaultAuthor,
    content: '# GPU Enabled\n\nContent coming soon...'
  },
  {
    slug: 'performance-monitoring',
    title: 'Performance Monitoring: Real-Time Resource Metrics',
    excerpt: 'Track CPU, memory, and network usage in real-time. Get instant visibility into your machine performance at a glance.',
    date: '2023-12-18',
    readTime: '5 min read',
    category: 'Platform',
    featured: false,
    author: defaultAuthor,
    content: '# Performance Monitoring\n\nContent coming soon...'
  },
  {
    slug: 'high-availability',
    title: 'High Availability: Reliable Infrastructure with Automatic Failover',
    excerpt: 'Built on reliable infrastructure with automatic failover - your machines stay available when you need them most.',
    date: '2023-12-15',
    readTime: '7 min read',
    category: 'Platform',
    featured: false,
    author: defaultAuthor,
    content: '# High Availability\n\nContent coming soon...'
  },
  {
    slug: 'sub-30s-startup',
    title: 'Sub-30s Startup: No Waiting, No Builds, Just Code',
    excerpt: 'Workspace environments ready in under 30 seconds. No waiting, no builds, no setup - just start coding immediately.',
    date: '2023-12-12',
    readTime: '4 min read',
    category: 'Platform',
    featured: false,
    author: defaultAuthor,
    content: '# Sub-30s Startup\n\nContent coming soon...'
  },

  // Development Workflow
  {
    slug: 'zero-setup-development',
    title: 'Zero Setup: Every Developer Gets Production-Like Environments',
    excerpt: 'No more "works on my machine". Every developer gets the same production-like environment without any local setup.',
    date: '2023-12-10',
    readTime: '8 min read',
    category: 'Workflow',
    featured: true,
    author: defaultAuthor,
    content: '# Zero Setup Development\n\nContent coming soon...'
  },
  {
    slug: 'real-services-vs-mocks',
    title: 'Real Services vs Mocks: Why Mocks Waste Developer Time',
    excerpt: 'Connect to actual databases, queues, and APIs instead of mocks. No compromises, no surprises when you deploy.',
    date: '2023-12-08',
    readTime: '9 min read',
    category: 'Workflow',
    featured: false,
    author: defaultAuthor,
    content: '# Real Services vs Mocks\n\nContent coming soon...'
  },
  {
    slug: 'faster-feedback-loops',
    title: 'Faster Feedback Loops: Find Bugs Before Production',
    excerpt: 'Test against live data and find bugs before they reach production. Ship with confidence knowing your code works with real services.',
    date: '2023-12-05',
    readTime: '7 min read',
    category: 'Workflow',
    featured: false,
    author: defaultAuthor,
    content: '# Faster Feedback Loops\n\nContent coming soon...'
  },
  {
    slug: 'private-network-access',
    title: 'Private Network Access: Secure VPN to Your Environments',
    excerpt: 'Secure VPN connection to your environments - access internal services as if you were on the same network.',
    date: '2023-12-03',
    readTime: '6 min read',
    category: 'Workflow',
    featured: false,
    author: defaultAuthor,
    content: '# Private Network Access\n\nContent coming soon...'
  },
  {
    slug: 'workspace-environment-connections',
    title: 'Workspace Connections: DNS and Routing Handled Automatically',
    excerpt: 'Connect any workspace to access environment services by name - DNS resolution and routing handled automatically.',
    date: '2023-12-01',
    readTime: '5 min read',
    category: 'Workflow',
    featured: false,
    author: defaultAuthor,
    content: '# Workspace Environment Connections\n\nContent coming soon...'
  },
  {
    slug: 'clone-fork-environments',
    title: 'Clone & Fork: Create Isolated Testing Environments',
    excerpt: 'Create copies of environments for isolated testing. Fork production to debug without affecting users.',
    date: '2023-11-28',
    readTime: '7 min read',
    category: 'Workflow',
    featured: false,
    author: defaultAuthor,
    content: '# Clone & Fork Environments\n\nContent coming soon...'
  },
  {
    slug: 'live-monitoring',
    title: 'Live Monitoring: Service Health, Logs, and Metrics in Real-Time',
    excerpt: 'Monitor service health, logs, and metrics in real-time. Get instant visibility into your environment status.',
    date: '2023-11-25',
    readTime: '6 min read',
    category: 'Workflow',
    featured: false,
    author: defaultAuthor,
    content: '# Live Monitoring\n\nContent coming soon...'
  },
  {
    slug: 'instant-deployment',
    title: 'Instant Deployment: Push Changes Without Waiting',
    excerpt: 'Deploy changes instantly without waiting. Push new images and see them running in seconds, not minutes.',
    date: '2023-11-22',
    readTime: '5 min read',
    category: 'Workflow',
    featured: false,
    author: defaultAuthor,
    content: '# Instant Deployment\n\nContent coming soon...'
  },

  // Problem/Solution Narrative
  {
    slug: 'distributed-apps-localhost-problem',
    title: 'The Problem: Why Localhost Development Fails for Microservices',
    excerpt: 'Modern applications are distributed across microservices, databases, queues, and APIs. But developers still code on localhost, disconnected from reality.',
    date: '2023-11-20',
    readTime: '10 min read',
    category: 'Philosophy',
    featured: false,
    author: defaultAuthor,
    content: '# The Problem with Localhost Development\n\nContent coming soon...'
  },
  {
    slug: 'mocks-dont-match-production',
    title: 'The Gap: Docker Compose is Slow and Mocks Behave Differently',
    excerpt: 'Docker Compose is slow. Mocked services behave differently than real ones. By the time you find bugs in staging, you have wasted hours.',
    date: '2023-11-18',
    readTime: '9 min read',
    category: 'Philosophy',
    featured: false,
    author: defaultAuthor,
    content: '# The Gap: Mocks Don\'t Match Production\n\nContent coming soon...'
  },
  {
    slug: 'cloud-dev-environments-solution',
    title: 'The Solution: Cloud Workspaces Connected to Real Services',
    excerpt: 'Kloudlite gives you cloud-hosted workspaces that connect directly to staging, QA, or production - no mocks, no waiting.',
    date: '2023-11-15',
    readTime: '11 min read',
    category: 'Philosophy',
    featured: false,
    author: defaultAuthor,
    content: '# The Solution: Cloud Dev Environments\n\nContent coming soon...'
  },

  // Philosophy & Values
  {
    slug: 'speed-above-all',
    title: 'Speed Above All: How We Obsess Over Reducing Latency',
    excerpt: 'Every millisecond matters. From workspace startup (<30s) to service intercepts (instant), we obsess over reducing latency at every step.',
    date: '2023-11-12',
    readTime: '8 min read',
    category: 'Philosophy',
    featured: false,
    author: defaultAuthor,
    content: '# Speed Above All\n\nContent coming soon...'
  },
  {
    slug: 'zero-configuration',
    title: 'Zero Configuration: No YAML Hell, No DevOps Degree Required',
    excerpt: 'Developers shouldn\'t need a degree in DevOps to write code. Our tools work out of the box - no YAML hell, no infrastructure expertise.',
    date: '2023-11-10',
    readTime: '7 min read',
    category: 'Philosophy',
    featured: false,
    author: defaultAuthor,
    content: '# Zero Configuration\n\nContent coming soon...'
  },
  {
    slug: 'open-by-default',
    title: 'Open by Default: Building in Public, Welcoming Contributions',
    excerpt: 'Our core platform is open source and always will be. We build in public, welcome contributions, and believe transparency creates better software.',
    date: '2023-11-08',
    readTime: '6 min read',
    category: 'Philosophy',
    featured: false,
    author: defaultAuthor,
    content: '# Open by Default\n\nContent coming soon...'
  },

  // Architecture Deep Dive
  {
    slug: 'kubernetes-native-architecture',
    title: 'Kubernetes-Native Architecture: Built on CRDs and Controllers',
    excerpt: 'Built on Kubernetes with custom CRDs and controllers. Your workspaces and environments are declared as resources and reconciled automatically.',
    date: '2023-11-05',
    readTime: '12 min read',
    category: 'Architecture',
    featured: false,
    author: defaultAuthor,
    content: '# Kubernetes-Native Architecture\n\nContent coming soon...'
  },
  {
    slug: 'custom-resource-definitions',
    title: 'Custom Resource Definitions: Workspace, Environment, ServiceIntercept',
    excerpt: 'Deep dive into Kloudlite CRDs - how Workspace, Environment, and ServiceIntercept resources work under the hood.',
    date: '2023-11-03',
    readTime: '15 min read',
    category: 'Architecture',
    featured: false,
    author: defaultAuthor,
    content: '# Custom Resource Definitions\n\nContent coming soon...'
  },
  {
    slug: 'kubernetes-controllers',
    title: 'Kubernetes Controllers: Reconciling Desired State with Infrastructure',
    excerpt: 'Learn how Kloudlite controllers watch CRDs and reconcile desired state with actual infrastructure using Kubernetes controller patterns.',
    date: '2023-11-01',
    readTime: '14 min read',
    category: 'Architecture',
    featured: false,
    author: defaultAuthor,
    content: '# Kubernetes Controllers\n\nContent coming soon...'
  },
  {
    slug: 'service-mesh-integration',
    title: 'Service Mesh Integration: SOCAT-Based Traffic Forwarding',
    excerpt: 'How we use SOCAT-based traffic forwarding to implement service intercepts without complex service mesh installations.',
    date: '2023-10-28',
    readTime: '13 min read',
    category: 'Architecture',
    featured: false,
    author: defaultAuthor,
    content: '# Service Mesh Integration\n\nContent coming soon...'
  },
]

// Create a map for easy lookup by slug
export const blogPostsMap = blogPostsData.reduce((acc, post) => {
  acc[post.slug] = post
  return acc
}, {} as Record<string, BlogPost>)

// Get featured posts
export const featuredPosts = blogPostsData.filter(post => post.featured)

// Get regular posts
export const regularPosts = blogPostsData.filter(post => !post.featured)
