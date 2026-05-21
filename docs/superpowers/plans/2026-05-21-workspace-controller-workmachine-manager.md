# Workspace Controller in WorkMachine Manager Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Move Workspace reconciliation into the per-WorkMachine `workmachine-manager` and ensure each manager reconciles only Workspaces belonging to its own WorkMachine.

**Architecture:** Add explicit scope fields and scope checks to `WorkspaceReconciler`, register a scoped instance from `newMachineScopedManager`, and remove Workspace registration from the API manager path. Keep existing Workspace behavior unchanged for unscoped unit-test/legacy construction while making partial scope fail closed.

**Tech Stack:** Go, controller-runtime, Kubernetes client-go, Kloudlite Workspace/WorkMachine CRDs, Go unit tests with fake controller-runtime clients.

---

## File Structure

- Modify `controllers/workspace/workspace_controller.go`
  - Add `OwnNamespace` and `WorkMachineName` fields to `WorkspaceReconciler`.
  - Add `scopeConfigured()` and `shouldReconcileWorkspace()` helpers.
  - Gate `Reconcile` after fetching the Workspace and before mutation.
  - Filter `findWorkspacesForEnvironment` results through the same scope helper.
- Modify `controllers/workspace/pod_lifecycle.go`
  - Guard `getWorkMachine(ctx, name)` so scoped reconcilers reject non-current WorkMachine names before API lookup.
- Modify `controllers/manager.go`
  - Add `MachineScopedControllerNames()`.
  - Register a scoped `workspace.WorkspaceReconciler` inside `newMachineScopedManager`.
  - Remove Workspace controller setup from the non-platform API manager path.
- Modify `controllers/workspace/workspace_controller_test.go`
  - Add unit tests for Workspace scope behavior and WorkMachine lookup guard.
- Modify `controllers/manager_test.go`
  - Add unit test for machine-scoped controller name list.

---

### Task 1: Add Workspace scope tests

**Files:**
- Modify: `controllers/workspace/workspace_controller_test.go`
- Modify: `controllers/manager_test.go`

- [ ] **Step 1: Add failing Workspace scope tests**

Add these tests near the top of `controllers/workspace/workspace_controller_test.go`, after `TestWorkspaceReconciler_Reconcile_NotFound`:

```go
func TestWorkspaceReconcilerScopeAllowsCurrentWorkMachineWorkspace(t *testing.T) {
	reconciler := &WorkspaceReconciler{OwnNamespace: "wm-alice", WorkMachineName: "alice-dev"}
	workspace := &workspacev1.Workspace{
		ObjectMeta: metav1.ObjectMeta{Name: "ws", Namespace: "wm-alice"},
		Spec:       workspacev1.WorkspaceSpec{WorkmachineName: "alice-dev"},
	}

	if !reconciler.shouldReconcileWorkspace(workspace) {
		t.Fatalf("shouldReconcileWorkspace returned false, want true")
	}
}

func TestWorkspaceReconcilerScopeSkipsWrongNamespace(t *testing.T) {
	reconciler := &WorkspaceReconciler{OwnNamespace: "wm-alice", WorkMachineName: "alice-dev"}
	workspace := &workspacev1.Workspace{
		ObjectMeta: metav1.ObjectMeta{Name: "ws", Namespace: "wm-bob"},
		Spec:       workspacev1.WorkspaceSpec{WorkmachineName: "alice-dev"},
	}

	if reconciler.shouldReconcileWorkspace(workspace) {
		t.Fatalf("shouldReconcileWorkspace returned true for workspace in another namespace")
	}
}

func TestWorkspaceReconcilerScopeSkipsWrongWorkMachine(t *testing.T) {
	reconciler := &WorkspaceReconciler{OwnNamespace: "wm-alice", WorkMachineName: "alice-dev"}
	workspace := &workspacev1.Workspace{
		ObjectMeta: metav1.ObjectMeta{Name: "ws", Namespace: "wm-alice"},
		Spec:       workspacev1.WorkspaceSpec{WorkmachineName: "bob-dev"},
	}

	if reconciler.shouldReconcileWorkspace(workspace) {
		t.Fatalf("shouldReconcileWorkspace returned true for workspace on another WorkMachine")
	}
}

func TestWorkspaceReconcilerScopeFailsClosedWhenConfiguredIncomplete(t *testing.T) {
	workspace := &workspacev1.Workspace{
		ObjectMeta: metav1.ObjectMeta{Name: "ws", Namespace: "wm-alice"},
		Spec:       workspacev1.WorkspaceSpec{WorkmachineName: "alice-dev"},
	}

	if (&WorkspaceReconciler{OwnNamespace: "wm-alice"}).shouldReconcileWorkspace(workspace) {
		t.Fatalf("shouldReconcileWorkspace returned true with missing WorkMachineName")
	}
	if (&WorkspaceReconciler{WorkMachineName: "alice-dev"}).shouldReconcileWorkspace(workspace) {
		t.Fatalf("shouldReconcileWorkspace returned true with missing OwnNamespace")
	}
}

func TestWorkspaceReconcilerUnscopedAllowsLegacyController(t *testing.T) {
	workspace := &workspacev1.Workspace{
		ObjectMeta: metav1.ObjectMeta{Name: "ws", Namespace: "wm-alice"},
		Spec:       workspacev1.WorkspaceSpec{WorkmachineName: "alice-dev"},
	}

	if !(&WorkspaceReconciler{}).shouldReconcileWorkspace(workspace) {
		t.Fatalf("unscoped shouldReconcileWorkspace returned false, want true")
	}
}

func TestWorkspaceReconcilerScopedGetWorkMachineRejectsOtherWorkMachine(t *testing.T) {
	reconciler := &WorkspaceReconciler{WorkMachineName: "alice-dev"}

	_, err := reconciler.getWorkMachine(context.Background(), "bob-dev")

	if err == nil {
		t.Fatalf("getWorkMachine returned nil error for another WorkMachine")
	}
	if !strings.Contains(err.Error(), "outside controller scope") {
		t.Fatalf("error = %v, want outside controller scope", err)
	}
}
```

Add this test to `controllers/manager_test.go`:

```go
func TestMachineScopedControllerNames(t *testing.T) {
	want := []string{"workmachine-manager", "environment", "workspace"}
	got := MachineScopedControllerNames()

	if len(got) != len(want) {
		t.Fatalf("MachineScopedControllerNames() = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("MachineScopedControllerNames() = %v, want %v", got, want)
		}
	}
}
```

- [ ] **Step 2: Run tests to verify they fail for missing implementation**

Run:

```bash
go test ./controllers/workspace -run 'TestWorkspaceReconcilerScope|TestWorkspaceReconcilerUnscoped|TestWorkspaceReconcilerScopedGetWorkMachine' -v
go test ./controllers -run TestMachineScopedControllerNames -v
```

Expected: compile/test failure mentioning missing `OwnNamespace`, `WorkMachineName`, `shouldReconcileWorkspace`, or `MachineScopedControllerNames`.

- [ ] **Step 3: Commit the red tests**

```bash
git add controllers/workspace/workspace_controller_test.go controllers/manager_test.go
git commit -m "test: define workspace manager scoping"
```

---

### Task 2: Implement Workspace scope helpers and WorkMachine lookup guard

**Files:**
- Modify: `controllers/workspace/workspace_controller.go`
- Modify: `controllers/workspace/pod_lifecycle.go`

- [ ] **Step 1: Add scope fields and helpers**

In `controllers/workspace/workspace_controller.go`, update `WorkspaceReconciler`:

```go
type WorkspaceReconciler struct {
	client.Client
	Scheme          *runtime.Scheme
	Logger          *zap.Logger
	Config          *rest.Config
	Clientset       *kubernetes.Clientset
	JWTSecret       string            // JWT secret (kept for compatibility, no longer used for registry)
	Cfg             *ControllerConfig // Controller configuration
	OwnNamespace    string
	WorkMachineName string
}
```

Add helpers near `ensureConfig()` or immediately after the struct:

```go
func (r *WorkspaceReconciler) scopeConfigured() bool {
	return r.OwnNamespace != "" || r.WorkMachineName != ""
}

func (r *WorkspaceReconciler) shouldReconcileWorkspace(workspace *workspacev1.Workspace) bool {
	if !r.scopeConfigured() {
		return true
	}
	if r.OwnNamespace == "" || r.WorkMachineName == "" {
		return false
	}
	return workspace.Namespace == r.OwnNamespace && workspace.Spec.WorkmachineName == r.WorkMachineName
}
```

- [ ] **Step 2: Guard WorkMachine lookup**

In `controllers/workspace/pod_lifecycle.go`, update `getWorkMachine`:

```go
func (r *WorkspaceReconciler) getWorkMachine(ctx context.Context, name string) (*machinesv1.WorkMachine, error) {
	if r.WorkMachineName != "" && name != r.WorkMachineName {
		return nil, fmt.Errorf("workmachine %s is outside controller scope %s", name, r.WorkMachineName)
	}

	wm := &machinesv1.WorkMachine{}
	if err := r.Get(ctx, client.ObjectKey{Name: name}, wm); err != nil {
		return nil, fmt.Errorf("failed to get WorkMachine %s: %w", name, err)
	}
	return wm, nil
}
```

- [ ] **Step 3: Run Task 1 tests to verify green for scope helpers**

Run:

```bash
go test ./controllers/workspace -run 'TestWorkspaceReconcilerScope|TestWorkspaceReconcilerUnscoped|TestWorkspaceReconcilerScopedGetWorkMachine' -v
```

Expected: PASS.

- [ ] **Step 4: Commit scope helper implementation**

```bash
git add controllers/workspace/workspace_controller.go controllers/workspace/pod_lifecycle.go
git commit -m "feat: scope workspace reconciler by workmachine"
```

---

### Task 3: Gate reconciliation and event mapping by scope

**Files:**
- Modify: `controllers/workspace/workspace_controller.go`
- Test: `controllers/workspace/workspace_controller_test.go`

- [ ] **Step 1: Add a failing Reconcile skip test**

Add this test after the scope helper tests in `controllers/workspace/workspace_controller_test.go`:

```go
func TestWorkspaceReconcilerReconcileSkipsOutOfScopeWorkspaceWithoutMutation(t *testing.T) {
	scheme := testutil.NewTestScheme()
	workspace := &workspacev1.Workspace{
		ObjectMeta: metav1.ObjectMeta{Name: "ws", Namespace: "wm-bob"},
		Spec: workspacev1.WorkspaceSpec{
			DisplayName:     "Bob Workspace",
			OwnedBy:         "bob@example.com",
			WorkmachineName: "bob-dev",
		},
	}
	k8sClient := testutil.NewFakeClient(scheme, workspace).
		WithStatusSubresource(&packagesv1.PackageRequest{}, &workspacev1.Workspace{}).
		Build()
	reconciler := &WorkspaceReconciler{
		Client:          k8sClient,
		Scheme:          scheme,
		Logger:          zaptest.NewLogger(t),
		OwnNamespace:    "wm-alice",
		WorkMachineName: "alice-dev",
	}

	result, err := reconciler.Reconcile(context.Background(), reconcile.Request{NamespacedName: types.NamespacedName{Name: "ws", Namespace: "wm-bob"}})

	require.NoError(t, err)
	assert.False(t, result.Requeue)
	updated := &workspacev1.Workspace{}
	require.NoError(t, k8sClient.Get(context.Background(), types.NamespacedName{Name: "ws", Namespace: "wm-bob"}, updated))
	assert.Empty(t, updated.Finalizers)
	assert.Empty(t, updated.Labels["kloudlite.io/hash"])
}
```

- [ ] **Step 2: Run the new test to verify it fails**

Run:

```bash
go test ./controllers/workspace -run TestWorkspaceReconcilerReconcileSkipsOutOfScopeWorkspaceWithoutMutation -v
```

Expected: FAIL because the current reconciler adds finalizers/hash before scope gating.

- [ ] **Step 3: Gate `Reconcile` after fetching the Workspace**

In `WorkspaceReconciler.Reconcile`, immediately after successfully fetching `workspace`, add:

```go
	if !r.shouldReconcileWorkspace(workspace) {
		logger.Info("skipping workspace outside controller scope",
			zap.String("own_namespace", r.OwnNamespace),
			zap.String("workmachine_name", r.WorkMachineName),
			zap.String("workspace_namespace", workspace.Namespace),
			zap.String("workspace_workmachine", workspace.Spec.WorkmachineName))
		return reconcile.Result{}, nil
	}
```

This block must be before deletion handling, finalizer changes, label changes, RBAC setup, owner reference changes, and lifecycle handling.

- [ ] **Step 4: Filter environment-triggered workspace enqueueing**

In `findWorkspacesForEnvironment`, change the loop condition from:

```go
		if ws.Spec.EnvironmentConnection != nil &&
			ws.Spec.EnvironmentConnection.EnvironmentRef.Name == env.Name {
```

to:

```go
		if !r.shouldReconcileWorkspace(&ws) {
			continue
		}
		if ws.Spec.EnvironmentConnection != nil &&
			ws.Spec.EnvironmentConnection.EnvironmentRef.Name == env.Name {
```

- [ ] **Step 5: Run focused workspace tests**

Run:

```bash
go test ./controllers/workspace -run 'TestWorkspaceReconcilerScope|TestWorkspaceReconcilerUnscoped|TestWorkspaceReconcilerScopedGetWorkMachine|TestWorkspaceReconcilerReconcileSkipsOutOfScopeWorkspaceWithoutMutation' -v
```

Expected: PASS.

- [ ] **Step 6: Commit reconcile gating**

```bash
git add controllers/workspace/workspace_controller.go controllers/workspace/workspace_controller_test.go
git commit -m "fix: skip out-of-scope workspaces"
```

---

### Task 4: Register Workspace controller in workmachine-manager and remove API registration

**Files:**
- Modify: `controllers/manager.go`
- Test: `controllers/manager_test.go`

- [ ] **Step 1: Add `MachineScopedControllerNames`**

In `controllers/manager.go`, after `PlatformAPIControllerNames`, add:

```go
func MachineScopedControllerNames() []string {
	return []string{"workmachine-manager", "environment", "workspace"}
}
```

- [ ] **Step 2: Register scoped Workspace controller in `newMachineScopedManager`**

In `newMachineScopedManager`, after Environment controller registration and before `machinescoped.Register`, create a clientset and register Workspace:

```go
	clientset, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("unable to create kubernetes clientset: %w", err)
	}

	workspaceReconciler := &workspace.WorkspaceReconciler{
		Client:          mgr.GetClient(),
		Scheme:          mgr.GetScheme(),
		Logger:          logger.With(zap.String("controller", "workspace")),
		Config:          cfg,
		Clientset:       clientset,
		JWTSecret:       authCfg.JWTSecret,
		Cfg:             controllerCfg,
		OwnNamespace:    os.Getenv("POD_NAMESPACE"),
		WorkMachineName: os.Getenv("WORKMACHINE_NAME"),
	}
	if workspaceReconciler.OwnNamespace == "" {
		workspaceReconciler.OwnNamespace = os.Getenv("NAMESPACE")
	}
	if err := workspaceReconciler.SetupWithManager(mgr); err != nil {
		return nil, fmt.Errorf("unable to setup scoped Workspace controller: %w", err)
	}
```

Update the success log to use `MachineScopedControllerNames()`:

```go
	logger.Info("controllers initialized successfully", zap.Strings("controllers", MachineScopedControllerNames()))
```

- [ ] **Step 3: Remove Workspace controller registration from API manager path**

In `newManager`, remove the block that creates `clientset`, creates `workspaceReconciler`, and calls `workspaceReconciler.SetupWithManager(mgr)`. Leave snapshot controller setup intact.

If `k8s.io/client-go/kubernetes` is now only used by `newMachineScopedManager`, keep the import. If it is unused, `go test` will identify it.

- [ ] **Step 4: Run manager tests**

Run:

```bash
go test ./controllers -run 'TestPlatformAPIControllerNames|TestMachineScopedControllerNames' -v
```

Expected: PASS.

- [ ] **Step 5: Commit manager registration changes**

```bash
git add controllers/manager.go controllers/manager_test.go
git commit -m "feat: run workspace controller in workmachine manager"
```

---

### Task 5: Full verification and cleanup

**Files:**
- No required file edits unless verification reveals issues.

- [ ] **Step 1: Format changed Go files**

Run:

```bash
gofmt -w controllers/workspace/workspace_controller.go controllers/workspace/pod_lifecycle.go controllers/workspace/workspace_controller_test.go controllers/manager.go controllers/manager_test.go
```

Expected: no output.

- [ ] **Step 2: Run focused verification**

Run:

```bash
go test ./controllers/workspace ./controllers -v
```

Expected: PASS.

- [ ] **Step 3: Run full Go verification**

Run:

```bash
go test ./...
```

Expected: PASS.

- [ ] **Step 4: Inspect final status and commit any formatting-only changes**

Run:

```bash
git status --short
```

If files changed only due to `gofmt`, commit them:

```bash
git add controllers/workspace/workspace_controller.go controllers/workspace/pod_lifecycle.go controllers/workspace/workspace_controller_test.go controllers/manager.go controllers/manager_test.go
git commit -m "chore: format workspace manager changes"
```

If no files changed, do not create an empty commit.

---

## Plan Self-Review

- Spec coverage: Tasks cover scoped Workspace reconciliation, WorkMachine lookup guard, machine manager registration, platform/API deregistration, tests, and verification.
- Placeholder scan: no placeholder implementation steps remain; each code change has exact snippets and commands.
- Type consistency: uses existing `WorkspaceSpec.WorkmachineName`, `WorkspaceReconciler`, `MachineScopedControllerNames`, and controller-runtime patterns from the codebase.
