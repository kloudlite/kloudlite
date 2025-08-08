"use client";

import { Users, ChevronRight, Shield, Clock, CheckCircle } from "lucide-react";
import { useRouter } from "next/navigation";

import { Badge } from "@/components/ui/badge";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { cn } from "@/lib/utils";

interface TeamCardProps {
  team: {
    accountid: string;
    name: string;
    slug?: string;
    description?: string;
    status: string;
    role?: string;
    memberCount?: number;
    resourceCount?: number;
    createdAt?: string;
  };
}

export function TeamCard({ team }: TeamCardProps) {
  const router = useRouter();

  const handleClick = () => {
    // Only navigate if team is active
    if (team.status === "active") {
      // Use the actual slug if available, otherwise derive from name
      const teamSlug = team.slug || team.name.toLowerCase().replace(/\s+/g, '-').replace(/[^a-z0-9-]/g, '');
      router.push(`/${teamSlug}`);
    }
  };

  const getStatusIcon = () => {
    switch (team.status) {
      case "active":
        return <CheckCircle className="h-3.5 w-3.5" />;
      case "pending":
        return <Clock className="h-3.5 w-3.5" />;
      default:
        return null;
    }
  };

  const getStatusVariant = () => {
    switch (team.status) {
      case "active":
        return "default";
      case "pending":
        return "secondary";
      default:
        return "outline";
    }
  };

  const getRoleBadge = () => {
    if (!team.role) {return null;}
    
    switch (team.role) {
      case "owner":
        return (
          <Badge variant="default" className="gap-1">
            <Shield className="h-3 w-3" />
            Owner
          </Badge>
        );
      case "admin":
        return (
          <Badge variant="secondary" className="gap-1">
            <Shield className="h-3 w-3" />
            Admin
          </Badge>
        );
      default:
        return (
          <Badge variant="outline" className="capitalize">
            {team.role}
          </Badge>
        );
    }
  };

  return (
    <Card 
      className={cn(
        "group relative transition-all duration-200",
        team.status === "active" && "hover:shadow-md hover:border-primary/20 cursor-pointer",
        team.status !== "active" && "opacity-75"
      )}
      onClick={handleClick}
    >
      <CardHeader className="pb-4">
        <div className="flex items-start justify-between">
          <div className="space-y-1">
            <CardTitle className="text-lg font-medium group-hover:text-primary transition-colors">
              {team.name}
            </CardTitle>
            {team.description && (
              <CardDescription className="line-clamp-2">
                {team.description}
              </CardDescription>
            )}
          </div>
          <ChevronRight className="h-5 w-5 text-muted-foreground group-hover:text-primary transition-all group-hover:translate-x-0.5" />
        </div>
      </CardHeader>
      <CardContent>
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-4 text-sm text-muted-foreground">
            <div className="flex items-center gap-1.5">
              <Users className="h-4 w-4" />
              <span>{team.memberCount || 1} members</span>
            </div>
            {team.resourceCount !== undefined && (
              <div className="flex items-center gap-1.5">
                <span>{team.resourceCount} resources</span>
              </div>
            )}
          </div>
          <div className="flex items-center gap-2">
            {getRoleBadge()}
            <Badge variant={getStatusVariant()} className="gap-1">
              {getStatusIcon()}
              {team.status}
            </Badge>
          </div>
        </div>
      </CardContent>
    </Card>
  );
}