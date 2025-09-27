import { type Metadata } from "next"

export const metadata: Metadata = {
  title: "System Status",
  description: "Real-time status and uptime information for Kloudlite services.",
}

export default function StatusLayout({
  children,
}: {
  children: React.ReactNode
}) {
  return children
}