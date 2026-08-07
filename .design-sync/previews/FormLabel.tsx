import * as React from 'react'
import { useForm } from 'react-hook-form'
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
  Input,
} from '@kloudlite/ui'

export const Default = () => {
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
              <FormMessage />
            </FormItem>
          )}
        />
      </form>
    </Form>
  )
}

export const ErrorState = () => {
  const form = useForm({ defaultValues: { name: 'Staging EU' } })

  React.useEffect(() => {
    form.setError('name', { message: 'Environment names must be lowercase.' })
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
