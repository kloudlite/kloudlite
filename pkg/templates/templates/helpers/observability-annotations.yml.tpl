{{- define "observability-annotations" -}}

{{- $resourceType := get . "resource-type" }} 
{{- $resourceName := get . "resource-name" }} 
{{- $resourceComponent := get . "resource-component" }} 

{{- $workspaceName := get . "workspace-name" }} 
{{- $workspaceTargetNs := get . "workspace-target-ns" }} 

{{- $projectName := get . "project-name" }} 
{{- $projectTargetNs := get . "project-target-ns" }} 

{{- if not $resourceType -}}
{{- fail "resource-type must be provided, when using template: observability-annotations" -}}
{{- end -}}

{{- if not $resourceName -}}
{{- fail "resource-name must be provided, when using template: observability-annotations" -}}
{{- end -}}

{{- if not $resourceComponent -}}
{{- fail "resource-component must be provided, when using template: observability-annotations" -}}
{{- end -}}

kloudlite.io/resource_name: {{$resourceName}}
kloudlite.io/resource_component: {{$resourceComponent}}
kloudlite.io/resource_type: {{$resourceType}}

{{- if $workspaceName }}
kloudlite.io/workspace_name: "{{$workspaceName}}"
{{- end }}

{{- if $workspaceTargetNs }}
kloudlite.io/workspace_target_ns: "{{$workspaceTargetNs}}"
{{- end }}

{{- if $projectName}}
kloudlite.io/project_name: "{{$projectName}}"
{{- end }}

{{- if $projectTargetNs}}
kloudlite.io/project_target_ns: "{{$projectTargetNs}}"
{{- end }}

{{- end -}}
