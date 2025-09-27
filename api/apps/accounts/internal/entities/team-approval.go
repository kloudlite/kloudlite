package entities

import (
	"time"
	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/pkg/repos"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ApprovalStatus string

const (
	ApprovalStatusPending  ApprovalStatus = "pending"
	ApprovalStatusApproved ApprovalStatus = "approved"
	ApprovalStatusRejected ApprovalStatus = "rejected"
)

type TeamApprovalRequest struct {
	repos.BaseEntity  `json:",inline" graphql:"noinput"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	
	common.ResourceMetadata `json:",inline"`

	// Team details from the request
	TeamSlug        string         `json:"teamSlug"`
	TeamDescription string         `json:"teamDescription,omitempty"`
	TeamRegion      string         `json:"teamRegion"`
	
	// Request metadata
	RequestedBy     repos.ID       `json:"requestedBy"`
	RequestedByEmail string        `json:"requestedByEmail"`
	RequestedAt     time.Time      `json:"requestedAt"`
	
	// Approval metadata
	Status          ApprovalStatus `json:"status"`
	ReviewedBy      *repos.ID      `json:"reviewedBy,omitempty"`
	ReviewedByEmail *string        `json:"reviewedByEmail,omitempty"`
	ReviewedAt      *time.Time     `json:"reviewedAt,omitempty"`
	RejectionReason *string        `json:"rejectionReason,omitempty"`
	
	// Created team ID if approved
	CreatedTeamId   *repos.ID      `json:"createdTeamId,omitempty"`
}

var TeamApprovalRequestIndices = []repos.IndexField{
	{
		Field: []repos.IndexKey{
			{Key: "id", Value: repos.IndexAsc},
		},
		Unique: true,
	},
	{
		Field: []repos.IndexKey{
			{Key: "teamSlug", Value: repos.IndexAsc},
			{Key: "status", Value: repos.IndexAsc},
		},
	},
	{
		Field: []repos.IndexKey{
			{Key: "requestedBy", Value: repos.IndexAsc},
		},
	},
	{
		Field: []repos.IndexKey{
			{Key: "status", Value: repos.IndexAsc},
		},
	},
	{
		Field: []repos.IndexKey{
			{Key: "reviewedBy", Value: repos.IndexAsc},
		},
	},
}