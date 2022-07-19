package main

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/anypb"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/jseval"
	"net"
	"os"
	v8 "rogchap.com/v8go"
	"sync"
)

type JsServer struct {
	jseval.UnimplementedJSEvalServer
}

func (s *JsServer) Eval(c context.Context, in *jseval.EvalIn) (*jseval.EvalOut, error) {
	f := in.FunName + `(` + string(in.Inputs.Value) + `)`
	ctx := v8.NewContext()
	ctx.RunScript(in.Init, "eval.js")
	val, err := ctx.RunScript(f, "eval.js")
	if err != nil {
		return nil, err
	}
	marshalJSON, err := val.MarshalJSON()
	return &jseval.EvalOut{Output: &anypb.Any{
		TypeUrl: "",
		Value:   marshalJSON,
	}}, nil
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}
	listen, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		panic(err)
	} else {
		server := grpc.NewServer()
		jseval.RegisterJSEvalServer(server, &JsServer{})
		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			err := server.Serve(listen)
			wg.Done()
			if err != nil {
				panic(err)
			}
		}()
		wg.Wait()
	}
}
