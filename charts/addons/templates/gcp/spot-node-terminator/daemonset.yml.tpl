{{- $gcpSpotTerminationHandler := "gcp-spot-termination-handler" }}
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: {{$gcpSpotTerminationHandler}}
  namespace: {{.Release.Namespace}}
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: {{$gcpSpotTerminationHandler}}-rb
subjects:
  - kind: ServiceAccount
    name: {{$gcpSpotTerminationHandler}}
    namespace: {{.Release.Namespace}}
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: "ClusterRole"
  name: cluster-admin
---
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: &name {{$gcpSpotTerminationHandler}}
  namespace: {{.Release.Namespace}}
spec:
  selector:
    matchLabels:
      name: *name
  template:
    metadata:
      labels:
        name: *name
    spec:
      serviceAccountName: {{$gcpSpotTerminationHandler}}
      nodeSelector:
        kloudlite.io/node.is-spot: "true"
      containers:
      - name: main
        image: {{.Values.gcp.spot_node_terminator.configuration.image.repository}}:{{.Values.gcp.spot_node_terminator.configuration.image.tag | default (include "image-tag" .) }}
        env:
          - name: DEBUG
            value: "false"
          - name: NODE_NAME
            valueFrom:
              fieldRef: 
                fieldPath: spec.nodeName
        resources:
          limits:
            memory: 100Mi
            cpu: 100m
          requests:
            memory: 20Mi
            cpu: 20m
      terminationGracePeriodSeconds: 10
