FROM ubuntu:22.04

# Avoid prompts from apt
ENV DEBIAN_FRONTEND=noninteractive

# Install basic dependencies
# Note: build-essential removed - users can install via Nix if needed for specific projects
RUN apt-get update && apt-get install -y \
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
