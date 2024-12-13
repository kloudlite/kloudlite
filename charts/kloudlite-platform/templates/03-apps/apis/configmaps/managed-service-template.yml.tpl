apiVersion: v1
kind: ConfigMap
metadata:
  name: managed-svc-template
  namespace: {{.Release.Namespace}}
data:
managed-svc-templates.yml: |+
  - category: ""
    items:
      - plugin: HelmChart
        meta:
          logo: "https://www.vectorlogo.zone/logos/helmsh/helmsh-icon.svg"
        spec:
          apiVersion: "plugin-helm-chart.kloudlite.github.com/v1"
          services:
            - kind: "HelmChart"
              description: "Manages Helm Lifeycle"
              inputs:
                - input: "chart.url"
                  label: "Helm Chart Repo URL"
                  type: "text"
                  required: true

                - input: "chart.name"
                  label: "Helm Chart Name"
                  type: "text"
                  required: true

                - input: "chart.version"
                  label: "Helm Chart Version"
                  type: "text"
                  required: false

                - input: "helmValues"
                  label: "values.yaml"
                  type: "text/yaml"
                  required: true
                  default: {}

                # - input: "preInstall"
                #   type: "text/bash"
                #   required: true
                #   default: ""
                #
                # - input: "postInstall"
                #   type: "text/bash"
                #   required: false
                #   default: ""
                #
                # - input: "preUninstall"
                #   type: "text/bash"
                #   required: false
                #   default: ""
                #
                # - input: "postUninstall"
                #   type: "text/bash"
                #   required: false
                #   default: ""

      # - kind: "DockerCompose"
      #   description: "create/manage resources directly from a docker-compose file"
      #   inputs:
      #     - input: "nodeSelector"
      #       label: "Node Selector"
      #       description: "Node Selector"
      #       type: "nodeSelector"
      #       required: false
      #
      #     - input: "tolerations"
      #       label: "Tolerations"
      #       description: "Tolerations"
      #       type: array
      #       required: false
      #
      #     - input: "resources.cpu"
      #       label: "CPU"
      #       description: "Allocates specified CPU resources"
      #       type: float-range
      #       required: false
      #
      #     - input: "resources.memory"
      #       label: "Memory"
      #       description: "Allocates specified CPU resources"
      #       type: float-range
      #       required: false
      #
      #     - input: "resources.storage"
      #       label: "Storage"
      #       description: "Allocates specified CPU resources"
      #       type: float-range
      #       required: false

  - category: Databases
    items:
      - plugin: MongoDB
        meta:
          logo: "https://cdn.iconscout.com/icon/free/png-128/mongodb-5-1175140.png"
        spec:
          apiVersion: "plugin-mongodb.kloudlite.github.com/v1"
          services:
            - kind: "StandaloneService"
              description: "creates a single instance mongodb server"
              inputs:
                - input: "nodeSelector"
                  label: "Node Selector"
                  description: "[read more](https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/#label-selectors)"
                  type: "text/yaml"
                  required: false

                - input: "tolerations"
                  label: "Pod Tolerations"
                  description: "[read more](https://kubernetes.io/docs/concepts/scheduling-eviction/taint-and-toleration/#concepts)"
                  type: "text/yaml"
                  required: false

                - input: "resources.cpu"
                  unit: "m"
                  displayUnit: "vCPU"
                  label: "Allocate CPU"
                  description: "Allocates specified CPU resources"
                  type: int-range
                  required: true
                  default: 400
                  min: 200
                  step: 100
                  max: 2000

                - input: "resources.memory"
                  unit: "Mi"
                  displayUnit: "MB"
                  label: "Allocate Memory"
                  description: "Allocates specified CPU resources"
                  type: int-range
                  required: true
                  default: 400
                  min: 200
                  step: 100
                  max: 2000

                - input: "resources.storage"
                  label: "Storage"
                  description: "Allocates specified CPU resources"
                  type: int-range
                  unit: "GiB"
                  displayUnit: "GB"
                  required: true
                  default: 5
                  min: 1
                  step: 1
                  max: 10

              resources:
                - kind: "StandaloneDatabase"
                  description: "creates a mongodb database on standalone service"
                  inputs: []

                # - kind: "StandaloneServiceBackup"
                #   description: ""
                #   inputs: []
                #
                # - kind: "StandaloneDatabaseBackup"
                #   description: ""
                #   inputs: []
                #
  {{- /* managed-svc-templates.yml: |+ #yaml */}}
  {{- /*   - category: db */}}
  {{- /*     displayName: Databases */}}
  {{- /*     items: */}}
  {{- /*       - apiVersion: mongodb.msvc.kloudlite.io/v1 */}}
  {{- /*         kind: StandaloneService */}}
  {{- /*         name: mongo_standalone */}}
  {{- /*         logoUrl: https://img.icons8.com/color/344/mongodb.png */}}
  {{- /*         displayName: MongoDB */}}
  {{- /*         description: MongoDB Server in a standalone instance */}}
  {{- /*         active: true */}}
  {{- /*         fields:  &input-fields */}}
  {{- /*           - name: resources.storage.size */}}
  {{- /*             label: Storage */}}
  {{- /*             inputType: Number */}}
  {{- /*             defaultValue: 1 */}}
  {{- /*             min: 1 */}}
  {{- /*             step: 1 */}}
  {{- /*             max: 1000 */}}
  {{- /*             required: true */}}
  {{- /*             displayUnit: GB */}}
  {{- /*             unit: Gi */}}
  {{- /**/}}
  {{- /*           - name: resources.cpu */}}
  {{- /*             label: Cpu */}}
  {{- /*             inputType: Resource */}}
  {{- /*             defaultValue: 0.4 */}}
  {{- /*             min: 0.4 */}}
  {{- /*             step: 0.1 */}}
  {{- /*             max: 2 */}}
  {{- /*             required: true */}}
  {{- /*             displayUnit: vCPU */}}
  {{- /*             multiplier: 1000 */}}
  {{- /*             unit: "m" */}}
  {{- /**/}}
  {{- /*           - name: resources.memory */}}
  {{- /*             label: Memory */}}
  {{- /*             inputType: Resource */}}
  {{- /*             defaultValue: 0.4 */}}
  {{- /*             min: 0.4 */}}
  {{- /*             step: 0.1 */}}
  {{- /*             max: 2 */}}
  {{- /*             required: true */}}
  {{- /*             multiplier: 1000 */}}
  {{- /*             displayUnit: GB */}}
  {{- /*             unit: "Mi" */}}
  {{- /**/}}
  {{- /*         resources: */}}
  {{- /*           - name: db */}}
  {{- /*             apiVersion: mongodb.msvc.kloudlite.io/v1 */}}
  {{- /*             kind: StandaloneDatabase */}}
  {{- /*             displayName: Database */}}
  {{- /*             description: Creates and Manages a mongodb database and user with proper access to this database */}}
  {{- /**/}}
  {{- /*       - name: postgresql_standalone */}}
  {{- /*         kind: Standalone */}}
  {{- /*         logoUrl: https://upload.wikimedia.org/wikipedia/commons/2/29/Postgresql_elephant.svg */}}
  {{- /*         apiVersion: postgres.msvc.kloudlite.io/v1 */}}
  {{- /*         displayName: Postgres */}}
  {{- /*         description: Postgres Standalond distribution */}}
  {{- /*         active: true */}}
  {{- /*         fields: *input-fields */}}
  {{- /**/}}
  {{- /*         resources: */}}
  {{- /*           - name: db */}}
  {{- /*             apiVersion: postgres.msvc.kloudlite.io/v1 */}}
  {{- /*             kind: StandaloneDatabase */}}
  {{- /*             displayName: Database */}}
  {{- /*             description: Creates and Manages a PostgreSQL database and user with proper access to this database */}}
  {{- /**/}}
  {{- /*       - apiVersion: mysql.msvc.kloudlite.io/v1 */}}
  {{- /*         kind: StandaloneService */}}
  {{- /*         name: mysql_standalone */}}
  {{- /*         logoUrl: https://mariadb.com/wp-content/uploads/2019/11/vertical-logo_black.svg */}}
  {{- /*         displayName: MySQL */}}
  {{- /*         description: MySQL/MariaDB running in standalone fashion */}}
  {{- /**/}}
  {{- /*         active: true */}}
  {{- /**/}}
  {{- /*         fields: *input-fields */}}
  {{- /**/}}
  {{- /*         resources: */}}
  {{- /*           - apiVersion: mysql.msvc.kloudlite.io/v1 */}}
  {{- /*             kind: StandaloneDatabase */}}
  {{- /*             name: db */}}
  {{- /*             displayName: Database */}}
  {{- /*             description: Creates and Manages a MySQL/MariaDB database and user with proper access to this database */}}
  {{- /**/}}
  {{- /*   - category: kv */}}
  {{- /*     displayName: Caches */}}
  {{- /*     logoUrl: https://img.icons8.com/external-others-pike-picture/344/external-cache-data-scientist-worker-others-pike-picture-2.png */}}
  {{- /*     items: */}}
  {{- /*       - apiVersion: redis.msvc.kloudlite.io/v1 */}}
  {{- /*         kind: StandaloneService */}}
  {{- /*         name: redis_standalone */}}
  {{- /*         logoUrl: https://img.icons8.com/color/344/redis.png */}}
  {{- /*         displayName: Redis */}}
  {{- /*         description: Redis running as a standalone database */}}
  {{- /**/}}
  {{- /*         active: true */}}
  {{- /**/}}
  {{- /*         fields: *input-fields */}}
  {{- /**/}}
  {{- /*         resources: [] */}}
  {{- /**/}}
