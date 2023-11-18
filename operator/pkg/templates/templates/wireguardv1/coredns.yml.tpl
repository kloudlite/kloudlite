{{- $configExists := get . "corednsConfigExists"}}
{{- $name := get . "name"}}
{{- $namespace := get . "namespace"}}
{{- $ownerRefs := get . "ownerRefs"}}

{{- if not $configExists }}
apiVersion: v1
data:
  devices: "[]"
  Corefile: |
    .:53 {
        errors
        health
        ready

        forward . 10.96.0.10 
        cache 30
        loop
        reload
        loadbalance
    }
    import /etc/coredns/custom/*.server
kind: ConfigMap
metadata:
  name: coredns
  namespace: {{$namespace}}
  ownerReferences: {{ $ownerRefs| toJson}}

---
{{- end}}

apiVersion: apps/v1
kind: Deployment
metadata:
  name: coredns
  namespace: {{$namespace}}
  ownerReferences: {{ $ownerRefs| toJson}}
spec:
  replicas: 1
  selector:
    matchLabels: &r
      app: dns
  template:
    metadata: 
      labels: *r
    spec:
      containers:
      - args:
        - -conf
        - /etc/coredns/Corefile
        image: rancher/mirrored-coredns-coredns:1.9.1
        imagePullPolicy: IfNotPresent
        livenessProbe:
          failureThreshold: 3
          httpGet:
            path: /health
            port: 8080
            scheme: HTTP
          initialDelaySeconds: 60
          periodSeconds: 10
          successThreshold: 1
          timeoutSeconds: 1
        name: coredns
        ports:
        - containerPort: 53
          name: dns
          protocol: UDP
        - containerPort: 53
          name: dns-tcp
          protocol: TCP
        - containerPort: 9153
          name: metrics
          protocol: TCP
        readinessProbe:
          failureThreshold: 3
          httpGet:
            path: /ready
            port: 8181
            scheme: HTTP
          periodSeconds: 2
          successThreshold: 1
          timeoutSeconds: 1
        resources:
          limits:
            # cpu: 100m
            memory: 170Mi
          requests:
            # cpu: 100m
            memory: 70Mi
        securityContext:
          allowPrivilegeEscalation: false
          capabilities:
            add:
            - NET_BIND_SERVICE
            drop:
            - all
          readOnlyRootFilesystem: true
        terminationMessagePath: /dev/termination-log
        terminationMessagePolicy: File
        volumeMounts:
        - mountPath: /etc/coredns
          name: config-volume
          readOnly: true
        - mountPath: /etc/coredns/custom
          name: custom-config-volume
          readOnly: true
      dnsPolicy: Default
      priorityClassName: system-cluster-critical
      restartPolicy: Always
      schedulerName: default-scheduler
      terminationGracePeriodSeconds: 30
      volumes:
      - configMap:
          defaultMode: 420
          items:
          - key: Corefile
            path: Corefile
          name: coredns
        name: config-volume
      - configMap:
          defaultMode: 420
          name: coredns-custom
          optional: true
        name: custom-config-volume
