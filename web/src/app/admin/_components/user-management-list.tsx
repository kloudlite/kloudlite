'use client'

import { useState, useTransition } from 'react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import {
  Plus,
  MoreHorizontal,
  Edit,
  Trash2,
  Key,
  Mail,
  Loader2,
  UserCheck,
  UserX,
} from 'lucide-react'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { createUser, updateUser, deleteUser, resetUserPassword } from '@/lib/actions/user-actions'
import { UserDisplay, CreateUserFormData } from '@/types/user'
import { toast } from 'sonner'

// Helper function to get available roles based on current user's role
function getAvailableRoles(currentUserRole: 'super-admin' | 'admin'): string[] {
  if (currentUserRole === 'super-admin') {
    return ['user', 'admin', 'super-admin']
  } else {
    // Admin can only create regular users
    return ['user']
  }
}

interface UserManagementListProps {
  users: UserDisplay[]
  currentUserRole: 'super-admin' | 'admin'
}

export function UserManagementList({
  users: initialUsers,
  currentUserRole,
}: UserManagementListProps) {
  const [users, setUsers] = useState(initialUsers)
  const [roleFilter, setRoleFilter] = useState<'all' | 'super-admin' | 'admin' | 'user'>('all')
  const [statusFilter, setStatusFilter] = useState<'all' | 'active'>('all')
  const [searchQuery, setSearchQuery] = useState('')
  const [isAddUserOpen, setIsAddUserOpen] = useState(false)
  const [editingUser, setEditingUser] = useState<UserDisplay | null>(null)
  const [deletingUser, setDeletingUser] = useState<UserDisplay | null>(null)
  const [resettingPasswordUser, setResettingPasswordUser] = useState<UserDisplay | null>(null)
  const [newPassword, setNewPassword] = useState('')
  const [isPending, startTransition] = useTransition()
  const [formError, setFormError] = useState<string | null>(null)
  const [hasAttemptedSubmit, setHasAttemptedSubmit] = useState(false)

  // Form state
  const [formData, setFormData] = useState<CreateUserFormData>({
    email: '',
    displayName: '',
    roles: [],
  })

  const resetForm = () => {
    setFormData({
      email: '',
      displayName: '',
      roles: [],
    })
    setFormError(null)
    setHasAttemptedSubmit(false)
    setEditingUser(null)
    setIsAddUserOpen(false)
  }

  const resetDeleteDialog = () => {
    setDeletingUser(null)
  }

  const resetPasswordDialog = () => {
    setResettingPasswordUser(null)
    setNewPassword('')
  }

  const handleEditUser = (user: UserDisplay) => {
    // Clear any previous errors first
    setFormError(null)
    setHasAttemptedSubmit(false)

    setEditingUser(user)

    // Parse roles safely with fallback
    let roles: string[] = []
    if (user.role && typeof user.role === 'string') {
      roles = user.role.split(', ').filter((role) => role.trim() !== '')
    }

    setFormData({
      email: user.email,
      displayName: user.displayName || '',
      roles: roles.length > 0 ? roles : ['user'], // Ensure at least one role
    })
  }

  // Helper function to check if current user can edit another user
  const canEditUser = (targetUser: UserDisplay): boolean => {
    if (currentUserRole === 'super-admin') {
      return true // Super admin can edit anyone
    }

    if (currentUserRole === 'admin') {
      // Admin can only edit regular users, not other admins or super-admins
      const targetRoles = targetUser.role.split(', ')
      return !targetRoles.includes('admin') && !targetRoles.includes('super-admin')
    }

    return false
  }

  const handleSubmit = async () => {
    // Mark that user has attempted to submit
    setHasAttemptedSubmit(true)

    // Clear previous errors
    setFormError(null)

    // Validate required fields
    if (!formData.email) {
      setFormError('Email is required')
      return
    }

    if (!formData.roles || formData.roles.length === 0) {
      setFormError('At least one role is required')
      return
    }

    startTransition(async () => {
      try {
        if (editingUser) {
          // Update existing user
          const result = await updateUser(editingUser.id, formData)
          if (result.success && result.user) {
            setUsers((prev) => prev.map((u) => (u.id === editingUser.id ? result.user! : u)))
            toast.success('User updated successfully')
            resetForm()
          } else {
            setFormError(result.error || 'Failed to update user')
            toast.error(result.error || 'Failed to update user')
          }
        } else {
          // Create new user
          const result = await createUser(formData)
          if (result.success && result.user) {
            setUsers((prev) => [...prev, result.user!])
            toast.success('User created successfully')
            resetForm()
          } else {
            setFormError(result.error || 'Failed to create user')
            toast.error(result.error || 'Failed to create user')
          }
        }
      } catch (error) {
        const errorMessage = 'An unexpected error occurred'
        setFormError(errorMessage)
        toast.error(errorMessage)
        console.error('Form submission error:', error)
      }
    })
  }

  const handleDeleteUser = async () => {
    if (!deletingUser) return

    startTransition(async () => {
      try {
        const result = await deleteUser(deletingUser.id)
        if (result.success) {
          setUsers((prev) => prev.filter((u) => u.id !== deletingUser.id))
          toast.success('User deleted successfully')
          resetDeleteDialog()
        } else {
          toast.error(result.error || 'Failed to delete user')
        }
      } catch (error) {
        toast.error('An unexpected error occurred')
        console.error('Delete error:', error)
      }
    })
  }

  const handleResetPassword = async () => {
    if (!resettingPasswordUser || !newPassword) return

    if (newPassword.length < 8) {
      toast.error('Password must be at least 8 characters long')
      return
    }

    startTransition(async () => {
      try {
        const result = await resetUserPassword(resettingPasswordUser.id, newPassword)
        if (result.success) {
          toast.success('Password reset successfully')
          resetPasswordDialog()
        } else {
          toast.error(result.error || 'Failed to reset password')
        }
      } catch (error) {
        toast.error('An unexpected error occurred')
        console.error('Reset password error:', error)
      }
    })
  }

  const handleToggleUserStatus = async (user: UserDisplay, enable: boolean) => {
    startTransition(async () => {
      try {
        const action = enable ? 'activate' : 'deactivate'
        const result = await updateUser(user.id, { isActive: enable })

        if (result.success) {
          // Update the user status in the local state
          setUsers((prev) =>
            prev.map((u) =>
              u.id === user.id ? { ...u, status: enable ? 'active' : 'inactive' } : u,
            ),
          )
          toast.success(`User ${enable ? 'enabled' : 'disabled'} successfully`)
        } else {
          toast.error(result.error || `Failed to ${action} user`)
        }
      } catch (error) {
        toast.error('An unexpected error occurred')
        console.error('Toggle user status error:', error)
      }
    })
  }

  let filteredUsers = users

  // Apply role filter
  if (roleFilter !== 'all') {
    filteredUsers = filteredUsers.filter((user) => user.role === roleFilter)
  }

  // Apply status filter
  if (statusFilter === 'active') {
    filteredUsers = filteredUsers.filter((user) => user.status === 'active')
  }

  // Apply search
  if (searchQuery) {
    filteredUsers = filteredUsers.filter(
      (user) =>
        user.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
        user.email.toLowerCase().includes(searchQuery.toLowerCase()),
    )
  }

  return (
    <div className="space-y-4">
      {/* Filter and Actions */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-4">
          {/* Role Filter */}
          <div className="flex items-center gap-1 rounded-md bg-muted p-1">
            <button
              onClick={() => setRoleFilter('all')}
              className={`rounded px-3 py-1 text-sm transition-colors ${
                roleFilter === 'all'
                  ? 'bg-card text-foreground shadow-sm'
                  : 'text-muted-foreground hover:text-foreground'
              }`}
            >
              All
            </button>
            <button
              onClick={() => setRoleFilter('admin')}
              className={`rounded px-3 py-1 text-sm transition-colors ${
                roleFilter === 'admin'
                  ? 'bg-card text-foreground shadow-sm'
                  : 'text-muted-foreground hover:text-foreground'
              }`}
            >
              Admin
            </button>
            <button
              onClick={() => setRoleFilter('user')}
              className={`rounded px-3 py-1 text-sm transition-colors ${
                roleFilter === 'user'
                  ? 'bg-card text-foreground shadow-sm'
                  : 'text-muted-foreground hover:text-foreground'
              }`}
            >
              User
            </button>
          </div>

          {/* Status Filter */}
          <div className="flex items-center gap-1 rounded-md bg-muted p-1">
            <button
              onClick={() => setStatusFilter('all')}
              className={`rounded px-3 py-1 text-sm transition-colors ${
                statusFilter === 'all'
                  ? 'bg-card text-foreground shadow-sm'
                  : 'text-muted-foreground hover:text-foreground'
              }`}
            >
              All
            </button>
            <button
              onClick={() => setStatusFilter('active')}
              className={`rounded px-3 py-1 text-sm transition-colors ${
                statusFilter === 'active'
                  ? 'bg-card text-foreground shadow-sm'
                  : 'text-muted-foreground hover:text-foreground'
              }`}
            >
              Active
            </button>
          </div>

          {/* Search */}
          <Input
            placeholder="Search users..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="w-64"
          />
        </div>

        <Button onClick={() => setIsAddUserOpen(true)}>
          <Plus className="mr-2 h-4 w-4" />
          Add User
        </Button>
      </div>

      {/* Users Table */}
      <div className="rounded-lg border bg-card">
        <table className="w-full">
          <thead>
            <tr className="border-b">
              <th className="p-4 text-left text-sm font-medium text-foreground">User</th>
              <th className="p-4 text-left text-sm font-medium text-foreground">Role</th>
              <th className="p-4 text-left text-sm font-medium text-foreground">Status</th>
              <th className="p-4 text-left text-sm font-medium text-foreground">Last Login</th>
              <th className="p-4 text-left text-sm font-medium text-foreground">Created</th>
              <th className="p-4 text-left text-sm font-medium text-foreground">Actions</th>
            </tr>
          </thead>
          <tbody>
            {filteredUsers.map((user) => (
              <tr key={user.id} className="border-b hover:bg-muted">
                <td className="p-4">
                  <div>
                    <div className="text-sm font-medium">{user.name}</div>
                    <div className="text-sm text-muted-foreground">{user.email}</div>
                  </div>
                </td>
                <td className="p-4">
                  <div className="flex flex-wrap gap-1">
                    {user.role.split(', ').map((role, index) => (
                      <span
                        key={index}
                        className={`inline-flex rounded-full px-2 py-1 text-xs font-medium ${
                          role === 'super-admin'
                            ? 'bg-purple-100 text-purple-700'
                            : role === 'admin'
                              ? 'bg-info/10 text-info'
                              : 'bg-muted text-foreground'
                        }`}
                      >
                        {role}
                      </span>
                    ))}
                  </div>
                </td>
                <td className="p-4">
                  <span
                    className={`inline-flex rounded-full px-2 py-1 text-xs font-medium ${
                      user.status === 'active'
                        ? 'bg-success/10 text-success'
                        : user.status === 'suspended'
                          ? 'bg-destructive/10 text-destructive'
                          : 'bg-warning/10 text-warning'
                    }`}
                  >
                    {user.status}
                  </span>
                </td>
                <td className="p-4">
                  <span className="text-sm text-muted-foreground">{user.lastLogin}</span>
                </td>
                <td className="p-4">
                  <span className="text-sm text-muted-foreground">{user.created}</span>
                </td>
                <td className="p-4">
                  <DropdownMenu>
                    <DropdownMenuTrigger asChild>
                      <Button variant="ghost" size="sm">
                        <MoreHorizontal className="h-4 w-4" />
                      </Button>
                    </DropdownMenuTrigger>
                    <DropdownMenuContent align="end">
                      {canEditUser(user) && (
                        <>
                          <DropdownMenuItem onClick={() => handleEditUser(user)}>
                            <Edit className="mr-2 h-4 w-4" />
                            Edit User
                          </DropdownMenuItem>
                          <DropdownMenuItem onClick={() => setResettingPasswordUser(user)}>
                            <Key className="mr-2 h-4 w-4" />
                            Reset Password
                          </DropdownMenuItem>
                          <DropdownMenuItem>
                            <Mail className="mr-2 h-4 w-4" />
                            Send Invite
                          </DropdownMenuItem>
                          <DropdownMenuSeparator />
                          {user.status === 'active' ? (
                            <DropdownMenuItem
                              className="text-warning"
                              onClick={() => handleToggleUserStatus(user, false)}
                            >
                              <UserX className="mr-2 h-4 w-4" />
                              Disable User
                            </DropdownMenuItem>
                          ) : (
                            <DropdownMenuItem
                              className="text-success"
                              onClick={() => handleToggleUserStatus(user, true)}
                            >
                              <UserCheck className="mr-2 h-4 w-4" />
                              Enable User
                            </DropdownMenuItem>
                          )}
                          <DropdownMenuSeparator />
                          <DropdownMenuItem
                            className="text-destructive"
                            onClick={() => setDeletingUser(user)}
                          >
                            <Trash2 className="mr-2 h-4 w-4" />
                            Delete User
                          </DropdownMenuItem>
                        </>
                      )}
                      {!canEditUser(user) && (
                        <DropdownMenuItem disabled>
                          <Edit className="mr-2 h-4 w-4" />
                          View Only (Insufficient Permissions)
                        </DropdownMenuItem>
                      )}
                    </DropdownMenuContent>
                  </DropdownMenu>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {/* Add/Edit User Dialog */}
      <Dialog
        open={isAddUserOpen || !!editingUser}
        onOpenChange={(open) => {
          if (!open) {
            resetForm()
          }
        }}
      >
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{editingUser ? 'Edit User' : 'Add New User'}</DialogTitle>
            <DialogDescription>
              {editingUser
                ? 'Update user information and permissions'
                : 'Create a new user account'}
              {!editingUser && currentUserRole === 'admin' && (
                <span className="text-muted-foreground mt-1 block text-sm">
                  As an admin, you can only create regular users. Contact a super admin to create
                  admin users.
                </span>
              )}
            </DialogDescription>
          </DialogHeader>
          {formError && (
            <div className="bg-destructive/10 border-destructive/20 mx-6 rounded-md border p-3">
              <p className="text-destructive text-sm">{formError}</p>
            </div>
          )}
          <div className="space-y-6 py-4">
            <div className="space-y-2">
              <Label htmlFor="email">Email *</Label>
              <Input
                id="email"
                type="email"
                value={formData.email}
                onChange={(e) => setFormData((prev) => ({ ...prev, email: e.target.value }))}
                placeholder="Enter email address"
                required
                disabled={!!editingUser}
                className={editingUser ? 'bg-muted cursor-not-allowed' : ''}
              />
              {editingUser && (
                <p className="text-muted-foreground mt-1 text-sm">
                  Email cannot be changed after user creation
                </p>
              )}
            </div>
            <div className="space-y-2">
              <Label htmlFor="displayName">Display Name</Label>
              <Input
                id="displayName"
                value={formData.displayName || ''}
                onChange={(e) => setFormData((prev) => ({ ...prev, displayName: e.target.value }))}
                placeholder="Enter display name (optional)"
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="roles">Roles *</Label>
              <div className="mt-2 flex flex-wrap gap-2">
                {getAvailableRoles(currentUserRole).map((role) => (
                  <button
                    key={role}
                    type="button"
                    onClick={() => {
                      const isSelected = formData.roles.includes(role)
                      if (isSelected) {
                        setFormData((prev) => ({
                          ...prev,
                          roles: prev.roles.filter((r) => r !== role),
                        }))
                      } else {
                        setFormData((prev) => ({
                          ...prev,
                          roles: [...prev.roles, role],
                        }))
                      }
                    }}
                    className={`rounded-md border px-3 py-2 text-sm transition-colors ${
                      formData.roles.includes(role)
                        ? 'bg-primary text-primary-foreground border-primary'
                        : 'bg-card hover:bg-muted border-border'
                    }`}
                  >
                    {role === 'super-admin'
                      ? 'Super Admin'
                      : role.charAt(0).toUpperCase() + role.slice(1)}
                  </button>
                ))}
              </div>
              {hasAttemptedSubmit && formData.roles.length === 0 && (
                <p className="text-destructive mt-2 text-sm">At least one role is required</p>
              )}
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={resetForm} disabled={isPending}>
              Cancel
            </Button>
            <Button
              onClick={handleSubmit}
              disabled={isPending || !formData.email || formData.roles.length === 0}
            >
              {isPending ? (
                <>
                  <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                  {editingUser ? 'Updating...' : 'Creating...'}
                </>
              ) : editingUser ? (
                'Update'
              ) : (
                'Create'
              )}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Delete User Confirmation Dialog */}
      <Dialog
        open={!!deletingUser}
        onOpenChange={(open) => {
          if (!open) {
            resetDeleteDialog()
          }
        }}
      >
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Delete User</DialogTitle>
            <DialogDescription>
              Are you sure you want to delete user &quot;{deletingUser?.name}&quot;? This action
              cannot be undone.
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button variant="outline" onClick={resetDeleteDialog} disabled={isPending}>
              Cancel
            </Button>
            <Button variant="destructive" onClick={handleDeleteUser} disabled={isPending}>
              {isPending ? (
                <>
                  <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                  Deleting...
                </>
              ) : (
                'Delete User'
              )}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Reset Password Dialog */}
      <Dialog
        open={!!resettingPasswordUser}
        onOpenChange={(open) => {
          if (!open) {
            resetPasswordDialog()
          }
        }}
      >
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Reset Password</DialogTitle>
            <DialogDescription>
              Set a new password for user &quot;{resettingPasswordUser?.email}&quot;. The password
              must be at least 8 characters long.
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-4 py-4">
            <div className="space-y-2">
              <Label htmlFor="newPassword">New Password</Label>
              <Input
                id="newPassword"
                type="password"
                value={newPassword}
                onChange={(e) => setNewPassword(e.target.value)}
                placeholder="Enter new password (min 8 characters)"
                minLength={8}
                required
                className="w-full"
              />
              {newPassword && newPassword.length < 8 && (
                <p className="text-sm text-destructive">Password must be at least 8 characters long</p>
              )}
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={resetPasswordDialog} disabled={isPending}>
              Cancel
            </Button>
            <Button onClick={handleResetPassword} disabled={isPending || newPassword.length < 8}>
              {isPending ? (
                <>
                  <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                  Resetting...
                </>
              ) : (
                'Reset Password'
              )}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}
