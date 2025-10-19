import React from 'react'

interface OpenCodeIconProps {
  className?: string
}

export function OpenCodeIcon({ className = "h-4 w-4" }: OpenCodeIconProps) {
  return (
    <svg
      viewBox="0 0 24 24"
      className={className}
      fill="none"
      xmlns="http://www.w3.org/2000/svg"
    >
      <rect
        x="2"
        y="2"
        width="20"
        height="20"
        rx="5"
        fill="currentColor"
      />
      <rect
        x="6"
        y="6"
        width="8"
        height="12"
        fill="white"
      />
      <rect
        x="9"
        y="9"
        width="6"
        height="6"
        fill="currentColor"
      />
    </svg>
  )
}
