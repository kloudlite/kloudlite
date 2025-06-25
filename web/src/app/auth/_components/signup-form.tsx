"use client";

import { signup } from "@/actions/auth";
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
import { HoverCardContent } from "@/components/ui/hover-card";
import { HoverCard, HoverCardTrigger } from "@radix-ui/react-hover-card";
import { LockIcon, LogIn } from "lucide-react";
import Link from "next/link";
import { useForm } from "react-hook-form";
import { toast } from "sonner";

export default function SignupForm() {
  const form = useForm();
  return (
    <Form {...form}>
      <form
        onSubmit={form.handleSubmit(async (data) => {
          const [_, error] = await signup(data.name, data.email, data.password);
          if (error) {
            toast.error(`Signup failed: ${error.message}`);
          }
        })}
      >
        <Card className="w-[400px] mx-auto mt-20">
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <LockIcon />
              Sign up
            </CardTitle>
            <CardDescription>to setup your account</CardDescription>
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
            <FormField
              control={form.control}
              name="password"
              render={() => {
                return (
                  <FormItem>
                    <FormLabel>Password</FormLabel>
                    <FormControl>
                      <input
                        type="password"
                        placeholder="Password"
                        className="border p-2 rounded-md w-full"
                        {...form.register("password", {
                          required: "Password is required",
                          minLength: {
                            value: 6,
                            message: "Password must be at least 6 characters",
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
            <div className="w-full flex justify-between items-center">
              <HoverCard>
                <HoverCardTrigger className="cursor-pointer text-sm font-medium">
                  Terms of Service
                </HoverCardTrigger>
                <HoverCardContent className="text-sm">
                  By signing up, you agree to our{" "}
                  <Link href="/terms" className="text-blue-500">
                    Terms of Service
                  </Link>
                  .
                </HoverCardContent>
              </HoverCard>
              <Button type="submit">
                <LogIn />
                Signup
              </Button>
            </div>
          </CardFooter>
        </Card>
        <div className="p-4 text-sm flex gap-2 justify-center">
          <span>
            Already have an account?
          </span>
          <Link href="/auth/login" className="text-blue-500">
            Login
          </Link>
        </div>
      </form>
    </Form>
  );
}
