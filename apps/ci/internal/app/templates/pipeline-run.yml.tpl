{{- $svcAccountName := "kloudlite-svc-account" }}
{{/*{{- $dockerConfigName := "kloudlite-docker-registry" }}*/}}
{{- $dockerConfigName := "kloudlite-harbor-creds" }}

{{- $tolerationKey := "only-ci" }}
{{- $tolerationEffect := "NoSchedule" }}

{{- $varPipelineId := "pipeline-id" }}
{{- $varTaskNamespace := "task-namespace" }}
{{- $varGitRepo := "git-repo" }}
{{- $varGitUser := "git-user" }}
{{- $varGitPassword := "git-password" }}
{{- $varGitBranch := "git-branch" }}
{{- $varGitCommitHash := "git-commit-hash" }}

{{- $varIsDockerBuild := "is-docker-build" }}
{{- $varDockerContextDir := "docker-context-dir"}}
{{- $varDockerFile := "docker-file"}}
{{- $varDockerBuildArgs := "docker-build-args"}}

{{- $varBuildBaseImage := "build-base-image" }}
{{- $varBuildCmd := "build-cmd" }}
{{- $varBuildOutputDir := "build-output-dir" }}

{{- $varRunBaseImage := "run-base-image"}}
{{- $varRunCmd := "run-cmd"}}

{{- $varArtifactRefDockerImageName := "artifact_ref-docker_image_name"}}
{{- $varArtifactRefDockerImageTag := "artifact_ref-docker_image_tag"}}
{{/*{{- $varAppName := get . "app-name"  -}}*/}}

{{/*input Variables*/}}
{{- $pipelineRuns := get . "pipeline-runs" }}

{{- range $pRun := $pipelineRuns}}
{{- with $pRun }}
{{- /*gotype: kloudlite.io/apps/ci/internal/domain.TektonVars*/ -}}
{{""}}

---
apiVersion: tekton.dev/v1beta1
kind: PipelineRun
metadata:
  generateName: {{.PipelineRunId}}-
  namespace: {{.TaskNamespace}}
  labels:
    app: {{.PipelineId}}
    component: {{.PipelineRunId}}
spec:
  workspaces:
    - name: p-output
      emptyDir: {}

    - name: p-docker-config
      secret:
        secretName: {{$dockerConfigName}}
  serviceAccountName: {{$svcAccountName}}
  podTemplate:
    hostNetwork: true
    nodeSelector:
{{/*      kloudlite.io/auto-scaler: "true"*/}}
      kloudlite.io/region: "reg-20chh-dzgubg4o-nczwa93eciqbs"
    tolerations:
{{/*    - key: kloudlite.io/auto-scaler*/}}
{{/*      operator: Exists*/}}
{{/*      effect: NoExecute*/}}
      - effect: NoExecute
        key: kloudlite.io/region
        operator: Equal
        value: reg-20chh-dzgubg4o-nczwa93eciqbs
    volumes:
      - name: host-ssh-root
        hostPath:
          path: /root/.ssh

      - name: host-dir
        hostPath:
          path: /tmp/{{.PipelineRunId}}
          type: DirectoryOrCreate

  pipelineSpec:
    workspaces:
      - name: p-output
      - name: p-docker-config
    tasks:
      - name: build-and-push
        workspaces:
          - name: output
            workspace: p-output
          - name: docker-config
            workspace: p-docker-config
        params:
          - name: {{$varGitRepo}}
            value: {{.GitRepo}}

          - name: {{$varGitBranch}}
            value: {{.GitBranch}}

          - name: {{$varGitUser}}
            value: {{.GitUser}}

          - name: {{$varGitPassword}}
            value: {{.GitPassword}}

          - name: {{$varGitCommitHash}}
            value: {{.GitCommitHash}}

            {{/*          Docker configurations*/}}
          - name: {{$varIsDockerBuild}}
            value: {{.IsDockerBuild}}

          - name: {{$varDockerFile}}
            value: {{.DockerFile}}

          - name: {{$varDockerContextDir}}
            value: {{.DockerContextDir}}

          - name: {{$varDockerBuildArgs}}
            value: {{.DockerBuildArgs}}

            {{/*            Build Configurations*/}}
          - name: {{$varBuildBaseImage}}
            value: {{.BuildBaseImage}}

          - name: {{$varBuildCmd}}
            value: {{.BuildCmd}}

          - name: {{$varBuildOutputDir}}
            value: {{.BuildOutputDir}}

            {{/*          Run Configurations*/}}
          - name: {{$varRunBaseImage}}
            value: {{.RunBaseImage}}

          - name: {{$varRunCmd}}
            value: {{.RunCmd}}

            {{/*          Artifacts*/}}
          - name: {{$varArtifactRefDockerImageName}}
            value: {{.ArtifactDockerImageName}}

          - name: {{$varArtifactRefDockerImageTag}}
            value: {{.ArtifactDockerImageTag}}
        taskSpec:
{{/*          computeResources:*/}}
{{/*            requests:*/}}
{{/*              cpu: 100m*/}}
{{/*              memory: 400Mi*/}}
{{/*            limits:*/}}
{{/*              cpu: 200m*/}}
{{/*              memory: 400Mi*/}}

{{/*          sidecars:*/}}
{{/*            - name: dind*/}}
{{/*              image: docker:19.03.5-dind*/}}
{{/*              args:*/}}
{{/*                - dockerd*/}}
{{/*                - --host*/}}
{{/*                - tcp://127.0.0.1:2375*/}}
{{/*                - --max-concurrent-downloads*/}}
{{/*                - "1"*/}}
{{/*              securityContext:*/}}
{{/*                privileged: true*/}}
{{/*              volumeMounts:*/}}
{{/*                - name: $(workspaces.output.volume)*/}}
{{/*                  mountPath: /var/lib/docker*/}}
{{/*                  subPath: docker*/}}

          workspaces:
            - name: output
            - name: docker-config
          steps:
            - name: clone-git
              image: registry.kloudlite.io/public/git:latest
              script: |+
                cat > clone-git.sh <<'EOF'
                workdir="$(workspaces.output.path)"
                mkdir -p $workdir
                gitUser="$(params.{{$varGitUser}})"
                gitPassword="$(params.{{$varGitPassword}})"
                gitRepo="$(params.{{$varGitRepo}})"
                gitRepoUrl=$(echo $(params.{{$varGitRepo}}) | sed -E "s/https:[/]{2}(.*?\/)/https\:\/\/${gitUser}\:${gitPassword}@\1/g")
                gitBranch="$(params.{{$varGitBranch}})"

                echo "cloning git repo: $gitRepo"

                # if not using git submodules
                cd $workdir
{{/*                echo git clone --depth=1 --branch ${gitBranch} --single-branch "$gitRepo" "./repo"*/}}
                git clone --depth=1 --branch ${gitBranch} --single-branch "$gitRepoUrl" "./repo"
                ls -l repo
                echo "successfully cloned git repo: $gitRepo"
                echo "STEP (clone-git) FINISHED"
                EOF

                sh clone-git.sh
{{/*                ssh root@localhost bash -c "$(cat clone-git.sh)"*/}}
                # sh clone-git.sh | sed 's|.*|[kl-build:clone-git] &|'

            - name: build-image
              image: docker.io/nxtcoder17/alpine.ssh:root
              imagePullPolicy: Always
              env:
                - name: DOCKER_HOST
                  value: 127.0.0.1:2375
              securityContext:
                runAsUser: 0
              volumeMounts:
                - name: host-ssh-root
                  mountPath: /root/.ssh
                - name: host-dir
                  mountPath: /app
              script: |+
{{/*                ls -al $(workspaces.docker-config.path)*/}}
{{/*                ls -al $(workspaces.output.path)*/}}

{{/*                [ -f /root/.ssh/id_rsa.pub ] || ssh-keygen -t rsa -N "" -f /root/.ssh/id_rsa*/}}
{{/*                echo "here"*/}}
{{/*                touch /root/.ssh/authorized_keys*/}}
{{/*                cat /root/.ssh/authorized_keys | grep -i "$(cat /root/.ssh/id_rsa.pub)"*/}}
{{/*                [ $? -ne 0 ] && {*/}}
{{/*                  cat /root/.ssh/id_rsa.pub >> /root/.ssh/authorized_keys*/}}
{{/*                }*/}}
{{/*                bash -x sx.sh*/}}

                cp $(workspaces.docker-config.path)/.dockerconfigjson /app/docker-config.json
                cp -r $(workspaces.output.path)/repo /app/repo

                cat > /app/build-image.sh <<'EOF'

                set -o errexit
                set -o pipefail
                set -o nounset

                DOCKER="podman"

                mkdir -p /etc/containers
                cat > /etc/containers/registries.conf <<'END1'
                [registries.search]
                registries = ['docker.io']
                END1

                # ls -l $(workspaces.docker-config.path)
                mkdir -p ~/.docker
                cp /tmp/{{.PipelineRunId}}/docker-config.json ~/.docker/config.json
{{/*                cp $(workspaces.docker-config.path)/.dockerconfigjson ~/.docker/config.json*/}}

                cd /tmp/{{.PipelineRunId}}/repo
{{/*                cd "$(workspaces.output.path)/repo"*/}}

                buildBaseImage='$(params.{{$varBuildBaseImage}})'
                buildCmd='$(params.{{$varBuildCmd}})'
                buildOutputDir='$(params.{{$varBuildOutputDir}})'

                runBaseImage='$(params.{{$varRunBaseImage}})'
                runCmd='$(params.{{$varRunCmd}})'

                isDockerBuild='$(params.{{$varIsDockerBuild}})'
                dockerfile='$(params.{{$varDockerFile}})'
                dockerContextDir='$(params.{{$varDockerContextDir}})'
                dockerBuildArgs='$(params.{{$varDockerBuildArgs}})'

                dockerImageName='$(params.{{$varArtifactRefDockerImageName}})'
                dockerImageTag='$(params.{{$varArtifactRefDockerImageTag}})'

                gitCommitHash='$(params.{{$varGitCommitHash}})'

                if [ "$isDockerBuild" == "true" ]; then
                  # eval "$buildCmd" | envsubst
                  pushd $dockerContextDir
{{/*                  echo "listing files in context dir"*/}}
{{/*                  ls -al*/}}
                  echo $DOCKER build -f $dockerfile $dockerBuildArgs -t $dockerImageName:$dockerImageTag .
                  $DOCKER build -f $dockerfile $dockerBuildArgs -t $dockerImageName:$dockerImageTag . || exit 1
                  popd
                else
                  IFS=','; read -ra arr <<< $buildOutputDir
                  copyCmds=""
                  for item in ${arr[@]}
                  do
                  item=$(echo $item | xargs echo -n)
                  copyCmds+="COPY --from=build /app/$item ./$item\n"
                done

                cat > /tmp/Dockerfile <<EOF2
                FROM $buildBaseImage AS build
                WORKDIR /app
                COPY . ./
                RUN $buildCmd
                ####

                FROM $runBaseImage
                WORKDIR /app
                RUN ls -al
                $(printf $copyCmds)
                ENTRYPOINT [ "sh", "-c", "$runCmd"]
                EOF2

                  cat /tmp/Dockerfile
                  timeout 2700 $DOCKER build -f /tmp/Dockerfile -t $dockerImageName:$dockerImageTag . ||exit 1
                fi

                echo "pushing docker image: $dockerImageName:$dockerImageTag"
                $DOCKER push "$dockerImageName:$dockerImageTag"
                [ -n "$gitCommitHash" ] && {
                    $DOCKER tag $dockerImageName:$dockerImageTag $dockerImageName:$gitCommitHash
                    echo "pushing docker image: $dockerImageName:$gitCommitHash"
                    $DOCKER push $dockerImageName:$gitCommitHash
                }
                echo "STEP (build-image) FINISHED"
                EOF

                ssh root@localhost "bash /tmp/{{.PipelineRunId}}/build-image.sh"
                # ssh root@localhost bash -c "$(cat build-image.sh)"
{{/*                bash build-image.sh*/}}
                # bash build-image.sh | sed 's|.*|[kl-build:build-image] &|'
{{- end }}
{{- end }}
