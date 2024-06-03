{{- $debug := false }}

apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{.Name}}
  namespace: {{.Namespace}}
  labels: &labels {{.Labels  | toYAML | nindent 4}}
  annotations: {{.Annotations | toYAML | nindent 4}}
spec:
  selector:
    matchLabels: *labels
  template:
    metadata:
      labels: *labels
    spec:
      serviceAccountName: {{.ServiceAccountName}}
      containers:
      {{- /* # wireguard container  */}}
      {{- /* - name: "wireguard" */}}
      {{- /*   image: "ghcr.io/linuxserver/wireguard" */}}
      {{- /*   command: */}}
      {{- /*     - sh */}}
      {{- /*     - -c */}}
      {{- /*     - |+ */}}
      {{- /*       echo "" > /tmp/old-hash */}}
      {{- /*       mkdir -p /config/wg_confs */}}
      {{- /*       trap 'exit 0' SIGTERM SIGINT */}}
      {{- /*       while true; do */}}
      {{- /*         sleep 2 */}}
      {{- /*         curl --silent '{{.GatewayWgConfigHashURI}}' > /tmp/new-hash */}}
      {{- /*         diff /tmp/new-hash /tmp/old-hash */}}
      {{- /*         status=$? */}}
      {{- /*         if [ $status -ne 0 ]; then */}}
      {{- /*           echo "diff detected, will pull new updated wg config" */}}
      {{- /*           cp /tmp/new-hash /tmp/old-hash */}}
      {{- /*           curl --silent '{{.GatewayWgConfigURI}}' > /tmp/new-config */}}
      {{- /*           [ -f /etc/wireguard/wg0.conf ] && wg-quick down wg0 */}}
      {{- /*           mv /tmp/new-config /etc/wireguard/wg0.conf */}}
      {{- /*           wg-quick up wg0 */}}
      {{- /*         fi */}}
      {{- /*       done */}}
      {{- /*       tail -f /dev/null */}}
      {{- /*   securityContext: */}}
      {{- /*     capabilities: */}}
      {{- /*       add: */}}
      {{- /*         - NET_ADMIN */}}

      {{- /* # mutation webhook container */}}
      - name: webhook-server
        {{- if $debug }}
        image: ghcr.io/kloudlite/hub/socat:latest
        command:
          - sh
          - -c
          - |+
            (socat -dd tcp4-listen:8443,fork,reuseaddr tcp4:baby.default.svc.cluster.local:8443 2>&1 | grep -iE --line-buffered 'listening|exiting') &
            pid="$pid $!"

            trap "eval kill -9 $pid || exit 0" EXIT SIGINT SIGTERM
            eval wait $pid
        {{ else }}
        image: {{.WebhookServerImage}}
        imagePullPolicy: Always
        env: 
          - name: GATEWAY_ADMIN_API_ADDR
            value: http://{{.Name}}.{{.Namespace}}.svc.cluster.local:8080
            {{- /* value: http://localhost:8080 */}}
        args:
          - --debug
        ports:
        - containerPort: 8443
        volumeMounts:
        - name: webhook-cert
          mountPath: /tmp/tls
          readOnly: true
        {{- end }}

      # runs, wireguard, nginx, and gateway-admin-api
      - name: gateway-admin-api
        {{- if $debug }}
        image: ghcr.io/kloudlite/hub/socat:latest
        command:
          - sh
          - -c
          - |+
            (socat -dd tcp4-listen:8080,fork,reuseaddr tcp4:baby.default.svc.cluster.local:8090 2>&1 | grep -iE --line-buffered 'listening|exiting') &
            pid="$pid $!"

            trap "eval kill -9 $pid || exit 0" EXIT SIGINT SIGTERM
            eval wait $pid
        {{- else }}
        image: {{.GatewayAdminAPIImage}}
        {{- end }}
        imagePullPolicy: Always
        env:
          - name: GATEWAY_WG_PUBLIC_KEY
            valueFrom:
              secretKeyRef:
                name: {{.GatewayWgSecretName}}
                key: public_key

          - name: GATEWAY_WG_PRIVATE_KEY
            valueFrom:
              secretKeyRef:
                name: {{.GatewayWgSecretName}}
                key: private_key

          - name: GATEWAY_WG_ENDPOINT
            value: {{.Name}}.{{.Namespace}}.svc.cluster.local:51820

          - name: GATEWAY_GLOBAL_IP
            value: {{.GatewayGlobalIP}}

          - name: CLUSTER_CIDR
            value: {{.ClusterCIDR}}

          - name: SERVICE_CIDR
            value: {{.ServiceCIDR}}

          - name: IP_MANAGER_CONFIG_NAME
            value: {{.IPManagerConfigName}}

          - name: IP_MANAGER_CONFIG_NAMESPACE
            value: {{.IPManagerConfigNamespace}}

          - name: POD_ALLOWED_IPS
            value: "100.64.0.0/10"

        {{- /* volumeMounts: */}}
        {{- /*   - name: nginx-stream-configs */}}
        {{- /*     mountPath: /etc/nginx/streams-enabled */}}

        securityContext:
          capabilities:
            add:
              - NET_ADMIN

      # - name: service-bind-controller
      #   image: ""
      #   imagePullPolicy: Always

      volumes:
        - name: wireguard-secret
          secret:
            secretName: wireguard

        {{- /* - name: nginx-stream-configs */}}
        {{- /*   secret: */}}
        {{- /*     secretName: nginx-stream-configs */}}

        {{- /* - name: nginx-stream-configs */}}
        {{- /*   emptyDir: {} */}}

        {{- if not $debug }}
        - name: webhook-cert
          secret:
            secretName: {{.Name}}-webhook-cert
            items:
              - key: tls.crt
                path: tls.crt

              - key: tls.key
                path: tls.key
        {{- end }}

---

apiVersion: v1
kind: Service
metadata:
  name: &name {{.Name}}
  namespace: {{.Namespace}}
  labels: {{.Labels | toYAML | nindent 4}}
spec:
  selector: {{.Labels | toYAML | nindent 4}}
  ports:
    - name: wireguard
      port: 51820
      protocol: UDP
      targetPort: 51820

    - name: webhook
      port: 443
      protocol: TCP
      targetPort: 8443

    - name: ip-manager
      port: 8080
      protocol: TCP
      targetPort: 8080
