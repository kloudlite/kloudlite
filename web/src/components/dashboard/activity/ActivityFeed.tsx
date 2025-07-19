import { Activity } from '@/types/dashboard';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { ActivityItem } from './ActivityItem';
import { Clock } from 'lucide-react';

interface ActivityFeedProps {
  activities: Activity[];
  maxItems?: number;
}

export function ActivityFeed({ activities, maxItems = 10 }: ActivityFeedProps) {
  const displayedActivities = activities.slice(0, maxItems);
  
  if (activities.length === 0) {
    return (
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2 text-xl font-bold">
            <Clock className="h-6 w-6" />
            Recent Activity
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="flex flex-col items-center justify-center py-8 text-center">
            <Clock className="h-12 w-12 text-muted-foreground mb-4" />
            <p className="text-muted-foreground">No recent activity</p>
            <p className="text-sm text-muted-foreground mt-1">
              Activity will appear here as team members work with environments, workspaces, and services.
            </p>
          </div>
        </CardContent>
      </Card>
    );
  }
  
  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <Clock className="h-5 w-5" />
          Recent Activity
          <span className="text-sm font-normal text-muted-foreground ml-auto">
            {activities.length} {activities.length === 1 ? 'activity' : 'activities'}
          </span>
        </CardTitle>
      </CardHeader>
      <CardContent>
        <div className="space-y-0 relative">
          {/* Continuous timeline line */}
          <div className="absolute left-4 top-0 bottom-0 w-0.5 bg-border" />
          
          {displayedActivities.map((activity, index) => (
            <div key={activity.id} className="relative">
              <ActivityItem activity={activity} />
            </div>
          ))}
          
          {activities.length > maxItems && (
            <div className="pt-4 mt-4 border-t">
              <p className="text-sm text-muted-foreground text-center">
                Showing {maxItems} of {activities.length} activities
              </p>
            </div>
          )}
        </div>
      </CardContent>
    </Card>
  );
}