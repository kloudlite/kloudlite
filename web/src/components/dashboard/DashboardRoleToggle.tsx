'use client';

import { useState } from 'react';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Users, User } from 'lucide-react';
import { RoleDashboard } from './stats/RoleDashboard';
import { 
  mockUserDashboardStats, 
  mockAdminDashboardStats,
  mockUserContext,
  mockAdminContext
} from '@/lib/mock-data/overview-stats';

interface DashboardRoleToggleProps {
  teamName: string;
}

export function DashboardRoleToggle({ teamName }: DashboardRoleToggleProps) {
  const [currentRole, setCurrentRole] = useState<'user' | 'admin'>('admin');

  const context = currentRole === 'admin' ? mockAdminContext : mockUserContext;

  return (
    <div className="space-y-6">
      {/* Role Toggle Demo */}
      <div className="flex items-center justify-between p-4 bg-muted/50 rounded-lg border-2 border-dashed border-muted-foreground/20">
        <div className="flex items-center gap-2">
          <Badge variant="outline">Demo Mode</Badge>
          <span className="text-sm text-muted-foreground">
            Toggle between user and admin dashboard views
          </span>
        </div>
        <div className="flex items-center gap-2">
          <Button
            variant={currentRole === 'user' ? 'default' : 'outline'}
            size="sm"
            onClick={() => setCurrentRole('user')}
            className="flex items-center gap-2"
          >
            <User className="h-4 w-4" />
            User View
          </Button>
          <Button
            variant={currentRole === 'admin' ? 'default' : 'outline'}
            size="sm"
            onClick={() => setCurrentRole('admin')}
            className="flex items-center gap-2"
          >
            <Users className="h-4 w-4" />
            Admin View
          </Button>
        </div>
      </div>

      {/* Role-based Dashboard */}
      <RoleDashboard
        context={context}
        userStats={currentRole === 'user' ? mockUserDashboardStats : undefined}
        adminStats={currentRole === 'admin' ? mockAdminDashboardStats : undefined}
        teamName={teamName}
      />
    </div>
  );
}