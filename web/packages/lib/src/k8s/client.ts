import * as k8s from '@kubernetes/client-node';
import {
  isInCluster,
  loadServiceAccountToken,
  loadServiceAccountCACert,
  getK8sApiUrl,
} from './auth';

/**
 * Kubernetes client singleton
 */
class K8sClient {
  private static instance: K8sClient | null = null;
  private kc: k8s.KubeConfig;
  public core: k8s.CoreV1Api;
  public apps: k8s.AppsV1Api;
  public custom: k8s.CustomObjectsApi;
  public metrics: k8s.Metrics;

  private constructor() {
    this.kc = new k8s.KubeConfig();

    if (isInCluster()) {
      // In-cluster configuration using ServiceAccount
      this.configureInCluster();
    } else {
      // Out-of-cluster configuration using kubeconfig
      this.configureOutOfCluster();
    }

    // Initialize API clients
    this.core = this.kc.makeApiClient(k8s.CoreV1Api);
    this.apps = this.kc.makeApiClient(k8s.AppsV1Api);
    this.custom = this.kc.makeApiClient(k8s.CustomObjectsApi);
    this.metrics = new k8s.Metrics(this.kc);
  }

  private configureInCluster() {
    const token = loadServiceAccountToken();
    const caCert = loadServiceAccountCACert();
    const apiUrl = getK8sApiUrl();

    if (!token) {
      throw new Error('ServiceAccount token not found. Ensure the pod has a ServiceAccount mounted.');
    }

    // Create cluster configuration
    const cluster: k8s.Cluster = {
      name: 'in-cluster',
      server: apiUrl,
      skipTLSVerify: false,
      caData: caCert ? Buffer.from(caCert).toString('base64') : undefined,
    };

    // Create user configuration with token auth
    const user: k8s.User = {
      name: 'serviceaccount',
      token,
    };

    // Create context
    const context: k8s.Context = {
      name: 'in-cluster',
      cluster: 'in-cluster',
      user: 'serviceaccount',
    };

    // Apply configuration
    this.kc.loadFromOptions({
      clusters: [cluster],
      users: [user],
      contexts: [context],
      currentContext: 'in-cluster',
    });

    console.log('✓ Kubernetes client configured for in-cluster access');
  }

  private configureOutOfCluster() {
    try {
      // Try to load from default kubeconfig location
      const kubeconfigPath = process.env.KUBECONFIG ||
        (process.env.HOME ? `${process.env.HOME}/.kube/config` : undefined);

      if (kubeconfigPath) {
        this.kc.loadFromFile(kubeconfigPath);
      } else {
        // Fallback to default config
        this.kc.loadFromDefault();
        console.log('✓ Kubernetes client configured from default kubeconfig');
      }
    } catch (err) {
      console.error('Failed to configure Kubernetes client:', err);
      throw new Error('Failed to configure Kubernetes client. Ensure KUBECONFIG is set or running in-cluster.');
    }
  }

  /**
   * Get singleton instance
   */
  public static getInstance(): K8sClient {
    if (!K8sClient.instance) {
      K8sClient.instance = new K8sClient();
    }
    return K8sClient.instance;
  }

  /**
   * Get KubeConfig instance
   */
  public getKubeConfig(): k8s.KubeConfig {
    return this.kc;
  }

  /**
   * Create a watch stream for resources
   */
  public createWatch() {
    return new k8s.Watch(this.kc);
  }

  /**
   * Create a log stream for pod logs
   */
  public createLog() {
    return new k8s.Log(this.kc);
  }
}

/**
 * Get the singleton Kubernetes client instance
 */
export function getK8sClient(): K8sClient {
  return K8sClient.getInstance();
}

/**
 * Export k8s types for convenience
 */
export * from '@kubernetes/client-node';
