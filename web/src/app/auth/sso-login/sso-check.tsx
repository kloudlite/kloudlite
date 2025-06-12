"use client";

import { loginWithSSO } from "@/actions/auth";
import { Button } from "@/components/ui/button";
import { LoaderCircle } from "lucide-react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { useEffect, useState } from "react";

export const SSOCheck = ({ token }: { token: string }) => {
  const [err, setErr] = useState<string | null>(null);
  const router = useRouter();
  useEffect(() => {
    (async () => {
      const [, err] = await loginWithSSO(token);
      if (!err) {
        router.push("/teams");
      } else {
        setErr(err.message);
      }
    })();
  }, [token]);
  return (
    <>
      {err
        ? (
          <div className="text-center">
            <p className="text-red-500 flex gap-1">
              <span className="font-medium">
                Errored while logging in with SSO:
              </span>
              {err}
            </p>
            <Button variant="link" className="p-0" asChild>
              <Link href="/dev-only/external-login">
                Please try again
              </Link>
            </Button>
          </div>
        )
        : (
          <div className="flex gap-1 items-center">
            <LoaderCircle className="animate-spin" />
            <span>Checking SSO...</span>
          </div>
        )}
    </>
  );
};
