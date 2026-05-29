import { ArrowLeft, ArrowRight } from 'lucide-react'
import { useEffect, useState } from 'react'

interface NavIndicatorProps {
  direction: 'back' | 'forward' | null
}

export function NavIndicator({ direction }: NavIndicatorProps) {
  const [visible, setVisible] = useState(false)
  const [currentDir, setCurrentDir] = useState<'back' | 'forward' | null>(null)

  useEffect(() => {
    if (direction) {
      setCurrentDir(direction)
      setVisible(true)
      const timer = setTimeout(() => setVisible(false), 400)
      return () => clearTimeout(timer)
    }
  }, [direction])

  if (!visible || !currentDir) return null

  return (
    <div
      className={`pointer-events-none absolute top-1/2 z-20 -translate-y-1/2 animate-[nav-flash_0.4s_ease-out_forwards] ${
        currentDir === 'back' ? 'left-3' : 'right-3'
      }`}
    >
      <div className="flex h-9 w-9 items-center justify-center rounded-full bg-primary/90 text-primary-foreground shadow-lg">
        {currentDir === 'back' ? (
          <ArrowLeft className="h-5 w-5" />
        ) : (
          <ArrowRight className="h-5 w-5" />
        )}
      </div>
    </div>
  )
}
