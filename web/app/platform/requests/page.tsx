import { redirect } from "next/navigation";
import { getServerSession } from "next-auth";

import { PageHeader } from "@/components/platform/page-header";
import { getAuthOptions } from "@/lib/auth/get-auth-options";
import { fetchPlatformData } from "@/lib/data/platform-data";

import TeamRequests from "../team-requests";

export default async function RequestsPage() {
  const authOpts = await getAuthOptions();
  const session = await getServerSession(authOpts);
  
  if (!session?.user) {
    redirect("/auth/login");
  }

  try {
    const { platformRole, teamRequests } = await fetchPlatformData();

    if (!platformRole.canManagePlatform) {
      redirect("/overview");
    }

    return (
      <>
        <PageHeader 
          title="Team Requests"
          description="Review and approve pending team creation requests"
        />
        <TeamRequests 
          requests={teamRequests} 
          session={session}
        />
      </>
    );
  } catch (error) {
    console.error("Failed to load team requests:", error);
    redirect("/overview");
  }
}