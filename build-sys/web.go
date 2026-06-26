package main

import (
	"context"
	"dagger/kloudlite/internal/dagger"
	"fmt"
)

// +---------------------------------------+
// |  Web App Builds                       |
// +---------------------------------------+

// buildWebApp builds a Next.js web app and returns the built .next directory.
func (m *Kloudlite) buildWebApp(
	ctx context.Context,
	source *dagger.Directory,
	appName, envFile string,
) *dagger.Directory {
	webDir := source.Directory("web")
	bunCache := dag.CacheVolume("bun-install-v1")
	return dag.Container().
		From("oven/bun:alpine").
		WithMountedDirectory("/web", webDir).
		WithMountedCache("/web/node_modules", bunCache).
		WithWorkdir("/web").
		WithExec([]string{"bun", "install"}).
		WithExec([]string{"sh", "-c", fmt.Sprintf("cat > apps/%s/.env.local << 'EOF'\n%s\nEOF", appName, envFile)}).
		WithEnvVariable("NODE_ENV", "production").
		WithExec([]string{"bun", "run", "--filter", appName, "build"}).
		Directory(fmt.Sprintf("/web/apps/%s/.next", appName))
}

// BuildConsole builds the console web app (returns .next directory).
func (m *Kloudlite) BuildConsole(
	ctx context.Context,
	// +defaultPath="/"
	source *dagger.Directory,
) *dagger.Directory {
	return m.buildWebApp(ctx, source, "console",
		"NEXT_PUBLIC_BASE_URL=https://console.kloudlite.io\n"+
			"NEXT_PUBLIC_CONSOLE_URL=https://console.kloudlite.io\n"+
			"NEXT_PUBLIC_INSTALLATION_DOMAIN=khost.dev\n"+
			"NEXT_PUBLIC_TURNSTILE_SITE_KEY=0x4AAAAAACVzkPXwxkSHKMwt\n"+
			"NEXT_PUBLIC_SUPABASE_URL=https://vsrhxnpsxpqlyyetvrhy.supabase.co\n"+
			"NEXT_PUBLIC_SUPABASE_ANON_KEY=placeholder\n")
}

// BuildDashboard builds the dashboard web app (returns .next directory).
func (m *Kloudlite) BuildDashboard(
	ctx context.Context,
	// +defaultPath="/"
	source *dagger.Directory,
) *dagger.Directory {
	return m.buildWebApp(ctx, source, "dashboard",
		"NEXT_PUBLIC_BASE_URL=https://dashboard.kloudlite.io\n"+
			"NEXT_PUBLIC_SUPABASE_URL=https://vsrhxnpsxpqlyyetvrhy.supabase.co\n"+
			"NEXT_PUBLIC_SUPABASE_ANON_KEY=placeholder\n"+
			"NEXT_PUBLIC_GITHUB_CLIENT_ID=placeholder\n"+
			"NEXT_PUBLIC_GOOGLE_CLIENT_ID=placeholder\n"+
			"NEXT_PUBLIC_MICROSOFT_ENTRA_CLIENT_ID=placeholder\n"+
			"NEXT_PUBLIC_MICROSOFT_ENTRA_TENANT_ID=placeholder\n")
}

// BuildWebsite builds the website web app (returns .next directory).
func (m *Kloudlite) BuildWebsite(
	ctx context.Context,
	// +defaultPath="/"
	source *dagger.Directory,
) *dagger.Directory {
	return m.buildWebApp(ctx, source, "website",
		"NEXT_PUBLIC_BASE_URL=https://kloudlite.io\n")
}

// buildWebRuntimeImage creates the runtime container for a web app from its .next build output.
func (m *Kloudlite) buildWebRuntimeImage(nextBuild *dagger.Directory, appName string, source *dagger.Directory) *dagger.Container {
	return dag.Container().
		From("node:23-alpine").
		WithExec([]string{"apk", "add", "--no-cache", "--no-scripts", "openssl"}).
		WithExec([]string{"addgroup", "--system", "--gid", "1001", "nodejs"}).
		WithExec([]string{"adduser", "--system", "--uid", "1001", "nextjs"}).
		WithDirectory("/app", nextBuild.Directory("standalone")).
		WithDirectory(fmt.Sprintf("/app/apps/%s/.next/static", appName), nextBuild.Directory("static")).
		WithDirectory(fmt.Sprintf("/app/apps/%s/public", appName), source.Directory(fmt.Sprintf("web/apps/%s/public", appName))).
		WithExposedPort(3000).
		WithEnvVariable("PORT", "3000").
		WithEnvVariable("HOSTNAME", "0.0.0.0").
		WithEnvVariable("NODE_ENV", "production").
		WithWorkdir("/app")
}

// ImageConsole builds the console Docker runtime image.
func (m *Kloudlite) ImageConsole(
	ctx context.Context,
	// +defaultPath="/"
	source *dagger.Directory,
	// +default="dev-local"
	tag string,
	// +default="ghcr.io/kloudlite/kloudlite"
	registry string,
	// +optional
	nextBuild *dagger.Directory,
) *dagger.Container {
	if nextBuild == nil {
		nextBuild = m.BuildConsole(ctx, source)
	}
	return m.buildWebRuntimeImage(nextBuild, "console", source).
		WithExec([]string{"chown", "-R", "nextjs:nodejs", "/app"}).
		WithUser("nextjs").
		WithDefaultArgs([]string{"node", "apps/console/server.js"})
}

// ImageDashboard builds the dashboard Docker runtime image.
func (m *Kloudlite) ImageDashboard(
	ctx context.Context,
	// +defaultPath="/"
	source *dagger.Directory,
	// +default="dev-local"
	tag string,
	// +default="ghcr.io/kloudlite/kloudlite"
	registry string,
	// +optional
	nextBuild *dagger.Directory,
) *dagger.Container {
	if nextBuild == nil {
		nextBuild = m.BuildDashboard(ctx, source)
	}
	ctr := m.buildWebRuntimeImage(nextBuild, "dashboard", source).
		WithFile("/app/apps/dashboard/proxy-server.js",
			source.File("web/apps/dashboard/proxy-server.js")).
		WithFile("/app/apps/dashboard/start.sh",
			source.File("web/apps/dashboard/start.sh")).
		WithExec([]string{"chmod", "+x", "/app/apps/dashboard/start.sh"})
	ctr = ctr.
		WithExec([]string{"mkdir", "-p", "/tmp/ws-install"}).
		WithWorkdir("/tmp/ws-install").
		WithExec([]string{"npm", "install", "ws"}).
		WithExec([]string{"cp", "-r", "node_modules/ws", "/app/apps/dashboard/node_modules/"}).
		WithExec([]string{"rm", "-rf", "/tmp/ws-install"}).
		WithWorkdir("/app")
	return ctr
}

// ImageWebsite builds the website Docker runtime image.
func (m *Kloudlite) ImageWebsite(
	ctx context.Context,
	// +defaultPath="/"
	source *dagger.Directory,
	// +default="dev-local"
	tag string,
	// +default="ghcr.io/kloudlite/kloudlite"
	registry string,
	// +optional
	nextBuild *dagger.Directory,
) *dagger.Container {
	if nextBuild == nil {
		nextBuild = m.BuildWebsite(ctx, source)
	}
	return m.buildWebRuntimeImage(nextBuild, "website", source).
		WithExec([]string{"chown", "-R", "nextjs:nodejs", "/app"}).
		WithUser("nextjs").
		WithDefaultArgs([]string{"node", "apps/website/server.js"})
}
