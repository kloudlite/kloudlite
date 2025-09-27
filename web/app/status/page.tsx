'use client'

import { useState } from "react"

import { CheckCircle2, AlertCircle, XCircle, Minus, Circle } from "lucide-react"
import Link from "next/link"

import { ThemeToggle } from "@/components/theme-toggle"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { ScrollArea } from "@/components/ui/scroll-area"

interface ServiceStatus {
  name: string
  status: "operational" | "degraded" | "partial" | "major"
  incidents?: number
}

interface DayStatus {
  date: string
  status: "operational" | "degraded" | "partial" | "major" | "no-data"
}

interface Incident {
  id: string
  date: string
  title: string
  status: "resolved" | "monitoring" | "investigating" | "identified"
  impact: "none" | "minor" | "major" | "critical"
  affectedServices: string[]
  updates: {
    time: string
    status: string
    message: string
  }[]
}

const services: ServiceStatus[] = [
  { name: "API", status: "operational" },
  { name: "Console", status: "operational" },
  { name: "claude.ai", status: "operational" },
  { name: "Development Environments", status: "operational" },
  { name: "Build Pipeline", status: "operational" },
]

// Generate 90 days of uptime data
const generateUptimeData = (): DayStatus[] => {
  const days = []
  const today = new Date()
  
  for (let i = 89; i >= 0; i--) {
    const date = new Date(today)
    date.setDate(date.getDate() - i)
    
    // Simulate mostly operational with occasional issues
    let status: DayStatus['status'] = "operational"
    if (i === 15) {status = "partial"}
    else if (i === 45) {status = "degraded"}
    
    days.push({
      date: date.toISOString().split('T')[0],
      status
    })
  }
  
  return days
}

const uptimeData = generateUptimeData()

const recentIncidents: Incident[] = [
  {
    id: "inc-001",
    date: "January 15, 2024",
    title: "Elevated API response times",
    status: "resolved",
    impact: "minor",
    affectedServices: ["API"],
    updates: [
      {
        time: "Jan 15, 14:45 UTC",
        status: "Resolved",
        message: "The issue has been resolved and API response times have returned to normal."
      },
      {
        time: "Jan 15, 14:30 UTC", 
        status: "Monitoring",
        message: "We've implemented a fix and are monitoring the results."
      },
      {
        time: "Jan 15, 14:15 UTC",
        status: "Identified",
        message: "We've identified the cause of elevated response times and are working on a fix."
      },
      {
        time: "Jan 15, 14:00 UTC",
        status: "Investigating",
        message: "We're investigating reports of elevated API response times."
      }
    ]
  }
]

function getStatusColor(status: string) {
  switch (status) {
    case "operational":
      return "bg-green-500"
    case "degraded":
      return "bg-yellow-500"
    case "partial":
      return "bg-orange-500"
    case "major":
      return "bg-red-500"
    case "no-data":
      return "bg-muted"
    default:
      return "bg-muted"
  }
}

function getStatusIcon(status: string, size = "h-5 w-5") {
  const className = `${size}`
  switch (status) {
    case "operational":
      return <CheckCircle2 className={`${className} text-green-500`} />
    case "degraded":
      return <Minus className={`${className} text-yellow-500 rounded-full bg-yellow-500`} />
    case "partial":
      return <AlertCircle className={`${className} text-orange-500`} />
    case "major":
      return <XCircle className={`${className} text-red-500`} />
    default:
      return <Circle className={`${className} text-muted-foreground`} />
  }
}

function getOverallStatus(services: ServiceStatus[]) {
  const hasIssues = services.some(s => s.status !== "operational")
  if (!hasIssues) {return { 
    text: "All Systems Operational", 
    status: "operational"
  }}
  
  const hasMajor = services.some(s => s.status === "major")
  if (hasMajor) {return { 
    text: "Major Outage", 
    status: "major"
  }}
  
  const hasPartial = services.some(s => s.status === "partial")
  if (hasPartial) {return { 
    text: "Partial Outage", 
    status: "partial"
  }}
  
  return { 
    text: "Minor Issues", 
    status: "degraded"
  }
}

export default function StatusPage() {
  const overallStatus = getOverallStatus(services)
  const [hoveredDay, setHoveredDay] = useState<{ service: string; day: DayStatus } | null>(null)
  
  // Calculate uptime percentage
  const operationalDays = uptimeData.filter(d => d.status === "operational").length
  const uptimePercentage = ((operationalDays / uptimeData.length) * 100).toFixed(2)
  
  return (
    <div className="flex h-screen flex-col">
      <ScrollArea className="flex-1">
        <div className="container mx-auto max-w-4xl px-6 py-12">
          {/* Logo and Title */}
          <div className="flex items-center gap-4 mb-12">
            <Link href="/" className="transition-opacity hover:opacity-80">
              <svg height="28" viewBox="0 0 628 131" fill="none" xmlns="http://www.w3.org/2000/svg" className="h-7 w-auto">
                <path d="M51.9912 66.6496C51.2636 65.9244 51.2636 64.7486 51.9912 64.0235L89.4072 26.7312C90.1348 26.006 91.3145 26.006 92.042 26.7312L129.458 64.0237C130.186 64.7489 130.186 65.9246 129.458 66.6498L92.0423 103.942C91.3147 104.667 90.135 104.667 89.4074 103.942L51.9912 66.6496Z" className="fill-primary"></path>
                <path d="M66.5331 1.04291C65.8055 0.317729 64.6259 0.317729 63.8983 1.04291L0.545688 64.186C-0.181896 64.9111 -0.181896 66.0869 0.545688 66.8121L63.8983 129.955C64.6259 130.68 65.8055 130.68 66.5331 129.955L76.9755 119.547C77.7031 118.822 77.7031 117.646 76.9755 116.921L26.4574 66.5701C25.7298 65.8449 25.7298 64.6692 26.4574 63.944L76.7327 13.8349C77.4603 13.1097 77.4603 11.934 76.7327 11.2088L66.5331 1.04291Z" className="fill-primary"></path>
                <path d="M164.241 113.166V17.8325H180.841V73.6742L201.591 45.6076H220.333L195.968 78.3597L220.868 113.166H202.126L180.841 83.4467V113.166H164.241Z" className="fill-foreground"></path>
                <path d="M588.188 86.6906C588.274 90.651 589.308 93.5352 591.288 95.3432C593.354 97.0652 596.281 97.9261 600.07 97.9261C608.077 97.9261 615.223 97.6678 621.508 97.1513L625.124 96.7638L625.382 109.549C615.481 111.96 606.527 113.165 598.52 113.165C588.791 113.165 581.731 110.582 577.34 105.416C572.949 100.251 570.754 91.8564 570.754 80.2334C570.754 57.0736 580.268 45.4937 599.295 45.4937C618.064 45.4937 627.448 55.2225 627.448 74.6802L626.157 86.6906H588.188ZM610.401 73.5179C610.401 68.3521 609.583 64.7792 607.947 62.7989C606.312 60.7326 603.427 59.6995 599.295 59.6995C595.248 59.6995 592.364 60.7757 590.642 62.9281C589.006 64.9944 588.145 68.5243 588.059 73.5179H610.401Z" className="fill-foreground"></path>
                <path d="M560.42 61.7669H544.536V88.2414C544.536 90.8243 544.579 92.6754 544.665 93.7946C544.837 94.8278 545.311 95.7318 546.086 96.5067C546.946 97.2815 548.238 97.669 549.96 97.669L559.775 97.4107L560.55 111.229C554.781 112.521 550.39 113.166 547.377 113.166C539.628 113.166 534.333 111.444 531.492 108C528.651 104.471 527.23 98.0133 527.23 88.6289V61.7669V45.4948V17.8574H544.536V45.4948H560.42V61.7669Z" className="fill-foreground"></path>
                <path d="M496.661 113.166V45.4948H513.966V113.166H496.661ZM496.661 35.421V17.8574H513.966V35.421H496.661Z" className="fill-foreground"></path>
                <path d="M466.091 113.165L466.091 17.8667H483.396L483.397 113.165H466.091Z" className="fill-foreground"></path>
                <path d="M452.826 17.8667L452.826 113.165H435.65V108.904C429.624 111.745 424.415 113.165 420.024 113.165C410.639 113.165 404.096 110.453 400.394 105.029C396.692 99.6052 394.841 91.0387 394.841 79.3296C394.841 67.5345 397.036 58.9679 401.427 53.63C405.904 48.2059 412.62 45.4939 421.574 45.4939C424.329 45.4939 428.16 45.9244 433.067 46.7854L435.521 47.3019L435.521 17.8667H452.826ZM433.713 96.1183L435.521 95.7309V61.7661C430.786 60.9051 426.567 60.4746 422.865 60.4746C415.891 60.4746 412.404 66.6735 412.404 79.0714C412.404 85.7868 413.179 90.5652 414.729 93.4063C416.279 96.2475 418.819 97.6681 422.348 97.6681C425.965 97.6681 429.753 97.1515 433.713 96.1183Z" className="fill-foreground"></path>
                <path d="M367.331 45.4937H384.636V113.165H367.46V107.999C361.261 111.443 355.88 113.165 351.317 113.165C342.363 113.165 336.337 110.711 333.237 105.804C330.138 100.81 328.588 92.5021 328.588 80.8791V45.4937H345.893V81.1374C345.893 87.5085 346.41 91.8563 347.443 94.1809C348.476 96.5055 350.973 97.6678 354.933 97.6678C358.721 97.6678 362.295 97.0652 365.652 95.8598L367.331 95.3432V45.4937Z" className="fill-foreground"></path>
                <path d="M265.823 54.4046C270.386 48.464 278.006 45.4937 288.682 45.4937C299.358 45.4937 306.977 48.464 311.54 54.4046C316.103 60.2591 318.385 68.5243 318.385 79.2002C318.385 101.844 308.484 113.165 288.682 113.165C268.88 113.165 258.979 101.844 258.979 79.2002C258.979 68.5243 261.26 60.2591 265.823 54.4046ZM279.125 93.7935C280.933 96.893 284.119 98.4427 288.682 98.4427C293.245 98.4427 296.387 96.893 298.109 93.7935C299.917 90.694 300.821 85.8296 300.821 79.2002C300.821 72.5708 299.917 67.7495 298.109 64.7361C296.387 61.7227 293.245 60.2161 288.682 60.2161C284.119 60.2161 280.933 61.7227 279.125 64.7361C277.403 67.7495 276.542 72.5708 276.542 79.2002C276.542 85.8296 277.403 90.694 279.125 93.7935Z" className="fill-foreground"></path>
                <path d="M231.468 113.165L231.071 17.8667H248.377L248.774 113.165H231.468Z" className="fill-foreground"></path>
              </svg>
            </Link>
            <h1 className="text-3xl font-light">Status</h1>
          </div>

          {/* Overall Status */}
          <div className="mb-12">
            <div className="flex items-center gap-3 mb-2">
              {getStatusIcon(overallStatus.status, "h-6 w-6")}
              <h2 className="text-2xl font-normal">{overallStatus.text}</h2>
            </div>
            <p className="text-sm text-muted-foreground">
              Last updated {new Date().toLocaleTimeString('en-US', { 
                hour: '2-digit', 
                minute: '2-digit',
                second: '2-digit',
                timeZone: 'UTC' 
              })} UTC
            </p>
          </div>

          {/* Services Status */}
          <div className="mb-12">
            <div className="space-y-3">
              {services.map((service) => (
                <div key={service.name} className="flex items-center justify-between py-3 border-b last:border-0">
                  <span className="text-base">{service.name}</span>
                  <div className="flex items-center gap-2">
                    {getStatusIcon(service.status)}
                    <span className="text-sm text-muted-foreground capitalize">{service.status}</span>
                  </div>
                </div>
              ))}
            </div>
          </div>

          {/* 90 day uptime */}
          <div className="mb-12">
            <div className="flex items-center justify-between mb-4">
              <h3 className="text-lg font-normal">90 day uptime</h3>
              <span className="text-sm text-muted-foreground">{uptimePercentage}%</span>
            </div>
            
            <div className="space-y-6">
              {services.map((service) => (
                <div key={service.name}>
                  <div className="flex items-center justify-between mb-2">
                    <span className="text-sm">{service.name}</span>
                  </div>
                  <div className="relative">
                    <div className="flex gap-[1px]">
                      {uptimeData.map((day, index) => (
                        <div
                          key={index}
                          className={`h-8 flex-1 rounded-sm cursor-pointer transition-all ${getStatusColor(day.status)} hover:scale-110 hover:rounded`}
                          onMouseEnter={() => setHoveredDay({ service: service.name, day })}
                          onMouseLeave={() => setHoveredDay(null)}
                          title={`${day.date}: ${day.status}`}
                        />
                      ))}
                    </div>
                    {hoveredDay && hoveredDay.service === service.name && (
                      <div className="absolute top-10 left-1/2 transform -translate-x-1/2 bg-popover text-popover-foreground p-2 rounded shadow-md text-xs whitespace-nowrap z-10">
                        {hoveredDay.day.date}: {hoveredDay.day.status}
                      </div>
                    )}
                  </div>
                </div>
              ))}
            </div>
            
            <div className="flex items-center gap-6 mt-4 text-xs text-muted-foreground">
              <span>90 days ago</span>
              <div className="flex items-center gap-4 flex-1">
                <div className="flex items-center gap-1">
                  <div className="w-3 h-3 rounded-sm bg-green-500" />
                  <span>No issues</span>
                </div>
                <div className="flex items-center gap-1">
                  <div className="w-3 h-3 rounded-sm bg-yellow-500" />
                  <span>Degraded</span>
                </div>
                <div className="flex items-center gap-1">
                  <div className="w-3 h-3 rounded-sm bg-orange-500" />
                  <span>Partial outage</span>
                </div>
                <div className="flex items-center gap-1">
                  <div className="w-3 h-3 rounded-sm bg-red-500" />
                  <span>Major outage</span>
                </div>
              </div>
              <span>Today</span>
            </div>
          </div>

          {/* Recent Updates */}
          <div className="mb-12">
            <h3 className="text-lg font-normal mb-6">Recent updates</h3>
            
            {recentIncidents.length === 0 ? (
              <Card>
                <CardContent className="py-8 text-center">
                  <p className="text-sm text-muted-foreground">No incidents reported in the last 90 days.</p>
                </CardContent>
              </Card>
            ) : (
              <div className="space-y-8">
                {recentIncidents.map((incident) => (
                  <div key={incident.id}>
                    <div className="flex items-start justify-between mb-4">
                      <div>
                        <h4 className="font-medium mb-1">{incident.title}</h4>
                        <p className="text-sm text-muted-foreground">{incident.date}</p>
                      </div>
                      <Badge 
                        variant={incident.status === "resolved" ? "success" : "warning"} 
                        className="text-xs"
                      >
                        {incident.status}
                      </Badge>
                    </div>
                    
                    <div className="ml-6 space-y-4">
                      {incident.updates.map((update, idx) => (
                        <div key={idx} className="relative">
                          <div className="absolute left-0 top-2 w-2 h-2 rounded-full bg-muted-foreground/20" />
                          <div className="pl-6">
                            <div className="flex items-center gap-2 mb-1">
                              <span className="text-sm font-medium">{update.status}</span>
                              <span className="text-xs text-muted-foreground">{update.time}</span>
                            </div>
                            <p className="text-sm text-muted-foreground">{update.message}</p>
                          </div>
                        </div>
                      ))}
                    </div>
                  </div>
                ))}
              </div>
            )}
          </div>

          {/* Subscribe */}
          <Card>
            <CardHeader>
              <CardTitle className="text-base font-medium">Subscribe to updates</CardTitle>
              <CardDescription className="text-sm">
                Get email notifications whenever Kloudlite creates, updates or resolves an incident.
              </CardDescription>
            </CardHeader>
            <CardContent>
              <Button variant="outline" className="w-full sm:w-auto">
                Subscribe
              </Button>
            </CardContent>
          </Card>
        </div>
      </ScrollArea>

      {/* Footer */}
      <footer className="border-t">
        <div className="container mx-auto px-6 py-4">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-6 text-sm">
              <Link href="/" className="text-muted-foreground hover:text-foreground transition-colors">
                Kloudlite
              </Link>
              <Link href="/legal/privacy" className="text-muted-foreground hover:text-foreground transition-colors">
                Privacy
              </Link>
              <Link href="/legal/terms" className="text-muted-foreground hover:text-foreground transition-colors">
                Terms
              </Link>
            </div>
            <ThemeToggle />
          </div>
        </div>
      </footer>
    </div>
  )
}