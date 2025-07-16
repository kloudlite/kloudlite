import { AuthCard } from '@/components/auth/auth-card'
import { ResetPasswordForm } from '@/components/auth/reset-password-form'
import { KeyRound } from 'lucide-react'

export default function ResetPasswordPage() {
  return (
    <AuthCard
      title="Reset your password"
      description="Enter your new password below"
      icon={KeyRound}
    >
      <ResetPasswordForm />
    </AuthCard>
  )
}