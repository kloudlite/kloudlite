package entities

import (
	"time"
	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/pkg/repos"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NotificationType defines the type of notification
type NotificationType string

const (
	// Team-related notifications
	NotificationTypeTeamRequest     NotificationType = "team_request"
	NotificationTypeTeamApproved    NotificationType = "team_approved"
	NotificationTypeTeamRejected    NotificationType = "team_rejected"
	NotificationTypeTeamInvite      NotificationType = "team_invite"
	NotificationTypeMemberAdded     NotificationType = "member_added"
	NotificationTypeMemberRemoved   NotificationType = "member_removed"
	
	// Platform-related notifications
	NotificationTypePlatformInvite  NotificationType = "platform_invite"
	
	// General notifications
	NotificationTypeAnnouncement    NotificationType = "announcement"
	NotificationTypeSystem          NotificationType = "system"
)

// TargetType defines how notifications are targeted
type TargetType string

const (
	TargetTypeUser         TargetType = "user"
	TargetTypeTeamRole     TargetType = "team_role"
	TargetTypePlatformRole TargetType = "platform_role"
)

// Role hierarchy definitions
var (
	TeamRoleHierarchy = map[string]int{
		"member": 1,
		"admin":  2,
		"owner":  3,
	}
	
	PlatformRoleHierarchy = map[string]int{
		"user":        1,
		"admin":       2,
		"super_admin": 3,
	}
)

// NotificationTarget defines who should receive the notification
type NotificationTarget struct {
	Type            TargetType `json:"type" bson:"type"`
	UserId          *repos.ID  `json:"userId,omitempty" bson:"userId,omitempty"`
	TeamId          *repos.ID  `json:"teamId,omitempty" bson:"teamId,omitempty"`
	MinTeamRole     *string    `json:"minTeamRole,omitempty" bson:"minTeamRole,omitempty"`
	MinPlatformRole *string    `json:"minPlatformRole,omitempty" bson:"minPlatformRole,omitempty"`
}

// ReadStatus tracks who has read the notification
type ReadStatus struct {
	UserId repos.ID  `json:"userId" bson:"userId"`
	ReadAt time.Time `json:"readAt" bson:"readAt"`
}

// ActionStatus tracks who has taken action on the notification
type ActionStatus struct {
	UserId   repos.ID  `json:"userId" bson:"userId"`
	ActionAt time.Time `json:"actionAt" bson:"actionAt"`
	Action   string    `json:"action" bson:"action"` // Which action was taken
}

// NotificationAction defines an available action on a notification
type NotificationAction struct {
	Id       string `json:"id" bson:"id"`             // e.g., "approve", "reject", "accept", "decline"
	Label    string `json:"label" bson:"label"`       // e.g., "Approve", "Reject"
	Style    string `json:"style" bson:"style"`       // e.g., "primary", "danger", "default"
	Endpoint string `json:"endpoint" bson:"endpoint"` // e.g., "/api/teams/approve"
	Method   string `json:"method" bson:"method"`     // e.g., "POST", "DELETE"
	Data     map[string]string `json:"data,omitempty" bson:"data,omitempty"` // Additional data to send
}

// Notification represents a system notification
type Notification struct {
	repos.BaseEntity  `json:",inline" bson:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty" bson:"metadata,omitempty"`
	
	common.ResourceMetadata `json:",inline" bson:",inline"`

	// Targeting
	Target NotificationTarget `json:"target" bson:"target"`
	
	// Content
	Type        NotificationType `json:"type" bson:"type"`
	Title       string           `json:"title" bson:"title"`
	Description string           `json:"description" bson:"description"`
	
	// Metadata
	TeamId      *repos.ID `json:"teamId,omitempty" bson:"teamId,omitempty"`
	TeamName    *string   `json:"teamName,omitempty" bson:"teamName,omitempty"`
	RequestId   *repos.ID `json:"requestId,omitempty" bson:"requestId,omitempty"`
	InviteId    *repos.ID `json:"inviteId,omitempty" bson:"inviteId,omitempty"`
	
	// Deduplication
	DedupeKey string `json:"dedupeKey,omitempty" bson:"dedupeKey,omitempty"`
	
	// Read tracking (for group notifications)
	ReadByUsers []ReadStatus `json:"readByUsers,omitempty" bson:"readByUsers,omitempty"`
	
	// For single user notifications (optimization)
	Read   bool       `json:"read" bson:"read"`
	ReadAt *time.Time `json:"readAt,omitempty" bson:"readAt,omitempty"`
	
	// Action tracking
	ActionRequired  bool                 `json:"actionRequired" bson:"actionRequired"`
	Actions         []NotificationAction `json:"actions,omitempty" bson:"actions,omitempty"` // 1-4 actions
	ActionTakenBy   []ActionStatus       `json:"actionTakenBy,omitempty" bson:"actionTakenBy,omitempty"`
}

var NotificationIndices = []repos.IndexField{
	{
		Field: []repos.IndexKey{
			{Key: "id", Value: repos.IndexAsc},
		},
		Unique: true,
	},
	{
		Field: []repos.IndexKey{
			{Key: "target.userId", Value: repos.IndexAsc},
		},
	},
	{
		Field: []repos.IndexKey{
			{Key: "target.teamId", Value: repos.IndexAsc},
		},
	},
	{
		Field: []repos.IndexKey{
			{Key: "target.type", Value: repos.IndexAsc},
		},
	},
	{
		Field: []repos.IndexKey{
			{Key: "dedupeKey", Value: repos.IndexAsc},
		},
		Unique: true,
	},
	{
		Field: []repos.IndexKey{
			{Key: "createdAt", Value: repos.IndexDesc},
		},
	},
	{
		Field: []repos.IndexKey{
			{Key: "actionRequired", Value: repos.IndexAsc},
		},
	},
}

// Helper functions

func (n *Notification) IsReadByUser(userId repos.ID) bool {
	// For user-targeted notifications
	if n.Target.Type == TargetTypeUser && n.Target.UserId != nil && *n.Target.UserId == userId {
		return n.Read
	}
	
	// For group notifications
	for _, readStatus := range n.ReadByUsers {
		if readStatus.UserId == userId {
			return true
		}
	}
	return false
}

func (n *Notification) HasUserTakenAction(userId repos.ID) bool {
	for _, action := range n.ActionTakenBy {
		if action.UserId == userId {
			return true
		}
	}
	return false
}

func (n *Notification) MarkReadByUser(userId repos.ID) {
	// For user-targeted notifications
	if n.Target.Type == TargetTypeUser && n.Target.UserId != nil && *n.Target.UserId == userId {
		n.Read = true
		now := time.Now()
		n.ReadAt = &now
		return
	}
	
	// For group notifications, avoid duplicates
	for _, readStatus := range n.ReadByUsers {
		if readStatus.UserId == userId {
			return
		}
	}
	
	n.ReadByUsers = append(n.ReadByUsers, ReadStatus{
		UserId: userId,
		ReadAt: time.Now(),
	})
}

func (n *Notification) MarkActionTakenByUser(userId repos.ID, actionId string) {
	// Update or add action status
	for i, action := range n.ActionTakenBy {
		if action.UserId == userId {
			// Update existing action
			n.ActionTakenBy[i] = ActionStatus{
				UserId:   userId,
				ActionAt: time.Now(),
				Action:   actionId,
			}
			return
		}
	}
	
	// Add new action
	n.ActionTakenBy = append(n.ActionTakenBy, ActionStatus{
		UserId:   userId,
		ActionAt: time.Now(),
		Action:   actionId,
	})
}