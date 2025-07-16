import { Link } from '@/components/ui/link'
import { Button } from '@/components/ui/button'
import { Users, Calendar, Activity, Crown, Shield } from 'lucide-react'
import type { Team, TeamRole } from '@/lib/teams/types'
import { formatDistanceToNow } from 'date-fns'

interface TeamCardProps {
  team: Team & { userRole: TeamRole }
  userRole: TeamRole
}

const roleIcons = {
  owner: Crown,
  admin: Shield,
  member: Users,
}

export function TeamCard({ team, userRole }: TeamCardProps) {
  const RoleIcon = roleIcons[userRole]
  
  return (
    <div className="group relative border border-border rounded-none bg-card hover:border-primary/50 transition-all duration-200">
      {/* Role Badge */}
      <div className="absolute top-4 right-4 z-10">
        <div className={`
          flex items-center gap-1.5 px-3 py-1.5 text-xs font-medium rounded-none
          ${userRole === 'owner' ? 'bg-primary text-primary-foreground' : 
            userRole === 'admin' ? 'bg-primary/10 text-primary' : 
            'bg-muted text-muted-foreground'}
        `}>
          <RoleIcon className="h-3.5 w-3.5" />
          <span className="capitalize">{userRole}</span>
        </div>
      </div>

      <div className="p-6 space-y-6">
        {/* Header */}
        <div className="pr-20">
          <h3 className="text-xl font-semibold mb-2">{team.name}</h3>
          {team.description && (
            <p className="text-muted-foreground line-clamp-2">
              {team.description}
            </p>
          )}
        </div>

        {/* Stats Grid */}
        <div className="grid grid-cols-2 gap-4 py-4 border-y border-border">
          <div>
            <div className="flex items-center text-muted-foreground mb-1">
              <Users className="h-4 w-4 mr-2" />
              <span className="text-sm">Members</span>
            </div>
            <p className="text-2xl font-semibold">{team.memberCount}</p>
          </div>
          <div>
            <div className="flex items-center text-muted-foreground mb-1">
              <Activity className="h-4 w-4 mr-2" />
              <span className="text-sm">Last Active</span>
            </div>
            <p className="text-sm font-medium">
              {team.lastActivity ? formatDistanceToNow(team.lastActivity, { addSuffix: true }) : 'Never'}
            </p>
          </div>
        </div>

        {/* Metadata */}
        <div className="flex items-center justify-between text-sm text-muted-foreground">
          <div className="flex items-center">
            <Calendar className="h-4 w-4 mr-1.5" />
            <span>Created {formatDistanceToNow(team.createdAt, { addSuffix: true })}</span>
          </div>
          <span className={`
            px-2 py-0.5 text-xs font-medium rounded-none
            ${team.visibility === 'private' ? 'bg-muted' : 'bg-green-100 text-green-700 dark:bg-green-900/20 dark:text-green-400'}
          `}>
            {team.visibility}
          </span>
        </div>

        {/* Actions */}
        <div className="pt-2">
          <Button asChild variant="outline" className="w-full rounded-none h-11 font-medium group-hover:border-primary group-hover:text-primary transition-colors">
            <Link href={`/teams/${team.id}`}>
              View Team Details
            </Link>
          </Button>
        </div>
      </div>
    </div>
  )
}