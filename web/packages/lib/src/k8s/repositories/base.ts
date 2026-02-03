import type { V1DeleteOptions, V1Status } from "@kubernetes/client-node";
import { getK8sClient } from "../client";
import { parseK8sError, NotFoundError } from "../errors";
import type { K8sResource, K8sList } from "../types/common";

/**
 * Options for list operations
 */
export interface ListOptions {
  labelSelector?: string;
  fieldSelector?: string;
  limit?: number;
  continue?: string;
  timeoutSeconds?: number;
}

/**
 * Options for delete operations
 */
export interface DeleteOptions {
  gracePeriodSeconds?: number;
  propagationPolicy?: "Orphan" | "Background" | "Foreground";
  dryRun?: string[];
}

/**
 * Options for patch operations
 */
export interface PatchOptions {
  dryRun?: string[];
  fieldManager?: string;
  force?: boolean;
}

/**
 * Patch types
 */
export type PatchType =
  | "application/json-patch+json"
  | "application/merge-patch+json"
  | "application/strategic-merge-patch+json";

/**
 * Base repository for Kubernetes resources
 * Provides generic CRUD operations for both namespaced and cluster-scoped resources
 */
export abstract class BaseRepository<T extends K8sResource> {
  protected client = getK8sClient();

  constructor(
    protected readonly group: string,
    protected readonly version: string,
    protected readonly plural: string,
    protected readonly namespaced: boolean = true,
  ) {}

  /**
   * Get a single resource by name
   */
  async get(namespace: string, name: string): Promise<T>;
  async get(name: string): Promise<T>;
  async get(namespaceOrName: string, name?: string): Promise<T> {
    try {
      if (this.namespaced && name) {
        // Namespaced resource
        const response = await this.client.custom.getNamespacedCustomObject({
          group: this.group,
          version: this.version,
          namespace: namespaceOrName,
          plural: this.plural,
          name,
        });
        return response as unknown as T;
      } else {
        // Cluster-scoped resource
        const response = await this.client.custom.getClusterCustomObject({
          group: this.group,
          version: this.version,
          plural: this.plural,
          name: namespaceOrName,
        });
        return response as unknown as T;
      }
    } catch (err) {
      throw parseK8sError(err);
    }
  }

  /**
   * List resources
   */
  async list(namespace: string, options?: ListOptions): Promise<K8sList<T>>;
  async list(options?: ListOptions): Promise<K8sList<T>>;
  async list(
    namespaceOrOptions?: string | ListOptions,
    options?: ListOptions,
  ): Promise<K8sList<T>> {
    try {
      let namespace: string | undefined;
      let opts: ListOptions | undefined;

      if (typeof namespaceOrOptions === "string") {
        namespace = namespaceOrOptions;
        opts = options;
      } else {
        opts = namespaceOrOptions;
      }

      if (this.namespaced && namespace) {
        // Namespaced resource
        const response = await this.client.custom.listNamespacedCustomObject({
          group: this.group,
          version: this.version,
          namespace,
          plural: this.plural,
          labelSelector: opts?.labelSelector,
          fieldSelector: opts?.fieldSelector,
          limit: opts?.limit,
          _continue: opts?.continue,
          timeoutSeconds: opts?.timeoutSeconds,
        });
        return response as unknown as K8sList<T>;
      } else {
        // Cluster-scoped resource
        // Use object parameters (new API style)
        const response = await this.client.custom.listClusterCustomObject({
          group: this.group,
          version: this.version,
          plural: this.plural,
          labelSelector: opts?.labelSelector,
          fieldSelector: opts?.fieldSelector,
          limit: opts?.limit,
          _continue: opts?.continue,
          timeoutSeconds: opts?.timeoutSeconds,
        });

        return response as unknown as K8sList<T>;
      }
    } catch (err) {
      throw parseK8sError(err);
    }
  }

  /**
   * Create a new resource
   */
  async create(namespace: string, resource: Partial<T>): Promise<T>;
  async create(resource: Partial<T>): Promise<T>;
  async create(
    namespaceOrResource: string | Partial<T>,
    resource?: Partial<T>,
  ): Promise<T> {
    try {
      if (
        this.namespaced &&
        typeof namespaceOrResource === "string" &&
        resource
      ) {
        // Namespaced resource - use object parameters API
        const response = await this.client.custom.createNamespacedCustomObject({
          group: this.group,
          version: this.version,
          namespace: namespaceOrResource,
          plural: this.plural,
          body: resource as object,
        });
        return response as unknown as T;
      } else if (!this.namespaced && typeof namespaceOrResource === "object") {
        // Cluster-scoped resource
        const response = await this.client.custom.createClusterCustomObject({
          group: this.group,
          version: this.version,
          plural: this.plural,
          body: namespaceOrResource as object,
        });
        return response as unknown as T;
      } else {
        throw new Error("Invalid arguments for create operation");
      }
    } catch (err) {
      throw parseK8sError(err);
    }
  }

  /**
   * Update an existing resource (replaces entire resource)
   */
  async update(
    namespace: string,
    name: string,
    resource: Partial<T>,
  ): Promise<T>;
  async update(name: string, resource: Partial<T>): Promise<T>;
  async update(
    namespaceOrName: string,
    nameOrResource: string | Partial<T>,
    resource?: Partial<T>,
  ): Promise<T> {
    try {
      if (this.namespaced && typeof nameOrResource === "string" && resource) {
        // Namespaced resource - use object parameters API
        const response = await this.client.custom.replaceNamespacedCustomObject(
          {
            group: this.group,
            version: this.version,
            namespace: namespaceOrName,
            plural: this.plural,
            name: nameOrResource,
            body: resource as object,
          },
        );
        return response as unknown as T;
      } else if (!this.namespaced && typeof nameOrResource === "object") {
        // Cluster-scoped resource
        const response = await this.client.custom.replaceClusterCustomObject({
          group: this.group,
          version: this.version,
          plural: this.plural,
          name: namespaceOrName,
          body: nameOrResource as object,
        });
        return response as unknown as T;
      } else {
        throw new Error("Invalid arguments for update operation");
      }
    } catch (err) {
      throw parseK8sError(err);
    }
  }

  /**
   * Patch an existing resource (partial update)
   */
  async patch(
    namespace: string,
    name: string,
    patch: object,
    patchType?: PatchType,
    options?: PatchOptions,
  ): Promise<T>;
  async patch(
    name: string,
    patch: object,
    patchType?: PatchType,
    options?: PatchOptions,
  ): Promise<T>;
  async patch(
    namespaceOrName: string,
    nameOrPatch: string | object,
    patchOrType?: object | PatchType,
    _patchTypeOrOptions?: PatchType | PatchOptions,
    _options?: PatchOptions,
  ): Promise<T> {
    try {
      if (this.namespaced && typeof nameOrPatch === "string") {
        // Namespaced resource - use object parameters API
        const patch = patchOrType as object;

        const response = await this.client.custom.patchNamespacedCustomObject({
          group: this.group,
          version: this.version,
          namespace: namespaceOrName,
          plural: this.plural,
          name: nameOrPatch,
          body: patch,
        });
        return response as unknown as T;
      } else if (!this.namespaced && typeof nameOrPatch === "object") {
        // Cluster-scoped resource
        const patch = nameOrPatch;

        const response = await this.client.custom.patchClusterCustomObject({
          group: this.group,
          version: this.version,
          plural: this.plural,
          name: namespaceOrName,
          body: patch,
        });
        return response as unknown as T;
      } else {
        throw new Error("Invalid arguments for patch operation");
      }
    } catch (err) {
      throw parseK8sError(err);
    }
  }

  /**
   * Update resource status subresource
   */
  async updateStatus(
    namespace: string,
    name: string,
    resource: Partial<T>,
  ): Promise<T>;
  async updateStatus(name: string, resource: Partial<T>): Promise<T>;
  async updateStatus(
    namespaceOrName: string,
    nameOrResource: string | Partial<T>,
    resource?: Partial<T>,
  ): Promise<T> {
    try {
      if (this.namespaced && typeof nameOrResource === "string" && resource) {
        // Namespaced resource - use object parameters API
        const response =
          await this.client.custom.patchNamespacedCustomObjectStatus({
            group: this.group,
            version: this.version,
            namespace: namespaceOrName,
            plural: this.plural,
            name: nameOrResource,
            body: resource as object,
          });
        return response as unknown as T;
      } else if (!this.namespaced && typeof nameOrResource === "object") {
        // Cluster-scoped resource
        const response =
          await this.client.custom.patchClusterCustomObjectStatus({
            group: this.group,
            version: this.version,
            plural: this.plural,
            name: namespaceOrName,
            body: nameOrResource as object,
          });
        return response as unknown as T;
      } else {
        throw new Error("Invalid arguments for updateStatus operation");
      }
    } catch (err) {
      throw parseK8sError(err);
    }
  }

  /**
   * Delete a resource
   */
  async delete(
    namespace: string,
    name: string,
    options?: DeleteOptions,
  ): Promise<V1Status>;
  async delete(name: string, options?: DeleteOptions): Promise<V1Status>;
  async delete(
    namespaceOrName: string,
    nameOrOptions?: string | DeleteOptions,
    options?: DeleteOptions,
  ): Promise<V1Status> {
    try {
      if (this.namespaced && typeof nameOrOptions === "string") {
        // Namespaced resource - use object parameters API
        const deleteOptions: V1DeleteOptions = {
          gracePeriodSeconds: options?.gracePeriodSeconds,
          propagationPolicy: options?.propagationPolicy,
          dryRun: options?.dryRun,
        };

        const response = await this.client.custom.deleteNamespacedCustomObject({
          group: this.group,
          version: this.version,
          namespace: namespaceOrName,
          plural: this.plural,
          name: nameOrOptions,
          body: deleteOptions,
        });
        return response.body as V1Status;
      } else {
        // Cluster-scoped resource
        const opts = (nameOrOptions as DeleteOptions) || {};
        const deleteOptions: V1DeleteOptions = {
          gracePeriodSeconds: opts.gracePeriodSeconds,
          propagationPolicy: opts.propagationPolicy,
          dryRun: opts.dryRun,
        };

        const response = await this.client.custom.deleteClusterCustomObject({
          group: this.group,
          version: this.version,
          plural: this.plural,
          name: namespaceOrName,
          body: deleteOptions,
        });
        return response.body as V1Status;
      }
    } catch (err) {
      throw parseK8sError(err);
    }
  }

  /**
   * Check if a resource exists
   */
  async exists(namespace: string, name: string): Promise<boolean>;
  async exists(name: string): Promise<boolean>;
  async exists(namespaceOrName: string, name?: string): Promise<boolean> {
    try {
      if (this.namespaced && name) {
        await this.get(namespaceOrName, name);
      } else {
        await this.get(namespaceOrName);
      }
      return true;
    } catch (err) {
      if (err instanceof NotFoundError) {
        return false;
      }
      throw err;
    }
  }

  /**
   * Create or update a resource (upsert)
   */
  async createOrUpdate(
    namespace: string,
    name: string,
    resource: Partial<T>,
  ): Promise<T>;
  async createOrUpdate(name: string, resource: Partial<T>): Promise<T>;
  async createOrUpdate(
    namespaceOrName: string,
    nameOrResource: string | Partial<T>,
    resource?: Partial<T>,
  ): Promise<T> {
    try {
      if (this.namespaced && typeof nameOrResource === "string" && resource) {
        // Namespaced resource
        const exists = await this.exists(namespaceOrName, nameOrResource);
        if (exists) {
          return await this.update(namespaceOrName, nameOrResource, resource);
        } else {
          return await this.create(namespaceOrName, resource);
        }
      } else if (!this.namespaced && typeof nameOrResource === "object") {
        // Cluster-scoped resource
        const exists = await this.exists(namespaceOrName);
        if (exists) {
          return await this.update(namespaceOrName, nameOrResource);
        } else {
          return await this.create(nameOrResource);
        }
      } else {
        throw new Error("Invalid arguments for createOrUpdate operation");
      }
    } catch (err) {
      throw parseK8sError(err);
    }
  }
}
