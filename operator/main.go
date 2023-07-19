package main

//
// import (
// 	"context"
// 	"encoding/json"
// 	"flag"
// 	"fmt"
// 	"io/ioutil"
// 	"net/http"
// 	"os"
//
// 	"k8s.io/apimachinery/pkg/labels"
// 	"k8s.io/client-go/rest"
// 	"sigs.k8s.io/controller-runtime/pkg/client"
// 	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
// 	"sigs.k8s.io/controller-runtime/pkg/manager"
//
// 	artifactsControllers "github.com/kloudlite/operator/controllers/artifacts"
// 	influxDbControllers "github.com/kloudlite/operator/controllers/influxdb.msvc"
// 	opensearchControllers "github.com/kloudlite/operator/controllers/opensearch.msvc"
// 	s3awsControllers "github.com/kloudlite/operator/controllers/s3.aws"
// 	serverlessControllers "github.com/kloudlite/operator/controllers/serverless"
// 	watchercontrollers "github.com/kloudlite/operator/controllers/watcher"
// 	""
// 	"github.com/kloudlite/operator/lib/constants"
// 	fn "github.com/kloudlite/operator/lib/functions"
// 	"github.com/kloudlite/operator/lib/harbor"
// 	rApi "github.com/kloudlite/operator/lib/operator"
//
// 	"github.com/kloudlite/operator/lib/logging"
// 	"github.com/kloudlite/operator/lib/redpanda"
//
// 	crdsv1 "github.com/kloudlite/operator/apis/crds/v1"
// 	"github.com/kloudlite/operator/controllers/crds"
//
// 	_ "k8s.io/client-go/plugin/pkg/client/auth"
//
// 	"k8s.io/apimachinery/pkg/runtime"
// 	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
// 	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
// 	ctrl "sigs.k8s.io/controller-runtime"
// 	"sigs.k8s.io/controller-runtime/pkg/healthz"
// 	"sigs.k8s.io/controller-runtime/pkg/log/zap"
//
// 	artifactsv1 "github.com/kloudlite/operator/apis/artifacts/v1"
// 	csiv1 "github.com/kloudlite/operator/apis/csi/v1"
// 	elasticsearchmsvcv1 "github.com/kloudlite/operator/apis/elasticsearch.msvc/v1"
// 	extensionsv1 "github.com/kloudlite/operator/apis/extensions/v1"
// 	influxdbmsvcv1 "github.com/kloudlite/operator/apis/influxdb.msvc/v1"
// 	mongodbexternalv1 "github.com/kloudlite/operator/apis/mongodb.external/v1"
// 	mysqlexternalv1 "github.com/kloudlite/operator/apis/mysql.external/v1"
// 	neo4jmsvcv1 "github.com/kloudlite/operator/apis/neo4j.msvc/v1"
// 	opensearchmsvcv1 "github.com/kloudlite/operator/apis/opensearch.msvc/v1"
// 	redpandamsvcv1 "github.com/kloudlite/operator/apis/redpanda.msvc/v1"
// 	s3awsv1 "github.com/kloudlite/operator/apis/s3.aws/v1"
// 	serverlessv1 "github.com/kloudlite/operator/apis/serverless/v1"
// 	mongodbexternalcontrollers "github.com/kloudlite/operator/controllers/mongodb.external"
// 	mysqlexternalcontrollers "github.com/kloudlite/operator/controllers/mysql.external"
// 	redpandamsvccontrollers "github.com/kloudlite/operator/controllers/redpanda.msvc"
// 	// +kubebuilder:scaffold:imports
// )
//
// var (
// 	scheme   = runtime.NewScheme()
// 	setupLog = ctrl.Log.WithName("operator")
// )
//
// func init() {
// 	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
// 	utilruntime.Must(crdsv1.AddToScheme(scheme))
// 	utilruntime.Must(influxdbmsvcv1.AddToScheme(scheme))
// 	utilruntime.Must(serverlessv1.AddToScheme(scheme))
// 	utilruntime.Must(elasticsearchmsvcv1.AddToScheme(scheme))
// 	utilruntime.Must(opensearchmsvcv1.AddToScheme(scheme))
// 	utilruntime.Must(s3awsv1.AddToScheme(scheme))
// 	utilruntime.Must(artifactsv1.AddToScheme(scheme))
// 	utilruntime.Must(redpandamsvcv1.AddToScheme(scheme))
// 	utilruntime.Must(mongodbexternalv1.AddToScheme(scheme))
// 	utilruntime.Must(mysqlexternalv1.AddToScheme(scheme))
// 	// utilruntime.Must(mongodbmsvcv1.AddToScheme(scheme))
// 	// utilruntime.Must(redismsvcv1.AddToScheme(scheme))
// 	// utilruntime.Must(mysqlmsvcv1.AddToScheme(scheme))
// 	// utilruntime.Must(zookeepermsvcv1.AddToScheme(scheme))
// 	utilruntime.Must(extensionsv1.AddToScheme(scheme))
// 	utilruntime.Must(neo4jmsvcv1.AddToScheme(scheme))
// 	utilruntime.Must(csiv1.AddToScheme(scheme))
// 	// +kubebuilder:scaffold:scheme
// }
//
// type arrayFlags []string
//
// func (i *arrayFlags) String() string {
// 	return "<nothing>"
// }
//
// func (i *arrayFlags) Set(value string) error {
// 	*i = append(*i, value)
// 	return nil
// }
//
// func main() {
// 	// time.AfterFunc(
// 	// 	time.Second*10, func() {
// 	// 		// MemProfileAllocs changes which type of memory to profile
// 	// 		// allocations.
// 	// 		defer profile.Start(profile.MemProfile).Stop()
// 	// 	},
// 	// )
//
// 	var metricsAddr string
// 	var enableLeaderElection bool
// 	var probeAddr string
// 	var isDev bool
// 	var devServerHost string
// 	var enableForArgs arrayFlags
// 	var skipControllerArgs arrayFlags
// 	var isAllEnabled bool
//
// 	// flag.StringVar(&metricsAddr, "metrics-bind-address", ":9091", "The address the metric endpoint binds to.")
// 	// flag.StringVar(&probeAddr, "health-probe-bind-address", ":9092", "The address the probe endpoint binds to.")
// 	flag.StringVar(&metricsAddr, "metrics-bind-address", ":12345", "The address the metric endpoint binds to.")
// 	flag.StringVar(&probeAddr, "health-probe-bind-address", ":12346", "The address the probe endpoint binds to.")
// 	flag.BoolVar(
// 		&enableLeaderElection, "leader-elect", false,
// 		"Enable leader election for controllers manager. "+
// 			"Enabling this will ensure there is only one active controllers manager.",
// 	)
//
// 	flag.BoolVar(&isDev, "dev", false, "--dev")
// 	opts := zap.Options{
// 		Development: true,
// 	}
// 	opts.BindFlags(flag.CommandLine)
//
// 	flag.StringVar(&devServerHost, "serverHost", "localhost:8080", "--serverHost <host:port>")
// 	flag.Var(&enableForArgs, "for", "--for item1 --for item2")
// 	flag.Var(&skipControllerArgs, "skip", "--skip item1 --skip item2")
// 	flag.BoolVar(&isAllEnabled, "all", false, "--for")
// 	flag.Parse()
//
// 	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))
//
// 	logger := logging.NewOrDie(&logging.Options{Dev: isDev})
//
// 	mgr, err := func() (manager.Manager, error) {
// 		cOpts := ctrl.Options{
// 			Scheme:                     scheme,
// 			MetricsBindAddress:         metricsAddr,
// 			Port:                       9443,
// 			HealthProbeBindAddress:     probeAddr,
// 			LeaderElection:             enableLeaderElection,
// 			LeaderElectionID:           "operator.kloudlite.io",
// 			LeaderElectionResourceLock: "configmaps",
// 		}
// 		if isDev {
// 			// cOpts.LeaderElectionID = "nxtcoder17.dev.kloudlite.io"
// 			return ctrl.NewManager(&rest.Config{Host: devServerHost}, cOpts)
// 		}
// 		return ctrl.NewManager(ctrl.GetConfigOrDie(), cOpts)
// 	}()
// 	if err != nil {
// 		setupLog.Error(err, "unable to start manager")
// 		os.Exit(1)
// 	}
//
// 	envVars := env.GetEnvOrDie()
//
// 	controllers := []rApi.Reconciler{
// 		&crds.ProjectReconciler{Name: "project"},
// 		&crds.AppReconciler{Name: "app"},
// 		&crds.RouterReconciler{Name: "router"},
// 		// &crds.AccountRouterReconciler{Name: "account-router"},
// 		&crds.ManagedServiceReconciler{Name: "msvc"},
// 		&crds.ManagedResourceReconciler{Name: "mres"},
//
// 		&redpandamsvccontrollers.ServiceReconciler{Name: "msvc-redpanda-service"},
// 		&redpandamsvccontrollers.TopicReconciler{Name: "msvc-redpanda-topic"},
//
// 		&serverlessControllers.LambdaReconciler{Name: "lambda"},
//
// 		// &elasticsearchControllers.ServiceReconciler{Name: "msvc-elasticsearch-service"},
// 		&opensearchControllers.ServiceReconciler{Name: "msvc-opensearch-service"},
//
// 		&influxDbControllers.ServiceReconciler{Name: "msvc-influxdb-service"},
// 		&influxDbControllers.BucketReconciler{Name: "msvc-influxdb-bucket"},
//
// 		&s3awsControllers.BucketReconciler{Name: "s3-aws-bucket"},
//
// 		&artifactsControllers.HarborProjectReconciler{Name: "artifacts-harbor-project"},
// 		&artifactsControllers.HarborUserAccountReconciler{Name: "artifacts-harbor-user-account"},
//
// 		&mongodbexternalcontrollers.DatabaseReconciler{Name: "external-mongodb-database"},
// 		&mysqlexternalcontrollers.DatabaseReconciler{Name: "external-mysql-database"},
// 	}
// 	// +kubebuilder:scaffold:builder
//
// 	producer, err := redpanda.NewProducer(envVars.KafkaBrokers)
// 	if err != nil {
// 		setupLog.Error(err, "creating redpanda producer")
// 		panic(err)
// 	}
// 	defer producer.Close()
//
// 	statusNotifier := watchercontrollers.NewNotifier(envVars.ClusterId, producer, envVars.KafkaStatusReplyTopic)
// 	billingNotifier := watchercontrollers.NewNotifier(envVars.ClusterId, producer, envVars.KafkaBillingReplyTopic)
//
// 	controllers = append(
// 		controllers,
// 		&watchercontrollers.StatusWatcherReconciler{Name: "status", Notifier: statusNotifier},
// 		&watchercontrollers.BillingWatcherReconciler{Name: "billing-watcher-watcher", Notifier: billingNotifier},
// 	)
//
// 	enabledForControllers := map[string]bool{}
// 	for _, arg := range enableForArgs {
// 		enabledForControllers[arg] = true
// 	}
//
// 	skippedControllers := map[string]bool{}
// 	for _, arg := range skipControllerArgs {
// 		skippedControllers[arg] = true
// 	}
//
// 	for _, rc := range controllers {
// 		if skippedControllers[rc.GetName()] {
// 			logger.Infof("skipping %s controllers (by flag) ", rc.GetName())
// 			continue
// 		}
// 		if isAllEnabled || enabledForControllers[rc.GetName()] {
// 			if err := rc.SetupWithManager(mgr, logger); err != nil {
// 				setupLog.Error(err, "unable to create controllers", "controllers", rc.GetName())
// 				os.Exit(1)
// 			}
// 		}
// 	}
//
// 	if err = mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
// 		setupLog.Error(err, "unable to set up health check")
// 		os.Exit(1)
// 	}
//
// 	if err = mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
// 		setupLog.Error(err, "unable to set up ready check")
// 		os.Exit(1)
// 	}
//
// 	go func() {
// 		kClient := mgr.GetClient()
// 		mux := http.NewServeMux()
// 		mux.HandleFunc(
// 			"/healthy", func(w http.ResponseWriter, _ *http.Request) {
// 				w.WriteHeader(http.StatusOK)
// 			},
// 		)
//
// 		mux.HandleFunc(
// 			"/image-push", func(w http.ResponseWriter, req *http.Request) {
// 				logger.Infof("webhook event received")
// 				body, err2 := ioutil.ReadAll(req.Body)
// 				if err2 != nil {
// 					return
// 				}
// 				var hookBody harbor.WebhookBody
// 				if err := json.Unmarshal(body, &hookBody); err != nil {
// 					return
// 				}
//
// 				imageName := func() string {
// 					for _, v := range hookBody.EventData.Resources {
// 						if v.ResourceUrl != "" {
// 							return v.ResourceUrl
// 						}
// 					}
// 					return ""
// 				}()
// 				if err := restartApp(kClient, imageName); err != nil {
// 					logger.Errorf(err, "restarting apps")
// 					return
// 				}
//
// 				if err := restartLambda(kClient, imageName); err != nil {
// 					logger.Errorf(err, "restarting lambda")
// 					return
// 				}
// 			},
// 		)
//
// 		if err := http.ListenAndServe(envVars.WebhookAddr, mux); err != nil {
// 			logger.Errorf(err, "failed to start webhook server")
// 		}
// 	}()
//
// 	setupLog.Infof("starting manager")
// 	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
// 		setupLog.Error(err, "problem running manager")
// 		panic(err)
// 	}
// }
//
// func restartApp(kClient client.Client, imageName string) error {
// 	var apps crdsv1.AppList
// 	if err := kClient.List(
// 		context.TODO(), &apps, &client.ListOptions{
// 			LabelSelector: labels.SelectorFromValidatedSet(
// 				map[string]string{
// 					fmt.Sprintf("kloudlite.io/image-%s", fn.Sha1Sum([]byte(imageName))): "true",
// 				},
// 			),
// 		},
// 	); err != nil {
// 		return err
// 	}
//
// 	for _, item := range apps.Items {
// 		if _, err := controllerutil.CreateOrUpdate(
// 			context.TODO(), kClient, &item, func() error {
// 				ann := item.GetAnnotations()
// 				ann[constants.AnnotationKeys.Restart] = "true"
// 				item.SetAnnotations(ann)
// 				return nil
// 			},
// 		); err != nil {
// 			return err
// 		}
// 	}
// 	return nil
// }
//
// func restartLambda(kClient client.Client, imageName string) error {
// 	var lambdaList serverlessv1.LambdaList
// 	if err := kClient.List(
// 		context.TODO(), &lambdaList, &client.ListOptions{
// 			LabelSelector: labels.SelectorFromValidatedSet(
// 				map[string]string{
// 					fmt.Sprintf("kloudlite.io/image-%s", fn.Sha1Sum([]byte(imageName))): "true",
// 				},
// 			),
// 		},
// 	); err != nil {
// 		return err
// 	}
//
// 	for _, item := range lambdaList.Items {
// 		if _, err := controllerutil.CreateOrUpdate(
// 			context.TODO(), kClient, &item, func() error {
// 				ann := item.GetAnnotations()
// 				ann[constants.AnnotationKeys.Restart] = "true"
// 				item.SetAnnotations(ann)
// 				return nil
// 			},
// 		); err != nil {
// 			return err
// 		}
// 	}
// 	return nil
// }
