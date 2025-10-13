package cmd

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/kloudlite/kloudlite/api/cmd/kl/pkg/devbox"
	fuzzyfinder "github.com/ktr0731/go-fuzzyfinder"
	workspacesv1 "github.com/kloudlite/kloudlite/api/pkg/apis/workspaces/v1"
	"github.com/spf13/cobra"
)

var (
	pkgVersion string
	pkgChannel string
	pkgCommit  string
)

var pkgCmd = &cobra.Command{
	Use:     "pkg",
	Aliases: []string{"package", "p"},
	Short:   "Manage Nix packages in the workspace",
	Long:    `Search, install, and manage Nix packages in your workspace using the Devbox package registry.`,
	Example: `  # Search for packages
  kl pkg search nodejs
  kl p s vim

  # Add packages interactively
  kl pkg add
  kl p a

  # Add packages by name
  kl pkg add git vim curl
  kl p a nodejs

  # Install with specific version
  kl pkg install nodejs --version 20.0.0
  kl p i python --channel nixos-24.05

  # List installed packages
  kl pkg list
  kl p ls

  # Uninstall a package
  kl pkg uninstall git
  kl p rm vim`,
}

var pkgSearchCmd = &cobra.Command{
	Use:     "search <query>",
	Aliases: []string{"find", "s"},
	Short:   "Search for Nix packages",
	Long:    `Search for available Nix packages using the Devbox package registry.`,
	Example: `  kl pkg search nodejs
  kl p s python
  kl p find vim`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return handlePackageSearch(args[0])
	},
}

var pkgAddCmd = &cobra.Command{
	Use:     "add [package-names...]",
	Aliases: []string{"a"},
	Short:   "Add packages to the workspace",
	Long: `Add one or more packages to your workspace.

When run without arguments, enters interactive mode with fuzzy search and version selection.
When run with package names, automatically installs the latest version of each package.`,
	Example: `  # Interactive mode
  kl pkg add
  kl p a

  # Add multiple packages
  kl pkg add git vim curl
  kl p a nodejs python`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return handlePackageAdd(args)
	},
}

var pkgInstallCmd = &cobra.Command{
	Use:     "install <package-name>",
	Aliases: []string{"i", "in"},
	Short:   "Install a Nix package with specific version or channel",
	Long: `Install a specific Nix package with optional version, channel, or commit specification.

Use --version to install a specific semantic version.
Use --channel to install from a specific nixpkgs channel.
Use --commit to install from a specific nixpkgs commit.`,
	Example: `  # Install latest version
  kl pkg install nodejs
  kl p i python

  # Install specific version
  kl pkg install nodejs --version 20.0.0
  kl p i python --version 3.11.0

  # Install from channel
  kl pkg install vim --channel nixos-24.05
  kl p i git --channel unstable

  # Install from specific commit
  kl pkg install curl --commit abc123def456`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return handlePackageInstall(args[0], pkgVersion, pkgChannel, pkgCommit)
	},
}

var pkgUninstallCmd = &cobra.Command{
	Use:     "uninstall <package-name>",
	Aliases: []string{"remove", "rm", "un"},
	Short:   "Uninstall a Nix package",
	Long:    `Remove a package from your workspace.`,
	Example: `  kl pkg uninstall git
  kl p rm vim
  kl p un nodejs`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return handlePackageUninstall(args[0])
	},
}

var pkgListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls", "l"},
	Short:   "List installed packages",
	Long:    `Display all packages in the workspace spec and their installation status.`,
	Example: `  kl pkg list
  kl p ls
  kl p l`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return handlePackageList()
	},
}

func init() {
	// Add flags to install command
	pkgInstallCmd.Flags().StringVar(&pkgVersion, "version", "", "Semantic version (e.g. 24.0.0)")
	pkgInstallCmd.Flags().StringVar(&pkgChannel, "channel", "", "Nixpkgs channel (e.g. nixos-24.05)")
	pkgInstallCmd.Flags().StringVar(&pkgCommit, "commit", "", "Exact nixpkgs commit hash")

	// Add subcommands
	pkgCmd.AddCommand(pkgSearchCmd)
	pkgCmd.AddCommand(pkgAddCmd)
	pkgCmd.AddCommand(pkgInstallCmd)
	pkgCmd.AddCommand(pkgUninstallCmd)
	pkgCmd.AddCommand(pkgListCmd)
}

func handlePackageSearch(query string) error {
	fmt.Printf("Searching for packages matching '%s'...\n\n", query)

	searchResp, err := devbox.SearchPackages(query)
	if err != nil {
		return err
	}

	if searchResp.NumResults == 0 {
		fmt.Println("No packages found")
		return nil
	}

	fmt.Printf("Found %d packages:\n\n", searchResp.NumResults)
	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "NAME\tVERSIONS")
	for _, pkg := range searchResp.Packages {
		fmt.Fprintf(tw, "%s\t%d\n", pkg.Name, pkg.NumVersions)
	}
	tw.Flush()

	return nil
}

func handlePackageAdd(args []string) error {
	if err := InitClient(); err != nil {
		return err
	}

	if len(args) > 0 {
		return packageAddByNames(args)
	}
	return packageAddInteractive()
}

func packageAddByNames(packageNames []string) error {
	ctx := context.Background()

	for _, pkgName := range packageNames {
		fmt.Printf("\n[+] Processing package: %s\n", pkgName)

		// Search for the package
		fmt.Printf(" Searching for '%s'...\n", pkgName)
		searchResp, err := devbox.SearchPackages(pkgName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "  [!] Failed to search for %s: %v\n", pkgName, err)
			continue
		}

		if len(searchResp.Packages) == 0 {
			fmt.Fprintf(os.Stderr, "  [!] Package not found: %s\n", pkgName)
			continue
		}

		// Use the first (best) match
		selectedPkg := searchResp.Packages[0]
		if len(selectedPkg.Versions) == 0 {
			fmt.Fprintf(os.Stderr, "  [!] No versions available for %s\n", pkgName)
			continue
		}

		// Use the latest version
		selectedVersion := selectedPkg.Versions[0].Version

		// Resolve version to commit hash
		fmt.Printf(" Resolving %s@%s...\n", selectedPkg.Name, selectedVersion)
		resolveResp, err := devbox.ResolvePackageVersion(selectedPkg.Name, selectedVersion)
		if err != nil {
			fmt.Fprintf(os.Stderr, "  [!] Failed to resolve %s: %v\n", pkgName, err)
			continue
		}

		if len(resolveResp.Systems) == 0 {
			fmt.Fprintf(os.Stderr, "  [!] No system information available for %s\n", pkgName)
			continue
		}

		// Get commit hash from first system
		var commitHash, attrPath string
		for _, sysInfo := range resolveResp.Systems {
			commitHash = sysInfo.FlakeInstallable.Ref.Rev
			attrPath = sysInfo.FlakeInstallable.AttrPath
			break
		}

		fmt.Printf("[✓] Resolved to commit %s\n", commitHash[:8])

		// Add to workspace
		workspace, err := WsClient.Get(ctx)
		if err != nil {
			fmt.Fprintf(os.Stderr, "  [!] Failed to get workspace: %v\n", err)
			continue
		}

		// Check if package already exists
		alreadyExists := false
		for _, pkg := range workspace.Spec.Packages {
			if pkg.Name == attrPath {
				fmt.Printf("  [!] Package %s is already in the workspace spec\n", attrPath)
				alreadyExists = true
				break
			}
		}

		if alreadyExists {
			continue
		}

		newPackage := workspacesv1.PackageSpec{
			Name:          attrPath,
			NixpkgsCommit: commitHash,
		}

		workspace.Spec.Packages = append(workspace.Spec.Packages, newPackage)

		if err := WsClient.Update(ctx, workspace); err != nil {
			fmt.Fprintf(os.Stderr, "  [!] Failed to update workspace: %v\n", err)
			continue
		}

		fmt.Printf("[✓] Package %s@%s added to workspace\n", selectedPkg.Name, selectedVersion)
		fmt.Println("[...] Waiting for installation to complete...")

		// Wait for package installation
		if err := waitForPackageInstallation(ctx, attrPath); err != nil {
			fmt.Fprintf(os.Stderr, "  [!] Installation failed: %v\n", err)
		}
	}

	return nil
}

func packageAddInteractive() error {
	// Step 1: Dynamic search with fuzzyfinder
	selectedPkg, err := searchPackagesDynamic()
	if err != nil {
		return err
	}

	if len(selectedPkg.Versions) == 0 {
		return fmt.Errorf("no versions available for %s", selectedPkg.Name)
	}

	// Step 2: Use fzf to select version
	selectedVersion, err := selectVersionWithFzf(selectedPkg.Versions, selectedPkg.Name)
	if err != nil {
		return err
	}

	// Resolve version to commit hash
	fmt.Printf("\n[+] Resolving %s@%s...\n", selectedPkg.Name, selectedVersion)
	resolveResp, err := devbox.ResolvePackageVersion(selectedPkg.Name, selectedVersion)
	if err != nil {
		return err
	}

	if len(resolveResp.Systems) == 0 {
		return fmt.Errorf("no system information available for this package version")
	}

	// Get commit hash from first system
	var commitHash, attrPath string
	for _, sysInfo := range resolveResp.Systems {
		commitHash = sysInfo.FlakeInstallable.Ref.Rev
		attrPath = sysInfo.FlakeInstallable.AttrPath
		break
	}

	fmt.Printf("[✓] Resolved to commit %s\n", commitHash[:8])

	// Add to workspace
	ctx := context.Background()
	workspace, err := WsClient.Get(ctx)
	if err != nil {
		return err
	}

	// Check if package already exists
	for _, pkg := range workspace.Spec.Packages {
		if pkg.Name == attrPath {
			return fmt.Errorf("package %s is already in the workspace spec", attrPath)
		}
	}

	newPackage := workspacesv1.PackageSpec{
		Name:          attrPath,
		NixpkgsCommit: commitHash,
	}

	workspace.Spec.Packages = append(workspace.Spec.Packages, newPackage)

	if err := WsClient.Update(ctx, workspace); err != nil {
		return fmt.Errorf("failed to update workspace: %w", err)
	}

	fmt.Printf("[✓] Package %s@%s added to workspace\n", selectedPkg.Name, selectedVersion)
	fmt.Println("[...] Waiting for installation to complete...")

	// Wait for package installation
	return waitForPackageInstallation(ctx, attrPath)
}

func searchPackagesDynamic() (*devbox.PackageSearchResult, error) {
	// Prompt for search query
	fmt.Print("Search packages: ")
	var query string
	fmt.Scanln(&query)

	if len(query) < 2 {
		return nil, fmt.Errorf("search query must be at least 2 characters")
	}

	// Search for packages
	fmt.Printf(" Searching for '%s'...\n", query)
	resp, err := devbox.SearchPackages(query)
	if err != nil {
		return nil, err
	}

	if len(resp.Packages) == 0 {
		return nil, fmt.Errorf("no packages found")
	}

	// Use fuzzyfinder to select from results
	idx, err := fuzzyfinder.Find(
		resp.Packages,
		func(i int) string {
			return fmt.Sprintf("%s (%d versions)", resp.Packages[i].Name, resp.Packages[i].NumVersions)
		},
		fuzzyfinder.WithPromptString(fmt.Sprintf("Select package (%d results): ", len(resp.Packages))),
		fuzzyfinder.WithPreviewWindow(func(i, w, h int) string {
			if i < 0 || i >= len(resp.Packages) {
				return ""
			}
			pkg := resp.Packages[i]
			preview := fmt.Sprintf("Package: %s\n", pkg.Name)
			preview += fmt.Sprintf("Versions: %d available\n", pkg.NumVersions)
			if len(pkg.Versions) > 0 {
				preview += "\nLatest versions:\n"
				for j := 0; j < 5 && j < len(pkg.Versions); j++ {
					preview += fmt.Sprintf("  - %s\n", pkg.Versions[j].Version)
				}
			}
			return preview
		}),
	)

	if err != nil {
		return nil, fmt.Errorf("selection cancelled")
	}

	return &resp.Packages[idx], nil
}

func selectVersionWithFzf(versions []devbox.PackageVersion, pkgName string) (string, error) {
	idx, err := fuzzyfinder.Find(
		versions,
		func(i int) string {
			return versions[i].Version
		},
		fuzzyfinder.WithPromptString(fmt.Sprintf("Select version for %s: ", pkgName)),
	)
	if err != nil {
		return "", fmt.Errorf("selection cancelled")
	}

	return versions[idx].Version, nil
}

func waitForPackageInstallation(ctx context.Context, packageName string) error {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	timeout := time.After(5 * time.Minute)

	for {
		select {
		case <-timeout:
			return fmt.Errorf("timeout waiting for package installation")

		case <-ticker.C:
			workspace, err := WsClient.Get(ctx)
			if err != nil {
				return fmt.Errorf("failed to get workspace status: %w", err)
			}

			// Check if package is installed
			for _, pkg := range workspace.Status.InstalledPackages {
				if pkg.Name == packageName {
					fmt.Printf("\n[✓] Package installed successfully: %s\n", packageName)
					if pkg.Version != "" {
						fmt.Printf("  Version: %s\n", pkg.Version)
					}
					if pkg.BinPath != "" {
						fmt.Printf("  Binary path: %s\n", pkg.BinPath)
					}
					return nil
				}
			}

			// Check if package failed
			for _, failedPkg := range workspace.Status.FailedPackages {
				if failedPkg == packageName {
					errMsg := workspace.Status.PackageInstallationMessage
					if errMsg != "" {
						return fmt.Errorf("package installation failed: %s", errMsg)
					}
					return fmt.Errorf("package installation failed")
				}
			}

			// Still installing - show a dot to indicate progress
			fmt.Print(".")
		}
	}
}

func handlePackageInstall(packageName, version, channel, commit string) error {
	if err := InitClient(); err != nil {
		return err
	}

	// If version is specified, resolve it to a commit
	if version != "" {
		fmt.Printf("Resolving %s@%s...\n", packageName, version)
		resolveResp, err := devbox.ResolvePackageVersion(packageName, version)
		if err != nil {
			return err
		}

		if len(resolveResp.Systems) == 0 {
			return fmt.Errorf("no system information available for this package version")
		}

		// Get commit hash and attr path from first system
		for _, sysInfo := range resolveResp.Systems {
			commit = sysInfo.FlakeInstallable.Ref.Rev
			packageName = sysInfo.FlakeInstallable.AttrPath
			break
		}

		fmt.Printf("[✓] Resolved to commit %s\n", commit[:8])
	}

	ctx := context.Background()
	workspace, err := WsClient.Get(ctx)
	if err != nil {
		return err
	}

	// Check if package is already installed
	for _, pkg := range workspace.Spec.Packages {
		if pkg.Name == packageName {
			return fmt.Errorf("package %s is already in the workspace spec", packageName)
		}
	}

	// Add package to the workspace spec
	newPackage := workspacesv1.PackageSpec{
		Name:          packageName,
		Channel:       channel,
		NixpkgsCommit: commit,
	}

	workspace.Spec.Packages = append(workspace.Spec.Packages, newPackage)

	if err := WsClient.Update(ctx, workspace); err != nil {
		return fmt.Errorf("failed to update workspace: %w", err)
	}

	fmt.Printf("[✓] Package %s added to workspace\n", packageName)
	fmt.Println("[...] Waiting for installation to complete...")

	// Wait for package installation
	return waitForPackageInstallation(ctx, packageName)
}

func handlePackageUninstall(packageName string) error {
	if err := InitClient(); err != nil {
		return err
	}

	ctx := context.Background()
	workspace, err := WsClient.Get(ctx)
	if err != nil {
		return err
	}

	// Find and remove the package
	found := false
	newPackages := []workspacesv1.PackageSpec{}
	for _, pkg := range workspace.Spec.Packages {
		if pkg.Name != packageName {
			newPackages = append(newPackages, pkg)
		} else {
			found = true
		}
	}

	if !found {
		return fmt.Errorf("package %s not found in workspace spec", packageName)
	}

	workspace.Spec.Packages = newPackages

	if err := WsClient.Update(ctx, workspace); err != nil {
		return fmt.Errorf("failed to update workspace: %w", err)
	}

	fmt.Printf("Package %s removed from workspace.\n", packageName)

	return nil
}

func handlePackageList() error {
	if err := InitClient(); err != nil {
		return err
	}

	ctx := context.Background()
	workspace, err := WsClient.Get(ctx)
	if err != nil {
		return err
	}

	fmt.Println("Packages in workspace spec:")
	if len(workspace.Spec.Packages) == 0 {
		fmt.Println("  No packages specified")
	} else {
		tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(tw, "NAME\tCHANNEL\tCOMMIT")
		for _, pkg := range workspace.Spec.Packages {
			fmt.Fprintf(tw, "%s\t%s\t%s\n", pkg.Name, pkg.Channel, pkg.NixpkgsCommit)
		}
		tw.Flush()
	}

	fmt.Println("\nInstalled packages:")
	if len(workspace.Status.InstalledPackages) == 0 {
		fmt.Println("  No packages installed yet")
	} else {
		tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(tw, "NAME\tVERSION\tBIN PATH\tINSTALLED AT")
		for _, pkg := range workspace.Status.InstalledPackages {
			installedAt := ""
			if !pkg.InstalledAt.IsZero() {
				installedAt = pkg.InstalledAt.Format("2006-01-02 15:04:05")
			}
			fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n", pkg.Name, pkg.Version, pkg.BinPath, installedAt)
		}
		tw.Flush()
	}

	if len(workspace.Status.FailedPackages) > 0 {
		fmt.Println("\nFailed packages:")
		for _, pkg := range workspace.Status.FailedPackages {
			fmt.Printf("  - %s\n", pkg)
		}
		if workspace.Status.PackageInstallationMessage != "" {
			fmt.Printf("\nError: %s\n", workspace.Status.PackageInstallationMessage)
		}
	}

	return nil
}
