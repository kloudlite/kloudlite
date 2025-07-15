import { AuthCard } from '@/components/auth/auth-card'
import { VerifyEmail } from '@/components/auth/verify-email'
import { MailCheck } from 'lucide-react'

export default function VerifyEmailPage() {
  return (
    <AuthCard
      title="Verify your email"
      description="We need to verify your email address to activate your account"
      icon={MailCheck}
    >
      <VerifyEmail />
    </AuthCard>
  )
}