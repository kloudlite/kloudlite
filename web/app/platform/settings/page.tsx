import { redirect } from "next/navigation";
import { getServerSession } from "next-auth";

import { PageHeader } from "@/components/platform/page-header";
import { getAuthOptions } from "@/lib/auth/get-auth-options";
import { fetchPlatformData } from "@/lib/data/platform-data";

import PlatformSettings from "../platform-settings";

export default async function SettingsPage() {
  const authOpts = await getAuthOptions();
  const session = await getServerSession(authOpts);
  
  if (!session?.user) {
    redirect("/auth/login");
  }

  try {
    const { platformRole, platformSettings } = await fetchPlatformData();

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
          title="Platform Settings"
          description="Configure platform-wide settings and integrations"
        />
        <PlatformSettings 
          settings={platformSettings} 
          session={session}
        />
      </>
    );
  } catch (error) {
    console.error("Failed to load platform settings:", error);
    redirect("/overview");
  }
}