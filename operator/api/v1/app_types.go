package v1

import (
	meta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ContainerResource struct {
	Min string `json:"min"`
	Max string `json:"max"`
}

type ContainerEnv struct {
	Key     string `json:"key"`
	Value   string `json:"value,omitempty"`
	Type    string `json:"type,omitempty"`
	RefName string `json:"refName,omitempty"`
	RefKey  string `json:"refKey,omitempty"`
}

type ContainerVolumeItem struct {
	Key      string `json:"key"`
	FileName string `json:"fileName"`
}

type ContainerVolume struct {
	Name      string                `json:"name"`
	MountPath string                `json:"mountPath"`
	Type      string                `json:"type"`
	RefName   string                `json:"refName"`
	Items     []ContainerVolumeItem `json:"items"`
}

type ImageFromGit struct {
}

type AppContainer struct {
	Name            string            `json:"name"`
	Image           string            `json:"image"`
	ImagePullPolicy string            `json:"imagePullPolicy"`
	Command         []string          `json:"command,omitempty"`
	Args            []string          `json:"args,omitempty"`
	ResourceCpu     ContainerResource `json:"resourceCpu"`
	ResourceMemory  ContainerResource `json:"resourceMemory"`
	Env             []ContainerEnv    `json:"env,omitempty"`
	Volumes         []ContainerVolume `json:"volumes,omitempty"`
}

type AppSvc struct {
	Port       uint16 `json:"port"`
	TargetPort uint16 `json:"targetPort,omitempty"`
	Type       string `json:"type,omitempty"`
}

// AppSpec defines the desired state of App
type AppSpec struct {
	Services   []AppSvc       `json:"services,omitempty"`
	Containers []AppContainer `json:"containers"`
}

type ReconPod struct {
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
	Failed    string `json:"failed"`
}

// AppStatus defines the observed state of App
type AppStatus struct {
	Generation         int64              `json:"generation"`
	DependencyCheck    Recon              `json:"dependency_check,omitempty"`
	ApplyJob           Recon              `json:"apply_job,omitempty"`
	DeleteJob          Recon              `json:"delete_job,omitempty"`
	ImagesAvailability Recon              `json:"images_availability,omitempty"`
	Conditions         []metav1.Condition `json:"conditions,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// App is the Schema for the apps API
type App struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AppSpec   `json:"spec,omitempty"`
	Status AppStatus `json:"status,omitempty"`
}

func (app *App) DefaultStatus() {
	app.Status.Generation = app.Generation
	app.Status.DependencyCheck = Recon{}
	app.Status.ImagesAvailability = Recon{}
	app.Status.ApplyJob = Recon{}
}

func (app *App) IsNewGeneration() bool {
	return app.Generation > app.Status.Generation
}

func (app *App) HasToBeDeleted() bool {
	return app.GetDeletionTimestamp() != nil
}

func (app *App) BuildConditions() {
	// DependencyCheck
	meta.SetStatusCondition(&app.Status.Conditions, metav1.Condition{
		Type:               "DependencyCheck",
		Status:             app.Status.DependencyCheck.ConditionStatus(),
		ObservedGeneration: app.Generation,
		Reason:             app.Status.DependencyCheck.Reason(),
		Message:            app.Status.DependencyCheck.Message,
	})

	// Images Available
	meta.SetStatusCondition(&app.Status.Conditions, metav1.Condition{
		Type:               "ImagesAvailable",
		Status:             app.Status.ImagesAvailability.ConditionStatus(),
		ObservedGeneration: app.Generation,
		Reason:             app.Status.ImagesAvailability.Reason(),
		Message:            app.Status.ImagesAvailability.Message,
	})

	// Ready
	meta.SetStatusCondition(&app.Status.Conditions, metav1.Condition{
		Type:               "Ready",
		Status:             app.Status.ApplyJob.ConditionStatus(),
		ObservedGeneration: app.Generation,
		Reason:             app.Status.ApplyJob.Reason(),
		Message:            app.Status.ApplyJob.Message,
	})
}

//+kubebuilder:object:root=true

// AppList contains a list of App
type AppList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []App `json:"items"`
}

func init() {
	SchemeBuilder.Register(&App{}, &AppList{})
}
