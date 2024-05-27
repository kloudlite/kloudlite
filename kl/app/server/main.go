package server

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	appconsts "github.com/kloudlite/kl/app-consts"
	"github.com/kloudlite/kl/constants"
	"github.com/kloudlite/kl/domain/client"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/fwd"
)

type Server struct {
	bin string
}

func New(binName string) *Server {
	return &Server{
		bin: binName,
	}
}
func portAvailable(port string) bool {
	address := fmt.Sprintf(":%s", port)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return false
	}
	defer listener.Close()
	return true
}

func (s *Server) Start() error {
	hdir, err := client.GetUserHomeDir()
	if err != nil {
		return err
	}

	keyPath := filepath.Join(hdir, ".ssh", "id_rsa")

	startCh, cancelCh, exitCh, lports, runner := fwd.GetController("kl", "localhost", keyPath)
	go runner()

	defer func() {
		exitCh <- struct{}{}
	}()

	app := http.NewServeMux()
	app.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		outputCh := make(chan string)
		errCh := make(chan error)

		command := strings.TrimPrefix(req.URL.Path, "/")

		switch command {
		case "healthy":
			w.WriteHeader(http.StatusOK)
			return

		case "list-proxy-ports":
			for _, sc := range lports {
				fmt.Fprintf(w, "ssh [%s] %s\n", sc.SshPort, sc.GetId())
			}

		case "remove-proxy-by-ssh":
			var chMsg fwd.StartCh
			err := json.NewDecoder(req.Body).Decode(&chMsg)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			for _, sc := range lports {
				if chMsg.SshPort == sc.SshPort {
					cancelCh <- sc
				}
			}

		case "add-proxy", "remove-proxy":
			var chMsg []fwd.StartCh
			err := json.NewDecoder(req.Body).Decode(&chMsg)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			switch command {
			case "add-proxy":
				for _, v := range chMsg {
					old, ok := lports[v.LocalPort]
					if ok {
						if old.GetId() != v.GetId() {
							cancelCh <- old

							time.Sleep(100 * time.Millisecond)
						} else {
							fn.Logf("port %s already forwarded to %s", v.LocalPort, old.GetId())
							continue
						}
					}

					if !portAvailable(v.LocalPort) {
						fn.Notify("error:", fmt.Sprintf("port %s already in use", v.LocalPort))
						fn.Logf("error: port %s already in use", v.LocalPort)
						return
					}

					startCh <- v
				}
			case "remove-proxy":
				for _, v := range chMsg {
					cancelCh <- v
				}
			}

		case "start", "stop", "status", "restart":

			if runtime.GOOS == constants.RuntimeWindows {
				http.Error(w, "this command not supported for windows", http.StatusBadRequest)
				return
			}

			go fn.StreamOutput(fmt.Sprintf("%s vpn %s", s.bin, command), nil, outputCh, errCh)

			for {
				select {
				case output := <-outputCh:
					w.Write([]byte(output))
					w.(http.Flusher).Flush()
				case err := <-errCh:
					if err != nil {
						http.Error(w, err.Error(), http.StatusInternalServerError)
					}
					return
				}
			}
		default:
			w.WriteHeader(http.StatusNotFound)
			return
		}
	})
	http.ListenAndServe(fmt.Sprintf(":%d", appconsts.AppPort), app)

	return nil
}
