import { cache } from 'react'

import { type NextAuthOptions } from "next-auth"
import AzureADProvider from "next-auth/providers/azure-ad"
import CredentialsProvider from "next-auth/providers/credentials"
import GitHubProvider from "next-auth/providers/github"
import GoogleProvider from "next-auth/providers/google"


import type { PlatformSettings } from "@/grpc/accounts.external"
import type { 
  LoginRequest, 
  LoginResponse, 
  RefreshTokenRequest, 
  RefreshTokenResponse,
  LoginWithOAuthRequest,
  LoginWithOAuthResponse,
  GetUserDetailsRequest,
  GetUserDetailsResponse 
} from "@/grpc/auth.external"
import { getAccountsClient } from "@/lib/grpc/accounts-client"

import { getAuthClient, promisifyGrpcCall } from "./grpc-client"

// Custom user type
interface User {
  id: string
  email: string
  name: string
  emailVerified: boolean
  accessToken: string
  refreshToken: string
}

// Cache platform settings using React cache
const getCachedPlatformSettings = cache(async (): Promise<PlatformSettings | null> => {
  try {
    const client = getAccountsClient()
    const metadata = new (await import('@grpc/grpc-js')).Metadata()
    metadata.add('x-internal-request', 'true')
    
    const response = await new Promise<any>((resolve, reject) => {
      client.getPlatformSettings({}, metadata, (error, response) => {
        if (error) {reject(error)}
        else {resolve(response)}
      })
    })
    
    return response?.settings || null
  } catch (_error) {
    // Failed to fetch platform settings
    return null
  }
})

async function refreshAccessToken(token: any) {
  try {
    const authClient = getAuthClient()
    const response = await promisifyGrpcCall<RefreshTokenRequest, RefreshTokenResponse>(
      authClient.refreshToken.bind(authClient),
      {
        refreshToken: token.refreshToken
      }
    )

    return {
      ...token,
      accessToken: response.token,
      refreshToken: response.refreshToken,
      accessTokenExpires: Date.now() + 15 * 60 * 1000, // 15 minutes
    }
  } catch (_error) {
    // Error refreshing access token
    return {
      ...token,
      error: "RefreshAccessTokenError",
    }
  }
}

// Main auth configuration builder
async function buildAuthOptions(): Promise<NextAuthOptions> {
  const platformSettings = await getCachedPlatformSettings()
  
  const providers: any[] = [
    CredentialsProvider({
      name: "credentials",
      credentials: {
        email: { label: "Email", type: "email" },
        password: { label: "Password", type: "password" }
      },
      async authorize(credentials) {
        if (!credentials?.email || !credentials?.password) {
          return null
        }

        try {
          const authClient = getAuthClient()
          const response = await promisifyGrpcCall<LoginRequest, LoginResponse>(
            authClient.login.bind(authClient),
            {
              email: credentials.email,
              password: credentials.password
            }
          )

          // Fetch user details to get emailVerified status
          const userDetails = await promisifyGrpcCall<GetUserDetailsRequest, GetUserDetailsResponse>(
            authClient.getUserDetails.bind(authClient),
            {
              userId: response.userId
            }
          )

          // Return user object that will be saved in JWT
          return {
            id: response.userId,
            email: credentials.email,
            name: userDetails.name || credentials.email,
            emailVerified: userDetails.emailVerified,
            accessToken: response.token,
            refreshToken: response.refreshToken
          } as User
        } catch (_error) {
          // Login error
          return null
        }
      }
    }),
  ]

  // Add OAuth providers based on platform settings
  if (platformSettings?.oauthProviders?.google?.enabled) {
    providers.push(
      GoogleProvider({
        clientId: platformSettings.oauthProviders.google.clientId,
        clientSecret: platformSettings.oauthProviders.google.clientSecret,
        authorization: {
          params: {
            prompt: "consent",
            access_type: "offline",
            response_type: "code"
          }
        }
      })
    )
  }

  if (platformSettings?.oauthProviders?.github?.enabled) {
    providers.push(
      GitHubProvider({
        clientId: platformSettings.oauthProviders.github.clientId,
        clientSecret: platformSettings.oauthProviders.github.clientSecret,
      })
    )
  }

  if (platformSettings?.oauthProviders?.microsoft?.enabled) {
    providers.push(
      AzureADProvider({
        clientId: platformSettings.oauthProviders.microsoft.clientId,
        clientSecret: platformSettings.oauthProviders.microsoft.clientSecret,
        tenantId: platformSettings.oauthProviders.microsoft.tenantId || "common",
      })
    )
  }

  return {
    providers,
    callbacks: {
      async jwt({ token, user, account, profile }) {
        // Initial sign in
        if (account && user) {
          // OAuth sign in
          if (account.type === "oauth") {
            try {
              const authClient = getAuthClient()
              const response = await promisifyGrpcCall<LoginWithOAuthRequest, LoginWithOAuthResponse>(
                authClient.loginWithOAuth.bind(authClient),
                {
                  provider: account.provider,
                  email: user.email!,
                  name: user.name || user.email!,
                }
              )

              // Fetch user details to get emailVerified status
              const userDetails = await promisifyGrpcCall<GetUserDetailsRequest, GetUserDetailsResponse>(
                authClient.getUserDetails.bind(authClient),
                {
                  userId: response.userId
                }
              )

              return {
                ...token,
                id: response.userId,
                email: user.email,
                name: user.name,
                emailVerified: userDetails.emailVerified,
                accessToken: response.token,
                refreshToken: response.refreshToken,
                accessTokenExpires: Date.now() + 15 * 60 * 1000, // 15 minutes
              }
            } catch (_error) {
              // OAuth login error
              return token
            }
          }
          
          // Credentials sign in
          return {
            ...token,
            id: user.id,
            emailVerified: (user as User).emailVerified,
            accessToken: (user as User).accessToken,
            refreshToken: (user as User).refreshToken,
            accessTokenExpires: Date.now() + 15 * 60 * 1000, // 15 minutes
          }
        }

        // Return previous token if the access token has not expired yet
        if (Date.now() < (token.accessTokenExpires as number)) {
          return token
        }

        // Access token has expired, try to refresh it
        return refreshAccessToken(token)
      },
      async session({ session, token }) {
        if (token) {
          session.user = {
            ...session.user,
            id: token.id as string,
            emailVerified: token.emailVerified as boolean,
          }
          session.accessToken = token.accessToken as string
          session.error = token.error as string | undefined
        }
        return session
      }
    },
    pages: {
      signIn: "/auth/login",
      signOut: "/auth/logout",
      error: "/auth/error",
    },
    session: {
      strategy: "jwt",
      maxAge: 7 * 24 * 60 * 60, // 7 days
    },
    secret: process.env.NEXTAUTH_SECRET,
  }
}

// Export cached auth options getter
export const getAuthOptions = cache(buildAuthOptions)

// Export static instance for backward compatibility
export const authOptions = buildAuthOptions()