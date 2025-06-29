"use client";

import { UserTable } from "@/components/organisms";
import { ComponentShowcase } from "../../_components/component-showcase";
import { toast } from "sonner";
import type { TeamMember } from "@/components/organisms/tables/user-table";

export default function TablesPage() {
  const mockTeamMembers: TeamMember[] = [
    {
      id: "1",
      name: "Sarah Chen",
      email: "sarah.chen@company.com",
      avatar: "https://api.dicebear.com/7.x/avataaars/svg?seed=Sarah",
      role: "owner",
      status: "active",
      joinedAt: "2024-01-15",
      lastActive: "2 minutes ago"
    },
    {
      id: "2",
      name: "Alex Rivera",
      email: "alex.rivera@company.com",
      avatar: "https://api.dicebear.com/7.x/avataaars/svg?seed=Alex",
      role: "admin",
      status: "active",
      joinedAt: "2024-02-20",
      lastActive: "1 hour ago"
    },
    {
      id: "3",
      name: "Jordan Kim",
      email: "jordan.kim@company.com",
      avatar: "https://api.dicebear.com/7.x/avataaars/svg?seed=Jordan",
      role: "developer",
      status: "active",
      joinedAt: "2024-03-10",
      lastActive: "3 hours ago"
    },
    {
      id: "4",
      name: "Taylor Smith",
      email: "taylor.smith@company.com",
      role: "developer",
      status: "pending",
      joinedAt: "2024-06-20"
    },
    {
      id: "5",
      name: "Morgan Lee",
      email: "morgan.lee@company.com",
      avatar: "https://api.dicebear.com/7.x/avataaars/svg?seed=Morgan",
      role: "viewer",
      status: "inactive",
      joinedAt: "2024-04-05",
      lastActive: "2 weeks ago"
    }
  ];

  const handleRoleChange = (memberId: string, newRole: string) => {
    toast.success(`Role updated to ${newRole}`);
  };

  const handleRemoveMember = (memberId: string) => {
    toast.success("Member removed from team");
  };

  const handleResendInvitation = (memberId: string) => {
    toast.success("Invitation resent");
  };

  const handleResetPassword = (memberId: string) => {
    toast.success("Password reset email sent");
  };

  return (
    <div className="space-y-8">
      <div>
        <h1 className="text-2xl font-bold text-slate-900 dark:text-white mb-4">
          Table Components
        </h1>
        <p className="text-slate-600 dark:text-slate-400">
          Complex table components for displaying data.
        </p>
      </div>

      <ComponentShowcase
        title="User Management Table"
        description="Complete table with user information and actions"
        contentClassName="p-0"
      >
        <UserTable
          members={mockTeamMembers}
          onRoleChange={handleRoleChange}
          onRemoveMember={handleRemoveMember}
          onResendInvitation={handleResendInvitation}
          onResetPassword={handleResetPassword}
        />
      </ComponentShowcase>

      <ComponentShowcase
        title="Table States"
        description="Different table states and variations"
      >
        <div className="space-y-6">
          <div>
            <p className="text-sm font-medium mb-3">Empty State</p>
            <UserTable
              members={[]}
              onRoleChange={handleRoleChange}
              onRemoveMember={handleRemoveMember}
            />
          </div>

          <div>
            <p className="text-sm font-medium mb-3">Single Row</p>
            <UserTable
              members={[mockTeamMembers[0]]}
              onRoleChange={handleRoleChange}
              onRemoveMember={handleRemoveMember}
            />
          </div>
        </div>
      </ComponentShowcase>
    </div>
  );
}