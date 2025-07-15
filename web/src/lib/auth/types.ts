export interface User {
  id: string
  email: string
  name: string
  verified: boolean
  createdAt: Date
  updatedAt: Date
}

export interface Session {
  user: User
  expiresAt: Date
}

export interface AuthResponse {
  success: boolean
  user?: User
  error?: string
}

export interface LoginCredentials {
  email: string
  password: string
  rememberMe?: boolean
}

export interface SignupCredentials {
  name: string
  email: string
  password: string
  acceptTerms: boolean
}

export interface ForgotPasswordData {
  email: string
}

export interface ResetPasswordData {
  token: string
  password: string
  confirmPassword: string
}