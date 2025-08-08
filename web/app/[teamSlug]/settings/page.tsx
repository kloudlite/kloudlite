import { Settings, Shield, Bell, Trash2, Download, Key, Globe, CreditCard } from "lucide-react"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Switch } from "@/components/ui/switch"
import { Separator } from "@/components/ui/separator"
import { Textarea } from "@/components/ui/textarea"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"

interface SettingsPageProps {
  params: Promise<{ teamSlug: string }>
}

export default async function SettingsPage({ params }: SettingsPageProps) {
  const { teamSlug } = await params

  // TODO: Fetch actual team settings
  const team = {
    name: "Engineering Team",
    slug: teamSlug,
    description: "Main engineering team for product development",
    visibility: "private",
    defaultRegion: "us-west-2",
    billing: {
      plan: "Pro",
      monthlyLimit: 5000,
      currentUsage: 2340
    }
  }

  return (
    <div className="space-y-4 md:space-y-6">
      {/* Page Header */}
      <div>
        <h1 className="text-2xl md:text-3xl font-extralight tracking-tight">Team Settings</h1>
        <p className="text-sm md:text-base text-muted-foreground mt-2">
          Manage your team's configuration and preferences
        </p>
      </div>

      {/* General Settings */}
      <Card>
        <CardHeader>
          <div className="flex items-center gap-2">
            <Settings className="h-5 w-5 text-muted-foreground" />
            <CardTitle>General Settings</CardTitle>
          </div>
          <CardDescription>
            Basic information about your team
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="grid gap-4 sm:grid-cols-2">
            <div className="space-y-2">
              <Label htmlFor="team-name">Team Name</Label>
              <Input id="team-name" defaultValue={team.name} />
            </div>
            <div className="space-y-2">
              <Label htmlFor="team-slug">URL Slug</Label>
              <Input id="team-slug" defaultValue={team.slug} disabled />
            </div>
          </div>
          <div className="space-y-2">
            <Label htmlFor="description">Description</Label>
            <Textarea 
              id="description" 
              defaultValue={team.description}
              rows={3}
            />
          </div>
          <div className="flex justify-end">
            <Button>Save Changes</Button>
          </div>
        </CardContent>
      </Card>

      {/* Security & Privacy */}
      <Card>
        <CardHeader>
          <div className="flex items-center gap-2">
            <Shield className="h-5 w-5 text-muted-foreground" />
            <CardTitle>Security & Privacy</CardTitle>
          </div>
          <CardDescription>
            Control access and visibility settings
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
            <div className="space-y-0.5">
              <Label>Team Visibility</Label>
              <p className="text-xs sm:text-sm text-muted-foreground">
                Make your team discoverable to other users
              </p>
            </div>
            <Select defaultValue={team.visibility}>
              <SelectTrigger className="w-24 sm:w-32">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="private">Private</SelectItem>
                <SelectItem value="public">Public</SelectItem>
              </SelectContent>
            </Select>
          </div>
          <Separator />
          <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
            <div className="space-y-0.5">
              <Label>Two-Factor Authentication</Label>
              <p className="text-xs sm:text-sm text-muted-foreground">
                Require 2FA for all team members
              </p>
            </div>
            <Switch />
          </div>
          <Separator />
          <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
            <div className="space-y-0.5">
              <Label>IP Allowlist</Label>
              <p className="text-xs sm:text-sm text-muted-foreground">
                Restrict access to specific IP addresses
              </p>
            </div>
            <Button variant="outline" size="sm" className="w-full sm:w-auto">Configure</Button>
          </div>
        </CardContent>
      </Card>

      {/* Notifications */}
      <Card>
        <CardHeader>
          <div className="flex items-center gap-2">
            <Bell className="h-5 w-5 text-muted-foreground" />
            <CardTitle>Notifications</CardTitle>
          </div>
          <CardDescription>
            Configure how your team receives notifications
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
            <div className="space-y-0.5">
              <Label>Email Notifications</Label>
              <p className="text-xs sm:text-sm text-muted-foreground">
                Send important updates via email
              </p>
            </div>
            <Switch defaultChecked />
          </div>
          <Separator />
          <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
            <div className="space-y-0.5">
              <Label>Deployment Alerts</Label>
              <p className="text-xs sm:text-sm text-muted-foreground">
                Notify on deployment status changes
              </p>
            </div>
            <Switch defaultChecked />
          </div>
          <Separator />
          <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
            <div className="space-y-0.5">
              <Label>Resource Limits</Label>
              <p className="text-xs sm:text-sm text-muted-foreground">
                Alert when approaching resource limits
              </p>
            </div>
            <Switch defaultChecked />
          </div>
        </CardContent>
      </Card>

      {/* Regional Settings */}
      <Card>
        <CardHeader>
          <div className="flex items-center gap-2">
            <Globe className="h-5 w-5 text-muted-foreground" />
            <CardTitle>Regional Settings</CardTitle>
          </div>
          <CardDescription>
            Default region for new resources
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="space-y-2">
            <Label htmlFor="default-region">Default Region</Label>
            <Select defaultValue={team.defaultRegion}>
              <SelectTrigger id="default-region">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="us-west-1">US West (N. California)</SelectItem>
                <SelectItem value="us-west-2">US West (Oregon)</SelectItem>
                <SelectItem value="us-east-1">US East (N. Virginia)</SelectItem>
                <SelectItem value="us-east-2">US East (Ohio)</SelectItem>
                <SelectItem value="eu-west-1">EU (Ireland)</SelectItem>
                <SelectItem value="eu-central-1">EU (Frankfurt)</SelectItem>
                <SelectItem value="ap-southeast-1">Asia Pacific (Singapore)</SelectItem>
              </SelectContent>
            </Select>
          </div>
        </CardContent>
      </Card>

      {/* Billing */}
      <Card>
        <CardHeader>
          <div className="flex items-center gap-2">
            <CreditCard className="h-5 w-5 text-muted-foreground" />
            <CardTitle>Billing & Usage</CardTitle>
          </div>
          <CardDescription>
            Monitor and control team spending
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="grid gap-4 sm:grid-cols-2">
            <div className="space-y-2">
              <Label>Current Plan</Label>
              <div className="font-medium">{team.billing.plan}</div>
            </div>
            <div className="space-y-2">
              <Label>Monthly Limit</Label>
              <div className="font-medium">${team.billing.monthlyLimit}</div>
            </div>
          </div>
          <div className="space-y-2">
            <Label>Current Usage</Label>
            <div className="flex items-center gap-2">
              <div className="flex-1 bg-muted rounded-full h-2">
                <div 
                  className="bg-primary rounded-full h-2"
                  style={{ width: `${(team.billing.currentUsage / team.billing.monthlyLimit) * 100}%` }}
                />
              </div>
              <span className="text-sm font-medium">
                ${team.billing.currentUsage} / ${team.billing.monthlyLimit}
              </span>
            </div>
          </div>
          <div className="flex flex-col sm:flex-row gap-2">
            <Button variant="outline" className="w-full sm:w-auto">View Details</Button>
            <Button variant="outline" className="w-full sm:w-auto">Update Plan</Button>
          </div>
        </CardContent>
      </Card>

      {/* API Keys */}
      <Card>
        <CardHeader>
          <div className="flex items-center gap-2">
            <Key className="h-5 w-5 text-muted-foreground" />
            <CardTitle>API Keys</CardTitle>
          </div>
          <CardDescription>
            Manage API keys for programmatic access
          </CardDescription>
        </CardHeader>
        <CardContent>
          <Button variant="outline" className="w-full sm:w-auto">Manage API Keys</Button>
        </CardContent>
      </Card>

      {/* Data Export */}
      <Card>
        <CardHeader>
          <div className="flex items-center gap-2">
            <Download className="h-5 w-5 text-muted-foreground" />
            <CardTitle>Export Data</CardTitle>
          </div>
          <CardDescription>
            Download your team's data
          </CardDescription>
        </CardHeader>
        <CardContent>
          <Button variant="outline" className="gap-2 w-full sm:w-auto">
            <Download className="h-4 w-4" />
            Export Team Data
          </Button>
        </CardContent>
      </Card>

      {/* Danger Zone */}
      <Card className="border-destructive/50">
        <CardHeader>
          <div className="flex items-center gap-2">
            <Trash2 className="h-5 w-5 text-destructive" />
            <CardTitle className="text-destructive">Danger Zone</CardTitle>
          </div>
          <CardDescription>
            Irreversible actions that affect your entire team
          </CardDescription>
        </CardHeader>
        <CardContent>
          <Button variant="destructive" className="gap-2 w-full sm:w-auto">
            <Trash2 className="h-4 w-4" />
            Delete Team
          </Button>
        </CardContent>
      </Card>
    </div>
  )
}