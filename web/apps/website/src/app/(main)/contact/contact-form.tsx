'use client'

import { Button } from '@kloudlite/ui'
import { Input } from '@kloudlite/ui'
import { Textarea } from '@kloudlite/ui'
import { useState, useEffect } from 'react'
import { toast } from 'sonner'
import { ArrowRight } from 'lucide-react'

export default function ContactForm() {
  const [formData, setFormData] = useState({
    name: '',
    email: '',
    subject: '',
    message: '',
  })
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [isSuccess, setIsSuccess] = useState(false)
  const [canSubmit, setCanSubmit] = useState(true)
  const [timeRemaining, setTimeRemaining] = useState(0)

  useEffect(() => {
    const lastSubmission = localStorage.getItem('lastContactSubmission')
    if (lastSubmission) {
      const timeSinceSubmission = Date.now() - parseInt(lastSubmission, 10)
      const oneHourInMs = 60 * 60 * 1000
      if (timeSinceSubmission < oneHourInMs) {
        setCanSubmit(false)
        setTimeRemaining(Math.ceil((oneHourInMs - timeSinceSubmission) / 60000))
      }
    }
  }, [])

  useEffect(() => {
    if (!canSubmit && timeRemaining > 0) {
      const timer = setInterval(() => {
        setTimeRemaining((prev) => {
          if (prev <= 1) {
            setCanSubmit(true)
            localStorage.removeItem('lastContactSubmission')
            return 0
          }
          return prev - 1
        })
      }, 60000)
      return () => clearInterval(timer)
    }
    return undefined
  }, [canSubmit, timeRemaining])

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()

    if (!canSubmit) {
      toast.error(`Please wait ${timeRemaining} minutes before sending another message.`)
      return
    }

    setIsSubmitting(true)

    try {
      const response = await fetch('/api/contact/submit', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(formData),
      })

      const data = await response.json()

      if (!response.ok) {
        throw new Error(data.error || 'Failed to submit form')
      }

      localStorage.setItem('lastContactSubmission', Date.now().toString())
      setCanSubmit(false)
      setTimeRemaining(60)
      setIsSuccess(true)
      toast.success(data.message || "Message sent! We'll get back to you soon.")
      setFormData({ name: '', email: '', subject: '', message: '' })
      setTimeout(() => setIsSuccess(false), 10000)
    } catch (error) {
      console.error('Form submission error:', error)
      toast.error(error instanceof Error ? error.message : 'Failed to send message.')
    } finally {
      setIsSubmitting(false)
    }
  }

  const handleChange = (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>) => {
    setFormData({ ...formData, [e.target.name]: e.target.value })
  }

  if (isSuccess) {
    return (
      <div className="flex flex-col items-center justify-center py-12 text-center">
        <div className="w-12 h-12 rounded-full bg-green-500/10 flex items-center justify-center mb-4">
          <svg className="w-6 h-6 text-green-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
          </svg>
        </div>
        <h3 className="text-foreground text-lg font-semibold">Message Sent</h3>
        <p className="text-foreground/50 mt-2 text-sm max-w-sm">
          Thank you for reaching out. We&apos;ll get back to you within 24 hours.
        </p>
      </div>
    )
  }

  if (!canSubmit) {
    return (
      <div className="flex flex-col items-center justify-center py-12 text-center">
        <p className="text-foreground/50 text-sm">
          You can send another message in <span className="text-foreground font-medium">{timeRemaining}</span> minute{timeRemaining !== 1 ? 's' : ''}.
        </p>
      </div>
    )
  }

  return (
    <form onSubmit={handleSubmit} className="space-y-6">
      <div className="grid gap-6 sm:grid-cols-2">
        <div className="space-y-2">
          <label htmlFor="name" className="text-foreground/70 text-sm">
            Name
          </label>
          <Input
            id="name"
            name="name"
            type="text"
            placeholder="Your name"
            value={formData.name}
            onChange={handleChange}
            required
            className="rounded-none border-foreground/10 bg-transparent focus:border-foreground/30 h-11"
          />
        </div>

        <div className="space-y-2">
          <label htmlFor="email" className="text-foreground/70 text-sm">
            Email
          </label>
          <Input
            id="email"
            name="email"
            type="email"
            placeholder="you@company.com"
            value={formData.email}
            onChange={handleChange}
            required
            className="rounded-none border-foreground/10 bg-transparent focus:border-foreground/30 h-11"
          />
        </div>
      </div>

      <div className="space-y-2">
        <label htmlFor="subject" className="text-foreground/70 text-sm">
          Subject
        </label>
        <Input
          id="subject"
          name="subject"
          type="text"
          placeholder="How can we help?"
          value={formData.subject}
          onChange={handleChange}
          required
          className="rounded-none border-foreground/10 bg-transparent focus:border-foreground/30 h-11"
        />
      </div>

      <div className="space-y-2">
        <label htmlFor="message" className="text-foreground/70 text-sm">
          Message
        </label>
        <Textarea
          id="message"
          name="message"
          placeholder="Tell us more about your inquiry..."
          rows={5}
          value={formData.message}
          onChange={handleChange}
          required
          className="rounded-none border-foreground/10 bg-transparent focus:border-foreground/30 resize-none"
        />
      </div>

      <Button
        type="submit"
        className="rounded-none h-11 px-6"
        disabled={isSubmitting}
      >
        {isSubmitting ? 'Sending...' : 'Send Message'}
        {!isSubmitting && <ArrowRight className="ml-2 h-4 w-4" />}
      </Button>
    </form>
  )
}
