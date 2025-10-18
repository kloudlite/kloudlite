FROM kloudlite/workspace-base:latest

USER root

# Install system dependencies including shells first
# Removed: python3-pip (users can install via Nix), zsh (keeping bash + fish only)
RUN apt-get update && apt-get install -y \
    openssh-server \
    python3 \
    supervisor \
    fish \
    curl \
    ca-certificates \
    gnupg \
    unzip \
    fontconfig \
    && rm -rf /var/lib/apt/lists/*

# Create kl user with home directory and fish as default shell
RUN useradd -m -s /usr/bin/fish -u 1001 kl && \
    echo "kl ALL=(ALL) NOPASSWD:ALL" >> /etc/sudoers && \
    mkdir -p /workspace && \
    chown -R kl:kl /workspace

# Install Node.js 20.x (LTS)
RUN curl -fsSL https://deb.nodesource.com/setup_20.x | bash - \
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

# Install code-server
RUN curl -fsSL https://code-server.dev/install.sh | sh

# Install ttyd binary
RUN curl -fsSL https://github.com/tsl0922/ttyd/releases/download/1.7.4/ttyd.aarch64 -o /usr/local/bin/ttyd && \
    chmod +x /usr/local/bin/ttyd

# Install Nerd Fonts (JetBrainsMono and FiraCode)
RUN mkdir -p /usr/share/fonts/truetype/nerd-fonts && \
    curl -fsSL https://github.com/ryanoasis/nerd-fonts/releases/download/v3.1.1/JetBrainsMono.zip -o /tmp/JetBrainsMono.zip && \
    curl -fsSL https://github.com/ryanoasis/nerd-fonts/releases/download/v3.1.1/FiraCode.zip -o /tmp/FiraCode.zip && \
    unzip -qo /tmp/JetBrainsMono.zip -d /usr/share/fonts/truetype/nerd-fonts/ && \
    unzip -qo /tmp/FiraCode.zip -d /usr/share/fonts/truetype/nerd-fonts/ && \
    rm -f /tmp/JetBrainsMono.zip /tmp/FiraCode.zip && \
    fc-cache -f -v

# Install starship prompt
RUN curl -sS https://starship.rs/install.sh | sh -s -- -y

USER root

# Create kloudlite config directory with shell initialization scripts
RUN mkdir -p /etc/kloudlite
COPY starship.toml /etc/kloudlite/starship.toml
COPY kl-shell-init.sh /etc/kloudlite/shell-init.sh
COPY kl-shell-init.fish /etc/kloudlite/shell-init.fish
COPY kloudlite-context.sh /etc/kloudlite/kloudlite-context.sh
RUN chmod +x /etc/kloudlite/kloudlite-context.sh

# Disable Ubuntu's default MOTD scripts and install custom Kloudlite MOTD
RUN chmod -x /etc/update-motd.d/* && \
    rm -f /etc/legal /etc/update-motd.d/60-unminimize
COPY kloudlite-motd.sh /etc/update-motd.d/00-kloudlite-motd
RUN chmod +x /etc/update-motd.d/00-kloudlite-motd

# Configure fish shell for kl user with starship
RUN mkdir -p /home/kl/.config/fish && \
    echo 'source /etc/kloudlite/shell-init.fish' > /home/kl/.config/fish/config.fish && \
    ln -sf /etc/kloudlite/starship.toml /home/kl/.config/starship.toml && \
    chown -R kl:kl /home/kl/.config

# Fix TERM for SSH sessions - set it before fish starts
RUN echo '# Fix TERM for SSH login' >> /home/kl/.bashrc && \
    echo 'case "$TERM" in' >> /home/kl/.bashrc && \
    echo '  xterm-kitty|xterm-256color|screen-256color)' >> /home/kl/.bashrc && \
    echo '    export TERM=xterm-256color' >> /home/kl/.bashrc && \
    echo '    ;;' >> /home/kl/.bashrc && \
    echo '  xterm*)' >> /home/kl/.bashrc && \
    echo '    export TERM=xterm' >> /home/kl/.bashrc && \
    echo '    ;;' >> /home/kl/.bashrc && \
    echo 'esac' >> /home/kl/.bashrc && \
    chown kl:kl /home/kl/.bashrc

# Create supervisor configuration
RUN mkdir -p /etc/supervisor/conf.d
COPY supervisord.conf /etc/supervisor/conf.d/supervisord.conf

# Expose all required ports
EXPOSE 22 8080 7681 7682

# Start supervisor to manage all services
CMD ["/usr/bin/supervisord", "-c", "/etc/supervisor/conf.d/supervisord.conf"]
