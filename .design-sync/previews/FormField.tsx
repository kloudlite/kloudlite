import * as React from 'react'
import { useForm } from 'react-hook-form'
import {
  Form,
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
  Input,
  Switch,
} from '@kloudlite/ui'

export const TextField = () => {
  const form = useForm({ defaultValues: { name: 'staging-eu' } })

  return (
    <Form {...form}>
      <form className="w-96">
        <FormField
          control={form.control}
          name="name"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Environment name</FormLabel>
              <FormControl>
                <Input {...field} />
              </FormControl>
              <FormDescription>
                Used in the workspace URL and kubeconfig context.
              </FormDescription>
              <FormMessage />
            </FormItem>
          )}
        />
      </form>
    </Form>
  )
}

export const SwitchField = () => {
  const form = useForm({ defaultValues: { autoSleep: true } })

  return (
    <Form {...form}>
      <form className="w-96">
        <FormField
          control={form.control}
          name="autoSleep"
          render={({ field }) => (
            <FormItem className="flex items-center justify-between rounded-md border border-input p-4">
              <div className="space-y-1">
                <FormLabel>Auto-sleep</FormLabel>
                <FormDescription>
                  Suspend workspaces after 2 hours of inactivity.
                </FormDescription>
              </div>
              <FormControl>
                <Switch
                  checked={field.value}
                  onCheckedChange={field.onChange}
                />
              </FormControl>
            </FormItem>
          )}
        />
      </form>
    </Form>
  )
}

export const FieldWithError = () => {
  const form = useForm({ defaultValues: { name: 'Staging EU' } })

  React.useEffect(() => {
    form.setError('name', {
      message: 'Only lowercase letters, numbers and dashes are allowed.',
    })
  }, [form])

  return (
    <Form {...form}>
      <form className="w-96">
        <FormField
          control={form.control}
          name="name"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Environment name</FormLabel>
              <FormControl>
                <Input {...field} />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />
      </form>
    </Form>
  )
}
