import { UserDashboard } from './UserDashboard';
import { AdminDashboard } from './AdminDashboard';
import {
  DashboardContext,
  UserDashboardStats,
  AdminDashboardStats,
} from '@/types/dashboard';

interface RoleDashboardProps {
  context: DashboardContext;
  userStats?: UserDashboardStats;
  adminStats?: AdminDashboardStats;
  teamName: string;
}

export function RoleDashboard({
  context,
  userStats,
  adminStats,
  teamName,
}: RoleDashboardProps) {
  if (context.role === 'admin') {
    if (!adminStats) {
      return (
        <div className="text-center py-8">
          <p className="text-muted-foreground">Loading admin dashboard...</p>
        </div>
      );
    }
    
    return <AdminDashboard stats={adminStats} teamName={teamName} />;
  }

  if (!userStats) {
    return (
      <div className="text-center py-8">
        <p className="text-muted-foreground">Loading your dashboard...</p>
      </div>
    );
  }

  return <UserDashboard stats={userStats} userName={context.userName} />;
}