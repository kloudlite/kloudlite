"use client";

import { useState } from "react";
import { Button } from "@/components/atoms";
import { 
  Dialog, 
  DialogContent, 
  DialogDescription, 
  DialogFooter, 
  DialogHeader, 
  DialogTitle,
  DialogTrigger,
  CreateServiceDialog,
  InviteMemberDialog
} from "@/components/molecules";
import { ComponentShowcase } from "../../_components/component-showcase";
import { toast } from "sonner";

export default function DialogsPage() {
  const [dialogOpen, setDialogOpen] = useState(false);

  return (
    <div className="space-y-8">
      <div>
        <h1 className="text-2xl font-bold text-slate-900 dark:text-white mb-4">
          Dialogs & Modals
        </h1>
        <p className="text-slate-600 dark:text-slate-400">
          Modal dialog components for user interactions.
        </p>
      </div>

      <ComponentShowcase
        title="Basic Dialog"
        description="Standard dialog with header, content, and footer"
      >
        <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
          <DialogTrigger asChild>
            <Button>Open Dialog</Button>
          </DialogTrigger>
          <DialogContent>
            <DialogHeader>
              <DialogTitle>Edit Profile</DialogTitle>
              <DialogDescription>
                Make changes to your profile here. Click save when you're done.
              </DialogDescription>
            </DialogHeader>
            <div className="py-4">
              <p className="text-sm text-slate-600 dark:text-slate-400">
                Dialog content goes here. You can add forms, text, or any other content.
              </p>
            </div>
            <DialogFooter>
              <Button variant="outline" onClick={() => setDialogOpen(false)}>
                Cancel
              </Button>
              <Button onClick={() => {
                setDialogOpen(false);
                toast.success("Changes saved!");
              }}>
                Save changes
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      </ComponentShowcase>

      <ComponentShowcase
        title="Confirmation Dialogs"
        description="Dialogs for confirming destructive actions"
      >
        <div className="flex gap-4">
          <Dialog>
            <DialogTrigger asChild>
              <Button variant="destructive">Delete Item</Button>
            </DialogTrigger>
            <DialogContent>
              <DialogHeader>
                <DialogTitle>Are you absolutely sure?</DialogTitle>
                <DialogDescription>
                  This action cannot be undone. This will permanently delete your
                  item and remove it from our servers.
                </DialogDescription>
              </DialogHeader>
              <DialogFooter>
                <Button variant="outline" onClick={() => {}}>
                  Cancel
                </Button>
                <Button variant="destructive" onClick={() => toast.success("Item deleted")}>
                  Delete
                </Button>
              </DialogFooter>
            </DialogContent>
          </Dialog>

          <Dialog>
            <DialogTrigger asChild>
              <Button variant="outline">Reset Settings</Button>
            </DialogTrigger>
            <DialogContent>
              <DialogHeader>
                <DialogTitle>Reset to defaults?</DialogTitle>
                <DialogDescription>
                  This will reset all settings to their default values. Your data will not be affected.
                </DialogDescription>
              </DialogHeader>
              <DialogFooter>
                <Button variant="outline">
                  Cancel
                </Button>
                <Button onClick={() => toast.success("Settings reset")}>
                  Reset
                </Button>
              </DialogFooter>
            </DialogContent>
          </Dialog>
        </div>
      </ComponentShowcase>

      <ComponentShowcase
        title="Create Service Dialog"
        description="Specialized dialog for creating services"
      >
        <CreateServiceDialog
          onCreateService={(type, name) => {
            toast.success(`Creating ${type} service: ${name}`);
          }}
        />
      </ComponentShowcase>

      <ComponentShowcase
        title="Invite Member Dialog"
        description="Dialog for inviting team members"
      >
        <InviteMemberDialog
          onInviteMember={(email, role) => {
            toast.success(`Invitation sent to ${email} as ${role}`);
          }}
        />
      </ComponentShowcase>
    </div>
  );
}