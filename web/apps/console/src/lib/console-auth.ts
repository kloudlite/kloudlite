import { cookies } from 'next/headers'
import { jwtVerify } from 'jose'

interface RegistrationSession {
  user: {
    id: string
    email: string
    name: string
    image?: string
  }
  provider: string
  installationKey?: string
}

/**
 * Get the current registration session from the registration_session cookie
 * This is used for the access-console/installation flow
 */
export async function getRegistrationSession(): Promise<RegistrationSession | null> {
  const cookieStore = await cookies()
  const token = cookieStore.get('registration_session')?.value

  if (!token) {
    return null
  }

  try {
    const secret = new TextEncoder().encode(process.env.NEXTAUTH_SECRET)
    const { payload } = await jwtVerify(token, secret)

    return {
      user: {
        id: payload.userId as string,
        email: payload.email as string,
        name: payload.name as string,
        image: payload.image as string | undefined,
      },
      provider: payload.provider as string,
      installationKey: payload.installationKey as string | undefined,
    }
  } catch {
    return null
  }
}
