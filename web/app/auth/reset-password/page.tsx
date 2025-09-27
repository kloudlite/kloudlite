"use client"

import { useState, useEffect } from "react"

import { Loader2 } from "lucide-react"
import Link from "next/link"
import { useSearchParams, useRouter } from "next/navigation"
import { toast } from "sonner"

import { resetPasswordAction } from "@/app/actions/auth"
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



export default function ResetPasswordPage() {
  const searchParams = useSearchParams()
  const router = useRouter()
  const token = searchParams.get("token")
  
  const [password, setPassword] = useState("")
  const [confirmPassword, setConfirmPassword] = useState("")
  const [isLoading, setIsLoading] = useState(false)

  useEffect(() => {
    if (!token) {
      toast.error("Invalid reset link. Please request a new one.")
      router.push("/auth/forgot-password")
    }
  }, [token, router])

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    
    if (!password || !confirmPassword) {
      toast.error("Please fill in all fields")
      return
    }

    if (password !== confirmPassword) {
      toast.error("Passwords do not match")
      return
    }

    if (password.length < 8) {
      toast.error("Password must be at least 8 characters long")
      return
    }

    setIsLoading(true)

    try {
      const result = await resetPasswordAction(token!, password)
      
      if (result.success) {
        toast.success("Your password has been reset successfully")
        router.push("/auth/reset-success")
      } else {
        toast.error(result.error || "Failed to reset password")
      }
    } catch (error) {
      toast.error("An unexpected error occurred")
    } finally {
      setIsLoading(false)
    }
  }

  if (!token) {
    return null
  }

  return (
    <Card>
      <CardHeader className="text-center">
        <CardTitle className="text-xl">Reset your password</CardTitle>
        <CardDescription>
          Enter your new password below
        </CardDescription>
      </CardHeader>
      <CardContent>
        <form onSubmit={handleSubmit}>
          <div className="grid gap-6">
            <div className="grid gap-3">
              <Label htmlFor="password">New Password</Label>
              <Input
                id="password"
                type="password"
                placeholder="Enter new password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                disabled={isLoading}
                required
              />
            </div>
            
            <div className="grid gap-3">
              <Label htmlFor="confirm-password">Confirm Password</Label>
              <Input
                id="confirm-password"
                type="password"
                placeholder="Confirm new password"
                value={confirmPassword}
                onChange={(e) => setConfirmPassword(e.target.value)}
                disabled={isLoading}
                required
              />
            </div>
            
            <Button type="submit" className="w-full" disabled={isLoading}>
              {isLoading ? (
                <>
                  <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                  Resetting...
                </>
              ) : (
                "Reset Password"
              )}
            </Button>
            
            <div className="text-center text-sm">
              Remember your password?{" "}
              <Link href="/auth/login" className="underline underline-offset-4">
                Sign in
              </Link>
            </div>
          </div>
        </form>
      </CardContent>
    </Card>
  )
}