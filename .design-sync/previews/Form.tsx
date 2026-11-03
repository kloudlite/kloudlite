import * as React from 'react'
import { useForm } from 'react-hook-form'
import {
  Button,
  Form,
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
  Input,
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@kloudlite/ui'

export const CreateEnvironment = () => {
  const form = useForm({
    defaultValues: { name: 'staging-eu', cluster: 'prod-ap-south-1' },
  })

  return (
    <Form {...form}>
      <form className="w-96 space-y-6">
        <FormField
          control={form.control}
          name="name"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Environment name</FormLabel>
              <FormControl>
                <Input placeholder="staging-eu" {...field} />
              </FormControl>
              <FormDescription>
                Used in the workspace URL and kubeconfig context.
              </FormDescription>
              <FormMessage />
            </FormItem>
          )}
        />
        <FormField
          control={form.control}
          name="cluster"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Cluster</FormLabel>
              <Select value={field.value} onValueChange={field.onChange}>
                <FormControl>
                  <SelectTrigger>
                    <SelectValue placeholder="Select a cluster" />
                  </SelectTrigger>
                </FormControl>
                <SelectContent>
                  <SelectItem value="prod-ap-south-1">prod-ap-south-1</SelectItem>
                  <SelectItem value="dev-eu-west-1">dev-eu-west-1</SelectItem>
                </SelectContent>
              </Select>
              <FormMessage />
            </FormItem>
          )}
        />
        <Button type="submit">Create environment</Button>
      </form>
    </Form>
  )
}

export const WithValidationError = () => {
  const form = useForm({ defaultValues: { name: 'Staging EU' } })

  React.useEffect(() => {
    form.setError('name', {
      type: 'pattern',
      message: 'Only lowercase letters, numbers and dashes are allowed.',
    })
  }, [form])

  return (
    <Form {...form}>
      <form className="w-96 space-y-6">
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
        <Button type="submit">Create environment</Button>
      </form>
    </Form>
  )
}

export const InviteTeammate = () => {
  const form = useForm({
    defaultValues: { email: 'priya@kloudlite.io', role: 'developer' },
  })

  return (
    <Form {...form}>
      <form className="w-96 space-y-6">
        <FormField
          control={form.control}
          name="email"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Work email</FormLabel>
              <FormControl>
                <Input type="email" {...field} />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />
        <FormField
          control={form.control}
          name="role"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Role</FormLabel>
              <Select value={field.value} onValueChange={field.onChange}>
                <FormControl>
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                </FormControl>
                <SelectContent>
                  <SelectItem value="developer">Developer</SelectItem>
                  <SelectItem value="admin">Admin</SelectItem>
                </SelectContent>
              </Select>
              <FormDescription>
                Developers can open environments but not manage billing.
              </FormDescription>
              <FormMessage />
            </FormItem>
          )}
        />
        <Button type="submit">Send invite</Button>
      </form>
    </Form>
  )
}
