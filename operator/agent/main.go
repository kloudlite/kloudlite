package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/yaml"

	"github.com/kloudlite/operator/agent/internal/env"
	t "github.com/kloudlite/operator/agent/types"
	"github.com/kloudlite/operator/grpc-interfaces/grpc/messages"
	libGrpc "github.com/kloudlite/operator/pkg/grpc"
	"github.com/kloudlite/operator/pkg/kubectl"
	"github.com/kloudlite/operator/pkg/logging"
)

type grpcHandler struct {
	inMemCounter   int64
	yamlClient     *kubectl.YAMLClient
	logger         logging.Logger
	ev             *env.Env
	errorsCli      messages.MessageDispatchService_ReceiveErrorsClient
	msgDispatchCli messages.MessageDispatchServiceClient
}

func (g *grpcHandler) handleErrorOnApply(ctx context.Context, err error, msg t.AgentMessage) error {
	g.logger.Debugf("[ERROR]: %s", err.Error())

	b, err := json.Marshal(t.AgentErrMessage{
		AccountName: msg.AccountName,
		ClusterName: msg.ClusterName,
		Error:       err.Error(),
		Action:      msg.Action,
		Object:      msg.Object,
	})

	if err != nil {
		return err
	}

	return g.errorsCli.Send(&messages.ErrorData{
		AccessToken: g.ev.AccessToken,
		ClusterName: g.ev.ClusterName,
		AccountName: g.ev.AccountName,
		Data:        b,
	})
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
	// defer func() {
	// 	mLogger.Infof("processed message [%d]", g.inMemCounter)
	// }()

	if len(strings.TrimSpace(msg.AccountName)) == 0 {
		return g.handleErrorOnApply(ctx, fmt.Errorf("field 'accountName' must be defined in message"), msg)
	}

	switch msg.Action {
	case "apply", "delete":
		{
			b, err := yaml.Marshal(msg.Object)
			if err != nil {
				return g.handleErrorOnApply(ctx, err, msg)
			}

			if msg.Action == "apply" {
				_, err := g.yamlClient.ApplyYAML(ctx, b)
				if err != nil {
					mLogger.Infof("[%d] [error-on-apply]: %s", g.inMemCounter, err.Error())
					mLogger.Infof("[%d] failed to process message", g.inMemCounter)
					return g.handleErrorOnApply(ctx, err, msg)
				}
				mLogger.Infof("[%d] processed message", g.inMemCounter)
				return nil
			}

			if msg.Action == "delete" {
				err := g.yamlClient.DeleteYAML(ctx, b)
				if err != nil {
					mLogger.Infof("[%d] [error-on-delete]: %s", err.Error())
					return g.handleErrorOnApply(ctx, err, msg)
				}
				mLogger.Infof("[%d] processed message", g.inMemCounter)
				return nil
			}
			return nil
		}
	default:
		{
			err := fmt.Errorf("invalid action (%s)", msg.Action)
			mLogger.Infof("[%d] [error]: %s", err.Error())
			mLogger.Infof("[%d] failed to process message", g.inMemCounter)
			return g.handleErrorOnApply(ctx, err, msg)
		}
	}
}

func (g *grpcHandler) ensureAccessToken() error {
	if g.ev.AccessToken == "" {
		g.logger.Infof("waiting on clusterToken exchange for accessToken")
		out, err := g.msgDispatchCli.GetAccessToken(context.TODO(), &messages.GetClusterTokenIn{
			ClusterToken: g.ev.ClusterToken,
		})
		if err != nil {
			panic(err)
		}

		s, err := g.yamlClient.K8sClient.CoreV1().Secrets(g.ev.AccessTokenSecretNamespace).Get(context.TODO(), g.ev.AccessTokenSecretName, metav1.GetOptions{})
		if err != nil {
			panic(err)
		}
		delete(s.Data, "CLUSTER_TOKEN")
		s.Data["ACCESS_TOKEN"] = []byte(out.AccessToken)
		_, err = g.yamlClient.K8sClient.CoreV1().Secrets(g.ev.AccessTokenSecretNamespace).Update(context.TODO(), s, metav1.UpdateOptions{})
		if err != nil {
			panic(err)
		}

		g.ev.AccessToken = out.AccessToken
	}
	return nil
}

func (g *grpcHandler) run(conn *grpc.ClientConn) error {
	g.ensureAccessToken()

	ctx, cf := context.WithCancel(context.TODO())
	defer cf()

	errorsCli, err := g.msgDispatchCli.ReceiveErrors(ctx)
	if err != nil {
		return err
	}

	g.errorsCli = errorsCli

	msgActionsCli, err := g.msgDispatchCli.SendActions(ctx, &messages.StreamActionsRequest{
		AccessToken: g.ev.AccessToken,
		ClusterName: g.ev.ClusterName,
		AccountName: g.ev.AccountName,
	})
	if err != nil {
		return err
	}

	for {
		if err := ctx.Err(); err != nil {
			g.logger.Infof("connection context cancelled, will retry now...")
			return err
		}

		var msg t.AgentMessage
		a, err := msgActionsCli.Recv()
		if err != nil {
			g.logger.Errorf(err, "[ERROR] while receiving message")
			return err
		}

		if err := json.Unmarshal(a.Data, &msg); err != nil {
			g.logger.Errorf(err, "[ERROR] while json unmarshal")
			return err
		}

		g.handleMessage(msg)
	}
}

func main() {
	var isDev bool
	flag.BoolVar(&isDev, "dev", false, "--dev")
	flag.Parse()

	ev := env.GetEnvOrDie()

	fmt.Println(
		`
  ███████  ███████  █████  ██████  ██    ██ 
  ██   ██  ██      ██   ██ ██   ██  ██  ██  
  ██████   █████   ███████ ██   ██   ████   
  ██   ██  ██      ██   ██ ██   ██    ██    
  ██   ██  ███████ ██   ██ ██████     ██      `,
	)

	logger := logging.NewOrDie(&logging.Options{Name: "kloudlite-agent", Dev: isDev})

	logger.Infof("waiting for GRPC connection to happen")

	yamlClient := func() *kubectl.YAMLClient {
		if isDev {
			return kubectl.NewYAMLClientOrDie(&rest.Config{Host: "localhost:8080"})
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

	for {
		cc, err := func() (*grpc.ClientConn, error) {
			if isDev {
				return libGrpc.Connect(ev.GrpcAddr)
			}
			return libGrpc.ConnectSecure(ev.GrpcAddr)
		}()
		if err != nil {
			log.Fatalf("Failed to connect after retries: %v", err)
		}

		logger.Infof("GRPC connection successful")

		g.msgDispatchCli = messages.NewMessageDispatchServiceClient(cc)
		if err := g.run(cc); err != nil {
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
