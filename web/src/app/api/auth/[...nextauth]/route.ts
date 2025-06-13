import NextAuth, { AuthOptions } from "next-auth"
import GithubProvider from "next-auth/providers/github"
import AzureADProvider from "next-auth/providers/azure-ad"
import GoogleProvider from "next-auth/providers/google"
import { redirect } from "next/navigation"
import { loginWithOAuth } from "@/actions/auth"

export const authOptions = {
  callbacks: {
    async signIn({ user, account, profile }) {
      await loginWithOAuth(user.name || "", user.email || "", "oauth");
      return true;
    }
  },
  providers: [
    ...((process.env.OAUTH_GITHUB_CLIENT_ID && process.env.OAUTH_GITHUB_CLIENT_SECRET) ? [
      GithubProvider({
        clientId: process.env.OAUTH_GITHUB_CLIENT_ID || "",
        clientSecret: process.env.OAUTH_GITHUB_CLIENT_SECRET || "",
      }),
    ] : []),
    ...((process.env.OAUTH_AZURE_AD_CLIENT_ID && process.env.OAUTH_AZURE_AD_CLIENT_SECRET) ? [
      AzureADProvider({
        clientId: process.env.OAUTH_AZURE_AD_CLIENT_ID || "",
        clientSecret: process.env.OAUTH_AZURE_AD_CLIENT_SECRET || ""
      }),
    ] : []),
    ...((process.env.OAUTH_GOOGLE_CLIENT_ID && process.env.OAUTH_GOOGLE_CLIENT_SECRET) ? [
      GoogleProvider({
        clientId: process.env.OAUTH_GOOGLE_CLIENT_ID || "",
        clientSecret: process.env.OAUTH_GOOGLE_CLIENT_SECRET || "",
      }),
    ]:[])
  ],
} as AuthOptions

const handler = NextAuth(authOptions)

export { handler as GET, handler as POST }

