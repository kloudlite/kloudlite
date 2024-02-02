package zip

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"

	fn "github.com/kloudlite/kl/pkg/functions"
)

func Unzip(src, dest string) error {

	switch runtime.GOOS {
	case "darwin":
		_ = os.RemoveAll(path.Join(dest, "kloudlite.app"))
		_ = os.RemoveAll(path.Join(dest, "__MACOSX"))

		if err := fn.ExecCmd(fmt.Sprintf("unzip %q -d %q", src, dest), nil, false); err != nil {
			return err
		}

		_ = os.RemoveAll(path.Join(dest, "__MACOSX"))

	case "windows":
		_ = os.RemoveAll(path.Join(dest, "kloudlite"))

		if err := unzipForWindows(src, dest); err != nil {
			return err
		}

	default:
		return fmt.Errorf("not supported for platform %s", runtime.GOOS)
	}

	return nil
}

func unzipForWindows(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer func() {
		if err := r.Close(); err != nil {
			panic(err)
		}
	}()

	os.MkdirAll(dest, 0755)

	// Closure to address file descriptors issue with all the deferred .Close() methods
	extractAndWriteFile := func(f *zip.File) error {
		rc, err := f.Open()
		if err != nil {
			return err
		}
		defer func() {
			if err := rc.Close(); err != nil {
				panic(err)
			}
		}()

		path := filepath.Join(dest, f.Name)

		// Check for ZipSlip (Directory traversal)
		if !strings.HasPrefix(path, filepath.Clean(dest)+string(os.PathSeparator)) {
			return fmt.Errorf("illegal file path: %s", path)
		}

		if f.FileInfo().IsDir() {
			os.MkdirAll(path, f.Mode())
		} else {
			os.MkdirAll(filepath.Dir(path), f.Mode())
			f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return err
			}
			defer func() {
				if err := f.Close(); err != nil {
					panic(err)
				}
			}()

			_, err = io.Copy(f, rc)
			if err != nil {
				return err
			}
		}
		return nil
	}

	for _, f := range r.File {
		err := extractAndWriteFile(f)
		if err != nil {
			return err
		}
	}

	return nil
}
