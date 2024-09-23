apiVersion: scheduling.k8s.io/v1
kind: PriorityClass
metadata:
  name: stateful
value: 999999999
globalDefault: false
description: "This priority class should be used for stateful applications like db, kafka etc."

