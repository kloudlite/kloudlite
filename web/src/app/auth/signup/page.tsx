import { AuthCard } from '@/components/auth/auth-card'
import { SignupForm } from '@/components/auth/signup-form'
import { UserPlus } from 'lucide-react'

export default function SignupPage() {
  return (
    <AuthCard
      title="Create an account"
      description="Enter your information to create your account"
      icon={UserPlus}
    >
      <SignupForm />
    </AuthCard>
  )
}