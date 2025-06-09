package v1

import (
	ct "github.com/kloudlite/operator/apis/common-types"
	"github.com/kloudlite/operator/pkg/constants"
	rApi "github.com/kloudlite/operator/pkg/operator"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type BackupType string

const (
	BackupTypeOneShot BackupType = "one-shot"
	BackupTypeCron    BackupType = "cron"
)

// BackupSpec defines the desired state of Backup
type BackupSpec struct {
	MsvcRef ct.MsvcRef `json:"msvcRef"`

	BackupType `json:"backupType"`

	// must be set if backupType is cron
	Cron *CronSpec `json:"cron,omitempty"`
}

type CronSpec struct {
	// Cron schedule in cron format, refer https://crontab.guru
	Schedule string `json:"schedule"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Backup is the Schema for the backups API
type Backup struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   BackupSpec  `json:"spec,omitempty"`
	Status rApi.Status `json:"status,omitempty"`
}

func (bkp *Backup) EnsureGVK() {
	if bkp != nil {
		bkp.SetGroupVersionKind(GroupVersion.WithKind("Backup"))
	}
}

func (bkp *Backup) GetStatus() *rApi.Status {
	return &bkp.Status
}

func (bkp *Backup) GetEnsuredLabels() map[string]string {
	return map[string]string{
		constants.MsvcNameKey: bkp.Spec.MsvcRef.Name,
	}
}

func (bkp *Backup) GetEnsuredAnnotations() map[string]string {
	return map[string]string{}
}

//+kubebuilder:object:root=true

// BackupList contains a list of Backup
type BackupList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Backup `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Backup{}, &BackupList{})
}
