"use client";

import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import {
  Form,
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@/components/ui/form";
import { LockKeyhole, LogIn } from "lucide-react";
import { useRouter } from "next/navigation";
import { useForm } from "react-hook-form";

export default function ExternalLoginForm(
  { tokenGenerator }: {
    tokenGenerator: (email: string, name: string) => Promise<string>;
  },
) {
  const form = useForm();
  const router = useRouter();
  return (
    <Form {...form}>
      <form
        onSubmit={form.handleSubmit(async (data) => {
          try {
            const token = await tokenGenerator(data.email, data.name);
            router.push(`/auth/sso-login?token=${token}`);
          } catch (error) {
            console.error("Error generating token:", error);
          }
        })}
      >
        <Card className="w-[400px] mx-auto mt-20">
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <LockKeyhole />
              SSO Login Play Ground
            </CardTitle>
            <CardDescription>create SSO account to test</CardDescription>
          </CardHeader>
          <CardContent className="flex flex-col gap-4">
            <FormField
              control={form.control}
              name="name"
              render={() => {
                return (
                  <FormItem>
                    <FormLabel>
                      Name
                    </FormLabel>
                    <FormControl>
                      <input
                        type="name"
                        placeholder="Name"
                        className="border p-2 rounded-md w-full"
                        {...form.register("name", {
                          required: "Name is required",
                        })}
                      />
                    </FormControl>
                    <FormDescription />
                    <FormMessage />
                  </FormItem>
                );
              }}
            />
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
                            value:
                              /^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$/,
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
          </CardContent>
          <CardFooter className="flex flex-col gap-4 items-center">
            <div className="w-full flex justify-between text-sm items-center">
              <Button type="submit">
                <LogIn />
                Login
              </Button>
            </div>
          </CardFooter>
        </Card>
      </form>
    </Form>
  );
}
