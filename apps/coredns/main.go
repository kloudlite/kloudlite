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

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
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

	app := fiber.New()

	app.Get("/healthy", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	app.Post("/resync", func(c *fiber.Ctx) error {
		log.Println("resyncing corefile")

		b := c.Body()

		if err := os.WriteFile(newCorefilePath, b, fs.FileMode(os.O_WRONLY)); err != nil {
			return err
		}

		cf()
		ctx, cf = context.WithCancel(context.TODO())
		go runCoreDNS(ctx, newCorefilePath)

		return c.SendString("ok")
	})

	// Apply CORS middleware
	app.Use(cors.New(cors.Config{
		AllowOrigins: "https://*.kloudlite.io",
		AllowMethods: "GET",
		AllowHeaders: "Origin, X-Requested-With, Content-Type, Accept, Authorization",
	}))

	app.Get("/whoami", func(c *fiber.Ctx) error {

		if devi == "" {
			return fmt.Errorf("device info is not set")
		}

		d := devinfo.DeviceInfo{}

		if err := d.FromBase64(devi); err != nil {
			return err
		}

		return c.JSON(d)
	})

	app.Get("/*", func(c *fiber.Ctx) error {
		return c.SendStatus(http.StatusNotFound)
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
		if err := app.ListenTLS(tlsAddr, path.Join(tlsPath, "tls.crt"), path.Join(tlsPath, "tls.key")); err != nil {
			log.Printf("err occurred while starting https server: %v\n", err)
		}
	}()

	log.Printf("http server is starting on %s", addr)

	if err := app.Listen(addr); err != nil {
		log.Printf("err occurred while starting http server: %v\n", err)
	}
}
