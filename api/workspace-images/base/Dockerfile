# syntax=docker/dockerfile:1

FROM ubuntu:22.04

# Avoid prompts from apt
ENV DEBIAN_FRONTEND=noninteractive

# Install basic dependencies with apt cache mount
# Note: build-essential removed - users can install via Nix if needed for specific projects
RUN --mount=type=cache,target=/var/cache/apt,sharing=locked \
    --mount=type=cache,target=/var/lib/apt,sharing=locked \
    apt-get update && apt-get install -y \
    curl \
    git \
    sudo \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

# Create workspace user
RUN useradd -m -s /bin/bash -u 1000 workspace && \
    echo "workspace ALL=(ALL) NOPASSWD:ALL" >> /etc/sudoers

USER workspace
WORKDIR /home/workspace

# Set working directory to workspace mount point
WORKDIR /workspace

CMD ["/bin/bash"]
