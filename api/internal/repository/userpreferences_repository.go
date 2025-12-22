package repository

import (
	"context"

	platformv1alpha1 "github.com/kloudlite/kloudlite/api/internal/controllers/user/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// UserPreferencesRepository provides operations for UserPreferences resources (cluster-scoped)
type UserPreferencesRepository interface {
	ClusterRepository[*platformv1alpha1.UserPreferences, *platformv1alpha1.UserPreferencesList]

	// GetByUser retrieves preferences for a specific user (by username)
	// Returns empty preferences if not found (does not error)
	GetByUser(ctx context.Context, username string) (*platformv1alpha1.UserPreferences, error)

	// GetOrCreate retrieves preferences or creates empty ones if not found
	GetOrCreate(ctx context.Context, username string) (*platformv1alpha1.UserPreferences, error)

	// AddPinnedWorkspace adds a workspace to pinned list
	AddPinnedWorkspace(ctx context.Context, username string, wsRef platformv1alpha1.ResourceReference) error

	// RemovePinnedWorkspace removes a workspace from pinned list
	RemovePinnedWorkspace(ctx context.Context, username string, wsRef platformv1alpha1.ResourceReference) error

	// AddPinnedEnvironment adds an environment to pinned list
	AddPinnedEnvironment(ctx context.Context, username string, envName string) error

	// RemovePinnedEnvironment removes an environment from pinned list
	RemovePinnedEnvironment(ctx context.Context, username string, envName string) error
}

// userPreferencesRepository implements UserPreferencesRepository
type userPreferencesRepository struct {
	ClusterRepository[*platformv1alpha1.UserPreferences, *platformv1alpha1.UserPreferencesList]
	client client.Client
}

// NewUserPreferencesRepository creates a new UserPreferencesRepository
func NewUserPreferencesRepository(k8sClient client.Client) UserPreferencesRepository {
	baseRepo := NewK8sClusterRepository(
		k8sClient,
		func() *platformv1alpha1.UserPreferences { return &platformv1alpha1.UserPreferences{} },
		func() *platformv1alpha1.UserPreferencesList { return &platformv1alpha1.UserPreferencesList{} },
	)

	return &userPreferencesRepository{
		ClusterRepository: baseRepo,
		client:            k8sClient,
	}
}

// GetByUser retrieves preferences for a specific user
func (r *userPreferencesRepository) GetByUser(ctx context.Context, username string) (*platformv1alpha1.UserPreferences, error) {
	return r.Get(ctx, username)
}

// GetOrCreate retrieves preferences or creates empty ones if not found
func (r *userPreferencesRepository) GetOrCreate(ctx context.Context, username string) (*platformv1alpha1.UserPreferences, error) {
	prefs, err := r.Get(ctx, username)
	if err != nil {
		// If not found, create new empty preferences
		if IsNotFound(err) {
			newPrefs := &platformv1alpha1.UserPreferences{
				ObjectMeta: metav1.ObjectMeta{
					Name: username,
				},
				Spec: platformv1alpha1.UserPreferencesSpec{
					PinnedWorkspaces:   []platformv1alpha1.ResourceReference{},
					PinnedEnvironments: []string{},
				},
			}
			if err := r.Create(ctx, newPrefs); err != nil {
				return nil, err
			}
			return newPrefs, nil
		}
		return nil, err
	}
	return prefs, nil
}

// AddPinnedWorkspace adds a workspace to pinned list
func (r *userPreferencesRepository) AddPinnedWorkspace(ctx context.Context, username string, wsRef platformv1alpha1.ResourceReference) error {
	prefs, err := r.GetOrCreate(ctx, username)
	if err != nil {
		return err
	}

	// Check if already pinned
	for _, pw := range prefs.Spec.PinnedWorkspaces {
		if pw.Name == wsRef.Name && pw.Namespace == wsRef.Namespace {
			return nil // Already pinned, no-op
		}
	}

	// Add to pinned list
	prefs.Spec.PinnedWorkspaces = append(prefs.Spec.PinnedWorkspaces, wsRef)
	now := metav1.Now()
	prefs.Status.LastUpdated = &now

	return r.Update(ctx, prefs)
}

// RemovePinnedWorkspace removes a workspace from pinned list
func (r *userPreferencesRepository) RemovePinnedWorkspace(ctx context.Context, username string, wsRef platformv1alpha1.ResourceReference) error {
	prefs, err := r.GetByUser(ctx, username)
	if err != nil {
		if IsNotFound(err) {
			return nil // No preferences, nothing to remove
		}
		return err
	}

	// Find and remove from pinned list
	newPinned := make([]platformv1alpha1.ResourceReference, 0, len(prefs.Spec.PinnedWorkspaces))
	for _, pw := range prefs.Spec.PinnedWorkspaces {
		if pw.Name != wsRef.Name || pw.Namespace != wsRef.Namespace {
			newPinned = append(newPinned, pw)
		}
	}

	prefs.Spec.PinnedWorkspaces = newPinned
	now := metav1.Now()
	prefs.Status.LastUpdated = &now

	return r.Update(ctx, prefs)
}

// AddPinnedEnvironment adds an environment to pinned list
func (r *userPreferencesRepository) AddPinnedEnvironment(ctx context.Context, username string, envName string) error {
	prefs, err := r.GetOrCreate(ctx, username)
	if err != nil {
		return err
	}

	// Check if already pinned
	for _, pe := range prefs.Spec.PinnedEnvironments {
		if pe == envName {
			return nil // Already pinned, no-op
		}
	}

	// Add to pinned list
	prefs.Spec.PinnedEnvironments = append(prefs.Spec.PinnedEnvironments, envName)
	now := metav1.Now()
	prefs.Status.LastUpdated = &now

	return r.Update(ctx, prefs)
}

// RemovePinnedEnvironment removes an environment from pinned list
func (r *userPreferencesRepository) RemovePinnedEnvironment(ctx context.Context, username string, envName string) error {
	prefs, err := r.GetByUser(ctx, username)
	if err != nil {
		if IsNotFound(err) {
			return nil // No preferences, nothing to remove
		}
		return err
	}

	// Find and remove from pinned list
	newPinned := make([]string, 0, len(prefs.Spec.PinnedEnvironments))
	for _, pe := range prefs.Spec.PinnedEnvironments {
		if pe != envName {
			newPinned = append(newPinned, pe)
		}
	}

	prefs.Spec.PinnedEnvironments = newPinned
	now := metav1.Now()
	prefs.Status.LastUpdated = &now

	return r.Update(ctx, prefs)
}
