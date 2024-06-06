{{- $debugIPManager := false }}
{{- $debugWebhook := false }}
{{- $debugDNS := false }}

{{- $webhookServerHttpPort := "8443" }}
{{- $gatewayAdminHttpPort := "8080" }}
{{- $gatewayWgPort := "51820" }}

{{- /* {{- $dnsUDPPortWg := "5353" }} */}}
{{- /* {{- $dnsUDPPortLocal := "5354" }} */}}
{{- $dnsUDPPortWg := "53" }}
{{- $dnsUDPPortLocal := "54" }}
{{- $dnsHttpPort := "8082" }}

{{- $serviceBindControllerHealtCheckPort := "8081" }}
{{- $serviceBindControllerMetricsPort := "9090" }}

{{- $gatewayAdminApiAddr := printf "http://%s.%s.svc.cluster.local:%s" .Name .Namespace $gatewayAdminHttpPort }}

{{- define "pod-ip" -}}
- name: POD_IP
  valueFrom:
    fieldRef:
      apiVersion: v1
      fieldPath: status.podIP
{{- end -}}

apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{.Name}}
  namespace: {{.Namespace}}
  labels: &labels {{.Labels  | toYAML | nindent 4}}
  annotations: {{.Annotations | toYAML | nindent 4}}
  ownerReferences: {{.OwnerReferences | toYAML | nindent 4}}
spec:
  selector:
    matchLabels: *labels
  template:
    metadata:
      labels: *labels
    spec:
      serviceAccountName: {{.ServiceAccountName}}
      initContainers:
        - name: wg-hostnames
          image: ghcr.io/kloudlite/hub/wireguard:latest
          imagePullPolicy: IfNotPresent
          command:
            - sh
            - -c
            - |
              set -e
              cat > /etc/wireguard/wg0.conf <<EOF
              [Interface]

              PostUp = ip -4 address add {{.GatewayInternalDNSNameserver}}/32 dev wg0
              PostDown = ip -4 address add {{.GatewayInternalDNSNameserver}}/32 dev wg0
              EOF
              wg-quick down wg0 || echo "starting wg0"
              wg-quick up wg0
          securityContext:
            capabilities:
              add:
                - NET_ADMIN
      containers:
      {{- /* # mutation webhook container */}}
      - name: webhook-server
        {{- if $debugWebhook }}
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
          {{include "pod-ip" . | nindent 10}}

          - name: GATEWAY_ADMIN_API_ADDR
            {{- /* value: {{$gatewayAdminApiAddr}} */}}
            value: http://$(POD_IP):{{$gatewayAdminHttpPort}}
        args:
          - --debug
          - --addr
          - $(POD_IP):{{$webhookServerHttpPort}}

        volumeMounts:
        - name: webhook-cert
          mountPath: /tmp/tls
          readOnly: true
        {{- end }}

      {{- /* # runs, wireguard, nginx, and gateway-admin-api */}}
      - name: gateway-admin-api
        {{- if $debugIPManager }}
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
        args:
          - --addr
          - $(POD_IP):{{$gatewayAdminHttpPort}}
        {{- end }}
        imagePullPolicy: Always
        env:
          {{include "pod-ip" . | nindent 10}}

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

          - name: GATEWAY_INTERNAL_DNS_NAMESERVER
            value: "{{.GatewayInternalDNSNameserver}}" 

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

      - name: service-bind-controller
        imagePullPolicy: Always
        image: "ghcr.io/kloudlite/operator/networking/cmd/service-binding-controller:v1.0.7-nightly"
        args:
            - --health-probe-bind-address=$(POD_IP):8081
            - --metrics-bind-address=$(POD_IP):9090
            - --leader-elect
        env:
          {{include "pod-ip" . | nindent 10}}

          - name: MAX_CONCURRENT_RECONCILES
            value: "5"

          - name: GATEWAY_ADMIN_API_ADDR
            {{- /* value: {{$gatewayAdminApiAddr}} */}}
            value: http://$(POD_IP):{{$gatewayAdminHttpPort}}

          - name: SERVICE_DNS_HTTP_ADDR
            value: http://$(POD_IP):{{$dnsHttpPort}}

          - name: GATEWAY_DNS_SUFFIX
            value: {{.GatewayDNSSuffix}}

      - name: dns
        image: "ghcr.io/kloudlite/operator/networking/cmd/dns:v1.0.7-nightly"
        imagePullPolicy: Always
        args:
          - --wg-dns-addr
          {{- /* - $(POD_IP):{{$dnsUDPPortWg}} */}}
          - ":{{$dnsUDPPortWg}}"
          - --local-dns-addr
          {{- /* - "$(POD_IP):{{$dnsUDPPortLocal}}" */}}
          {{- /* - "{{.GatewayInternalDNSNameserver}}:{{$dnsUDPPortLocal}}" */}}
          - "{{.GatewayInternalDNSNameserver}}:{{$dnsUDPPortWg}}"
          - --local-gateway-dns
          - "{{.GatewayDNSSuffix}}"
          - --http-addr
          - $(POD_IP):{{$dnsHttpPort}}
          - --debug
        imagePullPolicy: Always
        {{- /* containerPorts: */}}
        {{- /*   - containerPort: 553 */}}
        {{- /*     name: dns */}}
        {{- /*     protocol: UDP */}}
        {{- /*   - containerPort: 553 */}}
        {{- /*     name: dns-tcp */}}
        {{- /*     protocol: TCP */}}
        {{- /**/}}
        {{- /**/}}
        {{- /*   - containerPort: 554 */}}
        {{- /*     name: dns-local */}}
        {{- /*     protocol: UDP */}}
        {{- /**/}}
        {{- /*   - containerPort: 554 */}}
        {{- /*     name: dns-local-tcp */}}
        {{- /*     protocol: TCP */}}

        env:
          {{include "pod-ip" . | nindent 10}}

          - name: MAX_CONCURRENT_RECONCILES
            value: "5"

          - name: GATEWAY_ADMIN_API_ADDR
            {{- /* value: {{$gatewayAdminApiAddr}} */}}
            value: http://$(POD_IP):{{$gatewayAdminHttpPort}}

      volumes:
        {{- if not $debugWebhook }}
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
  ownerReferences: {{.OwnerReferences | toYAML | nindent 4}}
spec:
  selector: {{.Labels | toYAML | nindent 4}}
  ports:
    - name: wireguard
      port: {{$gatewayWgPort}}
      protocol: UDP
      targetPort: {{$gatewayWgPort}}

    - name: webhook
      port: 443
      protocol: TCP
      targetPort: {{$webhookServerHttpPort}}

    - name: ip-manager
      port: {{$gatewayAdminHttpPort}}
      protocol: TCP
      targetPort: {{$gatewayAdminHttpPort}}

    - name: dns
      port: 53
      protocol: UDP
      targetPort: {{$dnsUDPPortWg}}

    - name: dns-tcp
      port: 53
      protocol: TCP
      targetPort: {{$dnsUDPPortWg}}

    - name: dns-http
      port: {{$dnsHttpPort}}
      protocol: TCP
      targetPort: {{$dnsHttpPort}}

---
apiVersion: v1
kind: Service
metadata:
  name: &name {{.Name}}-local-dns
  namespace: {{.Namespace}}
  labels: {{.Labels | toYAML | nindent 4}}
  ownerReferences: {{.OwnerReferences | toYAML | nindent 4}}
spec:
  selector: {{.Labels | toYAML | nindent 4}}
  ports:
    - name: dns
      port: 53
      protocol: UDP
      targetPort: {{$dnsUDPPortLocal}}
