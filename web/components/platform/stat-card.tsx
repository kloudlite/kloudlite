import { type LucideIcon } from "lucide-react";

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { cn } from "@/lib/utils";

interface StatCardProps {
  title: string;
  value: string | number;
  icon: LucideIcon;
  color?: "primary" | "orange" | "blue" | "green";
}

const colorMap = {
  primary: "border-l-primary text-primary",
  orange: "border-l-orange-500 text-orange-500",
  blue: "border-l-blue-500 text-blue-500",
  green: "border-l-green-500 text-green-500",
} as const;

export function StatCard({ title, value, icon: Icon, color = "primary" }: StatCardProps) {
  return (
    <Card className={cn(
      "border-l-4 transition-all duration-200 hover:shadow-lg", 
      colorMap[color].split(" ")[0]
    )}>
      <CardHeader className="pb-3">
        <CardTitle className="text-sm font-medium text-muted-foreground">
          {title}
        </CardTitle>
      </CardHeader>
      <CardContent>
        <div className="flex items-center justify-between">
          <p className="text-2xl font-semibold">{value}</p>
          <Icon className={cn("h-5 w-5 transition-colors duration-200", colorMap[color].split(" ")[1])} />
        </div>
      </CardContent>
    </Card>
  );
}