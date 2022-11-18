```json5
// minimum docker push/pull permissions
[
  {
    "action":   "push",
    "resource": "repository",
  },
  {
    "action":   "pull",
    "resource": "repository",
  },
]
```

```json5
// harbor account ACL example
[
  {
    "action":   "push",
    "resource": "repository",
  },
  {
    "action":   "pull",
    "resource": "repository",
  },
  {
    "action":   "delete",
    "resource": "artifact",
  },
  {
    "action":   "create",
    "resource": "helm-chart-version",
  },
  {
    "action":   "delete",
    "resource": "helm-chart-version",
  },
  {
    "action":   "create",
    "resource": "tag",
  },
  {
    "action":   "delete",
    "resource": "tag",
  },
  {
    "action":   "create",
    "resource": "artifact-label",
  },
  {
    "action":   "list",
    "resource": "artifact",
  },
  {
    "action":   "list",
    "resource": "repository",
  },
]
```
