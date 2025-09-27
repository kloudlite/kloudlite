# Frontend Conventions (Next.js/React)

This document outlines the conventions and standards for the frontend Next.js application in Kloudlite v2.

## 📁 Directory Structure

### Application Layout
```
/web/
├── app/                    # Next.js App Router
│   ├── (auth)/            # Route groups
│   │   └── auth/          # Auth pages
│   ├── [teamSlug]/        # Dynamic routes
│   │   └── page.tsx       # Team dashboard
│   ├── actions/           # Server actions
│   │   ├── auth.ts
│   │   ├── teams.ts
│   │   └── notifications.ts
│   ├── api/               # API routes (minimal use)
│   │   └── auth/
│   │       └── [...nextauth]/
│   ├── layout.tsx         # Root layout
│   ├── page.tsx           # Home page
│   └── globals.css        # Global styles
├── components/
│   ├── ui/                # shadcn/ui components
│   │   ├── button.tsx
│   │   ├── card.tsx
│   │   └── ...
│   ├── auth/              # Auth components
│   │   ├── login-form.tsx
│   │   └── signup-form.tsx
│   ├── overview/          # Feature components
│   │   ├── teams-list.tsx
│   │   └── resource-card.tsx
│   └── {component}.tsx    # Shared components
├── lib/
│   ├── auth/              # Auth utilities
│   │   ├── get-auth-options.ts
│   │   └── grpc-client.ts
│   ├── grpc/              # gRPC clients
│   │   └── generated/     # Proto-generated code
│   └── utils.ts           # Utility functions
├── hooks/                 # Custom React hooks
│   └── use-debounce.ts
├── types/                 # TypeScript types
│   └── next-auth.d.ts
└── public/                # Static assets
```

## 📝 File Naming Conventions

### Components
- **Format**: PascalCase
- **Extension**: `.tsx`
- **Examples**:
  - `TeamSelector.tsx`
  - `NotificationBell.tsx`
  - `ResourceCard.tsx`

### Pages & Layouts
- **Format**: lowercase
- **Files**: `page.tsx`, `layout.tsx`, `loading.tsx`, `error.tsx`
- **Examples**:
  - `/app/auth/login/page.tsx`
  - `/app/teams/[teamId]/page.tsx`

### Server Actions
- **Format**: lowercase with hyphens
- **Extension**: `.ts`
- **Examples**:
  - `team-approvals.ts`
  - `notifications.ts`

### Utilities & Hooks
- **Utilities**: lowercase with hyphens (`format-error.ts`)
- **Hooks**: camelCase with `use` prefix (`useDebounce.ts`)

## 🏗️ Component Architecture

### Server vs Client Components

#### Server Components (Default)
```tsx
// ✅ Default for pages and layouts
// app/overview/page.tsx
import { getServerSession } from "next-auth"
import { listUserTeams } from "@/app/actions/teams"

export default async function OverviewPage() {
  const session = await getServerSession()
  const teams = await listUserTeams()
  
  return (
    <div>
      <TeamsList teams={teams} />
    </div>
  )
}
```

#### Client Components
```tsx
// ✅ Only when needed for interactivity
// components/theme-toggle.tsx
"use client"

import { useState } from "react"

export function ThemeToggle() {
  const [theme, setTheme] = useState("light")
  
  return (
    <button onClick={() => setTheme(theme === "light" ? "dark" : "light")}>
      Toggle theme
    </button>
  )
}
```

### Component Patterns

#### 1. Composition Pattern
```tsx
// ✅ Prefer composition
export function Card({ children, className }: CardProps) {
  return (
    <div className={cn("rounded-lg border", className)}>
      {children}
    </div>
  )
}

export function CardHeader({ children }: { children: React.ReactNode }) {
  return <div className="p-6 pb-0">{children}</div>
}

// Usage
<Card>
  <CardHeader>Title</CardHeader>
  <CardContent>Content</CardContent>
</Card>
```

#### 2. Props Interface Pattern
```tsx
// ✅ Use interfaces for props
interface TeamSelectorProps {
  teams: Team[]
  selectedTeam: string | null
  onTeamSelect: (teamId: string | null) => void
  onCreateTeam?: () => void
}

export function TeamSelector({
  teams,
  selectedTeam,
  onTeamSelect,
  onCreateTeam = () => window.location.href = '/teams/new'
}: TeamSelectorProps) {
  // Component logic
}
```

## 🎨 Styling Conventions

### Tailwind CSS Usage
```tsx
// ✅ Use cn() utility for conditional classes
import { cn } from "@/lib/utils"

<button
  className={cn(
    "base-classes px-4 py-2",
    isActive && "active-classes",
    isDisabled && "disabled-classes"
  )}
/>

// ✅ Mobile-first responsive design
<div className="text-xs md:text-sm lg:text-base">
  Responsive text
</div>

// ✅ Use CSS variables for theming
<div className="bg-background text-foreground">
  Themed content
</div>
```

### Component Styling Patterns
```tsx
// Background gradients
<div className="absolute inset-0 gradient-primary -z-10" />

// Cards with blur effect
<Card className="border-border/50 bg-card/50 backdrop-blur">

// Headers with specific font weight
<h1 className="text-2xl font-extralight tracking-tight">

// Status indicators
<div className="h-2 w-2 rounded-full bg-green-500" /> // Active
<div className="h-2 w-2 rounded-full bg-yellow-500" /> // Pending
<div className="h-2 w-2 rounded-full bg-gray-400" /> // Inactive
```

## 📡 Data Fetching Patterns

### Server Actions (Preferred)
```typescript
// app/actions/teams.ts
'use server'

import { revalidatePath } from 'next/cache'

export async function createTeam(data: CreateTeamInput) {
  // 1. Validate session
  const session = await getServerSession()
  if (!session) throw new Error('Unauthorized')
  
  // 2. Call gRPC service
  const client = getAccountsClient()
  const metadata = await getAuthMetadata()
  
  const result = await new Promise((resolve, reject) => {
    client.createTeam(data, metadata, (error, response) => {
      if (error) reject(error)
      else resolve(response)
    })
  })
  
  // 3. Revalidate affected paths
  revalidatePath('/teams')
  revalidatePath('/overview')
  
  // 4. Return result
  return result
}
```

### Using Server Actions in Components
```tsx
// Client component
"use client"

import { createTeam } from '@/app/actions/teams'
import { useRouter } from 'next/navigation'

export function CreateTeamForm() {
  const router = useRouter()
  
  async function handleSubmit(formData: FormData) {
    try {
      const result = await createTeam({
        slug: formData.get('slug') as string,
        displayName: formData.get('displayName') as string,
      })
      
      if (result.pending) {
        router.push('/overview?teamPending=true')
      } else {
        router.push(`/${result.teamId}`)
      }
    } catch (error) {
      // Handle error
    }
  }
  
  return <form action={handleSubmit}>...</form>
}
```

## 🔄 State Management

### URL State
```tsx
// Use searchParams for filters/pagination
export default function TeamsPage({
  searchParams
}: {
  searchParams: Promise<{ page?: string; search?: string }>
}) {
  const params = await searchParams
  const page = params.page || '1'
  const search = params.search || ''
  
  // Fetch with filters
  const teams = await listTeams({ page, search })
}

// Client-side URL updates
const router = useRouter()
const searchParams = useSearchParams()

function updateFilter(key: string, value: string) {
  const params = new URLSearchParams(searchParams)
  params.set(key, value)
  router.push(`?${params.toString()}`)
}
```

### Local State
```tsx
// Simple state for UI interactions
const [isOpen, setIsOpen] = useState(false)
const [search, setSearch] = useState('')

// Complex state with reducer
const [state, dispatch] = useReducer(formReducer, initialState)
```

## 🔐 Authentication Patterns

### Protected Routes
```tsx
// Middleware protection
export function middleware(request: NextRequest) {
  const token = request.cookies.get('next-auth.session-token')
  
  if (!token && request.nextUrl.pathname.startsWith('/overview')) {
    return NextResponse.redirect(new URL('/auth/login', request.url))
  }
}
```

### Session Usage
```tsx
// Server component
const session = await getServerSession(authOptions)
if (!session) {
  redirect('/auth/login')
}

// Client component
import { useSession } from 'next-auth/react'

export function UserProfile() {
  const { data: session, status } = useSession()
  
  if (status === 'loading') return <Skeleton />
  if (!session) return null
  
  return <div>Welcome {session.user.name}</div>
}
```

## 🧩 Component Organization

### Feature-Based Structure
```
components/
├── auth/
│   ├── login-form.tsx
│   ├── signup-form.tsx
│   └── reset-password-form.tsx
├── teams/
│   ├── team-card.tsx
│   ├── team-list.tsx
│   └── team-selector.tsx
└── shared/
    ├── loading-spinner.tsx
    └── error-boundary.tsx
```

### Component Composition
```tsx
// Parent component handles data
export async function TeamsPage() {
  const teams = await listTeams()
  return <TeamsList teams={teams} />
}

// Child component is presentational
export function TeamsList({ teams }: { teams: Team[] }) {
  return (
    <div className="grid gap-4">
      {teams.map(team => (
        <TeamCard key={team.id} team={team} />
      ))}
    </div>
  )
}
```

## 🎯 TypeScript Best Practices

### Type Definitions
```typescript
// Define types in relevant files or types/ directory
interface Team {
  id: string
  slug: string
  displayName: string
  description?: string
  status: 'active' | 'inactive' | 'pending'
  role: 'owner' | 'admin' | 'member'
}

// Use type inference when obvious
const [search, setSearch] = useState('') // string inferred

// Be explicit for complex types
const [filters, setFilters] = useState<FilterOptions>({
  status: 'all',
  role: null
})
```

### Props Typing
```tsx
// Explicit prop types
interface ButtonProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: 'default' | 'destructive' | 'outline'
  size?: 'sm' | 'default' | 'lg'
  asChild?: boolean
}

// Children prop pattern
interface LayoutProps {
  children: React.ReactNode
}
```

## 🚀 Performance Optimization

### Image Optimization
```tsx
import Image from 'next/image'

<Image
  src="/logo.png"
  alt="Logo"
  width={120}
  height={40}
  priority // For above-fold images
/>
```

### Code Splitting
```tsx
// Dynamic imports for heavy components
const HeavyComponent = dynamic(() => import('./heavy-component'), {
  loading: () => <Skeleton />,
  ssr: false // If client-only
})
```

### Memoization
```tsx
// Memoize expensive computations
const expensiveValue = useMemo(() => {
  return computeExpensiveValue(data)
}, [data])

// Memoize callbacks
const handleClick = useCallback(() => {
  doSomething(id)
}, [id])
```

## 🧪 Error Handling

### Error Boundaries
```tsx
// app/error.tsx
'use client'

export default function Error({
  error,
  reset,
}: {
  error: Error & { digest?: string }
  reset: () => void
}) {
  return (
    <div className="flex min-h-screen items-center justify-center">
      <div className="text-center">
        <h2>Something went wrong!</h2>
        <button onClick={reset}>Try again</button>
      </div>
    </div>
  )
}
```

### Form Error Handling
```tsx
const [error, setError] = useState<string | null>(null)

async function handleSubmit(formData: FormData) {
  try {
    setError(null)
    await createTeam(formData)
  } catch (error) {
    setError(error instanceof Error ? error.message : 'An error occurred')
  }
}
```

## 📋 Best Practices

1. **Server-first approach** - Use Server Components by default
2. **Minimize client JavaScript** - Only use client components when necessary
3. **Type everything** - Leverage TypeScript for safety
4. **Mobile-first design** - Start with mobile, enhance for desktop
5. **Accessible by default** - Use semantic HTML and ARIA attributes
6. **Performance matters** - Optimize images, lazy load, code split
7. **Consistent styling** - Use design system tokens and patterns
8. **Error boundaries** - Handle errors gracefully at every level