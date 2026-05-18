'use client'

import Link from 'next/link'
import { AlertCircle, Shield } from 'lucide-react'
import { Alert, AlertDescription, Button, KloudliteLogo } from '@kloudlite/ui'
import { cn } from '@kloudlite/lib'

interface SignInFormProps {
  allowDevSuperAdmin: boolean
}

function CrossMarker({ className }: { className?: string }) {
  return (
    <div className={cn('absolute', className)}>
      <div className="absolute left-1/2 top-0 h-5 w-px -translate-x-1/2 bg-foreground/20" />
      <div className="absolute left-0 top-1/2 h-px w-5 -translate-y-1/2 bg-foreground/20" />
    </div>
  )
}

export function SignInForm({ allowDevSuperAdmin }: SignInFormProps) {
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
        .pulse-top { animation: pulseTopLeftToRight 4s ease-in-out infinite; }
        .pulse-right { animation: pulseRightTopToBottom 4s ease-in-out infinite 1s; }
        .pulse-bottom { animation: pulseBottomRightToLeft 4s ease-in-out infinite 2s; }
        .pulse-left { animation: pulseLeftBottomToTop 4s ease-in-out infinite 3s; }
      `}</style>

      <div className="relative mx-auto w-full max-w-[480px] border border-border/80">
        <div className="pointer-events-none absolute inset-0 overflow-hidden">
          <div className="pulse-top absolute top-0 h-px w-12 bg-gradient-to-r from-transparent via-primary to-transparent" />
          <div className="pulse-right absolute right-0 h-12 w-px bg-gradient-to-b from-transparent via-primary to-transparent" />
          <div className="pulse-bottom absolute bottom-0 h-px w-12 bg-gradient-to-r from-transparent via-primary to-transparent" />
          <div className="pulse-left absolute left-0 h-12 w-px bg-gradient-to-b from-transparent via-primary to-transparent" />
        </div>

        <div className="pointer-events-none absolute inset-0">
          <CrossMarker className="left-0 top-0 h-5 w-5 -translate-x-1/2 -translate-y-1/2" />
          <CrossMarker className="right-0 top-0 h-5 w-5 -translate-y-1/2 translate-x-1/2" />
          <CrossMarker className="bottom-0 left-0 h-5 w-5 -translate-x-1/2 translate-y-1/2" />
          <CrossMarker className="bottom-0 right-0 h-5 w-5 translate-x-1/2 translate-y-1/2" />
        </div>

        <div className="relative space-y-10 bg-background p-8 sm:p-12 lg:p-16">
          <div className="flex justify-center">
            <KloudliteLogo className="scale-125 transition-transform hover:scale-[1.35]" />
          </div>

          <div className="space-y-3 text-center">
            <h1 className="text-3xl font-bold tracking-tight text-foreground sm:text-4xl">
              Super Admin Login
            </h1>
            <p className="text-base text-muted-foreground">
              Use the login URL generated from console to access the admin dashboard.
            </p>
          </div>

          <Alert>
            <AlertCircle className="h-4 w-4" />
            <AlertDescription>
              Regular user login has moved out of this dashboard. This app is now admin-only.
            </AlertDescription>
          </Alert>

          {allowDevSuperAdmin && (
            <Button asChild size="lg" className="w-full gap-3 text-base font-medium">
              <Link href="/superadmin-login?token=dev-superadmin">
                <Shield className="h-5 w-5" />
                Continue as Dev Super Admin
              </Link>
            </Button>
          )}
        </div>
      </div>
    </>
  )
}
