# @kloudlite/ui

A comprehensive React component library for Kloudlite applications, built on Radix UI primitives with Tailwind CSS 4 styling. This package provides 47+ accessible, customizable UI components designed for use across the console, dashboard, and web applications.

## Features

- **Accessibility First**: Built on Radix UI primitives with full keyboard navigation and ARIA support
- **Fully Customizable**: Tailwind CSS 4 with CSS variables for easy theming
- **TypeScript**: Full type safety with comprehensive TypeScript definitions
- **Dark Mode**: Built-in dark mode support with theme-aware styling
- **Form Integration**: Seamless integration with react-hook-form and zod validation
- **Responsive**: Mobile-first design with responsive variants
- **Composable**: Components designed to work together with consistent patterns

## Installation

The package is already included in the Kloudlite monorepo as a workspace dependency. No additional installation is needed within the monorepo.

### Dependencies

Peer dependencies (must be installed in consuming applications):

```json
{
  "next": "^16.0.0",
  "react": "^19.0.0",
  "react-dom": "^19.0.0",
  "react-hook-form": "^7.71.1"
}
```

## Usage

Import components directly from `@kloudlite/ui`:

```tsx
import { Button, Card, Input } from '@kloudlite/ui'

export function MyComponent() {
  return (
    <Card>
      <CardHeader>
        <CardTitle>Welcome</CardTitle>
        <CardDescription>Get started with Kloudlite</CardDescription>
      </CardHeader>
      <CardContent>
        <Input placeholder="Enter your email" />
        <Button variant="default">Get Started</Button>
      </CardContent>
    </Card>
  )
}
```

## Styling Approach

### Tailwind CSS 4 Components

The package uses Tailwind CSS 4 with inline `@theme` configuration. All colors are defined as CSS variables for easy theming:

```css
--background: oklch(0.995 0 0);
--foreground: oklch(0.15 0.01 235);
--primary: oklch(0.548 0.204 251.6);
--primary-foreground: oklch(0.99 0 0);
/* ... and more */
```

### Dark Mode

Dark mode is supported via the `.dark` class. Components automatically adapt when this class is present on a parent element.

### Custom Styling

All components accept a `className` prop for additional styling:

```tsx
<Button className="w-full">Full Width Button</Button>
```

### Variants

Many components use `class-variance-authority` for variant styling:

```tsx
<Button variant="default">Default</Button>
<Button variant="destructive">Destructive</Button>
<Button variant="outline">Outline</Button>
<Button variant="secondary">Secondary</Button>
<Button variant="ghost">Ghost</Button>
<Button variant="link">Link</Button>
```

### Custom Utility

Use the `cn()` utility from `@kloudlite/lib` for conditional class merging:

```tsx
import { cn } from '@kloudlite/lib'

const className = cn(
  'base-class',
  isActive && 'active-class',
  isDisabled && 'opacity-50'
)
```

## Component Patterns

### forwardRef Pattern

All interactive components use `React.forwardRef` for ref forwarding:

```tsx
const Button = React.forwardRef<HTMLButtonElement, ButtonProps>(
  ({ className, variant, size, ...props }, ref) => {
    return (
      <button
        ref={ref}
        className={cn(buttonVariants({ variant, size, className }))}
        {...props}
      />
    )
  }
)
Button.displayName = "Button"
```

### asChild Pattern

Some components support the `asChild` pattern using Radix UI Slot:

```tsx
<Button asChild>
  <a href="/dashboard">Go to Dashboard</a>
</Button>
```

## Available Components

### Layout & Containers

- **Card**: Container with header, content, and footer sections
- **Separator**: Visual divider between content
- **ScrollArea**: Custom scrollable container
- **Resizable**: Resizable panels using react-resizable-panels
- **Sidebar**: Navigation sidebar with collapsible sections

### Typography & Text

- **Label**: Form label component
- **Kbd**: Keyboard shortcut styling
- **Badge**: Status indicators and tags

### Buttons & Actions

- **Button**: Primary action button with multiple variants
- **ButtonGroup**: Grouped buttons
- **IconButton**: Icon-only button (Button with size="icon")
- **Toggle**: Toggle switch component
- **ToggleGroup**: Group of toggle buttons

### Forms & Inputs

- **Input**: Text input field
- **Textarea**: Multi-line text input
- **Select**: Dropdown select component
- **Checkbox**: Checkbox input
- **RadioGroup**: Radio button group
- **Switch**: Toggle switch for boolean values
- **InputOtp**: OTP input with automatic focus management
- **InputGroup**: Input with icon/text prefix/suffix
- **Field**: Advanced form field with label, description, and error handling
- **Form**: Form components for react-hook-form integration

### Navigation

- **Breadcrumb**: Navigation breadcrumbs
- **Tabs**: Tabbed content with animated underline
- **Pagination**: Pagination controls
- **Menubar**: Application menu bar
- **NavigationMenu**: Navigation menu with dropdowns
- **Sidebar**: Sidebar navigation component

### Overlays & Modals

- **Dialog**: Modal dialog with overlay
- **Drawer**: Slide-out drawer panel
- **Sheet**: Sheet panel (side drawer)
- **AlertDialog**: Alert dialog for confirmations
- **Popover**: Popover content
- **DropdownMenu**: Dropdown menu with actions
- **ContextMenu**: Right-click context menu
- **HoverCard**: Card shown on hover

### Feedback & Indicators

- **Alert**: Alert banners for messages
- **AlertDialog**: Alert dialog
- **Progress**: Progress bar
- **Slider**: Range slider
- **Skeleton**: Loading placeholder
- **Spinner**: Loading spinner
- **Empty**: Empty state component
- **Badge**: Status badges

### Data Display

- **Table**: Data table with header, body, footer
- **Avatar**: User avatar with fallback
- **Carousel**: Image/content carousel
- **Chart**: Chart components using Recharts
- **Accordion**: Collapsible accordion sections
- **Collapsible**: Collapsible content
- **Tooltip**: Tooltip on hover
- **HoverCard**: Hoverable card
- **AspectRatio**: Aspect ratio container
- **Command**: Command palette / command menu

### Advanced Components

- **Calendar**: Date picker calendar
- **Command**: Command palette with search
- **ErrorBoundary**: Error boundary for catching React errors
- **GlowingStars**: Animated glowing stars effect
- **KloudliteLogo**: Kloudlite logo component
- **ThemeSwitcher**: Dark/light theme toggle
- **Sonner**: Toast notifications (wrapper around sonner)

## Component Examples

### Button

```tsx
import { Button } from '@kloudlite/ui'

// Variants
<Button variant="default">Default</Button>
<Button variant="destructive">Delete</Button>
<Button variant="outline">Outline</Button>
<Button variant="secondary">Secondary</Button>
<Button variant="ghost">Ghost</Button>
<Button variant="link">Link</Button>

// Sizes
<Button size="default">Default</Button>
<Button size="sm">Small</Button>
<Button size="lg">Large</Button>
<Button size="icon">
  <PlusIcon />
</Button>

// With icon
<Button>
  <PlusIcon className="w-4 h-4" />
  Add New
</Button>

// Disabled
<Button disabled>Disabled</Button>

// asChild pattern
<Button asChild>
  <a href="/new">Create New</a>
</Button>
```

### Card

```tsx
import { Card, CardHeader, CardTitle, CardDescription, CardContent, CardFooter } from '@kloudlite/ui'

<Card>
  <CardHeader>
    <CardTitle>Card Title</CardTitle>
    <CardDescription>Card description goes here</CardDescription>
  </CardHeader>
  <CardContent>
    <p>Card content goes here</p>
  </CardContent>
  <CardFooter>
    <Button>Action</Button>
  </CardFooter>
</Card>
```

### Form with Validation

```tsx
import { Form, FormField, FormItem, FormLabel, FormMessage } from '@kloudlite/ui'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'

const formSchema = z.object({
  username: z.string().min(2, 'Username must be at least 2 characters'),
  email: z.string().email('Invalid email address'),
})

export function MyForm() {
  const form = useForm({
    resolver: zodResolver(formSchema),
    defaultValues: {
      username: '',
      email: '',
    },
  })

  return (
    <Form {...form}>
      <form onSubmit={form.handleSubmit(onSubmit)}>
        <FormField
          control={form.control}
          name="username"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Username</FormLabel>
              <Input {...field} />
              <FormMessage />
            </FormItem>
          )}
        />
        <FormField
          control={form.control}
          name="email"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Email</FormLabel>
              <Input type="email" {...field} />
              <FormMessage />
            </FormItem>
          )}
        />
        <Button type="submit">Submit</Button>
      </form>
    </Form>
  )
}
```

### Dialog

```tsx
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle, DialogTrigger } from '@kloudlite/ui'

export function MyDialog() {
  return (
    <Dialog>
      <DialogTrigger asChild>
        <Button>Open Dialog</Button>
      </DialogTrigger>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Dialog Title</DialogTitle>
          <DialogDescription>
            Dialog description goes here
          </DialogDescription>
        </DialogHeader>
        <div>Dialog content</div>
        <DialogFooter>
          <Button variant="outline">Cancel</Button>
          <Button>Confirm</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
```

### Tabs

```tsx
import { Tabs, TabsList, TabsTrigger, TabsContent } from '@kloudlite/ui'

export function MyTabs() {
  return (
    <Tabs defaultValue="tab1">
      <TabsList>
        <TabsTrigger value="tab1">Tab 1</TabsTrigger>
        <TabsTrigger value="tab2">Tab 2</TabsTrigger>
        <TabsTrigger value="tab3">Tab 3</TabsTrigger>
      </TabsList>
      <TabsContent value="tab1">
        Content for tab 1
      </TabsContent>
      <TabsContent value="tab2">
        Content for tab 2
      </TabsContent>
      <TabsContent value="tab3">
        Content for tab 3
      </TabsContent>
    </Tabs>
  )
}
```

### Select

```tsx
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@kloudlite/ui'

export function MySelect() {
  return (
    <Select>
      <SelectTrigger>
        <SelectValue placeholder="Select an option" />
      </SelectTrigger>
      <SelectContent>
        <SelectItem value="option1">Option 1</SelectItem>
        <SelectItem value="option2">Option 2</SelectItem>
        <SelectItem value="option3">Option 3</SelectItem>
      </SelectContent>
    </Select>
  )
}
```

### Alert

```tsx
import { Alert, AlertDescription, AlertTitle } from '@kloudlite/ui'
import { AlertCircle } from 'lucide-react'

export function MyAlert() {
  return (
    <Alert>
      <AlertCircle className="h-4 w-4" />
      <AlertTitle>Error</AlertTitle>
      <AlertDescription>
        Something went wrong. Please try again.
      </AlertDescription>
    </Alert>
  )
}
```

### Table

```tsx
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@kloudlite/ui'

export function MyTable() {
  return (
    <Table>
      <TableHeader>
        <TableRow>
          <TableHead>Name</TableHead>
          <TableHead>Email</TableHead>
          <TableHead>Status</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        <TableRow>
          <TableCell>John Doe</TableCell>
          <TableCell>john@example.com</TableCell>
          <TableCell><Badge variant="success">Active</Badge></TableCell>
        </TableRow>
        <TableRow>
          <TableCell>Jane Smith</TableCell>
          <TableCell>jane@example.com</TableCell>
          <TableCell><Badge variant="warning">Pending</Badge></TableCell>
        </TableRow>
      </TableBody>
    </Table>
  )
}
```

### Field (Advanced Form Field)

```tsx
import { Field, FieldLabel, FieldDescription, FieldError } from '@kloudlite/ui'

export function MyField() {
  return (
    <Field>
      <FieldLabel htmlFor="name">Name</FieldLabel>
      <Input id="name" placeholder="Enter your name" />
      <FieldDescription>
        This will be displayed on your profile
      </FieldDescription>
      <FieldError errors={errors} />
    </Field>
  )
}
```

## Hooks

### useMobile

Detect if the current viewport is mobile (width < 768px):

```tsx
import { useIsMobile } from '@kloudlite/ui'

export function MyComponent() {
  const isMobile = useIsMobile()

  return (
    <div>
      {isMobile ? <MobileMenu /> : <DesktopMenu />}
    </div>
  )
}
```

## Icons

The package uses `lucide-react` for icons. Import icons directly:

```tsx
import { Plus, Trash, Settings, Search } from 'lucide-react'

<Button>
  <Plus className="w-4 h-4" />
  Add New
</Button>
```

## Custom Components

### ErrorBoundary

React Error Boundary for catching errors in component trees:

```tsx
import { ErrorBoundary } from '@kloudlite/ui'

export function MyComponent() {
  return (
    <ErrorBoundary fallback={<div>Something went wrong</div>}>
      <RiskyComponent />
    </ErrorBoundary>
  )
}
```

## Theme Colors

The package uses OKLCH color space for consistent colors across light and dark modes:

### Semantic Colors

- `primary` - Primary brand color (blue)
- `secondary` - Secondary color (gray)
- `accent` - Accent color (light gray)
- `destructive` - Error/danger state (red)
- `success` - Success state (green)
- `warning` - Warning state (yellow/orange)
- `info` - Information state (blue)
- `muted` - Muted background
- `card` - Card background
- `popover` - Popover background
- `border` - Border color
- `input` - Input border
- `ring` - Focus ring

### Social Provider Colors

- `github` - GitHub brand color
- `google` - Google brand color
- `microsoft` - Microsoft brand color

### Chart Colors

- `chart-1` through `chart-5` - Color palette for charts

## Design Patterns

### Kloudlite-Specific Patterns

1. **Sharp Corners**: Components use `rounded-none` for a sharp, modern look
2. **Minimal Borders**: Subtle borders using semantic border colors
3. **Focus Rings**: Consistent focus rings for accessibility
4. **Transitions**: Smooth transitions for interactive states
5. **Responsive**: Mobile-first approach with responsive variants

### Console App Patterns

- Stacked cards pattern for installation detail pages
- Tabs with animated underline at `bottom-0`
- Border divider line styling
- Button groups for related actions

### Dashboard App Patterns

- Form fields with labels, descriptions, and error handling
- Tables with consistent cell padding and borders
- Badges for status indicators
- Skeleton loaders for content

## Browser Support

- Modern browsers (Chrome, Firefox, Safari, Edge)
- Requires CSS custom properties (CSS variables)
- Requires ES6+ JavaScript features

## Accessibility

All components follow WCAG 2.1 AA guidelines:

- Full keyboard navigation
- Screen reader support via ARIA attributes
- Focus indicators
- Color contrast ratios
- Semantic HTML elements

## Contributing

When adding new components to this package:

1. Use Radix UI primitives when available
2. Follow the existing component patterns (forwardRef, displayName)
3. Use class-variance-authority for variants
4. Support dark mode via CSS variables
5. Ensure full TypeScript typing
6. Add to index.ts exports
7. Test keyboard navigation
8. Test screen reader compatibility

## License

Internal Kloudlite package - not for external distribution.
