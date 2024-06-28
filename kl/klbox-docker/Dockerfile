# syntax=docker/dockerfile:1.4
FROM --platform=$TARGETPLATFORM golang:1.21 as builder

COPY --from=project ./go.sum ./go.mod /kl-app/

WORKDIR /kl-app

RUN go mod download -x

COPY --from=project . /kl-app

ARG VERSION

RUN go build --tags=box -ldflags="-X github.com/kloudlite/kl/flags.Version=${VERSION} -X github.com/kloudlite/kl/flags.CliName=kl" -o ./bin/kl main.go


FROM --platform=$TARGETPLATFORM ubuntu:latest

RUN apt-get update && apt-get install -y openssh-server sudo curl xz-utils jq iproute2 nano zsh git dnsmasq bzip2

RUN mkdir /var/run/sshd

RUN userdel ubuntu || echo "ubuntu user already deleted"

RUN groupadd -g 1000 kl && useradd -u 1000 -g 1000 -m kl && usermod -aG sudo kl && chsh -s /bin/bash kl && \
  echo "kl ALL=(ALL) NOPASSWD: ALL" > /etc/sudoers.d/kl

# Optionally change the SSH port and other settings
#RUN sed -i 's/#PermitRootLogin prohibit-password/PermitRootLogin yes/' /etc/ssh/sshd_config
# SSH login fix. Otherwise user is kicked off after login
RUN sed -i 's@session\s*required\s*pam_loginuid.so@session optional pam_loginuid.so@g' /etc/pam.d/sshd

RUN echo "GatewayPorts yes" >> /etc/ssh/sshd_config


ENV HOSTNAME box
RUN mkdir -m 0755 /nix && chown kl /nix
RUN mkdir -p /kl-nix-profile && chown kl /kl-nix-profile

#RUN <<EOF
# export VERSION=2.23.0
# export SYSTEM=x86_64-linux
# cd /tmp
# curl -LO https://releases.nixos.org/nix/nix-$VERSION/nix-$VERSION-$SYSTEM.tar.xz
# tar xf nix-$VERSION-$SYSTEM.tar.xz
# mv nix-$VERSION-$SYSTEM /nix-installer
#EOF
#RUN chown kl /nix-installer

USER kl
RUN mkdir -p ~/.local/state/nix
RUN ln -sf /kl-nix-profile ~/.local/state/nix/profiles
RUN curl -L https://nixos.org/nix/install | sh -s --

RUN mkdir -p "/home/kl/.config/zsh"
RUN git clone https://github.com/sindresorhus/pure.git "/home/kl/.config/zsh/pure"

USER root


RUN git clone https://github.com/zsh-users/zsh-syntax-highlighting.git  "/usr/share/zsh/plugins/zsh-syntax-highlighting"

USER kl

USER root

# COPY --from=nix /nix/var/nix/profiles/default/etc/ssl/certs/ca-bundle.crt /etc/ssl/certs/ca-certificates.crt

# RUN mkdir -p /home/kl/.config/nix
# && echo "experimental-features = nix-command flakes" > /home/kl/.config/nix/nix.conf
# RUN chown -R kl:kl /home/kl/.config

RUN cp /etc/ssl/certs/ca-certificates.crt /no-cert-file.crt

RUN chmod -x /etc/update-motd.d/*
RUN cat > /etc/motd <<'EOF'
Welcome to KloudLite's Development Environment
EOF

COPY .bashrc .profile .zshrc aliasrc /tmp/

RUN chsh -s /bin/zsh kl

COPY ./start.sh /start.sh
COPY ./entrypoint.sh /entrypoint.sh


RUN mkdir /kl-tmp && chown -R kl:kl /kl-tmp

RUN chmod +x /start.sh /entrypoint.sh

COPY --from=builder /kl-app/bin/kl /usr/local/bin/kl

USER kl

ENTRYPOINT ["/entrypoint.sh"]
