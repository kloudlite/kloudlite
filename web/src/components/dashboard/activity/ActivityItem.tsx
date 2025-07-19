import {
  CheckCircle,
  XCircle,
  AlertCircle,
  Clock,
  Layers,
  FolderOpen,
  Share2,
  Globe,
  Server,
  User,
  Trash2,
  Plus,
  Play,
  Square,
  RefreshCw,
} from 'lucide-react';
import { Activity, ActivityType } from '@/types/dashboard';
import { cn } from '@/lib/utils';

interface ActivityItemProps {
  activity: Activity;
}

function getActivityIcon(type: ActivityType) {
  const iconMap: Record<ActivityType, React.ElementType> = {
    'environment.created': Plus,
    'environment.started': Play,
    'environment.stopped': Square,
    'environment.deleted': Trash2,
    'environment.deployed': Layers,
    'workspace.created': Plus,
    'workspace.started': Play,
    'workspace.stopped': Square,
    'workspace.archived': Square,
    'workspace.deleted': Trash2,
    'user.joined': User,
    'user.role_changed': User,
    'user.removed': Trash2,
    'service.shared.created': Plus,
    'service.shared.updated': RefreshCw,
    'service.shared.deleted': Trash2,
    'service.external.created': Plus,
    'service.external.updated': RefreshCw,
    'service.external.deleted': Trash2,
    'workmachine.status_changed': Server,
    'workmachine.capacity_updated': RefreshCw,
  };
  
  return iconMap[type] || AlertCircle;
}

function getStatusIcon(status: Activity['status']) {
  switch (status) {
    case 'success':
      return CheckCircle;
    case 'failed':
      return XCircle;
    case 'warning':
      return AlertCircle;
    case 'pending':
      return Clock;
    default:
      return Clock;
  }
}

function getStatusColor(status: Activity['status']) {
  switch (status) {
    case 'success':
      return 'text-green-500 bg-green-500/10';
    case 'failed':
      return 'text-red-500 bg-red-500/10';
    case 'warning':
      return 'text-yellow-500 bg-yellow-500/10';
    case 'pending':
      return 'text-blue-500 bg-blue-500/10';
    default:
      return 'text-muted-foreground bg-muted';
  }
}

function formatTimeAgo(date: Date) {
  const now = new Date();
  const diffInSeconds = Math.floor((now.getTime() - date.getTime()) / 1000);
  
  if (diffInSeconds < 60) {
    return 'just now';
  }
  
  const diffInMinutes = Math.floor(diffInSeconds / 60);
  if (diffInMinutes < 60) {
    return `${diffInMinutes}m ago`;
  }
  
  const diffInHours = Math.floor(diffInMinutes / 60);
  if (diffInHours < 24) {
    return `${diffInHours}h ago`;
  }
  
  const diffInDays = Math.floor(diffInHours / 24);
  return `${diffInDays}d ago`;
}

export function ActivityItem({ activity }: ActivityItemProps) {
  const ActivityIcon = getActivityIcon(activity.type);
  const StatusIcon = getStatusIcon(activity.status);
  const statusColor = getStatusColor(activity.status);
  
  return (
    <div className="flex items-start space-x-3 py-3 transition-colors duration-200 hover:bg-muted/20 rounded-lg px-2 -mx-2">
      {/* Activity Icon */}
      <div className="flex-shrink-0">
        <div className="w-8 h-8 rounded-full bg-background border-2 border-muted flex items-center justify-center relative z-10 transition-colors duration-200 hover:border-primary/30">
          <ActivityIcon className="h-4 w-4 text-muted-foreground" />
        </div>
      </div>
      
      {/* Activity Content */}
      <div className="flex-1 min-w-0">
        <div className="flex items-center justify-between">
          <div className="flex items-center space-x-2">
            <p className="text-sm font-medium text-foreground">
              {activity.title}
            </p>
            <div className={cn(
              'flex items-center justify-center w-4 h-4 rounded-full',
              statusColor
            )}>
              <StatusIcon className="h-3 w-3" />
            </div>
          </div>
          <time className="text-xs text-muted-foreground flex-shrink-0">
            {formatTimeAgo(activity.timestamp)}
          </time>
        </div>
        
        <p className="text-sm text-muted-foreground mt-1">
          {activity.description}
        </p>
        
        <div className="flex items-center space-x-2 mt-2">
          {/* User Avatar */}
          <div className="w-5 h-5 rounded-full bg-primary text-primary-foreground flex items-center justify-center text-xs font-medium">
            {activity.user.avatar || activity.user.name.charAt(0)}
          </div>
          <span className="text-xs text-muted-foreground">
            by {activity.user.name}
          </span>
          
          {/* Metadata Tags */}
          {activity.metadata && Object.keys(activity.metadata).length > 0 && (
            <div className="flex items-center space-x-1">
              {Object.entries(activity.metadata).slice(0, 2).map(([key, value]) => (
                <span
                  key={key}
                  className="inline-flex items-center px-2 py-0.5 rounded text-xs bg-muted text-muted-foreground"
                >
                  {Array.isArray(value) ? value.join(', ') : String(value)}
                </span>
              ))}
            </div>
          )}
        </div>
      </div>
    </div>
  );
}