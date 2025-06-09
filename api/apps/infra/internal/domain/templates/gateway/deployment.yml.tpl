---
apiVersion: v1
kind: Namespace
metadata:
  name: kl-gateway
---

apiVersion: v1
kind: Namespace
metadata:
  name: {{.GatewayWgSecret}}
  namespace: kl-gateway
data:
  private_key: "{{.GatewayPrivateKey}}"
  public_key: "{{.GatewayPublicKey}}"
---
{{- $webhookServerHttpPort := "8443" }}
{{- $gatewayAdminHttpPort := "8080" }}
{{- $gatewayWgPort := "51820" }}

{{- $dnsUDPPortWg := "53" }}
{{- $dnsUDPPortLocal := "54" }}
{{- $dnsHttpPort := "8082" }}
{{- $kubectlProxyHttpPort := "8383" }}

{{- $serviceBindControllerHealtCheckPort := "8081" }}
{{- $serviceBindControllerMetricsPort := "9090" }}

{{- /* INFO: should refrain from using it, as it requires coredns to be up and running */}}
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
metadata: {{.ObjectMeta | toJson }}
spec:
  selector:
    matchLabels: &labels
      kloudlite.io/deployment.name: {{.ObjectMeta.Name}}
  template:
    metadata:
      labels:
        kloudlite.io/deployment.name: {{.ObjectMeta.Name}}
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
      - name: webhook-server
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

      {{- /* # runs, wireguard, nginx, and gateway-admin-api */}}
      - name: ip-manager
        image: {{.ImageIPManager}}
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
          - --leader-elect
          - "false"
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
        resources:
          requests:
            cpu: 100m
            memory: 100Mi
          limits:
            cpu: 300m
            memory: 300Mi

      volumes:
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

---

apiVersion: admissionregistration.k8s.io/v1
kind: MutatingWebhookConfiguration
metadata:
  annotations:
    kloudlite.io/last-applied: '{"apiVersion":"admissionregistration.k8s.io/v1","kind":"MutatingWebhookConfiguration","metadata":{"name":"default-webhook","namespace":"kl-gateway-default","ownerReferences":[{"apiVersion":"networking.kloudlite.io/v1","blockOwnerDeletion":true,"controller":true,"kind":"Gateway","name":"default","uid":"24fad57f-836f-41b8-85af-129c3054c205"}]},"webhooks":[{"admissionReviewVersions":["v1"],"clientConfig":{"caBundle":"LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUJiVENDQVJPZ0F3SUJBZ0lCQVRBS0JnZ3Foa2pPUFFRREFqQWRNUnN3R1FZRFZRUUtFeEpOZVNCUGNtZGgKYm1sNllYUnBiMjRnUTBFd0lCY05NalF4TURBM01EWXdNelUyV2hnUE1qRXlOREE1TVRNd05qQXpOVFphTUIweApHekFaQmdOVkJBb1RFazE1SUU5eVoyRnVhWHBoZEdsdmJpQkRRVEJaTUJNR0J5cUdTTTQ5QWdFR0NDcUdTTTQ5CkF3RUhBMElBQkhqdEc5SEJiWGpKMEZvUjVCQXRxRmdhSVpWMlNIdDdOQjVVTFZ1Rnd3bHhXU25nSFg5MHhCMS8KM1NOK2pmandMaWtSOHVpQlltSGpWSkExdXd5UDd5cWpRakJBTUE0R0ExVWREd0VCL3dRRUF3SUJCakFQQmdOVgpIUk1CQWY4RUJUQURBUUgvTUIwR0ExVWREZ1FXQkJSeE9CcExvdmgyMkNQL1lXRmpqVVFVRTExYjZUQUtCZ2dxCmhrak9QUVFEQWdOSUFEQkZBaUEya1BxdEV3QjVmUjd4YnZjeTZvVHhtbW9VcWdNc09tYVhIZ2s1NlJwL05BSWgKQUtjTnBjeXFwNlE0NVpUWkUxUU1jSlJqdHc4VDZRUktwTGxQemM1RjRhV1YKLS0tLS1FTkQgQ0VSVElGSUNBVEUtLS0tLQo=","service":{"name":"default","namespace":"kl-gateway-default","path":"/mutate/pod"}},"name":"default-pod.kl-gateway-default.webhook.com","namespaceSelector":{"matchExpressions":[{"key":"kloudlite.io/gateway.enabled","operator":"In","values":["true"]}]},"rules":[{"apiGroups":[""],"apiVersions":["v1"],"operations":["CREATE","DELETE"],"resources":["pods"],"scope":"Namespaced"}],"sideEffects":"None"},{"admissionReviewVersions":["v1"],"clientConfig":{"caBundle":"LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUJiVENDQVJPZ0F3SUJBZ0lCQVRBS0JnZ3Foa2pPUFFRREFqQWRNUnN3R1FZRFZRUUtFeEpOZVNCUGNtZGgKYm1sNllYUnBiMjRnUTBFd0lCY05NalF4TURBM01EWXdNelUyV2hnUE1qRXlOREE1TVRNd05qQXpOVFphTUIweApHekFaQmdOVkJBb1RFazE1SUU5eVoyRnVhWHBoZEdsdmJpQkRRVEJaTUJNR0J5cUdTTTQ5QWdFR0NDcUdTTTQ5CkF3RUhBMElBQkhqdEc5SEJiWGpKMEZvUjVCQXRxRmdhSVpWMlNIdDdOQjVVTFZ1Rnd3bHhXU25nSFg5MHhCMS8KM1NOK2pmandMaWtSOHVpQlltSGpWSkExdXd5UDd5cWpRakJBTUE0R0ExVWREd0VCL3dRRUF3SUJCakFQQmdOVgpIUk1CQWY4RUJUQURBUUgvTUIwR0ExVWREZ1FXQkJSeE9CcExvdmgyMkNQL1lXRmpqVVFVRTExYjZUQUtCZ2dxCmhrak9QUVFEQWdOSUFEQkZBaUEya1BxdEV3QjVmUjd4YnZjeTZvVHhtbW9VcWdNc09tYVhIZ2s1NlJwL05BSWgKQUtjTnBjeXFwNlE0NVpUWkUxUU1jSlJqdHc4VDZRUktwTGxQemM1RjRhV1YKLS0tLS1FTkQgQ0VSVElGSUNBVEUtLS0tLQo=","service":{"name":"default","namespace":"kl-gateway-default","path":"/mutate/service"}},"name":"default-svc.kl-gateway-default.webhook.com","namespaceSelector":{"matchExpressions":[{"key":"kloudlite.io/gateway.enabled","operator":"In","values":["true"]}]},"rules":[{"apiGroups":[""],"apiVersions":["v1"],"operations":["CREATE","UPDATE","DELETE"],"resources":["services"],"scope":"Namespaced"}],"sideEffects":"None"}]}'
  creationTimestamp: "2024-10-07T06:03:57Z"
  generation: 1
  name: default-webhook
  ownerReferences:
  - apiVersion: networking.kloudlite.io/v1
    blockOwnerDeletion: true
    controller: true
    kind: Gateway
    name: default
    uid: 24fad57f-836f-41b8-85af-129c3054c205
  resourceVersion: "1184"
  uid: 3ab763bb-1f40-440b-b6ec-e8fb9a431c08
webhooks:
- admissionReviewVersions:
  - v1
  clientConfig:
    caBundle: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUJiVENDQVJPZ0F3SUJBZ0lCQVRBS0JnZ3Foa2pPUFFRREFqQWRNUnN3R1FZRFZRUUtFeEpOZVNCUGNtZGgKYm1sNllYUnBiMjRnUTBFd0lCY05NalF4TURBM01EWXdNelUyV2hnUE1qRXlOREE1TVRNd05qQXpOVFphTUIweApHekFaQmdOVkJBb1RFazE1SUU5eVoyRnVhWHBoZEdsdmJpQkRRVEJaTUJNR0J5cUdTTTQ5QWdFR0NDcUdTTTQ5CkF3RUhBMElBQkhqdEc5SEJiWGpKMEZvUjVCQXRxRmdhSVpWMlNIdDdOQjVVTFZ1Rnd3bHhXU25nSFg5MHhCMS8KM1NOK2pmandMaWtSOHVpQlltSGpWSkExdXd5UDd5cWpRakJBTUE0R0ExVWREd0VCL3dRRUF3SUJCakFQQmdOVgpIUk1CQWY4RUJUQURBUUgvTUIwR0ExVWREZ1FXQkJSeE9CcExvdmgyMkNQL1lXRmpqVVFVRTExYjZUQUtCZ2dxCmhrak9QUVFEQWdOSUFEQkZBaUEya1BxdEV3QjVmUjd4YnZjeTZvVHhtbW9VcWdNc09tYVhIZ2s1NlJwL05BSWgKQUtjTnBjeXFwNlE0NVpUWkUxUU1jSlJqdHc4VDZRUktwTGxQemM1RjRhV1YKLS0tLS1FTkQgQ0VSVElGSUNBVEUtLS0tLQo=
    service:
      name: default
      namespace: kl-gateway-default
      path: /mutate/pod
      port: 443
  failurePolicy: Fail
  matchPolicy: Equivalent
  name: default-pod.kl-gateway-default.webhook.com
  namespaceSelector:
    matchExpressions:
    - key: kloudlite.io/gateway.enabled
      operator: In
      values:
      - "true"
  objectSelector: {}
  reinvocationPolicy: Never
  rules:
  - apiGroups:
    - ""
    apiVersions:
    - v1
    operations:
    - CREATE
    - DELETE
    resources:
    - pods
    scope: Namespaced
  sideEffects: None
  timeoutSeconds: 10
- admissionReviewVersions:
  - v1
  clientConfig:
    caBundle: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUJiVENDQVJPZ0F3SUJBZ0lCQVRBS0JnZ3Foa2pPUFFRREFqQWRNUnN3R1FZRFZRUUtFeEpOZVNCUGNtZGgKYm1sNllYUnBiMjRnUTBFd0lCY05NalF4TURBM01EWXdNelUyV2hnUE1qRXlOREE1TVRNd05qQXpOVFphTUIweApHekFaQmdOVkJBb1RFazE1SUU5eVoyRnVhWHBoZEdsdmJpQkRRVEJaTUJNR0J5cUdTTTQ5QWdFR0NDcUdTTTQ5CkF3RUhBMElBQkhqdEc5SEJiWGpKMEZvUjVCQXRxRmdhSVpWMlNIdDdOQjVVTFZ1Rnd3bHhXU25nSFg5MHhCMS8KM1NOK2pmandMaWtSOHVpQlltSGpWSkExdXd5UDd5cWpRakJBTUE0R0ExVWREd0VCL3dRRUF3SUJCakFQQmdOVgpIUk1CQWY4RUJUQURBUUgvTUIwR0ExVWREZ1FXQkJSeE9CcExvdmgyMkNQL1lXRmpqVVFVRTExYjZUQUtCZ2dxCmhrak9QUVFEQWdOSUFEQkZBaUEya1BxdEV3QjVmUjd4YnZjeTZvVHhtbW9VcWdNc09tYVhIZ2s1NlJwL05BSWgKQUtjTnBjeXFwNlE0NVpUWkUxUU1jSlJqdHc4VDZRUktwTGxQemM1RjRhV1YKLS0tLS1FTkQgQ0VSVElGSUNBVEUtLS0tLQo=
    service:
      name: default
      namespace: kl-gateway-default
      path: /mutate/service
      port: 443
  failurePolicy: Fail
  matchPolicy: Equivalent
  name: default-svc.kl-gateway-default.webhook.com
  namespaceSelector:
    matchExpressions:
    - key: kloudlite.io/gateway.enabled
      operator: In
      values:
      - "true"
  objectSelector: {}
  reinvocationPolicy: Never
  rules:
  - apiGroups:
    - ""
    apiVersions:
    - v1
    operations:
    - CREATE
    - UPDATE
    - DELETE
    resources:
    - services
    scope: Namespaced
  sideEffects: None
  timeoutSeconds: 10
---
apiVersion: v1
kind: ServiceAccount
metadata:
  annotations:
    kloudlite.io/last-applied: '{"apiVersion":"v1","kind":"ServiceAccount","metadata":{"name":"default-svc-account","namespace":"kl-gateway-default"}}'
  creationTimestamp: "2024-10-07T06:03:56Z"
  name: default-svc-account
  namespace: kl-gateway-default
  resourceVersion: "1101"
  uid: 99520555-887b-48b7-baae-b72f37e23c81
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  annotations:
    kloudlite.io/last-applied: '{"apiVersion":"rbac.authorization.k8s.io/v1","kind":"ClusterRoleBinding","metadata":{"name":"default-svc-account-rb","namespace":"kl-gateway-default"},"roleRef":{"apiGroup":"rbac.authorization.k8s.io","kind":"ClusterRole","name":"cluster-admin"},"subjects":[{"kind":"ServiceAccount","name":"default-svc-account","namespace":"kl-gateway-default"}]}'
  creationTimestamp: "2024-10-07T06:03:56Z"
  name: default-svc-account-rb
  resourceVersion: "1104"
  uid: 66a02264-61e9-4786-b859-42ad078eaa8c
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: cluster-admin
subjects:
- kind: ServiceAccount
  name: default-svc-account
  namespace: kl-gateway-default

