'use server'

import { getAuthClient } from '@/lib/auth/grpc-client'

export async function verifyEmail(token: string) {
  try {
    const authClient = getAuthClient()
    
    return new Promise<{ success: boolean; userId?: string }>((resolve, reject) => {
      authClient.verifyEmail({ verificationToken: token }, (error, response) => {
        if (error) {
          console.error('Email verification error:', error)
          reject(error)
          return
        }
        
        resolve({
          success: response.success,
          userId: response.userId
        })
      })
    })
  } catch (error) {
    console.error('Email verification error:', error)
    throw error
  }
}

export async function resendVerificationEmail(email: string) {
  try {
    const authClient = getAuthClient()
    
    return new Promise<boolean>((resolve) => {
      authClient.resendEmailVerification({ email }, (error, response) => {
        if (error) {
          console.error('Resend verification error:', error)
          resolve(false)
          return
        }
        
        resolve(response.success)
      })
    })
  } catch (error) {
    console.error('Resend verification error:', error)
    return false
  }
}