package internal_crds

type DeviceSpec struct {
	Account      string  `json:"account,omitempty"`
	ActiveRegion string  `json:"activeRegion,omitempty"`
	Offset       int     `json:"offset,omitempty"`
	DeviceId     string  `json:"deviceId,omitempty"`
	Ports        []int32 `json:"ports,omitempty"`
}

type DeviceMetadata struct {
	Name        string            `json:"name,omitempty"`
	Namespace   string            `json:"namespace,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty"`
	Labels      map[string]string `json:"labels,omitempty"`
}

const DeviceAPIVersion = "management.kloudlite.io/v1"
const DeviceKind = "Device"

type Device struct {
	APIVersion string `json:"apiVersion,omitempty"`
	Kind       string `json:"kind,omitempty"`

	Metadata DeviceMetadata `json:"metadata"`
	Spec     DeviceSpec     `json:"spec,omitempty"`
}
