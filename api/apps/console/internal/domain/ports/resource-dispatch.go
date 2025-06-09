package ports

// import (
// 	"sigs.k8s.io/controller-runtime/pkg/client"
// )
//
// type ResourceDispatcher interface {
// 	ApplyResource(ctx InfraContext, clusterName string, obj client.Object, recordVersion int) error
// 	DeleteResource(ctx InfraContext, clusterName string, obj client.Object) error
// 	RestartResource(ctx InfraContext, clusterName string, obj client.Object) error
// }
//
// type ResourceUpdatesReceiver interface {
// 	OnResourceUpdate(ctx InfraContext, clusterName string, obj client.Object, recordVersion int) error
// }
