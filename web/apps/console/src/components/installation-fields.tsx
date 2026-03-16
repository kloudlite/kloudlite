'use client'

import type { Control } from 'react-hook-form'
import {
  Input,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
  FormDescription,
} from '@kloudlite/ui'
import { Loader2, CheckCircle2, AlertCircle } from 'lucide-react'

interface InstallationFieldsProps {
  control: Control<{ name: string; subdomain: string }>
  creating: boolean
  checkingSubdomain: boolean
  subdomainAvailable: boolean | null
  onSubdomainChange: (value: string) => void
}

export function InstallationFields({
  control,
  creating,
  checkingSubdomain,
  subdomainAvailable,
  onSubdomainChange,
}: InstallationFieldsProps) {
  return (
    <div className="rounded-lg border border-foreground/10 bg-background">
      <div className="border-b border-foreground/10 px-6 py-4">
        <h3 className="font-medium text-foreground">Installation Details</h3>
      </div>
      <div className="space-y-5 p-6">
        <FormField
          control={control}
          name="name"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Name</FormLabel>
              <FormControl>
                <Input placeholder="e.g., Production" {...field} disabled={creating} />
              </FormControl>
              <FormDescription>A friendly name for this installation</FormDescription>
              <FormMessage />
            </FormItem>
          )}
        />

        <FormField
          control={control}
          name="subdomain"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Domain</FormLabel>
              <FormControl>
                <div className="relative">
                  <Input
                    placeholder="your-company"
                    {...field}
                    disabled={creating}
                    className="font-mono"
                    onChange={(e) => {
                      const value = e.target.value.toLowerCase()
                      field.onChange(value)
                      onSubdomainChange(value)
                    }}
                  />
                  {checkingSubdomain && (
                    <div className="absolute top-1/2 right-3 -translate-y-1/2">
                      <Loader2 className="text-muted-foreground size-4 animate-spin" />
                    </div>
                  )}
                  {!checkingSubdomain && subdomainAvailable === true && (
                    <div className="absolute top-1/2 right-3 -translate-y-1/2">
                      <CheckCircle2 className="size-4 text-green-600" />
                    </div>
                  )}
                  {!checkingSubdomain && subdomainAvailable === false && (
                    <div className="absolute top-1/2 right-3 -translate-y-1/2">
                      <AlertCircle className="text-destructive size-4" />
                    </div>
                  )}
                </div>
              </FormControl>
              <FormDescription className="flex items-center justify-between gap-3">
                <span className="font-mono">
                  {field.value || 'your-subdomain'}.
                  {process.env.NEXT_PUBLIC_INSTALLATION_DOMAIN || 'khost.dev'}
                </span>
                <span className="whitespace-nowrap text-xs font-medium">
                  {!checkingSubdomain && subdomainAvailable === false && (
                    <span className="text-destructive">This domain is already taken</span>
                  )}
                  {!checkingSubdomain && subdomainAvailable === true && (
                    <span className="text-green-600 dark:text-green-500">Domain is available</span>
                  )}
                </span>
              </FormDescription>
              <FormMessage />
            </FormItem>
          )}
        />
      </div>
    </div>
  )
}
