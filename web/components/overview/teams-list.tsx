"use client";

import { useState, useMemo } from "react";

import { Search, Plus, Users } from "lucide-react";
import { useRouter } from "next/navigation";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Tabs, TabsList, TabsTrigger, TabsContent } from "@/components/ui/tabs";

import { TeamCard } from "./team-card";

interface Team {
  accountid: string;
  name: string;
  slug?: string;
  description?: string;
  status: string;
  role?: string;
  memberCount?: number;
  resourceCount?: number;
  createdAt?: string;
}

interface TeamsListProps {
  teams: Team[];
  canCreateTeam?: boolean;
}

export function TeamsList({ teams, canCreateTeam = true }: TeamsListProps) {
  const router = useRouter();
  const [searchQuery, setSearchQuery] = useState("");
  const [activeTab, setActiveTab] = useState("all");

  const filteredTeams = useMemo(() => {
    let filtered = teams;

    // Filter by search query
    if (searchQuery) {
      filtered = filtered.filter(team =>
        team.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
        team.description?.toLowerCase().includes(searchQuery.toLowerCase())
      );
    }

    // Filter by tab
    if (activeTab === "owned") {
      filtered = filtered.filter(team => team.role === "owner");
    } else if (activeTab === "member") {
      filtered = filtered.filter(team => team.role !== "owner");
    }

    return filtered;
  }, [teams, searchQuery, activeTab]);

  const stats = useMemo(() => ({
    all: teams.length,
    owned: teams.filter(t => t.role === "owner").length,
    member: teams.filter(t => t.role !== "owner").length,
  }), [teams]);

  return (
    <Tabs value={activeTab} onValueChange={setActiveTab} className="space-y-6">
      {/* Search, Tabs and Actions Row */}
      <div className="flex flex-col gap-4 lg:flex-row lg:items-center lg:justify-between">
        {/* Left side - Search and Tabs */}
        <div className="flex flex-col gap-4 sm:flex-row sm:items-center">
          <div className="relative w-full sm:w-auto sm:min-w-[300px]">
            <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
            <Input
              placeholder="Search teams..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="pl-9"
            />
          </div>
          
          <TabsList className="bg-muted/50 w-full sm:w-auto">
            <TabsTrigger value="all">
              All Teams
              <Badge variant="secondary" className="ml-2">
                {stats.all}
              </Badge>
            </TabsTrigger>
            <TabsTrigger value="owned">
              Owned
              <Badge variant="secondary" className="ml-2">
                {stats.owned}
              </Badge>
            </TabsTrigger>
            <TabsTrigger value="member">
              Member
              <Badge variant="secondary" className="ml-2">
                {stats.member}
              </Badge>
            </TabsTrigger>
          </TabsList>
        </div>

        {/* Right side - Create button */}
        {canCreateTeam && (
          <Button onClick={() => router.push("/teams/new")} className="gap-2 w-full sm:w-auto">
            <Plus className="h-4 w-4" />
            Create Team
          </Button>
        )}
      </div>

      {/* Tab Contents */}
      <TabsContent value={activeTab} className="mt-6">
        {filteredTeams.length === 0 ? (
          <div className="text-center py-12">
            <Users className="mx-auto h-12 w-12 text-muted-foreground/50" />
            <h3 className="mt-4 text-lg font-medium">No teams found</h3>
            <p className="mt-2 text-sm text-muted-foreground">
              {searchQuery 
                ? "Try adjusting your search query"
                : activeTab === "owned" 
                  ? "You don't own any teams yet"
                  : activeTab === "member"
                  ? "You're not a member of any teams"
                  : "Get started by creating your first team"
              }
            </p>
            {canCreateTeam && !searchQuery && (
              <Button onClick={() => router.push("/teams/new")} className="mt-4 gap-2">
                <Plus className="h-4 w-4" />
                Create Team
              </Button>
            )}
          </div>
        ) : (
          <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
            {filteredTeams.map((team) => (
              <TeamCard key={team.accountid} team={team} />
            ))}
          </div>
        )}
      </TabsContent>
    </Tabs>
  );
}