package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"
	"os"
	"strings"
	"time"

	"github.com/kloudlite/api/pkg/errors"

	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/metadata"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/yaml"

	"github.com/kloudlite/api/apps/tenant-agent/internal/env"
	proto_rpc "github.com/kloudlite/api/apps/tenant-agent/internal/proto-rpc"
	t "github.com/kloudlite/api/apps/tenant-agent/types"
	"github.com/kloudlite/operator/common"
	"github.com/kloudlite/operator/grpc-interfaces/grpc/messages"
	libGrpc "github.com/kloudlite/operator/pkg/grpc"
	"github.com/kloudlite/operator/pkg/kubectl"
	"github.com/kloudlite/operator/pkg/logging"
)

type grpcHandler struct {
	inMemCounter   int64
	yamlClient     kubectl.YAMLClient
	logger         logging.Logger
	ev             *env.Env
	errorsCli      messages.MessageDispatchService_ReceiveErrorsClient
	msgDispatchCli messages.MessageDispatchServiceClient
}

func (g *grpcHandler) handleErrorOnApply(err error, msg t.AgentMessage) error {
	g.logger.Debugf("[ERROR]: %s", err.Error())

	b, err := json.Marshal(t.AgentErrMessage{
		AccountName: msg.AccountName,
		ClusterName: msg.ClusterName,
		Error:       err.Error(),
		Action:      msg.Action,
		Object:      msg.Object,
	})
	if err != nil {
		return errors.NewE(err)
	}

	return g.errorsCli.Send(&messages.ErrorData{Message: b})
}

func (g *grpcHandler) handleMessage(msg t.AgentMessage) error {
	g.inMemCounter++
	ctx, cf := context.WithTimeout(context.TODO(), 3*time.Second)
	defer cf()

	if msg.Object == nil {
		g.logger.Infof("msg.Object is nil, could not process anything out of this kafka message, ignoring ...")
		return nil
	}

	obj := unstructured.Unstructured{Object: msg.Object}
	mLogger := g.logger.WithKV("gvk", obj.GetObjectKind().GroupVersionKind().String()).WithKV("clusterName", msg.ClusterName).WithKV("accountName", msg.AccountName).WithKV("action", msg.Action)

	mLogger.Infof("[%d] received message", g.inMemCounter)

	if len(strings.TrimSpace(msg.AccountName)) == 0 {
		return g.handleErrorOnApply(errors.Newf("field 'accountName' must be defined in message"), msg)
	}

	switch msg.Action {
	case "apply", "delete":
		{
			b, err := yaml.Marshal(msg.Object)
			if err != nil {
				return g.handleErrorOnApply(err, msg)
			}

			if msg.Action == "apply" {
				_, err := g.yamlClient.ApplyYAML(ctx, b)
				if err != nil {
					mLogger.Infof("[%d] [error-on-apply]: %s", g.inMemCounter, err.Error())
					mLogger.Infof("[%d] failed to process message", g.inMemCounter)
					return g.handleErrorOnApply(err, msg)
				}
				mLogger.Infof("[%d] processed message", g.inMemCounter)
				return nil
			}

			if msg.Action == "delete" {
				err := g.yamlClient.DeleteYAML(ctx, b)
				if err != nil {
					mLogger.Infof("[%d] [error-on-delete]: %s", err.Error())
					return g.handleErrorOnApply(err, msg)
				}
				mLogger.Infof("[%d] processed message", g.inMemCounter)
				return nil
			}
			return nil
		}
	default:
		{
			err := errors.Newf("invalid action (%s)", msg.Action)
			mLogger.Infof("[%d] [error]: %s", err.Error())
			mLogger.Infof("[%d] failed to process message", g.inMemCounter)
			return g.handleErrorOnApply(err, msg)
		}
	}
}

func (g *grpcHandler) ensureAccessToken() error {
	ctx, cf := context.WithTimeout(context.TODO(), 50*time.Second)
	defer cf()
	if g.ev.AccessToken == "" {
		g.logger.Infof("waiting on clusterToken exchange for accessToken")
	}

	validationOut, err := g.msgDispatchCli.ValidateAccessToken(ctx, &messages.ValidateAccessTokenIn{
		AccountName: g.ev.AccountName,
		ClusterName: g.ev.ClusterName,
		AccessToken: g.ev.AccessToken,
	})

	if err != nil || validationOut == nil || !validationOut.Valid {
		g.logger.Infof("accessToken is invalid, requesting new accessToken ...")
	}

	if validationOut != nil && validationOut.Valid {
		g.logger.Infof("accessToken is valid, proceeding with it ...")
		return nil
	}

	out, err := g.msgDispatchCli.GetAccessToken(ctx, &messages.GetClusterTokenIn{
		AccountName:  g.ev.AccountName,
		ClusterName:  g.ev.ClusterName,
		ClusterToken: g.ev.ClusterToken,
	})
	if err != nil {
		return errors.NewE(err)
	}

	g.logger.Infof("valid access token has been obtained, persisting it in k8s secret (%s/%s)...", g.ev.AccessTokenSecretNamespace, g.ev.AccessTokenSecretName)

	s, err := g.yamlClient.Client().CoreV1().Secrets(g.ev.AccessTokenSecretNamespace).Get(context.TODO(), g.ev.AccessTokenSecretName, metav1.GetOptions{})
	if err != nil {
		return errors.NewE(err)
	}

	s.Data["ACCESS_TOKEN"] = []byte(out.AccessToken)
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
			g.logger.Errorf(err, "failed to delete pods for resource watcher")
		}
		g.logger.Infof("deleted all pods for resource watcher, they will be recreated")
	}

	return nil
}

func (g *grpcHandler) run() error {
	ctx, cf := context.WithCancel(context.TODO())
	defer cf()

	outgoingCtx := metadata.NewOutgoingContext(ctx, metadata.Pairs("authorization", g.ev.AccessToken))

	errorsCli, err := g.msgDispatchCli.ReceiveErrors(outgoingCtx)
	if err != nil {
		return errors.NewE(err)
	}

	g.errorsCli = errorsCli

	g.logger.Infof("asking message office to start sending actions")
	msgActionsCli, err := g.msgDispatchCli.SendActions(outgoingCtx, &messages.Empty{})
	if err != nil {
		return errors.NewE(err)
	}

	for {
		if err := ctx.Err(); err != nil {
			g.logger.Infof("connection context cancelled, will retry now...")
			return errors.NewE(err)
		}

		var msg t.AgentMessage
		a, err := msgActionsCli.Recv()
		if err != nil {
			g.logger.Errorf(err, "[ERROR] while receiving message")
			return errors.NewE(err)
		}

		if err := json.Unmarshal(a.Message, &msg); err != nil {
			g.logger.Errorf(err, "[ERROR] while json unmarshal")
			return errors.NewE(err)
		}

		if err := g.handleMessage(msg); err != nil {
			g.logger.Errorf(err, "[ERROR] while handling message")
			return errors.NewE(err)
		}
	}
}

func main() {
	var isDev bool
	flag.BoolVar(&isDev, "dev", false, "--dev")
	flag.Parse()

	ev := env.GetEnvOrDie()

	logger := logging.NewOrDie(&logging.Options{Name: "kloudlite-agent", Dev: isDev})

	logger.Infof("waiting for GRPC connection to happen")

	yamlClient := func() kubectl.YAMLClient {
		if isDev {
			logger.Debugf("connecting to k8s over host addr (%s)", "localhost:8081")
			return kubectl.NewYAMLClientOrDie(&rest.Config{Host: "localhost:8081"})
		}
		config, err := rest.InClusterConfig()
		if err != nil {
			panic(err)
		}
		return kubectl.NewYAMLClientOrDie(config)
	}()

	g := grpcHandler{
		inMemCounter: 0,
		yamlClient:   yamlClient,
		logger:       logger,
		ev:           ev,
	}

	vps := &vectorGrpcProxyServer{
		realVectorClient: nil,
		logger:           logger,
		accessToken:      ev.AccessToken,
		accountName:      ev.AccountName,
		clusterName:      ev.ClusterName,
	}

	gs := libGrpc.NewGrpcServer(libGrpc.GrpcServerOpts{Logger: logger})
	proto_rpc.RegisterVectorServer(gs.GrpcServer, vps)

	go func() {
		err := gs.Listen(ev.VectorProxyGrpcServerAddr)
		if err != nil {
			logger.Error(err)
			os.Exit(1)
		}
	}()

	common.PrintReadyBanner()

	for {
		logger.Infof("trying to connect to message office grpc (%s)", ev.GrpcAddr)
		cc, err := func() (*grpc.ClientConn, error) {
			if isDev {
				logger.Infof("attempting grpc connect over %s", ev.GrpcAddr)
				return libGrpc.Connect(ev.GrpcAddr, libGrpc.ConnectOpts{
					SecureConnect: false,
					Timeout:       20 * time.Second,
				})
			}
			logger.Infof("attempting grpc connect over %s", ev.GrpcAddr)
			return libGrpc.ConnectSecure(ev.GrpcAddr)
		}()
		if err != nil {
			log.Fatalf("Failed to connect after retries: %v", err)
		}

		logger.Infof("GRPC connection to message-office (%s) successful", ev.GrpcAddr)

		g.msgDispatchCli = messages.NewMessageDispatchServiceClient(cc)

		if err := g.ensureAccessToken(); err != nil {
			logger.Errorf(err, "ensuring access token")
		}

		vps.accessToken = g.ev.AccessToken
		vps.realVectorClient = proto_rpc.NewVectorClient(cc)

		if err := g.run(); err != nil {
			logger.Errorf(err, "running grpc sendActions")
		}

		connState := cc.GetState()
		for connState != connectivity.Ready && connState != connectivity.Shutdown {
			log.Printf("Connection lost, trying to reconnect...")
			break
		}
		cc.Close()
	}
}
