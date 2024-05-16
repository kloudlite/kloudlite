package server

import (
	"fmt"
	"net/http"
	"strings"

	appconsts "github.com/kloudlite/kl/app-consts"
	fn "github.com/kloudlite/kl/pkg/functions"
)

type Server struct {
	bin string
}

func New(binName string) *Server {
	return &Server{
		bin: binName,
	}
}

func (s *Server) Start() {
	app := http.NewServeMux()
	app.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		outputCh := make(chan string)
		errCh := make(chan error)

		command := strings.TrimPrefix(req.URL.Path, "/")
		// fn.Log("command: " + command)

		if command == "healthy" {
			w.WriteHeader(http.StatusOK)
			return
		}

		go fn.StreamOutput(fmt.Sprintf("%s vpn %s", s.bin, command), nil, outputCh, errCh)

		for {
			select {
			case output := <-outputCh:
				// c.WriteString("data: " + output + "\n\n")
				w.Write([]byte(output))
				w.(http.Flusher).Flush()
			case err := <-errCh:
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
				}
				return
			}
		}

		// b, err := fn.ExecCmdOut("kl vpn start", nil)
		// if err != nil {
		// 	return err
		// }
		//
		// return c.Send(b)
	})
	http.ListenAndServe(fmt.Sprintf(":%d", appconsts.AppPort), app)
}
