package domain

import (
	"context"

	"github.com/kloudlite/api/apps/auth/internal/entities"
	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/pkg/repos"
)

type Domain interface {
	// Authentication
	Login(ctx context.Context, email string, password string) (*entities.User, error)
	LoginWithOAuth(ctx context.Context, email string, name string) (*entities.User, error)
	SignUp(ctx context.Context, name string, email string, password string) (*entities.User, error)
	
	// User Management
	GetUserById(ctx context.Context, id repos.ID) (*entities.User, error)
	GetUserByEmail(ctx context.Context, email string) (*entities.User, error)
	SetUserMetadata(ctx context.Context, userId repos.ID, metadata entities.UserMetadata) (*entities.User, error)
	
	// Email & Password
	VerifyEmail(ctx context.Context, token string) (*common.AuthSession, error)
	ResetPassword(ctx context.Context, token string, password string) (bool, error)
	RequestResetPassword(ctx context.Context, email string) (bool, error)
	ResendVerificationEmail(ctx context.Context, email string) (bool, error)
	ChangePassword(ctx context.Context, id repos.ID, currentPassword string, newPassword string) (bool, error)

	// Device Flow
	InitiateDeviceFlow(ctx context.Context, clientId string) (*entities.DeviceFlow, error)
	PollDeviceToken(ctx context.Context, deviceCode string, clientId string) (*entities.DeviceFlow, error)
	VerifyDeviceCode(ctx context.Context, userCode string, userId repos.ID) error

	// Platform User Management
	InitializePlatform(ctx context.Context, ownerEmail string) error
	GetPlatformUser(ctx context.Context, userId repos.ID) (*entities.PlatformUser, error)
	CreatePlatformUser(ctx context.Context, user entities.PlatformUser) (*entities.PlatformUser, error)
	UpdatePlatformUserRole(ctx context.Context, userId repos.ID, role entities.PlatformRole) (*entities.PlatformUser, error)
	ListPlatformUsers(ctx context.Context, role *entities.PlatformRole) ([]*entities.PlatformUser, error)
	CreateOrUpdatePlatformUser(ctx context.Context, user *entities.PlatformUser) error
	
	// Notifications
	CreateNotification(ctx context.Context, notification *entities.Notification) (*entities.Notification, error)
	ListUserNotifications(ctx context.Context, userId repos.ID, limit, offset int, unreadOnly, actionRequiredOnly bool) ([]*entities.Notification, int, error)
	GetUnreadNotificationCount(ctx context.Context, userId repos.ID) (int, error)
	MarkNotificationAsRead(ctx context.Context, notificationId, userId repos.ID) error
	MarkAllNotificationsAsRead(ctx context.Context, userId repos.ID) (int, error)
	MarkNotificationActionTaken(ctx context.Context, notificationId, userId repos.ID, actionId string) error
}

type Messenger interface {
	SendEmail(ctx context.Context, template string, payload map[string]any) error
}
