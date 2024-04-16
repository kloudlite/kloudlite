{{- $name := get . "name"}}
{{- $gatewayName := get . "gatewayName" -}}
{{- $namespace := get . "namespace" -}}
{{- $image := get . "image" -}}
{{- $resources := get . "resources" -}}
{{- $ownerRefs := get . "ownerRefs" -}}
{{- $coredns_svc_ip := get . "corednsSvcIp" -}}
{{- $interface := get . "interface" -}}

apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: {{ $name }}
  namespace: {{ $namespace }}
  ownerReferences: {{ $ownerRefs | toJson }}
  labels: &labels
    kloudlite.io/wg-cluster-connection.name: {{ $name }}
    kloudlite.io/wg-cluster-connection.resource: "agent"
spec:
  revisionHistoryLimit: 10
  selector:
    matchLabels: *labels
  template:
    metadata:
      labels: *labels
    spec:
      containers:
      - image: {{ $image }}
        imagePullPolicy: Always
        name: agent
        env:
        - name: WG_INTERFACE
          value: {{ $interface }}
        - name: SERVER_ADDR
          value: http://{{ $gatewayName }}.{{ $namespace }}.svc.cluster.local:3000
        - name: KUBE_DNS_IP
          value: {{ $coredns_svc_ip }}
        - name: MY_IP_ADDRESS
          valueFrom:
            fieldRef:
              fieldPath: status.podIP
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
        - mountPath: /etc/sysctl.conf
          name: sysctl
          subPath: sysctl.conf
      {{- /* priorityClassName: system-cluster-critical */}}
      restartPolicy: Always
      schedulerName: default-scheduler
      hostNetwork: true
      securityContext: {}
      terminationGracePeriodSeconds: 30
      volumes:
      - hostPath:
          path: /lib/modules
          type: Directory
        name: host-volumes
      - name: sysctl
        secret:
          defaultMode: 420
          items:
          - key: sysctl
            path: sysctl.conf
          secretName: {{ $gatewayName }}-configs
