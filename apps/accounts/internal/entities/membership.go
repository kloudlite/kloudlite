package entities

import (
	iamT "kloudlite.io/apps/iam/types"
	"kloudlite.io/pkg/repos"
)

type InvitationStatus struct {
	ThroughInvitation bool     `json:"throughInvitation"`
	InvitationId      repos.ID `json:"invitationId"`
}

type Membership struct {
	AccountName      string            `json:"accountName"`
	UserId           repos.ID          `json:"userId"`
	Role             iamT.Role         `json:"role"`
	InvitationStatus *InvitationStatus `json:"invitationStatus,omitempty"`
}
