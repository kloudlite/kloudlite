apiVersion: crds.kloudlite.io/v1
kind: App
metadata:
  name: {{.Values.apps.consoleApi.name}}
  namespace: {{.Release.Namespace}}
spec:
  region: {{.Values.region | default ""}}
  serviceAccount: {{.Values.clusterSvcAccount}}

  {{ include "node-selector-and-tolerations" . | nindent 2 }}
  
  services:
    - port: 80
      targetPort: {{.Values.apps.consoleApi.configuration.httpPort | int}}
      name: http
      type: tcp

    - port: {{.Values.apps.consoleApi.configuration.logsAndMetricsHttpPort | int }}
      targetPort: {{.Values.apps.consoleApi.configuration.logsAndMetricsHttpPort | int }}
      name: http
      type: tcp

  containers:
    - name: main
      image: {{.Values.apps.consoleApi.image}}
      imagePullPolicy: {{.Values.apps.consoleApi.ImagePullPolicy | default .Values.imagePullPolicy }}
      resourceCpu:
        min: "80m"
        max: "150m"
      resourceMemory:
        min: "80Mi"
        max: "150Mi"
      env:
        - key: HTTP_PORT
          value: "{{.Values.apps.consoleApi.configuration.httpPort}}"

        - key: LOGS_AND_METRICS_HTTP_PORT
          value: {{.Values.apps.consoleApi.configuration.logsAndMetricsHttpPort | squote}}
          {{- /* LOGS_AND_METRICS_HTTP_PORT=9999 */}}

        - key: COOKIE_DOMAIN
          value: "{{.Values.cookieDomain}}"

        - key: CONSOLE_DB_URI
          type: secret
          refName: "mres-{{.Values.managedResources.consoleDb}}"
          refKey: URI

        - key: CONSOLE_DB_NAME
          value: {{.Values.managedResources.consoleDb}}

        - key: AUTH_REDIS_HOSTS
          type: secret
          refName: "mres-{{.Values.managedResources.authRedis}}"
          refKey: HOSTS

        - key: AUTH_REDIS_PASSWORD
          type: secret
          refName: "mres-{{.Values.managedResources.authRedis}}"
          refKey: PASSWORD

        - key: AUTH_REDIS_PREFIX
          type: secret
          refName: "mres-{{.Values.managedResources.authRedis}}"
          refKey: PREFIX

        - key: AUTH_REDIS_USERNAME
          type: secret
          refName: "mres-{{.Values.managedResources.authRedis}}"
          refKey: USERNAME

        - key: ACCOUNT_COOKIE_NAME
          value: {{.Values.accountCookieName}}

        - key: CLUSTER_COOKIE_NAME
          value: {{.Values.clusterCookieName}}

        - key: KAFKA_BROKERS
          type: secret
          refName: "{{.Values.secretNames.redpandaAdminAuthSecret}}"
          refKey: KAFKA_BROKERS

        - key: KAFKA_USERNAME
          type: secret
          refName: "{{.Values.secretNames.redpandaAdminAuthSecret}}"
          refKey: USERNAME

        - key: KAFKA_PASSWORD
          type: secret
          refName: "{{.Values.secretNames.redpandaAdminAuthSecret}}"
          refKey: PASSWORD

        - key: KAFKA_STATUS_UPDATES_TOPIC
          value: {{.Values.kafka.topicStatusUpdates}}

        - key: KAFKA_ERROR_ON_APPLY_TOPIC
          value: {{.Values.kafka.topicErrorOnApply}}

        - key: KAFKA_CONSUMER_GROUP_ID
          value: {{.Values.kafka.consumerGroupId}}

        - key: IAM_GRPC_ADDR
          value: {{.Values.apps.iamApi.name}}.{{.Release.Namespace}}.{{.Values.clusterInternalDNS}}:{{.Values.apps.iamApi.configuration.grpcPort}}

        - key: INFRA_GRPC_ADDR
          value: {{.Values.apps.infraApi.name}}.{{.Release.Namespace}}.{{.Values.clusterInternalDNS}}:{{.Values.apps.infraApi.configuration.grpcPort}}

        - key: DEFAULT_PROJECT_WORKSPACE_NAME
          value: {{.Values.defaultProjectWorkspaceName}}

        - key: MSVC_TEMPLATE_FILE_PATH
          value: /console.d/templates/managed-svc-templates.yml

        - key: LOKI_SERVER_HTTP_ADDR
          value: http://{{ (index .Values.helmCharts "loki-stack").name }}.{{.Release.Namespace}}.{{.Values.clusterInternalDNS}}:3100

        - key: PROM_HTTP_ADDR
          value: http://{{ (index .Values.helmCharts "kube-prometheus").name }}-prometheus.{{.Release.Namespace}}.{{.Values.clusterInternalDNS}}:9090

        - key: VPN_DEVICES_MAX_OFFSET
          value: {{.Values.apps.consoleApi.configuration.vpnDevicesMaxOffset | squote}}

        - key: VPN_DEVICES_OFFSET_START
          value: {{.Values.apps.consoleApi.configuration.vpnDevicesOffsetStart | squote}}

        - key: PROM_HTTP_ADDR
          value: http://{{ (index .Values.helmCharts "kube-prometheus").name }}-prometheus.{{.Release.Namespace}}.{{.Values.clusterInternalDNS}}:9090

      volumes:
        - mountPath: /console.d/templates
          type: config
          refName: {{.Values.apps.consoleApi.name}}-managed-svc-template
          items:
            - key: managed-svc-templates.yml

---

apiVersion: v1
kind: ConfigMap
metadata:
  name: {{.Values.apps.consoleApi.name}}-managed-svc-template
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
            - name: db
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
          apiVersion: mongodb-standalone.msvc.kloudlite.io/v1
          kind: Service
          active: true
          fields:
            - name: cpu
              label: CPU Allocation
              inputType: Number
              defaultValue: 0.5
              min: 0.500
              max: 2
              required: true
              unit: vCpu

            - name: size
              label: Capacity in GB
              inputType: Number
              defaultValue: 10
              min: 1
              required: true
              unit: Gi
          inputMiddleware: |-
            const inputMiddleware = (inputs) => {
              return {
                annotation: {
                  "kloudlite.io/billing-plan": "Basic",
                  "kloudlite.io/billable-quantity": `${inputs.cpu}`,
                  "kloudlite.io/is-shared": "false",
                },
                inputs: {
                  resources: {
                    cpu: {
                      min: `${inputs.cpu * 1000}m`,
                      max: `${inputs.cpu * 1000}m`,
                    },
                    memory: `${inputs.cpu * 1000}Mi`,
                  },
                  storage: {
                    size: `${inputs.size}Gi`,
                  }
                },
                error: null,
              };
            }


          estimator: |-
            function (inputs, plans) {
              var computePrice = plans.compute["Basic"].dedicatedPrice * inputs.cpu;
              var storagePrice = plans.storage["Default"].pricePerGB * inputs.size;
              var totalPrice = computePrice + storagePrice;
              var numberOfSupportedConnections = inputs.cpu * 40 * 2;

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
              apiVersion: mongodb-standalone.msvc.kloudlite.io/v1
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
        - apiVersion: mongodb.msvc.kloudlite.io/v1
          kind: StandaloneService
          name: mysql_standalone
          logoUrl: https://img.icons8.com/material-two-tone/344/mysql-logo.png
          apiVersion: mysql-standalone.msvc.kloudlite.io/v1
          kind: Service
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

          inputMiddleware: |-
            const inputMiddleware = (inputs) =>{
              const plan = "Basic"
              return {
                annotation: {
                  "kloudlite.io/billing-plan": plan,
                  "kloudlite.io/billable-quantity": inputs.cpu,
                  "kloudlite.io/is-shared": true,
                },
                inputs: {
                  resources: {
                    cpu: {
                      min: `${inputs.cpu * 1000/2}m`,
                      max: `${inputs.cpu * 1000}m`,
                    },
                    memory: `${inputs.cpu * 1000}Mi`,
                  },
                  storage: {
                    size: `${inputs.size}Gi`,
                  },
                },
                error: null,
              };
            }

          estimator: |-
            function (inputs, plans) {
              const defaultPlan = "Basic"
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
            - name: DSN
              label: Mysql DSN

            - name: HOSTS
              label: Mysql Hosts

            - name: ROOT_PASSWORD
              label: Mysql Root Password

            - name: URI
              label: Mysql Root URI

          resources:
            - apiVersion: mongodb.msvc.kloudlite.io/v1
              kind: Database
              name: db
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


