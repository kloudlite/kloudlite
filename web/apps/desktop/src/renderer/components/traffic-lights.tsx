import { useState } from 'react'
import { cn } from '@/lib/utils'

interface TrafficLightsProps {
  className?: string
}

export function TrafficLights({ className }: TrafficLightsProps) {
  const [hovered, setHovered] = useState(false)

  return (
    <div
      className={cn('no-drag flex items-center gap-[7px]', className)}
      onMouseEnter={() => setHovered(true)}
      onMouseLeave={() => setHovered(false)}
    >
      <button
        className="flex h-[12px] w-[12px] items-center justify-center rounded-full bg-[#ff5f57]"
        onClick={() => window.electronAPI.windowControl('close')}
      >
        {hovered && (
          <svg width="6" height="6" viewBox="0 0 6 6">
            <path d="M0.5 0.5L5.5 5.5M5.5 0.5L0.5 5.5" stroke="#4a0002" strokeWidth="1.1" strokeLinecap="round" />
          </svg>
        )}
      </button>
      <button
        className="flex h-[12px] w-[12px] items-center justify-center rounded-full bg-[#febc2e]"
        onClick={() => window.electronAPI.windowControl('minimize')}
      >
        {hovered && (
          <svg width="6" height="1" viewBox="0 0 6 1">
            <path d="M0.5 0.5H5.5" stroke="#995700" strokeWidth="1.1" strokeLinecap="round" />
          </svg>
        )}
      </button>
      <button
        className="flex h-[12px] w-[12px] items-center justify-center rounded-full bg-[#28c840]"
        onClick={() => window.electronAPI.windowControl('maximize')}
      >
        {hovered && (
          <svg width="6" height="6" viewBox="0 0 8 8">
            <path d="M1.5 3L4 0.5L6.5 3M1.5 5L4 7.5L6.5 5" stroke="#006500" strokeWidth="1.1" strokeLinecap="round" strokeLinejoin="round" fill="none" />
          </svg>
        )}
      </button>
    </div>
  )
}
