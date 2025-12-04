'use client'

interface LocalTimeProps {
  date: string | Date
  format?: 'short' | 'long'
}

export function LocalTime({ date, format = 'short' }: LocalTimeProps) {
  const d = new Date(date)

  if (format === 'long') {
    return (
      <span>
        {d.toLocaleDateString('en-US', {
          year: 'numeric',
          month: 'short',
          day: 'numeric',
        })}
      </span>
    )
  }

  return (
    <span>
      {d.toLocaleString('en-US', {
        month: 'short',
        day: 'numeric',
        hour: '2-digit',
        minute: '2-digit',
      })}
    </span>
  )
}
