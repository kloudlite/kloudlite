'use client'

import { useState, useEffect } from 'react'
import { useRouter } from 'next/navigation'
import {
  Button,
  Tabs,
  TabsContent,
  TabsList,
  TabsTrigger,
  Popover,
  PopoverContent,
  PopoverTrigger,
  Command,
  CommandInput,
  CommandList,
  CommandEmpty,
  CommandGroup,
  CommandItem
} from '@kloudlite/ui'
import { Loader2, Copy, CheckCircle2, ChevronsUpDown, Check } from 'lucide-react'
import { toast } from 'sonner'
import { cn } from '@kloudlite/lib'

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

interface InstallCommandsProps {
  installationKey: string
  installationId: string | null
}

export function InstallCommands({ installationKey, installationId: initialInstallationId }: InstallCommandsProps) {
  const router = useRouter()
  const [selectedProvider, setSelectedProvider] = useState('aws')
  const [awsRegion, setAwsRegion] = useState('')
  const [gcpRegion, setGcpRegion] = useState('us-central1')
  const [azureLocation, setAzureLocation] = useState('eastus')
  const [awsOpen, setAwsOpen] = useState(false)
  const [gcpOpen, setGcpOpen] = useState(false)
  const [azureOpen, setAzureOpen] = useState(false)
  const [verificationStatus, setVerificationStatus] = useState<'waiting' | 'verified' | 'dns_pending' | 'complete' | 'error'>(
    'waiting',
  )
  const [installationId] = useState<string | null>(initialInstallationId)
  const [currentStep, setCurrentStep] = useState(0)
  const [totalSteps, setTotalSteps] = useState(0)
  const [stepDescription, setStepDescription] = useState('')
  const [jobActive, setJobActive] = useState(false)

  // Poll job-status for progress when we have an installation ID
  useEffect(() => {
    if (!installationId || verificationStatus === 'complete') return

    const pollProgress = async () => {
      try {
        const response = await fetch(`/api/installations/${installationId}/job-status`)
        if (!response.ok) return

        const data = await response.json()

        if (data.currentStep != null) setCurrentStep(data.currentStep)
        if (data.totalSteps != null) setTotalSteps(data.totalSteps)
        if (data.stepDescription) setStepDescription(data.stepDescription)

        const isActive = (data.status === 'running' || data.status === 'pending') &&
          (data.operation === 'install' || data.operation === 'uninstall')
        setJobActive(isActive)
      } catch {
        // Ignore polling errors
      }
    }

    pollProgress()
    const interval = setInterval(pollProgress, 5000)
    return () => clearInterval(interval)
  }, [installationId, verificationStatus])

  // Poll for deployment verification every 5 seconds
  useEffect(() => {
    if (!installationKey) return

    const checkVerification = async () => {
      try {
        const response = await fetch('/api/installations/check-installation', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ installationKey }),
        })

        const data = await response.json()

        // Check verification and DNS status
        if (data.verified && data.dnsConfigured) {
          setVerificationStatus('complete')
          // Auto-redirect to complete page after 2 seconds
          setTimeout(() => {
            router.push('/installations/new/complete')
          }, 2000)
        } else if (data.verified) {
          setVerificationStatus('dns_pending')
        }
      } catch (error) {
        console.error('Error checking verification:', error)
      }
    }

    // Check immediately
    checkVerification()

    // Then poll every 5 seconds
    const interval = setInterval(checkVerification, 5000)

    return () => clearInterval(interval)
  }, [installationKey, router])

  const copyCommand = (command: string) => {
    navigator.clipboard.writeText(command)
    toast.success('Command copied to clipboard')
  }

  const CLOUD_PROVIDERS = getCloudProviderCommands(installationKey, awsRegion, gcpRegion, azureLocation)

  return (
    <div className="lg:flex lg:gap-12">
      {/* Left Column - Information */}
      <div className="hidden lg:block lg:w-[400px] lg:flex-shrink-0">
        <div className="sticky top-6 space-y-6">
          {/* Background Icon */}
          <div className="absolute -top-4 -right-4 -z-10 opacity-5">
            <svg width="300" height="300" viewBox="0 0 24 24" fill="currentColor">
              <path d="M19 3H5c-1.1 0-2 .9-2 2v14c0 1.1.9 2 2 2h14c1.1 0 2-.9 2-2V5c0-1.1-.9-2-2-2zm-5 14h-2v-2h2v2zm0-4h-2V7h2v6z"/>
            </svg>
          </div>

          {/* What's Next Card */}
          <div className="border border-foreground/10 rounded-lg p-6 bg-muted/20">
            <h3 className="text-sm font-semibold text-foreground mb-3">What happens next?</h3>
            <div className="space-y-3">
              <div className="flex gap-3">
                <div className="flex-shrink-0 w-5 h-5 rounded-full bg-primary/10 text-primary flex items-center justify-center text-xs font-semibold">&#x2713;</div>
                <div>
                  <p className="text-sm font-medium text-foreground">Create installation</p>
                  <p className="text-xs text-muted-foreground mt-0.5">Set up your installation details</p>
                </div>
              </div>
              <div className="flex gap-3">
                <div className="flex-shrink-0 w-5 h-5 rounded-full bg-primary/10 text-primary flex items-center justify-center text-xs font-semibold">2</div>
                <div>
                  <p className="text-sm font-medium text-foreground">Deploy to cloud</p>
                  <p className="text-xs text-muted-foreground mt-0.5">Install Kloudlite in your infrastructure</p>
                </div>
              </div>
              <div className="flex gap-3">
                <div className="flex-shrink-0 w-5 h-5 rounded-full bg-muted text-muted-foreground flex items-center justify-center text-xs font-semibold">3</div>
                <div>
                  <p className="text-sm font-medium text-foreground">Verify & complete</p>
                  <p className="text-xs text-muted-foreground mt-0.5">Confirm your installation is ready</p>
                </div>
              </div>
            </div>
          </div>

          {/* Quick Tips Card */}
          <div className="border border-foreground/10 rounded-lg p-6 bg-background">
            <h3 className="text-sm font-semibold text-foreground mb-3">Quick Tips</h3>
            <ul className="space-y-2 text-sm text-muted-foreground">
              <li className="flex items-start gap-2">
                <span className="text-primary mt-0.5">&#x2022;</span>
                <span>Make sure your cloud provider CLI is configured</span>
              </li>
              <li className="flex items-start gap-2">
                <span className="text-primary mt-0.5">&#x2022;</span>
                <span>The installation takes approximately 10-15 minutes</span>
              </li>
              <li className="flex items-start gap-2">
                <span className="text-primary mt-0.5">&#x2022;</span>
                <span>Keep this window open while the deployment runs</span>
              </li>
            </ul>
          </div>
        </div>
      </div>

      {/* Right Column - Main Content */}
      <div className="space-y-6 lg:flex-1 lg:min-w-0">
        {/* Header */}
        <div>
          <h1 className="text-foreground text-2xl font-semibold tracking-tight">
            Install Kloudlite in Your Cloud
          </h1>
          <p className="text-muted-foreground mt-1 text-sm">
            Run the installation command on your cloud provider
          </p>
        </div>

        {/* Installation Commands Card */}
        <div className="border border-foreground/10 rounded-lg bg-background">
          <div className="border-b border-foreground/10 px-6 py-4">
            <h3 className="font-medium text-foreground">Installation Command</h3>
            <p className="text-muted-foreground mt-1 text-sm">
              Choose your cloud provider and run the command in your terminal
            </p>
          </div>
          <div className="p-6 space-y-6">
          <Tabs value={selectedProvider} onValueChange={setSelectedProvider}>
            <TabsList className="inline-flex gap-1 rounded-lg bg-muted/50 p-1">
              <TabsTrigger value="aws" className="rounded-md px-3.5 py-1.5 text-sm data-[state=active]:bg-background data-[state=active]:shadow-sm">
                AWS
              </TabsTrigger>
              <TabsTrigger value="gcp" className="rounded-md px-3.5 py-1.5 text-sm data-[state=active]:bg-background data-[state=active]:shadow-sm">
                GCP
              </TabsTrigger>
              <TabsTrigger value="azure" className="rounded-md px-3.5 py-1.5 text-sm data-[state=active]:bg-background data-[state=active]:shadow-sm">
                Azure
              </TabsTrigger>
            </TabsList>

            {Object.entries(CLOUD_PROVIDERS).map(([key, config]) => (
              <TabsContent key={key} value={key} className="mt-6 space-y-6">
                <div className="space-y-5">
                  {/* AWS Region Selector */}
                  {key === 'aws' && (
                    <div>
                      <p className="text-foreground mb-3 text-sm font-medium">Select AWS Region:</p>
                      <Popover open={awsOpen} onOpenChange={setAwsOpen}>
                        <PopoverTrigger asChild>
                          <Button
                            variant="outline"
                            role="combobox"
                            aria-expanded={awsOpen}
                            className="w-full md:w-80 justify-between rounded-sm"
                          >
                            {awsRegion
                              ? AWS_REGIONS.find((region) => region.value === awsRegion)?.label
                              : "Use AWS CLI default region"}
                            <ChevronsUpDown className="ml-2 h-4 w-4 shrink-0 opacity-50" />
                          </Button>
                        </PopoverTrigger>
                        <PopoverContent className="w-full md:w-80 p-0">
                          <Command>
                            <CommandInput placeholder="Search regions..." />
                            <CommandList>
                              <CommandEmpty>No region found.</CommandEmpty>
                              <CommandGroup>
                                {AWS_REGIONS.map((region) => (
                                  <CommandItem
                                    key={region.value || 'default'}
                                    value={region.label}
                                    onSelect={() => {
                                      setAwsRegion(region.value || '')
                                      setAwsOpen(false)
                                    }}
                                  >
                                    <Check
                                      className={cn(
                                        "mr-2 h-4 w-4",
                                        awsRegion === region.value ? "opacity-100" : "opacity-0"
                                      )}
                                    />
                                    {region.label}
                                  </CommandItem>
                                ))}
                              </CommandGroup>
                            </CommandList>
                          </Command>
                        </PopoverContent>
                      </Popover>
                    </div>
                  )}

                  {/* GCP Region Selector */}
                  {key === 'gcp' && (
                    <div>
                      <p className="text-foreground mb-3 text-sm font-medium">Select GCP Region:</p>
                      <Popover open={gcpOpen} onOpenChange={setGcpOpen}>
                        <PopoverTrigger asChild>
                          <Button
                            variant="outline"
                            role="combobox"
                            aria-expanded={gcpOpen}
                            className="w-full md:w-80 justify-between rounded-sm"
                          >
                            {gcpRegion
                              ? GCP_REGIONS.find((region) => region.value === gcpRegion)?.label
                              : "Select a region"}
                            <ChevronsUpDown className="ml-2 h-4 w-4 shrink-0 opacity-50" />
                          </Button>
                        </PopoverTrigger>
                        <PopoverContent className="w-full md:w-80 p-0">
                          <Command>
                            <CommandInput placeholder="Search regions..." />
                            <CommandList>
                              <CommandEmpty>No region found.</CommandEmpty>
                              <CommandGroup>
                                {GCP_REGIONS.map((region) => (
                                  <CommandItem
                                    key={region.value}
                                    value={region.label}
                                    onSelect={() => {
                                      setGcpRegion(region.value)
                                      setGcpOpen(false)
                                    }}
                                  >
                                    <Check
                                      className={cn(
                                        "mr-2 h-4 w-4",
                                        gcpRegion === region.value ? "opacity-100" : "opacity-0"
                                      )}
                                    />
                                    {region.label}
                                  </CommandItem>
                                ))}
                              </CommandGroup>
                            </CommandList>
                          </Command>
                        </PopoverContent>
                      </Popover>
                    </div>
                  )}

                  {/* Azure Location Selector */}
                  {key === 'azure' && (
                    <div>
                      <p className="text-foreground mb-3 text-sm font-medium">Select Azure Location:</p>
                      <Popover open={azureOpen} onOpenChange={setAzureOpen}>
                        <PopoverTrigger asChild>
                          <Button
                            variant="outline"
                            role="combobox"
                            aria-expanded={azureOpen}
                            className="w-full md:w-80 justify-between rounded-sm"
                          >
                            {azureLocation
                              ? AZURE_LOCATIONS.find((location) => location.value === azureLocation)?.label
                              : "Select a location"}
                            <ChevronsUpDown className="ml-2 h-4 w-4 shrink-0 opacity-50" />
                          </Button>
                        </PopoverTrigger>
                        <PopoverContent className="w-full md:w-80 p-0">
                          <Command>
                            <CommandInput placeholder="Search locations..." />
                            <CommandList>
                              <CommandEmpty>No location found.</CommandEmpty>
                              <CommandGroup>
                                {AZURE_LOCATIONS.map((location) => (
                                  <CommandItem
                                    key={location.value}
                                    value={location.label}
                                    onSelect={() => {
                                      setAzureLocation(location.value)
                                      setAzureOpen(false)
                                    }}
                                  >
                                    <Check
                                      className={cn(
                                        "mr-2 h-4 w-4",
                                        azureLocation === location.value ? "opacity-100" : "opacity-0"
                                      )}
                                    />
                                    {location.label}
                                  </CommandItem>
                                ))}
                              </CommandGroup>
                            </CommandList>
                          </Command>
                        </PopoverContent>
                      </Popover>
                    </div>
                  )}

                  <div>
                    <p className="text-foreground mb-2 text-sm font-medium">Prerequisites:</p>
                    <ul className="text-muted-foreground space-y-1.5 text-sm">
                      {config.requirements.map((req) => (
                        <li key={req} className="flex items-start gap-2">
                          <div className="bg-muted-foreground mt-1.5 size-1.5 flex-shrink-0 rounded-full" />
                          <span>{req}</span>
                        </li>
                      ))}
                    </ul>
                  </div>

                  <div>
                    <p className="text-foreground mb-2 text-sm font-medium">Run this command:</p>
                    <div className="space-y-3">
                      {config.commands.map((cmd) => (
                        <div key={cmd} className="bg-muted rounded-lg p-4">
                          <div className="flex items-start justify-between gap-4">
                            <code className="flex-1 font-mono text-sm leading-relaxed break-all">
                              {cmd}
                            </code>
                            <Button
                              variant="outline"
                              size="sm"
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
        </div>

        {/* Progress Bar - shown when CLI is actively running */}
        {jobActive && totalSteps > 0 && verificationStatus !== 'complete' && (
          <div className="border border-foreground/10 rounded-lg bg-background p-6">
            <div className="space-y-3">
              <div className="flex items-center justify-between text-sm">
                <span className="text-muted-foreground">
                  {currentStep > 0 ? `Step ${currentStep} of ${totalSteps}` : 'Starting...'}
                </span>
                <span className="text-muted-foreground font-medium">
                  {totalSteps > 0 ? `${Math.round((currentStep / totalSteps) * 100)}%` : '0%'}
                </span>
              </div>
              <div className="h-2 bg-foreground/[0.06] rounded-full overflow-hidden">
                <div
                  className="h-full bg-blue-600 dark:bg-blue-500 rounded-full transition-all duration-500 ease-out"
                  style={{ width: `${totalSteps > 0 ? (currentStep / totalSteps) * 100 : 0}%` }}
                />
              </div>
              {stepDescription && (
                <p className="text-xs text-muted-foreground truncate">
                  {stepDescription}
                </p>
              )}
            </div>
          </div>
        )}

        {/* Verification Status */}
        <div className="flex items-center justify-center gap-3 text-base">
          {verificationStatus === 'waiting' && (
            <>
              <Loader2 className="size-4 animate-spin text-blue-600" />
              <span className="text-muted-foreground">
                {jobActive && currentStep > 0
                  ? `Step ${currentStep} of ${totalSteps} — ${stepDescription || 'In progress...'}`
                  : 'Waiting for deployment...'}
              </span>
            </>
          )}
          {verificationStatus === 'dns_pending' && (
            <>
              <Loader2 className="size-4 animate-spin text-yellow-600" />
              <span className="text-yellow-600">Deployment verified. Configuring DNS...</span>
            </>
          )}
          {verificationStatus === 'complete' && (
            <>
              <CheckCircle2 className="size-4 text-green-600" />
              <span className="text-green-600 font-medium">Installation complete! Redirecting...</span>
            </>
          )}
        </div>
      </div>
    </div>
  )
}
