# Dockerfile for Nix installation image
# Used only by setup-nix init container to copy Nix to shared volume
FROM debian:bookworm-slim

# Install dependencies including nix with apt cache mount
RUN --mount=type=cache,target=/var/cache/apt,sharing=locked \
    --mount=type=cache,target=/var/lib/apt,sharing=locked \
    apt-get update && apt-get install -y \
    curl \
    xz-utils \
    ca-certificates \
    git \
    && rm -rf /var/lib/apt/lists/*

# Create /nix directory for Nix installation
RUN mkdir -m 0755 /nix && chown root /nix

# Create nixbld group and build users
RUN groupadd -g 30000 nixbld && \
    for i in $(seq 1 10); do \
        useradd -c "Nix build user $i" -d /var/empty -g nixbld -G nixbld -M -N -r -s "$(command -v nologin)" "nixbld$i"; \
    done

# Install Nix in single-user mode with cache mount for downloads
RUN --mount=type=cache,target=/tmp/nix-install-cache \
    curl -L https://nixos.org/nix/install | sh -s -- --no-daemon

# Source nix profile
ENV PATH="/root/.nix-profile/bin:${PATH}"

# No CMD - this image is used only for copying Nix to shared volumes
