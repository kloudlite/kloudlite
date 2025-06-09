#! /usr/bin/env bash

destination_path=$(realpath "$1")

SCRIPT_DIR=$(realpath $(dirname $0))

infra_template=$INFRA_TEMPLATE
if [ -z "$infra_template" ]; then
	templates_dir="$SCRIPT_DIR/../infrastructure-templates"
	infra_template=$(fd '' "$templates_dir" | fzf --prompt "Choose An Infrastructure template")
fi

[ -d "$destination_path" ] && echo "Directory $destination_path already exists" && exit 1

mkdir -p "$destination_path"

pushd "$destination_path" >/dev/null 2>&1 || exit
mkdir -p .secrets

touch .secrets/env

cat >Taskfile.yml <<EOF
version: 3

dotenv:
  - .secrets/env

vars:
  Varsfile: ".secrets/varfile.json"

  ApplyPlan: "./secrets/apply.plan"
  DestroyPlan: "./secrets/destroy.plan"

tasks:
  sync-from-template:
    vars:
      InfrastructureTemplate: $(realpath "${infra_template}" --relative-to="$destination_path")
    env:
      SHELL: bash
    silent: true
    cmds:
      - chmod -f 600 ./*.tf | true
      - cp {{.InfrastructureTemplate}}/*.tf ./
      - chmod 400 ./*.tf
      - echo "sync complete"

  init:
    cmds:
      - terraform init
    silent: true

  plan:
    dir: ./
    cmds:
      - cat ./varfile.template.yml | envsubst | yq > {{.Varsfile}}
      - terraform plan --var-file "{{.Varsfile}}" --out "{{.ApplyPlan}}"

  apply:
    dir: ./
    dotenv:
      - .secrets/env
    cmds:
      - terraform apply "{{.ApplyPlan}}"

  validate:
    dir: ./
    cmds:
      - terraform validate  -var-file={{.Varsfile}}

  destroy:plan:
    dir: ./
    dotenv:
      - .secrets/env
    cmds:
      - cat ./varfile.template.yml | envsubst | yq > {{.Varsfile}}
      - terraform plan --var-file={{.Varsfile}} --destroy --out "{{.DestroyPlan}}"

  destroy:apply:
    dir: ./
    dotenv:
      - .secrets/env
    cmds:
      - terraform apply "{{.DestroyPlan}}"
EOF

popd
