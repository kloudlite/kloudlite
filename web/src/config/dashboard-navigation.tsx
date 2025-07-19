import { 
  LayoutDashboard, 
  Users, 
  Settings, 
  Shield,
  Layers,
  CodeXml,
  Network,
  PieChart,
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
    href: '/overview', 
    icon: PieChart,
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
    icon: CodeXml,
    badge: '12',
  },
  { 
    label: 'Shared Services', 
    href: '/shared-services', 
    icon: Network,
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