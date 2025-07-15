# Web Application Development Guide

## Overview

This is a Next.js 15 web application built with modern React 19 and a comprehensive component library using shadcn/ui components.

## Tech Stack

- **Framework**: Next.js 15.3.5 with App Router
- **React**: 19.1.0
- **TypeScript**: 5.x
- **Styling**: Tailwind CSS v4.0.0-alpha.20
- **UI Components**: shadcn/ui with Radix UI primitives
- **Icons**: Lucide React & Heroicons
- **Theme**: next-themes for dark/light mode
- **Forms**: React Hook Form with Zod validation
- **Charts**: Recharts
- **Notifications**: Sonner toasts

## Getting Started

### Prerequisites

- Node.js 18+ 
- pnpm (preferred package manager)

### Installation

```bash
cd web
pnpm install
```

### Development

```bash
pnpm dev
```

The application will be available at `http://localhost:3000`

### Build

```bash
pnpm build
pnpm start
```

### Linting

```bash
pnpm lint
```

## Project Structure

```
web/
├── src/
│   ├── app/                 # Next.js App Router
│   │   ├── layout.tsx       # Root layout
│   │   ├── page.tsx         # Home page
│   │   ├── globals.css      # Global styles
│   │   └── theme-script.tsx # Theme initialization
│   ├── components/
│   │   └── ui/              # shadcn/ui components
│   └── lib/
│       ├── utils.ts         # Utility functions
│       └── theme-cookie.ts  # Theme cookie handling
├── public/                  # Static assets
└── package.json
```

## UI Components

The application uses shadcn/ui components which are based on Radix UI primitives. Available components include:

- **Layout**: Accordion, Collapsible, Resizable, Separator, Sidebar, Tabs
- **Forms**: Button, Input, Label, Textarea, Checkbox, Radio Group, Select, Switch, Toggle
- **Data Display**: Avatar, Badge, Card, Table, Progress, Chart
- **Feedback**: Alert, Dialog, Toast, Tooltip, Hover Card
- **Navigation**: Breadcrumb, Command, Dropdown Menu, Navigation Menu, Pagination
- **Overlays**: Dialog, Drawer, Popover, Sheet
- **Media**: Carousel, File Upload

## Theme System

The application supports dark/light mode switching using next-themes:

- Theme state is persisted in cookies
- Automatic system preference detection
- Custom theme script prevents flash on page load

## Development Guidelines

### Adding New Components

Use the existing shadcn/ui components as building blocks. For custom components, follow the established patterns in the `components/ui` directory.

### Styling

- Use Tailwind CSS classes for styling
- Leverage the component variants system with `class-variance-authority`
- Follow the existing design system patterns

### Forms

Use React Hook Form with Zod for form validation:

```tsx
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
```

### Icons

Use Lucide React for consistent iconography:

```tsx
import { ChevronDown } from 'lucide-react'
```

## Performance Considerations

- Next.js 15 with React 19 provides automatic optimizations
- Use React Server Components where appropriate
- Implement proper loading states and error boundaries
- Optimize images using Next.js Image component

## Deployment

The application is configured for deployment on platforms that support Next.js:

- Vercel (recommended)
- Netlify
- Custom Node.js servers

Ensure environment variables are properly configured for production deployment.