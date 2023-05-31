package main

import (
	"sort"
	"testing"

	crdsv1 "github.com/kloudlite/operator/apis/crds/v1"
	"github.com/maxatome/go-testdeep/td"
	"k8s.io/client-go/rest"

	"kloudlite.io/pkg/k8s"
	"kloudlite.io/pkg/repos"
	types "kloudlite.io/pkg/types"
)

func TestGenerateGraphQLSchema(t *testing.T) {
	type args struct {
		name string
		data interface{}
		kCli k8s.ExtendedK8sClient
	}
	kCli, err := k8s.NewExtendedK8sClient(&rest.Config{Host: "localhost:8080"})
	if err != nil {
		panic(err)
	}
	tests := []struct {
		name    string
		args    args
		want    map[string][]string
		wantErr bool
	}{
		{
			name: "Test Case 1",
			args: args{
				name: "TestSchema1",
				data: struct {
					ID   int    `json:"id,omitempty"`
					Name string `json:"name,omitempty"`
				}{},
				kCli: nil,
			},
			want: map[string][]string{
				"TestSchema1": {
					"id: Int",
					"name: String",
				},
			},
			wantErr: false,
		},
		{
			name: "Test Case 2",
			args: args{
				name: "TestSchema2",
				data: struct {
					Age     int    `json:"age,omitempty"`
					Address string `json:"address,omitempty"`
				}{},
				kCli: nil,
			},
			want: map[string][]string{
				"TestSchema2": {
					"age: Int",
					"address: String",
				},
			},
			wantErr: false,
		},
		{
			name: "Test Case 3",
			args: args{
				name: "TestSchema3",
				data: struct {
					FirstName string `json:"first_name,omitempty"`
					LastName  string `json:"last_name,omitempty"`
				}{},
				kCli: nil,
			},
			want: map[string][]string{
				"TestSchema3": {
					"first_name: String",
					"last_name: String",
				},
			},
			wantErr: false,
		},
		{
			name: "Test Case 4",
			args: args{
				name: "TestSchema4",
				data: struct {
					Email     string  `json:"email"`
					IsEnabled bool    `json:"is_enabled,omitempty"`
					Age       *int    `json:"age,omitempty"`
					Score     float64 `json:"score,omitempty"`
				}{},
				kCli: nil,
			},
			want: map[string][]string{
				"TestSchema4": {
					"email: String!",
					"is_enabled: Boolean",
					"age: Int",
					"score: Float",
				},
			},
			wantErr: false,
		},
		{
			name: "Test Case 5",
			args: args{
				name: "TestSchema5",
				data: struct {
					FullName string `json:"full_name,omitempty"`
				}{},
				kCli: nil,
			},
			want: map[string][]string{
				"TestSchema5": {
					"full_name: String",
				},
			},
			wantErr: false,
		},
		{
			name: "Test Case 6",
			args: args{
				name: "TestSchema6",
				data: struct {
					Username  string `json:"-"`
					Password  string `json:"password"`
					CreatedAt string `json:"created_at,omitempty"`
				}{},
				kCli: nil,
			},
			want: map[string][]string{
				"TestSchema6": {
					"password: String!",
					"created_at: String",
				},
			},
			wantErr: false,
		},
		{
			name: "Test Case 7",
			args: args{
				name: "TestSchema7",
				data: struct {
					FirstName string  `json:"first_name"`
					LastName  string  `json:"last_name"`
					Age       *int    `json:"age,omitempty"`
					Salary    float64 `json:"salary,omitempty"`
					IsMarried bool    `json:"is_married,omitempty"`
				}{},
				kCli: nil,
			},
			want: map[string][]string{
				"TestSchema7": {
					"first_name: String!",
					"last_name: String!",
					"age: Int",
					"salary: Float",
					"is_married: Boolean",
				},
			},
			wantErr: false,
		},
		{
			name: "Test Case 8",
			args: args{
				name: "TestSchema8",
				data: struct {
					Emails  []string `json:"emails,omitempty"`
					Phones  []string `json:"phones,omitempty"`
					Address struct {
						Street  string `json:"street"`
						City    string `json:"city"`
						Country string `json:"country"`
					} `json:"address,omitempty"`
				}{},
				kCli: nil,
			},
			want: map[string][]string{
				"TestSchema8": {
					"emails: [String]",
					"phones: [String]",
					"address: TestSchema8Address",
				},
				"TestSchema8Address": {
					"street: String!",
					"city: String!",
					"country: String!",
				},
			},
			wantErr: false,
		},
		{
			name: "Test Case 9",
			args: args{
				name: "TestSchema9",
				data: struct {
					ID        int      `json:"id,omitempty"`
					Name      string   `json:"name,omitempty"`
					IsEnabled bool     `json:"is_enabled,omitempty"`
					Score     float64  `json:"score,omitempty"`
					Tags      []string `json:"tags,omitempty"`
				}{},
				kCli: nil,
			},
			want: map[string][]string{
				"TestSchema9": {
					"id: Int",
					"name: String",
					"is_enabled: Boolean",
					"score: Float",
					"tags: [String]",
				},
			},
			wantErr: false,
		},
		{
			name: "Test Case 10",
			args: args{
				name: "TestSchema10",
				data: struct {
					Title     string `json:"title,omitempty"`
					Content   string `json:"content,omitempty"`
					Published bool   `json:"published,omitempty"`
				}{},
				kCli: nil,
			},
			want: map[string][]string{
				"TestSchema10": {
					"title: String",
					"content: String",
					"published: Boolean",
				},
			},
			wantErr: false,
		},
		{
			name: "Test Case 11",
			args: args{
				name: "TestSchema11",
				data: struct {
					Title   string `json:"title,omitempty"`
					Content string `json:"content,omitempty"`
					Author  struct {
						Name  string `json:"name"`
						Email string `json:"email"`
					} `json:",inline"`
				}{},
				kCli: nil,
			},
			want: map[string][]string{
				"TestSchema11": {
					"title: String",
					"content: String",
					"name: String!",
					"email: String!",
				},
			},
			wantErr: false,
		},

		// generate 4 more tests, similar to what i have in test case 11, but with different struct fields, and a different structure
		// and different field tags values, and generate it all in one go, not line by line
		// and make sure it works

		{
			name: "Test Case 14",
			args: args{
				name: "TestSchema14",
				data: struct {
					ID          int    `json:"identifier,omitempty"`
					FirstName   string `json:",omitempty"`
					LastName    string `json:",omitempty"`
					DateOfBirth string `json:"dob,omitempty"`
				}{},
				kCli: nil,
			},
			want: map[string][]string{
				"TestSchema14": {
					"identifier: Int",
					"FirstName: String",
					"LastName: String",
					"dob: String",
				},
			},
			wantErr: false,
		},
		{
			name: "Test Case 15",
			args: args{
				name: "TestSchema15",
				data: struct {
					Email   string `json:"email,omitempty"`
					Phone   string `json:"phone,omitempty"`
					Address struct {
						Street  string `json:"street,omitempty"`
						City    string `json:"city,omitempty"`
						Country string `json:"country,omitempty"`
					} `json:",inline"`
				}{},
				kCli: nil,
			},
			want: map[string][]string{
				"TestSchema15": {
					"email: String",
					"phone: String",
					"street: String",
					"city: String",
					"country: String",
				},
			},
			wantErr: false,
		},
		{
			name: "Test Case 16",
			args: args{
				name: "TestSchema16",
				data: struct {
					FirstName string `json:"first_name,omitempty"`
					LastName  string `json:"last_name,omitempty"`
					Address   struct {
						Street  string `json:"street"`
						City    string `json:"city"`
						Country string `json:"country,omitempty"`
					} `json:"address,omitempty"`
				}{},
				kCli: nil,
			},
			want: map[string][]string{
				"TestSchema16": {
					"first_name: String",
					"last_name: String",
					"address: TestSchema16Address",
				},
				"TestSchema16Address": {
					"street: String!",
					"city: String!",
					"country: String",
				},
			},
			wantErr: false,
		},
		{
			name: "Test Case 17",
			args: args{
				name: "TestSchema17",
				data: struct {
					ID         int    `json:"id,omitempty"`
					Username   string `json:"username,omitempty"`
					IsVerified bool   `json:"is_verified,omitempty"`
					Profile    struct {
						Bio   string `json:"bio,omitempty"`
						Email string `json:"email"`
					} `json:",inline"`
				}{},
				kCli: nil,
			},
			want: map[string][]string{
				"TestSchema17": {
					"id: Int",
					"username: String",
					"is_verified: Boolean",
					"bio: String",
					"email: String!",
				},
			},
			wantErr: false,
		},

		{
			name: "Test Case 18",
			args: args{
				name: "Project",
				data: struct {
					repos.BaseEntity `json:",inline"`
					crdsv1.Project   `json:",inline" json-schema:"k8s://projects.crds.kloudlite.io"`
					AccountName      string           `json:"accountName"`
					ClusterName      string           `json:"clusterName"`
					SyncStatus       types.SyncStatus `json:"syncStatus"`
				}{},
				kCli: kCli,
			},
			want: map[string][]string{"Project": {"accountName: String!", "apiVersion: String", "clusterName: String!", "creation_time: Date", "id: String!", "kind: String", "metadata: Metadata! @goField(name: \"objectMeta\")", "spec: ProjectSpec", "status: ProjectStatus", "syncStatus: ProjectSyncStatus", "update_time: Date"}, "ProjectSpec": {"accountName: String!", "clusterName: String!", "displayName: String", "logo: String", "targetNamespace: String!"}, "ProjectStatus": {"checks: Map", "generatedVars: Map", "lastReconcileTime: String", "message: Map", "messages: [ProjectStatusMessages]", "opsConditions: [ProjectStatusOpsConditions]", "childConditions: [ProjectStatusChildConditions]", "conditions: [ProjectStatusConditions]", "displayVars: Map", "isReady: Boolean", "resources: [ProjectStatusResources]"}, "ProjectStatusChildConditions": {"lastTransitionTime: String!", "message: String!", "observedGeneration: Int", "reason: String!", "status: String!", "type: String!"}, "ProjectStatusConditions": {"status: String!", "type: String!", "lastTransitionTime: String!", "message: String!", "observedGeneration: Int", "reason: String!"}, "ProjectStatusMessages": {"reason: String", "state: String", "container: String", "exitCode: Int", "message: String", "pod: String"}, "ProjectStatusOpsConditions": {"lastTransitionTime: String!", "message: String!", "observedGeneration: Int", "reason: String!", "status: String!", "type: String!"}, "ProjectStatusResources": {"namespace: String!", "apiVersion: String", "kind: String", "name: String!"}, "ProjectSyncStatus": {"action: String!", "error: String", "generation: Int!", "lastSyncedAt: Date", "state: String", "syncScheduledAt: Date"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GenerateGraphQLSchema(tt.args.name, tt.args.data, tt.args.kCli)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateGraphQLSchema() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			for k, v := range tt.want {
				sort.Strings(v)
				tt.want[k] = v
			}

			if !td.Cmp(t, got, tt.want) {
				t.Errorf("GenerateGraphQLSchema() \ngot = %#v\nwant = %#v\n", got, tt.want)
			}
		})
	}
}
