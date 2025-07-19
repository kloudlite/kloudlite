import { Activity } from '@/types/dashboard';

export const mockActivities: Activity[] = [
  {
    id: '1',
    type: 'environment.deployed',
    title: 'Environment deployed',
    description: 'staging environment deployed successfully',
    user: {
      name: 'Sarah Chen',
      email: 'sarah@team.com',
      avatar: 'SC',
    },
    timestamp: new Date(Date.now() - 10 * 60 * 1000), // 10 minutes ago
    status: 'success',
    metadata: {
      environment: 'staging',
      services: 3,
    },
  },
  {
    id: '2',
    type: 'workspace.started',
    title: 'Workspace started',
    description: 'api-service workspace started on work machine wm-dev-01',
    user: {
      name: 'Alex Kumar',
      email: 'alex@team.com',
      avatar: 'AK',
    },
    timestamp: new Date(Date.now() - 25 * 60 * 1000), // 25 minutes ago
    status: 'success',
    metadata: {
      workspace: 'api-service',
      workMachine: 'wm-dev-01',
    },
  },
  {
    id: '3',
    type: 'service.shared.created',
    title: 'Shared service created',
    description: 'Redis cache service added to shared services',
    user: {
      name: 'Mike Torres',
      email: 'mike@team.com',
      avatar: 'MT',
    },
    timestamp: new Date(Date.now() - 45 * 60 * 1000), // 45 minutes ago
    status: 'success',
    metadata: {
      serviceName: 'redis-cache',
      serviceType: 'Redis',
    },
  },
  {
    id: '4',
    type: 'environment.failed',
    title: 'Environment deployment failed',
    description: 'production environment deployment failed due to resource limits',
    user: {
      name: 'Sarah Chen',
      email: 'sarah@team.com',
      avatar: 'SC',
    },
    timestamp: new Date(Date.now() - 1 * 60 * 60 * 1000), // 1 hour ago
    status: 'failed',
    metadata: {
      environment: 'production',
      error: 'Resource limits exceeded',
    },
  },
  {
    id: '5',
    type: 'service.external.updated',
    title: 'External service updated',
    description: 'Stripe payment service configuration updated',
    user: {
      name: 'Jenny Park',
      email: 'jenny@team.com',
      avatar: 'JP',
    },
    timestamp: new Date(Date.now() - 2 * 60 * 60 * 1000), // 2 hours ago
    status: 'success',
    metadata: {
      serviceName: 'stripe-payments',
      changes: ['webhook_url', 'api_version'],
    },
  },
  {
    id: '6',
    type: 'workspace.created',
    title: 'Workspace created',
    description: 'frontend-app workspace created from template',
    user: {
      name: 'David Kim',
      email: 'david@team.com',
      avatar: 'DK',
    },
    timestamp: new Date(Date.now() - 3 * 60 * 60 * 1000), // 3 hours ago
    status: 'success',
    metadata: {
      workspace: 'frontend-app',
      template: 'react-nextjs',
    },
  },
  {
    id: '7',
    type: 'environment.started',
    title: 'Environment started',
    description: 'development environment started with 5 services',
    user: {
      name: 'Alex Kumar',
      email: 'alex@team.com',
      avatar: 'AK',
    },
    timestamp: new Date(Date.now() - 4 * 60 * 60 * 1000), // 4 hours ago
    status: 'success',
    metadata: {
      environment: 'development',
      services: 5,
    },
  },
  {
    id: '8',
    type: 'workmachine.capacity_updated',
    title: 'Work machine capacity updated',
    description: 'wm-dev-02 upgraded with additional 32GB RAM',
    user: {
      name: 'System Admin',
      email: 'admin@team.com',
      avatar: 'SA',
    },
    timestamp: new Date(Date.now() - 6 * 60 * 60 * 1000), // 6 hours ago
    status: 'success',
    metadata: {
      workMachine: 'wm-dev-02',
      upgrade: '32GB RAM added',
    },
  },
  {
    id: '9',
    type: 'service.shared.deleted',
    title: 'Shared service removed',
    description: 'Deprecated MongoDB service removed from shared services',
    user: {
      name: 'Mike Torres',
      email: 'mike@team.com',
      avatar: 'MT',
    },
    timestamp: new Date(Date.now() - 8 * 60 * 60 * 1000), // 8 hours ago
    status: 'success',
    metadata: {
      serviceName: 'mongodb-legacy',
      reason: 'deprecated',
    },
  },
  {
    id: '10',
    type: 'workspace.stopped',
    title: 'Workspace stopped',
    description: 'ml-training workspace stopped after completion',
    user: {
      name: 'Jenny Park',
      email: 'jenny@team.com',
      avatar: 'JP',
    },
    timestamp: new Date(Date.now() - 12 * 60 * 60 * 1000), // 12 hours ago
    status: 'success',
    metadata: {
      workspace: 'ml-training',
      duration: '4h 23m',
    },
  },
  {
    id: '11',
    type: 'environment.scaled',
    title: 'Environment scaled',
    description: 'production environment scaled up to handle increased traffic',
    user: {
      name: 'Sarah Chen',
      email: 'sarah@team.com',
      avatar: 'SC',
    },
    timestamp: new Date(Date.now() - 18 * 60 * 60 * 1000), // 18 hours ago
    status: 'success',
    metadata: {
      environment: 'production',
      replicas: '3 → 8',
    },
  },
  {
    id: '12',
    type: 'workspace.cloned',
    title: 'Workspace cloned',
    description: 'api-gateway workspace cloned for hotfix development',
    user: {
      name: 'Ryan Miller',
      email: 'ryan@team.com',
      avatar: 'RM',
    },
    timestamp: new Date(Date.now() - 20 * 60 * 60 * 1000), // 20 hours ago
    status: 'success',
    metadata: {
      sourceWorkspace: 'api-gateway',
      targetWorkspace: 'api-gateway-hotfix',
    },
  },
  {
    id: '13',
    type: 'service.external.created',
    title: 'External service added',
    description: 'Elasticsearch logging service connected to platform',
    user: {
      name: 'Lisa Wilson',
      email: 'lisa@team.com',
      avatar: 'LW',
    },
    timestamp: new Date(Date.now() - 24 * 60 * 60 * 1000), // 1 day ago
    status: 'success',
    metadata: {
      serviceName: 'elasticsearch-logs',
      endpoint: 'logs.example.com',
    },
  },
  {
    id: '14',
    type: 'environment.backed_up',
    title: 'Environment backed up',
    description: 'staging environment backup completed successfully',
    user: {
      name: 'System Admin',
      email: 'admin@team.com',
      avatar: 'SA',
    },
    timestamp: new Date(Date.now() - 26 * 60 * 60 * 1000), // 26 hours ago
    status: 'success',
    metadata: {
      environment: 'staging',
      backupSize: '2.4GB',
    },
  },
  {
    id: '15',
    type: 'workspace.restored',
    title: 'Workspace restored',
    description: 'mobile-app workspace restored from backup after corruption',
    user: {
      name: 'David Brown',
      email: 'david@team.com',
      avatar: 'DB',
    },
    timestamp: new Date(Date.now() - 30 * 60 * 60 * 1000), // 30 hours ago
    status: 'warning',
    metadata: {
      workspace: 'mobile-app',
      backupDate: '2 days ago',
    },
  },
];