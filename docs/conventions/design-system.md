# UI/UX Design System

This document defines the visual design standards, component guidelines, and interaction patterns for Kloudlite v2.

## 🎨 Design Principles

### Core Principles
1. **Clarity First** - Clean, uncluttered interfaces
2. **Mobile-First** - Design for small screens, enhance for larger
3. **Consistent** - Uniform patterns across the application
4. **Accessible** - WCAG 2.1 AA compliant
5. **Performance** - Fast loading, smooth interactions

## 🎭 Visual Identity

### Logo Usage
```tsx
import { KloudliteLogo } from "@/components/kloudlite-logo"

// Standard usage
<KloudliteLogo className="h-6 w-auto" />

// Responsive sizing
<KloudliteLogo className="h-5 md:h-6 lg:h-7" />
```

### Typography

#### Font Family
- **System Font Stack**: Uses native system fonts for performance
```css
font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, 
             "Helvetica Neue", Arial, sans-serif;
```

#### Type Scale
```tsx
// Headings
<h1 className="text-2xl font-extralight tracking-tight md:text-3xl lg:text-4xl">
  Page Title
</h1>

<h2 className="text-xl font-light md:text-2xl">
  Section Header
</h2>

<h3 className="text-lg font-normal md:text-xl">
  Subsection
</h3>

// Body text
<p className="text-sm text-muted-foreground md:text-base">
  Regular paragraph text
</p>

// Small text
<span className="text-xs text-muted-foreground">
  Caption or helper text
</span>
```

#### Font Weights
- **Extralight (200)**: Main headings
- **Light (300)**: Secondary headings
- **Normal (400)**: Body text
- **Medium (500)**: Buttons, emphasis
- **Semibold (600)**: Important labels

## 🎨 Color System

### Semantic Colors
All colors use CSS variables for theme support:

```css
/* Light mode (default) */
--background: 0 0% 100%;
--foreground: 240 10% 3.9%;
--card: 0 0% 100%;
--card-foreground: 240 10% 3.9%;
--primary: 240 5.9% 10%;
--primary-foreground: 0 0% 98%;
--secondary: 240 4.8% 95.9%;
--secondary-foreground: 240 5.9% 10%;
--muted: 240 4.8% 95.9%;
--muted-foreground: 240 3.8% 46.1%;
--accent: 240 4.8% 95.9%;
--accent-foreground: 240 5.9% 10%;
--destructive: 0 84.2% 60.2%;
--destructive-foreground: 0 0% 98%;
--border: 240 5.9% 90%;
--input: 240 5.9% 90%;
--ring: 240 5.9% 10%;
```

### Status Colors
```tsx
// Success - Green
<div className="bg-green-500" />
<div className="text-green-600" />
<div className="border-green-500" />

// Warning - Yellow/Amber
<div className="bg-yellow-500" />
<div className="text-yellow-600" />
<div className="border-yellow-500" />

// Error - Red
<div className="bg-red-500" />
<div className="text-red-600" />
<div className="border-red-500" />

// Info - Blue
<div className="bg-blue-500" />
<div className="text-blue-600" />
<div className="border-blue-500" />

// Inactive - Gray
<div className="bg-gray-400" />
<div className="text-gray-500" />
<div className="border-gray-400" />
```

### Gradient Patterns
```tsx
// Primary gradient overlay
<div className="absolute inset-0 gradient-primary -z-10" />

// Subtle background gradient
<div className="bg-gradient-to-br from-primary/5 via-transparent to-primary/5" />

// Card hover gradient
<div className="hover:bg-gradient-to-br hover:from-primary/5 hover:to-primary/10" />
```

## 🗂️ Component Patterns

### Cards
```tsx
// Standard card
<Card className="border-border/50 bg-card/50 backdrop-blur">
  <CardHeader>
    <CardTitle>Title</CardTitle>
    <CardDescription>Description</CardDescription>
  </CardHeader>
  <CardContent>Content</CardContent>
</Card>

// Interactive card
<Card className="cursor-pointer transition-all hover:shadow-md hover:border-border">
  {/* Content */}
</Card>

// Status card
<Card className="relative overflow-hidden">
  <div className="absolute inset-x-0 top-0 h-1 bg-green-500" />
  {/* Content */}
</Card>
```

### Buttons

#### Variants
```tsx
// Primary (default)
<Button>Create Team</Button>

// Secondary
<Button variant="secondary">Cancel</Button>

// Outline
<Button variant="outline">Learn More</Button>

// Destructive
<Button variant="destructive">Delete</Button>

// Ghost
<Button variant="ghost">
  <Settings className="h-4 w-4" />
</Button>

// Link
<Button variant="link">View Details</Button>
```

#### Sizes
```tsx
<Button size="sm">Small</Button>
<Button>Default</Button>
<Button size="lg">Large</Button>
<Button size="icon">
  <Plus className="h-4 w-4" />
</Button>
```

#### States
```tsx
// Loading
<Button disabled>
  <Loader2 className="mr-2 h-4 w-4 animate-spin" />
  Creating...
</Button>

// With icon
<Button>
  <Plus className="mr-2 h-4 w-4" />
  Add Item
</Button>
```

### Forms

#### Input Fields
```tsx
<div className="space-y-2">
  <Label htmlFor="email">Email</Label>
  <Input
    id="email"
    type="email"
    placeholder="name@example.com"
    className="w-full"
  />
  <p className="text-xs text-muted-foreground">
    We'll never share your email
  </p>
</div>

// With error
<div className="space-y-2">
  <Label htmlFor="slug">Team Slug</Label>
  <Input
    id="slug"
    aria-invalid={!!error}
    className={cn(error && "border-destructive")}
  />
  {error && (
    <p className="text-xs text-destructive">{error}</p>
  )}
</div>
```

#### Form Layout
```tsx
<form className="space-y-6">
  <div className="space-y-4">
    <div className="space-y-2">
      <Label>Field 1</Label>
      <Input />
    </div>
    
    <div className="space-y-2">
      <Label>Field 2</Label>
      <Textarea />
    </div>
  </div>
  
  <div className="flex gap-4">
    <Button type="submit">Submit</Button>
    <Button variant="outline" type="button">Cancel</Button>
  </div>
</form>
```

### Navigation

#### Header Pattern
```tsx
<header className="sticky top-0 z-50 border-b bg-background/80 backdrop-blur">
  <div className="container flex h-14 items-center">
    <KloudliteLogo />
    <nav className="ml-8 hidden md:flex gap-6">
      {/* Nav items */}
    </nav>
    <div className="ml-auto flex items-center gap-4">
      <NotificationBell />
      <UserMenu />
    </div>
  </div>
</header>
```

#### Sidebar Pattern
```tsx
<aside className="hidden lg:block w-64 border-r bg-muted/50">
  <nav className="space-y-1 p-4">
    <Link
      href="/dashboard"
      className={cn(
        "flex items-center gap-3 rounded-md px-3 py-2",
        "hover:bg-accent hover:text-accent-foreground",
        isActive && "bg-accent text-accent-foreground"
      )}
    >
      <Home className="h-4 w-4" />
      Dashboard
    </Link>
  </nav>
</aside>
```

### Data Display

#### Tables
```tsx
<Table>
  <TableHeader>
    <TableRow>
      <TableHead>Name</TableHead>
      <TableHead>Status</TableHead>
      <TableHead>Actions</TableHead>
    </TableRow>
  </TableHeader>
  <TableBody>
    <TableRow>
      <TableCell>Item Name</TableCell>
      <TableCell>
        <Badge variant="success">Active</Badge>
      </TableCell>
      <TableCell>
        <Button variant="ghost" size="sm">Edit</Button>
      </TableCell>
    </TableRow>
  </TableBody>
</Table>
```

#### Lists
```tsx
// Simple list
<div className="divide-y">
  {items.map(item => (
    <div key={item.id} className="py-4">
      <h4 className="font-medium">{item.name}</h4>
      <p className="text-sm text-muted-foreground">{item.description}</p>
    </div>
  ))}
</div>

// Card list
<div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
  {items.map(item => (
    <Card key={item.id}>
      {/* Card content */}
    </Card>
  ))}
</div>
```

### Feedback

#### Alerts
```tsx
// Success
<Alert variant="success">
  <CheckCircle className="h-4 w-4" />
  <AlertDescription>
    Operation completed successfully
  </AlertDescription>
</Alert>

// Warning
<Alert variant="warning">
  <AlertTriangle className="h-4 w-4" />
  <AlertDescription>
    This action requires approval
  </AlertDescription>
</Alert>

// Error
<Alert variant="destructive">
  <XCircle className="h-4 w-4" />
  <AlertDescription>
    An error occurred
  </AlertDescription>
</Alert>
```

#### Loading States
```tsx
// Full page
<div className="flex min-h-screen items-center justify-center">
  <Loader2 className="h-8 w-8 animate-spin text-primary" />
</div>

// Inline
<div className="flex items-center gap-2 text-muted-foreground">
  <Loader2 className="h-4 w-4 animate-spin" />
  <span>Loading...</span>
</div>

// Skeleton
<div className="space-y-4">
  <Skeleton className="h-4 w-3/4" />
  <Skeleton className="h-4 w-1/2" />
</div>
```

## 🎯 Interaction Patterns

### Hover Effects
```tsx
// Subtle hover
className="hover:bg-accent hover:text-accent-foreground"

// Shadow hover
className="transition-shadow hover:shadow-md"

// Border hover
className="border-transparent hover:border-border"

// Scale hover
className="transition-transform hover:scale-105"
```

### Focus States
```tsx
// Focus visible
className="focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary"

// Focus within
className="focus-within:ring-2 focus-within:ring-primary"
```

### Transitions
```tsx
// Standard timing
className="transition-all duration-200"

// Fast interactions
className="transition-colors duration-150"

// Smooth animations
className="transition-all duration-300 ease-in-out"
```

## 📱 Responsive Design

### Breakpoints
```css
/* Mobile: 0-639px (default) */
/* Tablet: 640px-1023px (sm:) */
/* Desktop: 1024px+ (lg:) */
/* Wide: 1280px+ (xl:) */
```

### Responsive Patterns
```tsx
// Text sizing
<h1 className="text-xl sm:text-2xl lg:text-3xl">

// Spacing
<div className="p-4 sm:p-6 lg:p-8">

// Grid layouts
<div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">

// Show/hide elements
<div className="hidden sm:block">Desktop only</div>
<div className="sm:hidden">Mobile only</div>
```

## ♿ Accessibility

### ARIA Patterns
```tsx
// Loading states
<Button disabled aria-busy="true">
  <span className="sr-only">Loading</span>
  <Loader2 className="animate-spin" />
</Button>

// Form validation
<Input
  aria-invalid={!!error}
  aria-describedby={error ? "error-message" : undefined}
/>
{error && <p id="error-message" role="alert">{error}</p>}

// Navigation
<nav aria-label="Main navigation">
  <ul role="list">
    <li><a href="/">Home</a></li>
  </ul>
</nav>
```

### Keyboard Navigation
- All interactive elements must be keyboard accessible
- Use proper focus indicators
- Implement skip links for main content
- Support Escape key for dismissible elements

## 🌙 Dark Mode

### Theme Toggle
```tsx
<ThemeToggle />

// CSS variables automatically adjust
className="bg-background text-foreground"
```

### Dark Mode Specific Styles
```tsx
// Only when needed
className="dark:bg-gray-800 dark:border-gray-700"

// Prefer semantic colors that adapt automatically
className="bg-card border-border"
```

## 📏 Spacing System

### Spacing Scale
```
0.5 = 0.125rem = 2px
1   = 0.25rem  = 4px
2   = 0.5rem   = 8px
3   = 0.75rem  = 12px
4   = 1rem     = 16px
6   = 1.5rem   = 24px
8   = 2rem     = 32px
12  = 3rem     = 48px
16  = 4rem     = 64px
```

### Common Patterns
```tsx
// Card padding
className="p-4 sm:p-6"

// Section spacing
className="space-y-6 sm:space-y-8"

// Grid gaps
className="gap-4 sm:gap-6"

// Margin between sections
className="mt-8 sm:mt-12"
```

## 🚀 Performance Guidelines

1. **Optimize Images** - Use Next.js Image component
2. **Lazy Load** - Heavy components and below-fold content
3. **Minimize Animations** - Use CSS transforms over position changes
4. **Reduce Layout Shifts** - Set explicit dimensions
5. **Bundle Size** - Import only needed components