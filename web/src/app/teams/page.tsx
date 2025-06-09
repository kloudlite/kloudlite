import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import {
  Pagination,
  PaginationContent,
  PaginationEllipsis,
  PaginationItem,
  PaginationLink,
  PaginationNext,
  PaginationPrevious,
} from "@/components/ui/pagination";
import {
  Table,
  TableBody,
  TableCaption,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import {
  ArrowRight,
  ChevronRight,
  Search,
  Terminal,
  Users,
} from "lucide-react";
import Link from "next/link";

export default function Home() {
  return (
    <div className="w-[700px] p-12 flex flex-col gap-4">
      <Card className="mt-20">
        <CardHeader>
          <CardTitle>
            Select Account
          </CardTitle>
          <CardDescription>Choose account to access resources</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="flex gap-4 items-center">
            <Input placeholder="Search" />
          </div>
          <div className="py-4">
            <Table className="border">
              <TableBody>
                {[1, 2].map((item) => (
                  <TableRow
                    key={item}
                    className="group"
                  >
                    <TableCell>
                      <div className="flex flex-col">
                        <div className="font-medium">Account Name</div>
                        <div className="text-muted-foreground">
                          #account_name
                        </div>
                      </div>
                    </TableCell>
                    <TableCell>
                      <div className="flex gap-2 justify-end">
                        <Button variant="default" size="sm">Accept</Button>
                        <Button variant="outline" size="sm">Reject</Button>
                      </div>
                    </TableCell>
                  </TableRow>
                ))}
                {[1, 2, 3, 4, 5].map((item) => (
                  <TableRow
                    key={item}
                    className="group"
                  >
                    <TableCell>
                      <div className="flex flex-col">
                        <div className="font-medium">Account Name</div>
                        <div className="text-muted-foreground">
                          #account_name
                        </div>
                      </div>
                    </TableCell>
                    <TableCell className="items-center justify-center group-hover:opacity-100 opacity-0 transition-all">
                      <div className="h-full flex items-center justify-end">
                        <ChevronRight size={18} />
                      </div>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </div>
          <Pagination className="justify-end !text-sm">
            <PaginationContent>
              <PaginationItem>
                <PaginationPrevious href="#" />
              </PaginationItem>
              <PaginationItem>
                <PaginationLink href="#">1</PaginationLink>
              </PaginationItem>
              <PaginationItem>
                <PaginationEllipsis />
              </PaginationItem>
              <PaginationItem>
                <PaginationNext href="#" />
              </PaginationItem>
            </PaginationContent>
          </Pagination>
        </CardContent>
      </Card>
      <div className="flex gap-4 items-center p-4 bg-slate-100 rounded-md justify-between">
        <div className="flex gap-2 items-center">
          <Users className="h-4 w-4" />
          <p className="font-medium text-sm">
            Want to use Kloudlite with a different team?
          </p>
        </div>
        <Button variant="link" className="text-sm" asChild>
          <Link href="/teams/new-team">
            <span>Create New Team</span>
            <ArrowRight className="h-4 w-4" />
          </Link>
        </Button>
      </div>
      <div className="text-center">
        Not able to see your team?
        <Button variant={"link"} className="underline text-sm">
          Try a different email
        </Button>
      </div>
    </div>
  );
}
