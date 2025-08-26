package controller

import (
	"context"
	"fmt"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"

	"github.com/codingconcepts/env"
	fn "github.com/kloudlite/kloudlite/operator/toolkit/functions"
	"github.com/kloudlite/kloudlite/operator/toolkit/kubectl"
	"github.com/kloudlite/kloudlite/operator/toolkit/reconciler"
	"github.com/kloudlite/operator/api/v1"
	"github.com/kloudlite/operator/internal/workmachine/internal/templates"
	"github.com/kloudlite/operator/pkg/ssh"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

type envVars struct {
	MaxConcurrentReconciles int `env:"MAX_CONCURRENT_RECONCILES" default:"5"`

	ImageSSH string `env:"WORKMACHINE_IMAGE_SSH" default:"ghcr.io/kloudlite/kloudlite/operator/workspace-ssh:debug"`

	// K3sParamsSecretRef      string `env:"K3S_PARAMS_SECRET_REF" required:"true"`
	//
	// IACJobsNamespace string `env:"IAC_JOBS_NAMESPACE" required:"true"`
	// IACJobImage      string `env:"IAC_JOB_IMAGE" required:"true"`
	//
	// TFStateSecretNamespace string `env:"TF_STATE_SECRET_NAMESPACE" required:"true" default:"kloudlite"`
	//
	// K3sVersion    string `env:"WORKMACHINE_K3S_VERSION" required:"true"`
	// K3sServerHost string `env:"WORKMACHINE_K3S_SERVER_HOST" required:"true"`
	// K3sAgentToken string `env:"WORKMACHINE_K3S_AGENT_TOKEN" required:"true"`
	//
	// AwsVpcName      string `env:"WORKMACHINE_AWS_VPC_NAME" required:"true"`
	// AwsVpcID        string `env:"WORKMACHINE_AWS_VPC_ID" required:"true"`
	// AwsRegion       string `env:"WORKMACHINE_AWS_REGION" required:"true"`
	// AwsPublicSubnet string `env:"WORKMACHINE_AWS_PUBLIC_SUBNET" required:"true"`
	// AwsNLBDnsHost   string `env:"WORKMACHINE_AWS_NLB_DNS_HOST" required:"true"`
}

// WorkmachineReconciler reconciles a Workmachine object
type Reconciler struct {
	client.Client
	Scheme *runtime.Scheme

	env envVars

	YAMLClient                   kubectl.YAMLClient
	templateJumpServerDeployment []byte
}

func (r *Reconciler) GetName() string {
	return v1.ProjectDomain + "/workmachine-controller"
}

// +kubebuilder:rbac:groups=kloudlite.io,resources=workmachines,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=kloudlite.io,resources=workmachines/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=kloudlite.io,resources=workmachines/finalizers,verbs=update

func (r *Reconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	req, err := reconciler.NewRequest(ctx, r.Client, request.NamespacedName, &v1.Workmachine{})
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	req.PreReconcile()
	defer req.PostReconcile()

	return reconciler.ReconcileSteps(req, []reconciler.Step[*v1.Workmachine]{
		{
			Name:     "setup namespace",
			Title:    "Setup Namespace",
			OnCreate: r.createNamespace,
			OnDelete: r.cleanupNamespace,
		},
		{
			Name:     "setup ssh keys",
			Title:    "Setup SSH Keys",
			OnCreate: r.createSSHKeys,
			OnDelete: r.cleanupSSHKeys,
		},
		{
			Name:     "setup ssh jump server",
			Title:    "Setup SSH Jump Server",
			OnCreate: r.createSSHJumpServer,
			OnDelete: r.cleanupSSHJumpServer,
		},
	})
}

func (r *Reconciler) createNamespace(check *reconciler.Check[*v1.Workmachine], obj *v1.Workmachine) reconciler.StepResult {
	if obj.Spec.TargetNamespace == "" {
		obj.Spec.TargetNamespace = "wm-" + obj.Name
		if err := r.Update(check.Context(), obj); err != nil {
			return check.Failed(err)
		}

		return check.Abort("waiting for .spec.targetNamespace to be set").RequeueAfter(100 * time.Millisecond)
	}

	ns := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: obj.Spec.TargetNamespace}}
	if _, err := controllerutil.CreateOrUpdate(check.Context(), r.Client, ns, func() error {
		ns.SetAnnotations(fn.MapMerge(ns.GetAnnotations(), map[string]string{
			reconciler.AnnotationDescriptionKey: fmt.Sprintf("created by %s for storing workmachine related stuffs", r.GetName()),
		}))
		return nil
	}); err != nil {
		return check.Failed(err)
	}

	return check.Passed()
}

func (r *Reconciler) cleanupNamespace(check *reconciler.Check[*v1.Workmachine], obj *v1.Workmachine) reconciler.StepResult {
	if obj.Spec.TargetNamespace == "" {
		return check.Passed()
	}

	ns := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: obj.Spec.TargetNamespace}}
	if err := fn.DeleteAndWait(check.Context(), r.Client, ns); err != nil {
		return check.Failed(err)
	}

	return check.Passed()
}

func (r *Reconciler) createSSHKeys(check *reconciler.Check[*v1.Workmachine], obj *v1.Workmachine) reconciler.StepResult {
	if obj.Spec.SSH.Secret.Name == "" {
		obj.Spec.SSH.Secret.Name = obj.Name + "-ssh-keypair"
		if err := r.Update(check.Context(), obj); err != nil {
			return check.Failed(err)
		}
	}

	secret := &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: obj.Spec.SSH.Secret.Name, Namespace: obj.Spec.TargetNamespace}}
	if _, err := controllerutil.CreateOrUpdate(check.Context(), r.Client, secret, func() error {
		secret.SetAnnotations(fn.MapMerge(secret.GetAnnotations(), map[string]string{
			reconciler.AnnotationDescriptionKey: fmt.Sprintf("this secret contains ssh public keys given by user in workmachine (%s) resource", obj.Name),
		}))
		secret.SetOwnerReferences([]metav1.OwnerReference{fn.AsOwner(obj, true)})

		if secret.StringData == nil {
			secret.StringData = make(map[string]string, 1)
		}

		secret.StringData["authorized_keys"] = strings.Join(obj.Spec.SSH.PublicKeys, "\n")
		return nil
	}); err != nil {
		return check.Failed(err)
	}

	if secret.Data["private_key"] == nil || secret.Data["public_key"] == nil {
		privateKeyPEM, publicKey, err := ssh.GenerateSSHKeyPair()
		if err != nil {
			return check.Failed(err)
		}

		if _, err := controllerutil.CreateOrUpdate(check.Context(), r.Client, secret, func() error {
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

	return check.Passed()
}

func (r *Reconciler) cleanupSSHKeys(check *reconciler.Check[*v1.Workmachine], obj *v1.Workmachine) reconciler.StepResult {
	if obj.Spec.SSH.Secret.Name == "" {
		return check.Passed()
	}

	if err := fn.DeleteAndWait(check.Context(), r.Client, &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      obj.Spec.SSH.Secret.Name,
			Namespace: obj.Namespace,
		},
	}); err != nil {
		return check.Failed(err)
	}

	return check.Passed()
}

func (r *Reconciler) createSSHJumpServer(check *reconciler.Check[*v1.Workmachine], obj *v1.Workmachine) reconciler.StepResult {
	sshJumpServerName := obj.Name + "-ssh-server"

	b, err := templates.ParseBytes(r.templateJumpServerDeployment, templates.JumpServerDeploymentTemplateArgs{
		Metadata: metav1.ObjectMeta{
			Name:      sshJumpServerName,
			Namespace: obj.Spec.TargetNamespace,
		},
		SSH: obj.Spec.SSH,
		SelectorLabels: map[string]string{
			"app": "jump-server",
		},
		ImageSSHServer:           r.env.ImageSSH,
		WorkMachineTolerationKey: v1.WorkMachineNameKey,
		WorkMachineName:          obj.Name,
	})
	if err != nil {
		return check.Failed(err)
	}

	objects, err := r.YAMLClient.ApplyYAML(check.Context(), b)
	if err != nil {
		return check.Failed(err)
	}

	if len(objects) != 1 {
		return check.Failed(fmt.Errorf("expected 1 object returned from YAMLClient.ApplyYAML, but got %d", len(objects)))
	}

	deployment, err := fn.FromUnstructured(objects[0], &appsv1.Deployment{})
	if err != nil {
		return check.Failed(err)
	}

	return check.ValidateDeploymentReady(deployment)
}

func (r *Reconciler) cleanupSSHJumpServer(check *reconciler.Check[*v1.Workmachine], obj *v1.Workmachine) reconciler.StepResult {
	sshJumpServerName := obj.Name + "-ssh-server"

	if err := fn.DeleteAndWait(check.Context(), r.Client, &appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{Name: sshJumpServerName, Namespace: obj.Spec.TargetNamespace}}); err != nil {
		return check.Failed(err)
	}

	return check.Passed()
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

// func (r *Reconciler) parseSpecIntoTFValues(ctx context.Context, obj *v1.Workmachine) ([]byte, error) {
// 	sp := strings.Split(r.Env.K3sParamsSecretRef, "/")
// 	if len(sp) != 2 {
// 		return nil, fmt.Errorf("invalid k3s params secret ref must be a valid <secret-namespace>/<secret-name> format")
// 	}
//
// 	secret, err := rApi.Get(ctx, r.Client, fn.NN(sp[0], sp[1]), &corev1.Secret{})
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	fmt.Printf("cluster-params.yml: \n---\n%s\n---\n", string(secret.Data["cluster-params.yml"]))
//
// 	cp := ClusterParams{}
// 	if err := yaml.Unmarshal(secret.Data["cluster-params.yml"], &cp); err != nil {
// 		return nil, err
// 	}
//
// 	fmt.Printf("cluster params: %+v\n", cp)
//
// 	switch obj.Spec.GetCloudProvider() {
// 	case ct.CloudProviderAWS:
// 		{
// 			return json.Marshal(map[string]any{
// 				"aws_region":      cp.AwsRegion,
// 				"trace_id":        "workmachine-" + obj.Name,
// 				"vpc_id":          cp.AwsVPCId,
// 				"name":            obj.Name,
// 				"k3s_server_host": cp.K3sServerHost,
// 				"k3s_agent_token": cp.K3sAgentToken,
// 				"k3s_version":     cp.K3sVersion,
// 				"ami":             obj.Spec.AWSMachineConfig.AMI,
// 				"instance_type":   obj.Spec.AWSMachineConfig.InstanceType,
// 				"instance_state": func() string {
// 					if obj.Spec.State == crdsv1.WorkMachineStateOn {
// 						return "running"
// 					}
//
// 					return "stopped"
// 				}(),
// 				"availability_zone": cp.AwsAvailblityZone,
// 				// "iam_instance_profile": func() string {
// 				// 	if obj.Spec.AWSMachineConfig.IAMInstanceProfileRole != nil {
// 				// 		return *obj.Spec.AWSMachineConfig.IAMInstanceProfileRole
// 				// 	}
// 				// 	return cp.AwsIAMInstanceProfileName
// 				// }(),
// 				"root_volume_size":   obj.Spec.AWSMachineConfig.RootVolumeSize,
// 				"root_volume_type":   obj.Spec.AWSMachineConfig.RootVolumeType,
// 				"security_group_ids": cp.AwsSecurityGroupIDs,
// 				"subnet_id":          cp.AwsPublicSubnet,
// 			})
// 		}
// 	default:
// 		return nil, fmt.Errorf("unsupported cloud provider (%s)", obj.Spec.GetCloudProvider())
// 	}
// }
//
// func (r *Reconciler) createWorkMachineCreationJob(check *reconciler.Check[*v1.Workmachine], obj *v1.Workmachine) reconciler.StepResult {
// 	if obj.Spec.Type == v1.NoOpWorkmachine {
// 		return check.Passed()
// 	}
//
// 	jobName := fmt.Sprintf("wm-%s-job", obj.Name)
//
// 	if v, ok := obj.Annotations[jobRefAnnotation]; !ok || v != jobName {
// 		fn.MapSet(&obj.Annotations, jobRefAnnotation, jobName)
// 		if err := r.Update(check.Context(), obj); err != nil {
// 			return check.Failed(err)
// 		}
// 	}
//
// 	varfileJSON, err := r.parseSpecIntoTFValues(ctx, obj)
// 	if err != nil {
// 		return check.Failed(err)
// 	}
//
// 	b, err := templates.ParseBytes(r.workmachineLifecycleTemplateSpec, templates.WorkMachineLifecycleVars{
// 		JobMetadata: metav1.ObjectMeta{
// 			Name:            jobName,
// 			Namespace:       r.Env.IACJobsNamespace,
// 			Labels:          fn.MapFilterWithPrefix(obj.GetLabels(), v1.ProjectDomain),
// 			Annotations:     fn.FilterObservabilityAnnotations(obj.GetAnnotations()),
// 			OwnerReferences: []metav1.OwnerReference{fn.AsOwner(obj, true)},
// 		},
// 		NodeSelector:         obj.Spec.JobParams.NodeSelector,
// 		Tolerations:          obj.Spec.JobParams.Tolerations,
// 		JobImage:             r.Env.IACJobImage,
// 		TFWorkspaceName:      obj.Name,
// 		TfWorkspaceNamespace: r.Env.TFStateSecretNamespace,
// 		CloudProvider:        obj.Spec.GetCloudProvider().String(),
// 		ValuesJSON:           string(varfileJSON),
//
// 		OutputSecretName:      obj.Name + "-tf-outputs",
// 		OutputSecretNamespace: r.Env.IACJobsNamespace,
//
// 		NodeName: obj.Name,
// 	})
// 	if err != nil {
// 		return check.Failed(err)
// 	}
//
// 	lf := &crdsv1.Lifecycle{ObjectMeta: metav1.ObjectMeta{Name: jobName, Namespace: r.Env.IACJobsNamespace}}
// 	if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, lf, func() error {
// 		lf.SetLabels(fn.MapMerge(fn.MapFilterWithPrefix(obj.GetLabels(), "kloudlite.io/"), lf.GetLabels()))
// 		lf.SetAnnotations(fn.MapMerge(fn.MapFilterWithPrefix(obj.GetAnnotations(), "kloudlite.io/observability"), lf.GetAnnotations()))
// 		lf.SetOwnerReferences([]metav1.OwnerReference{fn.AsOwner(obj, true)})
// 		return yaml.Unmarshal(b, &lf.Spec)
// 	}); err != nil {
// 		return check.Failed(err)
// 	}
//
// 	if !lf.HasCompleted() {
// 		return check.StillRunning(fmt.Errorf("waiting for lifecycle job to complete"))
// 	}
//
// 	if lf.Status.Phase == crdsv1.JobPhaseFailed {
// 		return check.Failed(fmt.Errorf("lifecycle job failed"))
// 	}
//
// 	req.AddToOwnedResources(rApi.ParseResourceRef(lf))
//
// 	return check.Completed()
// }

// SetupWithManager sets up the controller with the Manager.
func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.Client = mgr.GetClient()
	r.Scheme = mgr.GetScheme()

	if err := env.Set(&r.env); err != nil {
		return fmt.Errorf("failed to set env: %w", err)
	}

	if r.YAMLClient == nil {
		return fmt.Errorf("yaml client must be set")
	}

	var err error
	// r.workmachineLifecycleTemplateSpec, err = templates.Read(templates.WorkMachineLifecycleTemplate)
	// if err != nil {
	// 	return err
	// }
	//
	r.templateJumpServerDeployment, err = templates.Read(templates.JumpServerDeployment)
	if err != nil {
		return err
	}

	builder := ctrl.NewControllerManagedBy(mgr).For(&v1.Workmachine{})
	builder.WithOptions(controller.Options{MaxConcurrentReconciles: r.env.MaxConcurrentReconciles})
	// builder.Owns(&v1.Lifecycle{})
	builder.WithEventFilter(reconciler.ReconcileFilter(mgr.GetEventRecorderFor(r.GetName())))
	return builder.Complete(r)
}
