{{- /*gotype: github.com/kloudlite/api/apps/infra/internal/domain/templates.GVPNKloudliteDeviceVars*/ -}}
{{ with . }}

{{- $namespace := .Namespace }}
{{- $klDeviceWgSecretName := "kl-device-wg-secret-name" }}

---
apiVersion: v1
kind: Namespace
metadata:
  name: {{$namespace}}
---
apiVersion: v1
kind: Secret
metadata:
  name: {{$klDeviceWgSecretName | squote}}
  namespace: {{$namespace}}
data:
  wg0.conf: {{.WgConfig | b64enc }}
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{.Name | squote}}
  namespace: {{$namespace}}
spec:
  replicas: 1
  selector:
    matchLabels: &labels
      app: {{.Name | squote}}
  strategy:
    rollingUpdate:
      maxSurge: 25%
      maxUnavailable: 25%
    type: RollingUpdate
  template:
    metadata:
      labels: *labels
      annotations:
        "secret-ref": "{{.WgConfig | b64enc | sha256sum}}"
    spec:
      initContainers:
        - image: linuxserver/wireguard
          imagePullPolicy: IfNotPresent
          name: wg
          {{- /* resources: */}}
          {{- /*   limits: */}}
          {{- /*     cpu: 80m */}}
          {{- /*     memory: 100Mi */}}
          {{- /*   requests: */}}
          {{- /*     cpu: 50m */}}
          {{- /*     memory: 75Mi */}}
          securityContext:
            capabilities:
              add:
                - NET_ADMIN
                {{- /* - SYS_MODULE */}}
            {{- /* privileged: true */}}
          command:
            - wg-quick
            - up
            - wg0
          volumeMounts:
            - mountPath: /config/wg_confs/wg0.conf
              name: wg-config
              subPath: wg0.conf
          terminationMessagePath: /dev/termination-log
          terminationMessagePolicy: File
      containers:
        - name: kube-reverse-proxy
          image: {{.KubeReverseProxyImage}}
          args:
            - --addr
            - ":8080"
            - --proxy-addr
            - {{ printf "kubectl-proxy.kloudlite.svc.{{.CLUSTER_NAME}}.local:8080" }}
            - "--authz"
            - {{.AuthzToken}}
          imagePullPolicy: "IfNotPresent"
          resources:
            limits:
              cpu: 100m
              memory: 100Mi
            requests:
              cpu: 100m
              memory: 100Mi

      dnsPolicy: ClusterFirst
      volumes:
        - name: wg-config
          secret:
            defaultMode: 420
            items:
              - key: "wg0.conf"
                path: wg0.conf
            secretName: {{$klDeviceWgSecretName | squote}}

---
apiVersion: v1
kind: Service
metadata:
  name: {{.Name | squote }}
  namespace: {{.Namespace | squote}}
spec:
  selector:
    app: {{.Name | squote}}
  ports:
    - name: p-8080
      port: 8080
      targetPort: 8080
      protocol: TCP
---
{{ end }}
