import { AuthSessionProvider } from "@/components/registration/session-provider";

export default function RegisterLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <AuthSessionProvider>
      {children}
    </AuthSessionProvider>
  );
}
