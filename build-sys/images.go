package main

import (
	"context"
	"dagger/kloudlite/internal/dagger"
)

// +---------------------------------------+
// |  Docker Image Builds                  |
// +---------------------------------------+

// ImageAPIServer builds the api-server Docker image.
func (m *Kloudlite) ImageAPIServer(
	ctx context.Context,
	// +defaultPath="/"
	source *dagger.Directory,
	// +default="dev-local"
	tag string,
	// +default="ghcr.io/kloudlite"
	registry string,
) *dagger.Container {
	bin := m.BuildKloudlite(ctx, source, "linux", "amd64", "")
	return dag.Container().
		From("gcr.io/distroless/static-debian12:nonroot").
		WithFile("/app/api-server", bin).
		WithEntrypoint([]string{"/app/api-server"}).
		WithDefaultArgs([]string{"server", "api"})
}

// ImageWorkmachineManager builds the workmachine-manager Docker image.
func (m *Kloudlite) ImageWorkmachineManager(
	ctx context.Context,
	// +defaultPath="/"
	source *dagger.Directory,
	// +default="dev-local"
	tag string,
	// +default="ghcr.io/kloudlite"
	registry string,
) *dagger.Container {
	bin := m.BuildKloudlite(ctx, source, "linux", "amd64", "")
	aptCache := dag.CacheVolume("apt-debian-v1")
	return dag.Container().
		From("debian:bookworm-slim").
		WithMountedCache("/var/cache/apt", aptCache).
		WithExec([]string{"apt-get", "update"}).
		WithExec([]string{"apt-get", "install", "-y", "--no-install-recommends",
			"bash", "btrfs-progs", "ca-certificates", "coreutils", "curl", "git", "tar", "util-linux", "xz-utils",
			"wireguard-tools", "iptables", "iproute2"}).
		WithExec([]string{"rm", "-rf", "/var/lib/apt/lists/*"}).
		WithExec([]string{"mkdir", "-m", "0755", "/nix"}).
		WithExec([]string{"groupadd", "-g", "30000", "nixbld"}).
		WithExec([]string{"sh", "-c",
			`for i in $(seq 1 10); do useradd -c "Nix build user $i" -d /var/empty -g nixbld -G nixbld -M -N -r -s "$(command -v nologin)" "nixbld$i"; done`}).
		WithExec([]string{"sh", "-c", "curl -L https://nixos.org/nix/install | sh -s -- --no-daemon"}).
		WithEnvVariable("PATH", "/root/.nix-profile/bin:${PATH}", dagger.ContainerWithEnvVariableOpts{Expand: true}).
		WithFile("/app/workmachine-manager", bin).
		WithEntrypoint([]string{"/app/workmachine-manager"}).
		WithDefaultArgs([]string{"server", "workmachine-manager"})
}

// ImageWorkspaceBase builds the workspace-base image.
func (m *Kloudlite) ImageWorkspaceBase(
	ctx context.Context,
	// +defaultPath="/"
	source *dagger.Directory,
	// +default="dev-local"
	tag string,
	// +default="ghcr.io/kloudlite"
	registry string,
) *dagger.Container {
	return dag.Container().
		From("ubuntu:22.04").
		WithEnvVariable("DEBIAN_FRONTEND", "noninteractive").
		WithMountedCache("/var/cache/apt", dag.CacheVolume("apt-ubuntu-v1")).
		WithExec([]string{"apt-get", "update"}).
		WithExec([]string{"apt-get", "install", "-y", "curl", "git", "sudo", "ca-certificates"}).
		WithExec([]string{"rm", "-rf", "/var/lib/apt/lists/*"}).
		WithExec([]string{"useradd", "-m", "-s", "/bin/bash", "-u", "1000", "workspace"}).
		WithExec([]string{"sh", "-c", `echo "workspace ALL=(ALL) NOPASSWD:ALL" >> /etc/sudoers`}).
		WithWorkdir("/workspace").
		WithDefaultArgs([]string{"/bin/bash"})
}

// ImageWorkspaceComprehensive builds the workspace-comprehensive image.
// Depends on workspace-base — pass baseImage or it will be built automatically.
func (m *Kloudlite) ImageWorkspaceComprehensive(
	ctx context.Context,
	// +defaultPath="/"
	source *dagger.Directory,
	// +default="dev-local"
	tag string,
	// +default="ghcr.io/kloudlite"
	registry string,
	// +optional
	baseImage *dagger.Container,
) *dagger.Container {
	kl := m.BuildKlLinux(ctx, source)
	gw := m.buildGoBinary(ctx, source, "ghostty-web-server", "linux", "amd64", "", false)
	ctr := baseImage
	if ctr == nil {
		ctr = m.ImageWorkspaceBase(ctx, source, tag, registry)
	}
	ctr = ctr.
		WithExec([]string{"apt-get", "update"}).
		WithExec([]string{"apt-get", "install", "-y", "software-properties-common"}).
		WithExec([]string{"sh", "-c", "apt-add-repository -y ppa:fish-shell/release-3"}).
		WithExec([]string{"apt-get", "update"}).
		WithExec([]string{"apt-get", "install", "-y",
			"openssh-server", "python3", "supervisor",
			"bash", "bash-completion", "zsh", "fish",
			"curl", "ca-certificates", "gnupg", "unzip", "fontconfig"}).
		WithExec([]string{"rm", "-rf", "/var/lib/apt/lists/*"}).
		WithFile("/usr/local/bin/kl", kl).
		WithExec([]string{"chmod", "+x", "/usr/local/bin/kl"}).
		WithFile("/usr/local/bin/ghostty-web", gw).
		WithExec([]string{"chmod", "+x", "/usr/local/bin/ghostty-web"}).
		WithExec([]string{"npm", "install", "-g", "ghostty-web", "--prefix", "/home/kl/.local"}).
		WithExec([]string{"mkdir", "-p", "/usr/local/share/ghostty-web"}).
		WithExec([]string{"sh", "-c", "cp /home/kl/.local/lib/node_modules/ghostty-web/dist/ghostty-web.js /usr/local/share/ghostty-web/ && cp /home/kl/.local/lib/node_modules/ghostty-web/ghostty-vt.wasm /usr/local/share/ghostty-web/"}).
		WithExec([]string{"useradd", "-m", "-s", "/bin/bash", "-u", "1001", "kl"}).
		WithExec([]string{"sh", "-c", `echo "kl ALL=(ALL) NOPASSWD:ALL" >> /etc/sudoers`}).
		WithExec([]string{"mkdir", "-p", "/workspace"}).
		WithExec([]string{"chown", "-R", "kl:kl", "/workspace"}).
		WithEnvVariable("NPM_CONFIG_PREFIX", "/home/kl/.local").
		WithExec([]string{"sh", "-c", "curl -fsSL https://deb.nodesource.com/setup_20.x | bash -"}).
		WithExec([]string{"apt-get", "install", "-y", "nodejs"}).
		WithExec([]string{"rm", "-rf", "/var/lib/apt/lists/*"}).
		WithExec([]string{"mkdir", "-p", "/run/sshd"}).
		WithExec([]string{"sed", "-i", "s/#PermitRootLogin prohibit-password/PermitRootLogin no/", "/etc/ssh/sshd_config"}).
		WithExec([]string{"sed", "-i", "s/#PasswordAuthentication yes/PasswordAuthentication no/", "/etc/ssh/sshd_config"}).
		WithExec([]string{"sed", "-i", "s/#PubkeyAuthentication yes/PubkeyAuthentication yes/", "/etc/ssh/sshd_config"}).
		WithExec([]string{"sh", "-c", `echo 'AcceptEnv LANG LC_*' >> /etc/ssh/sshd_config`}).
		WithExec([]string{"sh", "-c", `echo 'SetEnv TERM=xterm-256color' >> /etc/ssh/sshd_config`}).
		WithExec([]string{"sh", "-c", `echo 'AuthorizedKeysFile /etc/ssh/kl-authorized-keys/authorized_keys' >> /etc/ssh/sshd_config`}).
		WithExec([]string{"sh", "-c", "curl -fsSL https://code-server.dev/install.sh | sh"}).
		WithExec([]string{"sh", "-c",
			"curl -fsSL https://github.com/tsl0922/ttyd/releases/download/1.7.4/ttyd.x86_64 -o /usr/local/bin/ttyd && chmod +x /usr/local/bin/ttyd"}).
		WithExec([]string{"sh", "-c", "curl -sS https://starship.rs/install.sh | sh -s -- -y"}).
		WithExec([]string{"mkdir", "-p", "/usr/share/fonts/truetype/nerd-fonts"}).
		WithExec([]string{"sh", "-c",
			"curl -fsSL https://github.com/ryanoasis/nerd-fonts/releases/download/v3.1.1/JetBrainsMono.zip -o /tmp/JetBrainsMono.zip && " +
				"unzip -qo /tmp/JetBrainsMono.zip -d /usr/share/fonts/truetype/nerd-fonts/"}).
		WithExec([]string{"sh", "-c",
			"curl -fsSL https://github.com/ryanoasis/nerd-fonts/releases/download/v3.1.1/FiraCode.zip -o /tmp/FiraCode.zip && " +
				"unzip -qo /tmp/FiraCode.zip -d /usr/share/fonts/truetype/nerd-fonts/"}).
		WithExec([]string{"fc-cache", "-f", "-v"}).
		WithExec([]string{"sh", "-c",
			"curl -fsSL https://download.docker.com/linux/static/stable/x86_64/docker-27.4.1.tgz -o /tmp/docker.tgz && " +
				"tar -xzf /tmp/docker.tgz --strip-components=1 -C /usr/local/bin docker/docker && chmod +x /usr/local/bin/docker"}).
		WithExec([]string{"mkdir", "-p", "/usr/local/lib/docker/cli-plugins"}).
		WithExec([]string{"sh", "-c",
			"curl -fsSL https://github.com/docker/buildx/releases/download/v0.19.3/buildx-v0.19.3.linux-amd64 -o /usr/local/lib/docker/cli-plugins/docker-buildx && " +
				"chmod +x /usr/local/lib/docker/cli-plugins/docker-buildx"})

	compDir := source.Directory("api/workspace-images/comprehensive")
	ctr = ctr.
		WithDirectory("/etc/kloudlite", compDir).
		WithExec([]string{"sh", "-c", "chmod +x /etc/kloudlite/kloudlite-context.sh /etc/kloudlite/init-shell-configs.sh"}).
		WithExec([]string{"sh", "-c", "chmod -x /etc/update-motd.d/* 2>/dev/null; rm -f /etc/legal /etc/update-motd.d/60-unminimize"}).
		WithExec([]string{"sh", "-c", "cp /etc/kloudlite/kloudlite-motd.sh /etc/update-motd.d/00-kloudlite-motd && chmod +x /etc/update-motd.d/00-kloudlite-motd"}).
		WithExec([]string{"mkdir", "-p", "/etc/supervisor/conf.d"}).
		WithExec([]string{"sh", "-c", "cp /etc/kloudlite/supervisord.conf /etc/supervisor/conf.d/supervisord.conf"})

	return ctr.
		WithExposedPort(22).WithExposedPort(8080).
		WithExposedPort(7681).WithExposedPort(7682).WithExposedPort(7683).WithExposedPort(7684).
		WithDefaultArgs([]string{"/usr/bin/supervisord", "-c", "/etc/supervisor/conf.d/supervisord.conf"})
}

// ImageOciInstaller builds the OCI installer Docker image.
func (m *Kloudlite) ImageOciInstaller(
	ctx context.Context,
	// +defaultPath="/"
	source *dagger.Directory,
	// +default="dev-local"
	tag string,
	// +default="ghcr.io/kloudlite"
	registry string,
) *dagger.Container {
	bin := m.BuildOciInstaller(ctx, source, "linux", "amd64")
	return dag.Container().
		From("alpine:3.20").
		WithExec([]string{"apk", "add", "--no-cache", "ca-certificates"}).
		WithFile("/app/oci-installer", bin).
		WithExec([]string{"chmod", "+x", "/app/oci-installer"}).
		WithEntrypoint([]string{"/app/oci-installer"})
}

// ImageCodeAnalyzer builds the code-analyzer Docker image.
func (m *Kloudlite) ImageCodeAnalyzer(
	ctx context.Context,
	// +defaultPath="/"
	source *dagger.Directory,
	// +default="dev-local"
	tag string,
	// +default="ghcr.io/kloudlite"
	registry string,
) *dagger.Container {
	bin := m.BuildCodeAnalyzer(ctx, source, "linux", "amd64")
	return dag.Container().
		From("python:3.11-alpine").
		WithExec([]string{"apk", "--no-cache", "add", "ca-certificates", "git"}).
		WithExec([]string{"pip", "install", "--no-cache-dir", "semgrep"}).
		WithFile("/usr/local/bin/code-analyzer", bin).
		WithExec([]string{"chmod", "+x", "/usr/local/bin/code-analyzer"}).
		WithExec([]string{"mkdir", "-p", "/var/lib/kloudlite/code-analysis"}).
		WithExposedPort(8082).
		WithEntrypoint([]string{"/usr/local/bin/code-analyzer"})
}

// ImageK3sBackup builds the k3s-backup Docker image.
func (m *Kloudlite) ImageK3sBackup(
	ctx context.Context,
	// +default="dev-local"
	tag string,
	// +default="ghcr.io/kloudlite"
	registry string,
) *dagger.Container {
	return dag.Container().
		From("amazon/aws-cli:latest").
		WithExec([]string{"yum", "install", "-y", "gzip"})
}

// ImageNix builds the Nix helper Docker image.
func (m *Kloudlite) ImageNix(ctx context.Context) *dagger.Container {
	return dag.Container().
		From("debian:bookworm-slim").
		WithMountedCache("/var/cache/apt", dag.CacheVolume("apt-debian-v1")).
		WithExec([]string{"apt-get", "update"}).
		WithExec([]string{"apt-get", "install", "-y", "--no-install-recommends",
			"curl", "xz-utils", "ca-certificates", "git"}).
		WithExec([]string{"rm", "-rf", "/var/lib/apt/lists/*"}).
		WithExec([]string{"mkdir", "-m", "0755", "/nix"}).
		WithExec([]string{"groupadd", "-g", "30000", "nixbld"}).
		WithExec([]string{"sh", "-c",
			`for i in $(seq 1 10); do useradd -c "Nix build user $i" -d /var/empty -g nixbld -G nixbld -M -N -r -s "$(command -v nologin)" "nixbld$i"; done`}).
		WithExec([]string{"sh", "-c", "curl -L https://nixos.org/nix/install | sh -s -- --no-daemon"}).
		WithEnvVariable("PATH", "/root/.nix-profile/bin:${PATH}", dagger.ContainerWithEnvVariableOpts{Expand: true})
}

// ImageNixServe builds the nix-serve Docker image.
func (m *Kloudlite) ImageNixServe(ctx context.Context) *dagger.Container {
	return dag.Container().
		From("debian:bookworm-slim").
		WithMountedCache("/var/cache/apt", dag.CacheVolume("apt-debian-v1")).
		WithExec([]string{"apt-get", "update"}).
		WithExec([]string{"apt-get", "install", "-y", "curl", "xz-utils", "ca-certificates"}).
		WithExec([]string{"rm", "-rf", "/var/lib/apt/lists/*"}).
		WithExec([]string{"mkdir", "-m", "0755", "/nix"}).
		WithExec([]string{"groupadd", "-g", "30000", "nixbld"}).
		WithExec([]string{"sh", "-c",
			`for i in $(seq 1 10); do useradd -c "Nix build user $i" -d /var/empty -g nixbld -G nixbld -M -N -r -s "$(command -v nologin)" "nixbld$i"; done`}).
		WithExec([]string{"sh", "-c", "curl -L https://nixos.org/nix/install | sh -s -- --no-daemon"}).
		WithEnvVariable("PATH", "/root/.nix-profile/bin:${PATH}", dagger.ContainerWithEnvVariableOpts{Expand: true}).
		WithExec([]string{"sh", "-c", ". /root/.nix-profile/etc/profile.d/nix.sh && nix-env -iA nixpkgs.nix-serve"}).
		WithExec([]string{"mkdir", "-p", "/nix/store"}).
		WithExec([]string{"chmod", "755", "/nix/store"}).
		WithExposedPort(5000).
		WithDefaultArgs([]string{"/bin/bash", "-c",
			". /root/.nix-profile/etc/profile.d/nix.sh && nix-serve --listen 0.0.0.0:5000"})
}

// ImageKlTunProxy builds the kl-tun-proxy Docker image.
func (m *Kloudlite) ImageKlTunProxy(
	ctx context.Context,
	// +defaultPath="/"
	source *dagger.Directory,
	// +default="dev-local"
	tag string,
	// +default="ghcr.io/kloudlite"
	registry string,
) *dagger.Container {
	bin := m.BuildKlTunProxy(ctx, source, "linux", "amd64")
	return dag.Container().
		From("gcr.io/distroless/static-debian12:nonroot").
		WithFile("/usr/local/bin/kl-tun-proxy", bin).
		WithExposedPort(51820).
		WithEntrypoint([]string{"/usr/local/bin/kl-tun-proxy"})
}

// ImageTunnelServer builds the tunnel-server Docker image.
func (m *Kloudlite) ImageTunnelServer(
	ctx context.Context,
	// +defaultPath="/"
	source *dagger.Directory,
	// +default="dev-local"
	tag string,
	// +default="ghcr.io/kloudlite"
	registry string,
) *dagger.Container {
	bin := m.BuildKloudlite(ctx, source, "linux", "amd64", "")
	aptCache := dag.CacheVolume("apt-debian-v1")
	return dag.Container().
		From("debian:bookworm-slim").
		WithMountedCache("/var/cache/apt", aptCache).
		WithExec([]string{"apt-get", "update"}).
		WithExec([]string{"apt-get", "install", "-y", "--no-install-recommends",
			"ca-certificates", "wireguard-tools", "iptables", "iproute2", "bash"}).
		WithExec([]string{"rm", "-rf", "/var/lib/apt/lists/*"}).
		WithFile("/app/tunnel-server", bin).
		WithEntrypoint([]string{"/app/tunnel-server", "server", "tunnel"})
}
