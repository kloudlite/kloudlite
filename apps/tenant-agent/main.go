package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/pkg/errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/yaml"

	"github.com/kloudlite/api/apps/tenant-agent/internal/env"
	proto_rpc "github.com/kloudlite/api/apps/tenant-agent/internal/proto-rpc"
	t "github.com/kloudlite/api/apps/tenant-agent/types"
	"github.com/kloudlite/operator/grpc-interfaces/grpc/messages"

	libGrpc "github.com/kloudlite/api/pkg/grpc"
	"github.com/kloudlite/operator/pkg/kubectl"

	"github.com/kloudlite/api/pkg/logging"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
)

type grpcHandler struct {
	mu             sync.Mutex
	inMemCounter   int64
	yamlClient     kubectl.YAMLClient
	logger         *slog.Logger
	ev             *env.Env
	msgDispatchCli messages.MessageDispatchServiceClient
	isDev          bool
}

func (g *grpcHandler) incrementCounter() {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.inMemCounter++
}

const (
	MaxConnectionDuration = 45 * time.Second
)

func (g *grpcHandler) handleErrorOnApply(ctx context.Context, err error, msg t.AgentMessage) error {
	b, err := json.Marshal(t.AgentErrMessage{
		AccountName: msg.AccountName,
		Error:       err.Error(),
		Action:      msg.Action,
		Object:      msg.Object,
	})
	if err != nil {
		return errors.NewE(err)
	}

	_, err = g.msgDispatchCli.ReceiveError(ctx, &messages.ErrorData{
		ProtocolVersion: g.ev.GrpcMessageProtocolVersion,
		Message:         b,
	})
	return err
}

func NewAuthorizedGrpcContext(ctx context.Context, accessToken string) context.Context {
	return metadata.NewOutgoingContext(ctx, metadata.Pairs("authorization", accessToken))
}

func (g *grpcHandler) handleMessage(_ context.Context, msg t.AgentMessage) error {
	g.incrementCounter()
	start := time.Now()

	logger := g.logger.With("counter", g.inMemCounter, "account", msg.AccountName, "action", msg.Action)
	ctx, cf := func() (context.Context, context.CancelFunc) {
		if g.isDev {
			return context.WithCancel(context.TODO())
		}
		return context.WithTimeout(context.TODO(), 2*time.Second)
	}()
	defer cf()

	if msg.Object == nil {
		logger.Info("msg.Object is nil, could not process anything out of this message, ignoring ...")
		return nil
	}

	obj := unstructured.Unstructured{Object: msg.Object}
	mLogger := logger.With("gvk", obj.GetObjectKind().GroupVersionKind().String()).With("NN", fmt.Sprintf("%s/%s", obj.GetNamespace(), obj.GetName()))

	mLogger.Info("received message")

	if len(strings.TrimSpace(msg.AccountName)) == 0 {
		return g.handleErrorOnApply(ctx, errors.Newf("field 'accountName' must be defined in message"), msg)
	}

	switch msg.Action {
	case t.ActionApply:
		{
			b, err := yaml.Marshal(msg.Object)
			if err != nil {
				return g.handleErrorOnApply(ctx, err, msg)
			}

			if _, err := g.yamlClient.ApplyYAML(ctx, b); err != nil {
				mLogger.Error("failed to process message, got", "err", err, "error-on-apply:YAML", fmt.Sprintf("\n%s\n", b))
				return g.handleErrorOnApply(ctx, err, msg)
			}
		}
	case t.ActionDelete:
		{
			if err := g.yamlClient.DeleteResource(ctx, &obj); err != nil {
				mLogger.Warn("while deleting resource, got", "err", err)
				if apiErrors.IsNotFound(err) {
					mLogger.Info("processed message, resource does not exist, might already be deleted")
					return g.handleErrorOnApply(ctx, err, msg)
				}

				mLogger.Error("failed to process message, got", "err", err)
				return g.handleErrorOnApply(ctx, err, msg)
			}
		}
	case t.ActionRestart:
		{
			if err := g.yamlClient.RolloutRestart(ctx, kubectl.Deployment, obj.GetNamespace(), obj.GetLabels()); err != nil {
				return err
			}
			mLogger.Info("rolled out deployments")

			if err := g.yamlClient.RolloutRestart(ctx, kubectl.StatefulSet, obj.GetNamespace(), obj.GetLabels()); err != nil {
				return err
			}

			mLogger.Info("rolled out statefulsets")
		}
	default:
		{
			err := errors.Newf("invalid action (%s)", msg.Action)
			mLogger.Info("failed to process message, got", "err", err)
			return g.handleErrorOnApply(ctx, err, msg)
		}
	}

	mLogger.Info("processed message", "took", fmt.Sprintf("%.3fs", time.Since(start).Seconds()))

	return nil
}

func (g *grpcHandler) ensureAccessToken() error {
	if g.ev.AccessToken == "" {
		g.logger.Info("waiting on clusterToken exchange for accessToken")
	}

	ctx := NewAuthorizedGrpcContext(context.TODO(), g.ev.AccessToken)

	validationOut, err := g.msgDispatchCli.ValidateAccessToken(ctx, &messages.ValidateAccessTokenIn{
		ProtocolVersion: g.ev.GrpcMessageProtocolVersion,
	})
	if err != nil {
		g.logger.Error("validating access token, got", "err", err)
		validationOut = nil
	}

	if validationOut != nil && validationOut.Valid {
		g.logger.Info("accessToken is valid, proceeding with it ...")
		return nil
	}

	g.logger.Debug("accessToken is invalid, requesting new accessToken ...")

	out, err := g.msgDispatchCli.GetAccessToken(ctx, &messages.GetAccessTokenIn{
		ProtocolVersion: g.ev.GrpcMessageProtocolVersion,
		ClusterToken:    g.ev.ClusterToken,
	})
	if err != nil {
		return errors.NewE(err)
	}

	g.logger.Info("valid access token has been obtained, persisting it in k8s secret (%s/%s)...", g.ev.AccessTokenSecretNamespace, g.ev.AccessTokenSecretName)

	s, err := g.yamlClient.Client().CoreV1().Secrets(g.ev.AccessTokenSecretNamespace).Get(context.TODO(), g.ev.AccessTokenSecretName, metav1.GetOptions{})
	if err != nil {
		return errors.NewE(err)
	}

	if s.Data == nil {
		s.Data = make(map[string][]byte, 1)
	}
	s.Data["ACCESS_TOKEN"] = []byte(out.AccessToken)
	s.Data["ACCOUNT_NAME"] = []byte(out.AccountName)
	s.Data["CLUSTER_NAME"] = []byte(out.ClusterName)
	_, err = g.yamlClient.Client().CoreV1().Secrets(g.ev.AccessTokenSecretNamespace).Update(context.TODO(), s, metav1.UpdateOptions{})
	if err != nil {
		return errors.NewE(err)
	}

	g.ev.AccessToken = out.AccessToken

	if g.ev.ResourceWatcherNamespace != "" {
		// need to restart resource watcher
		d, err := g.yamlClient.Client().AppsV1().Deployments(g.ev.ResourceWatcherNamespace).Get(ctx, g.ev.ResourceWatcherName, metav1.GetOptions{})
		if err != nil {
			return errors.NewE(err)
		}
		podLabelSelector := metav1.LabelSelector{}
		for k, v := range d.Spec.Selector.MatchLabels {
			metav1.AddLabelToSelector(&podLabelSelector, k, v)
		}

		if err := g.yamlClient.Client().CoreV1().Pods(g.ev.ResourceWatcherNamespace).DeleteCollection(ctx, metav1.DeleteOptions{}, metav1.ListOptions{LabelSelector: metav1.FormatLabelSelector(&podLabelSelector)}); err != nil {
			g.logger.Error("failed to delete pods for resource watcher, got", "err", err)
		}
		g.logger.Info("deleted all pods for resource watcher, they will be recreated")
	}

	return nil
}

func (g *grpcHandler) run(rctx context.Context) error {
	ctx := NewAuthorizedGrpcContext(rctx, g.ev.AccessToken)

	g.logger.Info("asking message office to start sending actions")
	msgActionsCli, err := g.msgDispatchCli.SendActions(ctx, &messages.Empty{})
	if err != nil {
		return errors.NewE(err)
	}

	for {
		if err := ctx.Err(); err != nil {
			return err
		}

		var msg t.AgentMessage
		a, err := msgActionsCli.Recv()
		if err != nil {
			if status.Code(err) == codes.Unavailable {
				g.logger.Info("server unavailable, (may be, Gateway Timed Out 504), reconnecting ...")
				return nil
			}
			if status.Code(err) == codes.DeadlineExceeded {
				g.logger.Info("Connection Timed Out, reconnecting ...")
				return nil
			}
			if status.Code(err) == codes.Canceled {
				g.logger.Info("client is being closed, will reconnect")
				return nil
			}
			return err
		}

		if err := json.Unmarshal(a.Message, &msg); err != nil {
			g.logger.Error("while unmarshalling agent message, got", "err", err)
			return errors.NewE(err)
		}

		if err := g.handleMessage(ctx, msg); err != nil {
			g.logger.Error("while handling agent message, got", "err", err)
			return errors.NewE(err)
		}
	}
}

func (g *grpcHandler) askForGatewayResource(rctx context.Context) error {
	ctx := NewAuthorizedGrpcContext(rctx, g.ev.AccessToken)

	g.logger.Info("asking message office to send gateway resource for this cluster")
	out, err := g.msgDispatchCli.SendClusterGatewayResource(ctx, &messages.Empty{})
	if err != nil {
		return errors.NewE(err)
	}

	if _, err := g.yamlClient.ApplyYAML(ctx, out.Gateway); err != nil {
		g.logger.Error("failed to process message, got", "err", err, "error-on-apply:YAML", fmt.Sprintf("\n%s\n", out.Gateway))
		return err
	}

	return nil
}

func main() {
	var isDev bool
	flag.BoolVar(&isDev, "dev", false, "--dev")

	var debug bool
	flag.BoolVar(&debug, "debug", false, "--debug")

	var kubeApiAddr string
	flag.StringVar(&kubeApiAddr, "kube-api-addr", "localhost:8081", "--kube-api-addr [host]:port")

	flag.Parse()

	start := time.Now()
	common.PrintBuildInfo()

	ev := env.GetEnvOrDie()

	logger := logging.NewSlogLogger(logging.SlogOptions{ShowCaller: true, ShowDebugLogs: debug})

	logger.Debug("waiting for GRPC connection to happen")

	yamlClient := func() kubectl.YAMLClient {
		if isDev {
			logger.Debug("connecting to k8s over", "local-addr", kubeApiAddr)
			return kubectl.NewYAMLClientOrDie(&rest.Config{Host: kubeApiAddr}, kubectl.YAMLClientOpts{Slogger: logger})
		}
		config, err := rest.InClusterConfig()
		if err != nil {
			panic(err)
		}
		return kubectl.NewYAMLClientOrDie(config, kubectl.YAMLClientOpts{Slogger: logger})
	}()

	g := grpcHandler{
		mu:           sync.Mutex{},
		inMemCounter: 0,
		yamlClient:   yamlClient,
		logger:       logger,
		ev:           ev,
		isDev:        isDev,
	}

	vps := &vectorGrpcProxyServer{
		realVectorClient: nil,
		logger:           logger,
		accessToken:      ev.AccessToken,
	}

	gs, err := libGrpc.NewGrpcServer(libGrpc.ServerOpts{Logger: logger.With("component", "vector-grpc-proxy")})
	if err != nil {
		logger.Error("failed to create grpc server, got", "err", err)
	}
	proto_rpc.RegisterVectorServer(gs, vps)

	go func() {
		err := gs.Listen(ev.VectorProxyGrpcServerAddr)
		if err != nil {
			logger.Error("failed to listen on vector grpc server, got", "err", err)
			os.Exit(1)
		}
	}()

	common.PrintReadyBanner2(time.Since(start))

	for {
		cc, err := libGrpc.NewGrpcClientV2(ev.GrpcAddr, libGrpc.GrpcConnectOpts{TLSConnect: !isDev, Logger: logger})
		if err != nil {
			logger.Error("failed to connect to message office, got", "err", err, "retrying after", "5s")
			<-time.After(5 * time.Second)
		}

		g.msgDispatchCli = messages.NewMessageDispatchServiceClient(cc)

		if err := g.ensureAccessToken(); err != nil {
			logger.Error("ensuring access token, got", "err", err)
			logger.Info("will retry after 5s")
			cc.Close()
			<-time.After(5 * time.Second)
			continue
		}

		ctx, cf := context.WithTimeout(context.TODO(), MaxConnectionDuration)

		vps.accessToken = g.ev.AccessToken
		vps.realVectorClient = proto_rpc.NewVectorClient(cc)
		vps.connCancelFn = cf

		if err := g.askForGatewayResource(ctx); err != nil {
			logger.Error("asking gateway resource, got", "err", err)
			cf()
			cleanup(ctx, cc, logger)
			logger.Info("will retry after 5s")
			<-time.After(5 * time.Second)
		}

		go func() {
			defer cf()
			if err := g.run(ctx); err != nil {
				logger.Error("running grpc sendActions, got", "err", err)
			}
		}()

		cleanup(ctx, cc, logger)
	}
}

func cleanup(ctx context.Context, cc libGrpc.Client, logger *slog.Logger) {
	<-ctx.Done()
	logger.Debug("MAX_CONNECTION_DURATION reached, will re-initialize connection")

	if err := cc.Close(); err != nil {
		logger.Error("Failed to close connection, got", "err", err)
	}
}
