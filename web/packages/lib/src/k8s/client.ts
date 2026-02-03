import * as k8s from '@kubernetes/client-node';
import {
  isInCluster,
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
    // Always use kube-proxy sidecar (bun can't handle mTLS)
    const apiUrl = getK8sApiUrl();

    const cluster: k8s.Cluster = {
      name: 'kube-proxy',
      server: apiUrl,
      skipTLSVerify: true, // kube-proxy doesn't use TLS
    };

    const user: k8s.User = {
      name: 'kube-proxy-user',
    };

    const context: k8s.Context = {
      name: 'kube-proxy',
      cluster: 'kube-proxy',
      user: 'kube-proxy-user',
    };

    this.kc.loadFromOptions({
      clusters: [cluster],
      users: [user],
      contexts: [context],
      currentContext: 'kube-proxy',
    });

    console.log(`✓ Kubernetes client configured to use kube-proxy sidecar at ${apiUrl}`);
  }

  private configureOutOfCluster() {
    try {
      // Use kubectl proxy for local development
      const proxyUrl = process.env.KUBECTL_PROXY_URL || 'http://localhost:8080';

      // Create cluster configuration for kubectl proxy
      const cluster: k8s.Cluster = {
        name: 'kubectl-proxy',
        server: proxyUrl,
        skipTLSVerify: true, // kubectl proxy doesn't use TLS
      };

      // Create user configuration (no auth needed for kubectl proxy)
      const user: k8s.User = {
        name: 'kubectl-proxy-user',
      };

      // Create context
      const context: k8s.Context = {
        name: 'kubectl-proxy',
        cluster: 'kubectl-proxy',
        user: 'kubectl-proxy-user',
      };

      // Apply configuration
      this.kc.loadFromOptions({
        clusters: [cluster],
        users: [user],
        contexts: [context],
        currentContext: 'kubectl-proxy',
      });

      console.log(`✓ Kubernetes client configured to use kubectl proxy at ${proxyUrl}`);
    } catch (err) {
      console.error('Failed to configure Kubernetes client:', err);
      throw new Error('Failed to configure Kubernetes client. Ensure kubectl proxy is running on port 8080.');
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
