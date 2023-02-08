package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/pkg/profile"
)

func main() {
	var isDev bool
	flag.BoolVar(&isDev, "dev", false, "--dev")
	flag.Parse()

	p := profile.Start(profile.MemProfile)
	time.AfterFunc(10*time.Second, func() {
		defer p.Stop()
	})

	AssetManager.Versions = listVersions()
	AssetManager.CrdsBytes = readAllCrds(isDev)

	for _, v := range AssetManager.Versions {
		if AssetManager.Operators == nil {
			AssetManager.Operators = make(map[string][]string, len(AssetManager.Versions))
		}
		AssetManager.Operators[v] = listOperators(v)
	}

	app := fiber.New()
	// app.Use(pprof.New())

	authMiddleware := func(ctx *fiber.Ctx) error {
		klToken := ctx.GetReqHeaders()["Token"]
		fmt.Println("token", ctx.GetReqHeaders())
		if len(klToken) == 0 {
			return fiber.ErrUnauthorized
		}
		ctx.Set("token", klToken)
		return ctx.Next()
	}

	// app.Get("/versions", authMiddleware, func(ctx *fiber.Ctx) error {
	// 	entries, err := assets.ReadDir("assets/operator")
	// 	if err != nil {
	// 		return err
	// 	}
	//
	// 	results := make([]string, 0, len(entries))
	// 	for i := range entries {
	// 		if entries[i].IsDir() {
	// 			results = append(results, entries[i].Name())
	// 		}
	// 	}
	// 	return ctx.JSON(results)
	// })
	//
	// app.Get("/crds", func(ctx *fiber.Ctx) error {
	// 	entries, err := func() ([]fs.DirEntry, error) {
	// 		if isDev {
	// 			return os.ReadDir("../../config/crd/bases")
	// 		}
	// 		return assets.ReadDir("assets/crds")
	// 	}()
	// 	if err != nil {
	// 		return err
	// 	}
	//
	// 	results := make([]string, 0, len(entries))
	// 	for i := range entries {
	// 		if !entries[i].IsDir() {
	// 			results = append(results, entries[i].Name())
	// 		}
	// 	}
	// 	return ctx.JSON(results)
	// })
	//
	// app.Get("/crds/all", func(ctx *fiber.Ctx) error {
	// 	ctx.Set("Content-Type", "text/plain")
	// 	if isDev {
	// 		files, err := os.ReadDir("../../config/crd/bases")
	// 		if err != nil {
	// 			return err
	// 		}
	// 		for i := range files {
	// 			reader, err := os.Open(path.Join("../../config/crd/bases", files[i].Name()))
	// 			if err != nil {
	// 				return err
	// 			}
	// 			if err := processFile(ctx, reader); err != nil {
	// 				return err
	// 			}
	// 		}
	// 		return nil
	// 	}
	//
	// 	entries, err := assets.ReadDir("assets/crds")
	// 	if err != nil {
	// 		return err
	// 	}
	// 	for i := range entries {
	// 		reader, err := assets.Open(path.Join("assets/crds", entries[i].Name()))
	// 		if err != nil {
	// 			return err
	// 		}
	// 		if err := processFile(ctx, reader); err != nil {
	// 			return err
	// 		}
	// 	}
	// 	//ctx.Write([]byte("\n"))
	// 	return nil
	// })

	app.Get("/operators", authMiddleware, func(c *fiber.Ctx) error {
		return c.JSON(AssetManager.Operators)
	})

	app.Get("/download/:version/:name", authMiddleware, func(ctx *fiber.Ctx) error {
		version := ctx.Params("version")
		name := ctx.Params("name")

		b, err := readOperator(version, name)
		if err != nil {
			return err
		}
		if _, err := ctx.Write(b); err != nil {
			return err
		}

		return nil
	})

	log.Fatal(app.Listen(":3111"))
}
