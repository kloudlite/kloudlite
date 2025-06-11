export default function AuthLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <div className="flex items-center justify-center h-screen">
      <div className="w-full max-w-md p-6">
        {children}
      </div>
    </div>
  );
}