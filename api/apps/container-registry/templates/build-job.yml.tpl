{{- $name := .Name -}}
{{- $namespace := .Namespace -}}
{{- $labels := .Labels -}}
{{- $annotations := .Annotations -}}
{{- $accountName := .AccountName -}}
{{- $registry := .Registry -}}
{{- $registryRepoName := .RegistryRepoName -}}
{{- $tags := .Tags -}}
{{- $gitRepoUrl := .GitRepoUrl -}}
{{- $branch := .Branch -}}
{{- $klAdmin := .KlAdmin -}}
{{- $dockerPassword := .DockerPassword -}}




apiVersion: batch/v1
kind: Job
metadata:
  name: {{ $name }}
  namespace: {{ $namespace }}
  labels: {{ $labels | toJson }}
  annotations: {{ $annotations | toJson }}

spec:
  backoffLimit: 0
  suspend: false
  template:
    metadata:
      name: {{ $name }}
    spec:
      shareProcessNamespace: true
      volumes:
      - name: docker-socket
        emptyDir: {}

      - name: hostpath-volume
        hostPath:
          path: /var/docker-data/{{ $accountName}}
          type: DirectoryOrCreate


      containers:
      - name: dind-server
        command:
        - /bin/sh
        - -c
        args:
        - |
          dockerd-entrypoint.sh > /dev/null 2>&1
        volumeMounts:
        - name: docker-socket
          mountPath: /var/run
        - name: hostpath-volume
          mountPath: /var/lib/docker

        image: ghcr.io/kloudlite/platform/apis/docker:dind
        securityContext:
          privileged: true
        resources:
          requests:
            memory: "2048Mi"
            cpu: "1"
          limits:
            memory: "2048Mi"
            cpu: "1"


      - name: build-and-push
        volumeMounts:
        - name: docker-socket
          mountPath: /var/run
        image: ghcr.io/kloudlite/image-builder:v1.0.5-nightly
        env:
        - name: DOCKER_PSW
          value: {{ $dockerPassword }}

        command: ["bash", "-c"]
        args:
        - |
          set -o errexit
          set -o pipefail

          trap 'pkill dockerd' SIGINT SIGTERM EXIT
          while ! docker info > /dev/null 2>&1 ; do sleep 1; done

          echo $DOCKER_PSW | docker login -u {{ $klAdmin }} --password-stdin {{ $registry }} > /dev/null 2>&1

          # temporary work dir
          TEMP_DIR=$(mktemp -d -t ci-XXXXXXXXXX)
          CONTEXT_DIR=$TEMP_DIR
          cd $TEMP_DIR

          git init > /dev/null 2>&1
          git fetch --depth=1 {{$gitRepoUrl}} {{$branch}}
          git checkout {{ $branch }} > /dev/null 2>&1

          TARGET_PLATFORMS=""

          DOCKER_FILE_PATH=./Dockerfile
          {{- if .BuildOptions }}# Operations for if BuildOptions Provided

          {{- if and  .BuildOptions.TargetPlatforms (ne (len .BuildOptions.TargetPlatforms) 0)}}# setting target platforms
          TARGET_PLATFORMS=--platforms '{{join "," .BuildOptions.TargetPlatforms}}'
          {{- end}}

          {{- if .BuildOptions.DockerfilePath}}# overriding dockerfile path
          $DOCKER_FILE_PATH={{ .BuildOptions.DockerfilePath }}
          {{- end}}

          {{- if .BuildOptions.ContextDir}}# setting context dir
          CONTEXT_DIR=$TEMP_DIR/'{{ .BuildOptions.ContextDir }}'
          {{- end}}

          {{- if .BuildOptions.DockerfileContent }}# writing dockerfile
          cat > $DOCKER_FILE_PATH <<EOF
          {{ .BuildOptions.DockerfileContent | indent 10 }}
          EOF
          {{- end}}

          {{if .BuildOptions.BuildContexts}}
          BUILD_CONTEXTS=""
          {{- range $key, $value := .BuildOptions.BuildContexts}}
          BUILD_CONTEXTS="$BUILD_CONTEXTS  --build-context '{{$key}}={{ $value }}'"{{end}}
          {{- end}}

          {{if .BuildOptions.BuildArgs}}
          BUILD_ARGS=""
          {{- range $key, $value := .BuildOptions.BuildArgs}}
          BUILD_ARGS="$BUILD_ARGS --build-arg {{$key}}={{ $value }}"
          {{- end}}
          {{- end}}

          {{- end}}
          {{/* docker buildx create --use > /dev/null 2>&1 */}}

          docker buildx build \
          {{- range $tags}}
          --tag '{{ $registry }}/{{ $registryRepoName }}:{{ . }}' \
          {{- end}}
          --file $DOCKER_FILE_PATH \
          $BUILD_CONTEXTS \
          $BUILD_ARGS \
          $TARGET_PLATFORMS \
          -o type=registry,oci-mediatypes=true,compression=estargz,force-compression=true \
          $CONTEXT_DIR  \
          2>&1 | grep -v '\[internal\]'
      restartPolicy: Never
