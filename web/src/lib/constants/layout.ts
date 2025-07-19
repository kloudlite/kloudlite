// Layout constants for consistent spacing and sizing across the application

export const LAYOUT = {
  // Container patterns
  CONTAINER: 'container mx-auto max-w-7xl',
  
  // Common padding/spacing patterns
  PADDING: {
    // Page level padding
    PAGE: 'px-4 sm:px-5 md:px-6 py-4 sm:py-5 md:py-6',
    PAGE_X: 'px-4 sm:px-5 md:px-6',
    PAGE_Y: 'py-4 sm:py-5 md:py-6',
    
    // Header padding
    HEADER: 'px-4 sm:px-5 md:px-6 py-3 sm:py-4 md:py-5 lg:py-6',
    HEADER_X: 'px-4 sm:px-5 md:px-6',
    HEADER_Y: 'py-3 sm:py-4 md:py-5 lg:py-6',
    
    // Section padding
    SECTION: 'px-6 py-4',
    SECTION_X: 'px-6',
    SECTION_Y: 'py-4',
    
    // Card padding
    CARD: 'p-4',
    CARD_LG: 'p-6',
    CARD_X: 'px-6',
    CARD_Y: 'py-4',
    
    // Tab padding
    TAB: 'px-2 sm:px-2.5 md:px-3 py-3 sm:py-3.5 md:py-4',
    TAB_X: 'px-2 sm:px-2.5 md:px-3',
    TAB_Y: 'py-3 sm:py-3.5 md:py-4',
    
    // Mobile specific
    MOBILE: 'px-4 py-3',
    MOBILE_X: 'px-4',
    MOBILE_Y: 'py-3',
  },
  
  // Grid patterns
  GRID: {
    COLS_2: 'grid gap-4 md:grid-cols-2',
    COLS_3: 'grid gap-4 md:grid-cols-2 lg:grid-cols-3',
    COLS_4: 'grid gap-4 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4',
    RESPONSIVE_COLS_2: 'grid grid-cols-1 sm:grid-cols-2 gap-3 sm:gap-4 md:gap-5',
    RESPONSIVE_COLS_3: 'grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 gap-3 sm:gap-4 md:gap-4 lg:gap-5',
    RESPONSIVE_COLS_4: 'grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-3 sm:gap-4 md:gap-5',
  },
  
  // Common spacing
  SPACING: {
    SECTION: 'space-y-4 sm:space-y-5 md:space-y-6',
    SECTION_LG: 'space-y-6 md:space-y-8 lg:space-y-10',
    ITEMS: 'space-y-4',
    COMPACT: 'space-y-2',
    TIGHT: 'space-y-1',
    LOOSE: 'space-y-8',
  },
  
  // Common gaps
  GAP: {
    XS: 'gap-1',
    SM: 'gap-2',
    MD: 'gap-3 sm:gap-4',
    LG: 'gap-4 sm:gap-6',
    XL: 'gap-6 sm:gap-8',
    RESPONSIVE: 'gap-1.5 sm:gap-2',
    RESPONSIVE_MD: 'gap-2 sm:gap-3 md:gap-4 lg:gap-6',
  },
  
  // Background patterns
  BACKGROUND: {
    PAGE: 'min-h-screen bg-muted/30 flex flex-col',
    CARD: 'bg-background border rounded-sm',
    SECTION: 'bg-background border rounded-sm',
  }
} as const

export type LayoutPattern = keyof typeof LAYOUT