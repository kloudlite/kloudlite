"use server"

import type { 
  SignupRequest, 
  SignupResponse,
  RequestResetPasswordRequest,
  RequestResetPasswordResponse,
  ResetPasswordRequest,
  ResetPasswordResponse,
  VerifyEmailRequest,
  VerifyEmailResponse,
  ResendEmailVerificationRequest,
  ResendEmailVerificationResponse,
  VerifyDeviceCodeRequest,
  VerifyDeviceCodeResponse
} from "@/grpc/auth.external"
import { getAuthClient, promisifyGrpcCall } from "@/lib/auth/grpc-client"
import { formatErrorMessage } from "@/lib/format-error"

export async function signUpAction(data: { name: string; email: string; password: string }) {
  try {
    const authClient = getAuthClient()
    const response = await promisifyGrpcCall<SignupRequest, SignupResponse>(
      authClient.signup.bind(authClient),
      {
        name: data.name,
        email: data.email,
        password: data.password
      }
    )

    return { 
      success: true, 
      userId: response.userId,
      token: response.token,
      refreshToken: response.refreshToken
    }
  } catch (error) {
    const errorMessage = error instanceof Error ? error.message : 'Failed to create account'
    return { 
      success: false, 
      error: formatErrorMessage(errorMessage) 
    }
  }
}

export async function requestPasswordResetAction(email: string) {
  try {
    const authClient = getAuthClient()
    const response = await promisifyGrpcCall<RequestResetPasswordRequest, RequestResetPasswordResponse>(
      authClient.requestResetPassword.bind(authClient),
      { email }
    )

    return { 
      success: response.success,
      message: response.resetToken || "Reset link sent to your email"
    }
  } catch (error) {
    const errorMessage = error instanceof Error ? error.message : 'Failed to request password reset'
    return { 
      success: false, 
      error: formatErrorMessage(errorMessage) 
    }
  }
}

export async function resetPasswordAction(resetToken: string, newPassword: string) {
  try {
    const authClient = getAuthClient()
    const response = await promisifyGrpcCall<ResetPasswordRequest, ResetPasswordResponse>(
      authClient.resetPassword.bind(authClient),
      { resetToken, newPassword }
    )

    return { success: response.success }
  } catch (error) {
    const errorMessage = error instanceof Error ? error.message : 'Failed to reset password'
    return { 
      success: false, 
      error: formatErrorMessage(errorMessage) 
    }
  }
}

export async function verifyEmailAction(verificationToken: string) {
  try {
    const authClient = getAuthClient()
    const response = await promisifyGrpcCall<VerifyEmailRequest, VerifyEmailResponse>(
      authClient.verifyEmail.bind(authClient),
      { verificationToken }
    )

    return { 
      success: response.success,
      userId: response.userId
    }
  } catch (error) {
    const errorMessage = error instanceof Error ? error.message : 'Failed to verify email'
    return { 
      success: false, 
      error: formatErrorMessage(errorMessage) 
    }
  }
}

export async function resendVerificationEmailAction(email: string) {
  try {
    const authClient = getAuthClient()
    const response = await promisifyGrpcCall<ResendEmailVerificationRequest, ResendEmailVerificationResponse>(
      authClient.resendEmailVerification.bind(authClient),
      { email }
    )

    return { 
      success: response.success,
      message: response.message
    }
  } catch (error) {
    const errorMessage = error instanceof Error ? error.message : 'Failed to resend verification email'
    return { 
      success: false, 
      error: formatErrorMessage(errorMessage) 
    }
  }
}

export async function verifyDeviceCodeAction({ userCode, userId }: { userCode: string; userId: string }) {
  try {
    const authClient = getAuthClient()
    const response = await promisifyGrpcCall<VerifyDeviceCodeRequest, VerifyDeviceCodeResponse>(
      authClient.verifyDeviceCode.bind(authClient),
      {
        userCode,
        userId
      }
    )

    return {
      success: response.success,
      error: response.success ? null : response.message
    }
  } catch (error) {
    const errorMessage = error instanceof Error ? error.message : 'Failed to verify device code'
    return {
      success: false,
      error: formatErrorMessage(errorMessage)
    }
  }
}

