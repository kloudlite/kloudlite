"use client"

import * as React from "react"

import { cn } from "@/lib/utils"

interface TextareaProps extends React.ComponentProps<"textarea"> {
  maxLength?: number
  showCharacterCount?: boolean
  autoResize?: boolean
  minRows?: number
  maxRows?: number
}

const Textarea = React.forwardRef<HTMLTextAreaElement, TextareaProps>(
  ({ className, maxLength, showCharacterCount, autoResize, minRows = 3, maxRows = 10, onChange, value, defaultValue, ...props }, ref) => {
    const [internalValue, setInternalValue] = React.useState(defaultValue || "")
    const [textareaHeight, setTextareaHeight] = React.useState<string | undefined>(undefined)
    const textareaRef = React.useRef<HTMLTextAreaElement | null>(null)
    
    // Use controlled or uncontrolled value
    const currentValue = value !== undefined ? value : internalValue
    const characterCount = String(currentValue).length
    
    // Calculate line height for auto-resize
    const calculateHeight = React.useCallback(() => {
      const textarea = textareaRef.current
      if (!textarea || !autoResize) return
      
      // Reset height to get accurate scrollHeight
      textarea.style.height = 'auto'
      
      // Calculate line height if not already done
      const computedStyle = window.getComputedStyle(textarea)
      const lineHeight = parseInt(computedStyle.lineHeight)
      const paddingTop = parseInt(computedStyle.paddingTop)
      const paddingBottom = parseInt(computedStyle.paddingBottom)
      const borderTop = parseInt(computedStyle.borderTopWidth)
      const borderBottom = parseInt(computedStyle.borderBottomWidth)
      
      // Calculate min and max heights
      const minHeight = lineHeight * minRows + paddingTop + paddingBottom + borderTop + borderBottom
      const maxHeight = lineHeight * maxRows + paddingTop + paddingBottom + borderTop + borderBottom
      
      // Set new height within bounds
      const newHeight = Math.min(Math.max(textarea.scrollHeight, minHeight), maxHeight)
      setTextareaHeight(`${newHeight}px`)
    }, [autoResize, minRows, maxRows])
    
    // Effect to handle auto-resize
    React.useEffect(() => {
      calculateHeight()
    }, [currentValue, calculateHeight])
    
    // Handle window resize
    React.useEffect(() => {
      if (!autoResize) return
      
      const handleResize = () => calculateHeight()
      window.addEventListener('resize', handleResize)
      
      return () => window.removeEventListener('resize', handleResize)
    }, [autoResize, calculateHeight])
    
    // Handle change event
    const handleChange = React.useCallback((e: React.ChangeEvent<HTMLTextAreaElement>) => {
      const newValue = e.target.value
      
      // Respect maxLength if set
      if (maxLength && newValue.length > maxLength) {
        return
      }
      
      if (value === undefined) {
        setInternalValue(newValue)
      }
      
      onChange?.(e)
    }, [onChange, value, maxLength])
    
    // Combine refs
    const combinedRef = React.useCallback((node: HTMLTextAreaElement | null) => {
      textareaRef.current = node
      if (ref) {
        if (typeof ref === 'function') {
          ref(node)
        } else {
          ref.current = node
        }
      }
    }, [ref])
    
    return (
      <div className="relative w-full">
        <textarea
          ref={combinedRef}
          value={currentValue}
          onChange={handleChange}
          data-slot="textarea"
          className={cn(
            "border-input placeholder:text-muted-foreground dark:bg-input/30 flex w-full rounded-none border bg-transparent px-3 py-2 text-base shadow-xs transition-all duration-200 disabled:cursor-not-allowed disabled:opacity-50 md:text-sm",
            "focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 focus-visible:ring-offset-background",
            "aria-invalid:border-destructive aria-invalid:focus-visible:ring-destructive",
            autoResize && "resize-none overflow-hidden",
            !autoResize && "min-h-16 field-sizing-content",
            className
          )}
          style={autoResize ? { height: textareaHeight } : undefined}
          {...props}
        />
        {showCharacterCount && (
          <div className="absolute bottom-2 right-2 text-xs text-muted-foreground pointer-events-none">
            <span className={cn(
              "transition-colors duration-200",
              maxLength && characterCount > maxLength * 0.9 && "text-warning",
              maxLength && characterCount >= maxLength && "text-destructive"
            )}>
              {characterCount}
              {maxLength && `/${maxLength}`}
            </span>
          </div>
        )}
      </div>
    )
  }
)

Textarea.displayName = "Textarea"

export { Textarea }
