apiVersion: batch/v1
kind: Job
metadata:
  name: {{ .Name }}-{{ .Tag }}
  namespace: {{ .Namespace }}
  labels: {{ .Labels | toJson }}
spec:
  backoffLimit: 0
  suspend: false
  template:
    metadata:
      name: {{ .Name }}-{{ .Tag }}
    spec:
      containers:
      - name: build-container
        image: docker:dind
        env:
        - name: DOCKER_HOST
          value: {{ .DockerHost}}
        - name: DOCKER_PSW
          value: {{ .DockerPassword }}

        command: ["sh", "-c"]
        args:
        - |
          tag={{ .Registry }}/{{ .RegistryRepoName }}:{{ .Tag }}
          docker build -t $tag {{ .PullUrl }} &&
          echo $DOCKER_PSW | docker login -u admin --password-stdin {{ .Registry }} &&
          docker push $tag
      restartPolicy: Never
