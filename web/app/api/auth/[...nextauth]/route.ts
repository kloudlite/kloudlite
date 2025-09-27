import NextAuth from "next-auth"

import { getAuthOptions } from "@/lib/auth/get-auth-options"

async function handler(req: Request, res: Response) {
  const authOptions = await getAuthOptions()
  return NextAuth(req, res, authOptions)
}

export { handler as GET, handler as POST }