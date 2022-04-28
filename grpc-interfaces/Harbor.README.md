-- projects
  {
    "chart_count": 0,
    "creation_time": "2022-04-14T08:57:56.686Z",
    "current_user_role_id": 1,
    "current_user_role_ids": [
      1
    ],
    "cve_allowlist": {
      "creation_time": "0001-01-01T00:00:00.000Z",
      "id": 14,
      "items": [],
      "project_id": 17,
      "update_time": "0001-01-01T00:00:00.000Z"
    },
    "metadata": {
      "public": "false"
    },
    "name": "ci",
    "owner_id": 3,
    "owner_name": "nxtcoder17",
    "project_id": 17,
    "repo_count": 3,
    "update_time": "2022-04-14T08:57:56.686Z"
  },


-- robot accounts
[
  {
    "creation_time": "2022-04-20T11:12:59.615Z",
    "description": "kloudlite operator uses these to verify images exist or not, and these secrets are added to the image pull secrets on project creation",
    "disable": false,
    "duration": -1,
    "editable": true,
    "expires_at": -1,
    "id": 243,
    "level": "system",
    "name": "robot_kloudlite-operator",
    "permissions": [
      {
        "access": [
          {
            "action": "push",
            "resource": "repository"
          },
          {
            "action": "pull",
            "resource": "repository"
          },
          {
            "action": "delete",
            "resource": "artifact"
          },
          {
            "action": "read",
            "resource": "helm-chart"
          },
          {
            "action": "create",
            "resource": "helm-chart-version"
          },
          {
            "action": "delete",
            "resource": "helm-chart-version"
          },
          {
            "action": "create",
            "resource": "tag"
          },
          {
            "action": "delete",
            "resource": "tag"
          },
          {
            "action": "create",
            "resource": "artifact-label"
          },
          {
            "action": "create",
            "resource": "scan"
          },
          {
            "action": "stop",
            "resource": "scan"
          },
          {
            "action": "list",
            "resource": "artifact"
          },
          {
            "action": "list",
            "resource": "repository"
          }
        ],
        "kind": "project",
        "namespace": "ci"
      }
    ],
    "update_time": "2022-04-20T11:12:59.615Z"
  }
]
