package utils

import (
	"archive/zip"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/otiai10/copy"
)

func ZipSource(source, target string) error {
	// 1. Create a ZIP file and zip.Writer
	f, err := os.Create(target)
	if err != nil {
		return err
	}
	defer f.Close()

	writer := zip.NewWriter(f)
	defer writer.Close()

	// 2. Go through all the files of the source
	return filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 3. Create a local file header
		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		// set compression
		header.Method = zip.Deflate

		// 4. Set relative path of a file as the header name
		header.Name, err = filepath.Rel(filepath.Dir(source), path)
		if err != nil {
			return err
		}
		if info.IsDir() {
			header.Name += "/"
		}

		// 5. Create writer for the file header and save content of the file
		headerWriter, err := writer.CreateHeader(header)
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()

		_, err = io.Copy(headerWriter, f)
		return err
	})
}

func Unzip(src string, destination string) ([]string, error) {

	var filenames []string
	r, err := zip.OpenReader(src)

	if err != nil {
		return filenames, err
	}
	defer r.Close()

	for _, f := range r.File {
		// Store "path/filename" for returning and using later on
		fpath := filepath.Join(destination, f.Name)

		// Checking for any invalid file paths
		if !strings.HasPrefix(fpath, filepath.Clean(destination)+string(os.PathSeparator)) {
			return filenames, fmt.Errorf("%s is an illegal filepath", fpath)
		}

		filenames = append(filenames, fpath)

		if f.FileInfo().IsDir() {

			os.MkdirAll(fpath, os.ModePerm)
			continue
		}

		if err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return filenames, err
		}

		outFile, err := os.OpenFile(fpath,
			os.O_WRONLY|os.O_CREATE|os.O_TRUNC,
			f.Mode())

		if err != nil {
			return filenames, err
		}

		rc, err := f.Open()
		if err != nil {
			return filenames, err
		}

		_, err = io.Copy(outFile, rc)

		outFile.Close()
		rc.Close()

		if err != nil {
			return filenames, err
		}
	}

	return filenames, nil
}

func ExtractZip(src, destination string) error {
	if _, err := os.Stat(destination); err == nil {
		if er := os.RemoveAll(destination); er != nil {
			return err
		}
	}

	if _, err := os.Stat(src); err != nil {
		if e := os.Mkdir(destination, os.ModePerm); e != nil {
			return e
		}
	} else {

		tempDirName, err := ioutil.TempDir("/tmp", "zip_")
		if err != nil {
			return err
		}
		defer os.RemoveAll(tempDirName)

		if names, err := Unzip(src, tempDirName); err != nil {
			return err
		} else {
			fmt.Println(names)
			if err := copy.Copy(path.Join(tempDirName, destination), destination); err != nil {
				return err
			}

		}

	}

	return nil
}

func mutateOperation() error {

	file, err := ioutil.TempFile("out", "prefix_")
	if err != nil {
		return err
	}

	return os.WriteFile(file.Name(), []byte("hi"), os.ModePerm)
}

func TestZip() error {
	zipName, dirName := "ram.zip", "out"
	if err := ExtractZip(zipName, dirName); err != nil {
		return err
	}

	if err := mutateOperation(); err != nil {
		return err
	}

	defer func() {
		if err := ZipSource(dirName, zipName); err != nil {
			log.Fatal(err)
		}
	}()
	return nil
}
