package main

import (
	"context"
	"encoding/json"
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
	marshal, err := json.Marshal(in.Inputs)
	if err != nil {
		return nil, err
	}
	f := `
` + in.Init + `
` + in.FunName + `(` + string(marshal) + `)`
	ctx := v8.NewContext()
	val, err := ctx.RunScript(f, "eval.js")
	if err != nil {
		return nil, err
	}
	m := make(map[string]*anypb.Any)
	marshalJSON, err := val.MarshalJSON()
	json.Unmarshal(marshalJSON, &m)
	return &jseval.EvalOut{Output: m}, nil
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}
	listen, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
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
