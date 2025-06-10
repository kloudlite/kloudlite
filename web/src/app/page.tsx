import { Button } from "@/components/ui/button";
import Link from "next/link";

export default function Home() {
  return (
    <div className="flex items-center justify-center h-screen flex-col">
      <h1 className="text-4xl font-bold">Welcome to Kloudlite!</h1>
      <p className="mt-4 text-lg">
        Access your development and workspaces
      </p>
      <p className="mt-2 text-lg">
        <Button variant="outline" className="mt-4">
          <Link href="/auth/login">
            Login
          </Link>
        </Button>
      </p>
    </div>
  );
}