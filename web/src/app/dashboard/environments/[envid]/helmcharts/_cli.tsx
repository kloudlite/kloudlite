"use client";

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
import { TableCell, TableRow } from "@/components/ui/table";
import { Copy, EllipsisVertical, Pause, Settings, Trash } from "lucide-react";
import { useRouter } from "next/navigation";

export const HelmChartItem = ({ item }: { item: number }) => {
  const router = useRouter();
  return (
    <ContextMenu key={item}>
      <ContextMenuTrigger asChild>
        <TableRow
          className="group cursor-pointer"
          onClick={() => {
            console.log("Row clicked", item);
            router.push(`/dashboard/environments/${item}`);
          }}
        >
          <TableCell>
            <div className="flex flex-col">
              <div className="font-medium">Helm Name</div>
              <div className="text-muted-foreground">#helm_name</div>
            </div>
          </TableCell>
          <TableCell>
            <Badge variant={"outline"}>Active</Badge>
          </TableCell>
          <TableCell className="text-muted-foreground">
            created by <span className="text-primary">karthik</span>{" "}
            <span>10 days ago</span>
          </TableCell>
          <TableCell>
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <Button
                  variant="ghost"
                  size="icon"
                  className="group-hover:opacity-100 opacity-0 transition-all"
                >
                  <EllipsisVertical size={18} />
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end">
                <DropdownMenuItem>
                  <Copy /> Clone
                </DropdownMenuItem>
                <DropdownMenuItem>
                  <Pause />
                  Pause
                </DropdownMenuItem>
                <DropdownMenuItem>
                  <Settings />
                  Settings
                </DropdownMenuItem>
                <DropdownMenuItem variant="destructive">
                  <Trash />
                  Delete
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
          </TableCell>
        </TableRow>
      </ContextMenuTrigger>
      <ContextMenuContent>
        <ContextMenuItem>
          <Copy />
          Clone
        </ContextMenuItem>
        <ContextMenuItem>
          <Pause />
          Pause
        </ContextMenuItem>
        <ContextMenuItem>
          <Settings />
          Settings
        </ContextMenuItem>
        <ContextMenuItem variant="destructive">
          <Trash />
          Delete
        </ContextMenuItem>
      </ContextMenuContent>
    </ContextMenu>
  );
};