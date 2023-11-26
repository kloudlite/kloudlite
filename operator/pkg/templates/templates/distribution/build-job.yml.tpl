{{- $name := .Name -}}
{{- $namespace := .Namespace -}}
{{- $labels := .Labels -}}
{{- $annotations := .Annotations -}}

{{- $accountName := .AccountName -}}

{{- $registryHost := .RegistryHost -}}
{{- $registryUsername := .RegistryUsername -}}
{{- $registryPassword := .RegistryPassword -}}

{{- $gitRepoUrl := .GitRepoUrl -}}
{{- $gitRepoBranch := .GitRepoBranch -}}

{{- $registryTags := .RegistryTags -}}

apiVersion: batch/v1
kind: Job
metadata:
  name: build-{{ $name }}
  namespace: {{ $namespace }}
  labels: {{ $labels | toJson }}
  annotations: {{ $annotations | toJson }}
spec:
  backoffLimit: 0
  suspend: false
  template:
    metadata:
      name: build-{{ $name }}
    spec:
      containers:
      - name: build-and-push
        args:
        - |
          set -o errexit
          set -o pipefail

          trap 'echo "[#] kill signal received" && pkill dockerd' SIGINT SIGTERM EXIT

          counter=0
          while [ $counter -lt 20 ]; do
            docker info > /dev/null 2>&1 && break
            echo "[#] waiting for docker to be available"
            counter=$((counter+1))
            sleep 3
          done

          if [ $counter -eq 10 ]; then
            echo "[#] docker not available after 60 seconds, exiting"
            exit 1
          fi

          echo "[#] logging into docker registry\n"
          echo $DOCKER_PSW | docker login -u {{ $registryUsername }} --password-stdin {{ $registryHost }} > /dev/null 2>&1

          # temporary work dir
          TEMP_DIR=$(mktemp -d -t ci-XXXXXXXXXX)
          cd $TEMP_DIR

          echo "[#] Cloning {{ $gitRepoBranch }}\n"
          git init > /dev/null 2>&1
          git fetch --depth=1 {{$gitRepoUrl}} {{$gitRepoBranch}}
          git checkout {{ $gitRepoBranch }} > /dev/null 2>&1

          DOCKER_FILE_PATH=$TEMP_DIR/{{.BuildOptions.DockerfilePath}}
          CONTEXT_DIR=$TEMP_DIR/{{.BuildOptions.ContextDir}}
          {{if .BuildOptions.DockerfileContent }}
          echo "[#] overwriting dockerfile with provided content\n"
          cat > $DOCKER_FILE_PATH <<EOF
          {{ .BuildOptions.DockerfileContent | indent 10 }}
          EOF
          {{- else}}
          echo "[#] using dockerfile from repo\n"
          {{- end}}

          {{/* docker buildx create --use > /dev/null 2>&1 */}}
          echo "[#] Initalizing build and push\n"
          docker buildx build \
          {{$registryTags}} \
          --file $DOCKER_FILE_PATH \
          {{.BuildOptions.BuildContexts}} \
          {{.BuildOptions.BuildArgs}} \
          {{.BuildOptions.TargetPlatforms}} \
          -o type=registry,oci-mediatypes=true,compression=estargz,force-compression=true \
          $CONTEXT_DIR  \
          2>&1 | grep -v '\[internal\]'

        command: ["bash", "-c"]
        volumeMounts:
        - name: docker-socket
          mountPath: /var/run
        image: ghcr.io/kloudlite/image-builder:v1.0.5-nightly
        env:
        - name: DOCKER_PSW
          value: {{ $registryPassword }}
        resources:
          requests:
            memory: "{{.ClientResource.Memory}}Mi"
            cpu: "{{.ClientResource.Cpu}}m"
          limits:
            memory: "{{.ClientResource.Memory}}Mi"
            cpu: "{{.ClientResource.Cpu}}m"
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
        {{- if .BuildCacheKey}}
        - name: cache-volume
          mountPath: /var/lib/docker
        {{- end}}
        image: ghcr.io/kloudlite/platform/apis/docker:dind
        securityContext:
          privileged: true
        resources:
          requests:
            memory: "{{.ServerResource.Memory}}Mi"
            cpu: "{{.ServerResource.Cpu}}m"
          limits:
            memory: "{{.ServerResource.Memory}}Mi"
            cpu: "{{.ServerResource.Cpu}}m"
      shareProcessNamespace: true
      volumes:
      - name: docker-socket
        emptyDir: {}
      {{- if .BuildCacheKey}}
      - name: cache-volume
        persistentVolumeClaim:
          claimName: {{ .BuildCacheKey }}
      {{- end}}
      restartPolicy: Never
