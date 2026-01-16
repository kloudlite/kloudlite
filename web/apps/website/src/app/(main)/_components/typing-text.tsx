'use client'

import { useState, useEffect } from 'react'

const personas = [
  'Frontend Developer',
  'Backend Developer',
  'Full-Stack Developer',
  'DevOps Engineer',
  'Data Engineer',
  'AI Engineer',
  'QA Engineer',
  'Cloud Architect',
  'Platform Engineer',
  'Site Reliability Engineer',
  'ML Engineer',
  'Security Engineer',
  'Database Administrator',
  'Software Architect',
  'API Developer',
]

const colors = [
  'text-primary',
  'text-primary',
  'text-primary',
  'text-primary',
  'text-primary',
  'text-primary',
  'text-primary',
  'text-primary',
  'text-primary',
  'text-primary',
  'text-primary',
  'text-primary',
  'text-primary',
  'text-primary',
  'text-primary',
  'text-primary',
]

export function TypingText() {
  const [text, setText] = useState('')
  const [personaIndex, setPersonaIndex] = useState(0)

  useEffect(() => {
    let currentIndex = 0
    let isDeleting = false
    let timeout: NodeJS.Timeout

    const type = () => {
      const currentPersona = personas[personaIndex]

      if (!isDeleting && currentIndex <= currentPersona.length) {
        setText(currentPersona.substring(0, currentIndex))
        currentIndex++
        timeout = setTimeout(type, 100)
      } else if (!isDeleting && currentIndex > currentPersona.length) {
        timeout = setTimeout(() => {
          isDeleting = true
          type()
        }, 2000)
      } else if (isDeleting && currentIndex >= 0) {
        setText(currentPersona.substring(0, currentIndex))
        currentIndex--
        timeout = setTimeout(type, 50)
      } else if (isDeleting && currentIndex < 0) {
        isDeleting = false
        currentIndex = 0
        setPersonaIndex((prev) => (prev + 1) % personas.length)
      }
    }

    type()

    return () => clearTimeout(timeout)
  }, [personaIndex])

  return (
    <span className="inline-block min-w-[280px] text-left">
      <span className={colors[personaIndex]}>{text}</span>
      <span className="animate-pulse">|</span>
    </span>
  )
}
