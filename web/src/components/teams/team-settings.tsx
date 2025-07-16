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
import { Loader2, Globe, Lock, Trash2 } from 'lucide-react'
import type { Team, TeamRole } from '@/lib/teams/types'

interface TeamSettingsProps {
  team: Team & { userRole: TeamRole }
  isOwner: boolean
}

const formSchema = z.object({
  name: z.string().min(3, 'Team name must be at least 3 characters').max(50),
  description: z.string().max(200).optional(),
  visibility: z.enum(['public', 'private']),
})

type FormData = z.infer<typeof formSchema>

export function TeamSettings({ team, isOwner }: TeamSettingsProps) {
  const router = useRouter()
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [isDeleting, setIsDeleting] = useState(false)

  const form = useForm<FormData>({
    resolver: zodResolver(formSchema),
    defaultValues: {
      name: team.name,
      description: team.description || '',
      visibility: team.visibility,
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
    <div className="space-y-8">
      {/* General Settings */}
      <div className="border border-border rounded-none p-6">
        <h3 className="text-lg font-semibold mb-6">General Settings</h3>
        
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
                      className="rounded-none h-11"
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
                      className="rounded-none resize-none"
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

            {/* Visibility */}
            <FormField
              control={form.control}
              name="visibility"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Visibility</FormLabel>
                  <FormControl>
                    <RadioGroup
                      onValueChange={field.onChange}
                      defaultValue={field.value}
                      className="space-y-3"
                    >
                      <div className="flex items-start space-x-3">
                        <RadioGroupItem value="private" id="private" className="mt-1" />
                        <Label htmlFor="private" className="cursor-pointer flex-1">
                          <div className="flex items-center mb-1">
                            <Lock className="h-4 w-4 mr-2" />
                            <span className="font-medium">Private</span>
                          </div>
                          <p className="text-sm text-muted-foreground">
                            Only invited members can see and join this team
                          </p>
                        </Label>
                      </div>
                      <div className="flex items-start space-x-3">
                        <RadioGroupItem value="public" id="public" className="mt-1" />
                        <Label htmlFor="public" className="cursor-pointer flex-1">
                          <div className="flex items-center mb-1">
                            <Globe className="h-4 w-4 mr-2" />
                            <span className="font-medium">Public</span>
                          </div>
                          <p className="text-sm text-muted-foreground">
                            Anyone in your organization can see and request to join
                          </p>
                        </Label>
                      </div>
                    </RadioGroup>
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />

            {/* Save Button */}
            <Button
              type="submit"
              disabled={isSubmitting || !form.formState.isDirty}
              className="rounded-none"
            >
              {isSubmitting && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
              Save Changes
            </Button>
          </form>
        </Form>
      </div>

      {/* Danger Zone */}
      {isOwner && (
        <div className="border border-destructive rounded-none p-6">
          <h3 className="text-lg font-semibold text-destructive mb-2">Danger Zone</h3>
          <p className="text-sm text-muted-foreground mb-4">
            Once you delete a team, there is no going back. All team data, including members and settings, will be permanently removed.
          </p>
          
          <AlertDialog>
            <AlertDialogTrigger asChild>
              <Button variant="destructive" className="rounded-none">
                <Trash2 className="h-4 w-4 mr-2" />
                Delete Team
              </Button>
            </AlertDialogTrigger>
            <AlertDialogContent className="rounded-none">
              <AlertDialogHeader>
                <AlertDialogTitle>Are you absolutely sure?</AlertDialogTitle>
                <AlertDialogDescription>
                  This action cannot be undone. This will permanently delete the team
                  "{team.name}" and remove all members from the team.
                </AlertDialogDescription>
              </AlertDialogHeader>
              <AlertDialogFooter>
                <AlertDialogCancel className="rounded-none">Cancel</AlertDialogCancel>
                <AlertDialogAction
                  onClick={handleDelete}
                  disabled={isDeleting}
                  className="rounded-none bg-destructive text-destructive-foreground hover:bg-destructive/90"
                >
                  {isDeleting && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
                  Delete Team
                </AlertDialogAction>
              </AlertDialogFooter>
            </AlertDialogContent>
          </AlertDialog>
        </div>
      )}
    </div>
  )
}