package pgp

import (
	"io"
	"os/exec"
)

func DecryptWithGPG(gpg string, input io.Reader, output io.Writer) error {
	args := []string{
		"--use-agent",
		"-d",
	}

	cmd := exec.Command(gpg, args...)

	cmd.Stdout = output

	cmd.Stderr = nil

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	if _, err := io.Copy(stdin, input); err == nil {
		stdin.Close()
	}

	if err := cmd.Wait(); err != nil {
		return err
	}

	return nil
}
