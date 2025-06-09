import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { ConfigItem } from "./_cli";
import { Button } from "@/components/ui/button";
import { EllipsisVertical, Plus } from "lucide-react";
import { Input } from "@/components/ui/input";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Badge } from "@/components/ui/badge";

export default function Page() {
  return (
    <div className="flex flex-col gap-4">
      <div className="flex items-center">
        <div className="text-2xl font-semibold py-4 flex-1">
          Configs & Secrets
        </div>
        <Button>
          <Plus className="mr-2" />
          Add Config
        </Button>
      </div>

      <div className="flex gap-2 items-center">
        <Input
          placeholder="Search Apps"
          className="max-w-[300px]"
        />
        
      </div>
      <div className="rounded-md border">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Name</TableHead>
              <TableHead>Type</TableHead>
              <TableHead>Created</TableHead>
              <TableHead></TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {[1, 2, 3].map((item) => <ConfigItem key={item} item={item} />)}
            {[1, 2, 3].map((item) => (
              <TableRow key={item} className="group">
                <TableCell>
                  <div className="flex flex-col">
                    <div className="font-medium">Name</div>
                    <div className="text-muted-foreground">#name</div>
                  </div>
                </TableCell>
                <TableCell>
                  Config
                </TableCell>
                <TableCell className="text-muted-foreground">
                  created by <span className="text-primary">karthik</span>{" "}
                  <span>10 days ago</span>
                </TableCell>
                <TableCell>
                  <Button
                    variant="ghost"
                    size="icon"
                    className="group-hover:opacity-100 opacity-0 transition-all"
                  >
                    <EllipsisVertical size={18} />
                  </Button>
                </TableCell>
              </TableRow>
            ))}
            {[1, 2, 3].map((item) => (
              <TableRow key={item} className="group">
                <TableCell>
                  <div className="flex flex-col">
                    <div className="font-medium">Name</div>
                    <div className="text-muted-foreground">#name</div>
                  </div>
                </TableCell>
                <TableCell>
                  Config
                </TableCell>
                <TableCell className="text-muted-foreground">
                  created by <span className="text-primary">karthik</span>{" "}
                  <span>10 days ago</span>
                </TableCell>
                <TableCell>
                  <Button
                    variant="ghost"
                    size="icon"
                    className="group-hover:opacity-100 opacity-0 transition-all"
                  >
                    <EllipsisVertical size={18} />
                  </Button>
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </div>
    </div>
  );
}
