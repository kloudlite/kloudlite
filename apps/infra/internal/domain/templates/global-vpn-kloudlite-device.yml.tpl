{{- /*gotype: github.com/kloudlite/api/apps/infra/internal/domain/templates.GVPNKloudliteDeviceVars*/ -}}
{{ with . }}

{{- $namespace := .Namespace }}
{{- $klDeviceWgSecretName := printf "%s-wg" .Name }}

{{- $isDebug := true }}

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
        "wg-secret-ref": "{{.WgConfig | b64enc | sha256sum}}"
    spec:
      {{- if not $isDebug }}
      initContainers:
        {{- /* - image: linuxserver/wireguard */}}
        - image: ghcr.io/kloudlite/hub/wireguard:latest
          imagePullPolicy: IfNotPresent
          name: wg
          resources:
            limits:
              cpu: 100m
              memory: 100Mi
            requests:
              cpu: 50m
              memory: 50Mi
          securityContext:
            capabilities:
              add:
                - NET_ADMIN
              drop:
                - all
          command:
            - sh
            - -c
            - |+
              wg-quick down wg0 || echo "[starting] wg-quick down wg0"
              wg-quick up wg0
          volumeMounts:
            - mountPath: /config/wg_confs/wg0.conf
              name: wg-config
              subPath: wg0.conf
          terminationMessagePath: /dev/termination-log
          terminationMessagePolicy: File
      {{- end }}

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

        {{- if $isDebug }}
        - image: ghcr.io/kloudlite/hub/wireguard:latest
          imagePullPolicy: IfNotPresent
          name: wg
          resources:
            limits:
              cpu: 100m
              memory: 100Mi
            requests:
              cpu: 50m
              memory: 50Mi
          securityContext:
            capabilities:
              add:
                - NET_ADMIN
              drop:
                - all
          command:
            - sh
            - -c
            - |+
              wg-quick down wg0 || echo "[starting] wg-quick down wg0"
              wg-quick up wg0
              tail -f /dev/null
              pid=$!
              trap "kill $pid" EXIT SIGINT SIGTERM
              wait $pid
          volumeMounts:
            - mountPath: /config/wg_confs/wg0.conf
              name: wg-config
              subPath: wg0.conf
        {{- end }}

        - name: dns
          image: "ghcr.io/kloudlite/operator/networking/cmd/dns:v1.0.7-nightly"
          imagePullPolicy: Always
          args:
            - --wg-dns-addr
            - :53

            - --dns-servers
            - {{.GatewayDNSServers}}

            - --service-hosts
            -  {{.GatewayServiceHosts}}

            - --debug
          imagePullPolicy: Always
          resources:
            requests:
              cpu: 50m
              memory: 50Mi
            limits:
              cpu: 100m
              memory: 100Mi

          {{- /* securityContext: */}}
          {{- /*   capabilities: */}}
          {{- /*     add: */}}
          {{- /*       - NET_BIND_SERVICE */}}
          {{- /*       - SETGID */}}
          {{- /*     drop: */}}
          {{- /*       - all */}}

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

{{- /* apiVersion: v1 */}}
{{- /* kind: Service */}}
{{- /* metadata: */}}
{{- /*   name: {{printf "%s-wg" .Name | squote }} */}}
{{- /*   namespace: {{.Namespace | squote}} */}}
{{- /* spec: */}}
{{- /*   type: LoadBalancer */}}
{{- /*   selector: */}}
{{- /*     app: {{.Name | squote}} */}}
{{- /*   ports: */}}
{{- /*     - name: wireguard */}}
{{- /*       port: {{.WireguardPort }} */}}
{{- /*       targetPort: {{.WireguardPort}} */}}
{{- /*       protocol: UDP */}}
{{ end }}
