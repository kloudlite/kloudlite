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
        - name: init
          image: busybox:1.32.0
          command:
            - sh
            - -c
            - sysctl -w net.ipv4.ip_forward=1 && sysctl -w net.ipv4.conf.all.forwarding=1
          securityContext:
            privileged: true
            capabilities:
              add:
                - NET_ADMIN
                - SYS_MODULE

      containers:
        - image: linuxserver/wireguard
          imagePullPolicy: Always
          name: wg
          resources:
            limits:
              cpu: 80m
              memory: 100Mi
            requests:
              cpu: 50m
              memory: 75Mi
          securityContext:
            capabilities:
              add:
                - NET_ADMIN
                - SYS_MODULE
            privileged: true
          volumeMounts:
            - mountPath: /config/wg_confs/wg0.conf
              name: wg-config
              subPath: wg0.conf
          terminationMessagePath: /dev/termination-log
          terminationMessagePolicy: File

        {{- /* - name: debug */}}
        {{- /*   image: ghcr.io/kloudlite/hub/socat:latest */}}
        {{- /*   imagePullPolicy: Always */}}
        {{- /*   resources: */}}
        {{- /*     limits: */}}
        {{- /*       cpu: 100m */}}
        {{- /*       memory: 100Mi */}}
        {{- /*     requests: */}}
        {{- /*       cpu: 100m */}}
        {{- /*       memory: 100Mi
        {{- /*   command: */}}
        {{- /*     - sh */}}
        {{- /*     - -c */}}
        {{- /*     - |+ */}}
        {{- /*       (socat -dd tcp4-listen:8080,fork,reuseaddr tcp4:kubectl-proxy.{{.Namespace}}.svc.example-test.local:8080 2>&1 | grep -iE --line-buffered 'listening|exiting') & */}}
        {{- /*       pid=$! */}}
        {{- /**/}}
        {{- /*       trap "kill -9 $pid" EXIT SIGINT SIGTERM */}}
        {{- /*       wait $pid */}}

        - name: kube-reverse-proxy
          image: {{.KubeReverseProxyImage}}
          args:
            - --addr
            - ":8080"
            - --proxy-addr
            # this %s will be replaced with real cluster name by reverse proxy
            - {{ printf "kubectl-proxy.kloudlite.svc.{{.CLUSTER_NAME}}.local:8080" }}
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
