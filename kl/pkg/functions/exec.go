package functions

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// StreamOutput executes a command and streams its output line by line.
func StreamOutput(cmdString string, env map[string]string, outputCh chan<- string, errCh chan<- error) {
	defer close(outputCh)
	defer close(errCh)

	r := csv.NewReader(strings.NewReader(cmdString))
	r.Comma = ' '
	cmdArr, err := r.Read()
	if err != nil {
		errCh <- err
		return
	}

	cmd := exec.Command(cmdArr[0], cmdArr[1:]...)

	if env != nil {
		for k, v := range env {
			cmd.Env = append(cmd.Env, k+"="+v)
		}
	}

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		errCh <- err
		return
	}

	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		errCh <- err
		return
	}

	if err := cmd.Start(); err != nil {
		errCh <- err
		return
	}

	var stdoutBuf, stderrBuf bytes.Buffer

	go func() {
		buf := make([]byte, 1024)
		for {
			n, err := stdoutPipe.Read(buf)
			if err != nil {
				break
			}
			stdoutBuf.Write(buf[:n])
			outputCh <- stdoutBuf.String()
			stdoutBuf.Reset()
		}
	}()

	go func() {
		buf := make([]byte, 1024)
		for {
			n, err := stderrPipe.Read(buf)
			if err != nil {
				break
			}
			stderrBuf.Write(buf[:n])
			outputCh <- stderrBuf.String()
			stderrBuf.Reset()
		}
	}()

	if err := cmd.Wait(); err != nil {
		errCh <- err
	}
}

func ExecCmd(cmdString string, env map[string]string, verbose bool) error {
	r := csv.NewReader(strings.NewReader(cmdString))
	r.Comma = ' '
	cmdArr, err := r.Read()
	if err != nil {
		return err
	}
	cmd := exec.Command(cmdArr[0], cmdArr[1:]...)
	if verbose {
		Log("[#] " + strings.Join(cmdArr, " "))
		cmd.Stdout = os.Stdout
	}

	if env == nil {
		env = map[string]string{}
	}

	for k, v := range env {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
	}

	cmd.Stderr = os.Stderr
	// s.Start()
	err = cmd.Run()
	// s.Stop()
	return err
}

// func ExecCmdOut(cmdString string, env map[string]string) ([]byte, error) {
// 	r := csv.NewReader(strings.NewReader(cmdString))
// 	r.Comma = ' '
// 	cmdArr, err := r.Read()
// 	if err != nil {
// 		return nil, err
// 	}
// 	cmd := exec.Command(cmdArr[0], cmdArr[1:]...)
//
// 	if env == nil {
// 		env = map[string]string{}
// 	}
//
// 	for k, v := range env {
// 		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
// 	}
//
// 	cmd.Stderr = os.Stderr
// 	out, err := cmd.Output()
//
// 	fmt.Println(string(out), err)
//
// 	return out, err
// }

func ExecCmdOut(cmdString string, env map[string]string) ([]byte, error) {
	r := csv.NewReader(strings.NewReader(cmdString))
	r.Comma = ' '
	cmdArr, err := r.Read()
	if err != nil {
		return nil, err
	}

	cmd := exec.Command(cmdArr[0], cmdArr[1:]...)

	if env == nil {
		env = make(map[string]string)
	}

	for k, v := range env {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()
	if err != nil {
		return nil, err
	}

	combinedOutput := append(stdout.Bytes(), stderr.Bytes()...)
	return combinedOutput, nil
}

func Exec(cmdString string, env map[string]string) ([]byte, error) {
	r := csv.NewReader(strings.NewReader(cmdString))
	r.Comma = ' '
	cmdArr, err := r.Read()
	if err != nil {
		return nil, err
	}
	cmd := exec.Command(cmdArr[0], cmdArr[1:]...)

	if env == nil {
		env = map[string]string{}
	}

	for k, v := range env {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
	}

	cmd.Stderr = os.Stderr
	out, err := cmd.Output()

	return out, err
}
