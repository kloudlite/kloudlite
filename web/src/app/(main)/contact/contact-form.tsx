'use client'

import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import { useState, useEffect } from 'react'
import { toast } from 'sonner'
import { CheckCircle2 } from 'lucide-react'

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

  // Check rate limit on component mount
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

  // Update time remaining every minute
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
      }, 60000) // Update every minute
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
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(formData),
      })

      const data = await response.json()

      if (!response.ok) {
        throw new Error(data.error || 'Failed to submit form')
      }

      // Store submission timestamp
      localStorage.setItem('lastContactSubmission', Date.now().toString())
      setCanSubmit(false)
      setTimeRemaining(60) // 60 minutes

      // Show success state
      setIsSuccess(true)
      toast.success(data.message || "Message sent successfully! We'll get back to you soon.", {
        duration: 5000,
      })
      setFormData({ name: '', email: '', subject: '', message: '' })

      // Hide success state after 10 seconds
      setTimeout(() => setIsSuccess(false), 10000)
    } catch (error) {
      console.error('Form submission error:', error)
      toast.error(
        error instanceof Error ? error.message : 'Failed to send message. Please try again.',
      )
    } finally {
      setIsSubmitting(false)
    }
  }

  const handleChange = (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>) => {
    setFormData({
      ...formData,
      [e.target.name]: e.target.value,
    })
  }

  return (
    <div className="bg-card border-border rounded-lg border p-8">
      <h2 className="text-foreground text-2xl font-semibold">Send us a message</h2>
      <p className="text-muted-foreground mt-2">
        Fill out the form below and we&apos;ll get back to you as soon as possible.
      </p>

      {/* Success Message */}
      {isSuccess && (
        <div className="mt-6 rounded-lg bg-green-50 p-6 dark:bg-green-950/20">
          <div className="flex items-start gap-4">
            <CheckCircle2 className="h-6 w-6 flex-shrink-0 text-green-600 dark:text-green-400" />
            <div>
              <h3 className="text-lg font-semibold text-green-900 dark:text-green-100">
                Message Sent Successfully!
              </h3>
              <p className="mt-1 text-sm text-green-800 dark:text-green-200">
                Thank you for reaching out. We&apos;ve received your message and will get back to
                you as soon as possible.
              </p>
            </div>
          </div>
        </div>
      )}

      {/* Rate Limit Warning */}
      {!canSubmit && !isSuccess && (
        <div className="mt-6 rounded-lg bg-amber-50 p-4 dark:bg-amber-950/20">
          <p className="text-sm text-amber-800 dark:text-amber-200">
            You can send another message in {timeRemaining} minute{timeRemaining !== 1 ? 's' : ''}.
          </p>
        </div>
      )}

      <form onSubmit={handleSubmit} className="mt-6 space-y-6">
        <div className="grid gap-6 sm:grid-cols-2">
          <div className="space-y-2">
            <label htmlFor="name" className="text-foreground text-sm font-medium">
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
            />
          </div>

          <div className="space-y-2">
            <label htmlFor="email" className="text-foreground text-sm font-medium">
              Email
            </label>
            <Input
              id="email"
              name="email"
              type="email"
              placeholder="your@email.com"
              value={formData.email}
              onChange={handleChange}
              required
            />
          </div>
        </div>

        <div className="space-y-2">
          <label htmlFor="subject" className="text-foreground text-sm font-medium">
            Subject
          </label>
          <Input
            id="subject"
            name="subject"
            type="text"
            placeholder="What is this about?"
            value={formData.subject}
            onChange={handleChange}
            required
          />
        </div>

        <div className="space-y-2">
          <label htmlFor="message" className="text-foreground text-sm font-medium">
            Message
          </label>
          <Textarea
            id="message"
            name="message"
            placeholder="Your message..."
            rows={6}
            value={formData.message}
            onChange={handleChange}
            required
          />
        </div>

        <Button
          type="submit"
          size="lg"
          className="w-full sm:w-auto"
          disabled={isSubmitting || !canSubmit}
        >
          {isSubmitting ? 'Sending...' : !canSubmit ? 'Please Wait...' : 'Send Message'}
        </Button>
        {!canSubmit && (
          <p className="text-muted-foreground text-sm">
            Rate limit: You can send another message in {timeRemaining} minute
            {timeRemaining !== 1 ? 's' : ''}
          </p>
        )}
      </form>
    </div>
  )
}
