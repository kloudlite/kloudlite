package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"

	devinfo "github.com/kloudlite/operator/apps/coredns/dev-info"
	"github.com/kloudlite/operator/common"
)

var debug bool

func runCoreDNS(ctx context.Context, corefile string) {
	log.Println("starting coredns")
	defer func() {
		log.Println("stopping coredns")
	}()

	if debug {
		f, err := os.Open(corefile)
		if err != nil {
			panic(err)
		}
		b, err := io.ReadAll(f)
		if err != nil {
			panic(err)
		}
		fmt.Printf("[DEBUG]: Corefile\n%s\n", b)
	}

	cmd := exec.CommandContext(ctx, "/coredns", "-conf", corefile)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout

	if err := cmd.Run(); err != nil {
		log.Printf("err occurred while running coredns: %v\n", err)
	}
}

func main() {
	var addr string
	var tlsAddr string
	var corefile string
	var devi string
	flag.StringVar(&addr, "addr", "", "--addr host:port")
	flag.StringVar(&tlsAddr, "tls-addr", "", "--tls-addr host:port")
	flag.StringVar(&corefile, "corefile", "", "--corefile <file-path>")
	flag.StringVar(&devi, "dev-info", "", "--dev-info <device-info>")
	flag.BoolVar(&debug, "debug", false, "--debug")
	flag.Parse()

	newCorefilePath := "./Corefile"
	f, err := os.Open(corefile)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	b, err := io.ReadAll(f)
	if err != nil {
		panic(err)
	}

	if err := os.WriteFile(newCorefilePath, b, fs.FileMode(os.O_WRONLY)); err != nil {
		panic(err)
	}

	ctx, cf := context.WithCancel(context.TODO())
	defer cf()

	go runCoreDNS(ctx, newCorefilePath)

	sm := http.NewServeMux()
	sm.HandleFunc("/healthy", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})

	sm.HandleFunc("/whoami", func(w http.ResponseWriter, r *http.Request) {
		if devi == "" {
			w.WriteHeader(400)
		}

		d := devinfo.DeviceInfo{}

		if err := d.FromBase64(devi); err != nil {
			return
		}

		w.Write([]byte(d.String()))
	})

	sm.HandleFunc("/resync", func(w http.ResponseWriter, r *http.Request) {
		log.Println("resyncing corefile")
		b, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer r.Body.Close()

		if err := os.WriteFile(newCorefilePath, b, fs.FileMode(os.O_WRONLY)); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		cf()
		ctx, cf = context.WithCancel(context.TODO())
		go runCoreDNS(ctx, newCorefilePath)

		w.WriteHeader(200)
	})

	common.PrintReadyBanner()

	log.Printf("use /healthy to check, if server is reachable")

	go func() {
		tlsPath, ok := os.LookupEnv("TLS_CERT_FILE_PATH")
		if !ok {
			log.Println("TLS_CERT_FILE_PATH is not set, ignoring https server")
			return
		}
		if tlsAddr == "" {
			log.Println("TLS_ADDR is not set, ignoring https server")
			return
		}

		log.Printf("https server is starting on %s", tlsAddr)
		if err := http.ListenAndServeTLS(tlsAddr, path.Join(tlsPath, "tls.crt"), path.Join(tlsPath, "tls.key"), sm); err != nil {
			log.Printf("err occurred while starting https server: %v\n", err)
		}
	}()

	log.Printf("http server is starting on %s", addr)

	if err := http.ListenAndServe(addr, sm); err != nil {
		log.Printf("err occurred while starting http server: %v\n", err)
	}
}
