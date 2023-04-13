package main

import (
	"context"
	"flag"
	"fmt"
	"time"

	"go.uber.org/fx"
	"kloudlite.io/pkg/logging"

	env "kloudlite.io/apps/message-office/internal/env"
	"kloudlite.io/apps/message-office/internal/framework"
	fn "kloudlite.io/pkg/functions"
)

func main() {
	var isDev bool
	flag.BoolVar(&isDev, "dev", false, "--dev")
	flag.Parse()

	app := fx.New(
		fx.NopLogger,

		fx.Provide(func() *env.Env {
			return env.LoadEnvOrDie()
		}),

		fx.Provide(
			func() (logging.Logger, error) {
				return logging.New(&logging.Options{Name: "message-office", Dev: isDev})
			},
		),
		fn.FxErrorHandler(),
		framework.Module,
	)

	ctx, cancelFn := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelFn()
	if err := app.Start(ctx); err != nil {
		panic(err)
	}

	fmt.Println(
		`
██████  ███████  █████  ██████  ██    ██ 
██   ██ ██      ██   ██ ██   ██  ██  ██  
██████  █████   ███████ ██   ██   ████   
██   ██ ██      ██   ██ ██   ██    ██    
██   ██ ███████ ██   ██ ██████     ██    
	`,
	)

	<-app.Done()
}

// func main() {
// 	// ev := env.LoadEnvOrDie()
// 	//
// 	// logger := logging.NewOrDie(&logging.Options{
// 	// 	Name: "message-office",
// 	// 	Dev:  true,
// 	// })
//
// 	// producer, err :=
// 	// if err != nil {
// 	// 	panic(err)
// 	// }
// 	// defer producer.Close()
//
// 	gServer := grpc.NewServer()
// 	grpcImpl := grpcServer{}
//
// 	messages.RegisterMessageDispatchServiceServer(gServer, grpcImpl)
// 	listener, err := net.Listen("tcp", ":50051")
// 	if err != nil {
// 		log.Fatalf("failed to listen: %v", err)
// 	}
// 	logger.Infof("[GRPC] server listening on addr: :%v", 50051)
// 	if err := gServer.Serve(listener); err != nil {
// 		log.Fatalf("failed to serve: %v", err)
// 	}
// }
