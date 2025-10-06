FROM debian:bookworm-slim

# Install dependencies
RUN apt-get update && apt-get install -y \
    curl \
    xz-utils \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

# Create /nix directory for Nix installation
RUN mkdir -m 0755 /nix && chown root /nix

# Create nixbld group and build users
RUN groupadd -g 30000 nixbld && \
    for i in $(seq 1 10); do \
        useradd -c "Nix build user $i" -d /var/empty -g nixbld -G nixbld -M -N -r -s "$(command -v nologin)" "nixbld$i"; \
    done

# Install Nix in single-user mode
RUN curl -L https://nixos.org/nix/install | sh -s -- --no-daemon

# Source nix profile
ENV PATH="/root/.nix-profile/bin:${PATH}"

# Install nix-serve
RUN . /root/.nix-profile/etc/profile.d/nix.sh && \
    nix-env -iA nixpkgs.nix-serve

# Create directory for nix store
RUN mkdir -p /nix/store && chmod 755 /nix/store

# Expose port
EXPOSE 5000

# Run nix-serve
CMD ["/bin/bash", "-c", ". /root/.nix-profile/etc/profile.d/nix.sh && nix-serve --listen 0.0.0.0:5000"]
