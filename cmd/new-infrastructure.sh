#! /usr/bin/env bash

dir=$1

script_dir=$(realpath $(dirname "$0"))

dir_path=$script_dir/../infrastructures/$dir

[ -d "$dir_path" ] && echo "Directory $dir_path already exists" && exit 1

infra_template=$(ls "$script_dir/../infrastructure-templates" | fzf)

mkdir -p "$dir_path"

pushd "$dir_path" > /dev/null 2>&1 || exit
mkdir -p .secrets

touch .secrets/env


cat > Taskfile.yml <<EOF
version: 3

dotenv:
  - .secrets/env

vars:
  Varsfile: ".secrets/varfile.json"

tasks:
  sync-from-template:
    vars:
      InfrastructureTemplate: ../../infrastructure-templates/$infra_template
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
    vars:
      PlanOutput: ".secrets/plan.out"
    cmds:
      - cat ./varfile.template.yml | envsubst | yq > {{.Varsfile}}
      - terraform plan --var-file "{{.Varsfile}}" --out "{{.PlanOutput}}"

  apply:
    dir: ./
    dotenv:
      - .secrets/env
    vars:
      PlanOutput: ".secrets/plan.out"
    cmds:
      - terraform apply "{{.PlanOutput}}"

  validate:
    dir: ./
    cmds:
      - terraform validate  -var-file={{.Varsfile}}

  destroy:
    dir: ./
    dotenv:
      - .secrets/env
    vars:
      PlanOutput: ".secrets/plan.destroy.out"
    cmds:
      - cat ./varfile.template.yml | envsubst | yq > {{.Varsfile}}
      - terraform plan --var-file={{.Varsfile}} --destroy --out "{{.PlanOutput}}"
      - terraform apply "{{.PlanOutput}}"
EOF

popd