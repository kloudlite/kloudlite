import { isLoggedIn } from "@/actions/auth";
import { redirect } from "next/navigation";


export default async function AuthLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  if(await isLoggedIn()){
    redirect("/teams");
  }
  return (
    <div className="flex items-center justify-center h-screen">
      <div className="w-full max-w-md p-6">
        {children}
      </div>
    </div>
  );
}