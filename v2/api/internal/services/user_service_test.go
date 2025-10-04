package services

import (
	"context"
	"errors"
	"testing"

	"github.com/kloudlite/kloudlite/v2/api/internal/repository"
	machinesv1 "github.com/kloudlite/kloudlite/v2/api/pkg/apis/machines/v1"
	platformv1alpha1 "github.com/kloudlite/kloudlite/v2/api/pkg/apis/platform/v1alpha1"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Mock UserRepository
type mockUserRepo struct {
	createFunc       func(ctx context.Context, user *platformv1alpha1.User) error
	getFunc          func(ctx context.Context, name string) (*platformv1alpha1.User, error)
	getByUsernameFunc func(ctx context.Context, username string) (*platformv1alpha1.User, error)
	updateFunc       func(ctx context.Context, user *platformv1alpha1.User) error
	deleteFunc       func(ctx context.Context, name string) error
	listFunc         func(ctx context.Context, opts ...repository.ListOption) (*platformv1alpha1.UserList, error)
	patchFunc        func(ctx context.Context, name string, patchData map[string]interface{}) (*platformv1alpha1.User, error)
	updateStatusFunc func(ctx context.Context, user *platformv1alpha1.User) error
}

func (m *mockUserRepo) Create(ctx context.Context, user *platformv1alpha1.User) error {
	if m.createFunc != nil {
		return m.createFunc(ctx, user)
	}
	return errors.New("not implemented")
}

func (m *mockUserRepo) Get(ctx context.Context, name string) (*platformv1alpha1.User, error) {
	if m.getFunc != nil {
		return m.getFunc(ctx, name)
	}
	return nil, errors.New("not implemented")
}

func (m *mockUserRepo) GetByUsername(ctx context.Context, username string) (*platformv1alpha1.User, error) {
	if m.getByUsernameFunc != nil {
		return m.getByUsernameFunc(ctx, username)
	}
	return nil, errors.New("not implemented")
}

func (m *mockUserRepo) Update(ctx context.Context, user *platformv1alpha1.User) error {
	if m.updateFunc != nil {
		return m.updateFunc(ctx, user)
	}
	return errors.New("not implemented")
}

func (m *mockUserRepo) Delete(ctx context.Context, name string) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(ctx, name)
	}
	return errors.New("not implemented")
}

func (m *mockUserRepo) List(ctx context.Context, opts ...repository.ListOption) (*platformv1alpha1.UserList, error) {
	if m.listFunc != nil {
		return m.listFunc(ctx, opts...)
	}
	return nil, errors.New("not implemented")
}

func (m *mockUserRepo) Patch(ctx context.Context, name string, patchData map[string]interface{}) (*platformv1alpha1.User, error) {
	if m.patchFunc != nil {
		return m.patchFunc(ctx, name, patchData)
	}
	return nil, errors.New("not implemented")
}

func (m *mockUserRepo) UpdateStatus(ctx context.Context, user *platformv1alpha1.User) error {
	if m.updateStatusFunc != nil {
		return m.updateStatusFunc(ctx, user)
	}
	return errors.New("not implemented")
}

func (m *mockUserRepo) GetByEmail(ctx context.Context, email string) (*platformv1alpha1.User, error) {
	return nil, errors.New("not implemented")
}

func (m *mockUserRepo) ListActive(ctx context.Context) (*platformv1alpha1.UserList, error) {
	return nil, errors.New("not implemented")
}

func (m *mockUserRepo) Watch(ctx context.Context, opts ...repository.WatchOption) (<-chan repository.WatchEvent[*platformv1alpha1.User], error) {
	return nil, errors.New("not implemented")
}

// Mock WorkMachineRepository
type mockWorkMachineRepo struct {
	createFunc func(ctx context.Context, wm *machinesv1.WorkMachine) error
	getFunc    func(ctx context.Context, name string) (*machinesv1.WorkMachine, error)
	updateFunc func(ctx context.Context, wm *machinesv1.WorkMachine) error
	deleteFunc func(ctx context.Context, name string) error
	listFunc   func(ctx context.Context, opts ...repository.ListOption) (*machinesv1.WorkMachineList, error)
}

func (m *mockWorkMachineRepo) Create(ctx context.Context, wm *machinesv1.WorkMachine) error {
	if m.createFunc != nil {
		return m.createFunc(ctx, wm)
	}
	return errors.New("not implemented")
}

func (m *mockWorkMachineRepo) Get(ctx context.Context, name string) (*machinesv1.WorkMachine, error) {
	if m.getFunc != nil {
		return m.getFunc(ctx, name)
	}
	return nil, errors.New("not implemented")
}

func (m *mockWorkMachineRepo) Update(ctx context.Context, wm *machinesv1.WorkMachine) error {
	if m.updateFunc != nil {
		return m.updateFunc(ctx, wm)
	}
	return errors.New("not implemented")
}

func (m *mockWorkMachineRepo) Delete(ctx context.Context, name string) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(ctx, name)
	}
	return errors.New("not implemented")
}

func (m *mockWorkMachineRepo) List(ctx context.Context, opts ...repository.ListOption) (*machinesv1.WorkMachineList, error) {
	if m.listFunc != nil {
		return m.listFunc(ctx, opts...)
	}
	return nil, errors.New("not implemented")
}

func (m *mockWorkMachineRepo) Patch(ctx context.Context, name string, patchData map[string]interface{}) (*machinesv1.WorkMachine, error) {
	return nil, errors.New("not implemented")
}

func (m *mockWorkMachineRepo) Watch(ctx context.Context, opts ...repository.WatchOption) (<-chan repository.WatchEvent[*machinesv1.WorkMachine], error) {
	return nil, errors.New("not implemented")
}

func (m *mockWorkMachineRepo) GetByOwner(ctx context.Context, owner string) (*machinesv1.WorkMachine, error) {
	return nil, errors.New("not implemented")
}

func (m *mockWorkMachineRepo) StartMachine(ctx context.Context, name string) error {
	return errors.New("not implemented")
}

func (m *mockWorkMachineRepo) StopMachine(ctx context.Context, name string) error {
	return errors.New("not implemented")
}

func (m *mockWorkMachineRepo) ListByMachineType(ctx context.Context, machineType string) (*machinesv1.WorkMachineList, error) {
	return nil, errors.New("not implemented")
}

func TestCreateUser(t *testing.T) {
	ctx := context.Background()
	activeTrue := true

	t.Run("should create user successfully", func(t *testing.T) {
		userRepo := &mockUserRepo{
			createFunc: func(ctx context.Context, user *platformv1alpha1.User) error {
				return nil
			},
		}

		wmRepo := &mockWorkMachineRepo{
			getFunc: func(ctx context.Context, name string) (*machinesv1.WorkMachine, error) {
				return nil, repository.ErrNotFound("not found")
			},
			createFunc: func(ctx context.Context, wm *machinesv1.WorkMachine) error {
				assert.Equal(t, "wm-test", wm.Name)
				assert.Equal(t, "test@example.com", wm.Spec.OwnedBy)
				assert.Equal(t, "wm-test", wm.Spec.TargetNamespace)
				return nil
			},
		}

		service := NewUserService(userRepo, wmRepo)

		user := &platformv1alpha1.User{
			ObjectMeta: metav1.ObjectMeta{Name: "test-user"},
			Spec: platformv1alpha1.UserSpec{
				Email:  "test@example.com",
				Active: &activeTrue,
				Roles:  []platformv1alpha1.RoleType{platformv1alpha1.RoleUser},
			},
		}

		createdUser, err := service.CreateUser(ctx, user)
		assert.NoError(t, err)
		assert.NotNil(t, createdUser)
		assert.Equal(t, "test@example.com", createdUser.Spec.Email)
	})

	t.Run("should handle user already exists", func(t *testing.T) {
		userRepo := &mockUserRepo{
			createFunc: func(ctx context.Context, user *platformv1alpha1.User) error {
				return repository.ErrAlreadyExists("user already exists")
			},
		}

		wmRepo := &mockWorkMachineRepo{}
		service := NewUserService(userRepo, wmRepo)

		user := &platformv1alpha1.User{
			ObjectMeta: metav1.ObjectMeta{Name: "test-user"},
			Spec: platformv1alpha1.UserSpec{
				Email: "test@example.com",
			},
		}

		createdUser, err := service.CreateUser(ctx, user)
		assert.Error(t, err)
		assert.Nil(t, createdUser)
		assert.Contains(t, err.Error(), "user already exists")
	})

	t.Run("should create user even if WorkMachine creation fails", func(t *testing.T) {
		userRepo := &mockUserRepo{
			createFunc: func(ctx context.Context, user *platformv1alpha1.User) error {
				return nil
			},
		}

		wmRepo := &mockWorkMachineRepo{
			getFunc: func(ctx context.Context, name string) (*machinesv1.WorkMachine, error) {
				return nil, repository.ErrNotFound("not found")
			},
			createFunc: func(ctx context.Context, wm *machinesv1.WorkMachine) error {
				return errors.New("workmachine creation failed")
			},
		}

		service := NewUserService(userRepo, wmRepo)

		user := &platformv1alpha1.User{
			ObjectMeta: metav1.ObjectMeta{Name: "test-user"},
			Spec: platformv1alpha1.UserSpec{
				Email:  "test@example.com",
				Active: &activeTrue,
			},
		}

		createdUser, err := service.CreateUser(ctx, user)
		assert.NoError(t, err) // User creation should succeed even if WorkMachine fails
		assert.NotNil(t, createdUser)
	})

	t.Run("should sanitize email for WorkMachine name", func(t *testing.T) {
		userRepo := &mockUserRepo{
			createFunc: func(ctx context.Context, user *platformv1alpha1.User) error {
				return nil
			},
		}

		wmRepo := &mockWorkMachineRepo{
			getFunc: func(ctx context.Context, name string) (*machinesv1.WorkMachine, error) {
				return nil, repository.ErrNotFound("not found")
			},
			createFunc: func(ctx context.Context, wm *machinesv1.WorkMachine) error {
				// Should sanitize john.doe@example.com to john-doe
				assert.Equal(t, "wm-john-doe", wm.Name)
				return nil
			},
		}

		service := NewUserService(userRepo, wmRepo)

		user := &platformv1alpha1.User{
			ObjectMeta: metav1.ObjectMeta{Name: "test-user"},
			Spec: platformv1alpha1.UserSpec{
				Email:  "john.doe@example.com",
				Active: &activeTrue,
			},
		}

		_, err := service.CreateUser(ctx, user)
		assert.NoError(t, err)
	})
}

func TestGetUser(t *testing.T) {
	ctx := context.Background()
	activeTrue := true

	t.Run("should get user by name", func(t *testing.T) {
		userRepo := &mockUserRepo{
			getFunc: func(ctx context.Context, name string) (*platformv1alpha1.User, error) {
				return &platformv1alpha1.User{
					ObjectMeta: metav1.ObjectMeta{Name: name},
					Spec: platformv1alpha1.UserSpec{
						Email:  "test@example.com",
						Active: &activeTrue,
					},
				}, nil
			},
		}

		wmRepo := &mockWorkMachineRepo{}
		service := NewUserService(userRepo, wmRepo)

		user, err := service.GetUser(ctx, "test-user")
		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, "test-user", user.Name)
		assert.Equal(t, "test@example.com", user.Spec.Email)
	})

	t.Run("should return error for non-existent user", func(t *testing.T) {
		userRepo := &mockUserRepo{
			getFunc: func(ctx context.Context, name string) (*platformv1alpha1.User, error) {
				return nil, repository.ErrNotFound("user not found")
			},
		}

		wmRepo := &mockWorkMachineRepo{}
		service := NewUserService(userRepo, wmRepo)

		user, err := service.GetUser(ctx, "nonexistent")
		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Contains(t, err.Error(), "user not found")
	})
}

func TestGetUserByEmail(t *testing.T) {
	ctx := context.Background()
	activeTrue := true

	t.Run("should get user by email", func(t *testing.T) {
		userRepo := &mockUserRepo{
			listFunc: func(ctx context.Context, opts ...repository.ListOption) (*platformv1alpha1.UserList, error) {
				return &platformv1alpha1.UserList{
					Items: []platformv1alpha1.User{
						{
							ObjectMeta: metav1.ObjectMeta{Name: "user1"},
							Spec: platformv1alpha1.UserSpec{
								Email:  "user1@example.com",
								Active: &activeTrue,
							},
						},
						{
							ObjectMeta: metav1.ObjectMeta{Name: "user2"},
							Spec: platformv1alpha1.UserSpec{
								Email:  "user2@example.com",
								Active: &activeTrue,
							},
						},
					},
				}, nil
			},
		}

		wmRepo := &mockWorkMachineRepo{}
		service := NewUserService(userRepo, wmRepo)

		user, err := service.GetUserByEmail(ctx, "user2@example.com")
		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, "user2", user.Name)
		assert.Equal(t, "user2@example.com", user.Spec.Email)
	})

	t.Run("should return error for non-existent email", func(t *testing.T) {
		userRepo := &mockUserRepo{
			listFunc: func(ctx context.Context, opts ...repository.ListOption) (*platformv1alpha1.UserList, error) {
				return &platformv1alpha1.UserList{
					Items: []platformv1alpha1.User{
						{
							ObjectMeta: metav1.ObjectMeta{Name: "user1"},
							Spec: platformv1alpha1.UserSpec{
								Email:  "user1@example.com",
								Active: &activeTrue,
							},
						},
					},
				}, nil
			},
		}

		wmRepo := &mockWorkMachineRepo{}
		service := NewUserService(userRepo, wmRepo)

		user, err := service.GetUserByEmail(ctx, "nonexistent@example.com")
		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Contains(t, err.Error(), "user not found with email")
	})

	t.Run("should handle list error", func(t *testing.T) {
		userRepo := &mockUserRepo{
			listFunc: func(ctx context.Context, opts ...repository.ListOption) (*platformv1alpha1.UserList, error) {
				return nil, errors.New("list failed")
			},
		}

		wmRepo := &mockWorkMachineRepo{}
		service := NewUserService(userRepo, wmRepo)

		user, err := service.GetUserByEmail(ctx, "test@example.com")
		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Contains(t, err.Error(), "failed to list users")
	})
}

// Note: GetUserByUsername test skipped as Username field is not yet implemented in UserSpec

func TestUpdateUser(t *testing.T) {
	ctx := context.Background()
	activeTrue := true
	activeFalse := false

	t.Run("should update user successfully", func(t *testing.T) {
		userRepo := &mockUserRepo{
			getFunc: func(ctx context.Context, name string) (*platformv1alpha1.User, error) {
				return &platformv1alpha1.User{
					ObjectMeta: metav1.ObjectMeta{
						Name:            name,
						ResourceVersion: "1",
					},
					Spec: platformv1alpha1.UserSpec{
						Email:  "test@example.com",
						Active: &activeTrue,
					},
				}, nil
			},
			updateFunc: func(ctx context.Context, user *platformv1alpha1.User) error {
				assert.Equal(t, "test-user", user.Name)
				assert.False(t, *user.Spec.Active)
				return nil
			},
		}

		wmRepo := &mockWorkMachineRepo{}
		service := NewUserService(userRepo, wmRepo)

		user := &platformv1alpha1.User{
			ObjectMeta: metav1.ObjectMeta{Name: "test-user"},
			Spec: platformv1alpha1.UserSpec{
				Email:  "test@example.com",
				Active: &activeFalse,
			},
		}

		updatedUser, err := service.UpdateUser(ctx, user)
		assert.NoError(t, err)
		assert.NotNil(t, updatedUser)
		assert.False(t, *updatedUser.Spec.Active)
	})

	t.Run("should return error for non-existent user", func(t *testing.T) {
		userRepo := &mockUserRepo{
			getFunc: func(ctx context.Context, name string) (*platformv1alpha1.User, error) {
				return nil, repository.ErrNotFound("user not found")
			},
		}

		wmRepo := &mockWorkMachineRepo{}
		service := NewUserService(userRepo, wmRepo)

		user := &platformv1alpha1.User{
			ObjectMeta: metav1.ObjectMeta{Name: "nonexistent"},
			Spec: platformv1alpha1.UserSpec{
				Email: "test@example.com",
			},
		}

		updatedUser, err := service.UpdateUser(ctx, user)
		assert.Error(t, err)
		assert.Nil(t, updatedUser)
		assert.Contains(t, err.Error(), "user not found")
	})

	t.Run("should handle update conflict", func(t *testing.T) {
		userRepo := &mockUserRepo{
			getFunc: func(ctx context.Context, name string) (*platformv1alpha1.User, error) {
				return &platformv1alpha1.User{
					ObjectMeta: metav1.ObjectMeta{
						Name:            name,
						ResourceVersion: "1",
					},
					Spec: platformv1alpha1.UserSpec{
						Email:  "test@example.com",
						Active: &activeTrue,
					},
				}, nil
			},
			updateFunc: func(ctx context.Context, user *platformv1alpha1.User) error {
				return repository.ErrConflict("conflict")
			},
		}

		wmRepo := &mockWorkMachineRepo{}
		service := NewUserService(userRepo, wmRepo)

		user := &platformv1alpha1.User{
			ObjectMeta: metav1.ObjectMeta{Name: "test-user"},
			Spec: platformv1alpha1.UserSpec{
				Email:  "test@example.com",
				Active: &activeFalse,
			},
		}

		updatedUser, err := service.UpdateUser(ctx, user)
		assert.Error(t, err)
		assert.Nil(t, updatedUser)
		assert.Contains(t, err.Error(), "modified by another process")
	})
}

func TestDeleteUser(t *testing.T) {
	ctx := context.Background()

	t.Run("should delete user successfully", func(t *testing.T) {
		userRepo := &mockUserRepo{
			deleteFunc: func(ctx context.Context, name string) error {
				assert.Equal(t, "test-user", name)
				return nil
			},
		}

		wmRepo := &mockWorkMachineRepo{}
		service := NewUserService(userRepo, wmRepo)

		err := service.DeleteUser(ctx, "test-user")
		assert.NoError(t, err)
	})

	t.Run("should return error for non-existent user", func(t *testing.T) {
		userRepo := &mockUserRepo{
			deleteFunc: func(ctx context.Context, name string) error {
				return repository.ErrNotFound("user not found")
			},
		}

		wmRepo := &mockWorkMachineRepo{}
		service := NewUserService(userRepo, wmRepo)

		err := service.DeleteUser(ctx, "nonexistent")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "user not found")
	})
}

func TestListUsers(t *testing.T) {
	ctx := context.Background()
	activeTrue := true

	t.Run("should list all users", func(t *testing.T) {
		userRepo := &mockUserRepo{
			listFunc: func(ctx context.Context, opts ...repository.ListOption) (*platformv1alpha1.UserList, error) {
				return &platformv1alpha1.UserList{
					Items: []platformv1alpha1.User{
						{
							ObjectMeta: metav1.ObjectMeta{Name: "user1"},
							Spec:       platformv1alpha1.UserSpec{Email: "user1@example.com", Active: &activeTrue},
						},
						{
							ObjectMeta: metav1.ObjectMeta{Name: "user2"},
							Spec:       platformv1alpha1.UserSpec{Email: "user2@example.com", Active: &activeTrue},
						},
					},
				}, nil
			},
		}

		wmRepo := &mockWorkMachineRepo{}
		service := NewUserService(userRepo, wmRepo)

		userList, err := service.ListUsers(ctx)
		assert.NoError(t, err)
		assert.NotNil(t, userList)
		assert.Equal(t, 2, len(userList.Items))
	})

	t.Run("should handle list error", func(t *testing.T) {
		userRepo := &mockUserRepo{
			listFunc: func(ctx context.Context, opts ...repository.ListOption) (*platformv1alpha1.UserList, error) {
				return nil, errors.New("list failed")
			},
		}

		wmRepo := &mockWorkMachineRepo{}
		service := NewUserService(userRepo, wmRepo)

		userList, err := service.ListUsers(ctx)
		assert.Error(t, err)
		assert.Nil(t, userList)
		assert.Contains(t, err.Error(), "failed to list users")
	})
}

func TestResetUserPassword(t *testing.T) {
	ctx := context.Background()
	activeTrue := true

	t.Run("should reset password successfully", func(t *testing.T) {
		userRepo := &mockUserRepo{
			getFunc: func(ctx context.Context, name string) (*platformv1alpha1.User, error) {
				return &platformv1alpha1.User{
					ObjectMeta: metav1.ObjectMeta{Name: name},
					Spec: platformv1alpha1.UserSpec{
						Email:    "test@example.com",
						Password: "old-hashed-password",
						Active:   &activeTrue,
					},
				}, nil
			},
			updateFunc: func(ctx context.Context, user *platformv1alpha1.User) error {
				assert.Equal(t, "newpassword123", user.Spec.PasswordString)
				return nil
			},
		}

		wmRepo := &mockWorkMachineRepo{}
		service := NewUserService(userRepo, wmRepo)

		err := service.ResetUserPassword(ctx, "test-user", "newpassword123")
		assert.NoError(t, err)
	})

	t.Run("should return error for non-existent user", func(t *testing.T) {
		userRepo := &mockUserRepo{
			getFunc: func(ctx context.Context, name string) (*platformv1alpha1.User, error) {
				return nil, repository.ErrNotFound("user not found")
			},
		}

		wmRepo := &mockWorkMachineRepo{}
		service := NewUserService(userRepo, wmRepo)

		err := service.ResetUserPassword(ctx, "nonexistent", "newpassword123")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "user not found")
	})

	t.Run("should handle update error", func(t *testing.T) {
		userRepo := &mockUserRepo{
			getFunc: func(ctx context.Context, name string) (*platformv1alpha1.User, error) {
				return &platformv1alpha1.User{
					ObjectMeta: metav1.ObjectMeta{Name: name},
					Spec: platformv1alpha1.UserSpec{
						Email:  "test@example.com",
						Active: &activeTrue,
					},
				}, nil
			},
			updateFunc: func(ctx context.Context, user *platformv1alpha1.User) error {
				return errors.New("update failed")
			},
		}

		wmRepo := &mockWorkMachineRepo{}
		service := NewUserService(userRepo, wmRepo)

		err := service.ResetUserPassword(ctx, "test-user", "newpassword123")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to update user password")
	})
}

func TestUpdateUserLastLogin(t *testing.T) {
	ctx := context.Background()
	activeTrue := true

	t.Run("should update last login successfully", func(t *testing.T) {
		userRepo := &mockUserRepo{
			getFunc: func(ctx context.Context, name string) (*platformv1alpha1.User, error) {
				return &platformv1alpha1.User{
					ObjectMeta: metav1.ObjectMeta{Name: name},
					Spec: platformv1alpha1.UserSpec{
						Email:  "test@example.com",
						Active: &activeTrue,
					},
				}, nil
			},
			updateStatusFunc: func(ctx context.Context, user *platformv1alpha1.User) error {
				assert.NotNil(t, user.Status.LastLogin)
				return nil
			},
		}

		wmRepo := &mockWorkMachineRepo{}
		service := NewUserService(userRepo, wmRepo)

		err := service.UpdateUserLastLogin(ctx, "test-user")
		assert.NoError(t, err)
	})

	t.Run("should return error for non-existent user", func(t *testing.T) {
		userRepo := &mockUserRepo{
			getFunc: func(ctx context.Context, name string) (*platformv1alpha1.User, error) {
				return nil, repository.ErrNotFound("user not found")
			},
		}

		wmRepo := &mockWorkMachineRepo{}
		service := NewUserService(userRepo, wmRepo)

		err := service.UpdateUserLastLogin(ctx, "nonexistent")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "user not found")
	})

	t.Run("should handle update status error", func(t *testing.T) {
		userRepo := &mockUserRepo{
			getFunc: func(ctx context.Context, name string) (*platformv1alpha1.User, error) {
				return &platformv1alpha1.User{
					ObjectMeta: metav1.ObjectMeta{Name: name},
					Spec: platformv1alpha1.UserSpec{
						Email:  "test@example.com",
						Active: &activeTrue,
					},
				}, nil
			},
			updateStatusFunc: func(ctx context.Context, user *platformv1alpha1.User) error {
				return errors.New("status update failed")
			},
		}

		wmRepo := &mockWorkMachineRepo{}
		service := NewUserService(userRepo, wmRepo)

		err := service.UpdateUserLastLogin(ctx, "test-user")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to update user last login")
	})
}

func TestActivateUser(t *testing.T) {
	ctx := context.Background()
	activeTrue := true

	t.Run("should activate user successfully", func(t *testing.T) {
		userRepo := &mockUserRepo{
			patchFunc: func(ctx context.Context, name string, patchData map[string]interface{}) (*platformv1alpha1.User, error) {
				assert.Equal(t, "test-user", name)
				specData := patchData["spec"].(map[string]interface{})
				assert.True(t, specData["active"].(bool))

				return &platformv1alpha1.User{
					ObjectMeta: metav1.ObjectMeta{Name: name},
					Spec: platformv1alpha1.UserSpec{
						Email:  "test@example.com",
						Active: &activeTrue,
					},
				}, nil
			},
		}

		wmRepo := &mockWorkMachineRepo{}
		service := NewUserService(userRepo, wmRepo)

		user, err := service.ActivateUser(ctx, "test-user")
		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.True(t, *user.Spec.Active)
	})

	t.Run("should return error for non-existent user", func(t *testing.T) {
		userRepo := &mockUserRepo{
			patchFunc: func(ctx context.Context, name string, patchData map[string]interface{}) (*platformv1alpha1.User, error) {
				return nil, repository.ErrNotFound("user not found")
			},
		}

		wmRepo := &mockWorkMachineRepo{}
		service := NewUserService(userRepo, wmRepo)

		user, err := service.ActivateUser(ctx, "nonexistent")
		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Contains(t, err.Error(), "user not found")
	})

	t.Run("should handle patch conflict", func(t *testing.T) {
		userRepo := &mockUserRepo{
			patchFunc: func(ctx context.Context, name string, patchData map[string]interface{}) (*platformv1alpha1.User, error) {
				return nil, repository.ErrConflict("conflict")
			},
		}

		wmRepo := &mockWorkMachineRepo{}
		service := NewUserService(userRepo, wmRepo)

		user, err := service.ActivateUser(ctx, "test-user")
		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Contains(t, err.Error(), "modified by another process")
	})
}

func TestDeactivateUser(t *testing.T) {
	ctx := context.Background()
	activeFalse := false

	t.Run("should deactivate user successfully", func(t *testing.T) {
		userRepo := &mockUserRepo{
			patchFunc: func(ctx context.Context, name string, patchData map[string]interface{}) (*platformv1alpha1.User, error) {
				assert.Equal(t, "test-user", name)
				specData := patchData["spec"].(map[string]interface{})
				assert.False(t, specData["active"].(bool))

				return &platformv1alpha1.User{
					ObjectMeta: metav1.ObjectMeta{Name: name},
					Spec: platformv1alpha1.UserSpec{
						Email:  "test@example.com",
						Active: &activeFalse,
					},
				}, nil
			},
		}

		wmRepo := &mockWorkMachineRepo{}
		service := NewUserService(userRepo, wmRepo)

		user, err := service.DeactivateUser(ctx, "test-user")
		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.False(t, *user.Spec.Active)
	})

	t.Run("should return error for non-existent user", func(t *testing.T) {
		userRepo := &mockUserRepo{
			patchFunc: func(ctx context.Context, name string, patchData map[string]interface{}) (*platformv1alpha1.User, error) {
				return nil, repository.ErrNotFound("user not found")
			},
		}

		wmRepo := &mockWorkMachineRepo{}
		service := NewUserService(userRepo, wmRepo)

		user, err := service.DeactivateUser(ctx, "nonexistent")
		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Contains(t, err.Error(), "user not found")
	})

	t.Run("should handle patch conflict", func(t *testing.T) {
		userRepo := &mockUserRepo{
			patchFunc: func(ctx context.Context, name string, patchData map[string]interface{}) (*platformv1alpha1.User, error) {
				return nil, repository.ErrConflict("conflict")
			},
		}

		wmRepo := &mockWorkMachineRepo{}
		service := NewUserService(userRepo, wmRepo)

		user, err := service.DeactivateUser(ctx, "test-user")
		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Contains(t, err.Error(), "modified by another process")
	})
}

func TestSanitizeForLabel(t *testing.T) {
	t.Run("should sanitize email with @ and .", func(t *testing.T) {
		result := sanitizeForLabel("test.user@example.com")
		assert.Equal(t, "test-dot-user-at-example-dot-com", result)
	})

	t.Run("should handle underscores", func(t *testing.T) {
		result := sanitizeForLabel("test_user@example.com")
		assert.Equal(t, "test-user-at-example-dot-com", result)
	})

	t.Run("should convert to lowercase", func(t *testing.T) {
		result := sanitizeForLabel("TestUser@Example.COM")
		assert.Equal(t, "testuser-at-example-dot-com", result)
	})

	t.Run("should trim leading and trailing hyphens", func(t *testing.T) {
		result := sanitizeForLabel("-test@example.com-")
		assert.Equal(t, "test-at-example-dot-com", result)
	})

	t.Run("should limit length to 63 characters", func(t *testing.T) {
		longEmail := "very.long.email.address.that.exceeds.sixtythree.characters@example.com"
		result := sanitizeForLabel(longEmail)
		assert.LessOrEqual(t, len(result), 63)
	})
}
