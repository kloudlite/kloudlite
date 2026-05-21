FROM debian:bookworm-slim

RUN --mount=type=cache,target=/var/cache/apt,sharing=locked \
    --mount=type=cache,target=/var/lib/apt,sharing=locked \
    apt-get update && \
    apt-get install -y --no-install-recommends \
      bash \
      btrfs-progs \
      ca-certificates \
      coreutils \
      curl \
      git \
      tar \
      util-linux \
      xz-utils && \
    rm -rf /var/lib/apt/lists/*

RUN mkdir -m 0755 /nix && \
    groupadd -g 30000 nixbld && \
    for i in $(seq 1 10); do \
      useradd -c "Nix build user $i" -d /var/empty -g nixbld -G nixbld -M -N -r -s "$(command -v nologin)" "nixbld$i"; \
    done && \
    curl -L https://nixos.org/nix/install | sh -s -- --no-daemon

ENV PATH="/root/.nix-profile/bin:${PATH}"

COPY bin/workmachine-manager /app/workmachine-manager

ENTRYPOINT ["/app/workmachine-manager"]
CMD ["server", "workmachine-manager"]
