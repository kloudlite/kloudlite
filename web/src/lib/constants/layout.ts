// Layout constants for consistent spacing and sizing across the application

export const LAYOUT = {
  // Container patterns
  CONTAINER: 'container mx-auto max-w-7xl',
  
  // Common padding/spacing patterns
  PADDING: {
    PAGE: 'px-6 py-6',
    HEADER: 'px-6 py-4', 
    SECTION: 'px-6 py-4',
    CARD: 'p-4',
    CARD_LG: 'p-6',
  },
  
  // Grid patterns
  GRID: {
    COLS_2: 'grid gap-4 md:grid-cols-2',
    COLS_3: 'grid gap-4 md:grid-cols-2 lg:grid-cols-3',
    COLS_4: 'grid gap-4 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4',
  },
  
  // Common spacing
  SPACING: {
    SECTION: 'space-y-6',
    ITEMS: 'space-y-4',
    COMPACT: 'space-y-2',
    LOOSE: 'space-y-8',
  },
  
  // Common gaps
  GAP: {
    SM: 'gap-2',
    MD: 'gap-4', 
    LG: 'gap-6',
  },
  
  // Background patterns
  BACKGROUND: {
    PAGE: 'min-h-screen bg-muted/30 flex flex-col',
    CARD: 'bg-background border rounded-sm',
    SECTION: 'bg-background border rounded-sm',
  }
} as const

export type LayoutPattern = keyof typeof LAYOUT