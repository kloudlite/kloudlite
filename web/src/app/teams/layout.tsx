import { isLoggedIn } from "@/actions/auth";
import { redirect } from "next/navigation";

export default async function TeamsLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  if(await isLoggedIn()){
    return children;
  }
  redirect("/auth/login");
}