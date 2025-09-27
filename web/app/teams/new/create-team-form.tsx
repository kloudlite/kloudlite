"use client"

import React, { useState } from "react"

import { zodResolver } from "@hookform/resolvers/zod"
import { AlertCircle, Loader2, CheckCircle2 } from "lucide-react"
import { useRouter } from "next/navigation"
import { useForm } from "react-hook-form"
import * as z from "zod"

import { createTeam, checkTeamSlugAvailability, generateTeamSlugSuggestions } from "@/app/actions/teams"
import { Alert, AlertDescription } from "@/components/ui/alert"
import { Button } from "@/components/ui/button"
import {
  Form,
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@/components/ui/form"
import { Input } from "@/components/ui/input"
import { RegionSelect } from "@/components/ui/region-select"
import { Textarea } from "@/components/ui/textarea"
import { useDebounce } from "@/hooks/use-debounce"

const formSchema = z.object({
  slug: z
    .string()
    .min(3, "Team slug must be at least 3 characters")
    .max(30, "Team slug must be less than 30 characters")
    .regex(
      /^[a-z0-9-]+$/,
      "Team slug can only contain lowercase letters, numbers, and hyphens"
    ),
  displayName: z
    .string()
    .min(3, "Display name must be at least 3 characters")
    .max(50, "Display name must be less than 50 characters"),
  description: z
    .string()
    .max(200, "Description must be less than 200 characters")
    .optional(),
  region: z.string().min(1, "Please select a region"),
})

type FormValues = z.infer<typeof formSchema>

interface CreateTeamFormProps {
  regions: Array<{
    id: string
    displayName: string
    value: string
  }>
}

export function CreateTeamForm({ regions }: CreateTeamFormProps) {
  const router = useRouter()
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [slugAvailable, setSlugAvailable] = useState<boolean | null>(null)
  const [checkingSlug, setCheckingSlug] = useState(false)
  const [suggestedSlugs, setSuggestedSlugs] = useState<string[]>([])
  const [initialSuggestions, setInitialSuggestions] = useState<string[]>([])
  const [userHasEditedSlug, setUserHasEditedSlug] = useState(false)

  const form = useForm<FormValues>({
    resolver: zodResolver(formSchema),
    defaultValues: {
      slug: "",
      displayName: "",
      description: "",
      region: "",
    },
  })

  const slug = form.watch("slug")
  const displayName = form.watch("displayName")
  const debouncedSlug = useDebounce(slug, 500)
  const debouncedDisplayName = useDebounce(displayName, 500)
  const [isGeneratingSlug, setIsGeneratingSlug] = useState(false)

  // Generate slug from display name using backend
  React.useEffect(() => {
    if (debouncedDisplayName) {
      setIsGeneratingSlug(true)
      generateTeamSlugSuggestions(debouncedDisplayName)
        .then((suggestions) => {
          setInitialSuggestions(suggestions)
          // Only auto-fill if user hasn't manually edited the slug
          if (!slug && !userHasEditedSlug && suggestions.length > 0) {
            form.setValue("slug", suggestions[0])
          }
        })
        .catch((_error) => {
          // Failed to generate slug suggestions
          setInitialSuggestions([])
        })
        .finally(() => {
          setIsGeneratingSlug(false)
        })
    } else {
      setInitialSuggestions([])
    }
  }, [debouncedDisplayName, slug, form, userHasEditedSlug])

  // Check slug availability
  React.useEffect(() => {
    if (!debouncedSlug || debouncedSlug.length < 3) {
      setSlugAvailable(null)
      // Don't clear suggestions here to prevent them from disappearing
      return
    }

    const checkSlug = async () => {
      setCheckingSlug(true)
      try {
        const result = await checkTeamSlugAvailability(debouncedSlug)
        setSlugAvailable(result.available)
        // Only set suggested slugs if the current one is not available
        if (!result.available) {
          setSuggestedSlugs(result.suggestions || [])
        } else {
          setSuggestedSlugs([])
        }
      } catch (_error) {
        // Failed to check slug
      } finally {
        setCheckingSlug(false)
      }
    }

    checkSlug()
  }, [debouncedSlug])

  async function onSubmit(values: FormValues) {
    setIsSubmitting(true)
    setError(null)

    try {
      const result: any = await createTeam(values)
      
      if (result?.pending) {
        // Team creation request submitted, pending approval
        router.push("/overview?teamPending=true")
      } else {
        // Team created successfully
        router.push("/overview?teamCreated=true")
      }
    } catch (error) {
      setError(error instanceof Error ? error.message : "Something went wrong")
    } finally {
      setIsSubmitting(false)
    }
  }

  return (
    <Form {...form}>
      <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-6">
        {error && (
          <Alert variant="destructive">
            <AlertCircle className="h-4 w-4" />
            <AlertDescription>{error}</AlertDescription>
          </Alert>
        )}

        <FormField
          control={form.control}
          name="displayName"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Display name</FormLabel>
              <FormControl>
                <Input 
                  placeholder="My Awesome Team" 
                  {...field}
                  autoFocus
                />
              </FormControl>
              <FormDescription>
                Choose a name that clearly identifies your team
              </FormDescription>
              <FormMessage />
            </FormItem>
          )}
        />

        <FormField
          control={form.control}
          name="slug"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Team URL slug</FormLabel>
              <FormControl>
                <div className="relative">
                  <Input
                    placeholder="my-awesome-team"
                    className="pr-10"
                    {...field}
                    onChange={(e) => {
                      const value = e.target.value.toLowerCase().replace(/[^a-z0-9-]/g, '')
                      field.onChange(value)
                      setUserHasEditedSlug(true)
                    }}
                  />
                  {(checkingSlug || isGeneratingSlug) && (
                    <Loader2 className="absolute right-3 top-1/2 -translate-y-1/2 h-4 w-4 animate-spin text-muted-foreground" />
                  )}
                  {!checkingSlug && !isGeneratingSlug && slug && slugAvailable === true && (
                    <CheckCircle2 className="absolute right-3 top-1/2 -translate-y-1/2 h-4 w-4 text-green-600" />
                  )}
                  {!checkingSlug && !isGeneratingSlug && slug && slugAvailable === false && (
                    <AlertCircle className="absolute right-3 top-1/2 -translate-y-1/2 h-4 w-4 text-destructive" />
                  )}
                </div>
              </FormControl>
              <FormDescription>
                {displayName ? (
                  <>This will be used in URLs: <span className="font-mono text-xs">/{slug || "..."}</span></>
                ) : (
                  "This will be used in your team's URL"
                )}
              </FormDescription>
              
              {/* Show initial suggestions based on display name */}
              {initialSuggestions.length > 0 && (
                <div className="mt-3 space-y-2">
                  <p className="text-xs text-muted-foreground">Suggested slugs based on your team name:</p>
                  <div className="flex flex-wrap gap-2">
                    {initialSuggestions.slice(0, 4).map((suggestion) => (
                      <Button
                        key={suggestion}
                        type="button"
                        variant={slug === suggestion ? "default" : "outline"}
                        size="sm"
                        className="h-7 px-2.5 font-mono text-xs"
                        onClick={() => {
                          form.setValue("slug", suggestion)
                          setUserHasEditedSlug(true)
                          // Pre-emptively set availability to prevent flicker
                          if (slug !== suggestion) {
                            setSlugAvailable(null)
                          }
                        }}
                      >
                        {suggestion}
                      </Button>
                    ))}
                  </div>
                </div>
              )}
              
              {/* Availability status - only show when slug is taken */}
              {slug && !checkingSlug && slugAvailable === false && (
                <p className="text-sm mt-2 text-destructive">
                  This slug is already taken
                </p>
              )}
              
              {/* Alternative suggestions container - smooth transitions */}
              <div className={`transition-all duration-200 ${suggestedSlugs.length > 0 && slugAvailable === false ? 'opacity-100' : 'opacity-0 h-0 overflow-hidden'}`}>
                {suggestedSlugs.length > 0 && (
                  <Alert className="mt-3">
                    <AlertCircle className="h-4 w-4" />
                    <AlertDescription>
                      <p className="mb-2 font-medium">Available alternatives:</p>
                      <div className="flex flex-wrap gap-2">
                        {suggestedSlugs.map((suggestion) => (
                          <Button
                            key={suggestion}
                            type="button"
                            variant="outline"
                            size="sm"
                            className="h-8 px-3 font-mono text-sm hover:bg-primary hover:text-primary-foreground"
                            onClick={() => {
                              form.setValue("slug", suggestion)
                              setSlugAvailable(true)
                              setSuggestedSlugs([])
                              setUserHasEditedSlug(true)
                            }}
                          >
                            {suggestion}
                          </Button>
                        ))}
                      </div>
                    </AlertDescription>
                  </Alert>
                )}
              </div>
              <FormMessage />
            </FormItem>
          )}
        />

        <FormField
          control={form.control}
          name="description"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Description <span className="text-muted-foreground text-xs font-normal">(optional)</span></FormLabel>
              <FormControl>
                <Textarea 
                  placeholder="Brief description of what this team is for..."
                  className="resize-none"
                  rows={3}
                  {...field} 
                />
              </FormControl>
              <FormDescription>
                Help team members understand the purpose of this team
              </FormDescription>
              <FormMessage />
            </FormItem>
          )}
        />

        <FormField
          control={form.control}
          name="region"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Region</FormLabel>
              <FormControl>
                <RegionSelect
                  regions={regions}
                  value={field.value}
                  onValueChange={field.onChange}
                  placeholder="Search and select a region"
                />
              </FormControl>
              <FormDescription>
                Choose the region closest to your users for better performance
              </FormDescription>
              <FormMessage />
            </FormItem>
          )}
        />

        <div className="flex gap-3">
          <Button
            type="submit"
            disabled={isSubmitting || !slugAvailable}
            className="flex-1"
          >
            {isSubmitting ? (
              <>
                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                Creating team...
              </>
            ) : (
              "Create team"
            )}
          </Button>
          <Button
            type="button"
            variant="ghost"
            onClick={() => router.back()}
            disabled={isSubmitting}
          >
            Cancel
          </Button>
        </div>
      </form>
    </Form>
  )
}