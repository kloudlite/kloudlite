# syntax=docker/dockerfile:1

# Build arguments that need to be available globally
ARG BASE_IMAGE=kloudlite/workspace-base:latest
ARG VERSION=dev

# Final workspace image
FROM ${BASE_IMAGE}

USER root

# Copy pre-built kl binary (built outside Docker)
# Build the binary with: CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o bin/kl-linux cmd/kl/main.go
COPY ../../bin/kl-linux /usr/local/bin/kl
RUN chmod +x /usr/local/bin/kl

# Install system dependencies including shells first with apt cache mount
# Add fish-shell PPA for fish 3.4+ (required for starship compatibility with set -f)
RUN --mount=type=cache,target=/var/cache/apt,sharing=locked \
  --mount=type=cache,target=/var/lib/apt,sharing=locked \
  apt-get update && apt-get install -y software-properties-common \
  && apt-add-repository -y ppa:fish-shell/release-3 \
  && apt-get update && apt-get install -y \
  openssh-server \
  python3 \
  supervisor \
  bash \
  zsh \
  fish \
  curl \
  ca-certificates \
  gnupg \
  unzip \
  fontconfig \
  && rm -rf /var/lib/apt/lists/*

# Create kl user with home directory and bash as default shell
RUN useradd -m -s /bin/bash -u 1001 kl && \
  echo "kl ALL=(ALL) NOPASSWD:ALL" >> /etc/sudoers && \
  mkdir -p /workspace && \
  chown -R kl:kl /workspace && \
  usermod -p '*' kl

# Install Node.js 20.x (LTS) with apt cache mount
RUN --mount=type=cache,target=/var/cache/apt,sharing=locked \
  --mount=type=cache,target=/var/lib/apt,sharing=locked \
  curl -fsSL https://deb.nodesource.com/setup_20.x | bash - \
  && apt-get install -y nodejs \
  && rm -rf /var/lib/apt/lists/*

# Configure SSH server and fix TERM handling
RUN mkdir -p /run/sshd && \
  sed -i 's/#PermitRootLogin prohibit-password/PermitRootLogin no/' /etc/ssh/sshd_config && \
  sed -i 's/#PasswordAuthentication yes/PasswordAuthentication no/' /etc/ssh/sshd_config && \
  sed -i 's/#PubkeyAuthentication yes/PubkeyAuthentication yes/' /etc/ssh/sshd_config && \
  echo 'AcceptEnv LANG LC_*' >> /etc/ssh/sshd_config && \
  echo 'SetEnv TERM=xterm-256color' >> /etc/ssh/sshd_config && \
  echo 'AuthorizedKeysFile /etc/ssh/kl-authorized-keys/authorized_keys' >> /etc/ssh/sshd_config

# Install code-server with cache mount
RUN --mount=type=cache,target=/tmp/code-server-cache \
  curl -fsSL https://code-server.dev/install.sh | sh

# Install ttyd binary (detect architecture and download appropriate version)
# Use TARGETARCH which is automatically set by Docker Buildx for multi-platform builds
ARG TARGETARCH
RUN --mount=type=cache,target=/tmp/ttyd-cache \
  if [ "$TARGETARCH" = "amd64" ]; then \
  curl -fsSL https://github.com/tsl0922/ttyd/releases/download/1.7.4/ttyd.x86_64 -o /usr/local/bin/ttyd; \
  elif [ "$TARGETARCH" = "arm64" ]; then \
  curl -fsSL https://github.com/tsl0922/ttyd/releases/download/1.7.4/ttyd.aarch64 -o /usr/local/bin/ttyd; \
  else \
  echo "ERROR: Unsupported architecture: $TARGETARCH" && exit 1; \
  fi && \
  chmod +x /usr/local/bin/ttyd && \
  /usr/local/bin/ttyd --version

# Install Nerd Fonts (JetBrainsMono and FiraCode) with persistent cache
RUN --mount=type=cache,target=/tmp/fonts-cache \
  mkdir -p /usr/share/fonts/truetype/nerd-fonts && \
  cd /tmp/fonts-cache && \
  if [ ! -f JetBrainsMono.zip ]; then \
  curl -fsSL https://github.com/ryanoasis/nerd-fonts/releases/download/v3.1.1/JetBrainsMono.zip -o JetBrainsMono.zip; \
  fi && \
  if [ ! -f FiraCode.zip ]; then \
  curl -fsSL https://github.com/ryanoasis/nerd-fonts/releases/download/v3.1.1/FiraCode.zip -o FiraCode.zip; \
  fi && \
  unzip -qo JetBrainsMono.zip -d /usr/share/fonts/truetype/nerd-fonts/ && \
  unzip -qo FiraCode.zip -d /usr/share/fonts/truetype/nerd-fonts/ && \
  fc-cache -f -v

# Install starship prompt with cache mount
RUN --mount=type=cache,target=/tmp/starship-cache \
  curl -sS https://starship.rs/install.sh | sh -s -- -y

USER root

# Create kloudlite config directory with shell initialization scripts
RUN mkdir -p /etc/kloudlite
COPY workspace-images/comprehensive/starship.toml /etc/kloudlite/starship.toml
COPY workspace-images/comprehensive/kl-shell-init.sh /etc/kloudlite/shell-init.sh
COPY workspace-images/comprehensive/kl-shell-init.zsh /etc/kloudlite/shell-init.zsh
COPY workspace-images/comprehensive/kl-shell-init.fish /etc/kloudlite/shell-init.fish
COPY workspace-images/comprehensive/kloudlite-context.sh /etc/kloudlite/kloudlite-context.sh
COPY workspace-images/comprehensive/init-shell-configs.sh /etc/kloudlite/init-shell-configs.sh
RUN chmod +x /etc/kloudlite/kloudlite-context.sh /etc/kloudlite/init-shell-configs.sh

# Disable Ubuntu's default MOTD scripts and install custom Kloudlite MOTD
RUN chmod -x /etc/update-motd.d/* && \
  rm -f /etc/legal /etc/update-motd.d/60-unminimize
COPY workspace-images/comprehensive/kloudlite-motd.sh /etc/update-motd.d/00-kloudlite-motd
RUN chmod +x /etc/update-motd.d/00-kloudlite-motd

# Note: Shell configurations (.bashrc, .zshrc, config.fish) are managed by
# /etc/kloudlite/init-shell-configs.sh which runs on container startup via supervisord
# This allows configurations to persist even when home directory is mounted from host

# Create supervisor configuration
RUN mkdir -p /etc/supervisor/conf.d
COPY workspace-images/comprehensive/supervisord.conf /etc/supervisor/conf.d/supervisord.conf

# Expose all required ports
EXPOSE 22 8080 7681 7682 7683 7684

# Start supervisor to manage all services
CMD ["/usr/bin/supervisord", "-c", "/etc/supervisor/conf.d/supervisord.conf"]
