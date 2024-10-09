{{- $debugIPManager := false }}
{{- $debugWebhook := false }}
{{- $debugDNS := false }}

{{- $webhookServerHttpPort := "8443" }}
{{- $gatewayAdminHttpPort := "8080" }}
{{- $gatewayWgPort := "51820" }}

{{- $dnsUDPPortWg := "53" }}
{{- $dnsUDPPortLocal := "54" }}
{{- $dnsHttpPort := "8082" }}
{{- $kubectlProxyHttpPort := "8383" }}

{{- $serviceBindControllerHealtCheckPort := "8081" }}
{{- $serviceBindControllerMetricsPort := "9090" }}

{{- /* should refrain from using it, as it requires coredns to be up and running */}}
{{- /* {{- $gatewayAdminApiAddr := printf "http://%s.%s.svc.cluster.local:%s" .Name .Namespace $gatewayAdminHttpPort }} */}}

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
      annotations:
        kloudlite.io/gateway-extra-peers-hash: {{.GatewayWgExtraPeersHash}}
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
              cat > /etc/wireguard/wg0.conf <<EOF
              [Interface]
              PostUp = ip -4 address add {{.GatewayInternalDNSNameserver}}/32 dev wg0
              PostDown = ip -4 address add {{.GatewayInternalDNSNameserver}}/32 dev wg0
              EOF
              wg-quick down wg0 || echo "starting wg0"
              wg-quick up wg0

              while true; do
                ip -4 addr | grep -i "{{.GatewayInternalDNSNameserver}}"
                exit_code=$?

                [ $exit_code -eq 0 ] && break
                echo "waiting for wireguard to come up"
                sleep 1
              done
          env:
            {{include "pod-ip" . | nindent 12}}
          resources:
            requests:
              cpu: 50m
              memory: 50Mi
            limits:
              cpu: 300m
              memory: 300Mi
          securityContext:
            capabilities:
              add:
                - NET_ADMIN
              drop:
                - all
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
        image: {{.ImageWebhookServer}}
        imagePullPolicy: Always
        env: 
          {{include "pod-ip" . | nindent 10}}

          - name: GATEWAY_ADMIN_API_ADDR
            value: http://$(POD_IP):{{$gatewayAdminHttpPort}}
        args:
          - --addr
          - $(POD_IP):{{$webhookServerHttpPort}}
          - --wg-image
          - ghcr.io/kloudlite/hub/wireguard:latest
        resources:
          requests:
            cpu: 50m
            memory: 50Mi
          limits:
            cpu: 300m
            memory: 300Mi

        volumeMounts:
        - name: webhook-cert
          mountPath: /tmp/tls
          readOnly: true
        {{- end }}

      {{- /* # runs, wireguard, nginx, and gateway-admin-api */}}
      - name: ip-manager
        {{- if $debugIPManager }}
        image: ghcr.io/kloudlite/hub/socat:latest
        command:
          - sh
          - -c
          - |+
            (socat -dd tcp4-listen:8080,fork,reuseaddr tcp4:baby.device.local:8090 2>&1 | grep -iE --line-buffered 'listening|exiting') &
            pid="$pid $!"

            trap "eval kill -9 $pid || exit 0" EXIT SIGINT SIGTERM
            eval wait $pid
        {{- else }}
        image: {{.ImageIPManager}}
        {{- /* args:  */}}
          {{- /* - --addr */}}
          {{- /* - $(POD_IP):{{$gatewayAdminHttpPort}} */}}
        command:
          - sh
          - -c
          - |+
            mkdir -p /etc/wireguard
            for file in `find /tmp/include-wg-interfaces -type f`; do
              basepath=`basename $file`
              cp $file /etc/wireguard/$basepath
              wg-quick up "${basepath%.*}"
            done

            /entrypoint.sh --addr $(POD_IP):{{$gatewayAdminHttpPort}}

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
            {{- /* value: {{.Name}}.{{.Namespace}}.svc.cluster.local:51820 */}}
            value: $(POD_IP):51820

          - name: EXTRA_WIREGUARD_PEERS_PATH
            value: "/tmp/peers.conf"

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

        volumeMounts:
          - name: gateway-wg-extra-peers
            mountPath: /tmp/peers.conf
            subPath: peers.conf

          - name: include-wg-interfaces
            mountPath: /tmp/include-wg-interfaces

        resources:
          requests:
            cpu: 100m
            memory: 100Mi
          limits:
            cpu: 300m
            memory: 300Mi

        securityContext:
          capabilities:
            add:
              - NET_ADMIN

      - name: ip-binding-controller
        imagePullPolicy: Always
        image: "{{.ImageIPBindingController}}"
        args:
          - --health-probe-bind-address=$(POD_IP):8081
          - --metrics-bind-address=$(POD_IP):9090
          {{- /* - --leader-elect */}}
        resources:
          requests:
            cpu: 100m
            memory: 100Mi
          limits:
            cpu: 300m
            memory: 300Mi

        env:
          {{include "pod-ip" . | nindent 10}}

          - name: MAX_CONCURRENT_RECONCILES
            value: "5"

          - name: GATEWAY_ADMIN_API_ADDR
            value: http://$(POD_IP):{{$gatewayAdminHttpPort}}

          - name: SERVICE_DNS_HTTP_ADDR
            value: http://$(POD_IP):{{$dnsHttpPort}}

          - name: GATEWAY_DNS_SUFFIX
            value: {{.GatewayDNSSuffix}}

        securityContext:
          capabilities:
            add:
              {{- /* - NET_BIND_SERVICE */}}
              - NET_RAW
            drop:
              - all

      - name: dns
        image: "{{.ImageDNS}}"
        imagePullPolicy: Always
        args:
          - --wg-dns-addr
          - :{{$dnsUDPPortWg}}

          - --enable-local-dns

          - --local-dns-addr
          - "{{.GatewayInternalDNSNameserver}}:{{$dnsUDPPortWg}}"

          - --local-gateway-dns
          - "{{.GatewayDNSSuffix}}"

          - --enable-http
          - --http-addr
          - $(POD_IP):{{$dnsHttpPort}}

          - --dns-servers
          - {{.GatewayDNSServers}}

          - --service-hosts
          - pod-logs-proxy.{{.Namespace}}.{{.GatewayDNSSuffix}}={{.GatewayGlobalIP}}

          - --debug
        imagePullPolicy: Always
        resources:
          requests:
            cpu: 50m
            memory: 50Mi
          limits:
            cpu: 300m
            memory: 300Mi

        securityContext:
          capabilities:
            add:
              - NET_BIND_SERVICE
              - SETGID
            drop:
              - all

        env:
          {{include "pod-ip" . | nindent 10}}

          - name: MAX_CONCURRENT_RECONCILES
            value: "5"

          - name: GATEWAY_ADMIN_API_ADDR
            value: http://$(POD_IP):{{$gatewayAdminHttpPort}}

      - name: logs-proxy
        image: "{{.ImageLogsProxy}}"
        imagePullPolicy: Always
        command:
          - sh
          - -c
          - |
            while true; do
              ip -4 addr | grep -i "{{.GatewayGlobalIP}}"
              exit_code=$?

              [ $exit_code -eq 0 ] && break
              echo "waiting for ip-manager to be ready"
              sleep 1
            done

            $EXECUTABLE --addr {{.GatewayGlobalIP}}:{{$kubectlProxyHttpPort}}
        {{- /* args: */}}
        {{- /*   - --addr */}}
        {{- /*   - {{.GatewayGlobalIP}}:{{$kubectlProxyHttpPort}} */}}
        resources:
          limits:
            cpu: 100m
            memory: 100Mi
          requests:
            cpu: 100m
            memory: 100Mi

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

        - name: gateway-wg-extra-peers
          configMap:
            name: {{.Name}}-wg-extra-peers
            items:
              - key: peers.conf
                path: peers.conf

        - name: include-wg-interfaces
          secret:
            secretName: include-wg-interfaces
            optional: true
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
  name: &name {{.Name}}-wg
  namespace: {{.Namespace}}
  labels: {{.Labels | toYAML | nindent 4}}
  ownerReferences: {{.OwnerReferences | toYAML | nindent 4}}
spec:
  type: {{.GatewayServiceType}}
  selector: {{.Labels | toYAML | nindent 4}}
  ports:
    - name: wireguard
      {{- if (and (eq .GatewayServiceType "NodePort") (ne .GatewayNodePort 0)) }}
      nodePort: {{.GatewayNodePort}}
      {{- end }}
      port: 31820
      protocol: UDP
      targetPort: {{$gatewayWgPort}}
