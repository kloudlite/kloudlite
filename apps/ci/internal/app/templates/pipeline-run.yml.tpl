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
    nodeSelector:
      kloudlite.io/auto-scaler: "true"
    tolerations:
    - key: kloudlite.io/auto-scaler
      operator: Exists
      effect: NoExecute
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
          sidecars:
            - name: dind
              image: docker:dind
              args:
                - dockerd
                - --host
                - tcp://127.0.0.1:2375
              securityContext:
                privileged: true
              volumeMounts:
                - name: $(workspaces.output.volume)
                  mountPath: /var/lib/docker
                  subPath: docker

          workspaces:
            - name: output
            - name: docker-config
          steps:
            - name: clone-git
              image: registry.kloudlite.io/public/git:latest
              script: |+
                cat > clone-git.sh <<'EOF'

                workdir="$(workspaces.output.path)"
                gitUser="$(params.{{$varGitUser}})"
                gitPassword="$(params.{{$varGitPassword}})"
                gitRepo=$(echo $(params.{{$varGitRepo}}) | sed -E "s/https:[/]{2}(.*?\/)/https\:\/\/${gitUser}\:${gitPassword}@\1/g")
                gitBranch="$(params.{{$varGitBranch}})"

                echo "cloning git repo: $gitRepo"

                # if not using git submodules

                cd $workdir
                {{/*        git clone --depth=1 --config remote.origin.pull="${gitBranch}" "$gitRepo" "./repo"*/}}
                echo git clone --depth=1 --branch ${gitBranch} --single-branch "$gitRepo" "./repo"
                git clone --depth=1 --branch ${gitBranch} --single-branch "$gitRepo" "./repo"
                ls -l repo
                echo "successfully cloned git repo: $gitRepo"

                echo "STEP (clone-git) FINISHED"
                EOF

                sh clone-git.sh
                # sh clone-git.sh | sed 's|.*|[kl-build:clone-git] &|'

            - name: build-image
              image: registry.kloudlite.io/public/kloudlite/tekton-builder:latest
              imagePullPolicy: Always
              env:
                - name: DOCKER_HOST
                  value: 127.0.0.1:2375
              securityContext:
                runAsUser: 0
              script: |+
                cat > build-image.sh <<'EOF'

                set -o errexit
                set -o pipefail
                set -o nounset

                # ls -l $(workspaces.docker-config.path)
                mkdir -p ~/.docker
                cp $(workspaces.docker-config.path)/.dockerconfigjson ~/.docker/config.json
                cd "$(workspaces.output.path)/repo"

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
                  echo "listing files in context dir"
                  ls -al
                  echo docker build -f $dockerfile $dockerBuildArgs -t $dockerImageName:$dockerImageTag .
                  {{/* docker buildx build -f $dockerfile $dockerBuildArgs -t $dockerImageName:$dockerImageTag .*/}}
                  docker build -f $dockerfile $dockerBuildArgs -t $dockerImageName:$dockerImageTag . || exit 1
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
                  timeout 2700 docker build -f /tmp/Dockerfile -t $dockerImageName:$dockerImageTag . ||exit 1
                fi

                echo "pushing docker image: $dockerImageName:$dockerImageTag"
                docker push "$dockerImageName:$dockerImageTag"
                [ -n "$gitCommitHash" ] && {
                    docker tag $dockerImageName:$dockerImageTag $dockerImageName:$gitCommitHash
                    echo "pushing docker image: $dockerImageName:$gitCommitHash"
                    docker push $dockerImageName:$gitCommitHash
                }

                echo "STEP (build-image) FINISHED"

                EOF

                bash build-image.sh
                # bash build-image.sh | sed 's|.*|[kl-build:build-image] &|'
{{- end }}
{{- end }}
