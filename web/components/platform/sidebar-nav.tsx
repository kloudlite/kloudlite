"use client";

import { type LucideIcon } from "lucide-react";

import { Badge } from "@/components/ui/badge";
import { cn } from "@/lib/utils";

interface NavItem {
  id: string;
  label: string;
  icon: LucideIcon;
  badge?: number;
}

interface SidebarNavProps {
  items: NavItem[];
  activeItem: string;
  onItemClick: (id: string) => void;
}

export function SidebarNav({ items, activeItem, onItemClick }: SidebarNavProps) {
  return (
    <nav className="space-y-1 px-1">
      {items.map((item) => {
        const Icon = item.icon;
        const isActive = activeItem === item.id;
        
        return (
          <button
            key={item.id}
            onClick={() => onItemClick(item.id)}
            className={cn(
              "relative flex w-full items-center gap-2.5 rounded-md px-3 py-2 text-[13px] font-medium border",
              "transition-all duration-200 ease-in-out",
              "focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary focus-visible:ring-offset-1",
              "ring-offset-background",
              "before:absolute before:left-0 before:top-1/2 before:h-0 before:w-0.5 before:-translate-y-1/2 before:bg-primary before:transition-all before:duration-200 before:ease-in-out",
              isActive
                ? "bg-background text-foreground border-border before:h-4"
                : "text-muted-foreground hover:bg-background/50 hover:text-foreground hover:before:h-4 border-transparent"
            )}
          >
            <Icon className="h-4 w-4 shrink-0" />
            <span className="flex-1 text-left">{item.label}</span>
            {item.badge !== undefined && item.badge > 0 && (
              <Badge variant="destructive" className="h-5 min-w-[20px] px-1 text-xs">
                {item.badge}
              </Badge>
            )}
          </button>
        );
      })}
    </nav>
  );
}