'use client'

import { Card, CardContent, CardDescription, CardHeader, CardTitle, Badge, Separator, Button } from '@kloudlite/ui'
import { Copy, ExternalLink, Cloud } from 'lucide-react'
import { toast } from 'sonner'
import type { Installation } from '@/lib/console/supabase-storage-service'

interface InstallationDetailsCardProps {
  installation: Installation
  status: {
    label: string
    color: string
    description: string
  }
  domain: string
  installationUrl: string | null
}

const CLOUD_PROVIDER_LABELS: Record<string, string> = {
  aws: 'AWS',
  gcp: 'Google Cloud',
  azure: 'Azure',
}

const AZURE_LOCATION_LABELS: Record<string, string> = {
  eastus: 'East US',
  eastus2: 'East US 2',
  westus: 'West US',
  westus2: 'West US 2',
  westus3: 'West US 3',
  centralus: 'Central US',
  northcentralus: 'North Central US',
  southcentralus: 'South Central US',
  westeurope: 'West Europe',
  northeurope: 'North Europe',
  uksouth: 'UK South',
  ukwest: 'UK West',
  francecentral: 'France Central',
  germanywestcentral: 'Germany West Central',
  swedencentral: 'Sweden Central',
  southeastasia: 'Southeast Asia',
  eastasia: 'East Asia',
  japaneast: 'Japan East',
  japanwest: 'Japan West',
  koreacentral: 'Korea Central',
  australiaeast: 'Australia East',
  australiasoutheast: 'Australia Southeast',
  centralindia: 'Central India',
  southindia: 'South India',
  brazilsouth: 'Brazil South',
  canadacentral: 'Canada Central',
  canadaeast: 'Canada East',
}

const AWS_REGION_LABELS: Record<string, string> = {
  'us-east-1': 'US East (N. Virginia)',
  'us-east-2': 'US East (Ohio)',
  'us-west-1': 'US West (N. California)',
  'us-west-2': 'US West (Oregon)',
  'eu-west-1': 'EU (Ireland)',
  'eu-west-2': 'EU (London)',
  'eu-west-3': 'EU (Paris)',
  'eu-central-1': 'EU (Frankfurt)',
  'eu-north-1': 'EU (Stockholm)',
  'ap-south-1': 'Asia Pacific (Mumbai)',
  'ap-southeast-1': 'Asia Pacific (Singapore)',
  'ap-southeast-2': 'Asia Pacific (Sydney)',
  'ap-northeast-1': 'Asia Pacific (Tokyo)',
  'ap-northeast-2': 'Asia Pacific (Seoul)',
  'sa-east-1': 'South America (São Paulo)',
  'ca-central-1': 'Canada (Central)',
}

export function InstallationDetailsCard({
  installation,
  status,
  domain,
  installationUrl,
}: InstallationDetailsCardProps) {
  const copyToClipboard = (text: string, label: string) => {
    navigator.clipboard.writeText(text)
    toast.success(`${label} copied to clipboard`)
  }

  // Generate resource group name for Azure
  const azureResourceGroup = installation.cloudProvider === 'azure'
    ? `kl-${installation.installationKey}-rg`
    : null

  // Get location label
  const getLocationLabel = () => {
    if (!installation.cloudLocation) return null
    if (installation.cloudProvider === 'azure') {
      return AZURE_LOCATION_LABELS[installation.cloudLocation] || installation.cloudLocation
    }
    if (installation.cloudProvider === 'aws') {
      return AWS_REGION_LABELS[installation.cloudLocation] || installation.cloudLocation
    }
    return installation.cloudLocation
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>Installation Details</CardTitle>
        <CardDescription>View and manage your installation configuration</CardDescription>
      </CardHeader>
      <CardContent className="space-y-6">
        {/* Status */}
        <div>
          <label className="text-foreground text-sm font-medium">Status</label>
          <div className="mt-2 flex items-center gap-3">
            <Badge className={status.color}>{status.label}</Badge>
            <span className="text-muted-foreground text-sm">{status.description}</span>
          </div>
        </div>

        <Separator />

        {/* Installation Key */}
        <div>
          <label className="text-foreground text-sm font-medium">Installation Key</label>
          <div className="mt-2 flex items-center gap-2">
            <code className="bg-muted flex-1 rounded-md px-3 py-2 font-mono text-sm">
              {installation.installationKey}
            </code>
            <Button
              variant="outline"
              size="sm"
              onClick={() => copyToClipboard(installation.installationKey, 'Installation key')}
            >
              <Copy className="h-4 w-4" />
            </Button>
          </div>
          <p className="text-muted-foreground mt-1 text-xs">
            Use this key during installation deployment
          </p>
        </div>

        {/* Domain */}
        {installation.subdomain && installationUrl && (
          <>
            <Separator />
            <div>
              <label className="text-foreground text-sm font-medium">Installation URL</label>
              <div className="mt-2 flex items-center gap-2">
                <a
                  href={installationUrl}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="text-primary flex-1 font-mono text-sm hover:underline"
                >
                  {installation.subdomain}.{domain}
                </a>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => window.open(installationUrl, '_blank')}
                >
                  <ExternalLink className="h-4 w-4" />
                </Button>
              </div>
              {installation.reservedAt && (
                <p className="text-muted-foreground mt-1 text-xs">
                  Reserved on {new Date(installation.reservedAt).toLocaleDateString()}
                </p>
              )}
            </div>
          </>
        )}

        {/* Cloud Provider Info */}
        {installation.cloudProvider && (
          <>
            <Separator />
            <div>
              <label className="text-foreground text-sm font-medium">Cloud Provider</label>
              <div className="mt-2 flex items-center gap-3">
                <Badge variant="outline" className="flex items-center gap-1.5">
                  <Cloud className="h-3 w-3" />
                  {CLOUD_PROVIDER_LABELS[installation.cloudProvider] || installation.cloudProvider}
                </Badge>
                {getLocationLabel() && (
                  <span className="text-muted-foreground text-sm">{getLocationLabel()}</span>
                )}
              </div>
            </div>

            {/* Azure Resource Group */}
            {installation.cloudProvider === 'azure' && azureResourceGroup && (
              <div className="mt-4">
                <label className="text-foreground text-sm font-medium">Resource Group</label>
                <div className="mt-2 flex items-center gap-2">
                  <code className="bg-muted flex-1 rounded-md px-3 py-2 font-mono text-sm">
                    {azureResourceGroup}
                  </code>
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => copyToClipboard(azureResourceGroup, 'Resource group')}
                  >
                    <Copy className="h-4 w-4" />
                  </Button>
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => window.open(`https://portal.azure.com/#@/resource/subscriptions//resourceGroups/${azureResourceGroup}`, '_blank')}
                  >
                    <ExternalLink className="h-4 w-4" />
                  </Button>
                </div>
                <p className="text-muted-foreground mt-1 text-xs">
                  All Kloudlite resources are in this resource group
                </p>
              </div>
            )}

            {/* Uninstall Instructions */}
            <div className="mt-4 rounded-lg border border-blue-200 bg-blue-50 p-4 dark:border-blue-900 dark:bg-blue-950">
              <p className="mb-2 text-sm font-semibold text-blue-900 dark:text-blue-200">
                How to Uninstall
              </p>
              {installation.cloudProvider === 'azure' && (
                <div className="space-y-2 text-sm text-blue-900 dark:text-blue-200">
                  <p>To uninstall Kloudlite from Azure:</p>
                  <ol className="list-decimal list-inside space-y-1 ml-2">
                    <li>Go to <a href="https://portal.azure.com/#blade/HubsExtension/BrowseResourceGroups" target="_blank" rel="noopener noreferrer" className="underline hover:no-underline">Azure Portal → Resource Groups</a></li>
                    <li>Find <code className="bg-blue-100 dark:bg-blue-900 px-1 rounded">{azureResourceGroup}</code></li>
                    <li>Click &quot;Delete resource group&quot;</li>
                  </ol>
                  <p className="mt-2 text-xs opacity-80">This will delete all Kloudlite resources including the VM, networking, and storage.</p>
                </div>
              )}
              {installation.cloudProvider === 'aws' && (
                <div className="space-y-2 text-sm text-blue-900 dark:text-blue-200">
                  <p>To uninstall Kloudlite from AWS, run:</p>
                  <div className="bg-blue-100 dark:bg-blue-900 rounded p-2 mt-2">
                    <code className="text-xs break-all">
                      curl -fsSL https://get.khost.dev/uninstall/aws | bash -s -- --key {installation.installationKey} --region {installation.cloudLocation}
                    </code>
                  </div>
                </div>
              )}
              {installation.cloudProvider === 'gcp' && (
                <div className="space-y-2 text-sm text-blue-900 dark:text-blue-200">
                  <p>To uninstall Kloudlite from GCP, run:</p>
                  <div className="bg-blue-100 dark:bg-blue-900 rounded p-2 mt-2">
                    <code className="text-xs break-all">
                      curl -fsSL https://get.khost.dev/uninstall/gcp | bash -s -- --key {installation.installationKey} --region {installation.cloudLocation}
                    </code>
                  </div>
                </div>
              )}
            </div>
          </>
        )}

        {/* Timestamps */}
        <Separator />
        <div className="grid grid-cols-2 gap-4 text-sm">
          <div>
            <label className="text-foreground font-medium">Created</label>
            <p className="text-muted-foreground mt-1">
              {new Date(installation.createdAt).toLocaleString()}
            </p>
          </div>
          {installation.lastHealthCheck && (
            <div>
              <label className="text-foreground font-medium">Last Health Check</label>
              <p className="text-muted-foreground mt-1">
                {new Date(installation.lastHealthCheck).toLocaleString()}
              </p>
            </div>
          )}
        </div>
      </CardContent>
    </Card>
  )
}
