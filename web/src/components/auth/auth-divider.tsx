export function AuthDivider({ text = 'OR' }: { text?: string }) {
  return (
    <div className="relative">
      <div className="absolute inset-0 flex items-center">
        <span className="w-full border-t border-border" />
      </div>
      <div className="relative flex justify-center text-sm">
        <span className="bg-background px-3 text-muted-foreground">{text}</span>
      </div>
    </div>
  )
}