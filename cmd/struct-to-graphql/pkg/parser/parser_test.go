package parser_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	rApi "github.com/kloudlite/operator/pkg/operator"
	apiExtensionsV1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"

	"sigs.k8s.io/yaml"

	exampleTypes "github.com/kloudlite/api/cmd/struct-to-graphql/internal/example/types"
	"github.com/kloudlite/api/cmd/struct-to-graphql/pkg/parser"
	types2 "github.com/kloudlite/api/cmd/struct-to-graphql/pkg/parser/testdata/types"
	"github.com/kloudlite/api/pkg/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ExampleJson struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              struct {
		ClusterName  string   `json:"clusterName"`
		NodePoolName string   `json:"nodePoolName"`
		NodeType     string   `json:"nodeType" graphql:"enum=worker;master;cluster"`
		Taints       []string `json:"taints"`
	}
}

type ProjectSpec struct {
	AccountName     string                    `json:"accountName"`
	ClusterName     string                    `json:"clusterName"`
	DisplayName     exampleTypes.SampleString `json:"displayName,omitempty"`
	TargetNamespace string                    `json:"targetNamespace"`
	Logo            string                    `json:"logo,omitempty"`
}

type Project struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              ProjectSpec `json:"spec"`
	Status            rApi.Status `json:"status,omitempty" graphql:"noinput"`
}

func exampleJsonSchema() ([]byte, error) {
	x := `description: Node is the Schema for the nodes API
properties:
  apiVersion:
    description: 'sample description'
    type: string
  kind:
    description: 'sample description'
    type: string
  metadata:
    type: object
  spec:
    properties:
      clusterName:
        type: string
      nodePoolName:
        type: string
      nodeType:
        enum:
          - worker
          - master
          - cluster
        type: string
      taints:
        items:
          type: string
        type: array
    required:
      - nodeType
      - clusterName
      - nodePoolName
    type: object
required:
  - spec
type: object
`

	return yaml.YAMLToJSON([]byte(x))
}

func exampleProjectCRD() ([]byte, error) {
	x := `apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  annotations:
    controller-gen.kubebuilder.io/version: v0.8.0
  creationTimestamp: "2023-06-27T06:11:21Z"
  generation: 5
  name: projects.crds.kloudlite.io
  resourceVersion: "141980035"
  uid: ba912e2c-1211-4e1c-bc26-84c11de0c46c
spec:
  conversion:
    strategy: None
  group: crds.kloudlite.io
  names:
    kind: Project
    listKind: ProjectList
    plural: projects
    singular: project
  scope: Cluster
  versions:
  - additionalPrinterColumns:
    - jsonPath: .spec.accountName
      name: AccountName
      type: string
    - jsonPath: .spec.clusterName
      name: ClusterName
      type: string
    - jsonPath: .spec.targetNamespace
      name: target-namespace
      type: string
    - jsonPath: .status.isReady
      name: Ready
      type: boolean
    - jsonPath: .metadata.creationTimestamp
      name: Age
      type: date
    name: v1
    schema:
      openAPIV3Schema:
        description: Project is the Schema for the projects API
        properties:
          apiVersion:
            description: 'APIVersion defines the versioned schema of this representation
              of an object. Servers should convert recognized schemas to the latest
              internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources'
            type: string
          kind:
            description: 'Kind is a string value representing the REST resource this
              object represents. Servers may infer this from the endpoint the client
              submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds'
            type: string
          metadata:
            type: object
          spec:
            description: ProjectSpec defines the desired state of Project
            properties:
              accountName:
                type: string
              clusterName:
                type: string
              displayName:
                type: string
              logo:
                type: string
              targetNamespace:
                type: string
            required:
            - accountName
            - clusterName
            - targetNamespace
            type: object
          status:
            properties:
              checks:
                additionalProperties:
                  properties:
                    generation:
                      format: int64
                      type: integer
                    message:
                      type: string
                    status:
                      type: boolean
                  required:
                  - status
                  type: object
                type: object
              isReady:
                type: boolean
              lastReconcileTime:
                format: date-time
                type: string
              message:
                type: object
                x-kubernetes-preserve-unknown-fields: true
              resources:
                items:
                  properties:
                    apiVersion:
                      description: 'APIVersion defines the versioned schema of this
                        representation of an object. Servers should convert recognized
                        schemas to the latest internal value, and may reject unrecognized
                        values. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources'
                      type: string
                    kind:
                      description: 'Kind is a string value representing the REST resource
                        this object represents. Servers may infer this from the endpoint
                        the client submits requests to. Cannot be updated. In CamelCase.
                        More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds'
                      type: string
                    name:
                      type: string
                    namespace:
                      type: string
                  required:
                  - name
                  - namespace
                  type: object
                type: array
            type: object
        required:
        - spec
        type: object
    served: true
    storage: true
    subresources:
      status: {}
`

	return yaml.YAMLToJSON([]byte(x))
}

type schemaClient struct{}

func (s schemaClient) GetK8sJsonSchema(name string) (*apiExtensionsV1.JSONSchemaProps, error) {
	if name == "projects.crds.kloudlite.io" {
		b, err := exampleProjectCRD()
		if err != nil {
			return nil, err
		}

		crd := apiExtensionsV1.CustomResourceDefinition{}
		if err := json.Unmarshal(b, &crd); err != nil {
			return nil, err
		}

		b2, err := json.Marshal(crd.Spec.Versions[0].Schema.OpenAPIV3Schema)
		if err != nil {
			return nil, err
		}

		var m apiExtensionsV1.JSONSchemaProps
		if err := json.Unmarshal(b2, &m); err != nil {
			return nil, err
		}
		return &m, nil
	}

	panic("unknown k8s crd resource name")
}

func (s schemaClient) GetHttpJsonSchema(url string) (*apiExtensionsV1.JSONSchemaProps, error) {
	if strings.HasSuffix(url, "example-json-schema") {
		schema, err := exampleJsonSchema()
		if err != nil {
			return nil, err
		}

		var m apiExtensionsV1.JSONSchemaProps
		if err := json.Unmarshal(schema, &m); err != nil {
			return nil, err
		}
		return &m, nil
	}
	panic("unknown http route")
}

func Test_GeneratedGraphqlSchema(t *testing.T) {
	//schemaCli, err := func() (kubernetes.Clientset, error) {
	//	// kc := kubernetesClient{}
	//	// return k8s.NewExtendedK8sClient()
	//	return kubernetesClient{}, nil
	//	// return k8s.NewExtendedK8sClient(&rest.Config{Host: "localhost:8080"})
	//}()

	schemaCli := &schemaClient{}

	type fields struct {
		structs   map[string]*parser.Struct
		schemaCli parser.SchemaClient
	}

	type args struct {
		name           string
		data           any
		withPagination []string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    map[string]*parser.Struct
		wantErr bool
	}{
		{
			name: "test 1. without any json tag",
			fields: fields{
				structs:   map[string]*parser.Struct{},
				schemaCli: schemaCli,
			},
			args: args{
				name: "User",
				data: struct {
					ID       int
					Username string
					Gender   string
				}{},
			},
			want: map[string]*parser.Struct{
				"User": {
					Types: map[string][]string{
						"User": {
							"ID: Int!",
							"Username: String!",
							"Gender: String!",
						},
					},
					Inputs: map[string][]string{
						"UserIn": {
							"ID: Int!",
							"Username: String!",
							"Gender: String!",
						},
					},
					Enums: map[string][]string{},
				},
			},
		},
		{
			name: "test 2. with json tags, for naming",
			fields: fields{
				structs:   map[string]*parser.Struct{},
				schemaCli: schemaCli,
			},
			args: args{
				name: "User",
				data: struct {
					ID       int    `json:"id,omitempty"`
					Username string `json:"username"`
					Gender   string `json:"gender"`
				}{},
			},
			want: map[string]*parser.Struct{
				"User": {
					Types: map[string][]string{
						"User": {
							"id: Int",
							"username: String!",
							"gender: String!",
						},
					},
					Inputs: map[string][]string{
						"UserIn": {
							"id: Int",
							"username: String!",
							"gender: String!",
						},
					},
					Enums: map[string][]string{},
				},
			},
		},
		{
			name: "test 3. with json tags for naming, and graphql enum tags",
			fields: fields{
				structs:   map[string]*parser.Struct{},
				schemaCli: schemaCli,
			},
			args: args{
				name: "User",
				data: struct {
					ID       int    `json:"id,omitempty"`
					Username string `json:"username"`
					Gender   string `json:"gender" graphql:"enum=MALE;FEMALE"`
				}{},
			},
			want: map[string]*parser.Struct{
				"User": {
					Types: map[string][]string{
						"User": {
							"id: Int",
							"username: String!",
							"gender: UserGender!",
						},
					},
					Inputs: map[string][]string{
						"UserIn": {
							"id: Int",
							"username: String!",
							"gender: UserGender!",
						},
					},
					Enums: map[string][]string{
						"UserGender": {
							"FEMALE",
							"MALE",
						},
					},
				},
			},
		},
		{
			name: "test 4. with struct containing slice field",
			fields: fields{
				structs:   map[string]*parser.Struct{},
				schemaCli: schemaCli,
			},
			args: args{
				name: "Post",
				data: struct {
					ID      int
					Title   string
					Content string
					Tags    []string
				}{},
			},
			want: map[string]*parser.Struct{
				"Post": {
					Types: map[string][]string{
						"Post": {
							"ID: Int!",
							"Title: String!",
							"Content: String!",
							"Tags: [String!]!",
						},
					},
					Inputs: map[string][]string{
						"PostIn": {
							"ID: Int!",
							"Title: String!",
							"Content: String!",
							"Tags: [String!]!",
						},
					},
					Enums: map[string][]string{},
				},
			},
		},
		{
			name: "test 5. with struct containing pointer field",
			fields: fields{
				structs:   map[string]*parser.Struct{},
				schemaCli: schemaCli,
			},
			args: args{
				name: "Address",
				data: struct {
					Street  string
					City    string
					Country *string
				}{},
			},
			want: map[string]*parser.Struct{
				"Address": {
					Types: map[string][]string{
						"Address": {
							"Street: String!",
							"City: String!",
							"Country: String",
						},
					},
					Inputs: map[string][]string{
						"AddressIn": {
							"Street: String!",
							"City: String!",
							"Country: String",
						},
					},
					Enums: map[string][]string{},
				},
			},
		},
		{
			name: "test 6. with struct containing nested anonymous struct field",
			fields: fields{
				structs:   map[string]*parser.Struct{},
				schemaCli: schemaCli,
			},
			args: args{
				name: "Employee",
				data: struct {
					ID      int
					Name    string
					Address struct {
						Street string
						City   string
					}
				}{},
			},
			want: map[string]*parser.Struct{
				"Employee": {
					Types: map[string][]string{
						"Employee": {
							"ID: Int!",
							"Name: String!",
							"Address: EmployeeAddress!",
						},
						"EmployeeAddress": {
							"Street: String!",
							"City: String!",
						},
					},
					Inputs: map[string][]string{
						"EmployeeIn": {
							"ID: Int!",
							"Name: String!",
							"Address: EmployeeAddressIn!",
						},
						"EmployeeAddressIn": {
							"Street: String!",
							"City: String!",
						},
					},
					Enums: map[string][]string{},
				},
			},
		},
		{
			name: "test 7. with struct containing nested struct field with json tags",
			fields: fields{
				structs:   map[string]*parser.Struct{},
				schemaCli: schemaCli,
			},
			args: args{
				name: "Employee",
				data: struct {
					ID      int
					Name    string
					Address struct {
						Street string `json:"street"`
						City   string `json:"city"`
					} `json:"address"`
				}{},
			},
			want: map[string]*parser.Struct{
				"Employee": {
					Types: map[string][]string{
						"Employee": {
							"ID: Int!",
							"Name: String!",
							"address: EmployeeAddress!",
						},
						"EmployeeAddress": {
							"street: String!",
							"city: String!",
						},
					},
					Inputs: map[string][]string{
						"EmployeeIn": {
							"ID: Int!",
							"Name: String!",
							"address: EmployeeAddressIn!",
						},
						"EmployeeAddressIn": {
							"street: String!",
							"city: String!",
						},
					},
					Enums: map[string][]string{},
				},
			},
		},
		{
			name: "test 8. with struct containing struct pointer field",
			fields: fields{
				structs:   map[string]*parser.Struct{},
				schemaCli: schemaCli,
			},
			args: args{
				name: "Company",
				data: struct {
					ID      int
					Name    string
					Address *struct {
						Street string
						City   string
					}
				}{},
			},
			want: map[string]*parser.Struct{
				"Company": {
					Types: map[string][]string{
						"Company": {
							"ID: Int!",
							"Name: String!",
							"Address: CompanyAddress",
						},
						"CompanyAddress": {
							"Street: String!",
							"City: String!",
						},
					},
					Inputs: map[string][]string{
						"CompanyIn": {
							"ID: Int!",
							"Name: String!",
							"Address: CompanyAddressIn",
						},
						"CompanyAddressIn": {
							"Street: String!",
							"City: String!",
						},
					},
					Enums: map[string][]string{},
				},
			},
		},
		{
			name: "test 9. with struct containing struct slice field",
			fields: fields{
				structs:   map[string]*parser.Struct{},
				schemaCli: schemaCli,
			},
			args: args{
				name: "Organization",
				data: struct {
					ID        int
					Name      string
					Employees []struct {
						ID   int
						Name string
					}
				}{},
			},
			want: map[string]*parser.Struct{
				"Organization": {
					Types: map[string][]string{
						"Organization": {
							"ID: Int!",
							"Name: String!",
							"Employees: [OrganizationEmployees!]!",
						},
						"OrganizationEmployees": {
							"ID: Int!",
							"Name: String!",
						},
					},
					Inputs: map[string][]string{
						"OrganizationIn": {
							"ID: Int!",
							"Name: String!",
							"Employees: [OrganizationEmployeesIn!]!",
						},
						"OrganizationEmployeesIn": {
							"ID: Int!",
							"Name: String!",
						},
					},
					Enums: map[string][]string{},
				},
			},
		},
		{
			name: "test 10. with struct containing struct slice field with json tags",
			fields: fields{
				structs:   map[string]*parser.Struct{},
				schemaCli: schemaCli,
			},
			args: args{
				name: "Organization",
				data: struct {
					ID        int
					Name      string
					Employees []struct {
						ID   int    `json:"employee_id"`
						Name string `json:"employee_name"`
					} `json:"employees"`
				}{},
			},
			want: map[string]*parser.Struct{
				"Organization": {
					Types: map[string][]string{
						"Organization": {
							"ID: Int!",
							"Name: String!",
							"employees: [OrganizationEmployees!]!",
						},
						"OrganizationEmployees": {
							"employee_id: Int!",
							"employee_name: String!",
						},
					},
					Inputs: map[string][]string{
						"OrganizationIn": {
							"ID: Int!",
							"Name: String!",
							"employees: [OrganizationEmployeesIn!]!",
						},
						"OrganizationEmployeesIn": {
							"employee_id: Int!",
							"employee_name: String!",
						},
					},
					Enums: map[string][]string{},
				},
			},
		},
		{
			name: "test 11. with struct containing enum field",
			fields: fields{
				structs:   map[string]*parser.Struct{},
				schemaCli: schemaCli,
			},
			args: args{
				name: "Product",
				data: struct {
					ID       int
					Name     string
					Category string `graphql:"enum=ELECTRONICS;FASHION;SPORTS"`
				}{},
			},
			want: map[string]*parser.Struct{
				"Product": {
					Types: map[string][]string{
						"Product": {
							"ID: Int!",
							"Name: String!",
							"Category: ProductCategory!",
						},
					},
					Inputs: map[string][]string{
						"ProductIn": {
							"ID: Int!",
							"Name: String!",
							"Category: ProductCategory!",
						},
					},
					Enums: map[string][]string{
						"ProductCategory": {
							"ELECTRONICS",
							"FASHION",
							"SPORTS",
						},
					},
				},
			},
		},
		{
			name: "test 12. with struct containing struct slice to pointer of a inline struct",
			fields: fields{
				structs:   map[string]*parser.Struct{},
				schemaCli: schemaCli,
			},
			args: args{
				name: "Organization",
				data: struct {
					ID        int
					Name      string
					Employees []*struct {
						ID   int    `json:"employee_id"`
						Name string `json:"employee_name"`
					} `json:"employees"`
				}{},
			},
			want: map[string]*parser.Struct{
				"Organization": {
					Types: map[string][]string{
						"Organization": {
							"ID: Int!",
							"Name: String!",
							"employees: [OrganizationEmployees]!",
						},
						"OrganizationEmployees": {
							"employee_id: Int!",
							"employee_name: String!",
						},
					},
					Inputs: map[string][]string{
						"OrganizationIn": {
							"ID: Int!",
							"Name: String!",
							"employees: [OrganizationEmployeesIn]!",
						},
						"OrganizationEmployeesIn": {
							"employee_id: Int!",
							"employee_name: String!",
						},
					},
					Enums: map[string][]string{},
				},
			},
		},
		{
			name: "test 13. with struct containing map field",
			fields: fields{
				structs:   map[string]*parser.Struct{},
				schemaCli: schemaCli,
			},
			args: args{
				name: "User",
				data: struct {
					ID    int
					Name  string
					Email string
					Tags  map[string]string
					KVs   map[string]any `json:"kvs"`
				}{},
			},
			want: map[string]*parser.Struct{
				"User": {
					Types: map[string][]string{
						"User": {
							"ID: Int!",
							"Name: String!",
							"Email: String!",
							"Tags: Map!",
							"kvs: Map!",
						},
					},
					Inputs: map[string][]string{
						"UserIn": {
							"ID: Int!",
							"Name: String!",
							"Email: String!",
							"Tags: Map!",
							"kvs: Map!",
						},
					},
					Enums: map[string][]string{},
				},
			},
		},
		{
			name: "test 14. with struct containing nested kloudlite CRD",
			fields: fields{
				structs:   map[string]*parser.Struct{},
				schemaCli: schemaCli,
			},
			args: args{
				name: "Project",
				data: struct {
					AccountName string
					Project     Project `json:",inline" graphql:"uri=k8s://projects.crds.kloudlite.io"`
				}{},
			},
			want: map[string]*parser.Struct{
				"Project": {
					Types: map[string][]string{
						"Project": {
							"AccountName: String!",
							"apiVersion: String!",
							"kind: String!",
							"metadata: Metadata! @goField(name: \"objectMeta\")",
							"spec: Kloudlite__io___cmd___struct____to____graphql___pkg___parser_test__ProjectSpec!",
							"status: Github__com___kloudlite___operator___pkg___operator__Status",
						},
					},
					Inputs: map[string][]string{
						"ProjectIn": {
							"AccountName: String!",
							"metadata: MetadataIn!",
							"spec: Kloudlite__io___cmd___struct____to____graphql___pkg___parser_test__ProjectSpecIn!",
						},
					},
					Enums: map[string][]string{},
				},
				"common-types": {
					Types: map[string][]string{
						"Kloudlite__io___cmd___struct____to____graphql___pkg___parser_test__ProjectSpec": {
							"accountName: String!",
							"clusterName: String!",
							"displayName: Kloudlite__io___cmd___struct____to____graphql___internal___example___types__SampleString",
							"logo: String",
							"targetNamespace: String!",
						},
						"Github__com___kloudlite___operator___pkg___operator__Check": {
							"status: Boolean!",
							"message: String",
							"generation: Int",
						},
						"Github__com___kloudlite___operator___pkg___operator__ResourceRef": {
							"apiVersion: String!",
							"kind: String!",
							"namespace: String!",
							"name: String!",
						},
						"Github__com___kloudlite___operator___pkg___operator__Status": {
							"isReady: Boolean!",
							"resources: [Github__com___kloudlite___operator___pkg___operator__ResourceRef!]",
							"message: Github__com___kloudlite___operator___pkg___raw____json__RawJson",
							"checks: Map",
							"lastReadyGeneration: Int",
							"lastReconcileTime: Date",
						},
						"Github__com___kloudlite___operator___pkg___raw____json__RawJson": {
							"RawMessage: Any",
						},
						"Metadata": {
							"name: String!",
							"namespace: String",
							"labels: Map",
							"annotations: Map",
							"generation: Int!",
							"creationTimestamp: Date!",
							"deletionTimestamp: Date",
						},
					},
					Inputs: map[string][]string{
						"Kloudlite__io___cmd___struct____to____graphql___pkg___parser_test__ProjectSpecIn": {
							"accountName: String!",
							"clusterName: String!",
							"displayName: Kloudlite__io___cmd___struct____to____graphql___internal___example___types__SampleString",
							"logo: String",
							"targetNamespace: String!",
						},
						"MetadataIn": {
							"name: String!",
							"namespace: String",
							"labels: Map",
							"annotations: Map",
						},
					},
					Enums: map[string][]string{
						"Kloudlite__io___cmd___struct____to____graphql___internal___example___types__SampleString": {
							"item_1",
							"item_2",
						},
					},
				},
			},
		},
		{
			name: "test 15. with pagination enabled",
			fields: fields{
				structs:   map[string]*parser.Struct{},
				schemaCli: schemaCli,
			},
			args: args{
				name: "User",
				data: struct {
					ID       int
					Username string
					Gender   string
				}{},
				withPagination: []string{"User"},
			},
			want: map[string]*parser.Struct{
				"User": {
					Types: map[string][]string{
						"User": {
							"ID: Int!",
							"Username: String!",
							"Gender: String!",
						},
						"UserPaginatedRecords": {
							"totalCount: Int!",
							"edges: [UserEdge!]!",
							"pageInfo: PageInfo!",
						},
						"UserEdge": {
							"node: User!",
							"cursor: String!",
						},
					},
					Inputs: map[string][]string{
						"UserIn": {
							"ID: Int!",
							"Username: String!",
							"Gender: String!",
						},
					},
					Enums: map[string][]string{},
				},
				"common-types": {
					Types: map[string][]string{
						"PageInfo": {
							"hasNextPage: Boolean",
							"hasPreviousPage: Boolean",
							"startCursor: String",
							"endCursor: String",
						},
					},
				},
			},
		},
		{
			name: "test 16. (with graphql (noinput))",
			fields: fields{
				structs:   map[string]*parser.Struct{},
				schemaCli: schemaCli,
			},
			args: args{
				name: "User",
				data: struct {
					SyncStatus types.SyncStatus `json:"syncStatus" graphql:"noinput"`
				}{},
			},
			want: map[string]*parser.Struct{
				"User": {
					Types: map[string][]string{
						"User": {
							"syncStatus: Kloudlite__io___pkg___types__SyncStatus!",
						},
					},
					Inputs: map[string][]string{},
					Enums:  map[string][]string{},
				},
				"common-types": {
					Types: map[string][]string{
						"Kloudlite__io___pkg___types__SyncStatus": {
							"action: Kloudlite__io___pkg___types__SyncStatusAction!",
							"error: String",
							"recordVersion: Int!",
							"lastSyncedAt: Date",
							"state: Kloudlite__io___pkg___types__SyncStatusState!",
							"syncScheduledAt: Date",
						},
					},
					Enums: map[string][]string{
						"Kloudlite__io___pkg___types__SyncStatusAction": {
							"APPLY",
							"DELETE",
						},
						"Kloudlite__io___pkg___types__SyncStatusState": {
							"IDLE",
							"APPLIED_AT_AGENT",
							"ERRORED_AT_AGENT",
							"IN_QUEUE",
							"RECEIVED_UPDATE_FROM_AGENT",
						},
					},
				},
			},
		},
		{
			name: "test 17. with json schema http uri, and Spec field with no json tag",
			fields: fields{
				structs:   map[string]*parser.Struct{},
				schemaCli: schemaCli,
			},
			args: args{
				name: "Example",
				data: struct {
					// Example ExampleJson `json:"example" graphql:"uri=http://localhost:30017/example-json-schema"`
					Example ExampleJson `json:"example" graphql:"uri=http://example.com/example-json-schema"`
				}{},
			},
			want: map[string]*parser.Struct{
				"Example": {
					Types: map[string][]string{
						"Example": {
							"example: Kloudlite__io___cmd___struct____to____graphql___pkg___parser_test__ExampleJson!",
						},
					},
					Inputs: map[string][]string{
						"ExampleIn": {
							"example: Kloudlite__io___cmd___struct____to____graphql___pkg___parser_test__ExampleJsonIn!",
						},
					},
					Enums: map[string][]string{},
				},
				"common-types": {
					Types: map[string][]string{
						"Kloudlite__io___cmd___struct____to____graphql___pkg___parser_test__ExampleJson": {
							"apiVersion: String!",
							"kind: String!",
							"metadata: Metadata! @goField(name: \"objectMeta\")",
							"Spec: Kloudlite__io___cmd___struct____to____graphql___pkg___parser_test__ExampleJsonSpec!",
						},
						"Kloudlite__io___cmd___struct____to____graphql___pkg___parser_test__ExampleJsonSpec": {
							"clusterName: String!",
							"nodePoolName: String!",
							"nodeType: Kloudlite__io___cmd___struct____to____graphql___pkg___parser_test__ExampleJsonSpecNodeType!",
							"taints: [String!]!",
						},
						"Metadata": {
							"annotations: Map",
							"labels: Map",
							"name: String!",
							"namespace: String",
							"creationTimestamp: Date!",
							"deletionTimestamp: Date",
							"generation: Int!",
						},
					},
					Inputs: map[string][]string{
						"Kloudlite__io___cmd___struct____to____graphql___pkg___parser_test__ExampleJsonIn": {
							"metadata: MetadataIn!",
							"Spec: Kloudlite__io___cmd___struct____to____graphql___pkg___parser_test__ExampleJsonSpecIn!",
						},
						"Kloudlite__io___cmd___struct____to____graphql___pkg___parser_test__ExampleJsonSpecIn": {
							"clusterName: String!",
							"nodePoolName: String!",
							"nodeType: Kloudlite__io___cmd___struct____to____graphql___pkg___parser_test__ExampleJsonSpecNodeType!",
							"taints: [String!]!",
						},
						"MetadataIn": {
							"annotations: Map",
							"labels: Map",
							"name: String!",
							"namespace: String",
						},
					},
					Enums: map[string][]string{
						"Kloudlite__io___cmd___struct____to____graphql___pkg___parser_test__ExampleJsonSpecNodeType": {
							"worker",
							"master",
							"cluster",
						},
					},
				},
			},
		},
		{
			name: "test 18. with some empty enums",
			fields: fields{
				structs:   map[string]*parser.Struct{},
				schemaCli: schemaCli,
			},
			args: args{
				name: "User",
				data: struct {
					Example string `json:"example" graphql:"enum=e1;e2;;e3;;e4"`
				}{},
			},
			want: map[string]*parser.Struct{
				"Example": {
					Types: map[string][]string{
						"User": {
							"example: UserExample!",
						},
					},
					Inputs: map[string][]string{
						"UserIn": {
							"example: UserExample!",
						},
					},
					Enums: map[string][]string{
						"UserExample": {
							"e1",
							"e2",
							"e3",
							"e4",
						},
					},
				},
			},
		},
		{
			name: "test 19. with default values for fields, with single-quoted string as default",
			fields: fields{
				structs:   map[string]*parser.Struct{},
				schemaCli: schemaCli,
			},
			args: args{
				name: "Account",
				data: struct {
					IsActive bool   `json:"isActive" graphql:"default=true"`
					Country  string `json:"country" graphql:"default='INDIA'"`
					Region   string `json:"region" graphql:"default=\"us-east-1\""`
				}{},
			},
			wantErr: true,
			want: map[string]*parser.Struct{
				"Account": {
					Types: map[string][]string{
						"Account": {
							"isActive: Boolean!",
							"country: String!",
							"region: String!",
						},
					},
					Inputs: map[string][]string{
						"AccountIn": {
							"isActive: Boolean! = true",
							"country: String! = 'INDIA'",
							"region: String! = \"us-east-1\"",
						},
					},
					Enums: map[string][]string{},
				},
			},
		},

		{
			name: "test 20. with default values for fields, with correct defaults",
			fields: fields{
				structs:   map[string]*parser.Struct{},
				schemaCli: schemaCli,
			},
			args: args{
				name: "Account",
				data: struct {
					IsActive bool   `json:"isActive" graphql:"default=true"`
					Country  string `json:"country" graphql:"default=\"INDIA\""`
					Region   string `json:"region" graphql:"default=\"us-east-1\""`
				}{},
			},
			want: map[string]*parser.Struct{
				"Account": {
					Types: map[string][]string{
						"Account": {
							"isActive: Boolean!",
							"country: String!",
							"region: String!",
						},
					},
					Inputs: map[string][]string{
						"AccountIn": {
							"isActive: Boolean! = true",
							"country: String! = \"INDIA\"",
							"region: String! = \"us-east-1\"",
						},
					},
					Enums: map[string][]string{},
				},
			},
		},

		{
			name:   "test 21. embedded struct with json inline tags",
			fields: fields{structs: map[string]*parser.Struct{}, schemaCli: schemaCli},
			args: args{
				name: "Sample",
				data: struct {
					Name        string `json:"name"`
					ProjectSpec `json:",inline"`
				}{},
			},
			want: map[string]*parser.Struct{
				"Sample": {
					Types: map[string][]string{
						"Sample": {
							"name: String!",
							"accountName: String!",
							"clusterName: String!",
							"displayName: Kloudlite__io___cmd___struct____to____graphql___internal___example___types__SampleString",
							"logo: String",
							"targetNamespace: String!",
						},
					},
					Inputs: map[string][]string{
						"SampleIn": {
							"name: String!",
							"accountName: String!",
							"clusterName: String!",
							"displayName: Kloudlite__io___cmd___struct____to____graphql___internal___example___types__SampleString",
							"logo: String",
							"targetNamespace: String!",
						},
					},
				},
				"common-types": {
					Enums: map[string][]string{
						"Kloudlite__io___cmd___struct____to____graphql___internal___example___types__SampleString": {
							"item_1",
							"item_2",
						},
					},
				},
			},
			wantErr: false,
		},

		{
			name:   "test 22. embedded struct (imported from other package) with json inline tags",
			fields: fields{structs: map[string]*parser.Struct{}, schemaCli: schemaCli},
			args: args{
				name: "Sample",
				data: struct {
					Name          string `json:"name"`
					types2.Sample `json:",inline"`
				}{},
			},
			want: map[string]*parser.Struct{
				"Sample": {
					Types: map[string][]string{
						"Sample": {
							"name: String!",
							"displayName: String!",
							"age: Int!",
							"createdBy: Kloudlite__io___cmd___struct____to____graphql___pkg___parser___testdata___types__ActionMeta!",
							"updatedBy: Kloudlite__io___cmd___struct____to____graphql___pkg___parser___testdata___types__ActionMeta!",
						},
					},
					Inputs: map[string][]string{
						"SampleIn": {
							"name: String!",
							"displayName: String!",
							"age: Int!",
							"createdBy: Kloudlite__io___cmd___struct____to____graphql___pkg___parser___testdata___types__ActionMetaIn!",
							"updatedBy: Kloudlite__io___cmd___struct____to____graphql___pkg___parser___testdata___types__ActionMetaIn!",
						},
					},
				},
				"common-types": {
					Types: map[string][]string{
						"Kloudlite__io___cmd___struct____to____graphql___pkg___parser___testdata___types__ActionMeta": {
							"firstName: String!",
							"lastName: String!",
						},
					},
					Inputs: map[string][]string{
						"Kloudlite__io___cmd___struct____to____graphql___pkg___parser___testdata___types__ActionMetaIn": {
							"firstName: String!",
							"lastName: String!",
						},
					},
				},
			},
			wantErr: false,
		},

		{
			name:   "test 23. string type as an enum of constants defined in pkg",
			fields: fields{structs: map[string]*parser.Struct{}, schemaCli: schemaCli},
			args: args{
				name: "Sample",
				data: struct {
					Name       string                    `json:"name"`
					SampleName exampleTypes.SampleString `json:"sampleName,omitempty"`
				}{},
			},
			want: map[string]*parser.Struct{
				"Sample": {
					Types: map[string][]string{
						"Sample": {
							"name: String!",
							"sampleName: Kloudlite__io___cmd___struct____to____graphql___internal___example___types__SampleString",
						},
					},
					Inputs: map[string][]string{
						"SampleIn": {
							"name: String!",
							"sampleName: Kloudlite__io___cmd___struct____to____graphql___internal___example___types__SampleString",
						},
					},
				},
				"common-types": {
					Enums: map[string][]string{
						"Kloudlite__io___cmd___struct____to____graphql___internal___example___types__SampleString": {
							"item_1",
							"item_2",
						},
					},
				},
			},
			wantErr: false,
		},
	}

	for _idx, _tt := range tests {
		idx := _idx
		tt := _tt
		t.Run(tt.name, func(t *testing.T) {
			// t.Parallel()

			p := parser.NewParser(tt.fields.schemaCli)
			err := p.LoadStruct(tt.args.name, tt.args.data)
			if err != nil {
				if tt.wantErr {
					return
				}
				t.Error(err)
			}

			p.WithPagination(tt.args.withPagination)

			testDir := filepath.Join(os.TempDir(), fmt.Sprintf("struct-to-graphql-testcase-%d", idx+1))
			os.Mkdir(testDir, 0o755)
			t.Logf("testcase output directory: %s", testDir)

			gbuft := new(bytes.Buffer)
			gbufc := new(bytes.Buffer)
			p.PrintTypes(gbuft)
			p.PrintCommonTypes(gbufc)
			gotTypes := gbuft.String()
			gotCommonTypes := gbufc.String()

			wantParser := parser.NewUnsafeParser(tt.want, nil)
			wbuft := new(bytes.Buffer)
			wbufc := new(bytes.Buffer)
			wantParser.PrintTypes(wbuft)
			wantParser.PrintCommonTypes(wbufc)
			wantTypes := wbuft.String()
			wantCommonTypes := wbufc.String()

			g, err := os.Create(filepath.Join(testDir, "./got.types.graphql"))
			if err != nil {
				t.Error(err)
			}
			g.WriteString(gotTypes)

			w, err := os.Create(filepath.Join(testDir, "./want.types.graphql"))
			if err != nil {
				t.Error(err)
			}
			w.WriteString(wantTypes)

			if gotTypes != wantTypes {
				t.Logf("diff %s %s", g.Name(), w.Name())
				b, err := exec.Command("diff", g.Name(), w.Name()).CombinedOutput()
				if err != nil {
					t.Error(err)
				}

				t.Errorf("diff output: \n%s\n", b)
			}

			g, err = os.Create(filepath.Join(testDir, "./got.common-types.graphql"))
			if err != nil {
				t.Error(err)
			}
			g.WriteString(gotCommonTypes)

			w, err = os.Create(filepath.Join(testDir, "./want.common-types.graphql"))
			if err != nil {
				t.Error(err)
			}
			w.WriteString(wantCommonTypes)

			if gotCommonTypes != wantCommonTypes {
				t.Logf("diff %s %s", g.Name(), w.Name())
				b, err := exec.Command("diff", g.Name(), w.Name()).CombinedOutput()
				if err != nil {
					t.Error(err)
				}

				t.Errorf("diff output: \n%s\n", b)
			}
		})
	}
}
