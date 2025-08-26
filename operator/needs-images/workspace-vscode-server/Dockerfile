FROM ubuntu:24.04
RUN apt-get update && apt-get install -y git sudo zsh xz-utils curl

WORKDIR /vscode
# [Read More](https://code.visualstudio.com/docs/remote/tunnels)
RUN curl -Lk 'https://code.visualstudio.com/sha/download?build=stable&os=cli-alpine-x64' --output vscode_cli.tar.gz && tar -xf vscode_cli.tar.gz && mv code /usr/local/bin

# Create a non-root user 'kl'
RUN userdel ubuntu || echo "ubuntu user already deleted"
RUN groupadd -g 1000 kl && useradd -u 1000 -g 1000 -m kl && usermod -aG sudo kl && chsh -s $(which zsh) kl
RUN echo "kl ALL=(ALL) NOPASSWD: ALL" > /etc/sudoers.d/kl

COPY ./entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh
USER kl
WORKDIR /home/kl
CMD ["/entrypoint.sh"]
