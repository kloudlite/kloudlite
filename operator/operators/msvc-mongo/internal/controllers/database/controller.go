package database

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	mongodbMsvcv1 "github.com/kloudlite/operator/apis/mongodb.msvc/v1"
	"github.com/kloudlite/operator/operators/msvc-mongo/internal/env"
	"github.com/kloudlite/operator/operators/msvc-mongo/internal/types"
	"github.com/kloudlite/operator/pkg/constants"
	"github.com/kloudlite/operator/pkg/errors"
	fn "github.com/kloudlite/operator/pkg/functions"
	"github.com/kloudlite/operator/pkg/kubectl"
	"github.com/kloudlite/operator/pkg/logging"
	libMongo "github.com/kloudlite/operator/pkg/mongo"
	rApi "github.com/kloudlite/operator/pkg/operator"
	stepResult "github.com/kloudlite/operator/pkg/operator/step-result"
	"github.com/kloudlite/operator/pkg/templates"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	apiErrors "k8s.io/apimachinery/pkg/api/errors"
)

type Reconciler struct {
	client.Client
	Scheme     *runtime.Scheme
	logger     logging.Logger
	Name       string
	Env        *env.Env
	yamlClient kubectl.YAMLClient

	templateJobUserCreate []byte
}

func (r *Reconciler) GetName() string {
	return r.Name
}

const (
	AccessCredsReady string = "access-creds"
	DBUserReady      string = "db-user-ready"
	IsOwnedByMsvc    string = "is-owned-by-msvc"

	DBUserDeleted   string = "db-user-deleted"
	DefaultsPatched string = "defaults-patched"
)

const (
	KeyMsvcOutput string = "msvc-output"
	KeyMresOutput string = "mres-output"
)

const (
	LabelResourceGeneration = "job-resource-generation"
	LabelUserCreateJob      = "user-create-job"
	LabelUserRemoveJob      = "user-remove-job"
)

func (r *Reconciler) newMongoContext(parent context.Context) (context.Context, context.CancelFunc) {
	if r.Env.IsDev {
		return context.WithCancel(parent)
	}
	return context.WithTimeout(parent, 5*time.Second)
}

// +kubebuilder:rbac:groups=mongodb.msvc.kloudlite.io,resources=databases,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=mongodb.msvc.kloudlite.io,resources=databases/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=mongodb.msvc.kloudlite.io,resources=databases/finalizers,verbs=update

func (r *Reconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	req, err := rApi.NewRequest(rApi.NewReconcilerCtx(ctx, r.logger), r.Client, request.NamespacedName, &mongodbMsvcv1.Database{})
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	req.PreReconcile()
	defer req.PostReconcile()

	if req.Object.GetDeletionTimestamp() != nil {
		if x := r.finalize(req); !x.ShouldProceed() {
			return x.ReconcilerResponse()
		}
		return ctrl.Result{}, nil
	}

	if step := req.ClearStatusIfAnnotated(); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.EnsureLabelsAndAnnotations(); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.EnsureFinalizers(constants.ForegroundFinalizer, constants.CommonFinalizer); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.reconDBCreds(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	req.Object.Status.IsReady = true
	return ctrl.Result{}, nil
}

func (r *Reconciler) finalize(req *rApi.Request[*mongodbMsvcv1.Database]) stepResult.Result {
	ctx, obj := req.Context(), req.Object

	check := rApi.NewRunningCheck("finalizing", req)

	msvcOutput, err := r.getMsvcConnectionParams(ctx, obj)
	if err != nil {
		if apiErrors.IsNotFound(err) {
			return req.Finalize()
		}
		return check.Failed(err)
	}

	mctx, cancel := r.newMongoContext(ctx)
	defer cancel()
	uri := msvcOutput.ClusterLocalURI
	if obj.IsGlobalVPNEnabled() {
		uri = msvcOutput.GlobalVpnURI
	}
	mongoCli, err := libMongo.NewClient(mctx, uri)
	if err != nil {
		return check.Failed(err)
	}
	defer mongoCli.Close()

	if err := mongoCli.DeleteUser(ctx, obj.Name, obj.Name); err != nil {
		return check.Failed(err)
	}

	return req.Finalize()
}

type MsvcOutput struct {
	ClusterLocalHosts string
	ClusterLocalURI   string

	GlobalVPNHosts string
	GlobalVpnURI   string

	ReplicasSetName *string
}

func (r *Reconciler) getGlobalVPNConnParams(ctx context.Context, obj *mongodbMsvcv1.Database) (map[string][]byte, error) {
	if !obj.IsGlobalVPNEnabled() {
		return nil, fmt.Errorf("global VPN support is not enabled")
	}

	if obj.Spec.MsvcRef.ClusterName == nil {
		return nil, fmt.Errorf(".spec.msvcRef.clusterName must be set")
	}

	svcURL := fmt.Sprintf("http://%s.%s.svc.%s.local", r.Env.MsvcCredsSvcName, r.Env.MsvcCredsSvcNamespace, *obj.Spec.MsvcRef.ClusterName)
	svcURL, err := url.JoinPath(svcURL, r.Env.MsvcCredsSvcRequestPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, svcURL, nil)
	if err != nil {
		return nil, err
	}

	qp := req.URL.Query()
	qp.Add("name", obj.Spec.MsvcRef.Name)
	qp.Add("namespace", func() string {
		if obj.Spec.MsvcRef.Namespace != "" {
			return obj.Spec.MsvcRef.Namespace
		}
		return obj.Namespace
	}())
	req.URL.RawQuery = qp.Encode()

	if obj.Spec.MsvcRef.SharedSecret != nil {
		req.Header.Add("kloudlite-shared-secret", *obj.Spec.MsvcRef.SharedSecret)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	m := map[string]string{}
	err = json.Unmarshal(b, &m)
	if err != nil {
		return nil, errors.NewEf(err, "unmarshalling %q msvc creds into map[string]string", b)
	}

	result := make(map[string][]byte, len(m))

	for k := range m {
		b, err := base64.StdEncoding.DecodeString(m[k])
		if err != nil {
			return nil, errors.NewEf(err, "base64 decoding %q", m[k])
		}
		result[k] = b
	}

	return result, nil
}

func (r *Reconciler) getMsvcConnectionParams(ctx context.Context, obj *mongodbMsvcv1.Database) (*MsvcOutput, error) {
	switch obj.Spec.MsvcRef.Kind {
	case mongodbMsvcv1.StandaloneServiceKind:
		{
			m, err := func() (map[string][]byte, error) {
				if obj.IsGlobalVPNEnabled() {
					return r.getGlobalVPNConnParams(ctx, obj)
				}

				msvc, err := rApi.Get(ctx, r.Client, fn.NN(obj.Spec.MsvcRef.Namespace, obj.Spec.MsvcRef.Name), &mongodbMsvcv1.StandaloneService{})
				if err != nil {
					return nil, err
				}

				s, err := rApi.Get(ctx, r.Client, fn.NN(obj.Spec.MsvcRef.Namespace, msvc.Output.CredentialsRef.Name), &corev1.Secret{})
				if err != nil {
					return nil, err
				}

				return s.Data, nil
			}()
			if err != nil {
				return nil, err
			}

			sso, err := fn.ParseFromSecretData[types.StandaloneSvcOutput](m)
			if err != nil {
				return nil, errors.NewEf(err, "unmarshalling msvc creds into types.StandaloneSvcOutput")
			}

			return &MsvcOutput{
				ClusterLocalHosts: sso.ClusterLocalHosts,
				ClusterLocalURI:   sso.ClusterLocalURI,
				GlobalVPNHosts:    sso.GlobalVPNHosts,
				GlobalVpnURI:      sso.GlobalVpnURI,
			}, nil
		}
	case mongodbMsvcv1.ClusterServiceKind:
		{
			m, err := func() (map[string][]byte, error) {
				if obj.IsGlobalVPNEnabled() {
					return r.getGlobalVPNConnParams(ctx, obj)
				}

				msvc, err := rApi.Get(ctx, r.Client, fn.NN(obj.Spec.MsvcRef.Namespace, obj.Spec.MsvcRef.Name), &mongodbMsvcv1.ClusterService{})
				if err != nil {
					return nil, err
				}

				s, err := rApi.Get(ctx, r.Client, fn.NN(msvc.Namespace, msvc.Output.CredentialsRef.Name), &corev1.Secret{})
				if err != nil {
					return nil, err
				}

				return s.Data, nil
			}()
			if err != nil {
				return nil, err
			}
			cso, err := fn.JsonConvert[types.ClusterSvcOutput](m)
			if err != nil {
				return nil, err
			}

			return &MsvcOutput{
				ClusterLocalHosts: cso.ClusterLocalHosts,
				ClusterLocalURI:   cso.ClusterLocalURI,
				GlobalVPNHosts:    cso.GlobalVpnHosts,
				GlobalVpnURI:      cso.GlobalVpnURI,

				ReplicasSetName: &cso.ReplicasSetName,
			}, nil
		}
	default:
		return nil, fmt.Errorf("unknown msvc kind: %s", obj.Spec.MsvcRef.Kind)
	}
}

func (r *Reconciler) reconDBCreds(req *rApi.Request[*mongodbMsvcv1.Database]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.NewRunningCheck(AccessCredsReady, req)

	secretName := obj.Output.CredentialsRef.Name
	secretNamespace := obj.Namespace

	scrt, err := rApi.Get(ctx, r.Client, fn.NN(secretNamespace, secretName), &corev1.Secret{})
	if err != nil {
		req.Logger.Infof("access credentials %s/%s does not exist, will be creating it now...", secretNamespace, secretName)
	}

	msvcOutput, err := r.getMsvcConnectionParams(ctx, obj)
	if err != nil {
		return check.Failed(err).Err(nil)
	}

	shouldGeneratePassword := scrt == nil

	if scrt != nil {
		shouldGeneratePassword = false
		mresOutput, err := fn.ParseFromSecret[types.DatabaseOutput](scrt)
		if err != nil {
			return check.Failed(err).Err(nil)
		}

		uri := mresOutput.ClusterLocalURI
		if obj.IsGlobalVPNEnabled() {
			uri = mresOutput.GlobalVpnURI
		}

		err = libMongo.ConnectAndPing(ctx, uri)
		if err != nil {
			if !libMongo.FailsWithAuthError(err) {
				return check.Failed(err)
			}
			req.Logger.Infof("Invalid Credentials in secret's .data.GlobalVpnURI, would need to be regenerated as connection failed with auth error")
			shouldGeneratePassword = true
		}
	}

	if shouldGeneratePassword {
		dbPasswd := fn.CleanerNanoid(40)

		if obj.Spec.MsvcRef.Kind == "ClusterService" && msvcOutput.ReplicasSetName == nil {
			return check.Failed(fmt.Errorf("%s: MsvcRef.Kind is ClusterService but MsvcRef.ReplicasSetName is nil", obj.Name))
		}

		mresOutput := types.DatabaseOutput{
			Username:          obj.Name,
			Password:          dbPasswd,
			ClusterLocalHosts: msvcOutput.ClusterLocalHosts,
			DbName:            obj.Name,
			ClusterLocalURI: func() string {
				baseURI := fmt.Sprintf("mongodb://%s:%s@%s/%s?authSource=%s", obj.Name, dbPasswd, msvcOutput.ClusterLocalHosts, obj.Name, obj.Name)
				if obj.Spec.MsvcRef.Kind == "ClusterService" {
					return baseURI + fmt.Sprintf("&replicaSet=%s", *msvcOutput.ReplicasSetName)
				}
				return baseURI
			}(),
		}
		if obj.IsGlobalVPNEnabled() {
			mresOutput.GlobalVpnURI = func() string {
				baseURI := fmt.Sprintf("mongodb://%s:%s@%s/%s?authSource=%s", obj.Name, dbPasswd, msvcOutput.GlobalVPNHosts, obj.Name, obj.Name)
				if obj.Spec.MsvcRef.Kind == "ClusterService" {
					return baseURI + fmt.Sprintf("&replicaSet=%s", *msvcOutput.ReplicasSetName)
				}
				return baseURI
			}()
		}

		b2, err := templates.Parse(
			templates.Secret, map[string]any{
				"name":        secretName,
				"namespace":   secretNamespace,
				"owner-refs":  obj.GetOwnerReferences(),
				"string-data": mresOutput,
			},
		)
		if err != nil {
			return check.Failed(err).Err(nil)
		}

		if _, err := r.yamlClient.ApplyYAML(ctx, b2); err != nil {
			return check.Failed(err)
		}

		mctx, cancel := r.newMongoContext(ctx)
		defer cancel()

		uri := msvcOutput.ClusterLocalURI
		if obj.IsGlobalVPNEnabled() {
			uri = msvcOutput.GlobalVpnURI
		}
		mongoCli, err := libMongo.NewClient(mctx, uri)
		if err != nil {
			return check.Failed(err)
		}

		if err := mongoCli.Ping(mctx); err != nil {
			return check.Failed(err)
		}

		defer mongoCli.Close()

		exists, err := mongoCli.UserExists(ctx, mresOutput.DbName, obj.Name)
		if err != nil {
			return check.Failed(err)
		}

		if !exists {
			if err := mongoCli.UpsertUser(ctx, mresOutput.DbName, mresOutput.Username, mresOutput.Password); err != nil {
				return check.Failed(err)
			}

			return check.StillRunning(nil)
		}

		if err := mongoCli.UpdateUserPassword(ctx, mresOutput.DbName, mresOutput.Username, mresOutput.Password); err != nil {
			return check.Failed(err)
		}
	}
	return check.Completed()
}

func (r *Reconciler) SetupWithManager(mgr ctrl.Manager, logger logging.Logger) error {
	r.Client = mgr.GetClient()
	r.Scheme = mgr.GetScheme()
	r.logger = logger.WithName(r.Name)
	r.yamlClient = kubectl.NewYAMLClientOrDie(mgr.GetConfig(), kubectl.YAMLClientOpts{Logger: r.logger})

	builder := ctrl.NewControllerManagedBy(mgr).For(&mongodbMsvcv1.Database{})
	builder.Owns(&corev1.Secret{})

	watchList := []client.Object{
		&mongodbMsvcv1.StandaloneService{},
		&mongodbMsvcv1.ClusterService{},
	}

	for _, obj := range watchList {
		builder.Watches(
			obj,
			handler.EnqueueRequestsFromMapFunc(
				func(ctx context.Context, obj client.Object) []reconcile.Request {
					msvcName, ok := obj.GetLabels()[constants.MsvcNameKey]
					if !ok {
						return nil
					}

					var dbList mongodbMsvcv1.DatabaseList
					if err := r.List(ctx, &dbList, &client.ListOptions{
						LabelSelector: labels.SelectorFromValidatedSet(
							map[string]string{constants.MsvcNameKey: msvcName},
						),
						Namespace: obj.GetNamespace(),
					},
					); err != nil {
						return nil
					}

					reqs := make([]reconcile.Request, 0, len(dbList.Items))
					for j := range dbList.Items {
						reqs = append(reqs, reconcile.Request{NamespacedName: fn.NN(dbList.Items[j].GetNamespace(), dbList.Items[j].GetName())})
					}

					return reqs
				},
			),
		)
	}

	builder.WithEventFilter(rApi.ReconcileFilter())
	builder.WithOptions(controller.Options{MaxConcurrentReconciles: r.Env.MaxConcurrentReconciles})
	return builder.Complete(r)
}
