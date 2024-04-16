#syntax=docker/dockerfile:1.4
FROM alpine:3.16
RUN apk add bash curl gettext jq lz4 helm kubectl zstd --repository=https://dl-cdn.alpinelinux.org/alpine/edge/community --no-cache 
RUN curl -L0 https://releases.hashicorp.com/terraform/1.5.7/terraform_1.5.7_linux_amd64.zip > tf.zip && unzip tf.zip && mv terraform /usr/local/bin && rm tf.zip
WORKDIR /app
COPY --chown=nonroot ./terraform ./terraform
RUN mkdir -p infrastructure-templates
COPY --chown=nonroot ./infrastructure-templates ./infrastructure-templates
ENV TF_PLUGIN_CACHE_DIR="/app/.terraform.d/plugin-cache"
# COPY .terraform.d.zip /app/terraform.zip
RUN mkdir -p $TF_PLUGIN_CACHE_DIR
SHELL ["/bin/bash", "-c"]
RUN <<'EOF'
  for dir in $(ls -d ./infrastructure-templates/{aws,gcp}/*); do
    pushd $dir
    terraform init -backend=false &
    popd
  done

  wait

  tdir=$(basename $(dirname $TF_PLUGIN_CACHE_DIR))
  tar cf - $tdir | zstd --compress > tf.zst && rm -rf $tdir
EOF

# ENV DECOMPRESS_CMD="lz4 -d tf.lz4 | tar xf -"
ENV DECOMPRESS_CMD="zstd --decompress tf.zst --stdout | tar xf -"
ENV TEMPLATES_DIR="/app/infrastructure-templates"
#
WORKDIR /app
ENV TF_PLUGIN_CACHE_DIR="/app/.terraform.d/plugin-cache"
RUN mkdir -p $TF_PLUGIN_CACHE_DIR
RUN cat > script.sh <<EOF
COPY --chown=nonroot ./terraform ./terraform
COPY --chown=nonroot ./infrastructure-templates ./infrastructure-templates
RUN --mount=type=cache,id=sample,target=$TF_PLUGIN_CACHE_DIR <<EOF
  #!/usr/bin/env bash
  echo "hi" >> log.file
  # ls -d ./infrastructure-templates/{gcp,aws}/* | tee log.file | xargs -I{} bash -c "echo name is {}; $(terraform init chdir={} -backend=false &)"
  for dir in $(ls -d ./infrastructure-templates/{gcp,aws}/*); do
    echo $dir >> log.file
    pushd $dir
    terraform init -backend=false &
    popd
  done

  wait

  tdir=$(basename $(dirname $TF_PLUGIN_CACHE_DIR))
  echo "hello" >> log.file
  # tar cf - $tdir | lz4 -v -5 > tf.lz4 && rm -rf $tdir
  tar cf - $tdir | zstd --compress > tf.zst && rm -rf $tdir
EOF
RUN adduser --disabled-password --home="/app" --uid 1717 nonroot
USER nonroot
# ENV DECOMPRESS_CMD="lz4 -d tf.lz4 | tar xf -"
ENV DECOMPRESS_CMD="zstd --decompress tf.zst --stdout | tar xf -"
ENV TEMPLATES_DIR="/app/infrastructure-templates"
