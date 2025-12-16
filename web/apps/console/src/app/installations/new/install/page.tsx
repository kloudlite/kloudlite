'use client'

import { useState, useEffect } from 'react'
import { useRouter } from 'next/navigation'
import { Button, Card, CardContent, CardDescription, CardHeader, CardTitle, Tabs, TabsContent, TabsList, TabsTrigger, Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@kloudlite/ui'
import { Loader2, Copy, CheckCircle2, ExternalLink } from 'lucide-react'
import { InstallationProgress } from '@/components/installation-progress'
import { WorldMap } from '@/components/world-map'
import { toast } from 'sonner'

interface SessionData {
  user: {
    email: string
    name: string
  }
  installationKey: string
}

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

export default function InstallPage() {
  const router = useRouter()
  const [selectedProvider, setSelectedProvider] = useState('aws')
  const [awsRegion, setAwsRegion] = useState('')
  const [gcpRegion, setGcpRegion] = useState('us-central1')
  const [azureLocation, setAzureLocation] = useState('eastus')
  const [session, setSession] = useState<SessionData | null>(null)
  const [loading, setLoading] = useState(true)
  const [verificationStatus, setVerificationStatus] = useState<'waiting' | 'verified' | 'dns_pending' | 'complete' | 'error'>(
    'waiting',
  )

  // Check session on mount
  useEffect(() => {
    const checkSession = async () => {
      try {
        const response = await fetch('/api/installations/session')
        if (response.ok) {
          const data = await response.json()

          if (!data.installationKey) {
            // No installation key, redirect to create installation page
            router.push('/installations/new')
            return
          }

          setSession(data)
        } else {
          router.push('/login')
        }
      } catch (error) {
        console.error('Error checking session:', error)
        router.push('/login')
      } finally {
        setLoading(false)
      }
    }
    checkSession()
  }, [router])

  // Poll for deployment verification every 5 seconds
  useEffect(() => {
    if (!session?.installationKey) return

    const checkVerification = async () => {
      try {
        const response = await fetch('/api/installations/check-installation', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ installationKey: session.installationKey }),
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
  }, [session, router])

  // Redirect if no session after loading completes
  useEffect(() => {
    if (!loading && (!session || !session.installationKey)) {
      router.push('/installations/new')
    }
  }, [loading, session, router])

  const copyCommand = (command: string) => {
    navigator.clipboard.writeText(command)
    toast.success('Command copied to clipboard')
  }

  if (loading) {
    return (
      <div className="bg-background flex min-h-screen items-center justify-center">
        <Loader2 className="text-primary size-8 animate-spin" />
      </div>
    )
  }

  if (!session || !session.installationKey) {
    return null
  }

  const CLOUD_PROVIDERS = getCloudProviderCommands(session.installationKey, awsRegion, gcpRegion, azureLocation)

  return (
    <>
      {/* Header */}
      <div className="mb-12 text-center">
        <h1 className="text-foreground mb-3 text-4xl font-bold tracking-tight">
          Install Kloudlite in Your Cloud
        </h1>
        <p className="text-muted-foreground text-lg">
          Run the installation command on your cloud provider
        </p>
      </div>

      <InstallationProgress currentStep={2} />

      {/* Installation Commands Card */}
      <Card>
        <CardHeader className="pb-6">
          <CardTitle className="text-2xl font-bold">Installation Command</CardTitle>
          <CardDescription className="text-base">
            Choose your cloud provider and run the command in your terminal
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-6">
          <Tabs value={selectedProvider} onValueChange={setSelectedProvider}>
            <TabsList className="grid w-full grid-cols-3 p-1">
              <TabsTrigger value="aws" className="text-sm font-semibold">
                AWS
              </TabsTrigger>
              <TabsTrigger value="gcp" className="text-sm font-semibold">
                GCP
              </TabsTrigger>
              <TabsTrigger value="azure" className="text-sm font-semibold">
                Azure
              </TabsTrigger>
            </TabsList>

            {Object.entries(CLOUD_PROVIDERS).map(([key, config]) => (
              <TabsContent key={key} value={key} className="mt-6 space-y-6">
                <div className="space-y-5">
                  {/* AWS Region Selector */}
                  {key === 'aws' && (
                    <>
                      <div>
                        <p className="text-foreground mb-3 text-sm font-semibold">Select AWS Region:</p>
                        <Select
                          value={awsRegion || 'default'}
                          onValueChange={(val) => setAwsRegion(val === 'default' ? '' : val)}
                        >
                          <SelectTrigger className="w-full md:w-80">
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

                      {/* One-Click CloudFormation */}
                      <div className="bg-muted/50 rounded-lg border p-4">
                        <div className="flex items-center justify-between gap-4">
                          <div>
                            <p className="text-foreground text-sm font-semibold">One-Click Install</p>
                            <p className="text-muted-foreground text-sm">
                              {awsRegion ? 'Launch directly in AWS CloudFormation' : 'Select a region to enable one-click install'}
                            </p>
                          </div>
                          <Button
                            variant="outline"
                            disabled={!awsRegion}
                            asChild={!!awsRegion}
                          >
                            {awsRegion ? (
                              <a
                                href={`https://console.aws.amazon.com/cloudformation/home?region=${awsRegion}#/stacks/create/review?templateURL=https://kloudlite-cloudformation-templates.s3.amazonaws.com/aws/aws-oneclick.yaml&stackName=kloudlite&param_InstallationKey=${session.installationKey}&param_Region=${awsRegion}`}
                                target="_blank"
                                rel="noopener noreferrer"
                              >
                                <ExternalLink className="mr-2 size-4" />
                                Launch Stack
                              </a>
                            ) : (
                              <>
                                <ExternalLink className="mr-2 size-4" />
                                Launch Stack
                              </>
                            )}
                          </Button>
                        </div>
                      </div>
                    </>
                  )}

                  {/* GCP Region Selector */}
                  {key === 'gcp' && (
                    <div>
                      <p className="text-foreground mb-3 text-sm font-semibold">Select GCP Region:</p>
                      <Select
                        value={gcpRegion}
                        onValueChange={setGcpRegion}
                      >
                        <SelectTrigger className="w-full md:w-80">
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
                    <>
                      <div>
                        <p className="text-foreground mb-3 text-sm font-semibold">Select Azure Location:</p>
                        <Select
                          value={azureLocation}
                          onValueChange={setAzureLocation}
                        >
                          <SelectTrigger className="w-full md:w-80">
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

                      {/* One-Click Azure Deploy */}
                      <div className="bg-muted/50 rounded-lg border p-4">
                        <div className="flex items-center justify-between gap-4">
                          <div>
                            <p className="text-foreground text-sm font-semibold">One-Click Install</p>
                            <p className="text-muted-foreground text-sm">
                              {azureLocation ? 'Deploy directly in Azure Portal' : 'Select a location to enable one-click install'}
                            </p>
                          </div>
                          <Button
                            variant="outline"
                            disabled={!azureLocation}
                            asChild={!!azureLocation}
                          >
                            {azureLocation ? (
                              <a
                                href={`https://portal.azure.com/#create/Microsoft.Template/uri/${encodeURIComponent('https://kloudlite-cloudformation-templates.s3.amazonaws.com/azure/azure-oneclick.json')}`}
                                target="_blank"
                                rel="noopener noreferrer"
                                onClick={() => {
                                  navigator.clipboard.writeText(session.installationKey)
                                  toast.success('Installation key copied! Paste it in Azure Portal.')
                                }}
                              >
                                <ExternalLink className="mr-2 size-4" />
                                Deploy to Azure
                              </a>
                            ) : (
                              <>
                                <ExternalLink className="mr-2 size-4" />
                                Deploy to Azure
                              </>
                            )}
                          </Button>
                        </div>
                      </div>
                    </>
                  )}

                  {/* World Map showing selected region */}
                  <WorldMap
                    selectedRegion={key === 'aws' ? awsRegion : key === 'gcp' ? gcpRegion : azureLocation}
                    provider={key as 'aws' | 'gcp' | 'azure'}
                  />

                  <div>
                    <p className="text-foreground mb-3 text-sm font-semibold">Prerequisites:</p>
                    <ul className="text-muted-foreground space-y-2 text-sm leading-relaxed">
                      {config.requirements.map((req, idx) => (
                        <li key={idx} className="flex items-start gap-3">
                          <div className="bg-muted-foreground mt-2 size-1.5 flex-shrink-0 rounded-full" />
                          <span>{req}</span>
                        </li>
                      ))}
                    </ul>
                  </div>

                  <div>
                    <p className="text-foreground mb-3 text-sm font-semibold">Run this command:</p>
                    <div className="space-y-3">
                      {config.commands.map((cmd, idx) => (
                        <div key={idx} className="bg-muted rounded-lg p-4">
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
        </CardContent>
      </Card>

      {/* Compact Verification Status */}
      <div className="mt-6 flex items-center justify-center gap-3 text-sm">
        {verificationStatus === 'waiting' && (
          <>
            <Loader2 className="size-4 animate-spin text-blue-600" />
            <span className="text-muted-foreground">Waiting for deployment...</span>
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
    </>
  )
}
