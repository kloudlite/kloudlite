import type { Metadata } from "next";
import { Geist, Geist_Mono } from "next/font/google";
import { cookies } from "next/headers";

import "./globals.css";
import { Toaster } from "sonner";

import { AuthProvider } from "@/components/providers/auth-provider";
import { ThemeProvider } from "@/components/theme-provider";
import { ThemeScript } from "@/components/theme-script";
import { type Theme } from "@/lib/theme";

const geistSans = Geist({
  variable: "--font-geist-sans",
  subsets: ["latin"],
});

const geistMono = Geist_Mono({
  variable: "--font-geist-mono",
  subsets: ["latin"],
});

export const metadata: Metadata = {
  title: {
    default: "Kloudlite - Development environments as a Service",
    template: "%s | Kloudlite",
  },
  description: "Instant, production-grade development environments powered by Kubernetes. Code, build, and deploy without limits.",
  keywords: ["development", "environments", "kubernetes", "cloud", "devops", "infrastructure"],
  authors: [{ name: "Kloudlite" }],
  creator: "Kloudlite",
  icons: {
    icon: "/favicon.ico",
    shortcut: "/favicon.ico",
    apple: "/favicon.ico",
  },
  openGraph: {
    type: "website",
    locale: "en_US",
    url: "https://kloudlite.io",
    siteName: "Kloudlite",
    title: "Kloudlite - Development environments as a Service",
    description: "Instant, production-grade development environments powered by Kubernetes.",
  },
  twitter: {
    card: "summary_large_image",
    title: "Kloudlite - Development environments as a Service",
    description: "Instant, production-grade development environments powered by Kubernetes.",
    creator: "@kloudlite",
  },
  robots: {
    index: true,
    follow: true,
  },
};

export default async function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  const cookieStore = await cookies();
  const themeCookie = cookieStore.get('theme');
  const theme = themeCookie ? themeCookie.value as Theme : 'system';
  
  // Resolve system theme on server side
  let resolvedTheme = theme;
  if (theme === 'system') {
    // Default to light on server for system theme
    resolvedTheme = 'light';
  }
  
  return (
    <html lang="en" className={resolvedTheme} suppressHydrationWarning>
      <head>
        <ThemeScript />
      </head>
      <body
        className={`${geistSans.variable} ${geistMono.variable} antialiased`}
      >
        <AuthProvider>
          <ThemeProvider defaultTheme={theme}>
            {children}
            <Toaster richColors closeButton />
          </ThemeProvider>
        </AuthProvider>
      </body>
    </html>
  );
}
