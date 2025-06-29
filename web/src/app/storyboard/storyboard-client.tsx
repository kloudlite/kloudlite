"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { cn } from "@/lib/utils";
import { useState } from "react";
import { 
  ChevronRight,
  ChevronDown,
  BookOpen, 
  Square,
  Circle,
  Triangle,
  Layout
} from "lucide-react";
import { KloudliteLogo, ThemeToggle } from "@/components/atoms";

interface StoryboardClientProps {
  children: React.ReactNode;
}

const navigation = [
  {
    title: "Overview",
    href: "/storyboard",
    icon: BookOpen,
  },
  {
    title: "Atoms",
    icon: Circle,
    items: [
      { label: "Tokens", href: "/storyboard/atoms/tokens" },
      { label: "Typography", href: "/storyboard/atoms/typography" },
      { label: "Buttons", href: "/storyboard/atoms/buttons" },
      { label: "Inputs", href: "/storyboard/atoms/inputs" },
      { label: "Date Picker", href: "/storyboard/atoms/date-picker" },
      { label: "Badges", href: "/storyboard/atoms/badges" },
      { label: "Avatars", href: "/storyboard/atoms/avatars" },
      { label: "Breadcrumb", href: "/storyboard/atoms/breadcrumb" },
      { label: "Card", href: "/storyboard/atoms/card" },
      { label: "Icons", href: "/storyboard/atoms/icons" },
      { label: "Loading", href: "/storyboard/atoms/loading" },
      { label: "Separator", href: "/storyboard/atoms/separator" },
      { label: "Skeletons", href: "/storyboard/atoms/skeletons" },
      { label: "Stat Displays", href: "/storyboard/atoms/stat-displays" },
      { label: "Tables", href: "/storyboard/atoms/tables" },
      { label: "Tooltip", href: "/storyboard/atoms/tooltip" },
    ],
  },
  {
    title: "Molecules",
    icon: Triangle,
    items: [
      { label: "Cards", href: "/storyboard/molecules/cards" },
      { label: "Form Components", href: "/storyboard/molecules/forms" },
      { label: "Dialogs & Modals", href: "/storyboard/molecules/dialogs" },
      { label: "Dropdowns", href: "/storyboard/molecules/dropdowns" },
      { label: "Status Components", href: "/storyboard/molecules/status" },
      { label: "Data Display", href: "/storyboard/molecules/data-display" },
    ],
  },
  {
    title: "Organisms",
    icon: Square,
    items: [
      { label: "Navigation", href: "/storyboard/organisms/navigation" },
      { label: "Headers", href: "/storyboard/organisms/headers" },
      { label: "Tables", href: "/storyboard/organisms/tables" },
      { label: "Lists", href: "/storyboard/organisms/lists" },
      { label: "Footers", href: "/storyboard/organisms/footers" },
    ],
  },
  {
    title: "Templates",
    icon: Layout,
    items: [
      { label: "Page Layouts", href: "/storyboard/templates/layouts" },
      { label: "Auth Pages", href: "/storyboard/templates/auth" },
      { label: "Dashboard Pages", href: "/storyboard/templates/dashboard" },
    ],
  },
];

export default function StoryboardClient({ children }: StoryboardClientProps) {
  const pathname = usePathname();
  const [expandedSections, setExpandedSections] = useState<string[]>([]);

  const toggleSection = (title: string) => {
    setExpandedSections(prev => 
      prev.includes(title) 
        ? prev.filter(t => t !== title)
        : [...prev, title]
    );
  };

  return (
    <div className="flex h-screen bg-background">
      {/* Sidebar */}
      <div className="w-64 bg-card border-r border-border flex flex-col">
        <div className="flex-1 overflow-y-auto">
          <div className="p-6">
            <Link href="/storyboard" className="flex items-center gap-3 mb-8 group">
              <KloudliteLogo height={24} className="fill-foreground transition-transform group-hover:scale-110" />
              <span className="font-semibold text-foreground">Storyboard</span>
            </Link>

            <nav className="space-y-6">
              {navigation.map((section) => {
                const Icon = section.icon;
                const isActive = pathname === section.href;
                const hasActiveChild = section.items?.some(item => pathname === item.href);

                return (
                  <div key={section.title}>
                    {section.href ? (
                      <Link
                        href={section.href}
                        className={cn(
                          "flex items-center gap-3 px-3 py-2 rounded-lg text-sm font-medium transition-all mb-2",
                          isActive
                            ? "bg-primary/10 text-primary shadow-sm"
                            : "text-muted-foreground hover:bg-muted hover:text-foreground"
                        )}
                      >
                        <Icon className="h-4 w-4" />
                        {section.title}
                      </Link>
                    ) : (
                      <button
                        onClick={() => toggleSection(section.title)}
                        className="flex items-center justify-between w-full px-3 py-2 text-sm font-semibold text-foreground mb-2 hover:bg-muted rounded-lg transition-colors"
                      >
                        <div className="flex items-center gap-3">
                          <Icon className="h-4 w-4 text-muted-foreground" />
                          {section.title}
                        </div>
                        {section.items && (
                          expandedSections.includes(section.title) 
                            ? <ChevronDown className="h-4 w-4 text-muted-foreground" />
                            : <ChevronRight className="h-4 w-4 text-muted-foreground" />
                        )}
                      </button>
                    )}

                    {section.items && expandedSections.includes(section.title) && (
                      <div className="space-y-1 mb-4">
                        {section.items.map((item) => {
                          const isItemActive = pathname === item.href;
                          return (
                            <Link
                              key={item.href}
                              href={item.href}
                              className={cn(
                                "block px-3 py-1.5 rounded-md text-sm transition-all truncate ml-7",
                                isItemActive
                                  ? "bg-primary/10 text-primary font-medium"
                                  : "text-muted-foreground hover:bg-muted hover:text-foreground"
                              )}
                            >
                              {item.label}
                            </Link>
                          );
                        })}
                      </div>
                    )}
                  </div>
                );
              })}
            </nav>
          </div>
        </div>

        {/* Theme Toggle - Fixed at bottom */}
        <div className="p-6 pt-3 border-t border-border">
          <ThemeToggle variant="sidebar" className="w-full justify-start" />
        </div>
      </div>

      {/* Main Content */}
      <div className="flex-1 overflow-auto bg-muted/20">
        <div className="container mx-auto p-8 max-w-7xl">
          {children}
        </div>
      </div>
    </div>
  );
}