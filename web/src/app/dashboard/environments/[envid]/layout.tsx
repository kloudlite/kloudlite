"use client";

import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbLink,
  BreadcrumbList,
} from "@/components/ui/breadcrumb";
import { Button } from "@/components/ui/button";
import { DatabaseZap, FileCode2, FileKey, FileKey2, LayoutGrid, Router, Settings } from "lucide-react";
import Link from "next/link";
import { useParams, usePathname } from "next/navigation";

export default function Page({ children }: { children: React.ReactNode }) {
  const { envid } = useParams();
  const pathName = usePathname();
  return (
    <div className="p-12 flex flex-col gap-4 container mx-auto">
      <div className="flex flex-col">
        <Breadcrumb>
          <BreadcrumbList>
            <BreadcrumbItem>
              <BreadcrumbLink href="/dashboard/environments">
                Environments
              </BreadcrumbLink>
            </BreadcrumbItem>
          </BreadcrumbList>
        </Breadcrumb>
        <h1 className="text-2xl font-bold">Environment Name</h1>
      </div>
      <div className="flex">
        <div className="flex gap-2 items-center flex-1">
          <Button
            variant={pathName.startsWith(
                `/dashboard/environments/${envid}/apps`,
              )
              ? "outline"
              : "link"}
            size={"sm"}
            asChild
          >
            <Link href={`/dashboard/environments/${envid}/apps`}>
              <LayoutGrid />
              Apps
            </Link>
          </Button>
          <Button
            variant={pathName.startsWith(
                `/dashboard/environments/${envid}/helmcharts`,
              )
              ? "outline"
              : "link"}
            size={"sm"}
            asChild
          >
            <Link href={`/dashboard/environments/${envid}/helmcharts`}>
              <FileCode2 />
              Helm Charts
            </Link>
          </Button>
          <Button
            variant={pathName.startsWith(
                `/dashboard/environments/${envid}/services`,
              )
              ? "outline"
              : "link"}
            size={"sm"}
            asChild
          >
            <Link href={`/dashboard/environments/${envid}/services`}>
              <Router />
              Services
            </Link>
          </Button>
          <Button
            variant={pathName.startsWith(
                `/dashboard/environments/${envid}/imports`,
              )
              ? "outline"
              : "link"}
            size={"sm"}
            asChild
          >
            <Link href={`/dashboard/environments/${envid}/imports`}>
              <DatabaseZap />
              Imports
            </Link>
          </Button>
          <Button
            variant={pathName.startsWith(
                `/dashboard/environments/${envid}/configs-secrets`,
              )
              ? "outline"
              : "link"}
            size={"sm"}
            asChild
          >
            <Link href={`/dashboard/environments/${envid}/configs-secrets`}>
              <FileKey2 />
              Configs & Secrets
            </Link>
          </Button>
        </div>

        <Button
          variant={pathName.startsWith(
              `/dashboard/environments/${envid}/settings`,
            )
            ? "outline"
            : "link"}
          size={"sm"}
          asChild
        >
          <Link href={`/dashboard/environments/${envid}/settings`}>
          <Settings />
            Settings
          </Link>
        </Button>
      </div>
      <div>
        {children}
      </div>
    </div>
  );
}
