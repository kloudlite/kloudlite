import { redirect } from "next/navigation";
import { getServerSession } from "next-auth";

import { PageHeader } from "@/components/platform/page-header";
import { getAuthOptions } from "@/lib/auth/get-auth-options";
import { fetchPlatformData } from "@/lib/data/platform-data";

import PlatformUsers from "../platform-users";

export default async function UsersPage() {
  const authOpts = await getAuthOptions();
  const session = await getServerSession(authOpts);
  
  if (!session?.user) {
    redirect("/auth/login");
  }

  try {
    const { platformRole, platformUsers, platformInvitations } = await fetchPlatformData();

    if (!platformRole.canManagePlatform) {
      redirect("/overview");
    }

    // Only super admins can access this page
    if (platformRole.role !== "super_admin") {
      redirect("/platform");
    }

    return (
      <>
        <PageHeader 
          title="Platform Users"
          description="Manage platform administrators and user permissions"
        />
        <PlatformUsers 
          users={platformUsers || []} 
          session={session}
          initialInvitations={platformInvitations || []}
        />
      </>
    );
  } catch (error) {
    console.error("Failed to load platform users:", error);
    redirect("/overview");
  }
}