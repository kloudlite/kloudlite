import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  ContextMenu,
  ContextMenuContent,
  ContextMenuItem,
  ContextMenuTrigger,
} from "@/components/ui/context-menu";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Input } from "@/components/ui/input";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import {
  Copy,
  EllipsisVertical,
  Pause,
  Plus,
  Settings,
  Trash,
} from "lucide-react";
import { EnvItem } from "./_cli";

export default function Environments() {
  return (
    <div className="p-12 flex flex-col container mx-auto gap-4">
      <div className="flex items-center justify-between mb-4">
        <div className="flex flex-col">
          <h1 className="text-2xl font-bold">Environments</h1>
          <p className="text-sm text-muted-foreground">
            List of development environments and templates.
          </p>
        </div>
        <Button>
          <Plus className="mr-2" />
          Create Environment
        </Button>
      </div>
      <div className="flex flex-col gap-4">
        <div className="flex gap-2 items-center">
          <Input
            placeholder="Search Environments / Templates"
            className="max-w-[300px]"
          />
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button className="border-dashed" variant={"outline"} size={"sm"}>
                State :
                <Badge variant={"secondary"}>Active</Badge>
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="start">
              <DropdownMenuItem>
                Any
              </DropdownMenuItem>
              <DropdownMenuItem>
                Active
              </DropdownMenuItem>
              <DropdownMenuItem>
                Paused
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button variant={"outline"} size={"sm"} className="border-dashed">
                Owner :
                <Badge variant={"secondary"}>Me</Badge>
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="start">
              <DropdownMenuItem>
                Any
              </DropdownMenuItem>
              <DropdownMenuItem>
                Me
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button variant={"outline"} size={"sm"} className="border-dashed">
                Type :
                <Badge variant={"secondary"}>Environment</Badge>
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="start">
              <DropdownMenuItem>
                Any
              </DropdownMenuItem>
              <DropdownMenuItem>
                Environment
              </DropdownMenuItem>
              <DropdownMenuItem>
                Template
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        </div>
        <div className="rounded-md border">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Name</TableHead>
                <TableHead>Type</TableHead>
                <TableHead>Status</TableHead>
                <TableHead>Created</TableHead>
                <TableHead></TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {[1, 2, 3].map((item) => <EnvItem key={item} item={item} />)}
              {[1, 2, 3].map((item) => (
                <TableRow key={item} className="group">
                  <TableCell>
                    <div className="flex flex-col">
                      <div className="font-medium">Environment Name</div>
                      <div className="text-muted-foreground">#env_name</div>
                    </div>
                  </TableCell>
                  <TableCell>Template</TableCell>
                  <TableCell></TableCell>
                  <TableCell className="text-muted-foreground">
                    created by <span className="text-primary">karthik</span>
                    {" "}
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
                      <div className="font-medium">Environment Name</div>
                      <div className="text-muted-foreground">#env_name</div>
                    </div>
                  </TableCell>
                  <TableCell>Template</TableCell>
                  <TableCell></TableCell>
                  <TableCell className="text-muted-foreground">
                    created by <span className="text-primary">karthik</span>
                    {" "}
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
    </div>
  );
}
