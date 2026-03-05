# Controller Configuration

This document describes the configurable values for Kloudlite controllers. All values can be set via environment variables to customize behavior for different deployment environments.

## Configuration Loading

Configuration is loaded from environment variables using the `github.com/codingconcepts/env` package. The configuration structure is defined in `/api/internal/controllers/config.go`.

## Environment Variables

### Workspace Controller (`WORKSPACE_` prefix)

| Variable | Default | Description |
|-----------|----------|-------------|
| `WORKSPACE_DEFAULT_IDLE_TIMEOUT_MINUTES` | `30` | Default idle timeout before auto-stopping a workspace (in minutes) |
| `WORKSPACE_REQUEUE_INTERVAL_MINUTES` | `1` | How often to requeue workspaces for idle checking (in minutes) |
| `WORKSPACE_RBAC_CLEANUP_INTERVAL_MINUTES` | `60` | How often to run orphaned RBAC cleanup (in minutes) |
| `WORKSPACE_KUBECTL_IMAGE` | `bitnami/kubectl:latest` | Image used for kubectl operations |
| `WORKSPACE_GIT_IMAGE` | `alpine/git:latest` | Image used for git operations |
| `WORKSPACE_ALPINE_IMAGE` | `alpine:latest` | Image used for Alpine-based operations |
| `WORKSPACE_CLEANUP_POD_TTL_SECONDS` | `300` | How long cleanup pods are kept (in seconds) |
| `WORKSPACE_VSCODE_VERSION` | `latest` | Default VS Code version for workspaces |

### Environment Controller (`ENVIRONMENT_` prefix)

| Variable | Default | Description |
|-----------|----------|-------------|
| `ENVIRONMENT_POD_TERMINATION_RETRY_INTERVAL` | `2s` | How long to wait between pod termination checks |
| `ENVIRONMENT_SNAPSHOT_RESTORE_RETRY_INTERVAL` | `2s` | How long to wait between snapshot restore retries |
| `ENVIRONMENT_SNAPSHOT_REQUEST_RETRY_INTERVAL` | `2s` | How long to wait between snapshot request retries |
| `ENVIRONMENT_FORK_RETRY_INTERVAL` | `5s` | How long to wait between fork operation retries |
| `ENVIRONMENT_STATUS_UPDATE_RETRY_INTERVAL` | `5s` | How long to wait between status update retries |
| `ENVIRONMENT_DELETION_RETRY_INTERVAL` | `5s` | How long to wait between deletion retries |
| `ENVIRONMENT_LIFECYCLE_RETRY_INTERVAL` | `5s` | How long to wait between lifecycle operation retries |

**Note:** Duration values should use Go duration format (e.g., `2s`, `5s`, `10s`)

### WorkMachine Controller (`WORKMACHINE_` prefix)

| Variable | Default | Description |
|-----------|----------|-------------|
| `WORKMACHINE_WM_INGRESS_CONTROLLER_IMAGE` | `ghcr.io/kloudlite/kloudlite/wm-ingress-controller:development` | Image for the wm-ingress-controller |
| `WORKMACHINE_SSH_USERNAME` | `kloudlite` | Username for SSH access to workmachine nodes |
| `WORKMACHINE_DEFAULT_WILDCARD_CERT_NAME` | `kloudlite-wildcard-cert-tls` | Default wildcard TLS certificate secret name |
| `WORKMACHINE_CLOUD_OPERATION_RETRY_INTERVAL` | `5s` | How long to wait between cloud operation retries |
| `WORKMACHINE_MACHINE_STATUS_CHECK_INTERVAL` | `5s` | How long to wait between machine status checks |
| `WORKMACHINE_MACHINE_STARTUP_RETRY_INTERVAL` | `10s` | How long to wait between machine startup retries |
| `WORKMACHINE_NODE_JOIN_RETRY_INTERVAL` | `10s` | How long to wait between node join checks |
| `WORKMACHINE_VOLUME_RESIZE_RETRY_INTERVAL` | `10s` | How long to wait between volume resize checks |
| `WORKMACHINE_MACHINE_TYPE_CHANGE_RETRY_INTERVAL` | `5s` | How long to wait between machine type change retries |
| `WORKMACHINE_AUTO_SHUTDOWN_CHECK_INTERVAL` | `5m` | How often to check for auto-shutdown |
| `WORKMACHINE_AUTO_SHUTDOWN_IDLE_THRESHOLD_MINUTES` | `30` | How long a workmachine can be idle before auto-shutdown (in minutes) |
| `WORKMACHINE_AUTO_SHUTDOWN_WARNING_MINUTES` | `5` | How many minutes before shutdown to send warning |

### WMIngress Controller (`WMINGRESS_` prefix)

| Variable | Default | Description |
|-----------|----------|-------------|
| `WMINGRESS_HTTP_PORT` | `80` | HTTP port for the ingress controller |
| `WMINGRESS_HTTPS_PORT` | `443` | HTTPS port for the ingress controller |
| `WMINGRESS_WILDCARD_DOMAIN` | `` (empty) | Wildcard domain for TLS certificates (e.g., `khost.dev`). Empty means no domain filtering |
| `WMINGRESS_WILDCARD_SECRET_NAME` | `kloudlite-wildcard-cert-tls` | Name of the wildcard TLS certificate secret |
| `WMINGRESS_WILDCARD_SECRET_NAMESPACE` | `kloudlite` | Namespace of the wildcard TLS certificate secret |
| `WMINGRESS_REGISTRY_USERNAME` | `` (empty) | Username for registry path access control. When set, write operations to `cr.*` domains are restricted to `/v2/{username}/*` |
| `WMINGRESS_FORCE_FULL_REBUILD` | `false` | Force full rebuild on every event (for debugging) |
| `WMINGRESS_PROXY_TIMEOUT` | `30s` | Timeout for HTTP proxy connections |
| `WMINGRESS_PROXY_KEEP_ALIVE` | `30s` | Keep-alive duration for HTTP proxy connections |
| `WMINGRESS_PROXY_IDLE_CONN_TIMEOUT` | `90s` | Idle connection timeout for HTTP proxy |
| `WMINGRESS_PROXY_TLS_HANDSHAKE_TIMEOUT` | `10s` | TLS handshake timeout for HTTP proxy |
| `WMINGRESS_PROXY_EXPECT_CONTINUE_TIMEOUT` | `1s` | Expect continue timeout for HTTP proxy |
| `WMINGRESS_PROXY_MAX_IDLE_CONNS` | `100` | Maximum number of idle connections |

## Production Considerations

### Image Tags
For production deployments, replace `:latest` tags with specific version tags:
- `WORKSPACE_KUBECTL_IMAGE`: Use a specific version like `bitnami/kubectl:1.31.0`
- `WORKSPACE_GIT_IMAGE`: Use a specific version like `alpine/git:2.45.2`
- `WORKSPACE_ALPINE_IMAGE`: Use a specific version like `alpine:3.19`
- `WORKMACHINE_WM_INGRESS_CONTROLLER_IMAGE`: Use a production tag like `ghcr.io/kloudlite/kloudlite/wm-ingress-controller:latest`

### Timeout Values
Adjust timeout values based on your infrastructure:
- **Cloud provider**: Increase timeouts if cloud operations are slower
- **Network**: Adjust proxy timeouts based on network latency
- **Workload**: Adjust idle thresholds based on typical usage patterns

### Resource Limits
Image pull policies and resource limits should be tuned for your cluster:
- **Development**: Can use `ImagePullPolicy: Always` for rapid iteration
- **Production**: Use `ImagePullPolicy: IfNotPresent` for stability

### Security
- Ensure wildcard TLS secrets are properly secured
- Use strong SSH keys for workmachine access
- Configure registry username appropriately for multi-tenant deployments

## Example Configuration

### Development Environment
```bash
# Use latest images for rapid iteration
WORKSPACE_KUBECTL_IMAGE="bitnami/kubectl:latest"
WORKSPACE_GIT_IMAGE="alpine/git:latest"

# Shorter timeouts for faster feedback
WORKMACHINE_MACHINE_STARTUP_RETRY_INTERVAL="5s"
ENVIRONMENT_POD_TERMINATION_RETRY_INTERVAL="1s"

# Enable debug mode
WMINGRESS_FORCE_FULL_REBUILD="true"
```

### Production Environment
```bash
# Use specific image versions for stability
WORKSPACE_KUBECTL_IMAGE="bitnami/kubectl:1.31.0"
WORKSPACE_GIT_IMAGE="alpine/git:2.45.2"
WORKSPACE_ALPINE_IMAGE="alpine:3.19"
WORKMACHINE_WM_INGRESS_CONTROLLER_IMAGE="ghcr.io/kloudlite/kloudlite/wm-ingress-controller:latest"

# Longer timeouts for production infrastructure
WORKMACHINE_MACHINE_STARTUP_RETRY_INTERVAL="30s"
ENVIRONMENT_POD_TERMINATION_RETRY_INTERVAL="10s"

# Production domain configuration
WMINGRESS_WILDCARD_DOMAIN="khost.dev"
WMINGRESS_WILDCARD_SECRET_NAMESPACE="kloudlite"

# Disable debug mode
# (default is false, no need to set)
```

## Configuration Validation

The configuration loader performs basic validation:
- Required fields must be provided
- Duration fields must be in valid Go duration format
- Integer fields must be valid numbers
- Invalid values will cause the controller to fail startup with a clear error message

For production deployments, consider adding additional validation in your deployment scripts to catch misconfigurations early.

## Migration from Hardcoded Values

This change replaces the following hardcoded values with configurable environment variables:

### Workspace Controller
- `defaultIdleTimeoutMinutes = 30` → `WORKSPACE_DEFAULT_IDLE_TIMEOUT_MINUTES`
- `rbacCleanupIntervalMinutes = 60` → `WORKSPACE_RBAC_CLEANUP_INTERVAL_MINUTES`
- `"latest"` (VSCode version) → `WORKSPACE_VSCODE_VERSION`
- `1 * time.Minute` (requeue interval) → `WORKSPACE_REQUEUE_INTERVAL_MINUTES`

### WorkMachine Controller
- `SSHUserName = "kloudlite"` → `WORKMACHINE_SSH_USERNAME`
- `wmIngressControllerImage` → `WORKMACHINE_WM_INGRESS_CONTROLLER_IMAGE`
- `kloudliteWildcardCertName` → `WORKMACHINE_DEFAULT_WILDCARD_CERT_NAME`
- Various timeout constants → `WORKMACHINE_*` prefixed variables

### Environment Controller
- `2 * time.Second` (pod termination) → `ENVIRONMENT_POD_TERMINATION_RETRY_INTERVAL`
- `5 * time.Second` (various operations) → `ENVIRONMENT_FORK_RETRY_INTERVAL`, etc.

### WMIngress Controller
- Proxy timeouts (30s, 90s, 10s, 1s) → `WMINGRESS_PROXY_*` variables
- Wildcard secret configuration → `WMINGRESS_WILDCARD_*` variables
- Force full rebuild flag → `WMINGRESS_FORCE_FULL_REBUILD`

## Troubleshooting

### Controllers Not Starting
If controllers fail to start, check:
1. All required environment variables are set
2. Duration values use correct format (e.g., `30s`, not `30`)
3. Integer values are valid numbers
4. Check controller logs for validation errors

### Timeouts Too Short/Long
Adjust timeout values based on your infrastructure:
- **Too short**: Increase retry interval values
- **Too long**: Decrease retry interval values for faster feedback

### Wrong Images Being Used
Verify environment variables are correctly set and loaded:
1. Check deployment configuration
2. Check process environment (e.g., `kubectl exec` into pod)
3. Review controller startup logs for loaded configuration

## Additional Resources

- [Go Duration Format](https://pkg.go.dev/time#ParseDuration)
- [Kubernetes Controller Runtime](https://book.kubebuilder.io/reference/controller-runtime.html)
- [Kloudlite Documentation](https://docs.kloudlite.io)
