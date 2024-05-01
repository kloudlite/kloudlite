{{- if .Values.byok.enabled }}
apiVersion: batch/v1
kind: Job
metadata:
  name: generate-byok-kubeconfig
  namespace: {{ .Release.Namespace }}
  annotations:
    "helm.sh/hook": post-install,post-upgrade
spec:
  template:
    spec:
      tolerations:
        - operator: Exists
      serviceAccountName: {{ include "serviceAccountName" . }}
      containers:
        - name: kubectl
          image: bitnami/kubectl:latest
          command: ["bash"]
          args:
          - -c
          - |+
            cat > k8s-user-account.sh <<EOKUA
            {{.Files.Get "k8s-user-account.sh"}}
            EOKUA

            bash k8s-user-account.sh kubeconfig.yml
            kubectl apply -f - <<EOF
            apiVersion: v1
            kind: Secret
            metadata:
              name: byok-kubeconfig
              namespace: kube-system
              annotations:
                kloudlite.io/dispatch-to-infra: "true"
                kloudlite.io/watch-secret: "true"
            data:
              kubeconfig: $(cat kubeconfig.yml | base64 | tr -d '\n')
            EOF
      restartPolicy: Never
  backoffLimit: 0
{{- end }}
