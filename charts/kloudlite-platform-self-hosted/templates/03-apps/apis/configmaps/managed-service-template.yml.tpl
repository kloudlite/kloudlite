apiVersion: v1
kind: ConfigMap
metadata:
  name: managed-svc-template
  namespace: {{.Release.Namespace}}
data:
  managed-svc-templates.yml: |+ #yaml
    - category: db
      displayName: Databases
      items:
        - apiVersion: mongodb.msvc.kloudlite.io/v1
          kind: StandaloneService
          name: mongo_standalone
          logoUrl: https://img.icons8.com/color/344/mongodb.png
          displayName: MongoDB
          description: MongoDB Server in a standalone instance
          active: true
          fields:  &input-fields
            - name: resources.storage.size
              label: Storage
              inputType: Number
              defaultValue: 1
              min: 1
              step: 1
              max: 1000
              required: true
              displayUnit: GB
              unit: Gi

            - name: resources.cpu
              label: Cpu
              inputType: Resource
              defaultValue: 0.4
              min: 0.4
              step: 0.1
              max: 2
              required: true
              displayUnit: vCPU
              multiplier: 1000
              unit: "m"

            - name: resources.memory
              label: Memory
              inputType: Resource
              defaultValue: 0.4
              min: 0.4
              step: 0.1
              max: 2
              required: true
              multiplier: 1000
              displayUnit: GB
              unit: "Mi"

          resources:
            - name: db
              apiVersion: mongodb.msvc.kloudlite.io/v1
              kind: StandaloneDatabase
              displayName: Database
              description: Creates and Manages a mongodb database and user with proper access to this database

        - name: postgresql_standalone
          kind: Standalone
          logoUrl: https://upload.wikimedia.org/wikipedia/commons/2/29/Postgresql_elephant.svg
          apiVersion: postgres.msvc.kloudlite.io/v1
          displayName: Postgres
          description: Postgres Standalond distribution
          active: true
          fields: *input-fields

          resources:
            - name: db
              apiVersion: postgres.msvc.kloudlite.io/v1
              kind: StandaloneDatabase
              displayName: Database
              description: Creates and Manages a PostgreSQL database and user with proper access to this database

        - apiVersion: mysql.msvc.kloudlite.io/v1
          kind: StandaloneService
          name: mysql_standalone
          logoUrl: https://mariadb.com/wp-content/uploads/2019/11/vertical-logo_black.svg
          displayName: MySQL
          description: MySQL/MariaDB running in standalone fashion

          active: true

          fields: *input-fields

          resources:
            - apiVersion: mysql.msvc.kloudlite.io/v1
              kind: StandaloneDatabase
              name: db
              displayName: Database
              description: Creates and Manages a MySQL/MariaDB database and user with proper access to this database

    - category: kv
      displayName: Caches
      logoUrl: https://img.icons8.com/external-others-pike-picture/344/external-cache-data-scientist-worker-others-pike-picture-2.png
      items:
        - apiVersion: redis.msvc.kloudlite.io/v1
          kind: StandaloneService
          name: redis_standalone
          logoUrl: https://img.icons8.com/color/344/redis.png
          displayName: Redis
          description: Redis running as a standalone database

          active: true

          fields: *input-fields

          resources: []

