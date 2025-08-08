"use client";

import { useState } from "react";

import { zodResolver } from "@hookform/resolvers/zod";
import { Cloud, Plus, Settings, Lock, CheckCircle2, AlertCircle, Info, Zap } from "lucide-react";
import { useForm } from "react-hook-form";
import { toast } from "sonner";
import * as z from "zod";

import { Alert, AlertDescription } from "@/components/ui/alert";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { Form, FormControl, FormDescription, FormField, FormItem, FormLabel, FormMessage } from "@/components/ui/form";
import { Input } from "@/components/ui/input";
import { cn } from "@/lib/utils";

// Cloud provider types
type CloudProviderType = "aws" | "gcp" | "azure" | "digitalocean";

interface CloudProviderInfo {
  id: CloudProviderType;
  name: string;
  shortName?: string;
  icon: React.ReactNode;
  status: "active" | "coming_soon";
  fields: {
    name: string;
    label: string;
    type: "text" | "password";
    placeholder: string;
    required: boolean;
  }[];
}

// Simple SVG icons for cloud providers
const AWSIcon = () => (
  <svg viewBox="0 0 24 24" className="w-8 h-8" fill="currentColor">
    <path d="M6.763 10.036c0 .296.032.535.088.71.064.176.144.368.256.576.04.063.056.127.056.183 0 .08-.048.16-.152.24l-.503.335a.383.383 0 0 1-.208.072c-.08 0-.16-.04-.239-.112a2.47 2.47 0 0 1-.287-.375 6.18 6.18 0 0 1-.248-.471c-.622.734-1.405 1.101-2.347 1.101-.67 0-1.205-.191-1.596-.574-.391-.384-.59-.894-.59-1.533 0-.678.239-1.23.726-1.644.487-.415 1.133-.623 1.955-.623.272 0 .551.024.846.064.296.04.6.104.918.176v-.583c0-.607-.127-1.03-.375-1.277-.255-.248-.686-.367-1.3-.367-.28 0-.568.031-.863.103-.295.072-.583.16-.862.272a2.287 2.287 0 0 1-.28.104.488.488 0 0 1-.127.023c-.112 0-.168-.08-.168-.247v-.391c0-.128.016-.224.056-.28a.597.597 0 0 1 .224-.167c.279-.144.614-.264 1.005-.36a4.84 4.84 0 0 1 1.246-.151c.95 0 1.644.216 2.091.647.439.43.662 1.085.662 1.963v2.586zm-3.24 1.214c.263 0 .534-.048.822-.144.287-.096.543-.271.758-.51.128-.152.224-.32.272-.512.047-.191.08-.423.08-.694v-.335a6.66 6.66 0 0 0-.735-.136 6.02 6.02 0 0 0-.75-.048c-.535 0-.926.104-1.19.32-.263.215-.39.518-.39.917 0 .375.095.655.295.846.191.2.47.296.838.296zm6.41.862c-.144 0-.24-.024-.304-.08-.064-.048-.12-.16-.168-.311L7.586 5.55a1.398 1.398 0 0 1-.072-.32c0-.128.064-.2.191-.2h.783c.151 0 .255.025.31.08.065.048.113.16.16.312l1.342 5.284 1.245-5.284c.04-.16.088-.264.151-.312a.549.549 0 0 1 .32-.08h.638c.152 0 .256.025.32.08.063.048.12.16.151.312l1.261 5.348 1.381-5.348c.048-.16.104-.264.16-.312a.52.52 0 0 1 .311-.08h.743c.127 0 .2.065.2.2 0 .04-.009.08-.017.128a1.137 1.137 0 0 1-.056.2l-1.923 6.17c-.048.16-.104.263-.168.311a.51.51 0 0 1-.303.08h-.687c-.151 0-.255-.024-.32-.08-.063-.056-.119-.16-.15-.32l-1.238-5.148-1.23 5.14c-.04.16-.087.264-.15.32-.065.056-.177.08-.32.08zm10.256.215c-.415 0-.83-.048-1.229-.143-.399-.096-.71-.2-.918-.32-.128-.071-.215-.151-.247-.223a.563.563 0 0 1-.048-.224v-.407c0-.167.064-.247.183-.247.048 0 .096.008.144.024.048.016.12.048.2.08.271.12.566.215.878.279.319.064.63.096.95.096.502 0 .894-.088 1.165-.264a.86.86 0 0 0 .415-.758.777.777 0 0 0-.215-.559c-.144-.151-.416-.287-.807-.415l-1.157-.36c-.583-.183-1.014-.454-1.277-.813a1.902 1.902 0 0 1-.4-1.158c0-.335.073-.63.216-.886.144-.255.335-.479.575-.654.24-.184.51-.32.83-.415.32-.096.655-.136 1.006-.136.175 0 .359.008.535.032.183.024.35.056.518.088.16.04.312.08.455.127.144.048.256.096.336.144a.69.69 0 0 1 .24.2.43.43 0 0 1 .071.263v.375c0 .168-.064.256-.184.256a.83.83 0 0 1-.303-.096 3.652 3.652 0 0 0-1.532-.311c-.455 0-.815.071-1.062.223-.248.152-.375.383-.375.694 0 .224.08.416.24.567.159.152.454.304.877.44l1.134.358c.574.184.99.44 1.237.767.247.327.367.702.367 1.117 0 .343-.072.655-.207.926-.144.272-.336.511-.583.703-.248.2-.543.343-.886.447-.36.111-.734.167-1.142.167z"/>
  </svg>
);

const GCPIcon = () => (
  <svg viewBox="0 0 24 24" className="w-8 h-8" fill="currentColor">
    <path d="M12.19 2.38a9.344 9.344 0 0 0-9.234 6.893c.053-.02.12-.038.174-.038h3.16c.24 0 .441.173.481.412.173-.704.522-1.356.985-1.92a5.151 5.151 0 0 1 4.26-2.297h.023a5.15 5.15 0 0 1 4.26 2.297l.107.146 2.67-1.885-.107-.147a9.344 9.344 0 0 0-6.864-3.46zM8.438 12.166a.523.523 0 0 0-.438.293l-1.73 3.002a.524.524 0 0 0 .186.708.525.525 0 0 0 .252.068h3.46c.413 0 .667-.48.437-.827l-1.73-3.003a.523.523 0 0 0-.438-.24zm3.619 0a.52.52 0 0 0-.44.241l-1.728 3.002a.527.527 0 0 0-.076.425c.056.187.209.325.396.361a.518.518 0 0 0 .265-.032.522.522 0 0 0 .252-.214l1.73-3.003a.524.524 0 0 0-.186-.71.532.532 0 0 0-.213-.068zm3.733 0a.524.524 0 0 0-.438.293l-1.73 3.002a.527.527 0 0 0-.033.485c.094.174.276.281.474.281h3.462a.525.525 0 0 0 .453-.786l-1.73-3.003a.524.524 0 0 0-.438-.24.506.506 0 0 0-.019-.002zm2.813.039a9.34 9.34 0 0 0 .024-.52c0-1.82-.52-3.512-1.42-4.945l-2.668 1.886.013.067c.33.56.533 1.198.586 1.881.053.704-.08 1.406-.387 2.026l-.093.186h3.924s.025-.12.025-.373v-.185l-.001-.022zm-7.198 4.022a.52.52 0 0 0-.44.241l-1.729 3.003a.531.531 0 0 0-.04.427.526.526 0 0 0 .279.307.519.519 0 0 0 .426-.014.524.524 0 0 0 .24-.21l1.73-3.003a.524.524 0 0 0-.186-.708.524.524 0 0 0-.252-.069l-.026.001v.024z"/>
  </svg>
);

const AzureIcon = () => (
  <svg viewBox="0 0 24 24" className="w-8 h-8" fill="currentColor">
    <path d="M5.483 12.645L10.303 1.872a.392.392 0 0 1 .35-.215c.152 0 .288.087.353.225l2.263 4.829c.035.073.051.154.05.235a.508.508 0 0 1-.036.18L6.071 21.344a.393.393 0 0 1-.357.23H.392A.392.392 0 0 1 0 21.182c0-.07.019-.14.054-.201l5.43-8.336zm6.174-8.467c.107-.187.347-.253.533-.146a.39.39 0 0 1 .135.139l10.204 17.75a.39.39 0 0 1-.337.596h-7.345a.504.504 0 0 1-.092-.007.394.394 0 0 1-.295-.242l-2.785-6.013a.39.39 0 0 1 .014-.268l1.967-11.809z"/>
  </svg>
);

const DigitalOceanIcon = () => (
  <svg viewBox="0 0 24 24" className="w-8 h-8" fill="currentColor">
    <path d="M12.04 0C5.408-.02.005 5.37.005 11.992h4.638c0-4.923 4.882-8.731 10.064-6.855a6.95 6.95 0 014.147 4.148c1.889 5.177-1.924 10.055-6.84 10.064v-4.667H7.391v4.667H2.73v-4.667h4.661v-4.657h4.657v4.667h4.667c6.631 0 12.024-5.402 11.999-12.034A12.02 12.02 0 0012.04 0z"/>
  </svg>
);

// Provider configurations
const providerInfos: CloudProviderInfo[] = [
  {
    id: "aws",
    name: "Amazon Web Services",
    shortName: "AWS",
    icon: <AWSIcon />,
    status: "active",
    fields: [
      {
        name: "accessKeyId",
        label: "Access Key ID",
        type: "text",
        placeholder: "AKIAIOSFODNN7EXAMPLE",
        required: true,
      },
      {
        name: "secretAccessKey",
        label: "Secret Access Key",
        type: "password",
        placeholder: "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
        required: true,
      },
      {
        name: "region",
        label: "Default Region",
        type: "text",
        placeholder: "us-east-1",
        required: true,
      },
    ],
  },
  {
    id: "gcp",
    name: "Google Cloud Platform",
    icon: <GCPIcon />,
    status: "coming_soon",
    fields: [],
  },
  {
    id: "azure",
    name: "Microsoft Azure",
    icon: <AzureIcon />,
    status: "coming_soon",
    fields: [],
  },
  {
    id: "digitalocean",
    name: "DigitalOcean",
    icon: <DigitalOceanIcon />,
    status: "coming_soon",
    fields: [],
  },
];

interface CloudProviderProps {
  cloudProvider: {
    provider?: string;
    aws?: any;
    gcp?: any;
    azure?: any;
    digitalocean?: any;
  };
  onUpdate: (settings: any) => Promise<void>;
  disabled?: boolean;
}

export function CloudProvider({ cloudProvider, onUpdate, disabled }: CloudProviderProps) {
  const [isConfigureDialogOpen, setIsConfigureDialogOpen] = useState(false);
  const [isEditDialogOpen, setIsEditDialogOpen] = useState(false);
  const [selectedProvider, setSelectedProvider] = useState<CloudProviderInfo | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);

  const currentProvider = cloudProvider?.provider;
  const currentProviderInfo = providerInfos.find(p => p.id === currentProvider);

  // Dynamic schema based on selected provider
  const getProviderSchema = (provider: CloudProviderInfo) => {
    const schemaFields: Record<string, z.ZodString> = {};
    provider.fields.forEach((field) => {
      schemaFields[field.name] = field.required 
        ? z.string().min(1, `${field.label} is required`)
        : z.string().optional();
    });
    return z.object(schemaFields);
  };

  const form = useForm({
    resolver: selectedProvider ? zodResolver(getProviderSchema(selectedProvider)) : undefined,
    defaultValues: {},
  });

  const handleConfigureProvider = async (values: any) => {
    if (!selectedProvider) {return;}

    setIsSubmitting(true);
    try {
      const updatedSettings = {
        cloudProvider: {
          provider: selectedProvider.id,
          [selectedProvider.id]: values,
        },
      };

      await onUpdate(updatedSettings);
      toast.success(`${selectedProvider.name} configured successfully`);
      setIsConfigureDialogOpen(false);
      form.reset({});
      setSelectedProvider(null);
    } catch (error: any) {
      toast.error(error.message || "Failed to configure cloud provider");
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleUpdateProvider = async (values: any) => {
    if (!currentProvider) {return;}

    setIsSubmitting(true);
    try {
      const updatedSettings = {
        cloudProvider: {
          provider: currentProvider,
          [currentProvider]: {
            ...(cloudProvider[currentProvider as keyof typeof cloudProvider] || {}),
            ...values,
          },
        },
      };

      await onUpdate(updatedSettings);
      toast.success("Cloud provider updated successfully");
      setIsEditDialogOpen(false);
      form.reset({});
    } catch (error: any) {
      toast.error(error.message || "Failed to update cloud provider");
    } finally {
      setIsSubmitting(false);
    }
  };

  const openEditDialog = () => {
    if (!currentProviderInfo || !currentProvider) {return;}
    const providerData = cloudProvider[currentProvider as keyof typeof cloudProvider];
    const currentValues: Record<string, string> = {};
    
    // Initialize all fields with empty strings or existing values
    currentProviderInfo.fields.forEach(field => {
      currentValues[field.name] = (providerData as any)?.[field.name] || "";
    });
    
    form.reset(currentValues);
    setIsEditDialogOpen(true);
  };

  return (
    <div className="space-y-6">


      {/* Current Provider or No Provider */}
      {currentProvider ? (
        <div className="space-y-4">
          <Card className="border-primary/20 bg-primary/5">
            <CardHeader>
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-3">
                  <div className="text-muted-foreground">
                    {currentProviderInfo?.icon}
                  </div>
                  <div>
                    <CardTitle className="text-base">{currentProviderInfo?.name}</CardTitle>
                    <CardDescription>Active cloud provider</CardDescription>
                  </div>
                </div>
                <div className="flex items-center gap-2">
                  <Badge variant="outline" className="gap-1">
                    <CheckCircle2 className="h-3 w-3" />
                    Configured
                  </Badge>
                  {!disabled && (
                    <Button
                      size="sm"
                      variant="ghost"
                      onClick={openEditDialog}
                    >
                      <Settings className="h-4 w-4" />
                    </Button>
                  )}
                </div>
              </div>
            </CardHeader>
            <CardContent>
              <div className="flex items-center gap-2 text-sm text-muted-foreground">
                <Lock className="h-3 w-3" />
                Credentials configured (hidden for security)
              </div>
            </CardContent>
          </Card>
          
          <Alert variant="warning">
            <AlertDescription>
              This cloud provider cannot be changed. To use a different provider, you must delete all teams first.
            </AlertDescription>
          </Alert>
        </div>
      ) : (
        <div className="space-y-4">
          <Alert variant="info">
            <AlertDescription>
              No cloud provider configured. Configure a provider to enable infrastructure deployment.
            </AlertDescription>
          </Alert>

          <Card className="border-dashed">
            <CardContent className="pt-6">
              <div className="text-center">
                <Cloud className="mx-auto h-12 w-12 text-muted-foreground/50" />
                <h3 className="mt-2 text-sm font-medium">No provider configured</h3>
                <p className="mt-1 text-sm text-muted-foreground">
                  Get started by configuring your cloud provider.
                </p>
                <div className="mt-6">
                  <Button onClick={() => setIsConfigureDialogOpen(true)} disabled={disabled}>
                    <Plus className="mr-2 h-4 w-4" />
                    Configure Provider
                  </Button>
                </div>
              </div>
            </CardContent>
          </Card>
        </div>
      )}

      {/* Configure Dialog */}
      <Dialog open={isConfigureDialogOpen} onOpenChange={(open) => {
        setIsConfigureDialogOpen(open);
        if (!open) {
          setSelectedProvider(null);
          form.reset({});
        }
      }}>
        <DialogContent className="max-w-2xl">
          <DialogHeader>
            <DialogTitle>{selectedProvider ? `Configure ${selectedProvider.name}` : 'Select Cloud Provider'}</DialogTitle>
            <DialogDescription>
              {selectedProvider 
                ? 'Enter your cloud provider credentials. This cannot be changed later.'
                : 'Choose your cloud provider. This decision cannot be changed without deleting all teams.'
              }
            </DialogDescription>
          </DialogHeader>

          {!selectedProvider ? (
            <div className="grid gap-4 py-4 md:grid-cols-2">
              {providerInfos.map((provider) => {
                const isDisabled = provider.status === "coming_soon";

                return (
                  <Card
                    key={provider.id}
                    className={cn(
                      "relative cursor-pointer transition-all duration-200",
                      isDisabled
                        ? "cursor-not-allowed opacity-60"
                        : "hover:border-primary hover:shadow-md"
                    )}
                    onClick={() => {
                      if (!isDisabled) {
                        setSelectedProvider(provider);
                        // Reset with empty values for all fields
                        const emptyValues: Record<string, string> = {};
                        provider.fields.forEach(field => {
                          emptyValues[field.name] = "";
                        });
                        form.reset(emptyValues);
                      }
                    }}
                  >
                    <CardHeader className="space-y-1">
                      <div className="flex items-start justify-between">
                        <div className="space-y-1">
                          <CardTitle className="text-base font-medium">
                            {provider.name}
                          </CardTitle>
                          {provider.status === "coming_soon" && (
                            <CardDescription className="text-sm">
                              Coming Soon
                            </CardDescription>
                          )}
                        </div>
                        <div className="text-muted-foreground">
                          {provider.icon}
                        </div>
                      </div>
                    </CardHeader>
                    {provider.status === "coming_soon" && (
                      <div className="absolute inset-0 bg-background/50 backdrop-blur-[1px] flex items-center justify-center">
                        <Badge variant="secondary" className="gap-1.5">
                          <AlertCircle className="h-3.5 w-3.5" />
                          Coming Soon
                        </Badge>
                      </div>
                    )}
                  </Card>
                );
              })}
            </div>
          ) : (
            <Form {...form}>
              <form onSubmit={form.handleSubmit(handleConfigureProvider)} className="space-y-4">
                <Alert variant="warning">
                  <AlertDescription>
                    This is a one-time configuration. Make sure you have the correct credentials.
                  </AlertDescription>
                </Alert>

                {selectedProvider.fields.map((field) => (
                  <FormField
                    key={field.name}
                    control={form.control}
                    name={field.name}
                    render={({ field: formField }) => (
                      <FormItem>
                        <FormLabel>{field.label}</FormLabel>
                        <FormControl>
                          <Input
                            type={field.type}
                            placeholder={field.placeholder}
                            {...formField}
                            value={formField.value || ""}
                          />
                        </FormControl>
                        <FormMessage />
                      </FormItem>
                    )}
                  />
                ))}

                <DialogFooter>
                  <Button
                    type="button"
                    variant="outline"
                    onClick={() => {
                      setSelectedProvider(null);
                      form.reset({});
                    }}
                  >
                    Back
                  </Button>
                  <Button type="submit" disabled={isSubmitting}>
                    {isSubmitting ? "Configuring..." : "Configure Provider"}
                  </Button>
                </DialogFooter>
              </form>
            </Form>
          )}
        </DialogContent>
      </Dialog>

      {/* Edit Dialog */}
      <Dialog open={isEditDialogOpen} onOpenChange={setIsEditDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Update {currentProviderInfo?.name}</DialogTitle>
            <DialogDescription>
              Update your cloud provider credentials. Leave fields empty to keep existing values.
            </DialogDescription>
          </DialogHeader>

          {currentProviderInfo && (
            <Form {...form}>
              <form
                onSubmit={form.handleSubmit(handleUpdateProvider)}
                className="space-y-4"
              >
                {currentProviderInfo.fields.map((field) => (
                  <FormField
                    key={field.name}
                    control={form.control}
                    name={field.name}
                    render={({ field: formField }) => (
                      <FormItem>
                        <FormLabel>{field.label}</FormLabel>
                        <FormControl>
                          <Input
                            type={field.type}
                            placeholder={field.placeholder}
                            {...formField}
                            value={formField.value || ""}
                          />
                        </FormControl>
                        <FormDescription>
                          {field.type === "password" &&
                            "Leave empty to keep the existing value"}
                        </FormDescription>
                        <FormMessage />
                      </FormItem>
                    )}
                  />
                ))}

                <DialogFooter>
                  <Button
                    type="button"
                    variant="outline"
                    onClick={() => {
                      setIsEditDialogOpen(false);
                      form.reset({});
                    }}
                  >
                    Cancel
                  </Button>
                  <Button type="submit" disabled={isSubmitting}>
                    {isSubmitting ? "Updating..." : "Update"}
                  </Button>
                </DialogFooter>
              </form>
            </Form>
          )}
        </DialogContent>
      </Dialog>
    </div>
  );
}