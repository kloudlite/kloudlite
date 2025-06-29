import { IconBox } from "@/components/atoms";
import { 
  Activity, 
  AlertCircle, 
  CheckCircle, 
  Database, 
  GitBranch, 
  Layers,
  Server,
  Settings,
  Terminal,
  Users,
  XCircle
} from "lucide-react";
import { ComponentShowcase } from "../../_components/component-showcase";

export default function IconsPage() {
  return (
    <div className="space-y-8">
      <div>
        <h1 className="text-2xl font-bold text-slate-900 dark:text-white mb-4">
          Icons
        </h1>
        <p className="text-slate-600 dark:text-slate-400">
          Icon components and icon boxes for consistent icon display.
        </p>
      </div>

      <ComponentShowcase
        title="Icon Box Sizes"
        description="Different sizes for icon containers"
      >
        <div className="flex flex-wrap items-end gap-6">
          {[
            { size: "xs", label: "Extra Small", pixels: "24px" },
            { size: "sm", label: "Small", pixels: "32px" },
            { size: "md", label: "Medium", pixels: "40px" },
            { size: "lg", label: "Large", pixels: "48px" },
            { size: "xl", label: "Extra Large", pixels: "56px" }
          ].map(({ size, label, pixels }) => (
            <div key={size} className="flex flex-col items-center gap-2">
              <IconBox icon={Terminal} size={size as any} />
              <div className="text-center">
                <p className="text-sm font-medium text-slate-900 dark:text-white">{size}</p>
                <p className="text-xs text-slate-600 dark:text-slate-400">{label}</p>
                <p className="text-xs text-slate-500 dark:text-slate-500 font-mono">{pixels}</p>
              </div>
            </div>
          ))}
        </div>
      </ComponentShowcase>

      <ComponentShowcase
        title="Icon Box Colors"
        description="All available color variants"
      >
        <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-6 gap-4">
          {[
            { color: "blue", icon: Database, label: "Blue" },
            { color: "green", icon: CheckCircle, label: "Green" },
            { color: "yellow", icon: AlertCircle, label: "Yellow" },
            { color: "red", icon: XCircle, label: "Red" },
            { color: "purple", icon: Settings, label: "Purple" },
            { color: "orange", icon: Activity, label: "Orange" },
            { color: "pink", icon: Users, label: "Pink" },
            { color: "indigo", icon: Layers, label: "Indigo" },
            { color: "slate", icon: Server, label: "Slate" },
            { color: "gradient", icon: GitBranch, label: "Gradient" },
            { color: "muted", icon: Terminal, label: "Muted" },
          ].map(({ color, icon, label }) => (
            <div key={color} className="flex flex-col items-center gap-2">
              <IconBox icon={icon} color={color as any} size="lg" />
              <p className="text-sm text-slate-700 dark:text-slate-300">{label}</p>
            </div>
          ))}
        </div>
      </ComponentShowcase>

      <ComponentShowcase
        title="Size and Color Combinations"
        description="All sizes with different color variants"
      >
        <div className="space-y-6">
          {["xs", "sm", "md", "lg", "xl"].map((size) => (
            <div key={size}>
              <p className="text-sm font-medium mb-3 text-slate-700 dark:text-slate-300">
                Size: {size}
              </p>
              <div className="flex flex-wrap items-center gap-3">
                <IconBox icon={Database} size={size as any} color="blue" />
                <IconBox icon={CheckCircle} size={size as any} color="green" />
                <IconBox icon={AlertCircle} size={size as any} color="yellow" />
                <IconBox icon={XCircle} size={size as any} color="red" />
                <IconBox icon={Settings} size={size as any} color="purple" />
                <IconBox icon={Activity} size={size as any} color="orange" />
                <IconBox icon={Users} size={size as any} color="pink" />
                <IconBox icon={Layers} size={size as any} color="indigo" />
                <IconBox icon={Server} size={size as any} color="slate" />
                <IconBox icon={GitBranch} size={size as any} color="gradient" />
                <IconBox icon={Terminal} size={size as any} color="muted" />
              </div>
            </div>
          ))}
        </div>
      </ComponentShowcase>

      <ComponentShowcase
        title="Common Icons"
        description="Frequently used icons in the system"
      >
        <div className="grid grid-cols-4 md:grid-cols-6 gap-4">
          {[
            { icon: Activity, label: "Activity" },
            { icon: AlertCircle, label: "Alert" },
            { icon: CheckCircle, label: "Success" },
            { icon: Database, label: "Database" },
            { icon: GitBranch, label: "Branch" },
            { icon: Layers, label: "Layers" },
            { icon: Server, label: "Server" },
            { icon: Settings, label: "Settings" },
            { icon: Terminal, label: "Terminal" },
            { icon: Users, label: "Users" },
            { icon: XCircle, label: "Error" },
          ].map(({ icon: Icon, label }) => (
            <div key={label} className="text-center">
              <div className="flex justify-center mb-2">
                <Icon className="h-6 w-6 text-slate-600 dark:text-slate-400" />
              </div>
              <p className="text-xs text-slate-600 dark:text-slate-400">{label}</p>
            </div>
          ))}
        </div>
      </ComponentShowcase>
    </div>
  );
}