import { type ReactNode } from 'react'

interface FormFieldProps {
  label: string
  hint?: string
  optional?: boolean
  children: ReactNode
}

export function FormField({ label, hint, optional, children }: FormFieldProps) {
  return (
    <div>
      <label className="mb-1.5 block text-[12px] font-medium text-foreground">
        {label}
        {optional && <span className="ml-1 text-muted-foreground">(optional)</span>}
      </label>
      {children}
      {hint && <p className="mt-1 text-[11px] text-muted-foreground">{hint}</p>}
    </div>
  )
}

export function TextInput(props: React.InputHTMLAttributes<HTMLInputElement>) {
  return (
    <input
      {...props}
      className="w-full rounded-lg border border-border bg-background px-3 py-2 text-[13px] text-foreground outline-none transition-colors focus:border-primary"
    />
  )
}
