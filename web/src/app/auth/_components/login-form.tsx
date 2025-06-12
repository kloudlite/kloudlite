"use client";

import { login, signup } from "@/actions/auth";
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
import { Lock, LogIn, ScanFace } from "lucide-react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { useCallback, useState } from "react";
import { useForm } from "react-hook-form";
import { toast } from "sonner";
import util from "util";

export default function LoginForm(
  { withSSO = false, emailCommEnabled = true }: {
    withSSO?: boolean;
    emailCommEnabled?: boolean;
  },
) {
  const form = useForm();
  const router = useRouter();
  const [loggingIn, setLoggingIn] = useState(false);
  const loginCall = useCallback(async (email: string, password: string) => {
    setLoggingIn(true);
    const [_, loginErr] = await login(email, password);
    setLoggingIn(false);
    if (loginErr) {
      toast.error(`Login failed. ${loginErr.message}`);
    } else {
      router.push("/teams");
    }
  }, []);
  return (
    <Form {...form}>
      <form
        onSubmit={form.handleSubmit((data) => {
          loginCall(data.email, data.password);
        })}
      >
        <Card className="w-[400px] mx-auto mt-20">
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <ScanFace />
              Sign in
            </CardTitle>
            <CardDescription>to access account</CardDescription>
          </CardHeader>
          <CardContent className="flex flex-col gap-4">
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
                        disabled={loggingIn}
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
                        disabled={loggingIn}
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
            <div className="w-full flex justify-between text-sm items-center">
              {emailCommEnabled && (
                <Link href="/auth/reset-password">
                  <Button
                    variant={"link"}
                    className="underline text-sm cursor-pointer"
                  >
                    Forgot Password?
                  </Button>
                </Link>
              )}
              {!emailCommEnabled && (
                <HoverCard>
                  <HoverCardTrigger className="cursor-pointer">
                    Forgot Password?
                  </HoverCardTrigger>
                  <HoverCardContent className="text-sm">
                    Contact administrator to reset your password.
                  </HoverCardContent>
                </HoverCard>
              )}

              <Button type="submit">
                <LogIn />
                Login
              </Button>
            </div>
            {withSSO && (
              <>
                <div className="flex items-center gap-2 text-sm">
                  <hr className="w-[100px]" />
                  OR
                  <hr className="w-[100px]" />
                </div>

                <div className="flex gap-2">
                  <Button variant={"secondary"} asChild disabled={loggingIn}>
                    <Link href="/dev-only/external-login">
                      <Lock />
                      Login With SSO
                    </Link>
                  </Button>
                </div>
              </>
            )}
          </CardFooter>
        </Card>
        {!withSSO && (
          <div className="p-4 text-sm flex gap-2 justify-center">
            <span>
              Don't have an account?
            </span>
            <Link href="/auth/signup" className="text-blue-500">
              Signup
            </Link>
          </div>
        )}
      </form>
    </Form>
  );
}
