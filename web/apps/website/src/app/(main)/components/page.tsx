"use client"

import {
  Accordion,
  AccordionContent,
  AccordionItem,
  AccordionTrigger,
  Alert,
  AlertDescription,
  AlertTitle,
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
  AspectRatio,
  Avatar,
  AvatarFallback,
  AvatarImage,
  Badge,
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbLink,
  BreadcrumbList,
  BreadcrumbPage,
  BreadcrumbSeparator,
  Button,
  ButtonGroup,
  Calendar,
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
  Checkbox,
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
  Command,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
  ContextMenu,
  ContextMenuContent,
  ContextMenuItem,
  ContextMenuTrigger,
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
  Drawer,
  DrawerContent,
  DrawerDescription,
  DrawerFooter,
  DrawerHeader,
  DrawerTitle,
  DrawerTrigger,
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
  Empty,
  EmptyHeader,
  EmptyMedia,
  EmptyTitle,
  EmptyDescription,
  EmptyContent,
  HoverCard,
  HoverCardContent,
  HoverCardTrigger,
  Input,
  InputOTP,
  InputOTPGroup,
  InputOTPSlot,
  Kbd,
  Label,
  Popover,
  PopoverContent,
  PopoverTrigger,
  Progress,
  RadioGroup,
  RadioGroupItem,
  ScrollArea,
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
  Separator,
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
  SheetTrigger,
  Skeleton,
  Slider,
  Spinner,
  Switch,
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
  Tabs,
  TabsContent,
  TabsList,
  TabsTrigger,
  Textarea,
  Toggle,
  ToggleGroup,
  ToggleGroupItem,
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@kloudlite/ui"

import {
  AlertCircle,
  Bold,
  ChevronDown,
  Home,
  Inbox,
  Italic,
  Settings,
  Underline,
  User,
} from "lucide-react"
import { useState } from "react"
import { ThemeSwitcher } from "@kloudlite/ui"

function ComponentSection({
  title,
  children,
}: {
  title: string
  children: React.ReactNode
}) {
  return (
    <div className="border border-foreground/10 rounded-lg p-6 space-y-4">
      <h2 className="text-xl font-semibold">{title}</h2>
      <div className="space-y-4">{children}</div>
    </div>
  )
}

export default function ComponentsPage() {
  const [date, setDate] = useState<Date | undefined>(new Date())
  const [sliderValue, setSliderValue] = useState([50])
  const [switchChecked, setSwitchChecked] = useState(false)
  const [checkboxChecked, setCheckboxChecked] = useState(false)

  return (
    <TooltipProvider>
      <div className="fixed top-4 right-4 z-50 bg-background/80 backdrop-blur-sm border border-foreground/10 rounded-lg p-3 shadow-lg">
        <div className="flex items-center gap-2">
          <span className="text-xs font-medium text-muted-foreground">Theme:</span>
          <ThemeSwitcher />
        </div>
      </div>
      <div className="container mx-auto py-10 px-4 max-w-6xl">
        <div className="mb-10">
          <h1 className="text-4xl font-bold mb-2">Component Library</h1>
          <p className="text-muted-foreground">
            Preview of all @kloudlite/ui components
          </p>
        </div>

        <div className="grid gap-8">
          {/* Accordion */}
          <ComponentSection title="Accordion">
            <Accordion type="single" collapsible className="w-full">
              <AccordionItem value="item-1">
                <AccordionTrigger>Is it accessible?</AccordionTrigger>
                <AccordionContent>
                  Yes. It adheres to the WAI-ARIA design pattern.
                </AccordionContent>
              </AccordionItem>
              <AccordionItem value="item-2">
                <AccordionTrigger>Is it styled?</AccordionTrigger>
                <AccordionContent>
                  Yes. It comes with default styles that match your theme.
                </AccordionContent>
              </AccordionItem>
            </Accordion>
          </ComponentSection>

          {/* Alert */}
          <ComponentSection title="Alert">
            <Alert>
              <AlertCircle className="h-4 w-4" />
              <AlertTitle>Heads up!</AlertTitle>
              <AlertDescription>
                You can add components to your app using the cli.
              </AlertDescription>
            </Alert>
            <Alert variant="destructive">
              <AlertCircle className="h-4 w-4" />
              <AlertTitle>Error</AlertTitle>
              <AlertDescription>
                Your session has expired. Please log in again.
              </AlertDescription>
            </Alert>
          </ComponentSection>

          {/* Alert Dialog */}
          <ComponentSection title="Alert Dialog">
            <AlertDialog>
              <AlertDialogTrigger asChild>
                <Button variant="outline">Open Alert Dialog</Button>
              </AlertDialogTrigger>
              <AlertDialogContent>
                <AlertDialogHeader>
                  <AlertDialogTitle>Are you absolutely sure?</AlertDialogTitle>
                  <AlertDialogDescription>
                    This action cannot be undone.
                  </AlertDialogDescription>
                </AlertDialogHeader>
                <AlertDialogFooter>
                  <AlertDialogCancel>Cancel</AlertDialogCancel>
                  <AlertDialogAction>Continue</AlertDialogAction>
                </AlertDialogFooter>
              </AlertDialogContent>
            </AlertDialog>
          </ComponentSection>

          {/* Aspect Ratio */}
          <ComponentSection title="Aspect Ratio">
            <div className="w-[300px]">
              <AspectRatio ratio={16 / 9} className="bg-muted rounded-md">
                <div className="flex items-center justify-center h-full text-muted-foreground">
                  16:9 Aspect Ratio
                </div>
              </AspectRatio>
            </div>
          </ComponentSection>

          {/* Avatar */}
          <ComponentSection title="Avatar">
            <div className="flex gap-4">
              <Avatar>
                <AvatarImage src="https://github.com/shadcn.png" alt="@shadcn" />
                <AvatarFallback>CN</AvatarFallback>
              </Avatar>
              <Avatar>
                <AvatarFallback>JD</AvatarFallback>
              </Avatar>
              <Avatar>
                <AvatarFallback>AB</AvatarFallback>
              </Avatar>
            </div>
          </ComponentSection>

          {/* Badge */}
          <ComponentSection title="Badge">
            <div className="flex gap-2 flex-wrap">
              <Badge>Default</Badge>
              <Badge variant="secondary">Secondary</Badge>
              <Badge variant="outline">Outline</Badge>
              <Badge variant="destructive">Destructive</Badge>
            </div>
          </ComponentSection>

          {/* Breadcrumb */}
          <ComponentSection title="Breadcrumb">
            <Breadcrumb>
              <BreadcrumbList>
                <BreadcrumbItem>
                  <BreadcrumbLink href="/">Home</BreadcrumbLink>
                </BreadcrumbItem>
                <BreadcrumbSeparator />
                <BreadcrumbItem>
                  <BreadcrumbLink href="/components">Components</BreadcrumbLink>
                </BreadcrumbItem>
                <BreadcrumbSeparator />
                <BreadcrumbItem>
                  <BreadcrumbPage>Breadcrumb</BreadcrumbPage>
                </BreadcrumbItem>
              </BreadcrumbList>
            </Breadcrumb>
          </ComponentSection>

          {/* Button */}
          <ComponentSection title="Button">
            <div className="flex gap-2 flex-wrap">
              <Button>Default</Button>
              <Button variant="secondary">Secondary</Button>
              <Button variant="outline">Outline</Button>
              <Button variant="ghost">Ghost</Button>
              <Button variant="link">Link</Button>
              <Button variant="destructive">Destructive</Button>
            </div>
            <div className="flex gap-2 flex-wrap">
              <Button size="sm">Small</Button>
              <Button size="default">Default</Button>
              <Button size="lg">Large</Button>
              <Button size="icon">
                <Settings className="h-4 w-4" />
              </Button>
            </div>
          </ComponentSection>

          {/* Button Group */}
          <ComponentSection title="Button Group">
            <ButtonGroup>
              <Button variant="outline">Left</Button>
              <Button variant="outline">Center</Button>
              <Button variant="outline">Right</Button>
            </ButtonGroup>
          </ComponentSection>

          {/* Calendar */}
          <ComponentSection title="Calendar">
            <Calendar
              mode="single"
              selected={date}
              onSelect={setDate}
              className="rounded-md border"
            />
          </ComponentSection>

          {/* Card */}
          <ComponentSection title="Card">
            <Card className="w-[350px]">
              <CardHeader>
                <CardTitle>Card Title</CardTitle>
                <CardDescription>Card Description</CardDescription>
              </CardHeader>
              <CardContent>
                <p>Card Content goes here.</p>
              </CardContent>
              <CardFooter>
                <Button>Action</Button>
              </CardFooter>
            </Card>
          </ComponentSection>

          {/* Checkbox */}
          <ComponentSection title="Checkbox">
            <div className="flex items-center space-x-2">
              <Checkbox
                id="terms"
                checked={checkboxChecked}
                onCheckedChange={(checked) =>
                  setCheckboxChecked(checked as boolean)
                }
              />
              <Label htmlFor="terms">Accept terms and conditions</Label>
            </div>
          </ComponentSection>

          {/* Collapsible */}
          <ComponentSection title="Collapsible">
            <Collapsible className="w-[350px] space-y-2">
              <div className="flex items-center justify-between space-x-4 px-4">
                <h4 className="text-sm font-semibold">
                  @peduarte starred 3 repositories
                </h4>
                <CollapsibleTrigger asChild>
                  <Button variant="ghost" size="sm">
                    <ChevronDown className="h-4 w-4" />
                  </Button>
                </CollapsibleTrigger>
              </div>
              <div className="rounded-md border px-4 py-2 text-sm">
                @radix-ui/primitives
              </div>
              <CollapsibleContent className="space-y-2">
                <div className="rounded-md border px-4 py-2 text-sm">
                  @radix-ui/colors
                </div>
                <div className="rounded-md border px-4 py-2 text-sm">
                  @stitches/react
                </div>
              </CollapsibleContent>
            </Collapsible>
          </ComponentSection>

          {/* Command */}
          <ComponentSection title="Command">
            <Command className="rounded-lg border shadow-md w-[350px]">
              <CommandInput placeholder="Type a command or search..." />
              <CommandList>
                <CommandEmpty>No results found.</CommandEmpty>
                <CommandGroup heading="Suggestions">
                  <CommandItem>
                    <Home className="mr-2 h-4 w-4" />
                    <span>Home</span>
                  </CommandItem>
                  <CommandItem>
                    <Settings className="mr-2 h-4 w-4" />
                    <span>Settings</span>
                  </CommandItem>
                  <CommandItem>
                    <User className="mr-2 h-4 w-4" />
                    <span>Profile</span>
                  </CommandItem>
                </CommandGroup>
              </CommandList>
            </Command>
          </ComponentSection>

          {/* Context Menu */}
          <ComponentSection title="Context Menu">
            <ContextMenu>
              <ContextMenuTrigger className="flex h-[150px] w-[300px] items-center justify-center rounded-md border border-dashed text-sm">
                Right click here
              </ContextMenuTrigger>
              <ContextMenuContent className="w-64">
                <ContextMenuItem>Back</ContextMenuItem>
                <ContextMenuItem>Forward</ContextMenuItem>
                <ContextMenuItem>Reload</ContextMenuItem>
              </ContextMenuContent>
            </ContextMenu>
          </ComponentSection>

          {/* Dialog */}
          <ComponentSection title="Dialog">
            <Dialog>
              <DialogTrigger asChild>
                <Button variant="outline">Open Dialog</Button>
              </DialogTrigger>
              <DialogContent>
                <DialogHeader>
                  <DialogTitle>Dialog Title</DialogTitle>
                  <DialogDescription>
                    This is a dialog description.
                  </DialogDescription>
                </DialogHeader>
                <div className="py-4">Dialog content goes here.</div>
                <DialogFooter>
                  <Button>Save changes</Button>
                </DialogFooter>
              </DialogContent>
            </Dialog>
          </ComponentSection>

          {/* Drawer */}
          <ComponentSection title="Drawer">
            <Drawer>
              <DrawerTrigger asChild>
                <Button variant="outline">Open Drawer</Button>
              </DrawerTrigger>
              <DrawerContent>
                <DrawerHeader>
                  <DrawerTitle>Drawer Title</DrawerTitle>
                  <DrawerDescription>Drawer description.</DrawerDescription>
                </DrawerHeader>
                <div className="p-4">Drawer content goes here.</div>
                <DrawerFooter>
                  <Button>Submit</Button>
                </DrawerFooter>
              </DrawerContent>
            </Drawer>
          </ComponentSection>

          {/* Dropdown Menu */}
          <ComponentSection title="Dropdown Menu">
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <Button variant="outline">Open Menu</Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent className="w-56">
                <DropdownMenuLabel>My Account</DropdownMenuLabel>
                <DropdownMenuSeparator />
                <DropdownMenuItem>
                  <User className="mr-2 h-4 w-4" />
                  <span>Profile</span>
                </DropdownMenuItem>
                <DropdownMenuItem>
                  <Settings className="mr-2 h-4 w-4" />
                  <span>Settings</span>
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
          </ComponentSection>

          {/* Empty */}
          <ComponentSection title="Empty State">
            <Empty>
              <EmptyHeader>
                <EmptyMedia variant="icon">
                  <Inbox className="h-6 w-6" />
                </EmptyMedia>
                <EmptyTitle>No items found</EmptyTitle>
                <EmptyDescription>Get started by creating a new item.</EmptyDescription>
              </EmptyHeader>
              <EmptyContent>
                <Button>Create Item</Button>
              </EmptyContent>
            </Empty>
          </ComponentSection>

          {/* Hover Card */}
          <ComponentSection title="Hover Card">
            <HoverCard>
              <HoverCardTrigger asChild>
                <Button variant="link">@nextjs</Button>
              </HoverCardTrigger>
              <HoverCardContent className="w-80">
                <div className="space-y-1">
                  <h4 className="text-sm font-semibold">@nextjs</h4>
                  <p className="text-sm">
                    The React Framework - created and maintained by @vercel.
                  </p>
                </div>
              </HoverCardContent>
            </HoverCard>
          </ComponentSection>

          {/* Input */}
          <ComponentSection title="Input">
            <div className="space-y-4 max-w-sm">
              <Input placeholder="Default input" />
              <Input type="email" placeholder="Email" />
              <Input type="password" placeholder="Password" />
              <Input disabled placeholder="Disabled" />
            </div>
          </ComponentSection>

          {/* Input OTP */}
          <ComponentSection title="Input OTP">
            <InputOTP maxLength={6}>
              <InputOTPGroup>
                <InputOTPSlot index={0} />
                <InputOTPSlot index={1} />
                <InputOTPSlot index={2} />
                <InputOTPSlot index={3} />
                <InputOTPSlot index={4} />
                <InputOTPSlot index={5} />
              </InputOTPGroup>
            </InputOTP>
          </ComponentSection>

          {/* Kbd */}
          <ComponentSection title="Keyboard Key (Kbd)">
            <div className="flex gap-2">
              <Kbd>⌘</Kbd>
              <Kbd>K</Kbd>
              <span className="text-muted-foreground">to open command</span>
            </div>
          </ComponentSection>

          {/* Label */}
          <ComponentSection title="Label">
            <div className="space-y-2">
              <Label htmlFor="email">Email</Label>
              <Input id="email" placeholder="Enter your email" />
            </div>
          </ComponentSection>

          {/* Popover */}
          <ComponentSection title="Popover">
            <Popover>
              <PopoverTrigger asChild>
                <Button variant="outline">Open Popover</Button>
              </PopoverTrigger>
              <PopoverContent className="w-80">
                <div className="grid gap-4">
                  <div className="space-y-2">
                    <h4 className="font-medium leading-none">Dimensions</h4>
                    <p className="text-sm text-muted-foreground">
                      Set the dimensions for the layer.
                    </p>
                  </div>
                </div>
              </PopoverContent>
            </Popover>
          </ComponentSection>

          {/* Progress */}
          <ComponentSection title="Progress">
            <Progress value={33} className="w-[60%]" />
            <Progress value={66} className="w-[60%]" />
            <Progress value={100} className="w-[60%]" />
          </ComponentSection>

          {/* Radio Group */}
          <ComponentSection title="Radio Group">
            <RadioGroup defaultValue="option-one">
              <div className="flex items-center space-x-2">
                <RadioGroupItem value="option-one" id="option-one" />
                <Label htmlFor="option-one">Option One</Label>
              </div>
              <div className="flex items-center space-x-2">
                <RadioGroupItem value="option-two" id="option-two" />
                <Label htmlFor="option-two">Option Two</Label>
              </div>
            </RadioGroup>
          </ComponentSection>

          {/* Scroll Area */}
          <ComponentSection title="Scroll Area">
            <ScrollArea className="h-[200px] w-[350px] rounded-md border p-4">
              <div className="space-y-4">
                {Array.from({ length: 20 }).map((_, i) => (
                  <p key={`scroll-item-${i + 1}`} className="text-sm">
                    Scroll item {i + 1}
                  </p>
                ))}
              </div>
            </ScrollArea>
          </ComponentSection>

          {/* Select */}
          <ComponentSection title="Select">
            <Select>
              <SelectTrigger className="w-[180px]">
                <SelectValue placeholder="Select a fruit" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="apple">Apple</SelectItem>
                <SelectItem value="banana">Banana</SelectItem>
                <SelectItem value="orange">Orange</SelectItem>
              </SelectContent>
            </Select>
          </ComponentSection>

          {/* Separator */}
          <ComponentSection title="Separator">
            <div>
              <div className="space-y-1">
                <h4 className="text-sm font-medium leading-none">
                  Radix Primitives
                </h4>
                <p className="text-sm text-muted-foreground">
                  An open-source UI component library.
                </p>
              </div>
              <Separator className="my-4" />
              <div className="flex h-5 items-center space-x-4 text-sm">
                <div>Blog</div>
                <Separator orientation="vertical" />
                <div>Docs</div>
                <Separator orientation="vertical" />
                <div>Source</div>
              </div>
            </div>
          </ComponentSection>

          {/* Sheet */}
          <ComponentSection title="Sheet">
            <Sheet>
              <SheetTrigger asChild>
                <Button variant="outline">Open Sheet</Button>
              </SheetTrigger>
              <SheetContent>
                <SheetHeader>
                  <SheetTitle>Sheet Title</SheetTitle>
                  <SheetDescription>Sheet description goes here.</SheetDescription>
                </SheetHeader>
                <div className="py-4">Sheet content goes here.</div>
              </SheetContent>
            </Sheet>
          </ComponentSection>

          {/* Skeleton */}
          <ComponentSection title="Skeleton">
            <div className="flex items-center space-x-4">
              <Skeleton className="h-12 w-12 rounded-full" />
              <div className="space-y-2">
                <Skeleton className="h-4 w-[250px]" />
                <Skeleton className="h-4 w-[200px]" />
              </div>
            </div>
          </ComponentSection>

          {/* Slider */}
          <ComponentSection title="Slider">
            <Slider
              value={sliderValue}
              onValueChange={setSliderValue}
              max={100}
              step={1}
              className="w-[60%]"
            />
            <p className="text-sm text-muted-foreground">Value: {sliderValue}</p>
          </ComponentSection>

          {/* Spinner */}
          <ComponentSection title="Spinner">
            <div className="flex gap-4 items-center">
              <Spinner className="size-3" />
              <Spinner />
              <Spinner className="size-6" />
            </div>
          </ComponentSection>

          {/* Switch */}
          <ComponentSection title="Switch">
            <div className="flex items-center space-x-2">
              <Switch
                id="airplane-mode"
                checked={switchChecked}
                onCheckedChange={setSwitchChecked}
              />
              <Label htmlFor="airplane-mode">Airplane Mode</Label>
            </div>
          </ComponentSection>

          {/* Table */}
          <ComponentSection title="Table">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Name</TableHead>
                  <TableHead>Status</TableHead>
                  <TableHead>Role</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                <TableRow>
                  <TableCell>John Doe</TableCell>
                  <TableCell>Active</TableCell>
                  <TableCell>Admin</TableCell>
                </TableRow>
                <TableRow>
                  <TableCell>Jane Smith</TableCell>
                  <TableCell>Pending</TableCell>
                  <TableCell>User</TableCell>
                </TableRow>
              </TableBody>
            </Table>
          </ComponentSection>

          {/* Tabs */}
          <ComponentSection title="Tabs">
            <Tabs defaultValue="account" className="w-[400px]">
              <TabsList>
                <TabsTrigger value="account">Account</TabsTrigger>
                <TabsTrigger value="password">Password</TabsTrigger>
              </TabsList>
              <TabsContent value="account">
                <p className="text-sm text-muted-foreground">
                  Make changes to your account here.
                </p>
              </TabsContent>
              <TabsContent value="password">
                <p className="text-sm text-muted-foreground">
                  Change your password here.
                </p>
              </TabsContent>
            </Tabs>
          </ComponentSection>

          {/* Textarea */}
          <ComponentSection title="Textarea">
            <Textarea placeholder="Type your message here." className="max-w-sm" />
          </ComponentSection>

          {/* Toggle */}
          <ComponentSection title="Toggle">
            <div className="flex gap-2">
              <Toggle aria-label="Toggle bold">
                <Bold className="h-4 w-4" />
              </Toggle>
              <Toggle aria-label="Toggle italic">
                <Italic className="h-4 w-4" />
              </Toggle>
              <Toggle aria-label="Toggle underline">
                <Underline className="h-4 w-4" />
              </Toggle>
            </div>
          </ComponentSection>

          {/* Toggle Group */}
          <ComponentSection title="Toggle Group">
            <ToggleGroup type="multiple">
              <ToggleGroupItem value="bold" aria-label="Toggle bold">
                <Bold className="h-4 w-4" />
              </ToggleGroupItem>
              <ToggleGroupItem value="italic" aria-label="Toggle italic">
                <Italic className="h-4 w-4" />
              </ToggleGroupItem>
              <ToggleGroupItem value="underline" aria-label="Toggle underline">
                <Underline className="h-4 w-4" />
              </ToggleGroupItem>
            </ToggleGroup>
          </ComponentSection>

          {/* Tooltip */}
          <ComponentSection title="Tooltip">
            <Tooltip>
              <TooltipTrigger asChild>
                <Button variant="outline">Hover me</Button>
              </TooltipTrigger>
              <TooltipContent>
                <p>Add to library</p>
              </TooltipContent>
            </Tooltip>
          </ComponentSection>
        </div>
      </div>
    </TooltipProvider>
  )
}
