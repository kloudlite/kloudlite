---
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: &name {{.Values.name}}
  namespace: {{.Release.Namespace}}
  labels:
    installed-by: kloudlite
spec:
  selector:
    matchLabels:
      name: *name
  template:
    metadata:
      labels:
        name: *name
    spec:
      serviceAccountName: {{.Values.name}}
      nodeSelector: {{ .Values.nodeSelector | toYaml | nindent 10 }}
      containers:
      - name: main
        image: {{.Values.image.name}}:{{.Values.kloudliteRelease}}
        imagePullPolicy: {{.Values.imagePullPolicy}}
        env:
         - name: NODE_NAME
           valueFrom:
             fieldRef:
               fieldPath: spec.nodeName
         - name: DEBUG
           value: "true"
        resources:
          limits:
            memory: 50Mi
            cpu: 50m
          requests:
            memory: 20Mi
            cpu: 20m
      terminationGracePeriodSeconds: 10
---
