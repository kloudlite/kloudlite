package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// InfluxDBSpec defines the desired state of InfluxDB
type InfluxDBSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of InfluxDB. Edit influxdb_types.go to remove/update
	Foo string `json:"foo,omitempty"`
}

// InfluxDBStatus defines the observed state of InfluxDB
type InfluxDBStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// InfluxDB is the Schema for the influxdbs API
type InfluxDB struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   InfluxDBSpec   `json:"spec,omitempty"`
	Status InfluxDBStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// InfluxDBList contains a list of InfluxDB
type InfluxDBList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []InfluxDB `json:"items"`
}

func init() {
	SchemeBuilder.Register(&InfluxDB{}, &InfluxDBList{})
}
