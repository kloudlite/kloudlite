"use client"

import { Plus, ChevronDown, Server, Code } from "lucide-react"

import { Button } from "@/components/ui/button"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"

interface CreateResourceMenuProps {
  onCreateResource: (type: "environment" | "workspace") => void
}

export function CreateResourceMenu({ onCreateResource }: CreateResourceMenuProps) {
  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button className="group w-full shadow-sm hover:shadow-md transition-all duration-200 sm:w-auto">
          <Plus className="mr-2 h-4 w-4" />
          <span className="sm:hidden">Create</span>
          <span className="hidden sm:inline">Create Resource</span>
          <ChevronDown className="ml-1 h-3 w-3 transition-transform group-data-[state=open]:rotate-180" />
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end" className="w-48">
        <DropdownMenuItem 
          onClick={() => onCreateResource("environment")}
          className="cursor-pointer"
        >
          <Server className="mr-2 h-4 w-4 text-muted-foreground" />
          New Environment
        </DropdownMenuItem>
        <DropdownMenuItem 
          onClick={() => onCreateResource("workspace")}
          className="cursor-pointer"
        >
          <Code className="mr-2 h-4 w-4 text-muted-foreground" />
          New Workspace
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  )
}