#syntax=docker/dockerfile:1.4
FROM alpine:3.16
RUN apk add bash curl gettext jq lz4 helm kubectl --repository=https://dl-cdn.alpinelinux.org/alpine/edge/community --no-cache 
RUN curl -L0 https://releases.hashicorp.com/terraform/1.5.7/terraform_1.5.7_linux_amd64.zip > tf.zip && unzip tf.zip && mv terraform /usr/local/bin && rm tf.zip
RUN adduser --disabled-password --home="/app" --uid 1717 nonroot
USER nonroot
WORKDIR /app
COPY --chown=nonroot ./terraform ./terraform
RUN mkdir -p infrastructure-templates
COPY --chown=nonroot ./infrastructure-templates ./infrastructure-templates
ENV TF_PLUGIN_CACHE_DIR="/app/.terraform.d/plugin-cache"
# COPY .terraform.d.zip /app/terraform.zip
RUN mkdir -p $TF_PLUGIN_CACHE_DIR
SHELL ["/bin/bash", "-c"]
RUN <<'EOF'
  for dir in $(ls -d ./infrastructure-templates/*); do
    pushd $dir
    terraform init -backend=false &
    popd
  done
  wait
  tdir=$(basename $(dirname $TF_PLUGIN_CACHE_DIR))
  # zip terraform.zip -r $tdir && rm -rf $tdir
  tar cf - $tdir | lz4 -v -5 > tf.lz4 && rm -rf $tdir
EOF
# RUN bash ./script.sh
# RUN mkdir build-scripts
# COPY .ci/scripts/terraform-module-cache.sh ./build-scripts/terraform-module-cache.sh
# RUN bash build-scripts/terraform-module-cache.sh ./infrastructure-templates
ENV DECOMPRESS_CMD="lz4 -d tf.lz4 | tar xf -"
# ENV TERRAFORM_ZIPFILE="/app/terraform.zip"
ENV TEMPLATES_DIR="/app/infrastructure-templates"
