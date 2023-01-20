package utils

import (
	"bytes"
	"fmt"
	"os/exec"
)

func ExecCommand(cmd string, params []string) ([]byte, error) {
	cmder := exec.Command(cmd, params...)

	var stdout, stderr bytes.Buffer
	cmder.Stdout = &stdout
	cmder.Stderr = &stderr

	if err := cmder.Start(); err != nil {
		return nil, err
	}
	if err := cmder.Wait(); err != nil {
		return nil, err
	}

	if len(stderr.Bytes()) != 0 {
		return nil, fmt.Errorf(stderr.String())
	}
	return stdout.Bytes(), nil
}
