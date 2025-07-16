export type TeamRole = 'owner' | 'admin' | 'member'
export type TeamVisibility = 'public' | 'private'
export type InvitationStatus = 'pending' | 'accepted' | 'declined' | 'expired'

export interface Team {
  id: string
  name: string
  description?: string
  visibility: TeamVisibility
  region: string
  memberCount: number
  createdAt: Date
  updatedAt: Date
  lastActivity?: Date
  joinedAt?: Date // When the current user joined this team
}

export interface TeamMember {
  id: string
  teamId: string
  userId: string
  role: TeamRole
  joinedAt: Date
  user: {
    id: string
    name: string
    email: string
    avatar?: string
  }
}

export interface TeamInvitation {
  id: string
  teamId: string
  team: {
    id: string
    name: string
    description?: string
  }
  inviterId: string
  inviter: {
    id: string
    name: string
    email: string
  }
  inviteeEmail: string
  role: TeamRole
  status: InvitationStatus
  createdAt: Date
  expiresAt: Date
}

export interface CreateTeamInput {
  name: string
  description?: string
  region: string
  invitations?: {
    email: string
    role: TeamRole
  }[]
}

export interface UpdateTeamInput {
  name?: string
  description?: string
  visibility?: TeamVisibility
}

export interface InviteTeamMemberInput {
  teamId: string
  email: string
  role: TeamRole
}