import { Shield, Clock, Users, Building } from "lucide-react"
import { redirect } from "next/navigation";
import { getServerSession } from "next-auth";

import { PageHeader } from "@/components/platform/page-header"
import { StatCard } from "@/components/platform/stat-card"
import { Badge } from "@/components/ui/badge"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { getAuthOptions } from "@/lib/auth/get-auth-options";
import { fetchPlatformData } from "@/lib/data/platform-data";

export default async function PlatformPage() {
  const authOpts = await getAuthOptions();
  const session = await getServerSession(authOpts);
  
  if (!session?.user) {
    redirect("/auth/login");
  }

  try {
    const { platformRole, platformSettings, platformUsers, teamRequests } = await fetchPlatformData();

    if (!platformRole.canManagePlatform) {
      redirect("/overview");
    }

    const isSuperAdmin = platformRole.role === "super_admin";

    return (
      <>
        {/* Page Header */}
        <PageHeader 
          title="Platform Overview"
          description="Monitor and manage your platform's health and configuration"
        />

        {/* Content */}
        <div className="space-y-6">
          {/* Stats Grid */}
          <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
            <StatCard
              title="Your Role"
              value={platformRole.role.replace("_", " ")}
              icon={Shield}
              color="primary"
            />
            
            <StatCard
              title="Pending Requests"
              value={teamRequests.length}
              icon={Clock}
              color="orange"
            />

            {isSuperAdmin && platformUsers && (
              <>
                <StatCard
                  title="Platform Users"
                  value={platformUsers.length}
                  icon={Users}
                  color="blue"
                />
                
                <StatCard
                  title="Registration"
                  value={platformSettings?.allowSignup ? "Open" : "Closed"}
                  icon={Building}
                  color="green"
                />
              </>
            )}
          </div>

          {/* Platform Configuration Card */}
          <Card className="transition-all duration-200 hover:shadow-lg">
            <CardHeader>
              <CardTitle>Platform Configuration</CardTitle>
              <CardDescription>
                Key platform settings and integrations
              </CardDescription>
            </CardHeader>
            <CardContent>
              <div className="space-y-6">
                <div className="grid gap-6 sm:grid-cols-2">
                  <div>
                    <label className="text-sm font-medium text-muted-foreground">
                      Platform Owner
                    </label>
                    <p className="mt-1 text-sm font-medium">
                      {platformSettings?.platformOwnerEmail || "Not configured"}
                    </p>
                  </div>
                  <div>
                    <label className="text-sm font-medium text-muted-foreground">
                      Support Email
                    </label>
                    <p className="mt-1 text-sm font-medium">
                      {platformSettings?.supportEmail || "Not configured"}
                    </p>
                  </div>
                  <div>
                    <label className="text-sm font-medium text-muted-foreground">
                      Team Approval
                    </label>
                    <p className="mt-1 text-sm font-medium">
                      {platformSettings?.teamSettings?.requireApproval
                        ? "Required for new teams"
                        : "Automatic approval"}
                    </p>
                  </div>
                  <div>
                    <label className="text-sm font-medium text-muted-foreground">
                      OAuth Providers
                    </label>
                    <div className="mt-1 flex flex-wrap gap-2">
                      {platformSettings?.oauthProviders?.google?.enabled && (
                        <Badge variant="secondary" className="gap-1 transition-all duration-200">
                          <div className="h-1.5 w-1.5 rounded-full bg-green-500" />
                          Google
                        </Badge>
                      )}
                      {platformSettings?.oauthProviders?.github?.enabled && (
                        <Badge variant="secondary" className="gap-1 transition-all duration-200">
                          <div className="h-1.5 w-1.5 rounded-full bg-green-500" />
                          GitHub
                        </Badge>
                      )}
                      {platformSettings?.oauthProviders?.microsoft?.enabled && (
                        <Badge variant="secondary" className="gap-1 transition-all duration-200">
                          <div className="h-1.5 w-1.5 rounded-full bg-green-500" />
                          Microsoft
                        </Badge>
                      )}
                      {!platformSettings?.oauthProviders?.google?.enabled && 
                       !platformSettings?.oauthProviders?.github?.enabled && 
                       !platformSettings?.oauthProviders?.microsoft?.enabled && (
                        <span className="text-sm text-muted-foreground">None enabled</span>
                      )}
                    </div>
                  </div>
                </div>
              </div>
            </CardContent>
          </Card>
        </div>
      </>
    );
  } catch (error) {
    console.error("Failed to load platform data:", error);
    redirect("/overview");
  }
}