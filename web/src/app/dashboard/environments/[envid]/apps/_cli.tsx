"use client";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import {
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
} from "@/components/ui/collapsible";
import {
  Command,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
} from "@/components/ui/command";
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
import {
  Form,
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@/components/ui/form";
import { Input } from "@/components/ui/input";
import { Popover, PopoverContent } from "@/components/ui/popover";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectLabel,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  Sheet,
  SheetContent,
  SheetHeader,
  SheetTitle,
  SheetTrigger,
} from "@/components/ui/sheet";
import { Slider } from "@/components/ui/slider";
import { Table, TableBody, TableCell, TableHeader, TableRow } from "@/components/ui/table";
import { Tabs, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { cn } from "@/lib/utils";
import { PopoverTrigger } from "@radix-ui/react-popover";
import { SelectGroup } from "@radix-ui/react-select";
import { TabsContent } from "@radix-ui/react-tabs";
import {
  ArrowDownLeft,
  ArrowLeft,
  ArrowRight,
  Copy,
  EllipsisVertical,
  Pause,
  Plus,
  Settings,
  Trash,
  X,
} from "lucide-react";
import { useRouter } from "next/navigation";
import { useState } from "react";
import { useForm } from "react-hook-form";

const FormSection = (
  { children, index, title, openIndex, currentIndex }: {
    children: React.ReactNode;
    index: number;
    title: string;
    currentIndex: number;
    openIndex: (index: number) => void;
  },
) => {
  const active = currentIndex >= index;
  console.log(currentIndex, index);

  return (
    <Collapsible className="relative py-3" open={currentIndex === index}>
      <div className="absolute top-[40px] left-[12px] bottom-[-18px] border-l border-dashed z-0">
      </div>
      <CollapsibleTrigger
        className={cn(
          "flex font-medium z-10 sticky top-0 w-full",
          {
            "opacity-30": !active,
            "opacity-100": active,
          },
        )}
        onClick={() => {
          if (active) {
            openIndex(index);
          }
        }}
      >
        <div className="flex flex-col w-full">
          <div className="flex gap-2 items-center flex-1 text-left pt-4">
            <div className="rounded-full bg-secondary block p-1 text-xs w-[24px] text-center">
              {index}
            </div>
            <span className=" bg-white flex-1">
              {title}
            </span>
          </div>
        </div>
      </CollapsibleTrigger>
      <CollapsibleContent className="flex flex-col gap-4 items-start pl-8 pt-6">
        {children}
      </CollapsibleContent>
    </Collapsible>
  );
};

export const AppItem = ({ item }: { item: number }) => {
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
              <div className="font-medium">App Name</div>
              <div className="text-muted-foreground">#app_name</div>
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

export const CreateApp = () => {
  const form = useForm();
  const [currentSection, setCurrentSection] = useState(1);
  return (
    <Sheet>
      <SheetTrigger asChild>
        <Button>
          <Plus className="mr-2" />
          Create App
        </Button>
      </SheetTrigger>
      <SheetContent
        className="h-screen px-12 flex flex-col gap-4 container mx-auto"
        side="bottom"
      >
        <SheetHeader>
          <SheetTitle className="hidden">Create App</SheetTitle>
        </SheetHeader>
        <div className="flex flex-col bg-white w-full">
          <h1 className="text-2xl font-bold">Create App</h1>
        </div>
        <div className="flex-1 overflow-y-auto">
          <Form {...form}>
            <div>
              <FormSection
                index={1}
                currentIndex={currentSection}
                title="Application Details"
                openIndex={setCurrentSection}
              >
                <FormField
                  control={form.control}
                  name="plan"
                  render={() => {
                    return (
                      <FormItem className="w-[400px]">
                        <FormLabel>
                          Application Name
                        </FormLabel>
                        <FormControl>
                          <Input
                            type="text"
                            placeholder="Application Name"
                            className="border p-2 rounded-md w-full"
                            {...form.register("email", {
                              required: "App name is required",
                              maxLength: 32,
                              pattern: {
                                value: /^[a-zA-Z0-9._%+-]$/,
                                message: "Invalid app name",
                              },
                            })}
                          />
                        </FormControl>
                        <FormDescription />
                        <FormMessage />
                      </FormItem>
                    );
                  }}
                />
                <FormField
                  control={form.control}
                  name="Image"
                  render={() => {
                    return (
                      <FormItem className="w-[400px]">
                        <FormLabel>
                          Image
                        </FormLabel>
                        <FormControl>
                          <Input
                            type="text"
                            placeholder="Image"
                            className="border p-2 rounded-md w-full"
                            {...form.register("email", {
                              required: "Image is required",
                            })}
                          />
                        </FormControl>
                        <FormDescription />
                        <FormMessage />
                      </FormItem>
                    );
                  }}
                />
                <Button onClick={() => setCurrentSection(2)}>
                  Continue <ArrowRight />
                </Button>
              </FormSection>
              <FormSection
                index={2}
                currentIndex={currentSection}
                title="Compute"
                openIndex={setCurrentSection}
              >
                <div>
                  <Tabs defaultValue="quick">
                    <TabsList>
                      <TabsTrigger value="quick">Quick</TabsTrigger>
                      <TabsTrigger value="manual">Manual</TabsTrigger>
                    </TabsList>
                    <TabsContent value="quick" className="pt-3">
                      <FormField
                        control={form.control}
                        name="plan"
                        render={() => {
                          return (
                            <FormItem className="w-[400px]">
                              <FormLabel>
                                Plan
                              </FormLabel>
                              <FormControl>
                                <Select>
                                  <SelectTrigger className="w-[400px]">
                                    <SelectValue placeholder="Select a plan" />
                                  </SelectTrigger>
                                  <SelectContent>
                                    <SelectGroup>
                                      <SelectLabel>
                                        Shared Compute
                                      </SelectLabel>
                                      <SelectItem value="shared-general-purpose">
                                        General Purpose
                                      </SelectItem>
                                      <SelectItem value="shared-memory-optimised">
                                        Memory Optimised
                                      </SelectItem>
                                    </SelectGroup>
                                    <SelectGroup>
                                      <SelectLabel>
                                        Dedicated Compute
                                      </SelectLabel>
                                      <SelectItem value="dedicated-general-purpose">
                                        General Purpose
                                      </SelectItem>
                                      <SelectItem value="dedicated-memory-optimised">
                                        CPU Optimised
                                      </SelectItem>
                                      <SelectItem value="dedicated-memory-optimised">
                                        Memory Optimised
                                      </SelectItem>
                                    </SelectGroup>
                                  </SelectContent>
                                </Select>
                              </FormControl>
                              <FormDescription />
                              <FormMessage />
                            </FormItem>
                          );
                        }}
                      />
                      <FormField
                        control={form.control}
                        name="plan"
                        render={() => {
                          return (
                            <Card>
                              <CardContent>
                                <FormItem className="w-[400px]">
                                  <FormLabel className="flex items-center">
                                    <span className="flex-1">Size</span>
                                    <span className="text-muted-foreground text-xs ml-2">
                                      1 vCPU & 2 GB Memory
                                    </span>
                                  </FormLabel>
                                  <FormControl>
                                    <Slider
                                      defaultValue={[50]}
                                      max={100}
                                      step={1}
                                    />
                                  </FormControl>
                                  <FormDescription />
                                  <FormMessage />
                                </FormItem>
                              </CardContent>
                            </Card>
                          );
                        }}
                      />
                    </TabsContent>
                    <TabsContent value="manual" className="pt-3">
                      <div className="grid grid-cols-2 gap-4 w-[600px]">
                        <FormField
                          control={form.control}
                          name="cpu_request"
                          render={() => {
                            return (
                              <FormItem>
                                <FormLabel>
                                  CPU Request
                                </FormLabel>
                                <FormControl>
                                  <Input type="number" prefix="m" />
                                </FormControl>
                                <FormDescription />
                                <FormMessage />
                              </FormItem>
                            );
                          }}
                        />
                        <FormField
                          control={form.control}
                          name="cpu_request"
                          render={() => {
                            return (
                              <FormItem>
                                <FormLabel>
                                  CPU Limit
                                </FormLabel>
                                <FormControl>
                                  <Input type="number" prefix="m" />
                                </FormControl>
                                <FormDescription />
                                <FormMessage />
                              </FormItem>
                            );
                          }}
                        />
                        <FormField
                          control={form.control}
                          name="cpu_request"
                          render={() => {
                            return (
                              <FormItem>
                                <FormLabel>
                                  Memory Request
                                </FormLabel>
                                <FormControl>
                                  <Input type="number" prefix="m" />
                                </FormControl>
                                <FormDescription />
                                <FormMessage />
                              </FormItem>
                            );
                          }}
                        />
                        <FormField
                          control={form.control}
                          name="cpu_request"
                          render={() => {
                            return (
                              <FormItem>
                                <FormLabel>
                                  Memory Limit
                                </FormLabel>
                                <FormControl>
                                  <Input type="number" prefix="m" />
                                </FormControl>
                                <FormDescription />
                                <FormMessage />
                              </FormItem>
                            );
                          }}
                        />
                      </div>
                    </TabsContent>
                  </Tabs>
                </div>
                <div className="flex gap-2">
                  <Button
                    variant={"outline"}
                    onClick={() => setCurrentSection(1)}
                  >
                    <ArrowLeft /> App Details
                  </Button>
                  <Button onClick={() => setCurrentSection(3)}>
                    Continue <ArrowRight />
                  </Button>
                </div>
              </FormSection>
              <FormSection
                index={3}
                currentIndex={currentSection}
                title="Env Vars"
                openIndex={setCurrentSection}
              >
                <Card>
                  <CardHeader>
                    <CardTitle>Environment Variables</CardTitle>
                  </CardHeader>
                  <CardContent>
                    <div className="grid grid-cols-2 gap-4 w-[600px] pb-2">
                      <FormField
                        control={form.control}
                        name="key"
                        render={() => {
                          return (
                            <FormItem>
                              <FormLabel>
                                Key
                              </FormLabel>
                              <FormControl>
                                <Input
                                  type="text"
                                  placeholder="Key"
                                  className="border p-2 rounded-md w-full"
                                  {...form.register("email", {
                                    required: "App name is required",
                                    maxLength: 32,
                                    pattern: {
                                      value: /^[a-zA-Z0-9._%+-]$/,
                                      message: "Invalid app name",
                                    },
                                  })}
                                />
                              </FormControl>
                              <FormDescription />
                              <FormMessage />
                            </FormItem>
                          );
                        }}
                      />
                      <FormField
                        control={form.control}
                        name="value"
                        render={() => {
                          return (
                            <FormItem>
                              <FormLabel>
                                Value
                              </FormLabel>
                              <FormControl>
                                <div className="flex relative items-center">
                                  <Input
                                    type="text"
                                    placeholder="Value"
                                    className="border p-2 rounded-md w-full"
                                    {...form.register("email", {
                                      required: "App name is required",
                                      maxLength: 32,
                                      pattern: {
                                        value: /^[a-zA-Z0-9._%+-]$/,
                                        message: "Invalid app name",
                                      },
                                    })}
                                  />
                                  <Popover>
                                    <PopoverTrigger asChild>
                                      <Button
                                        className="absolute right-1"
                                        size={"icon"}
                                        variant={"link"}
                                      >
                                        <ArrowDownLeft />
                                      </Button>
                                    </PopoverTrigger>
                                    <PopoverContent
                                      align="end"
                                      className="p-0 mt-1"
                                    >
                                      <Command>
                                        <CommandInput placeholder="Search" />
                                        <CommandList>
                                          <CommandEmpty>
                                            No results found.
                                          </CommandEmpty>
                                          <CommandGroup heading="Imported Resources">
                                            <CommandItem value="1">
                                              Database / CONNECTION_STRING
                                            </CommandItem>
                                            <CommandItem value="2">
                                              Database / USERNAME
                                            </CommandItem>
                                            <CommandItem value="3">
                                              Database / PASSWORD
                                            </CommandItem>
                                          </CommandGroup>
                                          <CommandGroup heading="Configs">
                                            <CommandItem value="4">
                                              Sample / isDev
                                            </CommandItem>
                                          </CommandGroup>
                                        </CommandList>
                                      </Command>
                                    </PopoverContent>
                                  </Popover>
                                </div>
                              </FormControl>
                              <FormDescription />
                              <FormMessage />
                            </FormItem>
                          );
                        }}
                      />
                      <div className="text-xs">
                        All config entries be mounted on path specified in the
                        container.
                      </div>
                      <div className="flex justify-end">
                        <Button variant={"outline"}>Add Entry</Button>
                      </div>
                    </div>
                    <div className="border rounded">
                      <div className="p-2 bg-muted text-sm flex items-center">
                        <span className="flex-1">
                          Variables
                        </span>
                        <span className="flex items-center gap-2">
                          <ArrowLeft size={12} />
                          <ArrowRight size={12} />
                        </span>
                      </div>
                      <Table>
                        <TableBody>
                          {[1, 2, 3].map((i) => {
                            return (
                              <TableRow className="group" key={i}>
                                <TableCell>KEY</TableCell>
                                <TableCell>KEY</TableCell>
                                <TableCell>
                                  <div className="flex justify-end text-right w-full">
                                    <Button size={"icon"} variant="ghost">
                                      <X size={12} className="group-hover:" />
                                    </Button>
                                  </div>
                                </TableCell>
                              </TableRow>
                            );
                          })}
                        </TableBody>
                      </Table>
                    </div>
                  </CardContent>
                </Card>

                <div className="flex gap-2">
                  <Button
                    variant={"outline"}
                    onClick={() => setCurrentSection(2)}
                  >
                    <ArrowLeft /> Compute
                  </Button>
                  <Button onClick={() => setCurrentSection(4)}>
                    Continue <ArrowRight />
                  </Button>
                </div>
              </FormSection>
              <FormSection
                index={4}
                currentIndex={currentSection}
                title="Config Mounts"
                openIndex={setCurrentSection}
              >
                <Card>
                  <CardHeader>
                    <CardTitle>Config Mounts</CardTitle>
                  </CardHeader>
                  <CardContent>
                    <div className="grid grid-cols-2 gap-4 w-[600px] pb-2">
                      <FormField
                        control={form.control}
                        name="key"
                        render={() => {
                          return (
                            <FormItem>
                              <FormLabel>
                                Path
                              </FormLabel>
                              <FormControl>
                                <Input
                                  type="text"
                                  placeholder="Mount Path"
                                  className="border p-2 rounded-md w-full"
                                  {...form.register("email", {
                                    required: "App name is required",
                                    maxLength: 32,
                                    pattern: {
                                      value: /^[a-zA-Z0-9._%+-]$/,
                                      message: "Invalid app name",
                                    },
                                  })}
                                />
                              </FormControl>
                              <FormDescription />
                              <FormMessage />
                            </FormItem>
                          );
                        }}
                      />
                      <FormField
                        control={form.control}
                        name="value"
                        render={() => {
                          return (
                            <FormItem>
                              <FormLabel>
                                Value
                              </FormLabel>
                              <FormControl>
                                <div className="flex relative items-center">
                                  <Input
                                    type="text"
                                    placeholder="Value"
                                    className="border p-2 rounded-md w-full"
                                    {...form.register("email", {
                                      required: "App name is required",
                                      maxLength: 32,
                                      pattern: {
                                        value: /^[a-zA-Z0-9._%+-]$/,
                                        message: "Invalid app name",
                                      },
                                    })}
                                  />
                                  <Button
                                    className="absolute right-0"
                                    size={"sm"}
                                    variant={"ghost"}
                                  >
                                    <ArrowDownLeft />
                                  </Button>
                                </div>
                              </FormControl>
                              <FormDescription />
                              <FormMessage />
                            </FormItem>
                          );
                        }}
                      />
                      <div className="text-xs">
                        All config entries be mounted on path specified in the
                        container.
                      </div>
                      <div className="flex justify-end">
                        <Button variant={"outline"}>Add Entry</Button>
                      </div>
                    </div>
                    <hr />
                    <div>
                      <Table>
                        <TableBody>
                          <TableRow className="group">
                            <TableCell>KEY</TableCell>
                            <TableCell>KEY</TableCell>
                            <TableCell>
                              <div className="flex justify-end text-right w-full">
                                <Button size={"icon"} variant="ghost">
                                  <X size={14} className="group-hover:" />
                                </Button>
                              </div>
                            </TableCell>
                          </TableRow>
                        </TableBody>
                      </Table>
                    </div>
                  </CardContent>
                </Card>

                <div className="flex gap-2">
                  <Button
                    variant={"outline"}
                    onClick={() => setCurrentSection(3)}
                  >
                    <ArrowLeft /> Compute
                  </Button>
                  <Button onClick={() => setCurrentSection(5)}>
                    Continue <ArrowRight />
                  </Button>
                </div>
              </FormSection>
              <FormSection
                index={5}
                currentIndex={currentSection}
                title="Network"
                openIndex={setCurrentSection}
              >
                <div className="text-muted-foreground text-sm">
                  The application streamlines project management through
                  intuitive task tracking and collaboration tools.
                </div>
                <FormField
                  control={form.control}
                  name="email"
                  render={() => {
                    return (
                      <FormItem className="w-[400px]">
                        <FormLabel>
                          Name
                        </FormLabel>
                        <FormControl>
                          <Input
                            type="text"
                            placeholder="Application Name"
                            className="border p-2 rounded-md w-full"
                            {...form.register("email", {
                              required: "App name is required",
                              maxLength: 32,
                              pattern: {
                                value: /^[a-zA-Z0-9._%+-]$/,
                                message: "Invalid app name",
                              },
                            })}
                          />
                        </FormControl>
                        <FormDescription />
                        <FormMessage />
                      </FormItem>
                    );
                  }}
                />
                <FormField
                  control={form.control}
                  name="Image"
                  render={() => {
                    return (
                      <FormItem className="w-[400px]">
                        <FormLabel>
                          Image
                        </FormLabel>
                        <FormControl>
                          <Input
                            type="text"
                            placeholder="Image"
                            className="border p-2 rounded-md w-full"
                            {...form.register("email", {
                              required: "Image is required",
                            })}
                          />
                        </FormControl>
                        <FormDescription />
                        <FormMessage />
                      </FormItem>
                    );
                  }}
                />
                <div className="flex gap-2">
                  <Button
                    variant={"outline"}
                    onClick={() => setCurrentSection(4)}
                  >
                    <ArrowLeft /> Compute
                  </Button>
                  <Button onClick={() => setCurrentSection(6)}>
                    Continue <ArrowRight />
                  </Button>
                </div>
              </FormSection>
              <FormSection
                index={6}
                currentIndex={currentSection}
                title="Review"
                openIndex={setCurrentSection}
              >
                <div className="text-muted-foreground text-sm">
                  The application streamlines project management through
                  intuitive task tracking and collaboration tools.
                </div>
                <FormField
                  control={form.control}
                  name="email"
                  render={() => {
                    return (
                      <FormItem className="w-[400px]">
                        <FormLabel>
                          Name
                        </FormLabel>
                        <FormControl>
                          <Input
                            type="text"
                            placeholder="Application Name"
                            className="border p-2 rounded-md w-full"
                            {...form.register("email", {
                              required: "App name is required",
                              maxLength: 32,
                              pattern: {
                                value: /^[a-zA-Z0-9._%+-]$/,
                                message: "Invalid app name",
                              },
                            })}
                          />
                        </FormControl>
                        <FormDescription />
                        <FormMessage />
                      </FormItem>
                    );
                  }}
                />
                <FormField
                  control={form.control}
                  name="Image"
                  render={() => {
                    return (
                      <FormItem className="w-[400px]">
                        <FormLabel>
                          Image
                        </FormLabel>
                        <FormControl>
                          <Input
                            type="text"
                            placeholder="Image"
                            className="border p-2 rounded-md w-full"
                            {...form.register("email", {
                              required: "Image is required",
                            })}
                          />
                        </FormControl>
                        <FormDescription />
                        <FormMessage />
                      </FormItem>
                    );
                  }}
                />
                <Button>Submit</Button>
              </FormSection>
            </div>
          </Form>
        </div>
      </SheetContent>
    </Sheet>
  );
};
