{{- $name := get . "name"}}
{{- $namespace := get . "namespace"}}
{{- $image := get . "image"}}
{{- $resources := get . "resources"}}
{{- $server_config := get . "serverConfig"}}
{{- $nodeport := get . "nodeport"}}
{{- $ownerRefs := get . "ownerRefs" -}}
{{- $interface := get . "interface" -}}

apiVersion: apps/v1
kind: Deployment
metadata:
  labels: &labels
    kloudlite.io/wg-cluster-connection.name: {{ $name }}
    kloudlite.io/wg-cluster-connection.resource: "gateway"
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
    spec:
      containers:
      - image: {{ $image }}
        imagePullPolicy: Always
        env:
        - name: WG_INTERFACE
          value: {{ $interface }}
        - name: ADDR
          value: :3000
        - name: CONFIG_PATH
          value: /tmp/server-config.yml
        - name: ENDPOINT
          value: {{ $name }}-external.{{ $namespace }}.svc.cluster.local:51820
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
      dnsPolicy: Default
      priorityClassName: system-cluster-critical
      restartPolicy: Always
      schedulerName: default-scheduler
      securityContext: {}
      terminationGracePeriodSeconds: 30
      tolerations:
      - operator: Exists
      volumes:
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

kind: Secret
metadata:
  name: {{ $name }}-configs
  namespace: {{ $namespace }}
  ownerReferences: {{ $ownerRefs | toJson }}
  labels:
    kloudlite.io/wg-cluster-connection.name: {{ $name }}
    kloudlite.io/wg-cluster-connection.resource: "gateway"
type: Opaque
---

apiVersion: v1
kind: Service
metadata:
  labels: &labels
    kloudlite.io/wg-cluster-connection.name: {{ $name }}
    kloudlite.io/wg-cluster-connection.resource: "gateway"
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
    kloudlite.io/wg-cluster-connection.name: {{ $name }}
    kloudlite.io/wg-cluster-connection.resource: "gateway"
  ownerReferences: {{ $ownerRefs | toJson }}
  name: {{ $name }}-external
  namespace: {{ $namespace }}
spec:
  ports:
  - port: 51820
    protocol: UDP
    targetPort: 51820
    nodePort: {{ $nodeport }}
  selector: *labels
  type: NodePort
