import { BaseCard } from "@/components/molecules";
import { cn } from "@/lib/utils";

interface ComponentShowcaseProps {
  title: string;
  description?: string;
  children: React.ReactNode;
  className?: string;
  contentClassName?: string;
  padding?: "none" | "sm" | "default" | "lg";
}

export function ComponentShowcase({
  title,
  description,
  children,
  className,
  contentClassName,
  padding = "default",
}: ComponentShowcaseProps) {
  return (
    <BaseCard 
      className={className}
      padding={padding}
    >
      <div className="space-y-4">
        <div>
          <h3 className="text-lg font-semibold text-foreground">{title}</h3>
          {description && (
            <p className="text-sm text-muted-foreground mt-1">
              {description}
            </p>
          )}
        </div>
        <div className={contentClassName}>
          {children}
        </div>
      </div>
    </BaseCard>
  );
}