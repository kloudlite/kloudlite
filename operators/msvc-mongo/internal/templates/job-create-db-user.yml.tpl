{{- $jobName := get . "job-name" }} 
{{- $jobNamespace := get . "job-namespace" }} 
{{- $labels := get . "labels" | default dict }} 
{{- $ownerRefs := get . "owner-refs" |default list }}

{{- $backoffLimit := get . "backoff-limit" | default 1 }} 

{{- $dbAdminUri := get . "db-admin-uri" }}
{{- $dbname := get . "dbname" }}
{{- $username := get . "username" }}
{{- $password := get . "password" }}

---
apiVersion: batch/v1
kind: Job
metadata:
  name: {{$jobName}}
  namespace: {{$jobNamespace}}
  labels: {{$labels | toYAML | nindent 4}}
  ownerReferences: {{$ownerRefs | toYAML| nindent 4}}
spec:
  template:
    metadata:
      annotations:
        kloudlite.io/job_name: {{$jobName}}
        kloudlite.io/job_type: "mongodb-create-user"
    spec:
      containers:
      - name: main
        image: docker.io/alpine/mongosh:2.0.2
        imagePullPolicy: Always
        command:
          - bash
          - -c
          - |+
            set -o pipefail
            cat > create-user.js <<EOF
            db.getSiblingDB("{{$dbname}}").createUser({
              user: "{{$username}}", 
              pwd: "{{$password}}", 
              roles: [
                { role: 'dbAdmin', db: "{{$dbName}}" },
                { role: 'readWrite', db: "{{$dbName}}" },
              ]
            })
            EOF
            mongosh "{{$dbAdminUri}}" create-user.js

      restartPolicy: Never
  backoffLimit: {{$backoffLimit | int}}
