'use client'

import { useEffect, useState } from 'react'

const roles = [
  'Frontend Developer',
  'Backend Developer',
  'Full Stack Developer',
  'DevOps Engineer',
  'Platform Engineer',
]

export function TypewriterText() {
  const [currentRoleIndex, setCurrentRoleIndex] = useState(0)
  const [currentText, setCurrentText] = useState('')
  const [isDeleting, setIsDeleting] = useState(false)

  useEffect(() => {
    const currentRole = roles[currentRoleIndex]

    const timeout = setTimeout(() => {
      if (!isDeleting) {
        if (currentText.length < currentRole.length) {
          setCurrentText(currentRole.slice(0, currentText.length + 1))
        } else {
          setTimeout(() => setIsDeleting(true), 2000)
        }
      } else if (currentText.length > 0) {
        setCurrentText(currentText.slice(0, -1))
      } else {
        setIsDeleting(false)
        setCurrentRoleIndex((prev) => (prev + 1) % roles.length)
      }
    }, isDeleting ? 50 : 100)

    return () => clearTimeout(timeout)
  }, [currentText, isDeleting, currentRoleIndex])

  return (
    <span className="text-primary">
      <span aria-hidden="true">
        {currentText}
        <span className="animate-pulse">|</span>
      </span>
      <span className="sr-only" role="status" aria-live="polite" aria-atomic="true">
        {`For ${roles[currentRoleIndex]}`}
      </span>
    </span>
  )
}
