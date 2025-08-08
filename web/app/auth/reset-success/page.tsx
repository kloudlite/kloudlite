import { CheckCircle } from "lucide-react"
import { type Metadata } from "next"
import Link from "next/link"

import { Button } from "@/components/ui/button"
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"

export const metadata: Metadata = {
  title: "Password Reset Successful",
  description: "Your password has been successfully reset.",
}

export default function ResetSuccessPage() {
  return (
    <Card>
      <CardHeader className="text-center">
        <div className="mx-auto mb-4 flex h-14 w-14 items-center justify-center rounded-full bg-primary/10">
          <CheckCircle className="h-6 w-6 text-primary" />
        </div>
        <CardTitle className="text-xl">Password reset successful</CardTitle>
        <CardDescription>
          Your password has been successfully updated
        </CardDescription>
      </CardHeader>
      <CardContent>
        <div className="grid gap-6">
          <Button asChild className="w-full">
            <Link href="/auth/login">
              Back to login
            </Link>
          </Button>
        </div>
      </CardContent>
    </Card>
  )
}