import { readFileSync } from 'fs';

/**
 * ServiceAccount token paths in Kubernetes pods
 */
const SERVICE_ACCOUNT_TOKEN_PATH = '/var/run/secrets/kubernetes.io/serviceaccount/token';
const SERVICE_ACCOUNT_CA_CERT_PATH = '/var/run/secrets/kubernetes.io/serviceaccount/ca.crt';
const SERVICE_ACCOUNT_NAMESPACE_PATH = '/var/run/secrets/kubernetes.io/serviceaccount/namespace';

/**
 * Load ServiceAccount token from mounted volume
 */
export function loadServiceAccountToken(): string | null {
  try {
    return readFileSync(SERVICE_ACCOUNT_TOKEN_PATH, 'utf8').trim();
  } catch (err) {
    console.warn('Failed to load ServiceAccount token:', err);
    return null;
  }
}

/**
 * Load ServiceAccount CA certificate
 */
export function loadServiceAccountCACert(): string | null {
  try {
    return readFileSync(SERVICE_ACCOUNT_CA_CERT_PATH, 'utf8');
  } catch (err) {
    console.warn('Failed to load ServiceAccount CA cert:', err);
    return null;
  }
}

/**
 * Load ServiceAccount namespace
 */
export function loadServiceAccountNamespace(): string | null {
  try {
    return readFileSync(SERVICE_ACCOUNT_NAMESPACE_PATH, 'utf8').trim();
  } catch (err) {
    console.warn('Failed to load ServiceAccount namespace:', err);
    return null;
  }
}

/**
 * Check if running in a Kubernetes cluster
 */
export function isInCluster(): boolean {
  try {
    readFileSync(SERVICE_ACCOUNT_TOKEN_PATH);
    return true;
  } catch {
    return false;
  }
}

/**
 * Get Kubernetes API server URL from environment or default
 * Always uses kube-proxy sidecar when in-cluster (bun can't handle mTLS)
 */
export function getK8sApiUrl(): string {
  // In-cluster: always use kube-proxy sidecar at localhost:8001
  if (isInCluster()) {
    const host = process.env.KUBERNETES_SERVICE_HOST || '127.0.0.1';
    const port = process.env.KUBERNETES_SERVICE_PORT || '8001';
    return `http://${host}:${port}`;
  }

  // Out-of-cluster (development): use kubectl proxy
  return process.env.KUBECTL_PROXY_URL || 'http://localhost:8080';
}
