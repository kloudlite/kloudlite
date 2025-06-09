"use client";
import {
  Form,
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@/components/ui/form";
import { useForm } from "react-hook-form";

export const FormDemo = () => {
  const form = useForm();
  return (
    <Form {...form}>
      <FormField
        control={form.control}
        name="email"
        render={() => {
          return (
            <FormItem>
              <FormLabel>
                Email
              </FormLabel>
              <FormControl>
                <input
                  type="email"
                  placeholder="Email"
                  className="border p-2 rounded-md w-full"
                  {...form.register("email", {
                    required: "Email is required",
                    pattern: {
                      value: /^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$/,
                      message: "Invalid email address",
                    },
                  })}
                />
              </FormControl>
              <FormDescription />
              <FormMessage />
            </FormItem>
          );
        }}
      />
    </Form>
  );
};
