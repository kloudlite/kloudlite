{{- $name := get . "name" -}}
{{- $namespace := get . "namespace" -}}
{{- $image := get . "image" -}}
{{- $svcAccount := get . "svc-account" -}}
{{- $clusterId := get . "cluster-id" -}}
{{- $kafkaSecretName := get . "kafka-secret-name" -}}

{{- $ownerRefs := get . "owner-refs" | default list -}}
{{- $accountRef := get . "account-ref" | default "kl-core" -}}
{{- $region := get . "region" | default "master" -}}
{{- $imagePullPolicy := get . "image-pull-policy" | default "Always" -}}

{{- $nodeSelector := get . "node-selector" | default dict -}}
{{- $tolerations := get . "tolerations" | default list -}}

apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{$name}}
  namespace: {{$namespace}}
  labels:
    app: {{$name}}
spec:
  replicas: 2
  selector:
    matchLabels:
      app: {{$name}}
  template:
    metadata:
      labels:
        app: {{$name}}
    spec:
      nodeSelector: {{ $nodeSelector | toYAML | nindent 8}}
      tolerations: {{ $tolerations | toYAML | nindent 8 }}
      serviceAccountName: {{$svcAccount}}
      containers:
        - name: main
          image: {{$image}}
          imagePullPolicy: {{$imagePullPolicy}}
          env:
            - name: KAFKA_BROKERS
              valueFrom:
                secretKeyRef:
                  name: {{$kafkaSecretName}}
                  key: KAFKA_BROKERS
            - name: KAFKA_CONSUMER_GROUP_ID
              value: "{{$clusterId}}-consumer-agent"
            - name: KAFKA_INCOMING_TOPIC
              value: "{{$clusterId}}-incoming"
            - name: KAFKA_ERROR_ON_APPLY_TOPIC
              value: "${KAFKA_ERROR_ON_APPLY_TOPIC}"
            - name: KAFKA_SASL_USER
              valueFrom:
                secretKeyRef:
                  name: {{$kafkaSecretName}}
                  key: USERNAME
            - name: KAFKA_SASL_PASSWORD
              valueFrom:
                secretKeyRef:
                  name: {{$kafkaSecretName}}
                  key: PASSWORD
          resources:
            requests:
              cpu: 50m
              memory: 100Mi
            limits:
              cpu: 100m
              memory: 200Mi
