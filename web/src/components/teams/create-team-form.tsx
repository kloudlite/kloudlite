'use client'

import { useState } from 'react'
import { useRouter } from 'next/navigation'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import { Label } from '@/components/ui/label'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import {
  Form,
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from '@/components/ui/form'
import { createTeam } from '@/actions/teams'
import { toast } from '@/components/ui/use-toast'
import { Loader2, Plus, X, MapPin } from 'lucide-react'
import type { CreateTeamInput, TeamRole } from '@/lib/teams/types'

// AWS Regions
const AWS_REGIONS = [
  { value: 'us-east-1', label: 'US East (N. Virginia)', flag: '🇺🇸' },
  { value: 'us-east-2', label: 'US East (Ohio)', flag: '🇺🇸' },
  { value: 'us-west-1', label: 'US West (N. California)', flag: '🇺🇸' },
  { value: 'us-west-2', label: 'US West (Oregon)', flag: '🇺🇸' },
  { value: 'ca-central-1', label: 'Canada (Central)', flag: '🇨🇦' },
  { value: 'eu-central-1', label: 'Europe (Frankfurt)', flag: '🇩🇪' },
  { value: 'eu-west-1', label: 'Europe (Ireland)', flag: '🇮🇪' },
  { value: 'eu-west-2', label: 'Europe (London)', flag: '🇬🇧' },
  { value: 'eu-west-3', label: 'Europe (Paris)', flag: '🇫🇷' },
  { value: 'eu-north-1', label: 'Europe (Stockholm)', flag: '🇸🇪' },
  { value: 'ap-south-1', label: 'Asia Pacific (Mumbai)', flag: '🇮🇳' },
  { value: 'ap-southeast-1', label: 'Asia Pacific (Singapore)', flag: '🇸🇬' },
  { value: 'ap-southeast-2', label: 'Asia Pacific (Sydney)', flag: '🇦🇺' },
  { value: 'ap-northeast-1', label: 'Asia Pacific (Tokyo)', flag: '🇯🇵' },
  { value: 'ap-northeast-2', label: 'Asia Pacific (Seoul)', flag: '🇰🇷' },
  { value: 'sa-east-1', label: 'South America (São Paulo)', flag: '🇧🇷' },
]

const formSchema = z.object({
  name: z.string().min(3, 'Team name must be at least 3 characters').max(50),
  description: z.string().max(200).optional(),
  region: z.string().min(1, 'Please select a region'),
  invitations: z.array(
    z.object({
      email: z.string().email('Invalid email address'),
      role: z.enum(['admin', 'member']),
    })
  ).optional(),
})

type FormData = z.infer<typeof formSchema>

export function CreateTeamForm() {
  const router = useRouter()
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [invitations, setInvitations] = useState<{ email: string; role: TeamRole }[]>([])
  const [newInviteEmail, setNewInviteEmail] = useState('')
  const [newInviteRole, setNewInviteRole] = useState<TeamRole>('member')

  const form = useForm<FormData>({
    resolver: zodResolver(formSchema),
    defaultValues: {
      name: '',
      description: '',
      region: '',
      invitations: [],
    },
  })

  const onSubmit = async (data: FormData) => {
    setIsSubmitting(true)
    try {
      const input: CreateTeamInput = {
        ...data,
        invitations: invitations.length > 0 ? invitations : undefined,
      }
      
      const result = await createTeam(input)
      
      if (result.success) {
        toast.success('Team created successfully')
        router.push(`/teams/${result.teamId}`)
      } else {
        toast.error(result.error || 'Failed to create team')
      }
    } catch (error) {
      toast.error('An unexpected error occurred')
    } finally {
      setIsSubmitting(false)
    }
  }

  const addInvitation = () => {
    if (newInviteEmail && z.string().email().safeParse(newInviteEmail).success) {
      if (!invitations.some(inv => inv.email === newInviteEmail)) {
        setInvitations([...invitations, { email: newInviteEmail, role: newInviteRole }])
        setNewInviteEmail('')
        setNewInviteRole('member')
      } else {
        toast.error('This email is already invited')
      }
    } else {
      toast.error('Please enter a valid email address')
    }
  }

  const removeInvitation = (email: string) => {
    setInvitations(invitations.filter(inv => inv.email !== email))
  }

  return (
    <Form {...form}>
      <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-6">
        {/* Team Name */}
        <FormField
          control={form.control}
          name="name"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Team Name</FormLabel>
              <FormControl>
                <Input
                  {...field}
                  placeholder="Enter team name"
                  autoComplete="off"
                />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />

        {/* Description */}
        <FormField
          control={form.control}
          name="description"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Description (Optional)</FormLabel>
              <FormControl>
                <Textarea
                  {...field}
                  placeholder="Brief description of your team"
                  className="resize-none"
                  rows={3}
                />
              </FormControl>
              <FormDescription>
                Help others understand what your team is about
              </FormDescription>
              <FormMessage />
            </FormItem>
          )}
        />

        {/* Region */}
        <FormField
          control={form.control}
          name="region"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Region</FormLabel>
              <Select onValueChange={field.onChange} defaultValue={field.value}>
                <FormControl>
                  <SelectTrigger className="w-full">
                    <SelectValue placeholder="Select a region" />
                  </SelectTrigger>
                </FormControl>
                <SelectContent>
                  {AWS_REGIONS.map((region) => (
                    <SelectItem key={region.value} value={region.value}>
                      <div className="flex items-center gap-2">
                        <span>{region.flag}</span>
                        <span>{region.label}</span>
                      </div>
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
              <FormDescription>
                Choose the AWS region where your team resources will be deployed
              </FormDescription>
              <FormMessage />
            </FormItem>
          )}
        />

        {/* Initial Invitations */}
        <div className="space-y-4">
          <div>
            <Label>Invite Members (Optional)</Label>
            <p className="text-sm text-muted-foreground mt-1">
              You can invite more members after creating the team
            </p>
          </div>

          {/* Add Invitation Form */}
          <div className="flex gap-2">
            <Input
              type="email"
              placeholder="Email address"
              value={newInviteEmail}
              onChange={(e) => setNewInviteEmail(e.target.value)}
              onKeyPress={(e) => e.key === 'Enter' && (e.preventDefault(), addInvitation())}
              className="flex-1"
            />
            <Select value={newInviteRole} onValueChange={(value) => setNewInviteRole(value as TeamRole)}>
              <SelectTrigger className="w-[120px] !h-11">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="member">Member</SelectItem>
                <SelectItem value="admin">Admin</SelectItem>
              </SelectContent>
            </Select>
            <Button
              type="button"
              onClick={addInvitation}
              size="icon-lg"
              variant="secondary"
            >
              <Plus className="h-4 w-4" />
            </Button>
          </div>

          {/* Invitation List */}
          {invitations.length > 0 && (
            <div className="space-y-2">
              {invitations.map((invitation) => (
                <div
                  key={invitation.email}
                  className="flex items-center justify-between p-3 border rounded-sm bg-muted/30"
                >
                  <div className="flex items-center gap-3">
                    <span className="text-sm">{invitation.email}</span>
                    <span className="text-xs px-2 py-1 bg-background border rounded-sm capitalize">
                      {invitation.role}
                    </span>
                  </div>
                  <Button
                    type="button"
                    variant="ghost"
                    size="icon"
                    onClick={() => removeInvitation(invitation.email)}
                  >
                    <X className="h-4 w-4" />
                  </Button>
                </div>
              ))}
            </div>
          )}
        </div>

        {/* Submit Button */}
        <Button
          type="submit"
          disabled={isSubmitting}
          className="w-full"
          size="lg"
        >
          {isSubmitting && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
          Create Team
        </Button>
      </form>
    </Form>
  )
}