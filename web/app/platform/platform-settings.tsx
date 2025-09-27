"use client";

import { useState } from "react";

import { zodResolver } from "@hookform/resolvers/zod";
import { Settings, Users, Shield, Zap, Save } from "lucide-react";
import { useRouter } from "next/navigation";
import { type Session } from "next-auth";
import { useForm } from "react-hook-form";
import { toast } from "sonner";
import * as z from "zod";

import { updatePlatformSettings } from "@/app/actions/teams";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
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
import { Input } from "@/components/ui/input";
import { Separator } from "@/components/ui/separator";
import { Switch } from "@/components/ui/switch";
import {
  Tabs,
  TabsContent,
  TabsList,
  TabsTrigger,
} from "@/components/ui/tabs";


const settingsSchema = z.object({
  platformOwnerEmail: z.string().email(),
  supportEmail: z.string().email(),
  allowSignup: z.boolean(),
  teamSettings: z.object({
    requireApproval: z.boolean(),
    autoApproveFirstTeam: z.boolean(),
    maxTeamsPerUser: z.number().min(1).max(100),
  }),
  oauthProviders: z.object({
    google: z.object({
      enabled: z.boolean(),
      clientId: z.string().optional(),
      clientSecret: z.string().optional(),
    }),
    github: z.object({
      enabled: z.boolean(),
      clientId: z.string().optional(),
      clientSecret: z.string().optional(),
    }),
    microsoft: z.object({
      enabled: z.boolean(),
      clientId: z.string().optional(),
      clientSecret: z.string().optional(),
      tenantId: z.string().optional(),
    }),
  }),
  features: z.object({
    enableDeviceFlow: z.boolean(),
    enableCLI: z.boolean(),
    enableAPI: z.boolean(),
  }),
});

type SettingsFormValues = z.infer<typeof settingsSchema>;

interface PlatformSettingsProps {
  settings: any;
  session: Session;
}

export default function PlatformSettings({
  settings,
}: PlatformSettingsProps) {
  const router = useRouter();
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [activeTab, setActiveTab] = useState("general");

  const form = useForm<SettingsFormValues>({
    resolver: zodResolver(settingsSchema),
    defaultValues: {
      platformOwnerEmail: settings?.platformOwnerEmail || "",
      supportEmail: settings?.supportEmail || "",
      allowSignup: settings?.allowSignup || false,
      teamSettings: {
        requireApproval: settings?.teamSettings?.requireApproval || false,
        autoApproveFirstTeam: settings?.teamSettings?.autoApproveFirstTeam || false,
        maxTeamsPerUser: settings?.teamSettings?.maxTeamsPerUser || 5,
      },
      oauthProviders: {
        google: {
          enabled: settings?.oauthProviders?.google?.enabled || false,
          clientId: settings?.oauthProviders?.google?.clientId || "",
          clientSecret: "",
        },
        github: {
          enabled: settings?.oauthProviders?.github?.enabled || false,
          clientId: settings?.oauthProviders?.github?.clientId || "",
          clientSecret: "",
        },
        microsoft: {
          enabled: settings?.oauthProviders?.microsoft?.enabled || false,
          clientId: settings?.oauthProviders?.microsoft?.clientId || "",
          clientSecret: "",
        },
      },
      features: {
        enableDeviceFlow: settings?.features?.enableDeviceFlow || false,
        enableCLI: settings?.features?.enableCLI || false,
        enableAPI: settings?.features?.enableAPI || false,
      },
    },
  });

  async function onSubmit(data: SettingsFormValues) {
    setIsSubmitting(true);
    try {
      // Merge with existing settings to preserve fields not in the form
      await updatePlatformSettings({
        ...settings,
        ...data,
      });
      toast.success("Platform settings updated successfully");
      router.refresh();
    } catch (error) {
      toast.error("Failed to update platform settings");
    } finally {
      setIsSubmitting(false);
    }
  }

  return (
    <Form {...form}>
      <form onSubmit={form.handleSubmit(onSubmit)}>
          <Tabs value={activeTab} onValueChange={setActiveTab}>
            <div className="flex items-center justify-between mb-6">
              <TabsList className="bg-muted/50">
                <TabsTrigger value="general">
                  <Settings className="h-4 w-4 mr-2" />
                  General
                </TabsTrigger>
                <TabsTrigger value="teams">
                  <Users className="h-4 w-4 mr-2" />
                  Teams
                </TabsTrigger>
                <TabsTrigger value="auth">
                  <Shield className="h-4 w-4 mr-2" />
                  Authentication
                </TabsTrigger>
                <TabsTrigger value="features">
                  <Zap className="h-4 w-4 mr-2" />
                  Features
                </TabsTrigger>
              </TabsList>
              
              <Button 
                type="submit" 
                disabled={isSubmitting}
                size="sm"
                className="gap-2"
              >
                <Save className="h-4 w-4" />
                {isSubmitting ? "Saving..." : "Save"}
              </Button>
            </div>

            <TabsContent value="general" className="mt-0">
              <div className="grid gap-6">
                <div>
                  <h3 className="text-lg font-medium mb-4">General Settings</h3>
                  <Card>
                    <CardContent className="pt-6 space-y-6">
                      <FormField
                        control={form.control}
                        name="platformOwnerEmail"
                        render={({ field }) => (
                          <FormItem>
                            <FormLabel>Platform Owner Email</FormLabel>
                            <FormControl>
                              <Input {...field} type="email" placeholder="owner@example.com" />
                            </FormControl>
                            <FormDescription>
                              The primary administrator email for this platform
                            </FormDescription>
                            <FormMessage />
                          </FormItem>
                        )}
                      />

                      <FormField
                        control={form.control}
                        name="supportEmail"
                        render={({ field }) => (
                          <FormItem>
                            <FormLabel>Support Email</FormLabel>
                            <FormControl>
                              <Input {...field} type="email" placeholder="support@example.com" />
                            </FormControl>
                            <FormDescription>
                              Contact email for user support inquiries
                            </FormDescription>
                            <FormMessage />
                          </FormItem>
                        )}
                      />

                      <Separator />

                      <FormField
                        control={form.control}
                        name="allowSignup"
                        render={({ field }) => (
                          <FormItem className="flex flex-row items-center justify-between">
                            <div className="space-y-0.5">
                              <FormLabel>Public Registration</FormLabel>
                              <FormDescription>
                                Allow new users to create accounts
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
                    </CardContent>
                  </Card>
                </div>
              </div>
            </TabsContent>

            <TabsContent value="teams" className="mt-0">
              <div className="grid gap-6">
                <div>
                  <h3 className="text-lg font-medium mb-4">Team Management</h3>
                  <Card>
                    <CardContent className="pt-6 space-y-6">
                      <FormField
                        control={form.control}
                        name="teamSettings.requireApproval"
                        render={({ field }) => (
                          <FormItem className="flex flex-row items-center justify-between">
                            <div className="space-y-0.5">
                              <FormLabel>Approval Required</FormLabel>
                              <FormDescription>
                                Require admin approval for new teams
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

                      <Separator />

                      <FormField
                        control={form.control}
                        name="teamSettings.autoApproveFirstTeam"
                        render={({ field }) => (
                          <FormItem className="flex flex-row items-center justify-between">
                            <div className="space-y-0.5">
                              <FormLabel>Auto-approve First Team</FormLabel>
                              <FormDescription>
                                Automatically approve user's first team
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

                      <Separator />

                      <FormField
                        control={form.control}
                        name="teamSettings.maxTeamsPerUser"
                        render={({ field }) => (
                          <FormItem>
                            <FormLabel>Team Limit per User</FormLabel>
                            <FormControl>
                              <Input
                                {...field}
                                type="number"
                                className="w-32"
                                min="1"
                                max="100"
                                onChange={(e) => field.onChange(parseInt(e.target.value))}
                              />
                            </FormControl>
                            <FormDescription>
                              Maximum number of teams a single user can create
                            </FormDescription>
                            <FormMessage />
                          </FormItem>
                        )}
                      />
                    </CardContent>
                  </Card>
                </div>
              </div>
            </TabsContent>

            <TabsContent value="auth" className="mt-0">
              <div className="grid gap-6">
                <div>
                  <h3 className="text-lg font-medium mb-4">Authentication Providers</h3>
                  <div className="space-y-4">
                    {/* Google OAuth */}
                    <Card>
                      <CardHeader className="pb-4">
                        <div className="flex items-center justify-between">
                          <div className="flex items-center gap-3">
                            <div className="w-8 h-8 rounded-full bg-background flex items-center justify-center">
                              <svg className="w-5 h-5" viewBox="0 0 24 24">
                                <path fill="#4285F4" d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z"/>
                                <path fill="#34A853" d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z"/>
                                <path fill="#FBBC05" d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l2.85-2.22.81-.62z"/>
                                <path fill="#EA4335" d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z"/>
                              </svg>
                            </div>
                            <CardTitle className="text-base">Google</CardTitle>
                          </div>
                          <FormField
                            control={form.control}
                            name="oauthProviders.google.enabled"
                            render={({ field }) => (
                              <FormControl>
                                <Switch
                                  checked={field.value}
                                  onCheckedChange={field.onChange}
                                />
                              </FormControl>
                            )}
                          />
                        </div>
                      </CardHeader>
                      {form.watch("oauthProviders.google.enabled") && (
                        <CardContent className="space-y-4">
                          <FormField
                            control={form.control}
                            name="oauthProviders.google.clientId"
                            render={({ field }) => (
                              <FormItem>
                                <FormLabel>Client ID</FormLabel>
                                <FormControl>
                                  <Input {...field} placeholder="Enter Google OAuth Client ID" />
                                </FormControl>
                                <FormMessage />
                              </FormItem>
                            )}
                          />
                          <FormField
                            control={form.control}
                            name="oauthProviders.google.clientSecret"
                            render={({ field }) => (
                              <FormItem>
                                <FormLabel>Client Secret</FormLabel>
                                <FormControl>
                                  <Input {...field} type="password" placeholder="Leave empty to keep current" />
                                </FormControl>
                                <FormMessage />
                              </FormItem>
                            )}
                          />
                        </CardContent>
                      )}
                    </Card>

                    {/* GitHub OAuth */}
                    <Card>
                      <CardHeader className="pb-4">
                        <div className="flex items-center justify-between">
                          <div className="flex items-center gap-3">
                            <div className="w-8 h-8 rounded-full bg-background flex items-center justify-center">
                              <svg className="w-5 h-5" fill="currentColor" viewBox="0 0 24 24">
                                <path d="M12 0c-6.626 0-12 5.373-12 12 0 5.302 3.438 9.8 8.207 11.387.599.111.793-.261.793-.577v-2.234c-3.338.726-4.033-1.416-4.033-1.416-.546-1.387-1.333-1.756-1.333-1.756-1.089-.745.083-.729.083-.729 1.205.084 1.839 1.237 1.839 1.237 1.07 1.834 2.807 1.304 3.492.997.107-.775.418-1.305.762-1.604-2.665-.305-5.467-1.334-5.467-5.931 0-1.311.469-2.381 1.236-3.221-.124-.303-.535-1.524.117-3.176 0 0 1.008-.322 3.301 1.23.957-.266 1.983-.399 3.003-.404 1.02.005 2.047.138 3.006.404 2.291-1.552 3.297-1.23 3.297-1.23.653 1.653.242 2.874.118 3.176.77.84 1.235 1.911 1.235 3.221 0 4.609-2.807 5.624-5.479 5.921.43.372.823 1.102.823 2.222v3.293c0 .319.192.694.801.576 4.765-1.589 8.199-6.086 8.199-11.386 0-6.627-5.373-12-12-12z"/>
                              </svg>
                            </div>
                            <CardTitle className="text-base">GitHub</CardTitle>
                          </div>
                          <FormField
                            control={form.control}
                            name="oauthProviders.github.enabled"
                            render={({ field }) => (
                              <FormControl>
                                <Switch
                                  checked={field.value}
                                  onCheckedChange={field.onChange}
                                />
                              </FormControl>
                            )}
                          />
                        </div>
                      </CardHeader>
                      {form.watch("oauthProviders.github.enabled") && (
                        <CardContent className="space-y-4">
                          <FormField
                            control={form.control}
                            name="oauthProviders.github.clientId"
                            render={({ field }) => (
                              <FormItem>
                                <FormLabel>Client ID</FormLabel>
                                <FormControl>
                                  <Input {...field} placeholder="Enter GitHub OAuth Client ID" />
                                </FormControl>
                                <FormMessage />
                              </FormItem>
                            )}
                          />
                          <FormField
                            control={form.control}
                            name="oauthProviders.github.clientSecret"
                            render={({ field }) => (
                              <FormItem>
                                <FormLabel>Client Secret</FormLabel>
                                <FormControl>
                                  <Input {...field} type="password" placeholder="Leave empty to keep current" />
                                </FormControl>
                                <FormMessage />
                              </FormItem>
                            )}
                          />
                        </CardContent>
                      )}
                    </Card>

                    {/* Microsoft OAuth */}
                    <Card>
                      <CardHeader className="pb-4">
                        <div className="flex items-center justify-between">
                          <div className="flex items-center gap-3">
                            <div className="w-8 h-8 rounded-full bg-background flex items-center justify-center">
                              <svg className="w-5 h-5" viewBox="0 0 23 23">
                                <path fill="#f25022" d="M1 1h10v10H1z"/>
                                <path fill="#00a4ef" d="M12 1h10v10H12z"/>
                                <path fill="#7fba00" d="M1 12h10v10H1z"/>
                                <path fill="#ffb900" d="M12 12h10v10H12z"/>
                              </svg>
                            </div>
                            <CardTitle className="text-base">Microsoft</CardTitle>
                          </div>
                          <FormField
                            control={form.control}
                            name="oauthProviders.microsoft.enabled"
                            render={({ field }) => (
                              <FormControl>
                                <Switch
                                  checked={field.value}
                                  onCheckedChange={field.onChange}
                                />
                              </FormControl>
                            )}
                          />
                        </div>
                      </CardHeader>
                      {form.watch("oauthProviders.microsoft.enabled") && (
                        <CardContent className="space-y-4">
                          <FormField
                            control={form.control}
                            name="oauthProviders.microsoft.clientId"
                            render={({ field }) => (
                              <FormItem>
                                <FormLabel>Client ID</FormLabel>
                                <FormControl>
                                  <Input {...field} placeholder="Enter Microsoft OAuth Client ID" />
                                </FormControl>
                                <FormMessage />
                              </FormItem>
                            )}
                          />
                          <FormField
                            control={form.control}
                            name="oauthProviders.microsoft.clientSecret"
                            render={({ field }) => (
                              <FormItem>
                                <FormLabel>Client Secret</FormLabel>
                                <FormControl>
                                  <Input {...field} type="password" placeholder="Leave empty to keep current" />
                                </FormControl>
                                <FormMessage />
                              </FormItem>
                            )}
                          />
                          <FormField
                            control={form.control}
                            name="oauthProviders.microsoft.tenantId"
                            render={({ field }) => (
                              <FormItem>
                                <FormLabel>Tenant ID</FormLabel>
                                <FormControl>
                                  <Input {...field} placeholder="common" />
                                </FormControl>
                                <FormDescription>
                                  Use 'common' for multi-tenant applications
                                </FormDescription>
                                <FormMessage />
                              </FormItem>
                            )}
                          />
                        </CardContent>
                      )}
                    </Card>
                  </div>
                </div>
              </div>
            </TabsContent>

            <TabsContent value="features" className="mt-0">
              <div className="grid gap-6">
                <div>
                  <h3 className="text-lg font-medium mb-4">Platform Features</h3>
                  <Card>
                    <CardContent className="pt-6 space-y-6">
                      <FormField
                        control={form.control}
                        name="features.enableDeviceFlow"
                        render={({ field }) => (
                          <FormItem className="flex flex-row items-center justify-between">
                            <div className="space-y-0.5">
                              <FormLabel>Device Flow Authentication</FormLabel>
                              <FormDescription>
                                Allow CLI tools to authenticate using device flow
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

                      <Separator />

                      <FormField
                        control={form.control}
                        name="features.enableCLI"
                        render={({ field }) => (
                          <FormItem className="flex flex-row items-center justify-between">
                            <div className="space-y-0.5">
                              <FormLabel>Command Line Interface</FormLabel>
                              <FormDescription>
                                Enable CLI access for developers
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

                      <Separator />

                      <FormField
                        control={form.control}
                        name="features.enableAPI"
                        render={({ field }) => (
                          <FormItem className="flex flex-row items-center justify-between">
                            <div className="space-y-0.5">
                              <FormLabel>API Access</FormLabel>
                              <FormDescription>
                                Enable programmatic API access
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
                    </CardContent>
                  </Card>
                </div>
              </div>
            </TabsContent>
          </Tabs>
        </form>
      </Form>
  );
}