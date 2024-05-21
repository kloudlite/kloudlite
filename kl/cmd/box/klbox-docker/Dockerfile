# Use the official Ubuntu base image
FROM nixos/nix:latest AS nix

FROM --platform=$TARGETPLATFORM golang:1.21 as builder

WORKDIR /app

COPY main.go main.go

RUN go mod init mounter.io

RUN go build -o mounter main.go

FROM --platform=$TARGETPLATFORM ubuntu:latest


RUN apt-get update && apt-get install -y openssh-server sudo curl xz-utils jq iproute2 libayatana-appindicator3-1 nano

RUN mkdir /var/run/sshd

RUN userdel ubuntu

RUN groupadd -g 1000 kl && useradd -u 1000 -g 1000 -m kl && usermod -aG sudo kl && chsh -s /bin/bash kl && \
  echo "kl ALL=(ALL) NOPASSWD: ALL" > /etc/sudoers.d/kl

# Optionally change the SSH port and other settings
#RUN sed -i 's/#PermitRootLogin prohibit-password/PermitRootLogin yes/' /etc/ssh/sshd_config
# SSH login fix. Otherwise user is kicked off after login
RUN sed -i 's@session\s*required\s*pam_loginuid.so@session optional pam_loginuid.so@g' /etc/pam.d/sshd

EXPOSE 22

ENV HOSTNAME box
RUN mkdir -m 0755 /nix && chown kl /nix
USER kl
RUN curl -L https://nixos.org/nix/install | sh -s -- --no-daemon --no-modify-profile

RUN curl -fsSL https://get.jetify.com/devbox > /tmp/devbox.sh && chmod +x /tmp/devbox.sh && sudo /tmp/devbox.sh -f && rm /tmp/devbox.sh && sudo chown kl:kl /usr/local/bin/devbox


USER root

COPY --from=nix /nix/var/nix/profiles/default/etc/ssl/certs/ca-bundle.crt /etc/ssl/certs/ca-certificates.crt

RUN mkdir -p /home/kl/.config/nix && echo "experimental-features = nix-command flakes" > /home/kl/.config/nix/nix.conf
RUN chown -R kl:kl /home/kl/.config

RUN cp /etc/ssl/certs/ca-certificates.crt /no-cert-file.crt

RUN chmod -x /etc/update-motd.d/*
RUN cat > /etc/motd <<'EOF'
Welcome to KloudLite's Development Environment
EOF

COPY .bashrc .profile /tmp/

COPY --from=builder /app/mounter /mounter


SHELL ["/bin/bash", "-c"]

COPY ./start.sh /start.sh
COPY ./entrypoint.sh /entrypoint.sh

RUN chmod +x /start.sh
RUN chmod +x /entrypoint.sh

USER kl
ENTRYPOINT ["/entrypoint.sh"]
