import { redirect } from 'next/navigation'
import { auth } from '@/lib/auth'
import { WorkMachinesContent } from '@/components/work-machines-content'

export default async function WorkMachinesPage() {
  const session = await auth()

  if (!session) {
    redirect('/auth/signin')
  }

  const currentUser = session.user?.email || 'user@example.com'
  const isAdmin = currentUser.endsWith('@kloudlite.io')

  // Mock data for work machines
  const workMachines = [
    {
      id: '1',
      owner: currentUser,
      name: `${currentUser.split('@')[0]}-machine`,
      status: 'active' as const,
      cpu: 45,
      memory: 62,
      disk: 38,
      uptime: '3 days, 14 hours',
      type: 'standard',
    },
    ...(isAdmin ? [
      {
        id: '2',
        owner: 'john@team.com',
        name: 'john-machine',
        status: 'active' as const,
        cpu: 78,
        memory: 85,
        disk: 45,
        uptime: '5 days, 2 hours',
        type: 'performance',
      },
      {
        id: '3',
        owner: 'sarah@team.com',
        name: 'sarah-machine',
        status: 'stopped' as const,
        cpu: 0,
        memory: 0,
        disk: 45,
        uptime: '0 minutes',
        type: 'basic',
      },
      {
        id: '4',
        owner: 'mike@team.com',
        name: 'mike-machine',
        status: 'idle' as const,
        cpu: 5,
        memory: 18,
        disk: 22,
        uptime: '12 days, 3 hours',
        type: 'standard',
      },
    ] : [])
  ]

  // Mock pinned workspaces data
  const pinnedWorkspaces = [
    {
      id: '1',
      name: 'web-app',
      environment: 'my-dev-env',
      status: 'active' as const,
      branch: 'main',
      language: 'TypeScript',
      framework: 'Next.js',
    },
    {
      id: '2',
      name: 'api-server',
      environment: 'my-dev-env',
      status: 'active' as const,
      branch: 'develop',
      language: 'Go',
      framework: 'Gin',
    },
    {
      id: '3',
      name: 'ml-service',
      environment: 'ml-dev',
      status: 'idle' as const,
      branch: 'feature/model-v2',
      language: 'Python',
      framework: 'TensorFlow',
    },
  ]

  const pinnedEnvironments = [
    {
      id: '1',
      name: 'my-dev-env',
      status: 'active' as const,
      services: 3,
      workspaces: 2,
      configs: 5,
      secrets: 8,
    },
    {
      id: '2',
      name: 'feature-auth',
      status: 'active' as const,
      services: 2,
      workspaces: 2,
      configs: 3,
      secrets: 6,
    },
  ]

  return (
    <WorkMachinesContent
      initialMachines={workMachines}
      currentUser={currentUser}
      isAdmin={isAdmin}
      pinnedWorkspaces={pinnedWorkspaces}
      pinnedEnvironments={pinnedEnvironments}
    />
  )
}