import { Globe, Server, Clock } from "lucide-react"

import { Badge } from "@/components/ui/badge"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { cn } from "@/lib/utils"

import { ResourceCardActions } from "./resource-card-actions"

interface ResourceCardProps {
  type: "environment" | "workspace"
  name: string
  team: string
  status: "active" | "inactive" | "deploying"
  description?: string
  lastUpdated?: string
  onClick?: () => void
}

export function ResourceCard({ 
  type, 
  name, 
  team, 
  status, 
  description, 
  lastUpdated,
  onClick 
}: ResourceCardProps) {
  const Icon = type === "environment" ? Globe : Server
  
  const statusConfig = {
    active: {
      variant: "default" as const,
      className: "bg-green-50 text-green-700 border-green-200 dark:bg-green-950 dark:text-green-400 dark:border-green-800"
    },
    inactive: {
      variant: "secondary" as const,
      className: "bg-gray-50 text-gray-700 border-gray-200 dark:bg-gray-950 dark:text-gray-400 dark:border-gray-800"
    },
    deploying: {
      variant: "outline" as const,
      className: "bg-blue-50 text-blue-700 border-blue-200 dark:bg-blue-950 dark:text-blue-400 dark:border-blue-800"
    }
  }[status]

  return (
    <Card 
      className={cn(
        "group relative overflow-hidden transition-all duration-200",
        "hover:shadow-md hover:-translate-y-0.5",
        "active:translate-y-0 active:shadow-sm",
        "cursor-pointer border-gray-200 dark:border-gray-800"
      )} 
      onClick={(e) => {
        // Don't trigger card click if clicking on interactive elements
        if ((e.target as HTMLElement).closest('button, a, [role="menuitem"]')) {
          return;
        }
        onClick?.();
      }}
    >
      <CardHeader className="pb-3">
        <div className="flex items-start justify-between">
          <div className="flex items-start gap-3">
            <div className={cn(
              "mt-0.5 rounded-lg p-2",
              "bg-gray-100 dark:bg-gray-800",
              "group-hover:bg-gray-200 dark:group-hover:bg-gray-700",
              "transition-colors duration-200"
            )}>
              <Icon className="h-4 w-4 text-gray-600 dark:text-gray-400" />
            </div>
            <div className="space-y-1 flex-1 min-w-0">
              <CardTitle className="text-base font-medium leading-none truncate">
                {name}
              </CardTitle>
              <CardDescription className="text-sm flex items-center gap-2">
                <span className="truncate">{team}</span>
                <span className="text-gray-400 dark:text-gray-600">•</span>
                <span className="capitalize text-xs">{type}</span>
              </CardDescription>
            </div>
          </div>
          <div className="flex items-center gap-2">
            <Badge 
              variant={statusConfig.variant} 
              className={cn("text-xs capitalize", statusConfig.className)}
            >
              {status === "deploying" && (
                <span className="mr-1 h-1.5 w-1.5 rounded-full bg-current animate-pulse" />
              )}
              {status}
            </Badge>
            <ResourceCardActions type={type} />
          </div>
        </div>
      </CardHeader>
      {(description || lastUpdated) && (
        <CardContent className="pt-0">
          {description && (
            <p className="text-sm text-muted-foreground line-clamp-2 mb-3">
              {description}
            </p>
          )}
          {lastUpdated && (
            <div className="flex items-center gap-1.5 text-xs text-muted-foreground">
              <Clock className="h-3 w-3" />
              <span>{lastUpdated}</span>
            </div>
          )}
        </CardContent>
      )}
    </Card>
  )
}