import { AuthCard } from '@/components/auth/auth-card'
import { LoginForm } from '@/components/auth/login-form'
import { LogIn } from 'lucide-react'

export default function LoginPage() {
  return (
    <AuthCard
      title="Sign in"
      description="Enter your email to sign in to your account"
      icon={LogIn}
    >
      <LoginForm />
    </AuthCard>
  )
}