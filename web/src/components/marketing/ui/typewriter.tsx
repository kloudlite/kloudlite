'use client'

import { useState, useEffect } from 'react'
import { cn } from '@/lib/utils'

interface TypewriterProps {
  words: string[]
  className?: string
  typingSpeed?: number
  deletingSpeed?: number
  pauseDuration?: number
}

export function Typewriter({ 
  words, 
  className,
  typingSpeed = 100,
  deletingSpeed = 50,
  pauseDuration = 2000
}: TypewriterProps) {
  const [currentIndex, setCurrentIndex] = useState(0)
  const [displayedText, setDisplayedText] = useState('')
  const [isTyping, setIsTyping] = useState(true)

  useEffect(() => {
    const word = words[currentIndex]
    
    if (isTyping) {
      if (displayedText.length < word.length) {
        const timeout = setTimeout(() => {
          setDisplayedText(word.slice(0, displayedText.length + 1))
        }, typingSpeed)
        return () => clearTimeout(timeout)
      } else {
        const timeout = setTimeout(() => {
          setIsTyping(false)
        }, pauseDuration)
        return () => clearTimeout(timeout)
      }
    } else {
      if (displayedText.length > 0) {
        const timeout = setTimeout(() => {
          setDisplayedText(displayedText.slice(0, -1))
        }, deletingSpeed)
        return () => clearTimeout(timeout)
      } else {
        setCurrentIndex((prev) => (prev + 1) % words.length)
        setIsTyping(true)
      }
    }
  }, [currentIndex, isTyping, displayedText, words, typingSpeed, deletingSpeed, pauseDuration])

  return (
    <span className={cn("text-primary font-semibold", className)}>
      {displayedText}
      <span className="animate-pulse">|</span>
    </span>
  )
}