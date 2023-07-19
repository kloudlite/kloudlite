{{- $name := get . "name"}}
{{- $isMaster := get . "isMaster"}}


apiVersion: apps/v1
kind: Deployment
metadata:
  name: "wireguard-deployment"
  annotations:
    kloudlite.io/account.name: {{ $name }}
  labels:
    kloudlite.io/wg-deployment: "true"
    kloudlite.io/wg-server.name: {{ $name }}
  namespace: "wg-{{ $name }}"
spec:
  replicas: 1
  strategy:
    type: Recreate
  selector:
    matchLabels:
      app: "wireguard"
  template:
    metadata:
      # annotations:
      #   linkerd.io/inject: disabled
      labels:
        app: "wireguard"
        kloudlite.io/wg-pod: "true"
    spec:
      {{- if $isMaster }}
      nodeSelector:
        node-role.kubernetes.io/master: "true"
      {{ end }}
      containers:
      - name: proxy
        imagePullPolicy: IfNotPresent
        image: ghcr.io/kloudlite/platform/apis/wg-proxy:v1.0.5-nightly
        env:
          - name: CONFIG_FILE
            value: /proxy-config/config.json

        volumeMounts:
          - mountPath: /proxy-config
            name: config-path

      - name: wireguard
        # image: ghcr.io/linuxserver/wireguard
        imagePullPolicy: IfNotPresent
        image: ghcr.io/kloudlite/platform/apis/wg-restart:v1.0.5-nightly
        securityContext:
          capabilities:
            add:
              - NET_ADMIN
              - SYS_MODULE
          privileged: true
        volumeMounts:
          - name: wg-config
            mountPath: /etc/wireguard/wg0.conf
            subPath: wg0.conf
          - name: host-volumes
            mountPath: /lib/modules
        ports:
        - containerPort: 51820
          protocol: UDP
        resources:
          requests:
            memory: 64Mi
            # cpu: "100m"
          limits:
            memory: "128Mi"
            # cpu: "200m"
      volumes:
        - name: wg-config
          secret:
            secretName: wg-server-config
            items:
              - key: data
                path: wg0.conf
        - name: host-volumes
          hostPath:
            path: /lib/modules
            type: Directory

        - name: config-path
          configMap:
            name: "device-proxy-config"
            items:
              - key: config.json
                path: config.json
---

kind: Service
apiVersion: v1
metadata:
  annotations:
    kloudlite.io/account.name: {{$name}}
  labels:
    k8s-app: wireguard
    kloudlite.io/wg-service: "true"
    kloudlite.io/wg-server.name: {{ $name }}
  name: "wireguard-service"
  namespace: "wg-{{$name}}"
spec:
  type: NodePort
  ports:
    - port: 51820
      protocol: UDP
      targetPort: 51820
  selector:
    app: "wireguard"

---

kind: Service
apiVersion: v1
metadata:
  annotations:
    kloudlite.io/account.name: {{$name}}
  labels:
    kloudlite.io/proxy-api: "true"
    kloudlite.io/wg-server.name: {{ $name }}
  name: "wg-api-service"
  namespace: "wg-{{$name}}"
spec:
  ports:
    - port: 2999
      name: proxy-restart
    - port: 2998
      name: wg-restart
  selector:
    app: "wireguard"
