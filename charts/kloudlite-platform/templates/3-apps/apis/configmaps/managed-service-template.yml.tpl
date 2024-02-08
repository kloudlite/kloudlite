apiVersion: v1
kind: ConfigMap
metadata:
  name: managed-svc-template
  namespace: {{.Release.Namespace}}
data:
  managed-svc-templates.yml: |+
    - category: db
      displayName: Databases
      items:
        - apiVersion: mongodb.msvc.kloudlite.io/v1
          kind: ClusterService
          name: mongo_cluster
          logoUrl: https://img.icons8.com/color/344/mongodb.png
          displayName: MongoDB cluster
          description: MongoDB cluster
          active: false
          fields:
            - name: capacity
              label: Capacity in GB
              inputType: Number
              defaultValue: 10
              min: 1
              required: true
              unit: Gi

            - name: primary_count
              label: Mongo Primary Count
              inputType: Number
              defaultValue: 1
              min: 1
              unit: Node

            - name: secondary_count
              label: Mongo Secondary Count
              inputType: Number
              defaultValue: 1
              min: 1
              unit: Node
          outputs:
            - name: ROOT_PASSWORD
              label: MongoRoot Password
            - name: HOSTS
              description: DB Hosts
            - name: DB_URL
              description: DB URL
            - name: JSON
              description: Configuration JSON
          resources:
            - apiVersion: mongodb.msvc.kloudlite.io/v1
              kind: Database
              name: db
              displayName: Database
              description: MongoDB
              fields:
                - name: name
                  label: Database Name
                  inputType: String
                  required: true
              outputs:
                - name: DB_USER
                  label: Username
                - name: DB_PASSWORD
                  label: Password
                - name: DB_URL
                  label: Connection String

        - name: mongo_standalone
          logoUrl: https://img.icons8.com/color/344/mongodb.png
          displayName: MongoDB Standalone
          description: MongoDB Standalone
          apiVersion: mongodb.msvc.kloudlite.io/v1
          kind: StandaloneService
          active: true
          fields:
            - name: resources.storage.size
              label: Storage
              inputType: Number
              defaultValue: 0.4
              min: 0.1
              step: 0.1
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

          outputs:
            - name: ROOT_PASSWORD
              label: MongoRoot Password
            - name: HOSTS
              description: DB Hosts
            - name: DB_URL
              description: DB URL
            - name: JSON
              description: Configuration JSON

          resources:
            - name: db
              apiVersion: mongodb.msvc.kloudlite.io/v1
              kind: Database
              displayName: Database
              description: MongoDB
              # fields:
              #   - name: name
              #     label: Database Name
              #     inputType: String
              #     required: true
              outputs:
                - name: DB_USER
                  label: Username
                - name: DB_PASSWORD
                  label: Password
                - name: DB_URL
                  label: Connection String

        - name: mysql_standalone
          kind: StandaloneService
          logoUrl: https://img.icons8.com/material-two-tone/344/mysql-logo.png
          apiVersion: mysql-standalone.msvc.kloudlite.io/v1
          displayName: MySQL Standalone
          description: MySQL Standalone
          fields:
            - name: size
              label: Capacity in GB
              inputType: Number
              defaultValue: 5
              min: 1
              max: 99999
              step: 0.1
              required: true
              unit: Gi

            - name: cpu
              label: Cpu
              inputType: Number
              defaultValue: 0.4
              min: 0.4
              step: 0.1
              max: 2
              required: true
              unit: vCpu

          outputs:
            - name: DSN
              label: Mysql DSN

            - name: HOSTS
              label: Mysql Hosts

            - name: ROOT_PASSWORD
              label: Mysql Root Password

            - name: URI
              label: Mysql Root URI

          resources:
            - name: db
              apiVersion: mysql-standalone.msvc.kloudlite.io/v1
              kind: Database
              displayName: Database
              description: MysqlDB
              # fields:
              #   - name: name
              #     label: Db Name
              #     inputType: String
              #     required: true
              outputs:
                - name: DB_NAME
                  label: Db Name

                - name: DSN
                  label: Mysql DSN

                - name: HOSTS
                  label: DB Hosts

                - name: URI
                  label: DB Uri

                - name: PASSWORD
                  label: Db password

                - name: USERNAME
                  label: DB Username

          active: true

        - apiVersion: mysql.msvc.kloudlite.io/v1
          kind: ClusterService
          name: mysql_cluster
          logoUrl: https://img.icons8.com/material-two-tone/344/mysql-logo.png
          displayName: MySQL Cluster
          description: MySQL Cluster
          active: false

        - name: object_storage
          displayName: Object Storage
          apiVersion: s3.aws.kloudlite.io/v1
          kind: Bucket
          description: S3 compatible object storage
          logoUrl: https://k21academy.com/wp-content/uploads/2021/07/Google-Cloud-Storage.png
          fields:
            # - name: name
            #   label: Bucket Name
            #   inputType: String
            #   required: true
            - name: region
              label: Bucket Region
              inputType: String
              default: ap-south-1
              required: true
            - name: publicFolders
              label: Bucket Public Folders
              inputType: String
              required: false

          inputMiddleware: |+
            const inputMiddleware =  (inputs) => {
              return {
                annotation: {},
                inputs: inputs,
                error: null,
              };
            }

          estimator: |+
            function (inputs, plans) {
              var storagePrice = plans.storage["Default"].pricePerGB * inputs.size;
              return {
                totalPrice: storagePrice,
                error: null,
                properties: [],
              };
            }

          output:
            - name: AWS_ACCESS_KEY_ID
              label: AWS Access Key ID
            - name: AWS_SECRET_ACCESS_KEY
              label: AWS Secret Access Key
            - name: AWS_REGION
              label: AWS Region
            - name: EXTERNAL_BUCKET_HOST
              label: External Bucket Host
            - name: INTERNAL_BUCKET_HOST
              label: Internal Bucket Host

    - category: cache
      displayName: Caches
      logoUrl: https://img.icons8.com/external-others-pike-picture/344/external-cache-data-scientist-worker-others-pike-picture-2.png
      items:
        - apiVersion: redis.msvc.kloudlite.io/v1
          kind: ClusterService
          name: redis_cluster
          logoUrl: https://img.icons8.com/color/344/redis.png
          displayName: Redis Cluster
          description: Redis Cluster
          active: false
        - apiVersion: redis.msvc.kloudlite.io/v1
          kind: StandaloneService
          name: redis_standalone
          logoUrl: https://img.icons8.com/color/344/redis.png
          displayName: Redis Standalone
          description: Redis Standalone
          fields:
            - name: cpu
              inputType: Number
              required: true
              min: 0.2
              max: 2
              default: 0.2
              unit: vCpu

            - name: size
              label: Capacity in GB
              inputType: Number
              required: true
              default: 5
              min: 1
              unit: Gi

          inputMiddleware: |+
            const inputMiddleware = (inputs) => {
              const defaultPlan = "General"
              return {
                annotation: {
                  "kloudlite.io/billing-plan": defaultPlan,
                  "kloudlite.io/billable-quantity": inputs.cpu,
                  "kloudlite.io/is-shared": true,
                },
                inputs: {
                  resources: {
                    cpu: {
                      min: `${inputs.cpu * 1000/2}m`,
                      max: `${inputs.cpu * 1000}m`,
                    },
                    memory: `${inputs.cpu * 2 * 1000}Mi`,
                  },
                  storage:{
                    size: `${inputs.size}Gi`,
                  },
                },
                error: null,
              };
            }

          estimator: |+
            function (inputs, plans) {
              const defaultPlan = "General"
              var computePrice = plans.compute[defaultPlan].sharedPrice * inputs.cpu;
              var storagePrice = plans.storage["Default"].pricePerGB * inputs.size;
              var totalPrice = computePrice + storagePrice;
              var numberOfSupportedConnections = inputs.cpu * 40;
              return {
                totalPrice: totalPrice,
                error: null,
                properties: [
                    {
                      name: "capacity",
                      items:[
                          "Supports around "  + numberOfSupportedConnections + " connections",
                      ]
                    }
                ]
              };
            }

          outputs:
            - name: HOSTS
              label: Redis Hosts

            - name: ROOT_PASSWORD
              label: Redis Root Password

            - name: URI
              label: Redis ROOT Uri
          resources:
            - apiVersion: redis.msvc.kloudlite.io/v1
              kind: ACLAccount
              name: ACLAccount
              fields:
                # - name: name
                #   label: ACL Account Name

                - name: prefix
                  label: Redis Key Prefix
              outputs:
                - name: HOSTS
                  label: Redis Hosts

                - name: PREFIX
                  label: Redis Prefix

                - name: URI
                  label: Redis Connection URI

                - name: USERNAME
                  label: Redis User

                - name: PASSWORD
                  label: Redis User Password

          active: true
        - name: memcached_cluster
          logoUrl: https://upload.wikimedia.org/wikipedia/en/thumb/2/27/Memcached.svg/200px-Memcached.svg.png
          displayName: Memcached Cluster
          description: Memcached Cluster
          active: false
        - name: memcached_standalone
          logoUrl: https://upload.wikimedia.org/wikipedia/en/thumb/2/27/Memcached.svg/200px-Memcached.svg.png
          displayName: Memcached Standalone
          description: Memcached Standalone
          active: false

    - category: messaging
      displayName: Messaging
      list:
        - name: kafka_cluster
          logoUrl: https://upload.wikimedia.org/wikipedia/commons/thumb/0/0a/Apache_kafka-icon.svg/1200px-Apache_kafka-icon.svg.png
          displayName: Kafka Cluster
          description: Kafka Cluster
          active: false
        - name: rabbitmq_cluster
          logoUrl: https://pbs.twimg.com/profile_images/1223261138059780097/eH73w5lN_400x400.jpg
          displayName: RabbitMQ Cluster
          description: RabbitMQ Cluster
          active: false



