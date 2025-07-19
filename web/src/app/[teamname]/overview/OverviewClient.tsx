'use client';

import { useState } from 'react';
import { UserDashboard } from '@/components/dashboard/UserDashboard';
import { AdminDashboard } from '@/components/dashboard/AdminDashboard';
import { AdminModeToggle } from '@/components/dashboard/AdminModeToggle';
import { ActivityFeed } from '@/components/dashboard/activity/ActivityFeed';
import { DashboardMode } from '@/types/dashboard';
import { 
  mockUserDashboardStats, 
  mockAdminDashboardStats,
  mockAdminUser,
  mockRegularUser, // For testing non-admin experience
  isAdmin
} from '@/lib/mock-data/overview-stats';
import { mockActivities } from '@/lib/mock-data/activity-feed';

export function OverviewClient() {
  const [mode, setMode] = useState<DashboardMode>('user'); // Always start with user view
  
  // In real app, this would come from authentication context
  // Change to mockRegularUser to test non-admin experience (no toggle button)
  const currentUser = mockAdminUser; // Admin user - will see toggle button
  const userIsAdmin = isAdmin(currentUser);

  return (
    <div className="space-y-10">
      {/* Page Header with Admin Toggle (only for admins) */}
      <div className="flex items-start justify-between border-b pb-6">
        <div className="space-y-1">
          <h1 className="text-3xl font-bold tracking-tight">Overview</h1>
          <p className="text-muted-foreground text-base">
            {mode === 'user' 
              ? 'Your personal development workspaces and environments' 
              : 'Team developer overview with infrastructure and resource metrics'
            }
          </p>
        </div>
        {userIsAdmin && (
          <AdminModeToggle mode={mode} onModeChange={setMode} />
        )}
      </div>

      {/* Dashboard Content based on Mode */}
      {mode === 'user' ? (
        <>
          <UserDashboard stats={mockUserDashboardStats} />
          
          {/* Recent Activity - only shown in user view */}
          <div className="space-y-6">
            <div className="border-b pb-4">
              <h2 className="text-xl font-semibold tracking-tight">Recent Activity</h2>
              <p className="text-sm text-muted-foreground mt-1">
                Your recent development activities and updates
              </p>
            </div>
            <ActivityFeed activities={mockActivities} maxItems={10} />
          </div>
        </>
      ) : (
        <AdminDashboard stats={mockAdminDashboardStats} />
      )}
    </div>
  );
}