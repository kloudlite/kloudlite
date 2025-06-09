apiVersion: v1
kind: Namespace
metadata:
  name: wg-{{ .Name }}
  ownerReferences:
    - apiVersion: {{.APIVersion}}
      kind: {{.Kind}}
      name: {{.Name}}
      uid: {{.UID}}
      controller: true
      blockOwnerDeletion: true
  labels:
    accountId: {{ .Name }}
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: wireguard-deployment
  namespace: wg-{{ .Name }}
spec:
  replicas: 1
  strategy:
    type: Recreate
  selector:
    matchLabels:
      app: wireguard
  template:
    metadata:
      labels:
        app: wireguard
    spec:
      nodeName: master
      containers:
        - name: wireguard
          image: ghcr.io/linuxserver/wireguard
          command:
            - bash
            - -c
            - tail -f /dev/null
          envFrom:
          # - configMapRef:
          #     name: wireguard-configmap
          securityContext:
            capabilities:
              add:
                - NET_ADMIN
                - SYS_MODULE
            privileged: true
          volumeMounts:
            - name: wg-config
              mountPath: /config
            - name: host-volumes
              mountPath: /lib/modules
          ports:
            - containerPort: 51820
              protocol: UDP
          resources:
            requests:
              memory: "64Mi"
              cpu: "100m"
            limits:
              memory: "128Mi"
              cpu: "200m"
      volumes:
        - name: wg-config
          hostPath:
            path: /wg-config/{{ .Name }}
            type: DirectoryOrCreate
        - name: host-volumes
          hostPath:
            path: /lib/modules
            type: Directory

---

kind: Service
apiVersion: v1
metadata:
  labels:
    k8s-app: wireguard
  name: wireguard-service
  namespace: wg-{{ .Name }}
spec:
  type: NodePort
  ports:
    - port: 51820
      protocol: UDP
      targetPort: 51820
  selector:
    app: wireguard
