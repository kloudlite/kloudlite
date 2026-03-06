'use client'

import { useState, useTransition, useEffect, useCallback } from 'react'
import { useRouter } from 'next/navigation'
import { useResourceWatch } from '@/lib/hooks/use-resource-watch'
import {
  Button,
  Input,
  Label,
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@kloudlite/ui'
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
  Check,
  X,
  AlertCircle,
  Cpu,
} from 'lucide-react'
import {
  createUser,
  updateUser,
  deleteUser,
  resetUserPassword,
  checkUsernameAvailability,
} from '@/app/actions/user.actions'
import { adminAssignMachineType } from '@/app/actions/work-machine.actions'
import { generateUsernameFromEmail } from '@/lib/utils/username'
import { UserDisplay, CreateUserFormData, UserResource, userToDisplay } from '@/types/user'
import { toast } from 'sonner'
import type { WorkMachine } from '@kloudlite/lib/k8s'

// Helper function to get available roles based on current user's role
function getAvailableRoles(currentUserRole: 'super-admin' | 'admin'): string[] {
  if (currentUserRole === 'super-admin') {
    return ['user', 'admin']
  } else {
    // Admin can only create regular users
    return ['user']
  }
}

interface MachineTypeOption {
  id: string
  name: string
  description?: string
  cpu?: string
  memory?: string
  tierSubtitle?: string
  tierPrice?: string
  tierPriceUnit?: string
  tierIncludedHours?: string
  tierExtraHourPrice?: string
  tierStorageGb?: string
  tierSuspendMinutes?: string
  tierPopular?: boolean
}

interface UserManagementListProps {
  users: UserDisplay[]
  currentUserRole: 'super-admin' | 'admin'
  isKloudliteCloud?: boolean
  machineTypes?: MachineTypeOption[]
  workMachines?: WorkMachine[]
}

export function UserManagementList({
  users: initialUsers,
  currentUserRole,
  isKloudliteCloud,
  machineTypes = [],
  workMachines = [],
}: UserManagementListProps) {
  const router = useRouter()
  const [users, setUsers] = useState(initialUsers)

  // Sync users state with server prop changes (e.g. after router.refresh from watch events)
  useEffect(() => { setUsers(initialUsers) }, [initialUsers])

  // Auto-refresh when K8s resources change via watch events
  useResourceWatch('users')
  useResourceWatch('workmachines')
  const [roleFilter, setRoleFilter] = useState<'all' | 'super-admin' | 'admin' | 'user'>('all')
  const [statusFilter, setStatusFilter] = useState<'all' | 'active'>('all')
  const [searchQuery, setSearchQuery] = useState('')
  const [isAddUserOpen, setIsAddUserOpen] = useState(false)
  const [editingUser, setEditingUser] = useState<UserDisplay | null>(null)
  const [deletingUser, setDeletingUser] = useState<UserDisplay | null>(null)
  const [resettingPasswordUser, setResettingPasswordUser] = useState<UserDisplay | null>(null)
  const [newPassword, setNewPassword] = useState('')
  const [isSubmitting, startSubmitTransition] = useTransition()
  const [isDeleting, startDeleteTransition] = useTransition()
  const [isResettingPassword, startResetPasswordTransition] = useTransition()
  const [_isTogglingStatus, startToggleStatusTransition] = useTransition()
  const [isAssigningMachine, startAssignMachineTransition] = useTransition()
  const [formError, setFormError] = useState<string | null>(null)
  const [hasAttemptedSubmit, setHasAttemptedSubmit] = useState(false)

  // Machine type assignment state (Kloudlite Cloud only)
  const [assigningUser, setAssigningUser] = useState<UserDisplay | null>(null)
  const [selectedMachineType, setSelectedMachineType] = useState('')
  const [createUserMachineType, setCreateUserMachineType] = useState('')

  // Helper to get machine type for a user from work machines
  const getUserMachineType = useCallback((username: string): string | null => {
    const wm = workMachines.find((m) => m.spec?.ownedBy === username)
    return wm?.spec?.machineType || null
  }, [workMachines])

  // Helper to get machine type display name
  const getMachineTypeName = useCallback((typeId: string): string => {
    const mt = machineTypes.find((t) => t.id === typeId)
    return mt?.name || typeId
  }, [machineTypes])

  // Form state
  const [formData, setFormData] = useState<CreateUserFormData>({
    username: '',
    email: '',
    displayName: '',
    password: '',
    roles: [],
  })

  // Username validation state
  const [usernameStatus, setUsernameStatus] = useState<{
    checking: boolean
    available: boolean | null
    suggested: string | null
  }>({
    checking: false,
    available: null,
    suggested: null,
  })
  const [usernameManuallyEdited, setUsernameManuallyEdited] = useState(false)

  const resetForm = () => {
    setFormData({
      username: '',
      email: '',
      displayName: '',
      roles: [],
    })
    setFormError(null)
    setHasAttemptedSubmit(false)
    setEditingUser(null)
    setIsAddUserOpen(false)
    setUsernameStatus({
      checking: false,
      available: null,
      suggested: null,
    })
    setUsernameManuallyEdited(false)
    setCreateUserMachineType('')
  }

  const resetDeleteDialog = () => {
    setDeletingUser(null)
  }

  const resetPasswordDialog = () => {
    setResettingPasswordUser(null)
    setNewPassword('')
  }

  // Auto-suggest username from email when email changes (only when creating new user)
  useEffect(() => {
    if (!editingUser && formData.email && !usernameManuallyEdited) {
      const suggested = generateUsernameFromEmail(formData.email)
      setFormData((prev) => ({ ...prev, username: suggested }))
    }
  }, [formData.email, editingUser, usernameManuallyEdited])

  // Debounced username availability check
  useEffect(() => {
    if (editingUser || !formData.username || formData.username.length < 3) {
      setUsernameStatus({ checking: false, available: null, suggested: null })
      return
    }

    const timeoutId = setTimeout(async () => {
      setUsernameStatus((prev) => ({ ...prev, checking: true }))

      const result = await checkUsernameAvailability(formData.username)

      if (result.success) {
        setUsernameStatus({
          checking: false,
          available: result.data?.available ?? null,
          suggested: null,
        })
      } else {
        setUsernameStatus({
          checking: false,
          available: null,
          suggested: null,
        })
      }
    }, 500) // 500ms debounce

    return () => clearTimeout(timeoutId)
  }, [formData.username, editingUser])

  // Handler for username change
  const handleUsernameChange = useCallback((value: string) => {
    setFormData((prev) => ({ ...prev, username: value }))
    setUsernameManuallyEdited(true)
  }, [])

  // Handler to use suggested username
  const useSuggestedUsername = useCallback(() => {
    if (usernameStatus.suggested) {
      setFormData((prev) => ({ ...prev, username: usernameStatus.suggested! }))
      setUsernameManuallyEdited(true)
    }
  }, [usernameStatus.suggested])

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
      username: user.username,
      email: user.email,
      displayName: user.displayName || '',
      roles: roles.length > 0 ? roles : ['user'], // Ensure at least one role
    })

    // Pre-populate machine type for editing
    if (isKloudliteCloud) {
      setCreateUserMachineType(getUserMachineType(user.username) || '')
    }
  }

  // Helper function to check if current user can edit another user
  const canEditUser = (targetUser: UserDisplay): boolean => {
    if (currentUserRole === 'super-admin') {
      return true // Super admin can edit anyone
    }

    if (currentUserRole === 'admin') {
      // Admin can only edit regular users, not other admins or super-admins
      const targetRoles = (targetUser.role || '').split(', ')
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
    if (!formData.username && !editingUser) {
      setFormError('Username is required')
      return
    }

    if (!formData.email) {
      setFormError('Email is required')
      return
    }

    if (!formData.roles || formData.roles.length === 0) {
      setFormError('At least one role is required')
      return
    }

    if (isKloudliteCloud && !editingUser && formData.roles.includes('user') && !createUserMachineType) {
      setFormError('Machine type is required when user role is assigned')
      return
    }

    startSubmitTransition(async () => {
      try {
        if (editingUser) {
          // Update existing user
          const result = await updateUser(editingUser.id, formData)
          if (result.success && result.data) {
            setUsers((prev) =>
              prev.map((u) =>
                u.id === editingUser.id ? userToDisplay(result.data as UserResource) : u,
              ),
            )
            // If machine type changed, assign the new one
            const currentMt = getUserMachineType(editingUser.username)
            if (isKloudliteCloud && createUserMachineType && createUserMachineType !== currentMt) {
              const assignResult = await adminAssignMachineType(editingUser.username, createUserMachineType)
              if (assignResult.success) {
                toast.success('User updated and machine type changed')
              } else {
                toast.success('User updated, but failed to change machine type')
              }
            } else {
              toast.success('User updated successfully')
            }
            resetForm()
            router.refresh()
          } else {
            setFormError(result.error || 'Failed to update user')
            toast.error(result.error || 'Failed to update user')
          }
        } else {
          // Create new user
          const result = await createUser(formData)
          if (result.success && result.data) {
            setUsers((prev) => [...prev, userToDisplay(result.data as UserResource)])
            // If cloud mode and machine type selected, assign it
            if (isKloudliteCloud && createUserMachineType && formData.username) {
              const assignResult = await adminAssignMachineType(formData.username, createUserMachineType)
              if (assignResult.success) {
                toast.success('User created and machine type assigned')
              } else {
                toast.error('User created, but failed to assign machine type')
              }
              setCreateUserMachineType('')
            } else {
              toast.success('User created successfully')
            }
            resetForm()
            router.refresh()
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

    startDeleteTransition(async () => {
      try {
        const result = await deleteUser(deletingUser.id)
        if (result.success) {
          setUsers((prev) => prev.filter((u) => u.id !== deletingUser.id))
          toast.success('User deleted successfully')
          resetDeleteDialog()
          router.refresh()
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

    startResetPasswordTransition(async () => {
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
    startToggleStatusTransition(async () => {
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

  const handleAssignMachineType = async () => {
    if (!assigningUser || !selectedMachineType) return

    startAssignMachineTransition(async () => {
      try {
        const result = await adminAssignMachineType(assigningUser.username, selectedMachineType)
        if (result.success) {
          toast.success(`Machine type assigned to ${assigningUser.name}`)
          setAssigningUser(null)
          setSelectedMachineType('')
          router.refresh()
        } else {
          toast.error(result.error || 'Failed to assign machine type')
        }
      } catch (error) {
        toast.error('An unexpected error occurred')
        console.error('Assign machine type error:', error)
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
          <div className="bg-muted flex items-center gap-1 rounded-md p-1">
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
          <div className="bg-muted flex items-center gap-1 rounded-md p-1">
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
      <div className="bg-card rounded-lg border">
        <table className="w-full">
          <thead>
            <tr className="border-b">
              <th className="text-foreground p-4 text-left text-sm font-medium">User</th>
              <th className="text-foreground w-40 p-4 text-left text-sm font-medium">Role</th>
              {isKloudliteCloud && (
                <th className="text-foreground w-40 p-4 text-left text-sm font-medium">Machine Type</th>
              )}
              <th className="text-foreground w-28 p-4 text-left text-sm font-medium">Status</th>
              <th className="text-foreground w-32 p-4 text-left text-sm font-medium">Last Login</th>
              <th className="text-foreground w-32 p-4 text-left text-sm font-medium">Created</th>
              <th className="text-foreground w-20 p-4 text-left text-sm font-medium">Actions</th>
            </tr>
          </thead>
          <tbody>
            {filteredUsers.length === 0 ? (
              <tr>
                <td colSpan={isKloudliteCloud ? 7 : 6} className="p-12 text-center">
                  <div className="text-muted-foreground">
                    <p className="text-sm font-medium">No users found</p>
                    <p className="mt-1 text-xs">Click &quot;Add User&quot; to create the first user.</p>
                  </div>
                </td>
              </tr>
            ) : (
              filteredUsers.map((user) => (
                <tr key={user.id} className="hover:bg-muted border-b">
                  <td className="p-4">
                    <div>
                      <div className="text-sm font-medium">{user.name}</div>
                      <div className="text-muted-foreground text-sm">{user.email}</div>
                    </div>
                  </td>
                  <td className="w-40 p-4">
                    <div className="flex flex-wrap gap-1">
                      {(user.role || '').split(', ').filter(Boolean).map((role) => (
                        <span
                          key={role}
                          className={`inline-flex rounded-md px-2 py-1 text-xs font-medium ${
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
                  {isKloudliteCloud && (
                    <td className="w-40 p-4">
                      {(() => {
                        const mt = getUserMachineType(user.username)
                        return mt ? (
                          <span className="inline-flex rounded-md bg-primary/10 text-primary px-2 py-1 text-xs font-medium">
                            {getMachineTypeName(mt)}
                          </span>
                        ) : (
                          <span className="text-muted-foreground text-xs italic">Not assigned</span>
                        )
                      })()}
                    </td>
                  )}
                  <td className="w-28 p-4">
                    <span
                      className={`inline-flex min-w-[70px] justify-center rounded-md px-2 py-1 text-xs font-medium ${
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
                  <td className="w-32 p-4">
                    <span className="text-muted-foreground text-sm">{user.lastLogin}</span>
                  </td>
                  <td className="w-32 p-4">
                    <span className="text-muted-foreground text-sm">{user.created}</span>
                  </td>
                  <td className="w-20 p-4">
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
                            {isKloudliteCloud && (
                              <DropdownMenuItem onClick={() => {
                                setAssigningUser(user)
                                setSelectedMachineType(getUserMachineType(user.username) || '')
                              }}>
                                <Cpu className="mr-2 h-4 w-4" />
                                Assign Machine Type
                              </DropdownMenuItem>
                            )}
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
              ))
            )}
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
              <Label htmlFor="displayName">Display Name</Label>
              <Input
                id="displayName"
                value={formData.displayName || ''}
                onChange={(e) => setFormData((prev) => ({ ...prev, displayName: e.target.value }))}
                placeholder="Enter display name (optional)"
              />
            </div>
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
              <Label htmlFor="username">Username *</Label>
              <div className="relative">
                <Input
                  id="username"
                  type="text"
                  value={formData.username}
                  onChange={(e) => handleUsernameChange(e.target.value)}
                  placeholder="Enter username (e.g., john-doe)"
                  required
                  disabled={!!editingUser}
                  className={editingUser ? 'bg-muted cursor-not-allowed pr-10' : 'pr-10'}
                  pattern="^[a-z0-9]([a-z0-9\-]*[a-z0-9])?$"
                  minLength={3}
                  maxLength={63}
                />
                {!editingUser && formData.username.length >= 3 && (
                  <div className="absolute right-3 top-1/2 -translate-y-1/2">
                    {usernameStatus.checking && (
                      <Loader2 className="h-4 w-4 animate-spin text-muted-foreground" />
                    )}
                    {!usernameStatus.checking && usernameStatus.available === true && (
                      <Check className="h-4 w-4 text-green-600" />
                    )}
                    {!usernameStatus.checking && usernameStatus.available === false && (
                      <X className="h-4 w-4 text-red-600" />
                    )}
                  </div>
                )}
              </div>
              {editingUser && (
                <p className="text-muted-foreground mt-1 text-sm">
                  Username cannot be changed after user creation
                </p>
              )}
              {!editingUser && !usernameStatus.suggested && (
                <p className="text-muted-foreground mt-1 text-sm">
                  Username is the resource name (3-63 chars, lowercase, numbers, hyphens only)
                </p>
              )}
              {!editingUser && usernameStatus.available === false && usernameStatus.suggested && (
                <div className="flex items-center gap-2 mt-1">
                  <AlertCircle className="h-4 w-4 text-amber-600" />
                  <p className="text-sm text-muted-foreground">
                    Username is taken.{' '}
                    <button
                      type="button"
                      onClick={useSuggestedUsername}
                      className="text-primary hover:underline font-medium"
                    >
                      Use suggested: {usernameStatus.suggested}
                    </button>
                  </p>
                </div>
              )}
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
                        // Clear machine type selection when "user" role is removed
                        if (role === 'user') {
                          setCreateUserMachineType('')
                        }
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
            {isKloudliteCloud && formData.roles.includes('user') && (
              <div className="space-y-2">
                <Label>Machine Type {!editingUser && '*'}</Label>
                <Select value={createUserMachineType} onValueChange={setCreateUserMachineType}>
                  <SelectTrigger className="w-full">
                    <SelectValue placeholder="Select a machine type" />
                  </SelectTrigger>
                  <SelectContent>
                    {machineTypes.map((mt) => (
                      <SelectItem key={mt.id} value={mt.id}>
                        <span className="font-medium">{mt.name}</span>
                        {mt.cpu && (
                          <span className="text-muted-foreground text-xs ml-2">
                            ({mt.cpu} vCPUs, {mt.memory} RAM{mt.tierPrice ? `, $${mt.tierPrice}${mt.tierPriceUnit}` : ''})
                          </span>
                        )}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
                {!editingUser && hasAttemptedSubmit && !createUserMachineType && (
                  <p className="text-destructive mt-1 text-sm">Machine type is required</p>
                )}
                <p className="text-muted-foreground text-sm">
                  {editingUser
                    ? 'Change the machine type for this user\'s work machine.'
                    : 'A work machine will be created in stopped state with this configuration.'}
                </p>
              </div>
            )}
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={resetForm} disabled={isSubmitting}>
              Cancel
            </Button>
            <Button
              onClick={handleSubmit}
              disabled={
                isSubmitting ||
                (!editingUser && !formData.username) ||
                !formData.email ||
                formData.roles.length === 0 ||
                (isKloudliteCloud && !editingUser && formData.roles.includes('user') && !createUserMachineType)
              }
            >
              {isSubmitting ? (
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
            <Button variant="outline" onClick={resetDeleteDialog} disabled={isDeleting}>
              Cancel
            </Button>
            <Button variant="destructive" onClick={handleDeleteUser} disabled={isDeleting}>
              {isDeleting ? (
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
                <p className="text-destructive text-sm">
                  Password must be at least 8 characters long
                </p>
              )}
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={resetPasswordDialog} disabled={isResettingPassword}>
              Cancel
            </Button>
            <Button onClick={handleResetPassword} disabled={isResettingPassword || newPassword.length < 8}>
              {isResettingPassword ? (
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

      {/* Assign Machine Type Dialog (Kloudlite Cloud only) */}
      {isKloudliteCloud && (
        <Dialog
          open={!!assigningUser}
          onOpenChange={(open) => {
            if (!open) {
              setAssigningUser(null)
              setSelectedMachineType('')
            }
          }}
        >
          <DialogContent className="sm:max-w-2xl">
            <DialogHeader>
              <DialogTitle>Assign Machine Type</DialogTitle>
              <DialogDescription>
                Select a machine type for user &quot;{assigningUser?.name}&quot;.
                {getUserMachineType(assigningUser?.username || '')
                  ? ' This will update their existing machine configuration.'
                  : ' A new work machine will be created in stopped state.'}
              </DialogDescription>
            </DialogHeader>
            <div className="space-y-3 py-4">
              {machineTypes.map((mt) => {
                const isSelected = selectedMachineType === mt.id
                const hasTierData = !!mt.tierPrice
                return (
                  <button
                    key={mt.id}
                    type="button"
                    onClick={() => setSelectedMachineType(mt.id)}
                    className={`w-full text-left rounded-lg border p-4 transition-colors ${
                      isSelected
                        ? 'border-primary bg-primary/5 ring-2 ring-primary/20'
                        : 'border-border hover:border-primary/40 hover:bg-muted/30'
                    }`}
                  >
                    <div className="flex items-start justify-between gap-4">
                      <div className="flex-1 min-w-0">
                        <div className="flex items-center gap-2">
                          <span className="font-semibold text-sm">{mt.name}</span>
                          {mt.tierPopular && (
                            <span className="rounded-md bg-primary/10 text-primary px-2 py-0.5 text-[10px] font-semibold uppercase tracking-wider">
                              Most Popular
                            </span>
                          )}
                        </div>
                        {mt.tierSubtitle && (
                          <p className="text-muted-foreground text-xs mt-0.5">{mt.tierSubtitle}</p>
                        )}
                        <div className="flex flex-wrap items-center gap-x-4 gap-y-1 mt-2 text-xs text-muted-foreground">
                          {mt.cpu && <span>{mt.cpu} vCPUs</span>}
                          {mt.memory && <span>{mt.memory} RAM</span>}
                          {mt.tierStorageGb && <span>{mt.tierStorageGb}GB storage</span>}
                          {mt.tierIncludedHours && <span>{mt.tierIncludedHours} hrs/mo</span>}
                          {mt.tierSuspendMinutes && (
                            <span>
                              {parseInt(mt.tierSuspendMinutes) >= 60
                                ? `${parseInt(mt.tierSuspendMinutes) / 60} hr suspend`
                                : `${mt.tierSuspendMinutes} min suspend`}
                            </span>
                          )}
                        </div>
                      </div>
                      {hasTierData && (
                        <div className="text-right shrink-0">
                          <div className="text-lg font-bold">${mt.tierPrice}</div>
                          <div className="text-muted-foreground text-[10px]">{mt.tierPriceUnit}</div>
                          {mt.tierExtraHourPrice && (
                            <div className="text-muted-foreground text-[10px] mt-0.5">
                              +${mt.tierExtraHourPrice}/hr extra
                            </div>
                          )}
                        </div>
                      )}
                    </div>
                  </button>
                )
              })}
            </div>
            <DialogFooter>
              <Button
                variant="outline"
                onClick={() => {
                  setAssigningUser(null)
                  setSelectedMachineType('')
                }}
                disabled={isAssigningMachine}
              >
                Cancel
              </Button>
              <Button
                onClick={handleAssignMachineType}
                disabled={isAssigningMachine || !selectedMachineType}
              >
                {isAssigningMachine ? (
                  <>
                    <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                    Assigning...
                  </>
                ) : (
                  'Assign'
                )}
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      )}
    </div>
  )
}
