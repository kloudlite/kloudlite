{{- $name := get . "name"}}
{{- $namespace := get . "namespace"}}
{{- $ownerRefs := get . "ownerRefs" }}

{{- $cloudProvider := get . "cloudProvider"}}
{{- $action := get . "action"}}
{{- $nodeConfig := get . "nodeConfig"}}

{{- $providerConfig := get . "providerConfig"}}

{{- $AwsProvider := get . "AwsProvider"}}
{{- $AzureProvider := get . "AzureProvider"}}
{{- $DoProvider := get . "DoProvider"}}
{{- $GCPProvider := get . "GCPProvider"}}


apiVersion: batch/v1
kind: Job
metadata:
  name: {{$name}}
  namespace: {{$namespace}}
  ownerReferences: {{$ownerRefs | toJson}}
  labels:
    kloudlite.io/is-nodectrl-job: "yes"
spec:
  template:
    spec:
      # nodeSelector:
      #   kloudlite.io/auto-scaler: "true"
      # tolerations:
      # - effect: NoExecute
      #   key: kloudlite.io/auto-scaler
      #   operator: Exists

      restartPolicy: Never
      # serviceAccount: cluster-kloudlite-svc-account
      # serviceAccountName: kloudlite-cluster-svc-account
      # nodeSelector:
      #   kloudlite.io/region: kl-blr1
      containers:

      - image: ghcr.io/kloudlite/platform/apis/nodectrl:v1.0.5-nightly
        name: nodectrl
        imagePullPolicy: Always

        # securityContext:
        #   runAsNonRoot: true
        #   runAsUser: 1000

        #  needed to add after inspection
        #resources:

        env:
        - name: NODE_CONFIG
          value: {{ $nodeConfig }}
        - name: CLOUD_PROVIDER
          value: {{ $cloudProvider }}
        - name: ACTION
          value: {{ $action }}

        - name: NODE_CONFIG
          value: {{ $nodeConfig }}

        - name: PROVIDER_CONFIG
          value: {{ $providerConfig }}

        - name: AWS_PROVIDER_CONFIG
          value: {{ $AwsProvider }}

        - name: AZURE_PROVIDER_CONFIG
          value: {{ $AzureProvider }}
        - name: DO_PROVIDER_CONFIG
          value: {{ $DoProvider }}

        - name: GCP_PROVIDER_CONFIG
          value: {{ $GCPProvider }}


        imagePullPolicy: Always
        resources:
          requests:
            cpu: 150m
            memory: 150Mi
          limits:
            cpu: 400m
            memory: 400Mi
