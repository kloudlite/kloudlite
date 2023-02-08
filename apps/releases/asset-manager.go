package main

import (
	"bufio"
	"bytes"
	"embed"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"strings"

	"github.com/kloudlite/operator/pkg/templates"
	corev1 "k8s.io/api/core/v1"
)

//go:embed assets/*
var assets embed.FS

var AssetManager struct {
	Versions  []string
	CrdsBytes []byte
	Operators map[string][]string
}

func processFile(writer io.Writer, reader fs.File) error {
	defer reader.Close()
	bReader := bufio.NewReader(reader)
	msg := make([]byte, 0xffff)
	for {
		n, err := bReader.Read(msg)
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		if _, err := writer.Write(msg[:n]); err != nil {
			return err
		}
	}
	//if _, err := ctx.Write([]byte("\n---\n")); err != nil {
	//	return err
	//}
	return nil
}

func listVersions() []string {
	entries, err := assets.ReadDir("assets/operator")
	if err != nil {
		panic(err)
	}

	results := make([]string, 0, len(entries))
	for i := range entries {
		if entries[i].IsDir() {
			results = append(results, entries[i].Name())
		}
	}
	return results
}

func readAllCrds(isDev bool) []byte {
	writer := bytes.NewBuffer(nil)
	if isDev {
		files, err := os.ReadDir("../../config/crd/bases")
		if err != nil {
			panic(err)
		}
		for i := range files {
			reader, err := os.Open(path.Join("../../config/crd/bases", files[i].Name()))
			if err != nil {
				panic(err)
			}
			if err := processFile(writer, reader); err != nil {
				panic(err)
			}
		}
		return writer.Bytes()
	}

	entries, err := assets.ReadDir("assets/crds")
	if err != nil {
		panic(err)
	}
	for i := range entries {
		reader, err := assets.Open(path.Join("assets/crds", entries[i].Name()))
		if err != nil {
			panic(err)
		}
		if err := processFile(writer, reader); err != nil {
			panic(err)
		}
	}

	return writer.Bytes()
}

func listOperators(version string) []string {
	entries, err := assets.ReadDir("assets/operator/" + version)
	if err != nil {
		panic(err)
	}

	results := make([]string, 0, len(entries))

	for i := range entries {
		if !entries[i].IsDir() {
			results = append(results, strings.Replace(entries[i].Name(), ".yml", "", 1))
		}
	}

	return results
}

func readOperator(version, name string) ([]byte, error) {
	b, err := assets.ReadFile(fmt.Sprintf("assets/operator/%s/%s.yml", version, name))
	if err != nil {
		return nil, err
	}

	return templates.ParseBytes(b, map[string]any{
		"Namespace":       "kl-init-operators",
		"SvcAccountName":  "kloudlite-svc-account",
		"ImagePullPolicy": "Always",
		"ImageTag":        version,
		"EnvName":         "development",
		"NodeSelector":    map[string]string{},
		"Tolerations":     []corev1.Toleration{},
	})
}
