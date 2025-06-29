import React, { useId } from "react"
import { cn } from "@/lib/utils"

interface KloudliteLogoProps extends React.SVGProps<SVGSVGElement> {
  height?: number
  variant?: "color" | "white" | "black" | "primary"
}

export function KloudliteLogo({ height = 24, variant = "color", className, ...props }: KloudliteLogoProps) {
  // Calculate width based on icon aspect ratio (approximately square)
  const width = height * 1.2
  
  // Use React's useId hook for stable IDs
  const id = useId()
  const maskId = `kloudlite-logo-mask-${id}`
  
  const getFillColor = () => {
    switch (variant) {
      case "white":
        return "white"
      case "black":
        return "black"
      case "primary":
        return "currentColor"
      case "color":
      default:
        return "#3b82f6" // blue-500
    }
  }
  
  return (
    <svg 
      height={height}
      width={width}
      viewBox="0 0 130 131" 
      fill="none" 
      xmlns="http://www.w3.org/2000/svg"
      className={cn("transition-colors", className)}
      {...props}
    >
      <defs>
        <mask id={maskId}>
          <path d="M51.9912 66.6496C51.2636 65.9244 51.2636 64.7486 51.9912 64.0235L89.4072 26.7312C90.1348 26.006 91.3145 26.006 92.042 26.7312L129.458 64.0237C130.186 64.7489 130.186 65.9246 129.458 66.6498L92.0423 103.942C91.3147 104.667 90.135 104.667 89.4074 103.942L51.9912 66.6496Z" fill="white"/>
          <path d="M66.5331 1.04291C65.8055 0.317729 64.6259 0.317729 63.8983 1.04291L0.545688 64.186C-0.181896 64.9111 -0.181896 66.0869 0.545688 66.8121L63.8983 129.955C64.6259 130.68 65.8055 130.68 66.5331 129.955L76.9755 119.547C77.7031 118.822 77.7031 117.646 76.9755 116.921L26.4574 66.5701C25.7298 65.8449 25.7298 64.6692 26.4574 63.944L76.7327 13.8349C77.4603 13.1097 77.4603 11.934 76.7327 11.2088L66.5331 1.04291Z" fill="white"/>
        </mask>
      </defs>
      <rect x="0" y="0" width="130" height="131" fill={getFillColor()} mask={`url(#${maskId})`}/>
    </svg>
  )
}