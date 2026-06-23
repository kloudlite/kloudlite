import { useState } from 'react'
import { Building2, KeyRound, Mail } from 'lucide-react'
import { Button, FormField, TextInput } from './ui'

const CONSOLE_URL = 'https://console.kloudlite.io'
const KLOUDLITE_LOGO = new URL('../assets/kloudlite-logo.svg', import.meta.url).href

interface AuthGateProps {
  onAuthenticated: (user: string) => void
}

export function AuthGate({ onAuthenticated }: AuthGateProps) {
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [ssoDomain, setSsoDomain] = useState('')
  const [error, setError] = useState<string | null>(null)

  function signIn() {
    if (!username.trim() || !password) {
      setError('Enter username and password')
      return
    }
    localStorage.setItem('kloudlite.desktop.auth', JSON.stringify({ user: username.trim(), signedInAt: Date.now() }))
    onAuthenticated(username.trim())
  }

  function openLogin(path: string) {
    window.electronAPI.openExternal(`${CONSOLE_URL}${path}`)
  }

  return (
    <div className="relative flex h-screen items-center justify-center overflow-hidden bg-background px-6">
      <img src={KLOUDLITE_LOGO} alt="" aria-hidden="true" className="pointer-events-none absolute -right-16 -top-16 h-72 w-72 opacity-[0.04]" />
      <img src={KLOUDLITE_LOGO} alt="" aria-hidden="true" className="pointer-events-none absolute -bottom-20 -left-20 h-80 w-80 opacity-[0.035]" />
      <div className="relative w-full max-w-[420px] rounded-2xl border border-border/50 bg-card/80 p-8 shadow-2xl">
        <div className="mb-8 text-center">
          <img src={KLOUDLITE_LOGO} alt="Kloudlite" className="mx-auto mb-4 h-14 w-14" />
          <h1 className="text-[22px] font-semibold text-foreground">Sign in to Kloudlite</h1>
          <p className="mt-1 text-[13px] text-muted-foreground">Authenticate before opening the desktop workspace.</p>
        </div>

        <div className="grid grid-cols-2 gap-2">
          <Button variant="secondary" onClick={() => openLogin('/api/oauth/google')}><Mail className="h-4 w-4" />Google</Button>
          <Button variant="secondary" onClick={() => openLogin('/api/oauth/microsoft-entra-id')}><Building2 className="h-4 w-4" />Microsoft</Button>
        </div>

        <div className="my-6 flex items-center gap-3">
          <div className="h-px flex-1 bg-border/60" />
          <span className="text-[11px] uppercase tracking-[0.18em] text-muted-foreground">or password</span>
          <div className="h-px flex-1 bg-border/60" />
        </div>

        <form className="space-y-4" onSubmit={(e) => { e.preventDefault(); signIn() }}>
          {error && <div className="rounded-lg border border-red-500/20 bg-red-500/[0.06] px-3 py-2 text-[12px] text-red-400">{error}</div>}
          <FormField label="Username or email">
            <TextInput value={username} onChange={(e) => setUsername(e.target.value)} placeholder="you@company.com" />
          </FormField>
          <FormField label="Password">
            <TextInput type="password" value={password} onChange={(e) => setPassword(e.target.value)} placeholder="••••••••" />
          </FormField>
          <Button className="w-full" type="submit">Sign in</Button>
        </form>

        <div className="mt-6 rounded-xl border border-border/50 bg-background/50 p-3">
          <div className="mb-2 flex items-center gap-2 text-[12px] font-medium text-foreground">
            <KeyRound className="h-3.5 w-3.5" /> Enterprise SSO
          </div>
          <div className="flex gap-2">
            <TextInput value={ssoDomain} onChange={(e) => setSsoDomain(e.target.value)} placeholder="company.com" />
            <Button variant="secondary" disabled={!ssoDomain.trim()} onClick={() => openLogin(`/login?sso=${encodeURIComponent(ssoDomain.trim())}`)}>Continue</Button>
          </div>
        </div>
      </div>
    </div>
  )
}
