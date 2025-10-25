'use client'

import { useState } from 'react'
import { useRouter } from 'next/navigation'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import * as z from 'zod'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
  FormDescription,
} from '@/components/ui/form'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Loader2 } from 'lucide-react'
import { InstallationProgress } from '@/components/installation-progress'
import { toast } from 'sonner'

const installationSchema = z.object({
  name: z
    .string()
    .min(3, 'Name must be at least 3 characters')
    .max(50, 'Name must be less than 50 characters')
    .regex(
      /^[a-zA-Z0-9\s-]+$/,
      'Name can only contain letters, numbers, spaces, and hyphens',
    ),
  description: z.string().max(200, 'Description must be less than 200 characters').optional(),
})

type InstallationFormData = z.infer<typeof installationSchema>

export default function NewInstallationPage() {
  const router = useRouter()
  const [creating, setCreating] = useState(false)

  const form = useForm<InstallationFormData>({
    resolver: zodResolver(installationSchema),
    defaultValues: {
      name: '',
      description: '',
    },
  })

  const onSubmit = async (data: InstallationFormData) => {
    setCreating(true)

    try {
      const response = await fetch('/api/installations/create-installation', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          name: data.name,
          description: data.description || undefined,
        }),
      })

      if (!response.ok) {
        const errorData = await response.json()
        throw new Error(errorData.error || 'Failed to create installation')
      }

      await response.json()
      toast.success('Installation created successfully!')

      // Redirect to install step
      router.push('/installations/new/install')
    } catch (err) {
      const error = err instanceof Error ? err : new Error('Failed to create installation')
      toast.error(error.message)
    } finally {
      setCreating(false)
    }
  }

  return (
    <>
      {/* Header */}
      <div className="mb-12 text-center">
        <h1 className="text-foreground mb-3 text-4xl font-bold tracking-tight">
          Create New Installation
        </h1>
        <p className="text-muted-foreground mx-auto max-w-2xl text-lg">
          Give your Kloudlite installation a name and description
        </p>
      </div>

      <InstallationProgress currentStep={1} />

      <Card>
        <CardHeader>
          <CardTitle className="text-xl">Installation Details</CardTitle>
          <CardDescription>
            Provide a name and optional description to help you identify this installation
          </CardDescription>
        </CardHeader>
        <CardContent>
          <Form {...form}>
            <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-6">
              <FormField
                control={form.control}
                name="name"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Installation Name</FormLabel>
                    <FormControl>
                      <Input
                        placeholder="Production Environment"
                        {...field}
                        disabled={creating}
                        className="h-11 px-4 text-base"
                      />
                    </FormControl>
                    <FormDescription>
                      A friendly name to identify this installation
                    </FormDescription>
                    <FormMessage />
                  </FormItem>
                )}
              />

              <FormField
                control={form.control}
                name="description"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Description (Optional)</FormLabel>
                    <FormControl>
                      <Textarea
                        placeholder="Main production deployment for our platform..."
                        {...field}
                        disabled={creating}
                        className="min-h-[100px] resize-none px-4 py-3 text-base"
                      />
                    </FormControl>
                    <FormDescription>
                      A brief description of this installation&apos;s purpose
                    </FormDescription>
                    <FormMessage />
                  </FormItem>
                )}
              />

              <Button
                type="submit"
                className="w-full"
                size="lg"
                disabled={creating}
              >
                {creating ? (
                  <>
                    <Loader2 className="mr-2 size-4 animate-spin" />
                    Creating installation...
                  </>
                ) : (
                  'Continue to Installation'
                )}
              </Button>
            </form>
          </Form>
        </CardContent>
      </Card>
    </>
  )
}
