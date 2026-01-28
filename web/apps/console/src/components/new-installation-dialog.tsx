'use client'

import { useState, useEffect, useCallback } from 'react'
import { useRouter } from 'next/navigation'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import * as z from 'zod'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@kloudlite/ui'
import {
  Button,
  Input,
  Textarea,
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
  FormDescription,
  Tabs,
  TabsContent,
  TabsList,
  TabsTrigger,
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@kloudlite/ui'
import { Loader2, CheckCircle2, AlertCircle, Copy, ExternalLink, PartyPopper, Clock } from 'lucide-react'
import { toast } from 'sonner'
import { InstallationProgress } from './installation-progress'
import { WorldMap } from './world-map'

const installationSchema = z.object({
  name: z
    .string()
    .min(3, 'Name must be at least 3 characters')
    .max(50, 'Name must be less than 50 characters')
    .regex(/^[a-zA-Z0-9\s-]+$/, 'Name can only contain letters, numbers, spaces, and hyphens'),
  description: z.string().max(200, 'Description must be less than 200 characters').optional(),
  subdomain: z
    .string()
    .min(3, 'Subdomain must be at least 3 characters')
    .max(63, 'Subdomain must be less than 63 characters')
    .regex(
      /^[a-z0-9]([a-z0-9-]*[a-z0-9])?$/,
      'Subdomain must start and end with alphanumeric characters and can only contain lowercase letters, numbers, and hyphens',
    ),
})

type InstallationFormData = z.infer<typeof installationSchema>

interface NewInstallationDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
}

// AWS Regions
const AWS_REGIONS = [
  { value: '', label: 'Use AWS CLI default region' },
  { value: 'us-east-1', label: 'US East (N. Virginia)' },
  { value: 'us-east-2', label: 'US East (Ohio)' },
  { value: 'us-west-1', label: 'US West (N. California)' },
  { value: 'us-west-2', label: 'US West (Oregon)' },
  { value: 'eu-west-1', label: 'EU (Ireland)' },
  { value: 'eu-west-2', label: 'EU (London)' },
  { value: 'eu-west-3', label: 'EU (Paris)' },
  { value: 'eu-central-1', label: 'EU (Frankfurt)' },
  { value: 'eu-north-1', label: 'EU (Stockholm)' },
  { value: 'ap-south-1', label: 'Asia Pacific (Mumbai)' },
  { value: 'ap-southeast-1', label: 'Asia Pacific (Singapore)' },
  { value: 'ap-southeast-2', label: 'Asia Pacific (Sydney)' },
  { value: 'ap-northeast-1', label: 'Asia Pacific (Tokyo)' },
  { value: 'ap-northeast-2', label: 'Asia Pacific (Seoul)' },
  { value: 'sa-east-1', label: 'South America (São Paulo)' },
  { value: 'ca-central-1', label: 'Canada (Central)' },
]

// GCP Regions
const GCP_REGIONS = [
  { value: 'us-central1', label: 'US Central (Iowa)' },
  { value: 'us-east1', label: 'US East (South Carolina)' },
  { value: 'us-east4', label: 'US East (N. Virginia)' },
  { value: 'us-west1', label: 'US West (Oregon)' },
  { value: 'us-west2', label: 'US West (Los Angeles)' },
  { value: 'us-west3', label: 'US West (Salt Lake City)' },
  { value: 'us-west4', label: 'US West (Las Vegas)' },
  { value: 'europe-west1', label: 'Europe West (Belgium)' },
  { value: 'europe-west2', label: 'Europe West (London)' },
  { value: 'europe-west3', label: 'Europe West (Frankfurt)' },
  { value: 'europe-west4', label: 'Europe West (Netherlands)' },
  { value: 'europe-north1', label: 'Europe North (Finland)' },
  { value: 'asia-east1', label: 'Asia East (Taiwan)' },
  { value: 'asia-east2', label: 'Asia East (Hong Kong)' },
  { value: 'asia-southeast1', label: 'Asia Southeast (Singapore)' },
  { value: 'asia-southeast2', label: 'Asia Southeast (Jakarta)' },
  { value: 'asia-south1', label: 'Asia South (Mumbai)' },
  { value: 'asia-northeast1', label: 'Asia Northeast (Tokyo)' },
  { value: 'asia-northeast2', label: 'Asia Northeast (Osaka)' },
  { value: 'asia-northeast3', label: 'Asia Northeast (Seoul)' },
  { value: 'australia-southeast1', label: 'Australia (Sydney)' },
  { value: 'southamerica-east1', label: 'South America (São Paulo)' },
]

// Azure Locations
const AZURE_LOCATIONS = [
  { value: 'eastus', label: 'East US (Virginia)' },
  { value: 'eastus2', label: 'East US 2 (Virginia)' },
  { value: 'westus', label: 'West US (California)' },
  { value: 'westus2', label: 'West US 2 (Washington)' },
  { value: 'westus3', label: 'West US 3 (Arizona)' },
  { value: 'centralus', label: 'Central US (Iowa)' },
  { value: 'northcentralus', label: 'North Central US (Illinois)' },
  { value: 'southcentralus', label: 'South Central US (Texas)' },
  { value: 'westeurope', label: 'West Europe (Netherlands)' },
  { value: 'northeurope', label: 'North Europe (Ireland)' },
  { value: 'uksouth', label: 'UK South (London)' },
  { value: 'ukwest', label: 'UK West (Cardiff)' },
  { value: 'francecentral', label: 'France Central (Paris)' },
  { value: 'germanywestcentral', label: 'Germany West Central (Frankfurt)' },
  { value: 'swedencentral', label: 'Sweden Central (Gävle)' },
  { value: 'southeastasia', label: 'Southeast Asia (Singapore)' },
  { value: 'eastasia', label: 'East Asia (Hong Kong)' },
  { value: 'japaneast', label: 'Japan East (Tokyo)' },
  { value: 'japanwest', label: 'Japan West (Osaka)' },
  { value: 'koreacentral', label: 'Korea Central (Seoul)' },
  { value: 'australiaeast', label: 'Australia East (Sydney)' },
  { value: 'australiasoutheast', label: 'Australia Southeast (Melbourne)' },
  { value: 'centralindia', label: 'Central India (Pune)' },
  { value: 'southindia', label: 'South India (Chennai)' },
  { value: 'brazilsouth', label: 'Brazil South (São Paulo)' },
  { value: 'canadacentral', label: 'Canada Central (Toronto)' },
  { value: 'canadaeast', label: 'Canada East (Quebec)' },
]

const getCloudProviderCommands = (installationKey: string, awsRegion: string, gcpRegion: string, azureLocation: string) => {
  const awsRegionFlag = awsRegion ? ` --region ${awsRegion}` : ''
  const gcpRegionFlag = gcpRegion ? ` --region ${gcpRegion}` : ''
  const azureLocationFlag = azureLocation ? ` --location ${azureLocation}` : ''
  return {
    aws: {
      name: 'AWS',
      commands: [`curl -fsSL https://get.khost.dev/install/aws | bash -s -- --key ${installationKey}${awsRegionFlag}`],
      requirements: [
        'AWS CLI configured',
        'IAM user with EC2 full access and iam:PassRole permissions',
        'Valid AWS access keys configured',
      ],
    },
    gcp: {
      name: 'Google Cloud',
      commands: [`curl -fsSL https://get.khost.dev/install/gcp | bash -s -- --key ${installationKey}${gcpRegionFlag}`],
      requirements: [
        'gcloud CLI configured with Application Default Credentials',
        'IAM permissions: Compute Admin, Service Account Admin, Storage Admin',
        'Billing enabled on the GCP project',
      ],
    },
    azure: {
      name: 'Azure',
      commands: [`curl -fsSL https://get.khost.dev/install/azure | bash -s -- --key ${installationKey}${azureLocationFlag}`],
      requirements: [
        'Azure CLI configured and logged in (az login)',
        'Subscription with VM, Network, and Storage permissions',
        'Billing enabled on the Azure subscription',
      ],
    },
  }
}

type VerificationStatus = 'waiting' | 'verified' | 'dns_pending' | 'complete' | 'error'
type ActiveStatus = 'checking' | 'active' | 'waiting' | 'error'

export function NewInstallationDialog({ open, onOpenChange }: NewInstallationDialogProps) {
  const router = useRouter()
  const [currentStep, setCurrentStep] = useState(1)
  const [creating, setCreating] = useState(false)
  const [checkingSubdomain, setCheckingSubdomain] = useState(false)
  const [subdomainAvailable, setSubdomainAvailable] = useState<boolean | null>(null)

  // Step 2 state
  const [installationKey, setInstallationKey] = useState<string>('')
  const [selectedProvider, setSelectedProvider] = useState('aws')
  const [awsRegion, setAwsRegion] = useState('')
  const [gcpRegion, setGcpRegion] = useState('us-central1')
  const [azureLocation, setAzureLocation] = useState('eastus')
  const [verificationStatus, setVerificationStatus] = useState<VerificationStatus>('waiting')

  // Step 3 state
  const [installationData, setInstallationData] = useState<{subdomain: string; url: string; installationId: string} | null>(null)
  const [activeStatus, setActiveStatus] = useState<ActiveStatus>('checking')
  const [checkCount, setCheckCount] = useState(0)

  const form = useForm<InstallationFormData>({
    resolver: zodResolver(installationSchema),
    defaultValues: {
      name: '',
      description: '',
      subdomain: '',
    },
  })

  const checkSubdomainAvailability = async (subdomain: string) => {
    if (!subdomain || subdomain.length < 3) {
      setSubdomainAvailable(null)
      return
    }

    const subdomainRegex = /^[a-z0-9]([a-z0-9-]*[a-z0-9])?$/
    if (!subdomainRegex.test(subdomain)) {
      setSubdomainAvailable(null)
      return
    }

    setCheckingSubdomain(true)
    try {
      const response = await fetch(`/api/installations/check-domain-kli?subdomain=${subdomain}`)
      const data = await response.json()
      setSubdomainAvailable(data.available)
    } catch (err) {
      console.error('Error checking subdomain:', err)
      setSubdomainAvailable(false)
    } finally {
      setCheckingSubdomain(false)
    }
  }

  const onSubmit = async (data: InstallationFormData) => {
    if (subdomainAvailable !== true) {
      toast.error('Please choose an available subdomain')
      return
    }

    setCreating(true)

    try {
      const response = await fetch('/api/installations/create-installation', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          name: data.name,
          description: data.description || undefined,
          subdomain: data.subdomain,
        }),
      })

      if (!response.ok) {
        const errorData = await response.json()
        throw new Error(errorData.error || 'Failed to create installation')
      }

      const result = await response.json()
      toast.success('Installation created successfully!')

      // Fetch session to get installation key
      const sessionResponse = await fetch('/api/installations/session')
      if (sessionResponse.ok) {
        const sessionData = await sessionResponse.json()
        if (sessionData.installationKey) {
          setInstallationKey(sessionData.installationKey)
          setCurrentStep(2)
        }
      }
    } catch (err) {
      const error = err instanceof Error ? err : new Error('Failed to create installation')
      toast.error(error.message)
    } finally {
      setCreating(false)
    }
  }

  // Poll for deployment verification (Step 2)
  useEffect(() => {
    if (currentStep !== 2 || !installationKey) return

    const checkVerification = async () => {
      try {
        const response = await fetch('/api/installations/check-installation', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ installationKey }),
        })

        const data = await response.json()

        if (data.verified && data.dnsConfigured) {
          setVerificationStatus('complete')
          // Move to step 3 after 1 second
          setTimeout(() => {
            // Fetch installation data for step 3
            fetchInstallationDataForComplete()
          }, 1000)
        } else if (data.verified) {
          setVerificationStatus('dns_pending')
        }
      } catch (error) {
        console.error('Error checking verification:', error)
      }
    }

    checkVerification()
    const interval = setInterval(checkVerification, 5000)

    return () => clearInterval(interval)
  }, [currentStep, installationKey])

  // Fetch installation data for step 3
  const fetchInstallationDataForComplete = async () => {
    try {
      const response = await fetch('/api/installations/session')
      if (response.ok) {
        const sessionData = await response.json()

        if (sessionData.installationKey) {
          const verifyResponse = await fetch('/api/installations/verify-key', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ installationKey: sessionData.installationKey }),
          })
          if (verifyResponse.ok) {
            const verifyData = await verifyResponse.json()
            const isValidSubdomain = verifyData.subdomain &&
              verifyData.subdomain !== '0.0.0.0' &&
              !verifyData.subdomain.includes('0.0.0.0')

            if (isValidSubdomain) {
              const domain = process.env.NEXT_PUBLIC_INSTALLATION_DOMAIN || 'khost.dev'
              setInstallationData({
                subdomain: verifyData.subdomain,
                url: `https://${verifyData.subdomain}.${domain}`,
                installationId: verifyData.installationId,
              })
              setCurrentStep(3)

              if (verifyData.installationId) {
                checkActiveStatus(verifyData.installationId)
              }
            }
          }
        }
      }
    } catch (error) {
      console.error('Error fetching installation data:', error)
    }
  }

  // Check if installation is active (Step 3)
  const checkActiveStatus = useCallback(async (installationId: string) => {
    try {
      const response = await fetch(`/api/installations/${installationId}/ping`)
      if (response.ok) {
        const data = await response.json()
        if (data.active) {
          setActiveStatus('active')
          return true
        } else {
          setActiveStatus('waiting')
          return false
        }
      }
      setActiveStatus('waiting')
      return false
    } catch {
      setActiveStatus('waiting')
      return false
    }
  }, [])

  // Poll for active status (Step 3)
  useEffect(() => {
    if (currentStep !== 3 || !installationData?.installationId || activeStatus === 'active') {
      return
    }

    const intervalId = setInterval(async () => {
      setCheckCount(c => c + 1)
      const isActive = await checkActiveStatus(installationData.installationId)
      if (isActive) {
        clearInterval(intervalId)
      }
    }, 5000)

    return () => clearInterval(intervalId)
  }, [currentStep, installationData?.installationId, activeStatus, checkActiveStatus])

  const copyCommand = (command: string) => {
    navigator.clipboard.writeText(command)
    toast.success('Command copied to clipboard')
  }

  const copyToClipboard = (text: string, label: string) => {
    navigator.clipboard.writeText(text)
    toast.success(`${label} copied to clipboard`)
  }

  const handleClose = () => {
    // Reset all state
    setCurrentStep(1)
    setCreating(false)
    setCheckingSubdomain(false)
    setSubdomainAvailable(null)
    setInstallationKey('')
    setSelectedProvider('aws')
    setAwsRegion('')
    setGcpRegion('us-central1')
    setAzureLocation('eastus')
    setVerificationStatus('waiting')
    setInstallationData(null)
    setActiveStatus('checking')
    setCheckCount(0)
    form.reset()
    onOpenChange(false)

    // Refresh the installations list
    router.refresh()
  }

  const CLOUD_PROVIDERS = installationKey ? getCloudProviderCommands(installationKey, awsRegion, gcpRegion, azureLocation) : null

  // Render Step 1: Create Installation Form
  const renderStep1 = () => (
    <>
      <div className="border-b border-foreground/10">
        <DialogHeader className="px-12 pt-10 pb-6">
          <div className="flex flex-col items-center gap-6">
            <InstallationProgress currentStep={1} />
            <div className="text-center">
              <DialogTitle className="text-2xl font-bold tracking-tight">Create New Installation</DialogTitle>
              <DialogDescription className="text-base mt-2">
                Deploy Kloudlite to your cloud infrastructure
              </DialogDescription>
            </div>
          </div>
        </DialogHeader>
      </div>

      <div className="px-12 pb-10 pt-8">
        <div className="max-w-2xl mx-auto">
          <Form {...form}>
            <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-10">
              <div className="space-y-8">
                <FormField
                  control={form.control}
                  name="name"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel className="text-base font-semibold text-foreground">Installation Name</FormLabel>
                      <FormControl>
                        <Input
                          placeholder="Production"
                          {...field}
                          disabled={creating}
                          className="h-11 w-full text-base"
                        />
                      </FormControl>
                      <FormDescription className="text-sm">
                        Choose a descriptive name for this installation
                      </FormDescription>
                      <FormMessage />
                    </FormItem>
                  )}
                />

                <FormField
                  control={form.control}
                  name="description"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel className="text-base font-semibold text-foreground">
                        Description <span className="text-muted-foreground font-normal text-sm">(optional)</span>
                      </FormLabel>
                      <FormControl>
                        <Textarea
                          placeholder="Production environment for customer workloads"
                          {...field}
                          disabled={creating}
                          rows={3}
                          className="resize-none w-full text-base"
                        />
                      </FormControl>
                      <FormDescription className="text-sm">
                        Add context about the purpose of this installation
                      </FormDescription>
                      <FormMessage />
                    </FormItem>
                  )}
                />

                <FormField
                  control={form.control}
                  name="subdomain"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel className="text-base font-semibold text-foreground">Domain</FormLabel>
                      <FormControl>
                        <div className="relative w-full">
                          <Input
                            placeholder="your-company"
                            {...field}
                            disabled={creating}
                            className="h-11 font-mono pr-10 text-base w-full"
                            onChange={(e) => {
                              const value = e.target.value.toLowerCase()
                              field.onChange(value)
                              checkSubdomainAvailability(value)
                            }}
                          />
                          {checkingSubdomain && (
                            <div className="absolute top-1/2 right-3 -translate-y-1/2">
                              <Loader2 className="text-muted-foreground size-4 animate-spin" />
                            </div>
                          )}
                          {!checkingSubdomain && subdomainAvailable === true && (
                            <div className="absolute top-1/2 right-3 -translate-y-1/2">
                              <CheckCircle2 className="size-4 text-green-600" />
                            </div>
                          )}
                          {!checkingSubdomain && subdomainAvailable === false && (
                            <div className="absolute top-1/2 right-3 -translate-y-1/2">
                              <AlertCircle className="text-destructive size-4" />
                            </div>
                          )}
                        </div>
                      </FormControl>
                      <FormDescription className="text-sm">
                        <span className="font-mono">
                          {field.value || 'your-subdomain'}.{process.env.NEXT_PUBLIC_INSTALLATION_DOMAIN || 'khost.dev'}
                        </span>
                      </FormDescription>
                      <div className="min-h-[20px]">
                        {!checkingSubdomain && subdomainAvailable === false && (
                          <p className="text-sm font-medium text-destructive">
                            This domain is already taken
                          </p>
                        )}
                        {!checkingSubdomain && subdomainAvailable === true && (
                          <p className="text-sm font-medium text-green-600 dark:text-green-500">
                            Domain is available
                          </p>
                        )}
                      </div>
                      <FormMessage />
                    </FormItem>
                  )}
                />
              </div>
            </form>
          </Form>
        </div>
      </div>

      <div className="border-t border-foreground/10">
        <div className="px-12 py-6">
          <div className="max-w-2xl mx-auto flex justify-end gap-3">
            <Button
              type="button"
              variant="outline"
              onClick={handleClose}
              disabled={creating}
              size="lg"
              className="min-w-32"
            >
              Cancel
            </Button>
            <Button
              type="submit"
              disabled={creating || subdomainAvailable !== true}
              size="lg"
              className="min-w-48"
              onClick={form.handleSubmit(onSubmit)}
            >
              {creating ? (
                <>
                  <Loader2 className="mr-2 size-4 animate-spin" />
                  Creating Installation...
                </>
              ) : (
                'Create Installation'
              )}
            </Button>
          </div>
        </div>
      </div>
    </>
  )

  // Render Step 2: Install in Cloud
  const renderStep2 = () => {
    if (!CLOUD_PROVIDERS) return null

    return (
      <>
        <div className="border-b border-foreground/10">
          <DialogHeader className="px-12 pt-10 pb-6">
            <div className="flex flex-col items-center gap-6">
              <InstallationProgress currentStep={2} />
              <div className="text-center">
                <DialogTitle className="text-2xl font-bold tracking-tight">Deploy to Cloud</DialogTitle>
                <DialogDescription className="text-base mt-2">
                  Run the installation command in your cloud provider terminal
                </DialogDescription>
              </div>
            </div>
          </DialogHeader>
        </div>

        <div className="px-12 pb-10 pt-8 max-h-[65vh] overflow-y-auto">
          <div className="max-w-3xl mx-auto">
          <div className="space-y-8">
            <div>
              <h3 className="text-base font-semibold text-foreground mb-4">Select Cloud Provider</h3>
              <Tabs value={selectedProvider} onValueChange={setSelectedProvider}>
                <TabsList className="grid w-full max-w-md grid-cols-3 p-1 rounded-sm">
                  <TabsTrigger value="aws" className="text-sm font-medium rounded-sm">
                    AWS
                  </TabsTrigger>
                  <TabsTrigger value="gcp" className="text-sm font-medium rounded-sm">
                    GCP
                  </TabsTrigger>
                  <TabsTrigger value="azure" className="text-sm font-medium rounded-sm">
                    Azure
                  </TabsTrigger>
                </TabsList>

                {Object.entries(CLOUD_PROVIDERS).map(([key, config]) => (
                  <TabsContent key={key} value={key} className="mt-8 space-y-8">
                    <div className="space-y-8">
                      {/* AWS Region Selector */}
                      {key === 'aws' && (
                        <div>
                          <p className="text-foreground mb-3 text-base font-semibold">Region</p>
                          <Select
                            value={awsRegion || 'default'}
                            onValueChange={(val) => setAwsRegion(val === 'default' ? '' : val)}
                          >
                            <SelectTrigger className="max-w-md rounded-sm border-foreground/10 h-11">
                              <SelectValue placeholder="Select a region" />
                            </SelectTrigger>
                            <SelectContent>
                              {AWS_REGIONS.map((region) => (
                                <SelectItem key={region.value || 'default'} value={region.value || 'default'}>
                                  {region.label}
                                </SelectItem>
                              ))}
                            </SelectContent>
                          </Select>
                        </div>
                      )}

                      {/* GCP Region Selector */}
                      {key === 'gcp' && (
                        <div>
                          <p className="text-foreground mb-3 text-base font-semibold">Region</p>
                          <Select value={gcpRegion} onValueChange={setGcpRegion}>
                            <SelectTrigger className="max-w-md rounded-sm border-foreground/10 h-11">
                              <SelectValue placeholder="Select a region" />
                            </SelectTrigger>
                            <SelectContent>
                              {GCP_REGIONS.map((region) => (
                                <SelectItem key={region.value} value={region.value}>
                                  {region.label}
                                </SelectItem>
                              ))}
                            </SelectContent>
                          </Select>
                        </div>
                      )}

                      {/* Azure Location Selector */}
                      {key === 'azure' && (
                        <div>
                          <p className="text-foreground mb-3 text-base font-semibold">Location</p>
                          <Select value={azureLocation} onValueChange={setAzureLocation}>
                            <SelectTrigger className="max-w-md rounded-sm border-foreground/10 h-11">
                              <SelectValue placeholder="Select a location" />
                            </SelectTrigger>
                            <SelectContent>
                              {AZURE_LOCATIONS.map((location) => (
                                <SelectItem key={location.value} value={location.value}>
                                  {location.label}
                                </SelectItem>
                              ))}
                            </SelectContent>
                          </Select>
                        </div>
                      )}

                      {/* World Map */}
                      <div>
                        <WorldMap
                          selectedRegion={key === 'aws' ? awsRegion : key === 'gcp' ? gcpRegion : azureLocation}
                          provider={key as 'aws' | 'gcp' | 'azure'}
                        />
                      </div>

                      <div className="border-t border-foreground/10 pt-8">
                        <h4 className="text-foreground mb-4 text-base font-semibold">Prerequisites</h4>
                        <ul className="text-muted-foreground space-y-3 text-base leading-relaxed">
                          {config.requirements.map((req, idx) => (
                            <li key={idx} className="flex items-start gap-3">
                              <div className="bg-primary mt-2 size-1.5 flex-shrink-0 rounded-full" />
                              <span>{req}</span>
                            </li>
                          ))}
                        </ul>
                      </div>

                      <div className="border-t border-foreground/10 pt-8">
                        <h4 className="text-foreground mb-4 text-base font-semibold">Installation Command</h4>
                        <div className="space-y-3">
                          {config.commands.map((cmd, idx) => (
                            <div key={idx} className="bg-muted/50 border border-foreground/10 p-5 rounded-sm">
                              <div className="flex items-start justify-between gap-4">
                                <code className="flex-1 font-mono text-sm leading-relaxed break-all text-foreground">
                                  {cmd}
                                </code>
                                <Button
                                  variant="outline"
                                  size="default"
                                  className="flex-shrink-0"
                                  onClick={() => copyCommand(cmd)}
                                >
                                  <Copy className="mr-2 size-4" />
                                  Copy
                                </Button>
                              </div>
                            </div>
                          ))}
                        </div>
                      </div>
                    </div>
                  </TabsContent>
                ))}
              </Tabs>
            </div>

            {/* Verification Status */}
            <div className="border-t border-foreground/10 pt-8">
              <div className="flex items-center justify-center gap-3 text-base py-4 bg-muted/30 rounded-sm">
                {verificationStatus === 'waiting' && (
                  <>
                    <Loader2 className="size-5 animate-spin text-blue-600" />
                    <span className="text-muted-foreground font-medium">Waiting for deployment...</span>
                  </>
                )}
                {verificationStatus === 'dns_pending' && (
                  <>
                    <Loader2 className="size-5 animate-spin text-yellow-600" />
                    <span className="text-yellow-600 font-medium">Deployment verified. Configuring DNS...</span>
                  </>
                )}
                {verificationStatus === 'complete' && (
                  <>
                    <CheckCircle2 className="size-5 text-green-600" />
                    <span className="text-green-600 font-semibold">Installation complete! Loading...</span>
                  </>
                )}
              </div>
            </div>
          </div>
        </div>
        </div>
      </>
    )
  }

  // Render Step 3: Complete
  const renderStep3 = () => {
    if (!installationData) return null

    const domain = process.env.NEXT_PUBLIC_INSTALLATION_DOMAIN || 'khost.dev'

    if (activeStatus === 'active') {
      // Active state
      return (
        <>
          <div className="border-b border-foreground/10">
            <DialogHeader className="px-12 pt-10 pb-8">
              <div className="flex flex-col items-center gap-6">
                <InstallationProgress currentStep={3} />
                <div className="flex flex-col items-center text-center">
                  <div className="mb-6 flex size-24 items-center justify-center bg-green-100 dark:bg-green-950 rounded-full">
                    <PartyPopper className="size-12 text-green-600 dark:text-green-400" />
                  </div>
                  <DialogTitle className="text-2xl font-bold tracking-tight">Installation Complete!</DialogTitle>
                  <DialogDescription className="text-base mt-2">
                    Your Kloudlite installation is ready to use
                  </DialogDescription>
                </div>
              </div>
            </DialogHeader>
          </div>

          <div className="px-12 pb-10 pt-8">
            <div className="max-w-3xl mx-auto">
              <div className="space-y-8">
              <div>
                <h3 className="text-lg font-semibold text-foreground mb-2">Access Your Installation</h3>
                <p className="text-muted-foreground text-base">
                  Your installation is live and ready to use
                </p>
              </div>

              <div className="bg-muted/30 border border-foreground/10 p-8 rounded-sm">
                <p className="mb-4 text-sm font-semibold text-muted-foreground uppercase tracking-wide">Installation URL</p>
                <div className="flex items-center gap-4 mb-6">
                  <a
                    href={installationData.url}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="text-primary flex items-center gap-2 font-mono text-lg hover:underline flex-1"
                  >
                    {installationData.subdomain}.{domain}
                    <ExternalLink className="size-5" />
                  </a>
                  <Button
                    variant="outline"
                    size="default"
                    onClick={() => copyToClipboard(installationData.url, 'URL')}
                  >
                    <Copy className="mr-2 size-4" />
                    Copy URL
                  </Button>
                </div>

                <div className="flex gap-4">
                  <Button
                    size="lg"
                    className="flex-1 max-w-xs"
                    onClick={() => window.open(installationData.url, '_blank')}
                  >
                    <ExternalLink className="mr-2 size-4" />
                    Open Installation Dashboard
                  </Button>
                </div>
              </div>

              <div className="border-t border-foreground/10 pt-8">
                <h4 className="text-base font-semibold text-foreground mb-3">What's Next?</h4>
                <p className="text-muted-foreground text-base leading-relaxed">
                  Access your installation dashboard to create and manage workspaces, environments, and work machines.
                  Your team members can log in using their own credentials at the installation URL.
                </p>
              </div>

              <div className="flex justify-end gap-3 pt-4">
                <Button
                  variant="outline"
                  size="lg"
                  className="min-w-32"
                  onClick={handleClose}
                >
                  Close
                </Button>
              </div>
              </div>
            </div>
          </div>
        </>
      )
    }

    // Waiting state
    return (
      <>
        <div className="border-b border-foreground/10">
          <DialogHeader className="px-12 pt-10 pb-8">
            <div className="flex flex-col items-center gap-6">
              <InstallationProgress currentStep={3} />
              <div className="flex flex-col items-center text-center">
                <div className="mb-6 flex size-24 items-center justify-center bg-blue-100 dark:bg-blue-950 rounded-full">
                  <Clock className="size-12 text-blue-600 dark:text-blue-400 animate-pulse" />
                </div>
                <DialogTitle className="text-2xl font-bold tracking-tight">Finalizing Installation</DialogTitle>
                <DialogDescription className="text-base mt-2">
                  Please wait while your installation becomes active
                </DialogDescription>
              </div>
            </div>
          </DialogHeader>
        </div>

        <div className="px-12 pb-10 pt-8">
          <div className="max-w-3xl mx-auto">
            <div className="space-y-8">
            <div className="bg-muted/30 border border-foreground/10 p-8 rounded-sm">
              <div className="flex items-center gap-3 mb-6">
                <Loader2 className="size-5 animate-spin text-primary" />
                <h3 className="text-lg font-semibold text-foreground">Activating Installation</h3>
              </div>
              <p className="text-muted-foreground text-base mb-8">
                Your installation is being set up. This process typically takes 1-3 minutes.
                The page will automatically update once your installation is ready.
              </p>

              <div className="bg-background border border-foreground/10 p-6 rounded-sm">
                <p className="mb-4 text-sm font-semibold text-muted-foreground uppercase tracking-wide">Installation URL</p>
                <div className="flex items-center justify-between gap-4">
                  <span className="text-foreground font-mono text-lg">
                    {installationData.subdomain}.{domain}
                  </span>
                  <Button
                    variant="outline"
                    size="default"
                    onClick={() => copyToClipboard(installationData.url, 'URL')}
                  >
                    <Copy className="mr-2 size-4" />
                    Copy URL
                  </Button>
                </div>
              </div>
            </div>

            <div className="bg-blue-50 dark:bg-blue-950 border border-blue-200 dark:border-blue-800 p-6 rounded-sm">
              <div className="flex items-start gap-4">
                <AlertCircle className="size-6 text-blue-600 dark:text-blue-400 mt-0.5 flex-shrink-0" />
                <div>
                  <p className="text-base font-semibold text-blue-900 dark:text-blue-200 mb-2">
                    Verification in Progress
                  </p>
                  <p className="text-sm text-blue-800 dark:text-blue-300 leading-relaxed">
                    We've checked {checkCount} time{checkCount !== 1 ? 's' : ''} so far.
                    The page will automatically refresh when your installation is active.
                  </p>
                </div>
              </div>
            </div>

            <div className="flex justify-end gap-3 pt-4">
              <Button
                variant="outline"
                size="lg"
                className="min-w-32"
                onClick={handleClose}
              >
                Close
              </Button>
            </div>
            </div>
          </div>
        </div>
      </>
    )
  }

  return (
    <Dialog open={open} onOpenChange={(isOpen) => !isOpen && handleClose()}>
      <DialogContent className={currentStep === 1 ? 'sm:max-w-[800px] p-0' : currentStep === 2 ? 'sm:max-w-[1000px] p-0' : 'sm:max-w-[900px] p-0'}>
        {currentStep === 1 && renderStep1()}
        {currentStep === 2 && renderStep2()}
        {currentStep === 3 && renderStep3()}
      </DialogContent>
    </Dialog>
  )
}
