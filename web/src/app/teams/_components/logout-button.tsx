"use client";
import { logout } from "@/actions/auth";
import { Button } from "@/components/ui/button";
import { useRouter } from "next/navigation";
import { useCallback } from "react";

export const LogoutButton = () => {
  const router =useRouter();
  const logoutHandler = useCallback(
    async () => {
      await logout();
      router.push("/auth/login");
    },
    [],
  );
  return (
    <Button variant={"link"} className="underline text-sm cursor-pointer" onClick={logoutHandler}>
      Try a different email
    </Button>
  );
};
