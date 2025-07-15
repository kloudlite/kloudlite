<<<<<<< HEAD
import { notFound } from "next/navigation";
import LoginForm from "../_components/login-form";
import { getProviders } from "next-auth/react";

export default async function Home() {
  const loginWithEmailEnabled = process.env.ALLOW_LOGIN_WITH_EMAIL === "true";
  if(!loginWithEmailEnabled){
    return notFound()
  }
  const withSSO = process.env.ALLOW_LOGIN_WITH_SSO=== "true";
  const emailCommEnabled = process.env.EMAIL_COMM_ENABLED === "true";
  const allowSignupWithEmail = process.env.ALLOW_SIGNUP_WITH_EMAIL === "true";
  const res =  await getProviders();
  return <LoginForm withSSO={withSSO} emailCommEnabled={emailCommEnabled} allowSignupWithEmail={allowSignupWithEmail} providers={res} />
}
=======
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
>>>>>>> 8ed8bd68d (feat(web): implement monospace design system with auth foundation)
