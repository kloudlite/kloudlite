{{- $name := get . "name"}}
{{- $namespace := get . "namespace"}}
{{- $image := get . "image"}}
{{- $corednsImage := get . "coredns-image"}}
{{- $resources := get . "resources"}}
{{- $server_config := get . "serverConfig"}}
{{- /* {{- $nodeport := get . "nodeport"}} */}}
{{- $ownerRefs := get . "ownerRefs" -}}
{{- $interface := get . "interface" -}}
{{- $corefile := get . "corefile" }}

apiVersion: apps/v1
kind: Deployment
metadata:
  labels: &labels
    kloudlite.io/wg-global-vpn.name: {{ $name }}
    kloudlite.io/wg-global-vpn.resource: "gateway"
  ownerReferences: {{ $ownerRefs | toJson }}
  name: {{ $name }}
  namespace: {{ $namespace }}
spec:
  progressDeadlineSeconds: 600
  revisionHistoryLimit: 10
  selector:
    matchLabels: *labels
  strategy:
    type: Recreate
  template:
    metadata:
      labels: *labels
      annotations:
        kloudlite.io/server-config-hash: {{printf "%s.%s.%s" $server_config $corefile | sha256sum}}
    spec:
      containers:
      - name: coredns
        image: ghcr.io/kloudlite/operator/components/coredns:v1.0.5-nightly
        args:
        - --addr
        - 0.0.0.0:17171
        - --corefile
        - /etc/coredns/Corefile
        - --debug
        imagePullPolicy: IfNotPresent
        resources:
          limits:
            # cpu: 100m
            memory: 20Mi
          requests:
            # cpu: 100m
            memory: 20Mi
        securityContext:
          allowPrivilegeEscalation: false
          capabilities:
            add:
            - NET_BIND_SERVICE
        terminationMessagePath: /dev/termination-log
        terminationMessagePolicy: File
        volumeMounts:
        - mountPath: /etc/coredns
          name: gateway-dns-config
          readOnly: true

      - image: {{ $image }}
        imagePullPolicy: Always
        env:
        - name: WG_INTERFACE
          value: {{ $interface }}
        {{- /* FIXME: might need to change it to something else */}}
        - name: AGENT_CIDR
          value: "10.43.0.0/21"
        - name: ADDR
          value: :3000
        - name: CONFIG_PATH
          value: /tmp/server-config.yml
        - name: ENDPOINT
          value: {{ $name }}-external.{{ $namespace }}.svc.cluster.local:31820
        name: gateway
        ports:
        - containerPort: 51820
          protocol: UDP
          name: wireguard
        - containerPort: 3000
          protocol: TCP
        resources: {{ $resources | toJson }}
        securityContext:
          capabilities:
            add:
            - NET_ADMIN
            - SYS_MODULE
          privileged: true
        terminationMessagePath: /dev/termination-log
        terminationMessagePolicy: File
        volumeMounts:
        - mountPath: /lib/modules
          name: host-volumes
        - mountPath: /tmp/server-config.yml
          name: gateway-configs
          subPath: server-config.yml
        - mountPath: /etc/sysctl.conf
          name: sysctl
          subPath: sysctl.conf
        - mountPath: /etc/coredns
          name: gateway-dns-config
          readOnly: true
      dnsPolicy: Default
      restartPolicy: Always
      schedulerName: default-scheduler
      securityContext: {}
      terminationGracePeriodSeconds: 30
      tolerations:
      - operator: Exists
      volumes:
      - name: gateway-dns-config
        secret:
          defaultMode: 420
          items:
          - key: Corefile
            path: Corefile
          secretName: {{ $name }}-configs
      - name: sysctl
        secret:
          defaultMode: 420
          items:
          - key: sysctl
            path: sysctl.conf
          secretName: {{ $name }}-configs
      - name: gateway-configs
        secret:
          defaultMode: 420
          items:
          - key: server-config
            path: server-config.yml
          secretName: {{ $name }}-configs
      - hostPath:
          path: /lib/modules
          type: Directory
        name: host-volumes
---

apiVersion: v1
stringData:
  server-config: |+
    {{ $server_config | nindent 4 }}
  sysctl: net.ipv4.ip_forward=1
  Corefile: |+
    {{ $corefile | nindent 4 }}
kind: Secret
metadata:
  name: {{ $name }}-configs
  namespace: {{ $namespace }}
  ownerReferences: {{ $ownerRefs | toJson }}
  labels:
    kloudlite.io/wg-global-vpn.name: {{ $name }}
    kloudlite.io/wg-global-vpn.resource: "gateway"
type: Opaque
---

apiVersion: v1
kind: Service
metadata:
  labels: &labels
    kloudlite.io/wg-global-vpn.name: {{ $name }}
    kloudlite.io/wg-global-vpn.resource: "gateway"
  ownerReferences: {{ $ownerRefs | toJson }}
  name: {{ $name }}
  namespace: {{ $namespace }}
spec:
  ports:
  - port: 3000
    targetPort: 3000
  selector: *labels

--- 
apiVersion: v1
kind: Service
metadata:
  labels: &labels
    kloudlite.io/wg-global-vpn.name: {{ $name }}
    kloudlite.io/wg-global-vpn.resource: "gateway"
  ownerReferences: {{ $ownerRefs | toJson }}
  name: {{ $name }}-external
  namespace: {{ $namespace }}
spec:
  ports:
  - port: 31820
    protocol: UDP
    name: "wireguard"
    targetPort: 51820
    {{- /* nodePort: {{ $nodeport }} */}}
  selector: *labels
  {{- /* type: NodePort */}}
  type: LoadBalancer
