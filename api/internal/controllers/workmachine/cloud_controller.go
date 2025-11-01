package workmachine

// import (
// 	"github.com/kloudlite/kloudlite/api/internal/controllers/workmachine/v1"
// 	"github.com/kloudlite/kloudlite/api/pkg/operator-toolkit/reconciler"
// )

// func (r *WorkMachineReconciler) handleVolumeSizeChange(
// 	check *reconciler.Check[*v1.WorkMachine],
// 	obj *v1.WorkMachine,
// ) reconciler.StepResult {
// 	var newSize int
//
// 	switch r.env.CloudProvider {
// 	case v1.AWS:
// 		{
// 			if obj.Spec.AWSProvider.VolumeSize > obj.Status.RootVolumeSize {
// 				r.cloudProviderAPI.IncreaseVolumeSize(check.Context(), obj.Status.MachineID, obj.Spec.AWSProvider.VolumeSize)
// 			}
// 		}
// 	}
//
// 	return check.Passed()
// }
//
// func (r *WorkMachineReconciler) handleMachineState(
// 	check *reconciler.Check[*v1.WorkMachine],
// 	obj *v1.WorkMachine,
// ) reconciler.StepResult {
// 	switch obj.Spec.Provider {
// 	case v1.AWS:
// 		{
// 			if obj.Spec.AWSProvider.VolumeSize > obj.Status.RootVolumeSize {
// 				r.cloudProviderAPI.IncreaseVolumeSize(check.Context(), obj.Status.MachineID, obj.Spec.AWSProvider.VolumeSize)
// 			}
// 		}
// 	}
//
// 	return check.Passed()
// }
