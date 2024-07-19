apiVersion: scheduling.k8s.io/v1
kind: PriorityClass
metadata:
  name: {{ include "priority-class.name" .}}
value: {{ include "priority-class.value" .}}
description: "This priority class is for kloudlite nodepools. it's value is higher, because it is criticial for scheduling/functioning of your apps"

