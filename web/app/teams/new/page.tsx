import { redirect } from "next/navigation"
import { getServerSession } from "next-auth"

import { listUserTeams } from "@/app/actions/teams"
import { Badge } from "@/components/ui/badge"
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"
import { getAuthOptions } from "@/lib/auth/get-auth-options"

import { CreateTeamForm } from "./create-team-form"


export default async function NewTeamPage() {
  const authOpts = await getAuthOptions()
  const session = await getServerSession(authOpts)

  if (!session) {
    redirect("/auth/login?callbackUrl=/teams/new")
  }

  if (!session.user.emailVerified) {
    redirect("/auth/email-verification-required")
  }

  // Check if user already has teams
  let userHasTeams = false
  try {
    const teams = await listUserTeams()
    userHasTeams = teams.length > 0
  } catch (_error) {
    // Failed to fetch teams
  }

  // TODO: Fetch available regions from backend
  const awsRegions = [
    // US Regions
    { id: "us-east-1", displayName: "US East (N. Virginia)", value: "us-east-1" },
    { id: "us-east-2", displayName: "US East (Ohio)", value: "us-east-2" },
    { id: "us-west-1", displayName: "US West (N. California)", value: "us-west-1" },
    { id: "us-west-2", displayName: "US West (Oregon)", value: "us-west-2" },
    
    // Europe Regions
    { id: "eu-west-1", displayName: "Europe (Ireland)", value: "eu-west-1" },
    { id: "eu-west-2", displayName: "Europe (London)", value: "eu-west-2" },
    { id: "eu-west-3", displayName: "Europe (Paris)", value: "eu-west-3" },
    { id: "eu-central-1", displayName: "Europe (Frankfurt)", value: "eu-central-1" },
    { id: "eu-central-2", displayName: "Europe (Zurich)", value: "eu-central-2" },
    { id: "eu-north-1", displayName: "Europe (Stockholm)", value: "eu-north-1" },
    { id: "eu-south-1", displayName: "Europe (Milan)", value: "eu-south-1" },
    { id: "eu-south-2", displayName: "Europe (Spain)", value: "eu-south-2" },
    
    // Asia Pacific Regions
    { id: "ap-south-1", displayName: "Asia Pacific (Mumbai)", value: "ap-south-1" },
    { id: "ap-south-2", displayName: "Asia Pacific (Hyderabad)", value: "ap-south-2" },
    { id: "ap-southeast-1", displayName: "Asia Pacific (Singapore)", value: "ap-southeast-1" },
    { id: "ap-southeast-2", displayName: "Asia Pacific (Sydney)", value: "ap-southeast-2" },
    { id: "ap-southeast-3", displayName: "Asia Pacific (Jakarta)", value: "ap-southeast-3" },
    { id: "ap-southeast-4", displayName: "Asia Pacific (Melbourne)", value: "ap-southeast-4" },
    { id: "ap-northeast-1", displayName: "Asia Pacific (Tokyo)", value: "ap-northeast-1" },
    { id: "ap-northeast-2", displayName: "Asia Pacific (Seoul)", value: "ap-northeast-2" },
    { id: "ap-northeast-3", displayName: "Asia Pacific (Osaka)", value: "ap-northeast-3" },
    { id: "ap-east-1", displayName: "Asia Pacific (Hong Kong)", value: "ap-east-1" },
    
    // Middle East Regions
    { id: "me-south-1", displayName: "Middle East (Bahrain)", value: "me-south-1" },
    { id: "me-central-1", displayName: "Middle East (UAE)", value: "me-central-1" },
    { id: "il-central-1", displayName: "Israel (Tel Aviv)", value: "il-central-1" },
    
    // South America Regions
    { id: "sa-east-1", displayName: "South America (São Paulo)", value: "sa-east-1" },
    
    // Africa Regions
    { id: "af-south-1", displayName: "Africa (Cape Town)", value: "af-south-1" },
    
    // Canada Regions
    { id: "ca-central-1", displayName: "Canada (Central)", value: "ca-central-1" },
    { id: "ca-west-1", displayName: "Canada (Calgary)", value: "ca-west-1" },
  ]

  return (
    <Card>
      <CardHeader className="space-y-1 text-center">
        <div className="flex items-center justify-center gap-2 mb-2">
          <CardTitle className="text-xl">Create your team</CardTitle>
          {!userHasTeams && (
            <Badge variant="secondary" className="text-xs">Required</Badge>
          )}
        </div>
        <CardDescription className="text-sm">
          {!userHasTeams 
            ? "Welcome to Kloudlite! Create your first team to get started."
            : "Set up a new team to organize your resources and collaborate."
          }
        </CardDescription>
      </CardHeader>
      <CardContent>
        <CreateTeamForm regions={awsRegions} />
      </CardContent>
    </Card>
  )
}