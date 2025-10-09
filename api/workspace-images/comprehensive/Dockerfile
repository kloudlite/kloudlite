FROM kloudlite/workspace-base:latest

USER root

# Create kl user with home directory
RUN useradd -m -s /bin/bash -u 1001 kl && \
    echo "kl ALL=(ALL) NOPASSWD:ALL" >> /etc/sudoers

# Install system dependencies including shells
RUN apt-get update && apt-get install -y \
    openssh-server \
    python3 \
    python3-pip \
    supervisor \
    fish \
    zsh \
    curl \
    ca-certificates \
    gnupg \
    && rm -rf /var/lib/apt/lists/*

# Install Node.js 20.x (LTS)
RUN curl -fsSL https://deb.nodesource.com/setup_20.x | bash - \
    && apt-get install -y nodejs \
    && rm -rf /var/lib/apt/lists/*

# Configure SSH server
RUN mkdir -p /run/sshd && \
    sed -i 's/#PermitRootLogin prohibit-password/PermitRootLogin no/' /etc/ssh/sshd_config && \
    sed -i 's/#PasswordAuthentication yes/PasswordAuthentication no/' /etc/ssh/sshd_config && \
    sed -i 's/#PubkeyAuthentication yes/PubkeyAuthentication yes/' /etc/ssh/sshd_config

# Install code-server
RUN curl -fsSL https://code-server.dev/install.sh | sh

# Install ttyd binary
RUN curl -fsSL https://github.com/tsl0922/ttyd/releases/download/1.7.4/ttyd.aarch64 -o /usr/local/bin/ttyd && \
    chmod +x /usr/local/bin/ttyd

# Install Claude Code via npm
RUN npm install -g @anthropic-ai/claude-code

# Install VS Code CLI (for tunnels)
RUN curl -Lk 'https://code.visualstudio.com/sha/download?build=stable&os=cli-alpine-x64' --output vscode_cli.tar.gz && \
    tar -xf vscode_cli.tar.gz -C /usr/local/bin && \
    rm vscode_cli.tar.gz

# Install starship prompt
RUN curl -sS https://starship.rs/install.sh | sh -s -- -y

USER root

# Create kloudlite config directory with shell initialization scripts
RUN mkdir -p /etc/kloudlite
COPY starship.toml /etc/kloudlite/starship.toml
COPY kl-shell-init.sh /etc/kloudlite/shell-init.sh
COPY kl-shell-init.fish /etc/kloudlite/shell-init.fish

# Create supervisor configuration
RUN mkdir -p /etc/supervisor/conf.d
COPY supervisord.conf /etc/supervisor/conf.d/supervisord.conf

# Expose all required ports
EXPOSE 22 8080 7681 8000

# Start supervisor to manage all services
CMD ["/usr/bin/supervisord", "-c", "/etc/supervisor/conf.d/supervisord.conf"]
