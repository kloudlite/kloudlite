'use client'

import { useState } from 'react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Building2 } from 'lucide-react'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '@/components/ui/dialog'
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from '@/components/ui/form'
import { useForm } from 'react-hook-form'

interface SSOFormData {
  email: string
}

export function SSOLogin() {
  const [open, setOpen] = useState(false)
  const [isLoading, setIsLoading] = useState(false)
  
  const form = useForm<SSOFormData>({
    defaultValues: {
      email: '',
    },
  })

  const onSubmit = async (data: SSOFormData) => {
    setIsLoading(true)
    // TODO: Implement actual SSO flow
    await new Promise(resolve => setTimeout(resolve, 1500))
    setIsLoading(false)
    setOpen(false)
    form.reset()
  }

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        <Button
          type="button"
          variant="outline"
          size="default"
          className="w-full h-11"
        >
          <Building2 className="h-5 w-5 mr-2" />
          Continue with SSO
        </Button>
      </DialogTrigger>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>Sign in with SSO</DialogTitle>
          <DialogDescription>
            Enter your work email address and we'll redirect you to your organization's SSO provider.
          </DialogDescription>
        </DialogHeader>
        <Form {...form}>
          <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
            <FormField
              control={form.control}
              name="email"
              rules={{
                required: 'Email is required',
                pattern: {
                  value: /^[A-Z0-9._%+-]+@[A-Z0-9.-]+\.[A-Z]{2,}$/i,
                  message: 'Invalid email address',
                },
              }}
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Work Email</FormLabel>
                  <FormControl>
                    <Input
                      type="email"
                      placeholder="you@company.com"
                      disabled={isLoading}
                      {...field}
                    />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />
            <Button 
              type="submit" 
              className="w-full" 
              size="auth"
              disabled={isLoading}
            >
              {isLoading ? 'Redirecting...' : 'Continue with SSO'}
            </Button>
          </form>
        </Form>
      </DialogContent>
    </Dialog>
  )
}