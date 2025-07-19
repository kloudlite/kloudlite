'use client'

import { useState } from 'react'
import { useRouter } from 'next/navigation'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import { RadioGroup, RadioGroupItem } from '@/components/ui/radio-group'
import { Label } from '@/components/ui/label'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import {
  Form,
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from '@/components/ui/form'
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
import { updateTeam, deleteTeam } from '@/actions/teams'
import { toast } from '@/components/ui/use-toast'
import { Loader2, Trash2, Calendar, Users, Activity, Layers, FolderOpen } from 'lucide-react'
import type { Team, TeamRole } from '@/lib/teams/types'

interface TeamGeneralSettingsProps {
  team: Team & { userRole: TeamRole }
  isOwner: boolean
}

const formSchema = z.object({
  name: z.string().min(3, 'Team name must be at least 3 characters').max(50),
  description: z.string().max(200).optional(),
})

type FormData = z.infer<typeof formSchema>

export function TeamGeneralSettings({ team, isOwner }: TeamGeneralSettingsProps) {
  const router = useRouter()
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [isDeleting, setIsDeleting] = useState(false)

  const form = useForm<FormData>({
    resolver: zodResolver(formSchema),
    defaultValues: {
      name: team.name,
      description: team.description || '',
    },
  })

  const onSubmit = async (data: FormData) => {
    setIsSubmitting(true)
    try {
      const result = await updateTeam(team.id, data)
      
      if (result.success) {
        toast.success('Team settings updated successfully')
        router.refresh()
      } else {
        toast.error(result.error || 'Failed to update team settings')
      }
    } catch (error) {
      toast.error('An unexpected error occurred')
    } finally {
      setIsSubmitting(false)
    }
  }

  const handleDelete = async () => {
    setIsDeleting(true)
    try {
      const result = await deleteTeam(team.id)
      
      if (result?.success === false) {
        toast.error(result.error || 'Failed to delete team')
        setIsDeleting(false)
      }
      // If successful, deleteTeam redirects to /teams
    } catch (error) {
      toast.error('An unexpected error occurred')
      setIsDeleting(false)
    }
  }

  return (
    <div className="space-y-4 sm:space-y-5 md:space-y-6">
      {/* Team Overview */}
      <div className="bg-background border rounded-lg overflow-hidden">
        <div className="border-b px-6 py-4">
          <h2 className="font-semibold flex items-center gap-2">
            <Activity className="h-5 w-5 text-primary" />
            Team Overview
          </h2>
        </div>
        <div className="px-6 py-4">
          <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-3 sm:gap-4 md:gap-5">
            <div className="flex items-center gap-3">
              <Users className="h-5 w-5 text-primary flex-shrink-0" />
              <div className="min-w-0">
                <p className="text-sm text-muted-foreground">Members</p>
                <p className="text-2xl font-bold">{team.memberCount || 0}</p>
              </div>
            </div>
            <div className="flex items-center gap-3">
              <Layers className="h-5 w-5 text-primary flex-shrink-0" />
              <div className="min-w-0">
                <p className="text-sm text-muted-foreground">Environments</p>
                <p className="text-2xl font-bold">4</p>
              </div>
            </div>
            <div className="flex items-center gap-3">
              <FolderOpen className="h-5 w-5 text-primary flex-shrink-0" />
              <div className="min-w-0">
                <p className="text-sm text-muted-foreground">Workspaces</p>
                <p className="text-2xl font-bold">8</p>
              </div>
            </div>
            <div className="flex items-center gap-3">
              <Calendar className="h-5 w-5 text-primary flex-shrink-0" />
              <div className="min-w-0">
                <p className="text-sm text-muted-foreground">Created</p>
                <p className="text-base font-semibold">
                  {new Date(team.createdAt).toLocaleDateString()}
                </p>
              </div>
            </div>
          </div>
        </div>
      </div>

      {/* General Settings */}
      <div className="bg-background border rounded-lg overflow-hidden">
        <div className="border-b px-6 py-4">
          <h2 className="font-semibold">General Settings</h2>
        </div>
        <div className="px-6 py-4">
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
                        className="h-11"
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
                    <FormLabel>Description</FormLabel>
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

              {/* Save Button */}
              <Button
                type="submit"
                disabled={isSubmitting || !form.formState.isDirty}
              >
                {isSubmitting && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
                Save Changes
              </Button>
            </form>
          </Form>
        </div>
      </div>

      {/* Danger Zone */}
      {isOwner && (
        <div className="bg-background border border-destructive rounded-lg overflow-hidden">
          <div className="border-b border-destructive px-6 py-4">
            <h2 className="font-semibold text-destructive">Danger Zone</h2>
          </div>
          <div className="px-6 py-4">
            <p className="text-sm text-muted-foreground mb-4">
              Once you delete a team, there is no going back. All team data, including members and settings, will be permanently removed.
            </p>
            
            <AlertDialog>
              <AlertDialogTrigger asChild>
                <Button variant="destructive">
                  <Trash2 className="h-4 w-4 mr-2" />
                  Delete Team
                </Button>
              </AlertDialogTrigger>
              <AlertDialogContent>
                <AlertDialogHeader>
                  <AlertDialogTitle>Are you absolutely sure?</AlertDialogTitle>
                  <AlertDialogDescription>
                    This action cannot be undone. This will permanently delete the team
                    "{team.name}" and remove all members from the team.
                  </AlertDialogDescription>
                </AlertDialogHeader>
                <AlertDialogFooter>
                  <AlertDialogCancel>Cancel</AlertDialogCancel>
                  <AlertDialogAction
                    onClick={handleDelete}
                    disabled={isDeleting}
                    className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
                  >
                    {isDeleting && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
                    Delete Team
                  </AlertDialogAction>
                </AlertDialogFooter>
              </AlertDialogContent>
            </AlertDialog>
          </div>
        </div>
      )}
    </div>
  )
}