import { cn } from "@/lib/utils"

// ==================== Types ====================

interface {{COMPONENT_NAME}}Props {
  // Required props
  id: string
  title: string
  
  // Optional props
  description?: string
  className?: string
  children?: React.ReactNode
  
  // Event handlers
  onClick?: () => void
  onClose?: () => void
  
  // State props
  isLoading?: boolean
  isDisabled?: boolean
}

// ==================== Component ====================

export function {{COMPONENT_NAME}}({
  id,
  title,
  description,
  className,
  children,
  onClick,
  onClose,
  isLoading = false,
  isDisabled = false,
}: {{COMPONENT_NAME}}Props) {
  // Component logic here
  
  return (
    <div
      id={id}
      className={cn(
        "base-classes-here",
        isDisabled && "opacity-50 cursor-not-allowed",
        className
      )}
    >
      {/* Header Section */}
      <div className="flex items-center justify-between">
        <div className="space-y-1">
          <h3 className="text-lg font-medium">{title}</h3>
          {description && (
            <p className="text-sm text-muted-foreground">{description}</p>
          )}
        </div>
        
        {onClose && (
          <button
            onClick={onClose}
            disabled={isDisabled}
            className="rounded-md p-2 hover:bg-accent"
            aria-label="Close"
          >
            <X className="h-4 w-4" />
          </button>
        )}
      </div>

      {/* Content Section */}
      <div className="mt-4">
        {isLoading ? (
          <div className="flex items-center justify-center py-8">
            <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
          </div>
        ) : (
          children
        )}
      </div>

      {/* Action Section */}
      {onClick && (
        <div className="mt-6 flex justify-end">
          <Button
            onClick={onClick}
            disabled={isDisabled || isLoading}
          >
            {isLoading ? (
              <>
                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                Processing...
              </>
            ) : (
              'Continue'
            )}
          </Button>
        </div>
      )}
    </div>
  )
}

// ==================== Compound Components (Optional) ====================

{{COMPONENT_NAME}}.Header = function {{COMPONENT_NAME}}Header({ 
  children, 
  className 
}: { 
  children: React.ReactNode
  className?: string 
}) {
  return (
    <div className={cn("space-y-1", className)}>
      {children}
    </div>
  )
}

{{COMPONENT_NAME}}.Content = function {{COMPONENT_NAME}}Content({ 
  children, 
  className 
}: { 
  children: React.ReactNode
  className?: string 
}) {
  return (
    <div className={cn("mt-4", className)}>
      {children}
    </div>
  )
}

{{COMPONENT_NAME}}.Footer = function {{COMPONENT_NAME}}Footer({ 
  children, 
  className 
}: { 
  children: React.ReactNode
  className?: string 
}) {
  return (
    <div className={cn("mt-6 flex items-center justify-end gap-4", className)}>
      {children}
    </div>
  )
}

// ==================== Usage Example ====================

/*
// Basic usage
<{{COMPONENT_NAME}}
  id="example"
  title="Example Component"
  description="This is an example"
  onClick={() => console.log('clicked')}
>
  <p>Content goes here</p>
</{{COMPONENT_NAME}}>

// With compound components
<{{COMPONENT_NAME}} id="compound-example">
  <{{COMPONENT_NAME}}.Header>
    <h3>Custom Header</h3>
  </{{COMPONENT_NAME}}.Header>
  
  <{{COMPONENT_NAME}}.Content>
    <p>Custom content</p>
  </{{COMPONENT_NAME}}.Content>
  
  <{{COMPONENT_NAME}}.Footer>
    <Button variant="outline">Cancel</Button>
    <Button>Save</Button>
  </{{COMPONENT_NAME}}.Footer>
</{{COMPONENT_NAME}}>
*/