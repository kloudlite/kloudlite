import { 
  LayoutDashboard, 
  Users, 
  Settings, 
  Shield,
  Layers,
  FolderOpen,
  Share2,
} from 'lucide-react'

export interface NavItem {
  label: string
  href: string
  icon: React.ElementType
  badge?: string | number
  badgeVariant?: 'default' | 'secondary' | 'destructive' | 'outline'
}

export const mainNavItems: NavItem[] = [
  { 
    label: 'Overview', 
    href: '', 
    icon: LayoutDashboard,
  },
  { 
    label: 'Environments', 
    href: '/environments', 
    icon: Layers,
    badge: '4',
  },
  { 
    label: 'Workspaces', 
    href: '/workspaces', 
    icon: FolderOpen,
    badge: '12',
  },
  { 
    label: 'Shared Services', 
    href: '/shared-services', 
    icon: Share2,
    badge: '7',
  },
]

export const teamNavItems: NavItem[] = [
  { 
    label: 'Team Settings', 
    href: '/settings', 
    icon: Settings,
  },
]