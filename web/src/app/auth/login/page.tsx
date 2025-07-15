import { AuthCard } from '@/components/auth/auth-card'
import { LoginForm } from '@/components/auth/login-form'

export default function LoginPage() {
  return (
    <AuthCard
      title="Welcome back"
      description="Enter your email to sign in to your account"
    >
      <LoginForm />
    </AuthCard>
  )
}