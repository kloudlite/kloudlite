package utils

import (
	"context"
	"encoding/base64"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/containerd/continuity/fs"
	"gopkg.in/yaml.v2"
	mongogridfs "kloudlite.io/pkg/mongo-gridfs"
)

const (
	Workdir string = "/tmp/tf-workdir"
)

func Base64YamlDecode(in string, out interface{}) error {
	rawDecodedText, err := base64.StdEncoding.DecodeString(in)
	if err != nil {
		return err
	}

	fmt.Println(string(rawDecodedText))

	return yaml.Unmarshal(rawDecodedText, out)
}

func SaveToDb(ctx context.Context, nodeId string, gfs mongogridfs.GridFs) error {
	/*
		Steps:
		  - compress the workdir into zip
		  - check if file present. if yes, upsert file else upload file
	*/

	dir := path.Join(Workdir, nodeId)
	filename := fmt.Sprintf("%s.zip", nodeId)

	// compress the workdir and upsert to db
	if err := func() error {
		if _, err := os.Stat(dir); err != nil {
			return err
		}

		source := fmt.Sprintf("%s.zip", dir)

		// compress
		if err := ZipSource(dir, source); err != nil {
			return err
		}

		if err := gfs.Upsert(ctx, filename, source); err != nil {
			return err
		}

		return nil
	}(); err != nil {
		fmt.Println(ColorText(fmt.Sprint("Error: ", err), 1))
		return err
	}

	return nil
}

const (
	enableClear bool = false
)

func CreateNodeWorkDir(nodeId string) error {
	dir := path.Join(Workdir, nodeId)
	if _, err := os.Stat(dir); err != nil {
		return os.Mkdir(dir, os.ModePerm)
	}

	if enableClear {
		if err := os.RemoveAll(dir); err != nil {
			return err
		}

		return os.Mkdir(dir, os.ModePerm)
	} else {
		return nil
	}
}

func SetupGetWorkDir() error {
	if _, err := os.Stat(Workdir); err != nil {
		return os.Mkdir(Workdir, os.ModePerm)
	}
	return nil
}

func MakeTfWorkFileReady(ctx context.Context, nodeId, tfPath string, gfs mongogridfs.GridFs, createIfNotExists bool) error {

	filename := fmt.Sprintf("%s.zip", nodeId)
	// check if file exists in db
	gf, err := gfs.FetchFileRef(ctx, filename)
	if err != nil {
		return err
	}

	// not found create new dir
	if gf == nil {
		if !createIfNotExists {
			return fmt.Errorf("no state file found with the nodeId %s to operate", nodeId)
		}

		if err := CreateNodeWorkDir(nodeId); err != nil {
			return err
		}

		// a.tfTemplates
		if err := fs.CopyDir(path.Join(Workdir, nodeId), tfPath); err != nil {
			return err
		}

		return nil
	}

	// found file in db, download and extract to the workdir
	fmt.Println(gf.Name, "found, extract it by downloading")

	source := path.Join(Workdir, filename)
	// Download from db
	if err := gfs.Download(ctx, filename, source); err != nil {
		return err
	}

	if s, err := Unzip(source, path.Join(Workdir)); err != nil {
		return err
	} else {
		for _, v := range s {
			fmt.Print(v, " \n")
		}
	}

	return nil
}

func ColorText(text interface{}, code int) string {
	return fmt.Sprintf("\033[38;05;%dm%v\033[0m", code, text)
}

func DownloadDir() error {
	return nil
}

func UploadDir() error {
	return nil
}

func ExecCmd(cmdString string, logStr string) error {
	r := csv.NewReader(strings.NewReader(cmdString))
	r.Comma = ' '
	cmdArr, err := r.Read()
	if err != nil {
		return err
	}

	if logStr != "" {
		fmt.Printf("[#] %s\n", logStr)
	} else {
		fmt.Printf("[#] %s\n", strings.Join(cmdArr, " "))
	}

	cmd := exec.Command(cmdArr[0], cmdArr[1:]...)
	cmd.Stderr = os.Stderr
	// cmd.Stdout = os.Stdout

	if err := cmd.Run(); err != nil {
		fmt.Printf("err occurred: %s\n", err.Error())
		return err
	}
	return nil
}

const (
	CLUSTER_ID = "kl"
)

// rmTFdir implements doProviderClient
// func rmdir(folder string) error {
// 	return execCmd(fmt.Sprintf("rm -rf %q", folder), "")
// }

// makeTFdir implements doProviderClient
func Mkdir(folder string) error {
	return ExecCmd(fmt.Sprintf("mkdir -p %q", folder), "mkdir <terraform_dir>")
}

func getOutput(folder, key string) ([]byte, error) {
	vars := []string{"output", "-json"}
	fmt.Printf("[#] terraform %s\n", strings.Join(vars, " "))
	cmd := exec.Command("terraform", vars...)
	cmd.Dir = folder

	// cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	out, err := cmd.Output()
	if err != nil {
		return nil, err

	}

	// fmt.Println(string(out))

	var resp map[string]struct {
		Value string `json:"value"`
	}

	err = json.Unmarshal(out, &resp)
	if err != nil {
		return nil, err
	}

	return []byte(resp[key].Value), nil
}

func InitTFdir(dir string) error {
	cmd := exec.Command("terraform", "init")
	cmd.Dir = dir

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// applyTF implements doProviderClient
func ApplyTF(folder string, values map[string]string) error {

	vars := []string{"apply", "-auto-approve"}

	fmt.Printf("[#] terraform %s", strings.Join(vars, " "))

	for k, v := range values {
		vars = append(vars, fmt.Sprintf("-var=%s=%s", k, v))
	}

	cmd := exec.Command("terraform", vars...)
	cmd.Dir = folder

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmd.Dir = folder

	return cmd.Run()
}

// destroyNode implements doProviderClient
func DestroyNode(nodeId string, values map[string]string) error {
	dest := path.Join(Workdir, nodeId)
	vars := []string{"destroy", "-auto-approve"}

	for k, v := range values {
		vars = append(vars, fmt.Sprintf("-var=%s=%s", k, v))
	}

	cmd := exec.Command("terraform", vars...)
	cmd.Dir = dest

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

func GetOutput(folder, key string) ([]byte, error) {
	vars := []string{"output", "-json"}
	fmt.Printf("[#] terraform %s\n", strings.Join(vars, " "))
	cmd := exec.Command("terraform", vars...)
	cmd.Dir = folder

	// cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	out, err := cmd.Output()
	if err != nil {
		return nil, err

	}

	// fmt.Println(string(out))

	var resp map[string]struct {
		Value string `json:"value"`
	}

	err = json.Unmarshal(out, &resp)
	if err != nil {
		return nil, err
	}

	return []byte(resp[key].Value), nil
}
