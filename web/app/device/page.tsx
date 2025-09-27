"use client"

import { useState, useEffect } from "react"

import { CheckCircle } from "lucide-react"
import { useSearchParams, useRouter } from "next/navigation"
import { useSession } from "next-auth/react"

import { verifyDeviceCodeAction } from "@/app/actions/auth"
import { Alert, AlertDescription } from "@/components/ui/alert"
import { Button } from "@/components/ui/button"
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"


export default function DevicePage() {
  const searchParams = useSearchParams()
  const router = useRouter()
  const { data: session } = useSession()
  const [code, setCode] = useState(searchParams.get("code") || "")
  const [error, setError] = useState("")
  const [success, setSuccess] = useState(false)
  const [loading, setLoading] = useState(false)

  useEffect(() => {
    // Auto-verify if code is in URL and user is logged in
    if (searchParams.get("code") && session?.user) {
      handleVerify()
    }
  }, [searchParams, session])

  const handleVerify = async () => {
    if (!session?.user) {
      setError("You must be logged in to authorize a device")
      return
    }

    if (!code || code.length !== 8) {
      setError("Please enter a valid 8-character code")
      return
    }

    setError("")
    setLoading(true)

    try {
      const result = await verifyDeviceCodeAction({
        userCode: code.toUpperCase(),
        userId: session.user.id
      })

      if (result.success) {
        setSuccess(true)
      } else {
        setError(result.error || "Failed to verify device code")
      }
    } catch (error: any) {
      setError(error?.message || "An error occurred. Please try again.")
    } finally {
      setLoading(false)
    }
  }

  if (!session) {
    return (
      <div className="flex min-h-screen items-center justify-center">
        <Card className="w-full max-w-md">
          <CardHeader className="text-center">
            <CardTitle>Device Authorization</CardTitle>
            <CardDescription>
              Please log in to authorize your device
            </CardDescription>
          </CardHeader>
          <CardContent>
            <Button 
              className="w-full" 
              onClick={() => {
                const codeParam = searchParams.get("code");
                const callbackUrl = codeParam ? `/device?code=${codeParam}` : '/device';
                router.push(`/auth/login?callbackUrl=${encodeURIComponent(callbackUrl)}`);
              }}
            >
              Log In
            </Button>
          </CardContent>
        </Card>
      </div>
    )
  }

  if (success) {
    return (
      <div className="flex min-h-screen items-center justify-center">
        <Card className="w-full max-w-md">
          <CardHeader className="text-center">
            <CardTitle className="flex items-center justify-center gap-2">
              <CheckCircle className="h-6 w-6 text-green-600" />
              Device Authorized
            </CardTitle>
            <CardDescription>
              Your device has been successfully authorized. You can now return to your CLI.
            </CardDescription>
          </CardHeader>
        </Card>
      </div>
    )
  }

  return (
    <div className="flex min-h-screen items-center justify-center">
      <Card className="w-full max-w-md">
        <CardHeader className="text-center">
          <CardTitle>Authorize Device</CardTitle>
          <CardDescription>
            Enter the code shown in your CLI to authorize the device
          </CardDescription>
        </CardHeader>
        <CardContent>
          <form onSubmit={(e) => {
            e.preventDefault()
            handleVerify()
          }}>
            <div className="grid gap-6">
              {error && (
                <Alert variant="destructive" className="animate-in fade-in-0 slide-in-from-top-1">
                  <AlertDescription>{error}</AlertDescription>
                </Alert>
              )}
              
              <div className="grid gap-3">
                <Label htmlFor="code">Device Code</Label>
                <Input
                  id="code"
                  type="text"
                  placeholder="XXXX-XXXX"
                  value={code}
                  onChange={(e) => setCode(e.target.value.toUpperCase())}
                  maxLength={8}
                  className="text-center text-2xl tracking-widest font-mono"
                  autoFocus
                  required
                />
                <p className="text-sm text-muted-foreground text-center">
                  Enter the 8-character code from your CLI
                </p>
              </div>

              <Button type="submit" className="w-full" disabled={loading}>
                {loading ? "Authorizing..." : "Authorize Device"}
              </Button>
            </div>
          </form>
        </CardContent>
      </Card>
    </div>
  )
}