# Auth Pages Design System

This document outlines the design principles and implementation guidelines for all authentication pages in the Kloudlite platform.

## Design Philosophy

The auth pages follow a **monospace/brutalist design aesthetic** characterized by:
- Zero border radius (sharp corners)
- High contrast
- Geometric shapes
- Functional minimalism
- IBM Plex Mono font for monospace elements

## Core Design Tokens

### Typography
- **Primary Font**: System UI sans-serif stack
- **Monospace Font**: IBM Plex Mono (when explicitly needed)
- **Title Size**: `text-3xl` (30px) with `font-semibold`
- **Body Text**: `text-base` (16px)
- **Small Text**: `text-sm` (14px)
- **Form Labels**: Default size with `font-medium`

### Colors
- **Background**: White (`#ffffff`) / Dark: `#0f172a`
- **Card Background**: White / Dark: `#1e293b`
- **Border**: `#e5e7eb` / Dark: `#4b5563`
- **Primary**: `#3b82f6` / Dark: `#60a5fa`
- **Destructive**: `#dc2626`
- **Muted Text**: `#6b7280`

### Spacing
- **Card Padding**: 24px (6 units)
- **Section Spacing**: 24px between major sections
- **Field Spacing**: 12px between form fields
- **Label Gap**: 4px from label to input

### Border Radius
- **ALL ELEMENTS**: `rounded-none` (0px)
- No rounded corners on cards, buttons, or inputs
- Sharp, geometric edges throughout

## Component Guidelines

### AuthCard
- Full width within `max-w-md` container
- White background with gray border
- `shadow-sm` for subtle depth
- Consistent padding: `p-6`
- Header spacing: `space-y-3`

### Form Inputs
- Height: `h-11` (44px)
- Border: 1px solid border color
- **MUST USE**: `rounded-none` (not `rounded-md`)
- Focus ring: 2px primary color with offset
- Consistent padding: `px-4 py-2`

### Buttons
- Primary button height: `h-11` (44px) for main CTAs
- Secondary button height: `h-9` (36px)
- **MUST USE**: `rounded-none`
- Full width for primary actions
- Active state: `scale-[0.99]`

### Form Layout
- `space-y-6` between major sections
- `space-y-3` between input groups
- `space-y-1` between label and input
- Error messages directly below inputs

## Page Structure

```tsx
<AuthCard
  title="Page Title"
  description="Optional description"
>
  <form className="space-y-6">
    {/* Alert for errors */}
    {/* Input fields with space-y-3 */}
    {/* Primary CTA button */}
    {/* Secondary links/text */}
  </form>
</AuthCard>
```

## Implementation Checklist

- [ ] Use `rounded-none` on ALL interactive elements
- [ ] Maintain consistent spacing using the defined system
- [ ] Use `h-11` for primary buttons and main inputs
- [ ] Include proper loading and error states
- [ ] Add focus rings for accessibility
- [ ] Test in both light and dark modes
- [ ] Ensure form validation with React Hook Form
- [ ] Include password strength indicators where applicable

## Security Considerations

- Server-side actions for all auth operations
- CSRF protection on all forms
- Rate limiting for sensitive operations
- Secure session management
- Email verification for new accounts
- Strong password requirements

## Accessibility

- Proper label associations
- Keyboard navigation support
- Screen reader friendly error messages
- Focus management
- High contrast ratios
- Clear visual feedback