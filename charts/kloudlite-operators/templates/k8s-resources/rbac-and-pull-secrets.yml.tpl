{{ $dockerSecretName := "kl-docker-creds" }}
---
apiVersion: v1
kind: Secret
metadata:
  name: {{$dockerSecretName}}
  namespace: {{.Release.Namespace}}
type: kubernetes.io/dockerconfigjson
data:
  .dockerconfigjson: {{.Values.imagePullSecret.dockerconfigjson}}
---

apiVersion: v1
kind: ServiceAccount
metadata:
  name: {{.Values.svcAccountName}}
  namespace: {{.Release.Namespace}}
imagePullSecrets:
  - name: {{$dockerSecretName}}

---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: {{.Release.Namespace}}-{{.Values.svcAccountName}}-rb
subjects:
  - kind: ServiceAccount
    name: {{.Values.svcAccountName}}
    namespace: {{.Release.Namespace}}
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: "ClusterRole"
  name: cluster-admin
---

