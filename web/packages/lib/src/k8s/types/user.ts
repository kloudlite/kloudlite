/**
 * User and UserPreferences CRD type definitions
 * Based on: api/internal/controllers/user/v1alpha1/user_types.go
 */

import type { K8sResource, K8sList, Condition } from './common';

export const UserGroup = 'platform.kloudlite.io';
export const UserVersion = 'v1alpha1';
export const UserPlural = 'users';

// User types
export type RoleType = 'super-admin' | 'admin' | 'user';

export interface ProviderAccount {
  provider: string;
  providerId: string;
  email: string;
  name?: string;
  image?: string;
  connectedAt: string;
}

export interface UserSpec {
  email: string;
  displayName?: string;
  avatarUrl?: string;
  providers?: ProviderAccount[];
  roles: RoleType[];
  active?: boolean;
  password?: string;
  passwordString?: string;
  metadata?: Record<string, string>;
}

export type UserPhase = 'active' | 'inactive' | 'suspended' | 'pending';

export interface UserStatus {
  phase?: UserPhase;
  lastLogin?: string;
  createdAt?: string;
  passwordHash?: string;
  conditions?: Condition[];
}

// Main User resource (cluster-scoped)
export interface User extends K8sResource<UserSpec, UserStatus> {
  apiVersion: 'platform.kloudlite.io/v1alpha1';
  kind: 'User';
}

export interface UserList extends K8sList<User> {
  apiVersion: 'platform.kloudlite.io/v1alpha1';
  kind: 'UserList';
}

// Helper types
export type UserCreateInput = Omit<User, 'apiVersion' | 'kind' | 'status'> & {
  apiVersion?: string;
  kind?: string;
};

export type UserUpdateInput = Partial<UserSpec>;

// ============================================================================
// UserPreferences - User pinned resources and settings
// ============================================================================

export const UserPreferencesGroup = 'platform.kloudlite.io';
export const UserPreferencesVersion = 'v1alpha1';
export const UserPreferencesPlural = 'userpreferences';

export interface ResourceReference {
  name: string;
  namespace?: string;
}

export interface UserPreferencesSpec {
  pinnedWorkspaces?: ResourceReference[];
  pinnedEnvironments?: string[];
}

export interface UserPreferencesStatus {
  lastUpdated?: string;
}

// Main UserPreferences resource (cluster-scoped)
// The resource name should match the username (User.metadata.name)
export interface UserPreferences extends K8sResource<UserPreferencesSpec, UserPreferencesStatus> {
  apiVersion: 'platform.kloudlite.io/v1alpha1';
  kind: 'UserPreferences';
}

export interface UserPreferencesList extends K8sList<UserPreferences> {
  apiVersion: 'platform.kloudlite.io/v1alpha1';
  kind: 'UserPreferencesList';
}

// Helper types
export type UserPreferencesCreateInput = Omit<UserPreferences, 'apiVersion' | 'kind' | 'status'> & {
  apiVersion?: string;
  kind?: string;
};

export type UserPreferencesUpdateInput = Partial<UserPreferencesSpec>;
