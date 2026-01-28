'use client'

import { KloudliteLogo } from '@/components/kloudlite-logo'
import { OAuthButtons } from '@/components/console/oauth-buttons'
import { AlertCircle } from 'lucide-react'
import { cn } from '@kloudlite/lib'

// Cross marker component for grid
function CrossMarker({ className }: { className?: string }) {
  return (
    <div className={cn('absolute', className)}>
      <div className="absolute left-1/2 top-0 -translate-x-1/2 w-px h-5 bg-foreground/20" />
      <div className="absolute top-1/2 left-0 -translate-y-1/2 h-px w-5 bg-foreground/20" />
    </div>
  )
}

export function LoginCard({ errorMessage }: { errorMessage: string | null }) {
  return (
    <>
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

      {/* Login card with animated border */}
      <div className="relative mx-auto w-full max-w-[480px] border border-border/80 shadow-xl shadow-foreground/[0.02]">
        {/* Animated border pulses */}
        <div className="absolute inset-0 pointer-events-none overflow-hidden">
          <div className="pulse-top absolute top-0 w-12 h-px bg-gradient-to-r from-transparent via-primary to-transparent" />
          <div className="pulse-right absolute right-0 h-12 w-px bg-gradient-to-b from-transparent via-primary to-transparent" />
          <div className="pulse-bottom absolute bottom-0 w-12 h-px bg-gradient-to-r from-transparent via-primary to-transparent" />
          <div className="pulse-left absolute left-0 h-12 w-px bg-gradient-to-b from-transparent via-primary to-transparent" />
        </div>

        {/* Corner markers */}
        <div className="absolute inset-0 pointer-events-none">
          <CrossMarker className="top-0 left-0 -translate-x-1/2 -translate-y-1/2 w-5 h-5" />
          <CrossMarker className="top-0 right-0 translate-x-1/2 -translate-y-1/2 w-5 h-5" />
          <CrossMarker className="bottom-0 left-0 -translate-x-1/2 translate-y-1/2 w-5 h-5" />
          <CrossMarker className="bottom-0 right-0 translate-x-1/2 translate-y-1/2 w-5 h-5" />
        </div>

        {/* Content */}
        <div className="relative bg-background p-8 sm:p-12 lg:p-16 space-y-10">
          {/* Logo */}
          <div className="flex justify-center">
            <KloudliteLogo className="scale-125 transition-transform hover:scale-[1.35]" />
          </div>

          {/* Heading */}
          <div className="space-y-3 text-center">
            <h1 className="text-foreground text-3xl sm:text-4xl font-bold tracking-tight">
              Welcome back
            </h1>
            <p className="text-muted-foreground text-base">
              Sign in to your installation console
            </p>
          </div>

          {/* Error message */}
          {errorMessage && (
            <div className="border-destructive bg-destructive/10 text-destructive flex items-start gap-3 border p-4 text-sm">
              <AlertCircle className="mt-0.5 h-5 w-5 flex-shrink-0" />
              <div>{errorMessage}</div>
            </div>
          )}

          {/* OAuth buttons */}
          <div className="space-y-3">
            <OAuthButtons />
          </div>

          {/* Divider */}
          <div className="relative py-2">
            <div className="absolute inset-0 flex items-center">
              <span className="border-border/50 w-full border-t" />
            </div>
            <div className="relative flex justify-center">
              <span className="bg-background text-muted-foreground/70 px-4 text-xs font-medium uppercase tracking-wider">
                New to Kloudlite?
              </span>
            </div>
          </div>

          {/* Features section - single line, centered */}
          <div className="text-center space-y-5">
            <p className="text-muted-foreground text-sm leading-relaxed">
              Get started with a free account and manage your cloud infrastructure.
            </p>
            <div className="flex flex-wrap items-center justify-center gap-x-8 gap-y-2.5">
              <div className="flex items-center gap-2 text-muted-foreground text-xs">
                <svg
                  className="text-success h-4 w-4"
                  fill="none"
                  viewBox="0 0 24 24"
                  stroke="currentColor"
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={2.5}
                    d="M5 13l4 4L19 7"
                  />
                </svg>
                <span>No credit card</span>
              </div>
              <div className="flex items-center gap-2 text-muted-foreground text-xs">
                <svg
                  className="text-success h-4 w-4"
                  fill="none"
                  viewBox="0 0 24 24"
                  stroke="currentColor"
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={2.5}
                    d="M5 13l4 4L19 7"
                  />
                </svg>
                <span>Free forever</span>
              </div>
              <div className="flex items-center gap-2 text-muted-foreground text-xs">
                <svg
                  className="text-success h-4 w-4"
                  fill="none"
                  viewBox="0 0 24 24"
                  stroke="currentColor"
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={2.5}
                    d="M5 13l4 4L19 7"
                  />
                </svg>
                <span>Deploy in minutes</span>
              </div>
              <div className="flex items-center gap-2 text-muted-foreground text-xs">
                <svg
                  className="text-success h-4 w-4"
                  fill="none"
                  viewBox="0 0 24 24"
                  stroke="currentColor"
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={2.5}
                    d="M5 13l4 4L19 7"
                  />
                </svg>
                <span>Enterprise security</span>
              </div>
            </div>
          </div>

          {/* Development backdoor */}
          {process.env.NODE_ENV !== 'production' && (
            <div className="border-t border-border/50 pt-6">
              <a
                href="/api/dev-login"
                className="text-muted-foreground/50 hover:text-foreground text-xs text-center block transition-colors"
              >
                [Dev] Quick login as karthik@kloudlite.io
              </a>
            </div>
          )}
        </div>
      </div>
    </>
  )
}
