import { useEffect, useState, useCallback } from 'react'
import { Shield, ShieldAlert, ShieldCheck, ChevronRight, X } from 'lucide-react'
import { cn } from '@/lib/utils'

interface CertInfo {
  subject: { CN?: string; O?: string }
  issuer: { CN?: string; O?: string }
  validFrom: string
  validTo: string
  fingerprint256: string
  serialNumber: string
  subjectaltname?: string
}

interface CertData {
  chain: CertInfo[]
  authorized: boolean
}

interface CertPopoverProps {
  url: string
  anchorRect: DOMRect | null
  onClose: () => void
}

export function CertPopover({ url, anchorRect, onClose }: CertPopoverProps) {
  const [certData, setCertData] = useState<CertData | null>(null)
  const [loading, setLoading] = useState(true)
  const [selectedIndex, setSelectedIndex] = useState(0)
  const [showDetails, setShowDetails] = useState(false)
  const [exiting, setExiting] = useState(false)

  const close = useCallback(() => {
    setExiting(true)
    setTimeout(onClose, 150)
  }, [onClose])

  useEffect(() => {
    setLoading(true)
    window.electronAPI.getCertificate(url).then((data) => {
      setCertData(data)
      setLoading(false)
    })
  }, [url])

  const isHttps = url.startsWith('https://')
  const cert = certData?.chain[selectedIndex]
  const leafCert = certData?.chain[0]

  const style: React.CSSProperties = anchorRect
    ? { top: anchorRect.bottom + 6, left: anchorRect.left }
    : { top: 80, left: 12 }

  return (
    <div className="fixed inset-0 z-50" onClick={close}>
      <div
        className="fixed w-[320px] origin-top-left overflow-hidden rounded-xl border border-border/50 bg-popover shadow-2xl shadow-black/20"
        style={{ ...style, animation: exiting ? 'popover-out 150ms ease-in forwards' : 'popover-in 150ms ease-out' }}
        onClick={(e) => e.stopPropagation()}
      >
        {/* Header */}
        <div className="flex items-center gap-2.5 border-b border-border/30 px-4 py-3">
          {!isHttps ? (
            <ShieldAlert className="h-5 w-5 shrink-0 text-amber-500" />
          ) : certData?.authorized ? (
            <ShieldCheck className="h-5 w-5 shrink-0 text-emerald-500" />
          ) : (
            <ShieldAlert className="h-5 w-5 shrink-0 text-red-500" />
          )}
          <div className="min-w-0 flex-1">
            <p className="text-[13px] font-semibold text-foreground">
              {!isHttps
                ? 'Not Secure'
                : certData?.authorized
                  ? 'Connection is Secure'
                  : 'Certificate Error'}
            </p>
            <p className="truncate text-[11px] text-muted-foreground">
              {isHttps ? leafCert?.subject?.CN || '' : 'This site does not use HTTPS'}
            </p>
          </div>
          <button
            className="shrink-0 rounded-md p-1 text-muted-foreground hover:bg-accent hover:text-foreground"
            onClick={close}
          >
            <X className="h-3.5 w-3.5" />
          </button>
        </div>

        {loading && (
          <div className="flex items-center justify-center py-8">
            <div className="h-5 w-5 animate-spin rounded-full border-2 border-muted-foreground/30 border-t-muted-foreground" />
          </div>
        )}

        {!loading && certData && (
          <>
            {/* Certificate chain */}
            <div className="border-b border-border/30 px-3 py-2">
              <p className="mb-1.5 text-[10px] font-medium uppercase tracking-wider text-muted-foreground/60">
                Certificate Chain
              </p>
              {certData.chain.map((c, i) => (
                <button
                  key={i}
                  className={cn(
                    'flex w-full items-center gap-1.5 rounded-md px-2 py-1 text-left text-[12px] transition-colors',
                    i === selectedIndex
                      ? 'bg-primary text-primary-foreground'
                      : 'text-foreground hover:bg-accent'
                  )}
                  style={{ paddingLeft: `${8 + i * 14}px` }}
                  onClick={() => setSelectedIndex(i)}
                >
                  {i > 0 && <span className="text-[10px] opacity-50">&#8627;</span>}
                  <Shield className="h-3 w-3 shrink-0" />
                  <span className="truncate">{c.subject?.CN || c.subject?.O || 'Unknown'}</span>
                </button>
              ))}
            </div>

            {/* Selected certificate details */}
            {cert && (
              <div className="px-4 py-3">
                <p className="text-[13px] font-semibold text-foreground">
                  {cert.subject?.CN || cert.subject?.O || 'Unknown'}
                </p>
                <p className="mt-0.5 text-[11px] text-muted-foreground">
                  Issued by: {cert.issuer?.CN || cert.issuer?.O || 'Unknown'}
                </p>
                <p className="mt-0.5 text-[11px] text-muted-foreground">
                  Expires: {new Date(cert.validTo).toLocaleDateString(undefined, {
                    weekday: 'long',
                    year: 'numeric',
                    month: 'long',
                    day: 'numeric'
                  })}
                </p>
                {selectedIndex === 0 && (
                  <p className={cn(
                    'mt-1.5 flex items-center gap-1 text-[11px] font-medium',
                    certData.authorized ? 'text-emerald-500' : 'text-red-500'
                  )}>
                    {certData.authorized ? '✓' : '✗'}
                    {certData.authorized ? ' This certificate is valid' : ' Certificate validation failed'}
                  </p>
                )}

                {/* Expandable details */}
                <button
                  className="mt-2 flex items-center gap-1 text-[11px] font-medium text-muted-foreground hover:text-foreground"
                  onClick={() => setShowDetails(!showDetails)}
                >
                  <ChevronRight className={cn('h-3 w-3 transition-transform', showDetails && 'rotate-90')} />
                  Details
                </button>
                {showDetails && (
                  <div className="mt-1.5 space-y-1 text-[10px] text-muted-foreground">
                    <p><span className="font-medium text-foreground/70">Serial:</span> {cert.serialNumber}</p>
                    <p className="break-all"><span className="font-medium text-foreground/70">SHA-256:</span> {cert.fingerprint256}</p>
                    {cert.subjectaltname && (
                      <p className="break-all"><span className="font-medium text-foreground/70">SANs:</span> {cert.subjectaltname}</p>
                    )}
                    <p><span className="font-medium text-foreground/70">Valid from:</span> {new Date(cert.validFrom).toLocaleString()}</p>
                    <p><span className="font-medium text-foreground/70">Valid to:</span> {new Date(cert.validTo).toLocaleString()}</p>
                  </div>
                )}
              </div>
            )}
          </>
        )}

        {!loading && !certData && isHttps && (
          <div className="px-4 py-4 text-center text-[12px] text-muted-foreground">
            Could not retrieve certificate information
          </div>
        )}
      </div>
    </div>
  )
}
