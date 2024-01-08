{{- $releaseName := get . "release-name" }}
{{- $releaseNamespace := get . "release-namespace" }}

{{- $labels := get . "labels" | default dict }}

{{- $ingressClassName := get . "ingress-class-name" }}

{{- $envIngress := "env-ingress" }}

apiVersion: crds.kloudlite.io/v1
kind: HelmChart
metadata:
  name: {{$releaseName}}
  namespace: {{$releaseNamespace}}
  labels: {{$labels}}
spec:
  chartRepoURL: https://kubernetes.github.io/ingress-nginx
  chartName: ingress-nginx
  chartVersion: 4.8.0

  jobVars:
    backOffLimit: 1

    tolerations:
      - operator: Exists

  postInstall: |+
    kubectl apply -f - <<EOF
    apiVersion: v1
    kind: Service
    metadata:
      name: {{$envIngress}}
      namespace: {{$releaseNamespace}}
    spec:
      ports:
      - appProtocol: http
        name: http
        port: 80
        protocol: TCP
        targetPort: http
      - appProtocol: https
        name: https
        port: 443
        protocol: TCP
        targetPort: https
      selector:
        app.kubernetes.io/component: controller
        app.kubernetes.io/instance: {{$releaseName}}
        app.kubernetes.io/name: {{$envIngress}}
      sessionAffinity: None
      type: ClusterIP
    EOF

    {{- /* kubectl delete svc/{{$releaseName}} -n {{$releaseNamespace}} */}}

  values:
    nameOverride: {{$envIngress}}

    rbac:
      create: true

    serviceAccount:
      create: true
      {{- /* name: "{{$envIngress}}-sa" */}}

    controller:
      kind: Deployment
      service:
        type: ClusterIP

      tolerations:
        - operator: Exists
    
      watchIngressWithoutClass: false
      ingressClassByName: true
      ingressClass: {{$ingressClassName}}
      electionID: {{$ingressClassName}}
      ingressClassResource:
        enabled: true
        name: {{$ingressClassName}}
        controllerValue: "k8s.io/{{$ingressClassName}}"

      resources:
        requests:
          cpu: 50m
          memory: 80Mi

      admissionWebhooks:
        enabled: false
        failurePolicy: Ignore

