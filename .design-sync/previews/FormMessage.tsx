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
} from '@kloudlite/ui'

export const ValidationError = () => {
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

export const ServerError = () => {
  const form = useForm({ defaultValues: { name: 'staging-eu' } })

  React.useEffect(() => {
    form.setError('name', {
      message: 'An environment named "staging-eu" already exists in this team.',
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

export const StaticChildren = () => {
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
              <FormMessage>
                This name is reserved for platform environments.
              </FormMessage>
            </FormItem>
          )}
        />
      </form>
    </Form>
  )
}
