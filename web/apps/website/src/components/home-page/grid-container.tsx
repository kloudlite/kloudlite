import type React from 'react'
import { cn } from '@kloudlite/lib'
import { CrossMarker } from './cross-marker'

interface GridContainerProps {
  children: React.ReactNode
  className?: string
}

export function GridContainer({ children, className }: GridContainerProps) {
  return (
    <div className={cn('relative mx-auto max-w-7xl', className)}>
      <style jsx>{`
        @keyframes pulseTopLeftToRight {
          0% { left: 0%; opacity: 0; }
          10% { opacity: 1; }
          90% { opacity: 1; }
          100% { left: 100%; opacity: 0; }
        }
        @keyframes pulseRightTopToBottom {
          0% { top: 0%; opacity: 0; }
          10% { opacity: 1; }
          90% { opacity: 1; }
          100% { top: 100%; opacity: 0; }
        }
        @keyframes pulseBottomRightToLeft {
          0% { right: 0%; opacity: 0; }
          10% { opacity: 1; }
          90% { opacity: 1; }
          100% { right: 100%; opacity: 0; }
        }
        @keyframes pulseLeftBottomToTop {
          0% { bottom: 0%; opacity: 0; }
          10% { opacity: 1; }
          90% { opacity: 1; }
          100% { bottom: 100%; opacity: 0; }
        }
        .pulse-top {
          animation: pulseTopLeftToRight 4s ease-in-out infinite;
        }
        .pulse-right {
          animation: pulseRightTopToBottom 4s ease-in-out infinite 1s;
        }
        .pulse-bottom {
          animation: pulseBottomRightToLeft 4s ease-in-out infinite 2s;
        }
        .pulse-left {
          animation: pulseLeftBottomToTop 4s ease-in-out infinite 3s;
        }
      `}</style>

      <div className="absolute inset-0 pointer-events-none overflow-visible">
        <div className="absolute inset-y-0 left-0 w-px bg-foreground/10" />
        <div className="absolute inset-y-0 right-0 w-px bg-foreground/10" />
        <div className="absolute inset-x-0 top-0 h-px bg-foreground/10" />
        <div className="absolute inset-x-0 bottom-0 h-px bg-foreground/10" />
        <div className="pulse-top absolute top-0 h-px w-12 bg-gradient-to-r from-transparent via-primary to-transparent" />
        <div className="pulse-right absolute right-0 h-12 w-px bg-gradient-to-b from-transparent via-primary to-transparent" />
        <div className="pulse-bottom absolute bottom-0 h-px w-12 bg-gradient-to-r from-transparent via-primary to-transparent" />
        <div className="pulse-left absolute left-0 h-12 w-px bg-gradient-to-b from-transparent via-primary to-transparent" />
        <CrossMarker className="-translate-x-1/2 -translate-y-1/2 left-0 top-0 h-5 w-5" />
        <CrossMarker className="-translate-y-1/2 right-0 top-0 h-5 w-5 translate-x-1/2" />
        <CrossMarker className="-translate-x-1/2 bottom-0 left-0 h-5 w-5 translate-y-1/2" />
        <CrossMarker className="bottom-0 right-0 h-5 w-5 translate-x-1/2 translate-y-1/2" />
      </div>

      <div className="relative">{children}</div>
    </div>
  )
}
