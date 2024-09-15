# vim: set ft=dockerfile:
FROM nixos/nix:latest AS nix
WORKDIR /app
COPY . ./

RUN cat > /tmp/script.sh <<EOF
  nix build .#container -o result
  mkdir -p /tmp/nix-store-closure
  cp -R $(nix-store -qR result/) /tmp/nix-store-closure

  export TF_PLUGIN_CACHE_DIR="$PWD/.terraform.d/plugin-cache"

  for dir in $(ls -d ./infrastructure-templates/{gcp,aws}/*); do
    terraform -chdir=$dir init -backend=false -upgrade &
  done

  echo "compressing"
  tdir=$(basename $(dirname $TF_PLUGIN_CACHE_DIR))
  # tar cf - $tdir | zstd -12 --compress > tf.zst
  tar cf - $tdir | zstd --compress > tf.zst
EOF
RUN nix --experimental-features "nix-command flakes" develop --command bash /tmp/script.sh

FROM busybox:latest

RUN mkdir -p /etc/ssl/certs
COPY --from=nix /nix/var/nix/profiles/default/etc/ssl/certs/ca-bundle.crt /etc/ssl/certs/ca-certificates.crt

WORKDIR /app

# RUN --mount=type=bind,source=context.tar,target=context.tar \
#   tar xf context.tar && \
#   mkdir -p /nix && mv nixstore /nix/store && \
#   mkdir -p /usr/local/bin && mv result/bin/* /usr/local/bin/ && rm -rf result && \
#   mv tf.zst /app/tf.zst

RUN mkdir -p /nix
COPY --from=nix /tmp/nix-store-closure /nix/store
COPY --from=nix /tmp/tf.zst /app/tf.zst
COPY --from=nix /app/result/bin/* /usr/local/bin/

RUN adduser --disabled-password --home="/app" --uid 1717 nonroot
COPY --chown=nonroot ./terraform ./terraform
COPY --chown=nonroot ./infrastructure-templates ./infrastructure-templates
ENV TF_PLUGIN_CACHE_DIR="/app/.terraform.d/plugin-cache"
ENV DECOMPRESS_CMD="zstd --decompress tf.zst --stdout | tar xf -"
ENV TEMPLATES_DIR="/app/infrastructure-templates"
