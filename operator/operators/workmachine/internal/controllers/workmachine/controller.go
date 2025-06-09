package workmachine

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"k8s.io/client-go/tools/record"

	ct "github.com/kloudlite/operator/apis/common-types"
	crdsv1 "github.com/kloudlite/operator/apis/crds/v1"
	"github.com/kloudlite/operator/operators/workmachine/internal/env"
	"github.com/kloudlite/operator/operators/workmachine/internal/templates"
	"github.com/kloudlite/operator/pkg/constants"
	"github.com/kloudlite/operator/pkg/ssh"
	fn "github.com/kloudlite/operator/toolkit/functions"
	"github.com/kloudlite/operator/toolkit/kubectl"
	rApi "github.com/kloudlite/operator/toolkit/reconciler"
	step_result "github.com/kloudlite/operator/toolkit/reconciler/step-result"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/yaml"
)

type Reconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Env    *env.Env

	YAMLClient kubectl.YAMLClient
	recorder   record.EventRecorder

	workmachineLifecycleTemplateSpec []byte
	templateWebhook                  []byte
	templateJumpServerDeploymentSpec []byte
}

func (r *Reconciler) GetName() string {
	return "workmachine"
}

const (
	createWorkMachineJob              string = "create-work-machine-job"
	createTargetNamespace             string = "create-target-namespace"
	createSSHPublicKeysSecret         string = "create-ssh-public-keys-secret"
	createMachinePublicPrivateKeyPair string = "create-machine-public-private-key-pair"
	createSSHJumpServerDeployment     string = "create-ssh-jumpserver-deployment"
)

const (
	sshPublicKeysSecretName string = "ssh-public-keys"
	authorizedKeysSecretKey string = "authorized_keys"
)

const (
	jobRefAnnotation string = "kloudlite.io/workmachine.job-ref"
)

// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=apps,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=apps/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=apps/finalizers,verbs=update

func (r *Reconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	req, err := rApi.NewRequest(ctx, r.Client, request.NamespacedName, &crdsv1.WorkMachine{})
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	req.PreReconcile()
	defer req.PostReconcile()

	req.Logger.Debug("RECONCILATION starting ...")

	if req.Object.GetDeletionTimestamp() != nil {
		if x := r.finalize(req); !x.ShouldProceed() {
			return x.ReconcilerResponse()
		}
		return ctrl.Result{}, nil
	}

	if step := req.ClearStatusIfAnnotated(); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.EnsureCheckList([]rApi.CheckMeta{
		{Name: createWorkMachineJob, Title: "Creates WorkMachine creation job"},
		{Name: createTargetNamespace, Title: "Creates a target namespace for workmachine"},
		{Name: createSSHPublicKeysSecret, Title: "store SSH public keys in a secret"},
	}); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.EnsureLabelsAndAnnotations(); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.EnsureFinalizers(rApi.ForegroundFinalizer, rApi.CommonFinalizer); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.createWorkMachineCreationJob(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.createTargetNamespace(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.createSSHPublicKeysSecret(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	// if step := r.createSSHJumpServer(req); !step.ShouldProceed() {
	// 	return step.ReconcilerResponse()
	// }

	req.Object.Status.IsReady = true
	return ctrl.Result{}, nil
}

func (r *Reconciler) finalize(req *rApi.Request[*crdsv1.WorkMachine]) step_result.Result {
	if step := req.EnsureCheckList([]rApi.CheckMeta{
		{Name: "uninstall workmachine"},
	}); !step.ShouldProceed() {
		return step
	}

	check := rApi.NewRunningCheck("uninstall workmachine", req)

	if step := req.CleanupOwnedResources(check); !step.ShouldProceed() {
		return step
	}

	return req.Finalize()
}

type ClusterParams struct {
	K3sServerHost string `json:"k3s_server_host"`

	K3sServerToken string `json:"k3s_server_token"`
	K3sAgentToken  string `json:"k3s_agent_token"`

	K3sVersion string `json:"k3s_version"`

	AwsVPCName        string `json:"aws_vpc_name"`
	AwsVPCId          string `json:"aws_vpc_id"`
	AwsPublicSubnet   string `json:"aws_public_subnet"`
	AwsRegion         string `json:"aws_region"`
	AwsAvailblityZone string `json:"aws_availblity_zone"`

	AwsNLBDNSHost string `json:"aws_nlb_dns_host"`

	AwsSecurityGroupIDs       []string `json:"aws_security_group_ids"`
	AwsIAMInstanceProfileName string   `json:"aws_iam_instance_profile_name"`
}

func (r *Reconciler) parseSpecIntoTFValues(ctx context.Context, obj *crdsv1.WorkMachine) ([]byte, error) {
	sp := strings.Split(r.Env.K3sParamsSecretRef, "/")
	if len(sp) != 2 {
		return nil, fmt.Errorf("invalid k3s params secret ref must be a valid <secret-namespace>/<secret-name> format")
	}

	secret, err := rApi.Get(ctx, r.Client, fn.NN(sp[0], sp[1]), &corev1.Secret{})
	if err != nil {
		return nil, err
	}

	fmt.Printf("cluster-params.yml: \n---\n%s\n---\n", string(secret.Data["cluster-params.yml"]))

	cp := ClusterParams{}
	if err := yaml.Unmarshal(secret.Data["cluster-params.yml"], &cp); err != nil {
		return nil, err
	}

	fmt.Printf("cluster params: %+v\n", cp)

	switch obj.Spec.GetCloudProvider() {
	case ct.CloudProviderAWS:
		{
			return json.Marshal(map[string]any{
				"aws_region":      cp.AwsRegion,
				"trace_id":        "workmachine-" + obj.Name,
				"vpc_id":          cp.AwsVPCId,
				"name":            obj.Name,
				"k3s_server_host": cp.K3sServerHost,
				"k3s_agent_token": cp.K3sAgentToken,
				"k3s_version":     cp.K3sVersion,
				"ami":             obj.Spec.AWSMachineConfig.AMI,
				"instance_type":   obj.Spec.AWSMachineConfig.InstanceType,
				"instance_state": func() string {
					if obj.Spec.State == crdsv1.WorkMachineStateOn {
						return "running"
					}

					return "stopped"
				}(),
				"availability_zone": cp.AwsAvailblityZone,
				// "iam_instance_profile": func() string {
				// 	if obj.Spec.AWSMachineConfig.IAMInstanceProfileRole != nil {
				// 		return *obj.Spec.AWSMachineConfig.IAMInstanceProfileRole
				// 	}
				// 	return cp.AwsIAMInstanceProfileName
				// }(),
				"root_volume_size":   obj.Spec.AWSMachineConfig.RootVolumeSize,
				"root_volume_type":   obj.Spec.AWSMachineConfig.RootVolumeType,
				"security_group_ids": cp.AwsSecurityGroupIDs,
				"subnet_id":          cp.AwsPublicSubnet,
			})
		}
	default:
		return nil, fmt.Errorf("unsupported cloud provider (%s)", obj.Spec.GetCloudProvider())
	}
}

func (r *Reconciler) createWorkMachineCreationJob(req *rApi.Request[*crdsv1.WorkMachine]) step_result.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.NewRunningCheck(createWorkMachineJob, req)

	jobName := fmt.Sprintf("wm-%s", obj.Name)

	if v, ok := obj.Annotations[jobRefAnnotation]; !ok || v != jobName {
		fn.MapSet(&obj.Annotations, jobRefAnnotation, jobName)
		if err := r.Update(ctx, obj); err != nil {
			return check.Failed(err)
		}
	}

	varfileJSON, err := r.parseSpecIntoTFValues(ctx, obj)
	if err != nil {
		return check.Failed(err)
	}

	b, err := templates.ParseBytes(r.workmachineLifecycleTemplateSpec, templates.WorkMachineLifecycleVars{
		JobMetadata: metav1.ObjectMeta{
			Name:            jobName,
			Namespace:       r.Env.IACJobsNamespace,
			Labels:          fn.MapFilterWithPrefix(obj.GetLabels(), "kloudlite.io/"),
			Annotations:     fn.FilterObservabilityAnnotations(obj.GetAnnotations()),
			OwnerReferences: []metav1.OwnerReference{fn.AsOwner(obj, true)},
		},
		NodeSelector:         obj.Spec.JobParams.NodeSelector,
		Tolerations:          obj.Spec.JobParams.Tolerations,
		JobImage:             r.Env.IACJobImage,
		TFWorkspaceName:      obj.Name,
		TfWorkspaceNamespace: r.Env.TFStateSecretNamespace,
		CloudProvider:        obj.Spec.GetCloudProvider().String(),
		ValuesJSON:           string(varfileJSON),

		OutputSecretName:      obj.Name + "-tf-outputs",
		OutputSecretNamespace: r.Env.IACJobsNamespace,

		NodeName: obj.Name,
	})
	if err != nil {
		return check.Failed(err)
	}

	lf := &crdsv1.Lifecycle{ObjectMeta: metav1.ObjectMeta{Name: jobName, Namespace: r.Env.IACJobsNamespace}}
	if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, lf, func() error {
		lf.SetLabels(fn.MapMerge(fn.MapFilterWithPrefix(obj.GetLabels(), "kloudlite.io/"), lf.GetLabels()))
		lf.SetAnnotations(fn.MapMerge(fn.MapFilterWithPrefix(obj.GetAnnotations(), "kloudlite.io/observability"), lf.GetAnnotations()))
		lf.SetOwnerReferences([]metav1.OwnerReference{fn.AsOwner(obj, true)})
		return yaml.Unmarshal(b, &lf.Spec)
	}); err != nil {
		return check.Failed(err)
	}

	if !lf.HasCompleted() {
		return check.StillRunning(fmt.Errorf("waiting for lifecycle job to complete"))
	}

	if lf.Status.Phase == crdsv1.JobPhaseFailed {
		return check.Failed(fmt.Errorf("lifecycle job failed"))
	}

	req.AddToOwnedResources(rApi.ParseResourceRef(lf))

	return check.Completed()
}

func (r *Reconciler) createTargetNamespace(req *rApi.Request[*crdsv1.WorkMachine]) step_result.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.NewRunningCheck(createTargetNamespace, req)

	hasUpdate := false
	if obj.Spec.TargetNamespace == "" {
		hasUpdate = true
		obj.Spec.TargetNamespace = "wm-" + obj.Name
	}

	if hasUpdate {
		if err := r.Update(ctx, obj); err != nil {
			return check.Failed(err)
		}

		return check.StillRunning(fmt.Errorf("waiting for reconcilation"))
	}

	ns := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: obj.Spec.TargetNamespace}}
	if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, ns, func() error {
		fn.MapSet(&ns.Annotations, "kloudlite.io/namespace.for", fmt.Sprintf("workmachine/%s", obj.Name))
		ns.SetOwnerReferences([]metav1.OwnerReference{fn.AsOwner(obj, true)})
		return nil
	}); err != nil {
		return check.Failed(err)
	}

	return check.Completed()
}

func (r *Reconciler) createSSHPublicKeysSecret(req *rApi.Request[*crdsv1.WorkMachine]) step_result.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.NewRunningCheck(createSSHPublicKeysSecret, req)

	secret := &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: sshPublicKeysSecretName, Namespace: obj.Spec.TargetNamespace}}
	if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, secret, func() error {
		fn.MapSet(&secret.Annotations, "kloudlite.io/description", "this secret contains ssh public keys given by user in workmachine resource")
		fn.MapSet(&secret.Annotations, "kloudlite.io/secret.for", fmt.Sprintf("workmachine/%s", obj.Name))
		secret.SetOwnerReferences([]metav1.OwnerReference{fn.AsOwner(obj, true)})

		if secret.StringData == nil {
			secret.StringData = make(map[string]string, 1)
		}

		secret.StringData[authorizedKeysSecretKey] = strings.Join(obj.Spec.SSHPublicKeys, "\n")
		return nil
	}); err != nil {
		return check.Failed(err)
	}

	if secret.Data["private_key"] == nil || secret.Data["public_key"] == nil {
		privateKeyPEM, publicKey, err := ssh.GenerateSSHKeyPair()
		if err != nil {
			return check.Failed(err)
		}

		if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, secret, func() error {
			if secret.Data == nil {
				secret.Data = make(map[string][]byte, 2)
			}
			secret.Data["public_key"] = publicKey
			secret.Data["private_key"] = privateKeyPEM
			return nil
		}); err != nil {
			return check.Failed(err)
		}
	}

	// if obj.Status.MachinePublicSSHKey == "" {
	// 	obj.Status.MachinePublicSSHKey = string(secret.Data["public_key"])
	// 	if err := r.Status().Update(ctx, obj); err != nil {
	// 		return check.Failed(err)
	// 	}
	// }

	return check.Completed()
}

func (r *Reconciler) createSSHJumpServer(req *rApi.Request[*crdsv1.WorkMachine]) step_result.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.NewRunningCheck(createSSHJumpServerDeployment, req)

	sshJumpServerName := "ssh-jump-server"

	b, err := templates.ParseBytes(r.templateJumpServerDeploymentSpec, templates.JumpServerDeploymentSpecTemplateArgs{
		SSHAuthorizedKeysSecretName: sshPublicKeysSecretName,
		SSHAuthorizedKeysSecretKey:  authorizedKeysSecretKey,
		SelectorLabels: map[string]string{
			"app": "jump-server",
		},
	})
	if err != nil {
		return check.Failed(err)
	}

	deployment := &appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{Name: sshJumpServerName, Namespace: obj.Spec.TargetNamespace}}
	if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, deployment, func() error {
		deployment.SetOwnerReferences([]metav1.OwnerReference{fn.AsOwner(obj, true)})
		fn.MapSet(&deployment.Annotations, constants.DescriptionKey, "this deployment is a ssh jump server used to allow users to jump to different workspaces")
		return yaml.Unmarshal(b, &deployment.Spec)
	}); err != nil {
		return check.Failed(err)
	}

	return check.Completed()
}

func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.Client = mgr.GetClient()
	r.Scheme = mgr.GetScheme()

	if r.YAMLClient == nil {
		return fmt.Errorf("yaml client must be set")
	}

	r.recorder = mgr.GetEventRecorderFor(r.GetName())

	var err error
	r.workmachineLifecycleTemplateSpec, err = templates.Read(templates.WorkMachineLifecycleTemplate)
	if err != nil {
		return err
	}

	r.templateJumpServerDeploymentSpec, err = templates.Read(templates.JumpServerDeploymentSpec)
	if err != nil {
		return err
	}

	builder := ctrl.NewControllerManagedBy(mgr).For(&crdsv1.WorkMachine{})
	builder.WithOptions(controller.Options{MaxConcurrentReconciles: r.Env.MaxConcurrentReconciles})
	builder.Owns(&crdsv1.Lifecycle{})
	builder.WithEventFilter(rApi.ReconcileFilter(r.recorder))
	return builder.Complete(r)
}
