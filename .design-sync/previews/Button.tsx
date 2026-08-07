import * as React from 'react'
import { Button } from '@kloudlite/ui'

export const Variants = () => (
  <div className="flex flex-wrap items-center gap-3">
    <Button>Create environment</Button>
    <Button variant="secondary">Save draft</Button>
    <Button variant="outline">Cancel</Button>
    <Button variant="ghost">Skip for now</Button>
    <Button variant="destructive">Delete workspace</Button>
    <Button variant="link">View documentation</Button>
  </div>
)

export const Sizes = () => (
  <div className="flex flex-wrap items-center gap-3">
    <Button size="sm">Small</Button>
    <Button size="default">Default</Button>
    <Button size="lg">Large</Button>
    <Button size="icon" aria-label="Add">
      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
        <path d="M12 5v14M5 12h14" />
      </svg>
    </Button>
  </div>
)

export const WithIcon = () => (
  <div className="flex flex-wrap items-center gap-3">
    <Button>
      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
        <path d="M12 5v14M5 12h14" />
      </svg>
      New workspace
    </Button>
    <Button variant="outline">
      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
        <path d="M21 12a9 9 0 1 1-6.2-8.6" />
        <path d="M21 3v6h-6" />
      </svg>
      Retry install
    </Button>
  </div>
)

export const Disabled = () => (
  <div className="flex flex-wrap items-center gap-3">
    <Button disabled>Create environment</Button>
    <Button variant="outline" disabled>
      Cancel
    </Button>
    <Button variant="destructive" disabled>
      Delete workspace
    </Button>
  </div>
)
