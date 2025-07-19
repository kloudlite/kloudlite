'use client'

import { useState } from 'react'
import { useRouter } from 'next/navigation'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar'
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from '@/components/ui/form'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
} from '@/components/ui/alert-dialog'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { inviteTeamMember, updateMemberRole, removeMember } from '@/actions/teams'
import { toast } from '@/components/ui/use-toast'
import { 
  Loader2, 
  UserPlus, 
  Mail, 
  Crown, 
  Shield, 
  User, 
  MoreVertical,
  Trash2,
  Calendar 
} from 'lucide-react'
import type { Team, TeamRole, TeamMember } from '@/lib/teams/types'

interface TeamUserManagementProps {
  team: Team & { userRole: TeamRole }
  isOwner: boolean
}

const inviteSchema = z.object({
  email: z.string().email('Please enter a valid email address'),
  role: z.enum(['member', 'admin']),
})

type InviteFormData = z.infer<typeof inviteSchema>

// Mock team members data
const mockTeamMembers: TeamMember[] = [
  {
    id: '1',
    userId: 'user-1',
    teamId: 'team-1',
    role: 'owner',
    joinedAt: new Date(Date.now() - 365 * 24 * 60 * 60 * 1000),
    user: {
      id: 'user-1',
      name: 'Sarah Chen',
      email: 'sarah@team.com',
      avatar: 'https://images.unsplash.com/photo-1494790108755-2616b612b786?w=150',
    }
  },
  {
    id: '2',
    userId: 'user-2',
    teamId: 'team-1',
    role: 'admin',
    joinedAt: new Date(Date.now() - 180 * 24 * 60 * 60 * 1000),
    user: {
      id: 'user-2',
      name: 'Alex Kumar',
      email: 'alex@team.com',
      avatar: null,
    }
  },
  {
    id: '3',
    userId: 'user-3',
    teamId: 'team-1',
    role: 'member',
    joinedAt: new Date(Date.now() - 90 * 24 * 60 * 60 * 1000),
    user: {
      id: 'user-3',
      name: 'Mike Torres',
      email: 'mike@team.com',
      avatar: null,
    }
  },
  {
    id: '4',
    userId: 'user-4',
    teamId: 'team-1',
    role: 'member',
    joinedAt: new Date(Date.now() - 30 * 24 * 60 * 60 * 1000),
    user: {
      id: 'user-4',
      name: 'Jenny Park',
      email: 'jenny@team.com',
      avatar: 'https://images.unsplash.com/photo-1438761681033-6461ffad8d80?w=150',
    }
  },
]

const roleIcons = {
  owner: Crown,
  admin: Shield,
  member: User,
}

function getRoleIcon(role: TeamRole) {
  switch (role) {
    case 'owner':
      return <Crown className="h-4 w-4 text-yellow-500" />
    case 'admin':
      return <Shield className="h-4 w-4 text-blue-500" />
    case 'member':
      return <User className="h-4 w-4 text-gray-500" />
  }
}

function getRoleBadgeVariant(role: TeamRole): "default" | "secondary" | "destructive" | "outline" {
  switch (role) {
    case 'owner':
      return 'default'
    case 'admin':
      return 'secondary'
    case 'member':
      return 'outline'
  }
}

export function TeamUserManagement({ team, isOwner }: TeamUserManagementProps) {
  const router = useRouter()
  const [isInviting, setIsInviting] = useState(false)
  const [members] = useState<TeamMember[]>(mockTeamMembers)

  const inviteForm = useForm<InviteFormData>({
    resolver: zodResolver(inviteSchema),
    defaultValues: {
      email: '',
      role: 'member',
    },
  })

  const onInvite = async (data: InviteFormData) => {
    setIsInviting(true)
    try {
      const result = await inviteTeamMember(team.id, data.email, data.role)
      
      if (result.success) {
        toast.success('Invitation sent successfully')
        inviteForm.reset()
        router.refresh()
      } else {
        toast.error(result.error || 'Failed to send invitation')
      }
    } catch (error) {
      toast.error('An unexpected error occurred')
    } finally {
      setIsInviting(false)
    }
  }

  const handleRoleChange = async (memberId: string, newRole: TeamRole) => {
    try {
      const result = await updateMemberRole(team.id, memberId, newRole)
      
      if (result.success) {
        toast.success('Member role updated successfully')
        router.refresh()
      } else {
        toast.error(result.error || 'Failed to update member role')
      }
    } catch (error) {
      toast.error('An unexpected error occurred')
    }
  }

  const handleRemoveMember = async (memberId: string) => {
    try {
      const result = await removeMember(team.id, memberId)
      
      if (result.success) {
        toast.success('Member removed successfully')
        router.refresh()
      } else {
        toast.error(result.error || 'Failed to remove member')
      }
    } catch (error) {
      toast.error('An unexpected error occurred')
    }
  }

  return (
    <div className="space-y-4 sm:space-y-5 md:space-y-6 lg:space-y-8">
      {/* Invite New Member */}
      <div className="bg-background border rounded-lg overflow-hidden">
        <div className="border-b px-6 py-4">
          <h2 className="font-semibold flex items-center gap-2">
            <UserPlus className="h-5 w-5" />
            Invite Team Member
          </h2>
        </div>
        <div className="px-6 py-4">
          <Form {...inviteForm}>
            <form onSubmit={inviteForm.handleSubmit(onInvite)} className="space-y-4">
              <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5 gap-3 sm:gap-4 md:gap-5 lg:gap-6">
                <div className="sm:col-span-1 md:col-span-2 lg:col-span-3 xl:col-span-4">
                  <FormField
                    control={inviteForm.control}
                    name="email"
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel>Email Address</FormLabel>
                        <FormControl>
                          <Input
                            {...field}
                            type="email"
                            placeholder="Enter email address"
                            className=""
                          />
                        </FormControl>
                        <FormMessage />
                      </FormItem>
                    )}
                  />
                </div>
                <FormField
                  control={inviteForm.control}
                  name="role"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>Role</FormLabel>
                      <Select onValueChange={field.onChange} defaultValue={field.value}>
                        <FormControl>
                          <SelectTrigger>
                            <SelectValue placeholder="Select role" />
                          </SelectTrigger>
                        </FormControl>
                        <SelectContent>
                          <SelectItem value="member">
                            <div className="flex items-center gap-2">
                              <User className="h-4 w-4 text-inherit" />
                              Member
                            </div>
                          </SelectItem>
                          <SelectItem value="admin">
                            <div className="flex items-center gap-2">
                              <Shield className="h-4 w-4 text-inherit" />
                              Admin
                            </div>
                          </SelectItem>
                        </SelectContent>
                      </Select>
                      <FormMessage />
                    </FormItem>
                  )}
                />
              </div>
              <Button type="submit" disabled={isInviting}>
                {isInviting && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
                <Mail className="mr-2 h-4 w-4" />
                Send Invitation
              </Button>
            </form>
          </Form>
        </div>
      </div>

      {/* Team Members */}
      <div className="bg-background border rounded-lg overflow-hidden">
        {/* Table Header */}
        <div className="border-b px-6 py-4">
          <div className="flex items-center justify-between">
            <h2 className="font-semibold">Team Members ({members.length})</h2>
          </div>
        </div>
          {/* Mobile: Card Layout */}
          <div className="block lg:hidden">
            {members.map((member, index) => (
              <div key={member.id} className={`p-6 space-y-4 ${index !== members.length - 1 ? 'border-b' : ''}`}>
                <div className="flex items-center gap-3">
                  <Avatar className="h-8 w-8">
                    <AvatarImage src={member.user.avatar || undefined} />
                    <AvatarFallback className="text-xs">
                      {member.user.name.split(' ').map(n => n[0]).join('')}
                    </AvatarFallback>
                  </Avatar>
                  <div className="flex-1 min-w-0">
                    <p className="font-medium text-sm truncate">{member.user.name}</p>
                    <p className="text-xs text-muted-foreground truncate">{member.user.email}</p>
                  </div>
                </div>
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-2">
                    {(isOwner || team.userRole === 'owner') && member.role !== 'owner' ? (
                      <Select
                        value={member.role}
                        onValueChange={(value: TeamRole) => handleRoleChange(member.id, value)}
                      >
                        <SelectTrigger className="w-28 h-8 text-xs">
                          <SelectValue />
                        </SelectTrigger>
                        <SelectContent>
                          <SelectItem value="member">
                            <div className="flex items-center gap-2">
                              <User className="h-3 w-3 text-inherit" />
                              <span className="text-xs">Member</span>
                            </div>
                          </SelectItem>
                          <SelectItem value="admin">
                            <div className="flex items-center gap-2">
                              <Shield className="h-3 w-3 text-inherit" />
                              <span className="text-xs">Admin</span>
                            </div>
                          </SelectItem>
                        </SelectContent>
                      </Select>
                    ) : (
                      <Badge variant={getRoleBadgeVariant(member.role)} className="flex items-center gap-1 text-xs">
                        {getRoleIcon(member.role)}
                        {member.role}
                      </Badge>
                    )}
                  </div>
                  <div className="flex items-center gap-2">
                    <div className="flex items-center gap-1 text-xs text-muted-foreground">
                      <Calendar className="h-3 w-3" />
                      {member.joinedAt.toLocaleDateString()}
                    </div>
                    {(isOwner || team.userRole === 'owner') && member.role !== 'owner' && (
                      <AlertDialog>
                        <AlertDialogTrigger asChild>
                          <Button variant="ghost" size="sm" className="h-7 w-7 p-0 text-destructive hover:text-destructive hover:bg-destructive/10">
                            <Trash2 className="h-3.5 w-3.5" />
                          </Button>
                        </AlertDialogTrigger>
                        <AlertDialogContent>
                          <AlertDialogHeader>
                            <AlertDialogTitle>Remove Team Member</AlertDialogTitle>
                            <AlertDialogDescription>
                              Are you sure you want to remove {member.user.name} from the team?
                              This action cannot be undone.
                            </AlertDialogDescription>
                          </AlertDialogHeader>
                          <AlertDialogFooter>
                            <AlertDialogCancel>Cancel</AlertDialogCancel>
                            <AlertDialogAction
                              onClick={() => handleRemoveMember(member.id)}
                              className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
                            >
                              Remove Member
                            </AlertDialogAction>
                          </AlertDialogFooter>
                        </AlertDialogContent>
                      </AlertDialog>
                    )}
                  </div>
                </div>
              </div>
            ))}
          </div>

        {/* Desktop: Table Layout */}
        <div className="hidden lg:block overflow-x-auto">
          <table className="w-full">
            <thead>
              <tr className="border-b text-sm text-muted-foreground">
                <th className="text-left font-medium px-6 py-3">Member</th>
                <th className="text-left font-medium px-6 py-3">Role</th>
                <th className="text-left font-medium px-6 py-3">Joined</th>
                <th className="text-left font-medium px-6 py-3 w-16"></th>
              </tr>
            </thead>
            <tbody className="divide-y">
              {members.map((member, index) => (
                <tr key={member.id} className="hover:bg-muted/50 transition-colors group focus-within:bg-muted/50">
                  <td className="px-6 py-4">
                    <div className="flex items-center gap-3">
                      <Avatar>
                        <AvatarImage src={member.user.avatar || undefined} />
                        <AvatarFallback>
                          {member.user.name.split(' ').map(n => n[0]).join('')}
                        </AvatarFallback>
                      </Avatar>
                      <div>
                        <div className="font-medium">{member.user.name}</div>
                        <div className="text-sm text-muted-foreground mt-0.5">{member.user.email}</div>
                      </div>
                    </div>
                  </td>
                  <td className="px-6 py-4">
                    <div className="flex items-center gap-2">
                      {isOwner && member.role !== 'owner' ? (
                        <Select
                          value={member.role}
                          onValueChange={(value: TeamRole) => handleRoleChange(member.id, value)}
                        >
                          <SelectTrigger className="w-32">
                            <SelectValue />
                          </SelectTrigger>
                          <SelectContent>
                            <SelectItem value="member">
                              <div className="flex items-center gap-2">
                                <User className="h-4 w-4" />
                                Member
                              </div>
                            </SelectItem>
                            <SelectItem value="admin">
                              <div className="flex items-center gap-2">
                                <Shield className="h-4 w-4" />
                                Admin
                              </div>
                            </SelectItem>
                          </SelectContent>
                        </Select>
                      ) : (
                        <div className="flex items-center gap-1.5">
                          {roleIcons[member.role] && 
                            (() => {
                              const IconComponent = roleIcons[member.role];
                              return <IconComponent className="h-3.5 w-3.5 text-muted-foreground" />;
                            })()
                          }
                          <span className="text-sm capitalize">{member.role}</span>
                        </div>
                      )}
                    </div>
                  </td>
                  <td className="px-6 py-4">
                    <span className="text-sm text-muted-foreground">
                      {member.joinedAt.toLocaleDateString()}
                    </span>
                  </td>
                  <td className="px-6 py-4">
                    {isOwner && member.role !== 'owner' && (
                      <div className="flex items-center justify-end gap-2 opacity-0 group-hover:opacity-100 focus-within:opacity-100 transition-opacity">
                        <AlertDialog>
                          <AlertDialogTrigger asChild>
                            <Button variant="ghost" size="icon-sm" aria-label={`Remove ${member.user.name}`}>
                              <Trash2 className="h-4 w-4" />
                            </Button>
                          </AlertDialogTrigger>
                          <AlertDialogContent>
                            <AlertDialogHeader>
                              <AlertDialogTitle>Remove Team Member</AlertDialogTitle>
                              <AlertDialogDescription>
                                Are you sure you want to remove {member.user.name} from the team?
                                This action cannot be undone.
                              </AlertDialogDescription>
                            </AlertDialogHeader>
                            <AlertDialogFooter>
                              <AlertDialogCancel>Cancel</AlertDialogCancel>
                              <AlertDialogAction
                                onClick={() => handleRemoveMember(member.id)}
                                className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
                              >
                                Remove Member
                              </AlertDialogAction>
                            </AlertDialogFooter>
                          </AlertDialogContent>
                        </AlertDialog>
                      </div>
                    )}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  )
}