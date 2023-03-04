apiVersion: crds.kloudlite.io/v1
kind: App
metadata:
  name: {{.Values.apps.consoleApi.name}}
  namespace: {{.Release.Namespace}}
  annotations:
    kloudlite.io/account-ref: {{.Values.accountName}}
spec:
  region: {{.Values.region}}
  serviceAccount: {{.Values.clusterSvcAccount}}
  services:
    - port: 80
      targetPort: 3000
      name: http
      type: tcp

    - port: 3001
      targetPort: 3001
      name: grpc
      type: tcp

    - port: 3002
      targetPort: 8192
      name: log
      type: tcp

    - port: 9191
      targetPort: 9191
      name: metrics
      type: tcp

  containers:
    - name: main
      image: {{.Values.apps.consoleApi.image}}
      imagePullPolicy: {{.Values.apps.consoleApi.ImagePullPolicy | default .Values.imagePullPolicy }}
      resourceCpu:
        min: "100m"
        max: "200m"
      resourceMemory:
        min: "100Mi"
        max: "200Mi"

      env:
        - key: PORT
          value: "3000"

        - key: GRPC_PORT
          value: '3001'
      
        - key: LOG_PORT
          value: "8192"

        - key: METRICS_HTTP_PORT
          value: "9191"

        - key: COOKIE_DOMAIN
          value: "{{.Values.cookieDomain}}"

        - key: IMAGE_REGISTRY_PREFIX
          value: registry.kloudlite.io
          
        - key: NOTIFIER_URL
          value: "http://{{.Values.apps.socketWeb.name}}.{{.Release.Namespace}}.svc.cluster.local:3001"

        - key: LOKI_URL
          value: loki-external.REPLACE_ME.clusters.{{.Values.baseDomain}}

        - key: LOKI_AUTH_PASSWORD
          value: ""

        - key: METRICS_HTTP_CORS
          value: "https://console.{{.Values.baseDomain}}"

        - key: PROMETHEUS_ENDPOINT
          value: https://prom-external.REPLACE_ME.clusters.{{.Values.baseDomain}}

        - key: PROMETHEUS_BASIC_AUTH_PASSWORD
          value: ""

        - key: KAFKA_WORKLOAD_STATUS_TOPIC
          value: kl-status-updates

        - key: KAFKA_GROUP_ID
          value: control-plane

        - key: MONGO_DB_NAME
          value: {{.Values.managedResources.consoleDb}}

        - key: REDIS_HOSTS
          type: secret
          refName: "mres-{{.Values.managedResources.consoleRedis}}"
          refKey: HOSTS

        - key: REDIS_PASSWORD
          type: secret
          refName: "mres-{{.Values.managedResources.consoleRedis}}"
          refKey: PASSWORD

        - key: REDIS_PREFIX
          type: secret
          refName: "mres-{{.Values.managedResources.consoleRedis}}"
          refKey: PREFIX

        - key: REDIS_USERNAME
          type: secret
          refName: "mres-{{.Values.managedResources.consoleRedis}}"
          refKey: USERNAME

        - key: REDIS_AUTH_HOSTS
          type: secret
          refName: "mres-{{.Values.managedResources.authRedis}}"
          refKey: HOSTS

        - key: REDIS_AUTH_PASSWORD
          type: secret
          refName: "mres-{{.Values.managedResources.authRedis}}"
          refKey: PASSWORD

        - key: REDIS_AUTH_PREFIX
          type: secret
          refName: "mres-{{.Values.managedResources.authRedis}}"
          refKey: PREFIX

        - key: REDIS_AUTH_USERNAME
          type: secret
          refName: "mres-{{.Values.managedResources.authRedis}}"
          refKey: USERNAME

        - key: MONGO_URI
          type: secret
          refName: "mres-{{.Values.managedResources.consoleDb}}"
          refKey: URI

        - key: MANAGED_TEMPLATES_PATH
          value: /console.d/templates/managed-svc-templates.yml

        - key: FINANCE_SERVICE
          value: {{.Values.apps.financeApi.name}}.{{.Release.Namespace}}.svc.cluster.local:3001

        - key: AUTH_SERVICE
          value: {{.Values.apps.authApi.name}}.{{.Release.Namespace}}.svc.cluster.local:3001

        - key: CI_SERVICE
          value: {{.Values.apps.ciApi.name}}.{{.Release.Namespace}}.svc.cluster.local:3001

        - key: IAM_SERVICE
          value: {{.Values.apps.iamApi.name}}.{{.Release.Namespace}}.svc.cluster.local:3001

        - key: DNS_SERVICE
          value: {{.Values.apps.dnsApi.name}}.{{.Release.Namespace}}.svc.cluster.local:3001

        - key: KAFKA_BOOTSTRAP_SERVERS
          type: secret
          refName: "{{.Values.redpandaAdminSecretName}}"
          refKey: KAFKA_BROKERS


        - key: KAFKA_WORKLOAD_STATUS_TOPIC
          value: kl-status-updates

        - key: KAFKA_GROUP_ID
          value: control-plane

        - key: COMPUTE_PLANS_PATH
          value: /console.d/templates/compute-plans.yaml

        - key: INVENTORY_PATH
          value: /console.d/templates

        - key: JSEVAL_SERVICE
          value: {{.Values.apps.jsEvalApi.name}}:3001

        - key: KUBE_API_ADDRESS
          value: http://127.0.0.1:2999

        - key: KAFKA_USERNAME
          type: secret
          refName: "{{.Values.redpandaAdminSecretName}}"
          refKey: USERNAME

        - key: KAFKA_PASSWORD
          type: secret
          refName: "{{.Values.redpandaAdminSecretName}}"
          refKey: PASSWORD

        - key: CLUSTER_CONFIGS_PATH
          value: /tmp/k8s

        - key: LOKI_AUTH_PASSWORD
          value: ""

        - key: PROMETHEUS_BASIC_AUTH_PASSWORD
          type: secret
          refName: "{{.PrometheusBasicAuthSecretName}}"
          refKey: "password"

        - key: KAFKA_AUDIT_EVENTS_TOPIC
          value: {{.KafkaTopicEvents}}

      {{/* envFrom: */}}
      {{/*   - type: secret */}}
      {{/*     refName: console-env */}}

      volumes:
        - mountPath: /console.d/templates
          type: config
          refName: {{.Values.apps.consoleApi.name}}-managed-svc-template

        - mountPath: /tmp/k8s
          type: secret
          refName: aggregated-kubeconfigs

{{/* --- */}}
{{/* apiVersion: v1 */}}
{{/* kind: Secret */}}
{{/* metadata: */}}
{{/*   name: {{.Values.apps.consoleApi.name}}-env */}}
{{/*   namespace: {{.Release.Namespace}} */}}
{{/* stringData: */}}
{{/*   IMAGE_REGISTRY_PREFIX: registry.kloudlite.io */}}
{{/*   LOG_PORT: "8192" */}}
{{/**/}}
{{/*   PORT: "3000" */}}
{{/*   GRPC_PORT: "3001" */}}
{{/**/}}
{{/*   NOTIFIER_URL:  */}}
{{/**/}}
{{/*   # LOKI_URL: loki-external.kl-01.$DOMAIN_1 */}}
{{/*   LOKI_URL: "loki-external.REPLACE_ME.clusters.{{.SubDomain}}" */}}
{{/*   LOKI_AUTH_PASSWORD: "$LOKI_EXTERNAL_PASSWORD" */}}
{{/**/}}
{{/*   METRICS_HTTP_PORT: "9191" */}}
{{/*   METRICS_HTTP_CORS: "https://console.{{.SubDomain}}" */}}
{{/*   PROMETHEUS_ENDPOINT: "https://prom-external.REPLACE_ME.clusters.{{.SubDomain}}" */}}
{{/*   PROMETHEUS_BASIC_AUTH_PASSWORD: "$PROM_EXTERNAL_PASSWORD" */}}
{{/**/}}
{{/*   LOG_SERVER_PORT: "8192" */}}
{{/*   COOKIE_DOMAIN: ".{{.CookieDomain}}" */}}
{{/**/}}
{{/*   # Kafka */}}
{{/*   KAFKA_WORKLOAD_STATUS_TOPIC: "kl-status-updates" */}}
{{/*   KAFKA_GROUP_ID: "control-plane" */}}
---
apiVersion: crds.kloudlite.io/v1
kind: Router
metadata:
  name: {{.Values.apps.consoleApi.name}}
  namespace: {{.Release.Namespace}}
  labels:
    kloudlite.io/account-ref: {{.Values.accountName}}
spec:
  domains:
    - "logs.{{.baseDomain}}"
  https:
    enabled: true
    forceRedirect: true
  routes:
    - app: {{.Values.apps.consoleApi.name}}
      path: /metrics
      port: 9191

    - app: {{.Values.apps.consoleApi.name}}
      path: /
      port: 3002

---
apiVersion: v1
kind: ConfigMap
metadata:
  name: {{.Values.apps.consoleApi.name}}-managed-svc-template
  namespace: {{.Release.Namespace}}
data:
  storage-plans.yaml: |+
    - name: BlockStorage
      desc: Block Storage

  compute-plans.yaml: |+
    - name: Basic
      desc: 1CPU, 1GB RAM
      sharingEnabled: true
      dedicatedEnabled: true
      memoryPerCPU: 1.0
      maxSharedCPUPerPod: 8.0
      maxDedicatedCPUPerPod: 16.0

    - name: General
      desc: 1CPU, 2GB RAM
      sharingEnabled: true
      dedicatedEnabled: true
      memoryPerCPU: 2.0
      maxSharedCPUPerPod: 20
      maxDedicatedCPUPerPod: 8

    - name: HighMemory
      desc: 1CPU, 4GB RAM
      sharingEnabled: true
      dedicatedEnabled: true
      memoryPerCPU: 4.0
      maxSharedCPUPerPod: 4
      maxDedicatedCPUPerPod: 8

  managed-svc-templates.yml: |+
    - category: db
      displayName: Databases
      list:
        - name: mongo_cluster
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
          apiVersion: mongodb.msvc.kloudlite.io/v1
          kind: StandaloneService
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
                  replicaCount: 1,
                  resources: {
                    cpu: {
                      min: `${inputs.cpu * 1000}m`,
                      max: `${inputs.cpu * 1000}m`,
                    },
                    memory: `${inputs.cpu * 1000}Mi`,
                    storage: {
                      size: `${inputs.size}Gi`,
                    }
                  },
                },
                error: null,
              };
            }


          estimator: |-
            function (inputs, plans) {
              var computePrice = plans.compute["Basic"].dedicatedPrice * inputs.cpu;
              var storagePrice = plans.storage["BlockStorage"].pricePerGB * inputs.size;
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
            - name: URI
              description: DB URI
          resources:
            # - name: root-creds
            #   displayName: "Default Credentials"
            #   default: true
            #   outputs:
            #     - name: USERNAME
            #       label: Username
            #     - name: PASSWORD
            #       label: Password
            #     - name: DB_NAME
            #       label: Database Name
            #     - name: HOSTS
            #       label: DB Service Hosts
            #     - name: URI
            #       label: Connection String
            #   getRefKey: |+
            #     function(installationId) {
            #       return `msvc-${installationId}`
            #     }

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
                - name: USERNAME
                  label: Username
                - name: PASSWORD
                  label: Password
                - name: DB_NAME
                  label: Database Name
                - name: HOSTS
                  label: DB Service Hosts
                - name: URI
                  label: Connection String

        - name: mysql_standalone
          logoUrl: https://img.icons8.com/material-two-tone/344/mysql-logo.png
          apiVersion: mysql.msvc.kloudlite.io/v1
          kind: StandaloneService
          displayName: MySQL Standalone
          description: MySQL Standalone
          active: true
          fields:
            - name: size
              label: Capacity in GB
              inputType: Number
              defaultValue: 5
              min: 1
              required: true
              unit: Gi

            - name: cpu
              label: Cpu
              inputType: Number
              defaultValue: 0.4
              min: 0.4
              max: 2
              required: true
              unit: vCpu

          inputMiddleware: |-
            const inputMiddleware = (inputs) =>{
              const plan = "Basic"
              return {
                annotation: {
                  "kloudlite.io/billing-plan": plan,
                  "kloudlite.io/billable-quantity": `${inputs.cpu}`,
                  "kloudlite.io/is-shared": "true",
                },
                inputs: {
                  replicaCount: 1,
                  resources: {
                    cpu: {
                      min: `${inputs.cpu * 1000/2}m`,
                      max: `${inputs.cpu * 1000}m`,
                    },
                    memory: `${inputs.cpu * 1000}Mi`,
                    storage: {
                      size: `${inputs.size}Gi`,
                    },
                  },
                },
                error: null,
              };
            }

          estimator: |-
            function (inputs, plans) {
              const defaultPlan = "Basic"
              var computePrice = plans.compute[defaultPlan].sharedPrice * inputs.cpu;
              var storagePrice = plans.storage["BlockStorage"].pricePerGB * inputs.size;
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
            - name: ROOT_PASSWORD
              label: Mysql Root Password

            - name: HOSTS
              label: Mysql Service Hosts

            - name: DSN
              label: Mysql DSN

            - name: URI
              label: Mysql Root URI

          resources:
            - name: db
              apiVersion: mysql.msvc.kloudlite.io/v1
              kind: Database
              displayName: Database
              description: MysqlDB
              # fields:
              #   - name: name
              #     label: Db Name
              #     inputType: String
              #     required: true
              outputs:
                - name: USERNAME
                  label: DB Username

                - name: PASSWORD
                  label: DB password

                - name: HOSTS
                  label: DB Hosts

                - name: DB_NAME
                  label: DB Name

                - name: DSN
                  label: Mysql DSN

                - name: URI
                  label: DB Uri
        - name: mysql_cluster
          logoUrl: https://img.icons8.com/material-two-tone/344/mysql-logo.png
          displayName: MySQL Cluster
          description: MySQL Cluster
          active: false

        - name: neo4j_database
          displayName: Neo4J Database
          apiVersion: neo4j.msvc.kloudlite.io/v1
          kind: StandaloneService
          description: "Neo4J Graph Data Platform"
          logoUrl: "https://dist.neo4j.com/wp-content/uploads/20210423072428/neo4j-logo-2020-1.svg"
          active: true
          fields:
            - name: cpu
              label: CPU
              inputType: Number
              required: true
              min: 1
              max: 2
              defaultValue: 1
              unit: vCpu

            - name: size
              label: Capacity in GB
              inputType: Number
              required: true
              defaultValue: 5
              min: 1
              unit: Gi

          inputMiddleware: |+
            const inputMiddleware = (inputs) =>{
              const plan = "General";

              return {
                annotation: {
                  "kloudlite.io/billing-plan": plan,
                  "kloudlite.io/billable-quantity": `${inputs.cpu}`,
                  "kloudlite.io/is-shared": "true",
                },
                inputs: {
                  replicaCount: 1,
                  resources: {
                    cpu: {
                      min: `${inputs.cpu * 1000/2}m`,
                      max: `${inputs.cpu * 1000}m`,
                    },
                    memory: `${inputs.cpu * 2000}Mi`,
                    storage: {
                      size: `${inputs.size}Gi`,
                    },
                  },
                },
                error: null,
              };
            }

          estimator: |+
            function (inputs, plans) {
              const plan = "General";
              var computePrice = plans.compute[plan].sharedPrice * inputs.cpu;
              var storagePrice = plans.storage["BlockStorage"].pricePerGB * inputs.size;
              var totalPrice = computePrice + storagePrice;
              return {
                totalPrice: totalPrice,
                error: null,
                properties: [],
              }
            }

          outputs: &neo4jOutput
            - name: ROOT_PASSWORD
              label: Neo4J Root Password
            - name: HOSTS
              label: Neo4J Service Hosts
            - name: ADMIN_HOSTS
              label: Neo4J Admin Service Hosts
            - name: PORT_BOLT
              label: Neo4J Bolt Service Port
            - name: PORT_HTTP
              label: Neo4J Http Service Port
            # - name: PORT_BACKUP
            #   label: Neo4J Backup Service Port

          resources:
            - name: root-creds
              displayName: "Default Credentials"
              default: true
              outputs: *neo4jOutput
              getRefKey: |+
                function(installationId) {
                  return `msvc-${installationId}`
                }

        - name: elasticsearch
          displayName: Elastic Search
          description: Search everything, anywhere
          logoUrl: 'https://assets.zabbix.com/img/brands/elastic.svg'
          apiVersion: elasticsearch.msvc.kloudlite.io/v1
          kind: Service
          active: true

          fields:
            - name: cpu
              label: CPU
              inputType: Number
              required: true
              min: 1
              max: 2
              defaultValue: 1
              unit: vCpu

            - name: size
              label: Capacity in GB
              inputType: Number
              required: true
              defaultValue: 2
              min: 1
              unit: Gi

          inputMiddleware: |+
            const inputMiddleware = (inputs) =>{
              const plan = "General";

              return {
                annotation: {
                  "kloudlite.io/billing-plan": plan,
                  "kloudlite.io/billable-quantity": `${inputs.cpu}`,
                  "kloudlite.io/is-shared": "true",
                },
                inputs: {
                  replicaCount: 1,
                  resources: {
                    cpu: {
                      min: `${inputs.cpu * 1000/2}m`,
                      max: `${inputs.cpu * 1000}m`,
                    },
                    memory: `${inputs.cpu * 2000}Mi`,
                    storage: {
                      size: `${inputs.size}Gi`,
                    },
                  },
                },
                error: null,
              };
            }

          estimator: |+
            function (inputs, plans) {
              const plan = "General";
              var computePrice = plans.compute[plan].sharedPrice * inputs.cpu;
              var storagePrice = plans.storage["BlockStorage"].pricePerGB * inputs.size;
              var totalPrice = computePrice + storagePrice;
              return {
                totalPrice: totalPrice,
                error: null,
                properties: [],
              }
            }

          outputs:
            - name: USERNAME
              label: Elastic Username
            - name: PASSWORD
              label: Elastic User Password
            - name: HOSTS
              label: Elastic Service Hosts
            - name: URI
              label: Elastic Service HTTP Uri

          resources:
            - name: root-creds
              displayName: "Default Credentials"
              default: true
              outputs:
                - name: USERNAME
                  label: Elastic Username
                - name: PASSWORD
                  label: Elastic User Password
                - name: HOSTS
                  label: Elastic Service Hosts
                - name: URI
                  label: Elastic Service HTTP Uri
              getRefKey: |+
                function(installationId) {
                  return `msvc-${installationId}`
                }


        - name: object_storage
          displayName: Object Storage
          apiVersion: s3.aws.kloudlite.io/v1
          kind: Bucket
          description: S3 compatible object storage
          active: false
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
              var storagePrice = plans.storage["ObjectStorage"].pricePerGB * inputs.size;
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
      list:
        - name: redis_cluster
          logoUrl: https://img.icons8.com/color/344/redis.png
          displayName: Redis Cluster
          description: Redis Cluster
          active: false
        - name: redis_standalone
          logoUrl: https://img.icons8.com/color/344/redis.png
          apiVersion: redis.msvc.kloudlite.io/v1
          kind: StandaloneService
          displayName: Redis Standalone
          description: Redis Standalone
          active: true
          fields:
            # - name: name
            #   label: Redis Instance Name
            #   inputType: String
            #   required: true

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
                  "kloudlite.io/billable-quantity": `${inputs.cpu}`,
                  "kloudlite.io/is-shared": "true",
                },
                inputs: {
                  replicaCount: 1,
                  resources: {
                    cpu: {
                      min: `${inputs.cpu * 1000/2}m`,
                      max: `${inputs.cpu * 1000}m`,
                    },
                    memory: `${inputs.cpu * 2 * 1000}Mi`,
                    storage: {
                      size: `${inputs.size}Gi`,
                    },
                  },
                },
                error: null,
              };
            }

          estimator: |+
            function (inputs, plans) {
              const defaultPlan = "General"
              var computePrice = plans.compute[defaultPlan].sharedPrice * inputs.cpu;
              var storagePrice = plans.storage["BlockStorage"].pricePerGB * inputs.size;
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
            - name: ACLAccount
              apiVersion: redis.msvc.kloudlite.io/v1
              kind: ACLAccount
              fields:
                # - name: name
                #   label: ACL Account Name

                - name: prefix
                  label: Redis Key Prefix
              outputs:
                - name: HOSTS
                  label: Redis Hosts

                - name: USERNAME
                  label: Redis User

                - name: PASSWORD
                  label: Redis User Password

                - name: PREFIX
                  label: Redis Prefix

                - name: URI
                  label: Redis Connection URI

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
