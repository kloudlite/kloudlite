package v1

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"operators.kloudlite.io/lib/errors"
)

type SecretRef struct {
	Name string `json:"name"`
}

type ConfigRef struct {
	Name string `json:"name"`
}

type ManagedSvcRef struct {
	Name      string     `json:"name"`
	ConfigRef *ConfigRef `json:"config_ref,omitempty"`
	SecretRef *SecretRef `json:"secret_ref,omitempty"`
}

// MongoDatabaseSpec defines the desired state of MongoDatabase
type MongoDatabaseSpec struct {
	ManagedSvc ManagedSvcRef `json:"managed_svc"`
	Outputs    []Output      `json:"outputs"`
}

// MongoDatabaseStatus defines the observed state of MongoDatabase
type MongoDatabaseStatus struct {
	SecretRef  SecretRef          `json:"secret_ref"`
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// MongoDatabase is the Schema for the mongodatabases API
type MongoDatabase struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MongoDatabaseSpec   `json:"spec,omitempty"`
	Status MongoDatabaseStatus `json:"status,omitempty"`
}

func (m *MongoDatabase) ConnectToDB(ctx context.Context, username, password, host, dbName string) (*mongo.Database, error) {
	client, err := mongo.NewClient(options.Client().ApplyURI(fmt.Sprintf("mongodb://%s:%s@%s", username, password, host)))
	if err != nil {
		return nil, errors.NewEf(err, "could not create mongodb client")
	}

	if err := client.Connect(ctx); err != nil {
		return nil, errors.NewEf(err, "could not connect to specified mongodb service")
	}
	db := client.Database(dbName)
	return db, nil
}

//+kubebuilder:object:root=true

// MongoDatabaseList contains a list of MongoDatabase
type MongoDatabaseList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MongoDatabase `json:"items"`
}

func init() {
	SchemeBuilder.Register(&MongoDatabase{}, &MongoDatabaseList{})
}
