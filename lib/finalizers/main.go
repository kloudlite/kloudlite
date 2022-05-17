package finalizers

type Finalizer string

const (
	Project           Finalizer = "finalizers.kloudlite.io/project"
	App               Finalizer = "finalizers.kloudlite.io/app"
	Router            Finalizer = "finalizers.kloudlite.io/router"
	ManagedService    Finalizer = "finalizers.kloudlite.io/managed-service"
	ManagedResource   Finalizer = "finalizers.kloudlite.io/managed-resource"
	MsvcCommonService Finalizer = "finalizers.kloudlite.io/msvc-common-service"
	// for foreground deletion of child resources
	Foreground Finalizer = "foreground"
)

func (f Finalizer) String() string {
	return string(f)
}
