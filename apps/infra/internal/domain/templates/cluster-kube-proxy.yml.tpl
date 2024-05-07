{{- /*gotype: github.com/kloudlite/api/apps/infra/internal/domain/templates.ClusterKubeProxyVars*/ -}}
{{ with . }}
---
apiVersion: v1
kind: Namespace
metadata:
  name: {{.Namespace | squote}}
---
apiVersion: v1
kind: Secret
metadata:
  name: kloudlite-gvpn-device-config
  namespace: {{.Namespace | squote}}
data:
  wg0.conf: {{.KloudliteDeviceWgConfig | b64enc }}
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: kube-access-globalvpn
  namespace: {{.Namespace | squote}}
spec:
  replicas: 1
  selector:
    matchLabels:
      app: app
  strategy:
    rollingUpdate:
      maxSurge: 25%
      maxUnavailable: 25%
    type: RollingUpdate
  template:
    metadata:
      labels:
        app: app
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
            privileged: true
          volumeMounts:
            - mountPath: /config/wg_confs/wg0.conf
              name: kloudlite-gvpn-device-config
              subPath: wg0.conf
          terminationMessagePath: /dev/termination-log
          terminationMessagePolicy: File
        - image: ghcr.io/kloudlite/hub/socat:latest
          command:
            - sh
            - -c
            - |+
              (socat -dd tcp4-listen:8080,fork,reuseaddr tcp4:kubectl-proxy.{{.Namespace}}.svc.example-test.local:8080 2>&1 | grep -iE --line-buffered 'listening|exiting') &
              pid=$!

              trap "kill -9 $pid" EXIT SIGINT SIGTERM
              wait $pid
          imagePullPolicy: Always
          name: app
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
              - key: wg0.conf
                path: wg0.conf
            secretName: wg-config

---
apiVersion: v1
kind: Service
metadata:
  name: kube-access
  namespace: test-kube-access-globalvpn
spec:
  selector:
    app: app
  ports:
    - name: p-8080
      port: 8080
      targetPort: 8080
      protocol: TCP
---
{{ end }}
