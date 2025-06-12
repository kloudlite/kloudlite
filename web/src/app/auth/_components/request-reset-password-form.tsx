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
import { Lock, LogIn, LogInIcon, RotateCcw, ScanFace, Send } from "lucide-react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { useCallback, useState } from "react";
import { useForm } from "react-hook-form";
import { toast } from "sonner";
import util from "util";

export default function RequestResetPasswordForm(
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
              <RotateCcw />
              Forgot Password?
            </CardTitle>
            <CardDescription>It happens don't worry</CardDescription>
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
                        placeholder="Registered Email"
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
            <div className="w-full flex justify-between text-xs items-center">
              <Button size={"icon"} className="rounded-full">
                <LogInIcon />
              </Button>
              <Button type="submit">
                <Send />
                Send Reset Instructions
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
                    <Link href="/teams">
                      <Lock />
                      Login With SSO
                    </Link>
                  </Button>
                </div>
              </>
            )}
          </CardFooter>
        </Card>
      </form>
    </Form>
  );
}
