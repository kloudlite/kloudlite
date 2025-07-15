'use client'

import { useState, useMemo } from 'react'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import { Eye, EyeOff } from 'lucide-react'
import { cn } from '@/lib/utils'

interface PasswordInputProps extends React.InputHTMLAttributes<HTMLInputElement> {
  showStrength?: boolean
}

function getPasswordStrength(password: string): {
  score: number
  label: string
  color: string
} {
  if (!password) return { score: 0, label: '', color: '' }
  
  let score = 0
  
  // Length check
  if (password.length >= 8) score++
  if (password.length >= 12) score++
  
  // Character variety checks
  if (/[a-z]/.test(password)) score++
  if (/[A-Z]/.test(password)) score++
  if (/[0-9]/.test(password)) score++
  if (/[^A-Za-z0-9]/.test(password)) score++
  
  // Determine strength label and color
  if (score <= 2) return { score: 1, label: 'Weak', color: 'bg-destructive' }
  if (score <= 4) return { score: 2, label: 'Fair', color: 'bg-yellow-500' }
  if (score <= 5) return { score: 3, label: 'Good', color: 'bg-blue-500' }
  return { score: 4, label: 'Strong', color: 'bg-green-500' }
}

export function PasswordInput({ showStrength, ...props }: PasswordInputProps) {
  const [showPassword, setShowPassword] = useState(false)
  
  const strength = useMemo(() => {
    if (!showStrength || !props.value) return null
    return getPasswordStrength(String(props.value))
  }, [showStrength, props.value])

  return (
    <div>
      <div className="relative">
        <Input
          {...props}
          type={showPassword ? 'text' : 'password'}
        />
        <div className="absolute right-0 top-0 h-full flex items-center px-3">
          <Button
            type="button"
            variant="ghost"
            size="sm"
            onClick={() => setShowPassword(!showPassword)}
          >
            {showPassword ? (
              <EyeOff className="h-4 w-4 text-muted-foreground" />
            ) : (
              <Eye className="h-4 w-4 text-muted-foreground" />
            )}
            <span className="sr-only">
              {showPassword ? 'Hide password' : 'Show password'}
            </span>
          </Button>
        </div>
      </div>
      
      {showStrength && (
        <div 
          className={cn(
            "mt-2 space-y-1 transition-all duration-300 ease-out",
            strength && strength.score > 0 ? "opacity-100 translate-y-0" : "opacity-0 -translate-y-1 pointer-events-none"
          )}
          style={{
            height: strength && strength.score > 0 ? 'auto' : '0px'
          }}
        >
          <div className="flex gap-1">
            {[1, 2, 3, 4].map((level) => (
              <div
                key={level}
                className={cn(
                  'h-1 flex-1 rounded-none bg-muted transition-all duration-500 ease-out transform origin-left',
                  strength && level <= strength.score ? [strength.color, 'scale-x-100'] : 'scale-x-100 opacity-50'
                )}
                style={{
                  transitionDelay: strength && level <= strength.score ? `${(level - 1) * 75}ms` : '0ms'
                }}
              />
            ))}
          </div>
          {strength && strength.label && (
            <p className="text-xs text-muted-foreground transition-opacity duration-300">
              Password strength: {strength.label}
            </p>
          )}
        </div>
      )}
    </div>
  )
}