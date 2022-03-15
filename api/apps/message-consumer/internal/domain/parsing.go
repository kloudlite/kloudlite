package domain

import (
	"bytes"
	"fmt"
	"os"
	"path"
	"text/template"

	"kloudlite.io/pkg/errors"
)

func readJobTemplate() *template.Template {
	tFile, ok := os.LookupEnv("JOB_TEMPLATE_FILE_PATH")
	errors.Assert(ok, fmt.Errorf("env key 'JOB_TEMPLATE_FILE_PATH' is not defined, exiting..."))
	t, err := template.New("job-template").ParseFiles(tFile)
	errors.AssertNoError(err, fmt.Errorf("Failed to parse template: %v", err))
	return t
}

func ParseJobTemplate(jobVars *JobVars) []byte {
	tFile := path.Join(os.Getenv("PWD"), "./src/templates/job-template.yml")
	fmt.Println(tFile)
	t, err := template.New("job-template").ParseFiles(tFile)
	if err != nil {
		fmt.Errorf("Failed to parse template: %v", err)
		panic(err)
	}

	w := new(bytes.Buffer)
	error := t.ExecuteTemplate(w, "job-template.yml", jobVars)
	if error != nil {
		panic(error)
	}

	return w.Bytes()
}
