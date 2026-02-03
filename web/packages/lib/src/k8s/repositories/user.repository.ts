import { BaseRepository, type ListOptions } from "./base";
import type { User, UserList, RoleType, ProviderAccount } from "../types/user";
import { buildLabelSelector } from "../utils";
import { parseK8sError, NotFoundError } from "../errors";

/**
 * Utility function to sanitize email for use in Kubernetes labels
 * Converts: user@kloudlite.io -> user-at-kloudlite-dot-io
 */
function sanitizeEmailForLabel(email: string): string {
  return email.replace(/@/g, "-at-").replace(/\./g, "-dot-").toLowerCase();
}

/**
 * User repository for managing User custom resources
 * Users are cluster-scoped (not namespaced)
 */
export class UserRepository extends BaseRepository<User> {
  constructor() {
    super("platform.kloudlite.io", "v1alpha1", "users", false); // false = cluster-scoped
  }

  /**
   * Get user by email address
   */
  async getByEmail(email: string): Promise<User> {
    try {
      // Convert email to sanitized label format
      const emailLabel = sanitizeEmailForLabel(email);
      const result = await this.list({
        labelSelector: buildLabelSelector({
          "kloudlite.io/user-email": emailLabel,
        }),
      });

      if (result.items.length === 0) {
        throw new NotFoundError("User", `with email ${email}`);
      }

      if (result.items.length > 1) {
        throw new Error(`Multiple users found with email ${email}`);
      }

      return result.items[0];
    } catch (err) {
      throw parseK8sError(err);
    }
  }

  /**
   * Get user by username (metadata.name)
   */
  async getByUsername(username: string): Promise<User> {
    try {
      return await this.get(username);
    } catch (err) {
      throw parseK8sError(err);
    }
  }

  /**
   * List all active users
   */
  async listActive(options?: ListOptions): Promise<UserList> {
    try {
      const all = await this.list(options);

      // Filter for active users (spec.active = true or undefined defaults to true)
      const filtered = all.items.filter((user) => user.spec?.active !== false);

      return {
        ...all,
        items: filtered,
      } as UserList;
    } catch (err) {
      throw parseK8sError(err);
    }
  }

  /**
   * List users by role
   */
  async listByRole(role: RoleType, options?: ListOptions): Promise<UserList> {
    try {
      const all = await this.list(options);

      // Filter by role (client-side since roles is an array in spec)
      const filtered = all.items.filter((user) =>
        user.spec?.roles?.includes(role),
      );

      return {
        ...all,
        items: filtered,
      } as UserList;
    } catch (err) {
      throw parseK8sError(err);
    }
  }

  /**
   * List users by provider (google, github, etc.)
   */
  async listByProvider(
    provider: string,
    options?: ListOptions,
  ): Promise<UserList> {
    try {
      const all = await this.list(options);

      // Filter by provider (client-side since providers is an array in spec)
      const filtered = all.items.filter((user) =>
        user.spec?.providers?.some((p) => p.provider === provider),
      );

      return {
        ...all,
        items: filtered,
      } as UserList;
    } catch (err) {
      throw parseK8sError(err);
    }
  }

  /**
   * Update user status (for controller use)
   */
  async updateUserStatus(
    name: string,
    status: Partial<User["status"]>,
  ): Promise<User> {
    try {
      const user = await this.get(name);

      user.status = {
        ...user.status,
        ...status,
      };

      return await this.updateStatus(name, user);
    } catch (err) {
      throw parseK8sError(err);
    }
  }

  /**
   * Activate a user
   */
  async activate(name: string): Promise<User> {
    try {
      const user = await this.get(name);

      user.spec = {
        ...user.spec!,
        active: true,
      };

      return await this.update(name, user);
    } catch (err) {
      throw parseK8sError(err);
    }
  }

  /**
   * Deactivate a user
   */
  async deactivate(name: string): Promise<User> {
    try {
      const user = await this.get(name);

      user.spec = {
        ...user.spec!,
        active: false,
      };

      return await this.update(name, user);
    } catch (err) {
      throw parseK8sError(err);
    }
  }

  /**
   * Add role to user
   */
  async addRole(name: string, role: User["spec"]["roles"][0]): Promise<User> {
    try {
      const user = await this.get(name);

      if (!user.spec!.roles.includes(role)) {
        user.spec!.roles.push(role);
      }

      return await this.update(name, user);
    } catch (err) {
      throw parseK8sError(err);
    }
  }

  /**
   * Remove role from user
   */
  async removeRole(
    name: string,
    role: User["spec"]["roles"][0],
  ): Promise<User> {
    try {
      const user = await this.get(name);

      user.spec!.roles = user.spec!.roles.filter((r) => r !== role);

      // Ensure at least one role remains
      if (user.spec!.roles.length === 0) {
        user.spec!.roles = ["user"];
      }

      return await this.update(name, user);
    } catch (err) {
      throw parseK8sError(err);
    }
  }

  /**
   * Update user display name
   */
  async updateDisplayName(name: string, displayName: string): Promise<User> {
    try {
      const user = await this.get(name);

      user.spec = {
        ...user.spec!,
        displayName,
      };

      return await this.update(name, user);
    } catch (err) {
      throw parseK8sError(err);
    }
  }

  /**
   * Update user avatar URL
   */
  async updateAvatarUrl(name: string, avatarUrl: string): Promise<User> {
    try {
      const user = await this.get(name);

      user.spec = {
        ...user.spec!,
        avatarUrl,
      };

      return await this.update(name, user);
    } catch (err) {
      throw parseK8sError(err);
    }
  }

  /**
   * Update last login timestamp
   */
  async updateLastLogin(name: string): Promise<User> {
    try {
      const user = await this.get(name);

      user.status = {
        ...user.status,
        lastLogin: new Date().toISOString(),
      };

      return await this.updateStatus(name, user);
    } catch (err) {
      throw parseK8sError(err);
    }
  }

  /**
   * Add OAuth provider account
   */
  async addProvider(name: string, provider: ProviderAccount): Promise<User> {
    try {
      const user = await this.get(name);

      if (!user.spec!.providers) {
        user.spec!.providers = [];
      }

      // Check if provider already exists
      const exists = user.spec!.providers.some(
        (p) =>
          p.provider === provider.provider &&
          p.providerId === provider.providerId,
      );

      if (!exists) {
        user.spec!.providers.push(provider);
      }

      return await this.update(name, user);
    } catch (err) {
      throw parseK8sError(err);
    }
  }

  /**
   * Remove OAuth provider account
   */
  async removeProvider(
    name: string,
    provider: string,
    providerId: string,
  ): Promise<User> {
    try {
      const user = await this.get(name);

      if (user.spec!.providers) {
        user.spec!.providers = user.spec!.providers.filter(
          (p) => !(p.provider === provider && p.providerId === providerId),
        );
      }

      return await this.update(name, user);
    } catch (err) {
      throw parseK8sError(err);
    }
  }

  /**
   * Check if user has a specific role
   */
  async hasRole(
    name: string,
    role: User["spec"]["roles"][0],
  ): Promise<boolean> {
    try {
      const user = await this.get(name);
      return user.spec?.roles?.includes(role) ?? false;
    } catch (err) {
      throw parseK8sError(err);
    }
  }

  /**
   * Check if user is active
   */
  async isActive(name: string): Promise<boolean> {
    try {
      const user = await this.get(name);
      return user.spec?.active !== false;
    } catch (err) {
      throw parseK8sError(err);
    }
  }
}

// Export singleton instance
export const userRepository = new UserRepository();
