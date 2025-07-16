import { AuthCard } from '@/components/auth/auth-card'
import { ForgotPasswordForm } from '@/components/auth/forgot-password-form'
import { KeyRound } from 'lucide-react'

export default function ForgotPasswordPage() {
  return (
    <AuthCard
      title="Forgot your password?"
      description="Enter your email address and we'll send you a link to reset your password"
      icon={KeyRound}
    >
      <ForgotPasswordForm />
    </AuthCard>
  )
}