apiVersion: scheduling.k8s.io/v1
kind: PriorityClass
metadata:
  name: kloudlite-critical
value: 999999
globalDefault: false
description: "This priority class should only be used for kloudlite critical applications, like agent, resource watcher etc."

