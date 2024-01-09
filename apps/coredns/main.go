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
	var corefile string
	flag.StringVar(&addr, "addr", "", "--addr host:port")
	flag.StringVar(&corefile, "corefile", "", "--corefile <file-path>")
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

	log.Printf("http server is starting on %s", addr)

	http.ListenAndServe(addr, sm)
}
