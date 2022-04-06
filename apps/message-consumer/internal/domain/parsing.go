package domain

import (
	"fmt"
	"kloudlite.io/pkg/errors"
	"os"
	"path"
	"text/template"
)

func readJobTemplate() *template.Template {
	tFile, ok := os.LookupEnv("JOB_TEMPLATE_FILE_PATH")
	filePath := path.Join(os.Getenv("PWD"), tFile)
	if path.IsAbs(tFile) {
		filePath = tFile
	}
	errors.Assert(ok, fmt.Errorf("env key 'JOB_TEMPLATE_FILE_PATH' is not defined, exiting..."))
	fmt.Println("tFile:", filePath)
	t, err := template.New("job-template").ParseFiles(filePath)
	errors.AssertNoError(err, fmt.Errorf("Failed to parse template: %v", err))
	return t
}
