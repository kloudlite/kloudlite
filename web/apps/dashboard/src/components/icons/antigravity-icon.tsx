import React from 'react'

interface AntigravityIconProps {
  className?: string
}

export function AntigravityIcon({ className = 'h-4 w-4' }: AntigravityIconProps) {
  return (
    <svg
      viewBox="0 0 100 100"
      className={className}
      xmlns="http://www.w3.org/2000/svg"
    >
      <defs>
        <linearGradient id="antigravity-gradient" x1="0%" y1="100%" x2="100%" y2="0%">
          <stop offset="0%" stopColor="#4A90D9" />
          <stop offset="40%" stopColor="#5B8ED9" />
          <stop offset="60%" stopColor="#E67E4A" />
          <stop offset="80%" stopColor="#F5A623" />
          <stop offset="100%" stopColor="#7ED321" />
        </linearGradient>
      </defs>
      <path
        d="M50 5 C35 5, 20 30, 5 95 C8 95, 15 95, 20 90 C30 60, 40 40, 50 30 C60 40, 70 60, 80 90 C85 95, 92 95, 95 95 C80 30, 65 5, 50 5 Z"
        fill="url(#antigravity-gradient)"
      />
    </svg>
  )
}
