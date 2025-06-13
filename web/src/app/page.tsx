import { Button } from "@/components/ui/button";
import Link from "next/link";
import { GoogleButton } from "./auth/_components/oauth-buttons";


export default function Home() {
  return (
    <div className="flex items-center justify-center h-screen flex-col">
      <h1 className="text-4xl font-bold">Welcome to Kloudlite!</h1>
      <p className="mt-4 text-lg">
        Access your development and workspaces
      </p>
      <p className="mt-2 text-lg">
        <Button variant="outline" className="mt-4">
          <Link href="/dev-only/external-login">
            SSO Login
          </Link>
        </Button>
        <GoogleButton />
      </p>
    </div>
  );
}