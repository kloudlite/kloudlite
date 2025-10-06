package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	packagesv1 "github.com/kloudlite/kloudlite/api/pkg/apis/packages/v1"
	workspacesv1 "github.com/kloudlite/kloudlite/api/pkg/apis/workspaces/v1"
	zap2 "go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/metrics/server"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	nixStorePath = "/var/lib/kloudlite/nix-store"
)

// CommandExecutor defines an interface for executing shell commands
// This allows for mocking in tests
type CommandExecutor interface {
	Execute(script string) ([]byte, error)
}

// RealCommandExecutor implements CommandExecutor using os/exec
type RealCommandExecutor struct{}

func (r *RealCommandExecutor) Execute(script string) ([]byte, error) {
	cmd := exec.Command("sh", "-c", script)
	cmd.Env = append(os.Environ(), fmt.Sprintf("NIX_STATE_DIR=%s", nixStorePath))
	return cmd.CombinedOutput()
}

type PackageManagerReconciler struct {
	client.Client
	Scheme    *runtime.Scheme
	Logger    *zap2.Logger
	Namespace string
	CmdExec   CommandExecutor
}

func (r *PackageManagerReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	logger := r.Logger.With(
		zap2.String("packageRequest", req.Name),
		zap2.String("namespace", req.Namespace),
	)

	logger.Info("Reconciling PackageRequest")

	// Fetch PackageRequest
	pkgReq := &packagesv1.PackageRequest{}
	if err := r.Get(ctx, req.NamespacedName, pkgReq); err != nil {
		logger.Error("Failed to get PackageRequest", zap2.Error(err))
		return reconcile.Result{}, client.IgnoreNotFound(err)
	}

	// Skip if already in terminal state (Ready or Failed)
	if pkgReq.Status.Phase == "Ready" || pkgReq.Status.Phase == "Failed" {
		logger.Info("PackageRequest already in terminal state", zap2.String("phase", pkgReq.Status.Phase))
		return reconcile.Result{}, nil
	}

	// Update status to Installing
	pkgReq.Status.Phase = "Installing"
	pkgReq.Status.Message = "Installing packages"
	pkgReq.Status.LastUpdated = metav1.Now()
	if err := r.Status().Update(ctx, pkgReq); err != nil {
		logger.Error("Failed to update status to Installing", zap2.Error(err))
		return reconcile.Result{}, err
	}

	logger.Info("Using profile", zap2.String("profile", pkgReq.Spec.ProfileName))

	// Detect and remove packages that are no longer in the spec
	specPackageNames := make(map[string]bool)
	for _, pkg := range pkgReq.Spec.Packages {
		specPackageNames[pkg.Name] = true
	}

	for _, installedPkg := range pkgReq.Status.InstalledPackages {
		if !specPackageNames[installedPkg.Name] {
			logger.Info("Removing package",
				zap2.String("package", installedPkg.Name),
				zap2.String("profile", pkgReq.Spec.ProfileName))

			if err := r.uninstallPackage(installedPkg.Name, pkgReq.Spec.ProfileName); err != nil {
				logger.Error("Failed to remove package",
					zap2.String("package", installedPkg.Name),
					zap2.String("profile", pkgReq.Spec.ProfileName),
					zap2.Error(err))
			} else {
				logger.Info("Successfully removed package",
					zap2.String("package", installedPkg.Name),
					zap2.String("profile", pkgReq.Spec.ProfileName))
			}
		}
	}

	// Install packages
	installedPackages := []workspacesv1.InstalledPackage{}
	failedPackages := []string{}

	for _, pkg := range pkgReq.Spec.Packages {
		logger.Info("Installing package",
			zap2.String("package", pkg.Name),
			zap2.String("profile", pkgReq.Spec.ProfileName))

		installedPkg, err := r.installPackage(pkg, pkgReq.Spec.ProfileName)
		if err != nil {
			logger.Error("Failed to install package",
				zap2.String("package", pkg.Name),
				zap2.String("profile", pkgReq.Spec.ProfileName),
				zap2.Error(err))
			failedPackages = append(failedPackages, pkg.Name)
			continue
		}

		installedPackages = append(installedPackages, installedPkg)
		logger.Info("Successfully installed package",
			zap2.String("package", pkg.Name),
			zap2.String("profile", pkgReq.Spec.ProfileName),
			zap2.String("storePath", installedPkg.StorePath))
	}

	// Update PackageRequest status
	pkgReq.Status.InstalledPackages = installedPackages
	pkgReq.Status.FailedPackages = failedPackages
	pkgReq.Status.LastUpdated = metav1.Now()

	if len(failedPackages) > 0 {
		pkgReq.Status.Phase = "Failed"
		pkgReq.Status.Message = fmt.Sprintf("Failed to install %d packages", len(failedPackages))
	} else {
		pkgReq.Status.Phase = "Ready"
		pkgReq.Status.Message = fmt.Sprintf("Successfully installed %d packages", len(installedPackages))
	}

	if err := r.Status().Update(ctx, pkgReq); err != nil {
		logger.Error("Failed to update status", zap2.Error(err))
		return reconcile.Result{}, err
	}

	logger.Info("PackageRequest reconciliation complete",
		zap2.String("phase", pkgReq.Status.Phase),
		zap2.Int("installed", len(installedPackages)),
		zap2.Int("failed", len(failedPackages)))

	return reconcile.Result{}, nil
}

func (r *PackageManagerReconciler) installPackage(pkg workspacesv1.PackageSpec, profileName string) (workspacesv1.InstalledPackage, error) {
	// Determine package attribute and install command
	profilePath := fmt.Sprintf("/nix/var/nix/profiles/per-user/root/%s", profileName)

	var installScript string
	if pkg.Version != "" {
		// Install specific version using nixpkgs flake with version constraint
		// Format: package@version (e.g., "vim@9.1.0")
		// Uses nix profile install with flake reference
		installScript = fmt.Sprintf(
			`. /root/.nix-profile/etc/profile.d/nix.sh && nix profile install --profile %s 'nixpkgs#%s' || nix-env -p %s -iA 'nixpkgs.%s'`,
			profilePath, pkg.Name, profilePath, pkg.Name,
		)
		r.Logger.Info("Installing package with version preference",
			zap2.String("package", pkg.Name),
			zap2.String("version", pkg.Version))
	} else {
		// Install latest version from nixpkgs
		pkgAttr := "nixpkgs." + pkg.Name
		installScript = fmt.Sprintf(". /root/.nix-profile/etc/profile.d/nix.sh && nix-env -p %s -iA %s", profilePath, pkgAttr)
	}

	output, err := r.CmdExec.Execute(installScript)
	if err != nil {
		return workspacesv1.InstalledPackage{}, fmt.Errorf("nix-env failed: %w, output: %s", err, string(output))
	}

	// Query package info to get store path and actual installed version from the named profile
	queryScript := fmt.Sprintf(". /root/.nix-profile/etc/profile.d/nix.sh && nix-env -p %s -q --out-path %s", profilePath, pkg.Name)

	queryOutput, err := r.CmdExec.Execute(queryScript)
	if err != nil {
		r.Logger.Warn("Failed to query package path, using default",
			zap2.String("package", pkg.Name),
			zap2.Error(err))
	}

	// Parse store path and version from output
	storePath := nixStorePath + "/store"
	installedVersion := pkg.Version // Default to requested version
	if len(queryOutput) > 0 {
		// Output format is typically: "package-version  /nix/store/hash-package-version"
		parts := strings.Fields(string(queryOutput))
		if len(parts) >= 1 {
			// First part is "package-version", extract version
			pkgWithVersion := parts[0]
			// Remove package name prefix to get version
			if strings.HasPrefix(pkgWithVersion, pkg.Name+"-") {
				installedVersion = strings.TrimPrefix(pkgWithVersion, pkg.Name+"-")
			} else {
				installedVersion = pkgWithVersion
			}
		}
		if len(parts) >= 2 {
			storePath = parts[1]
		}
	}

	binPath := fmt.Sprintf("%s/bin", profilePath)

	return workspacesv1.InstalledPackage{
		Name:        pkg.Name,
		Version:     installedVersion,
		BinPath:     binPath,
		StorePath:   storePath,
		InstalledAt: metav1.Now(),
	}, nil
}

func (r *PackageManagerReconciler) uninstallPackage(pkgName string, profileName string) error {
	// Remove package from the profile (doesn't delete from Nix store)
	// Using nix-env -e only removes the package from the user environment
	profilePath := fmt.Sprintf("/nix/var/nix/profiles/per-user/root/%s", profileName)
	uninstallScript := fmt.Sprintf(". /root/.nix-profile/etc/profile.d/nix.sh && nix-env -p %s -e %s", profilePath, pkgName)

	output, err := r.CmdExec.Execute(uninstallScript)
	if err != nil {
		return fmt.Errorf("nix-env uninstall failed: %w, output: %s", err, string(output))
	}

	return nil
}

func (r *PackageManagerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&packagesv1.PackageRequest{}).
		WithEventFilter(predicate.Funcs{
			// Only reconcile on Create and when not in terminal state
			UpdateFunc: func(e event.UpdateEvent) bool {
				newPR, ok := e.ObjectNew.(*packagesv1.PackageRequest)
				if !ok {
					return false
				}
				// Don't reconcile if already in terminal state
				if newPR.Status.Phase == "Ready" || newPR.Status.Phase == "Failed" {
					return false
				}
				return true
			},
			// Don't reconcile on delete
			DeleteFunc: func(e event.DeleteEvent) bool {
				return false
			},
		}).
		Complete(r)
}

func main() {
	// Get namespace from environment
	namespace := os.Getenv("NAMESPACE")
	if namespace == "" {
		fmt.Println("NAMESPACE environment variable must be set")
		os.Exit(1)
	}

	// Setup logger using controller-runtime's zap logger
	opts := zap.Options{
		Development: false,
		Level:       zapcore.InfoLevel,
	}
	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	// Create a native zap logger for our own use
	zapLogger, err := zap2.NewProduction()
	if err != nil {
		fmt.Printf("Failed to create logger: %v\n", err)
		os.Exit(1)
	}
	defer zapLogger.Sync()

	zapLogger.Info("Starting Package Manager",
		zap2.String("namespace", namespace),
		zap2.String("nixStorePath", nixStorePath))

	// Setup scheme
	scheme := runtime.NewScheme()
	_ = clientgoscheme.AddToScheme(scheme)
	_ = packagesv1.AddToScheme(scheme)
	_ = workspacesv1.AddToScheme(scheme)

	// Get in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		zapLogger.Fatal("Failed to get in-cluster config", zap2.Error(err))
	}

	// Create manager watching only the specified namespace
	mgr, err := ctrl.NewManager(config, ctrl.Options{
		Scheme:         scheme,
		LeaderElection: false, // Each Deployment manages only its namespace
		Metrics: server.Options{
			BindAddress: ":8080",
		},
		Cache: cache.Options{
			DefaultNamespaces: map[string]cache.Config{
				namespace: {},
			},
		},
	})
	if err != nil {
		zapLogger.Fatal("Failed to create manager", zap2.Error(err))
	}

	// Setup reconciler
	reconciler := &PackageManagerReconciler{
		Client:    mgr.GetClient(),
		Scheme:    mgr.GetScheme(),
		Logger:    zapLogger,
		Namespace: namespace,
		CmdExec:   &RealCommandExecutor{},
	}

	if err := reconciler.SetupWithManager(mgr); err != nil {
		zapLogger.Fatal("Failed to setup controller", zap2.Error(err))
	}

	zapLogger.Info("Starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		zapLogger.Fatal("Failed to start manager", zap2.Error(err))
	}
}
