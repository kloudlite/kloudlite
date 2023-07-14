package utils

import (
	"fmt"
	"os"
	"path"

	"github.com/containerd/continuity/fs"

	awss3 "kloudlite.io/pkg/aws-s3"
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

func MakeTfWorkFileReady(nodeId, tfPath string, awss3client awss3.AwsS3, createIfNotExists bool) error {
	filename := fmt.Sprintf("%s.zip", nodeId)
	// check if file exists in db
	err := awss3client.IsFileExists(filename)
	if err != nil {
		if !createIfNotExists {
			return fmt.Errorf("no state file found with the nodeId %s to operate", nodeId)
		}

		if err := CreateNodeWorkDir(nodeId); err != nil {
			return err
		}

		if err := fs.CopyDir(path.Join(Workdir, nodeId), tfPath); err != nil {
			return err
		}

		return nil
	}

	// found file in db, download and extract to the workdir
	fmt.Println("-> tfstate found in s3, downloading and extracting it")
	source := path.Join(Workdir, filename)
	// Download from db
	if err := awss3client.DownloadFile(source, filename); err != nil {
		return err
	}

	if _, err := Unzip(source, path.Join(Workdir)); err != nil {
		return err
	}

	return nil
}

func SaveToDb(nodeId string, awss3client awss3.AwsS3) error {

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

		if err := awss3client.UploadFile(source, filename); err != nil {
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
