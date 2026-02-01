import { readFileSync } from 'fs';
import { join } from 'path';

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
 */
export function getK8sApiUrl(): string {
  // In-cluster default
  if (isInCluster()) {
    return process.env.KUBERNETES_SERVICE_HOST && process.env.KUBERNETES_SERVICE_PORT
      ? `https://${process.env.KUBERNETES_SERVICE_HOST}:${process.env.KUBERNETES_SERVICE_PORT}`
      : 'https://kubernetes.default.svc';
  }

  // Out-of-cluster (development)
  return process.env.K8S_API_URL || 'https://localhost:6443';
}
