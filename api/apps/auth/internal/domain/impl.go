package domain

import (
	"context"
	"crypto/md5"
	crand "crypto/rand"
	"encoding/hex"
	"fmt"
	"log/slog"
	"math/big"
	"strings"
	"time"

	"github.com/kloudlite/api/apps/auth/internal/entities"
	"github.com/kloudlite/api/apps/auth/internal/env"

	"github.com/kloudlite/api/apps/auth/internal/app/email"

	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/pkg/errors"
	"github.com/kloudlite/api/pkg/functions"
	"github.com/kloudlite/api/pkg/kv"
	"github.com/kloudlite/api/pkg/repos"
)

func generateId(prefix string) string {
	id := functions.CleanerNanoidOrDie(28)
	return fmt.Sprintf("%s-%s", prefix, strings.ToLower(id))
}

func newAuthSession(userId repos.ID, userEmail string, userName string, userVerified bool, loginMethod string) *common.AuthSession {
	sessionId := generateId("sess")
	s := &common.AuthSession{
		UserId:       userId,
		UserEmail:    userEmail,
		UserVerified: userVerified,
		UserName:     userName,
		LoginMethod:  loginMethod,
	}
	s.SetId(repos.ID(sessionId))
	return s
}

type domainI struct {
	userRepo         repos.DbRepo[*entities.User]
	emailService     *email.EmailService
	verifyTokenRepo  kv.Repo[*entities.VerifyToken]
	resetTokenRepo   kv.Repo[*entities.ResetPasswordToken]
	logger           *slog.Logger
	deviceFlowRepo   repos.DbRepo[*entities.DeviceFlow]
	platformUserRepo repos.DbRepo[*entities.PlatformUser]
	notificationRepo repos.DbRepo[*entities.Notification]

	envVars *env.AuthEnv
}


func (d *domainI) GetUserById(ctx context.Context, id repos.ID) (*entities.User, error) {
	return d.userRepo.FindById(ctx, id)
}

func (d *domainI) GetUserByEmail(ctx context.Context, email string) (*entities.User, error) {
	return d.userRepo.FindOne(ctx, repos.Filter{"email": email})
}


func (d *domainI) LoginWithOAuth(ctx context.Context, email string, name string) (*entities.User, error) {
	user, err := d.userRepo.FindOne(ctx, repos.Filter{"email": email})
	if err != nil {
		return nil, errors.NewE(err)
	}

	if user == nil {
		user, err = d.userRepo.Create(ctx, &entities.User{
			Email:    email,
			Name:     name,
			Verified: true,
		})
		if err != nil {
			return nil, errors.NewE(err)
		}
		// Send welcome email asynchronously
		go func() {
			bgCtx := context.Background()
			if err := d.emailService.SendWelcomeEmail(bgCtx, user.Email, user.Name); err != nil {
				d.logger.Error("failed to send welcome email", "error", err, "email", user.Email)
			}
		}()
	}

	return user, nil
}

func (d *domainI) Login(ctx context.Context, email string, password string) (*entities.User, error) {
	user, err := d.userRepo.FindOne(ctx, repos.Filter{"email": email})
	if err != nil {
		return nil, errors.NewE(err)
	}

	if user == nil {
		d.logger.Warn("user not found", "email", email)
		return nil, errors.Newf("not valid credentials")
	}

	bytes := md5.Sum([]byte(password + user.PasswordSalt))
	// TODO (nxtcoder17): use crypto/subtle to compare hashes, to avoid timing attacks, also does not work now
	if user.Password != hex.EncodeToString(bytes[:]) {
		return nil, errors.New("not valid credentials")
	}

	return user, nil
}

func (d *domainI) SignUp(ctx context.Context, name string, email string, password string) (*entities.User, error) {
	matched, err := d.userRepo.FindOne(ctx, repos.Filter{"email": email})
	if err != nil {
		// If there's an error finding the user, we should return it
		return nil, errors.NewE(err)
	}

	// If we found a user with this email, return an error
	if matched != nil {
		return nil, errors.Newf("user(email=%q) already exists", email)
	}

	salt := generateId("salt")
	sum := md5.Sum([]byte(password + salt))
	user, err := d.userRepo.Create(
		ctx, &entities.User{
			Name:         name,
			Email:        email,
			Password:     hex.EncodeToString(sum[:]),
			Verified:     !d.envVars.UserEmailVerificationEnabled,
			Approved:     false,
			Metadata:     nil,
			Joined:       time.Now(),
			PasswordSalt: salt,
		},
	)
	if err != nil {
		return nil, errors.NewE(err)
	}

	// Check if this newly created user is a platform user
	platformUser, err := d.GetPlatformUser(ctx, user.Id)
	if err != nil && !errors.Is(err, repos.ErrNoDocuments) {
		// Log error but don't fail signup
		d.logger.Error("failed to check if new user is platform user", "error", err, "userId", user.Id)
	} else if platformUser != nil {
		d.logger.Info("newly signed up user is a platform user", "email", email, "role", platformUser.Role)
	}

	if d.envVars.UserEmailVerificationEnabled {
		// Send verification email asynchronously
		go func() {
			// Create a new context for the goroutine since the original context might be cancelled
			bgCtx := context.Background()
			d.logger.Info("attempting to send verification email", "email", user.Email)
			if err := d.generateAndSendVerificationToken(bgCtx, user); err != nil {
				d.logger.Error("failed to send verification email", "error", err, "email", user.Email)
			} else {
				d.logger.Info("verification email sent successfully", "email", user.Email)
			}
		}()
	} else {
		// Even if email verification is disabled, try to send welcome email asynchronously
		// Don't fail the signup if email sending fails
		go func() {
			bgCtx := context.Background()
			d.logger.Info("attempting to send welcome email", "email", user.Email)
			if err := d.emailService.SendWelcomeEmail(bgCtx, user.Email, user.Name); err != nil {
				d.logger.Error("failed to send welcome email", "error", err, "email", user.Email)
			} else {
				d.logger.Info("welcome email sent successfully", "email", user.Email)
			}
		}()
	}

	return user, nil
}



func (d *domainI) SetUserMetadata(ctx context.Context, userId repos.ID, metadata entities.UserMetadata) (*entities.User, error) {
	user, err := d.userRepo.FindById(ctx, userId)
	if err != nil {
		return nil, errors.NewE(err)
	}
	user.Metadata = metadata
	updated, err := d.userRepo.UpdateById(ctx, userId, user)
	if err != nil {
		return nil, errors.NewE(err)
	}
	return updated, nil
}


func (d *domainI) VerifyEmail(ctx context.Context, token string) (*common.AuthSession, error) {
	v, err := d.verifyTokenRepo.Get(ctx, token)
	if err != nil {
		return nil, errors.NewE(err)
	}
	user, err := d.userRepo.FindById(ctx, v.UserId)
	if err != nil {
		return nil, errors.NewE(err)
	}
	user.Verified = true
	u, err := d.userRepo.UpdateById(ctx, v.UserId, user)
	if err != nil {
		return nil, errors.NewE(err)
	}
	if err := d.emailService.SendWelcomeEmail(ctx, user.Email, user.Name); err != nil {
		d.logger.Error("error occurred", "error", err)
	}

	return newAuthSession(u.Id, u.Email, u.Name, u.Verified, "email/verify"), nil
}

func (d *domainI) ResetPassword(ctx context.Context, token string, password string) (bool, error) {
	get, err := d.resetTokenRepo.Get(ctx, token)
	if err != nil || get == nil {
		return false, errors.NewEf(err, "failed to verify reset password token")
	}

	user, err := d.userRepo.FindById(ctx, get.UserId)
	if err != nil {
		return false, errors.NewEf(err, "unable to find user")
	}
	salt := generateId("salt")
	sum := md5.Sum([]byte(password + salt))
	user.Password = hex.EncodeToString(sum[:])
	user.PasswordSalt = salt
	_, err = d.userRepo.UpdateById(ctx, repos.ID(get.UserId), user)
	if err != nil {
		return false, errors.NewE(err)
	}

	err = d.resetTokenRepo.Drop(ctx, token)
	if err != nil {
		// TODO silent fail
		d.logger.Error("could not delete resetPassword token", "error", err)
		return false, nil
	}
	return true, nil
}

func (d *domainI) RequestResetPassword(ctx context.Context, email string) (bool, error) {
	resetToken := generateId("reset")
	one, err := d.userRepo.FindOne(ctx, repos.Filter{"email": email})
	if err != nil {
		return false, errors.NewE(err)
	}
	if one == nil {
		return false, errors.New("no account present with provided email, register your account first.")
	}
	err = d.resetTokenRepo.SetWithExpiry(
		ctx,
		resetToken,
		&entities.ResetPasswordToken{Token: resetToken, UserId: one.Id},
		time.Second*24*60*60,
	)
	if err != nil {
		return false, errors.NewE(err)
	}
	err = d.sendResetPasswordEmail(ctx, resetToken, one)
	if err != nil {
		return false, errors.NewE(err)
	}
	return true, nil
}


func (d *domainI) ResendVerificationEmail(ctx context.Context, email string) (bool, error) {
	user, err := d.userRepo.FindOne(ctx, repos.Filter{"email": email})
	if err != nil {
		return false, errors.NewE(err)
	}
	err = d.generateAndSendVerificationToken(ctx, user)
	if err != nil {
		return false, errors.NewE(err)
	}
	return true, errors.NewE(err)
}

func (d *domainI) ChangePassword(ctx context.Context, id repos.ID, currentPassword string, newPassword string) (bool, error) {
	user, err := d.userRepo.FindById(ctx, id)
	if err != nil {
		return false, errors.NewE(err)
	}
	sum := md5.Sum([]byte(currentPassword + user.PasswordSalt))
	if user.Password == hex.EncodeToString(sum[:]) {
		salt := generateId("salt")
		user.PasswordSalt = salt
		newSum := md5.Sum([]byte(newPassword + user.PasswordSalt))
		user.Password = hex.EncodeToString(newSum[:])
		_, err := d.userRepo.UpdateById(ctx, id, user)
		if err != nil {
			return false, errors.NewE(err)
		}
		// TODO send comm
		return true, nil
	}
	return false, errors.New("invalid credentials")
}

func (d *domainI) sendResetPasswordEmail(ctx context.Context, token string, user *entities.User) error {
	err := d.emailService.SendPasswordResetEmail(ctx, user.Email, user.Name, token)
	if err != nil {
		return errors.NewEf(err, "could not send password reset email")
	}
	return nil
}

func (d *domainI) sendVerificationEmail(ctx context.Context, token string, user *entities.User) error {
	err := d.emailService.SendVerificationEmail(ctx, user.Email, user.Name, token)
	if err != nil {
		return errors.NewEf(err, "could not send verification email")
	}
	return nil
}

func (d *domainI) generateAndSendVerificationToken(ctx context.Context, user *entities.User) error {
	verificationToken := generateId("invite")
	err := d.verifyTokenRepo.SetWithExpiry(
		ctx, verificationToken, &entities.VerifyToken{
			Token:  verificationToken,
			UserId: user.Id,
		}, time.Second*24*60*60,
	)
	if err != nil {
		return errors.NewE(err)
	}
	err = d.sendVerificationEmail(ctx, verificationToken, user)
	if err != nil {
		return errors.NewE(err)
	}
	return nil
}

func fxDomain(
	userRepo repos.DbRepo[*entities.User],
	verifyTokenRepo kv.Repo[*entities.VerifyToken],
	resetTokenRepo kv.Repo[*entities.ResetPasswordToken],
	deviceFlowRepo repos.DbRepo[*entities.DeviceFlow],
	platformUserRepo repos.DbRepo[*entities.PlatformUser],
	notificationRepo repos.DbRepo[*entities.Notification],
	emailService *email.EmailService,
	ev *env.AuthEnv,
	logger *slog.Logger,
) Domain {
	return &domainI{
		emailService:     emailService,
		userRepo:         userRepo,
		verifyTokenRepo:  verifyTokenRepo,
		resetTokenRepo:   resetTokenRepo,
		deviceFlowRepo:   deviceFlowRepo,
		platformUserRepo: platformUserRepo,
		notificationRepo: notificationRepo,
		envVars:          ev,
		logger:           logger,
	}
}

// generateUserCode generates a user-friendly 8-character code
func generateUserCode() string {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 8)
	for i := range b {
		n, _ := crand.Int(crand.Reader, big.NewInt(int64(len(charset))))
		b[i] = charset[n.Int64()]
	}
	return string(b)
}

// Device Flow Implementation
func (d *domainI) InitiateDeviceFlow(ctx context.Context, clientId string) (*entities.DeviceFlow, error) {
	// Generate unique codes
	deviceCode := generateId("dev")
	userCode := generateUserCode() // 8 character user-friendly code
	
	deviceFlow := &entities.DeviceFlow{
		DeviceCode: deviceCode,
		UserCode:   userCode,
		ClientID:   clientId,
		Authorized: false,
		ExpiresAt:  time.Now().Add(15 * time.Minute), // 15 minute expiry
	}
	
	created, err := d.deviceFlowRepo.Create(ctx, deviceFlow)
	if err != nil {
		return nil, errors.NewE(err)
	}
	
	return created, nil
}

func (d *domainI) PollDeviceToken(ctx context.Context, deviceCode string, clientId string) (*entities.DeviceFlow, error) {
	deviceFlow, err := d.deviceFlowRepo.FindOne(ctx, repos.Filter{
		"deviceCode": deviceCode,
		"clientId":   clientId,
	})
	if err != nil {
		return nil, errors.NewE(err)
	}
	
	if deviceFlow == nil {
		return nil, errors.Newf("device flow not found")
	}
	
	return deviceFlow, nil
}

func (d *domainI) VerifyDeviceCode(ctx context.Context, userCode string, userId repos.ID) error {
	deviceFlow, err := d.deviceFlowRepo.FindOne(ctx, repos.Filter{
		"userCode": userCode,
	})
	if err != nil {
		d.logger.Error("Failed to find device flow", 
			slog.String("error", err.Error()), 
			slog.String("userCode", userCode))
		return errors.NewE(err)
	}
	
	if deviceFlow == nil {
		return errors.Newf("invalid user code")
	}
	
	// Check if expired
	if time.Now().After(deviceFlow.ExpiresAt) {
		d.logger.Error("Device flow expired", 
			slog.String("userCode", userCode),
			slog.Time("expiresAt", deviceFlow.ExpiresAt),
			slog.Time("now", time.Now()))
		return errors.Newf("user code expired")
	}
	
	// Update device flow with user authorization
	_, err = d.deviceFlowRepo.PatchOne(ctx, 
		repos.Filter{"id": deviceFlow.Id}, 
		repos.Document{
			"authorized": true,
			"userId": string(userId),
		})
	if err != nil {
		d.logger.Error("Failed to update device flow", 
			slog.String("error", err.Error()),
			slog.String("deviceFlowId", string(deviceFlow.Id)))
		return errors.NewE(err)
	}
	
	return nil
}

// Platform User Management
func (d *domainI) InitializePlatform(ctx context.Context, ownerEmail string) error {
	// Check if platform is already initialized (any super_admin exists)
	superAdminRole := entities.PlatformRoleSuperAdmin
	superAdmins, err := d.ListPlatformUsers(ctx, &superAdminRole)
	if err != nil {
		return err
	}
	if len(superAdmins) > 0 {
		d.logger.Info("platform already initialized", "existingSuperAdmins", len(superAdmins))
		return nil
	}

	// Check if user with this email already exists
	existingUser, err := d.GetUserByEmail(ctx, ownerEmail)
	if err != nil && !errors.Is(err, repos.ErrNoDocuments) {
		return err
	}

	var userId repos.ID
	if existingUser != nil {
		// User already exists, use their ID
		userId = existingUser.Id
		d.logger.Info("using existing user for platform owner", "email", ownerEmail, "userId", userId)
	} else {
		// Create a new user account for the platform owner
		// Generate a random password (owner will need to reset it)
		tempPassword := generateId("temp")
		salt := generateId("salt")
		sum := md5.Sum([]byte(tempPassword + salt))
		
		newUser := &entities.User{
			Name:         "Platform Owner",
			Email:        ownerEmail,
			Password:     hex.EncodeToString(sum[:]),
			PasswordSalt: salt,
			Verified:     true, // Auto-verify platform owner
			Approved:     true,
			Metadata:     nil,
			Joined:       time.Now(),
		}
		
		createdUser, err := d.userRepo.Create(ctx, newUser)
		if err != nil {
			return errors.NewE(err)
		}
		
		userId = createdUser.Id
		d.logger.Info("created user account for platform owner", "email", ownerEmail, "userId", userId)
		
		// Send password reset email so owner can set their password
		go func() {
			bgCtx := context.Background()
			resetToken := generateId("reset")
			err := d.resetTokenRepo.SetWithExpiry(
				bgCtx,
				resetToken,
				&entities.ResetPasswordToken{Token: resetToken, UserId: userId},
				time.Hour*24*7, // 7 days for initial setup
			)
			if err != nil {
				d.logger.Error("failed to create reset token for platform owner", "error", err)
				return
			}
			
			if err := d.sendResetPasswordEmail(bgCtx, resetToken, createdUser); err != nil {
				d.logger.Error("failed to send password reset email to platform owner", "error", err)
			} else {
				d.logger.Info("sent password reset email to platform owner", "email", ownerEmail)
			}
		}()
	}

	// Create platform user entry
	platformUser := entities.PlatformUser{
		UserId:   userId,
		Role:     entities.PlatformRoleSuperAdmin,
	}
	platformUser.Id = d.platformUserRepo.NewId()
	platformUser.CreatedBy = common.CreatedOrUpdatedByKloudlite
	platformUser.LastUpdatedBy = common.CreatedOrUpdatedByKloudlite

	_, err = d.platformUserRepo.Create(ctx, &platformUser)
	if err != nil {
		return errors.NewE(err)
	}

	d.logger.Info("platform initialized successfully", "ownerEmail", ownerEmail)
	if existingUser == nil {
		d.logger.Info("Platform owner account created. Password reset email sent to:", "email", ownerEmail)
	}
	
	return nil
}

func (d *domainI) GetPlatformUser(ctx context.Context, userId repos.ID) (*entities.PlatformUser, error) {
	platformUser, err := d.platformUserRepo.FindOne(ctx, repos.Filter{
		"userId": userId,
	})
	if err != nil {
		return nil, err
	}
	return platformUser, nil
}


func (d *domainI) CreatePlatformUser(ctx context.Context, user entities.PlatformUser) (*entities.PlatformUser, error) {
	user.Id = d.platformUserRepo.NewId()
	if user.CreatedBy.UserId == "" {
		user.CreatedBy = common.CreatedOrUpdatedByKloudlite
	}
	user.LastUpdatedBy = user.CreatedBy

	platformUser, err := d.platformUserRepo.Create(ctx, &user)
	if err != nil {
		return nil, err
	}

	return platformUser, nil
}

func (d *domainI) UpdatePlatformUserRole(ctx context.Context, userId repos.ID, role entities.PlatformRole) (*entities.PlatformUser, error) {
	platformUser, err := d.GetPlatformUser(ctx, userId)
	if err != nil {
		return nil, err
	}

	platformUser.Role = role
	platformUser.LastUpdatedBy = common.CreatedOrUpdatedByKloudlite

	updatedUser, err := d.platformUserRepo.UpdateById(ctx, platformUser.Id, platformUser)
	if err != nil {
		return nil, err
	}

	return updatedUser, nil
}

func (d *domainI) ListPlatformUsers(ctx context.Context, role *entities.PlatformRole) ([]*entities.PlatformUser, error) {
	filter := repos.Filter{}
	if role != nil {
		filter["role"] = *role
	}

	users, err := d.platformUserRepo.Find(ctx, repos.Query{
		Filter: filter,
		Sort: map[string]interface{}{
			"createdAt": -1,
		},
	})
	if err != nil {
		return nil, err
	}

	return users, nil
}

func (d *domainI) CreateOrUpdatePlatformUser(ctx context.Context, user *entities.PlatformUser) error {
	// Check if platform user already exists
	existingPlatformUser, err := d.GetPlatformUser(ctx, user.UserId)
	if err != nil && !errors.Is(err, repos.ErrNoDocuments) {
		return errors.NewE(err)
	}

	if existingPlatformUser != nil {
		// Update existing platform user
		existingPlatformUser.Role = user.Role
		existingPlatformUser.LastUpdatedBy = common.CreatedOrUpdatedByKloudlite

		_, err = d.platformUserRepo.UpdateById(ctx, existingPlatformUser.Id, existingPlatformUser)
		if err != nil {
			return errors.NewE(err)
		}
		d.logger.Info("updated existing platform user", "userId", user.UserId, "role", user.Role)
	} else {
		// Create new platform user
		newPlatformUser := &entities.PlatformUser{
			UserId:   user.UserId,
			Role:     user.Role,
		}
		newPlatformUser.Id = d.platformUserRepo.NewId()
		newPlatformUser.CreatedBy = common.CreatedOrUpdatedByKloudlite
		newPlatformUser.LastUpdatedBy = common.CreatedOrUpdatedByKloudlite

		_, err = d.platformUserRepo.Create(ctx, newPlatformUser)
		if err != nil {
			return errors.NewE(err)
		}
		d.logger.Info("created new platform user", "userId", user.UserId, "role", user.Role)
	}

	return nil
}

// Notification methods
func (d *domainI) CreateNotification(ctx context.Context, notification *entities.Notification) (*entities.Notification, error) {
	if notification.Id == "" {
		notification.Id = repos.ID(generateId("notif"))
	}
	
	notification.IncrementRecordVersion()
	
	// Handle deduplication
	if notification.DedupeKey != "" {
		existingNotif, err := d.notificationRepo.FindOne(ctx, repos.Filter{"dedupeKey": notification.DedupeKey})
		if err != nil && !errors.Is(err, repos.ErrNoDocuments) {
			return nil, errors.NewE(err)
		}
		if existingNotif != nil {
			// Return existing notification instead of creating duplicate
			return existingNotif, nil
		}
	}
	
	created, err := d.notificationRepo.Create(ctx, notification)
	if err != nil {
		return nil, errors.NewE(err)
	}
	
	return created, nil
}

func (d *domainI) ListUserNotifications(ctx context.Context, userId repos.ID, limit, offset int, unreadOnly, actionRequiredOnly bool) ([]*entities.Notification, int, error) {
	// First, check if user exists
	_, err := d.userRepo.FindById(ctx, userId)
	if err != nil {
		return nil, 0, errors.NewE(err)
	}
	
	// Get platform user to check platform role
	platformUser, _ := d.platformUserRepo.FindOne(ctx, repos.Filter{"userId": userId})
	
	// Build complex query for notifications
	orFilters := []repos.Filter{
		// Direct user notifications
		{
			"target.type": string(entities.TargetTypeUser),
			"target.userId": userId,
		},
	}
	
	// Platform role notifications
	if platformUser != nil {
		platformRoleValue := entities.PlatformRoleHierarchy[string(platformUser.Role)]
		// Add filters for each role level the user has access to
		for role, value := range entities.PlatformRoleHierarchy {
			if value <= platformRoleValue {
				orFilters = append(orFilters, repos.Filter{
					"target.type": string(entities.TargetTypePlatformRole),
					"target.minPlatformRole": role,
				})
			}
		}
	}
	
	// TODO: Add team role notifications when we have access to team memberships
	// This would require injecting a team service or having a way to query team memberships
	
	// Apply additional filters
	mainFilter := repos.Filter{
		"$or": orFilters,
	}
	
	if actionRequiredOnly {
		mainFilter["actionRequired"] = true
	}
	
	// Count total matching notifications
	totalCount, err := d.notificationRepo.Count(ctx, mainFilter)
	if err != nil {
		return nil, 0, errors.NewE(err)
	}
	
	// For pagination, we'll use FindPaginated with cursor pagination
	// First get all notifications to filter by unread status if needed
	query := repos.Query{
		Filter: mainFilter,
		Sort:   map[string]interface{}{"createdAt": -1},
	}
	
	allNotifications, err := d.notificationRepo.Find(ctx, query)
	if err != nil {
		return nil, 0, errors.NewE(err)
	}
	
	// Apply pagination manually since we need to filter by unread status
	start := offset
	end := offset + limit
	if start > len(allNotifications) {
		start = len(allNotifications)
	}
	if end > len(allNotifications) {
		end = len(allNotifications)
	}
	
	notifications := allNotifications
	if limit > 0 {
		notifications = allNotifications[start:end]
	}
	
	// Filter by unread if requested (post-query filtering)
	if unreadOnly {
		var unreadNotifs []*entities.Notification
		for _, notif := range notifications {
			if !notif.IsReadByUser(userId) {
				// For actionable notifications, also check if action is taken
				if notif.ActionRequired && notif.HasUserTakenAction(userId) {
					continue
				}
				unreadNotifs = append(unreadNotifs, notif)
			}
		}
		notifications = unreadNotifs
	}
	
	return notifications, int(totalCount), nil
}

func (d *domainI) GetUnreadNotificationCount(ctx context.Context, userId repos.ID) (int, error) {
	// Get all user notifications
	notifications, _, err := d.ListUserNotifications(ctx, userId, 1000, 0, false, false)
	if err != nil {
		return 0, err
	}
	
	unreadCount := 0
	for _, notif := range notifications {
		// Personal notifications count if unread
		if notif.Target.Type == entities.TargetTypeUser && !notif.Read {
			unreadCount++
			continue
		}
		
		// Group actionable notifications count if action not taken
		if notif.Target.Type != entities.TargetTypeUser && notif.ActionRequired && !notif.HasUserTakenAction(userId) {
			unreadCount++
		}
		// Non-actionable group notifications don't count
	}
	
	return unreadCount, nil
}

func (d *domainI) MarkNotificationAsRead(ctx context.Context, notificationId, userId repos.ID) error {
	notification, err := d.notificationRepo.FindById(ctx, notificationId)
	if err != nil {
		return errors.NewE(err)
	}
	
	notification.MarkReadByUser(userId)
	notification.IncrementRecordVersion()
	
	_, err = d.notificationRepo.UpdateById(ctx, notificationId, notification)
	if err != nil {
		return errors.NewE(err)
	}
	
	return nil
}

func (d *domainI) MarkAllNotificationsAsRead(ctx context.Context, userId repos.ID) (int, error) {
	// Get all unread notifications for the user
	notifications, _, err := d.ListUserNotifications(ctx, userId, 1000, 0, true, false)
	if err != nil {
		return 0, err
	}
	
	markedCount := 0
	for _, notif := range notifications {
		if err := d.MarkNotificationAsRead(ctx, notif.Id, userId); err != nil {
			d.logger.Error("failed to mark notification as read", "notificationId", notif.Id, "error", err)
			continue
		}
		markedCount++
	}
	
	return markedCount, nil
}

func (d *domainI) MarkNotificationActionTaken(ctx context.Context, notificationId, userId repos.ID, actionId string) error {
	notification, err := d.notificationRepo.FindById(ctx, notificationId)
	if err != nil {
		return errors.NewE(err)
	}
	
	notification.MarkActionTakenByUser(userId, actionId)
	notification.IncrementRecordVersion()
	
	_, err = d.notificationRepo.UpdateById(ctx, notificationId, notification)
	if err != nil {
		return errors.NewE(err)
	}
	
	return nil
}