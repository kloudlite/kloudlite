'use client'

import React from 'react'

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
  'text-blue-600 dark:text-blue-400',
  'text-emerald-600 dark:text-emerald-400',
  'text-purple-600 dark:text-purple-400',
  'text-orange-600 dark:text-orange-400',
  'text-pink-600 dark:text-pink-400',
  'text-cyan-600 dark:text-cyan-400',
  'text-amber-600 dark:text-amber-400',
  'text-rose-600 dark:text-rose-400',
  'text-indigo-600 dark:text-indigo-400',
  'text-teal-600 dark:text-teal-400',
  'text-violet-600 dark:text-violet-400',
  'text-lime-600 dark:text-lime-400',
  'text-fuchsia-600 dark:text-fuchsia-400',
  'text-sky-600 dark:text-sky-400',
  'text-red-600 dark:text-red-400',
  'text-green-600 dark:text-green-400',
]

export function TypingText() {
  const [text, setText] = React.useState('')
  const [personaIndex, setPersonaIndex] = React.useState(0)

  React.useEffect(() => {
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
