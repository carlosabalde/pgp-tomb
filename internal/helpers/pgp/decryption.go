package pgp

import (
	"io"
	"os/exec"

	"github.com/pkg/errors"
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
		return errors.Wrap(err, "failed to create GPG stdin pipe")
	}

	if err := cmd.Start(); err != nil {
		return errors.Wrap(err, "failed to start GPG command execution")
	}

	if _, err := io.Copy(stdin, input); err == nil {
		stdin.Close()
	} else {
		return errors.Wrap(err, "failed to close GPG stdin pipe")
	}

	if err := cmd.Wait(); err != nil {
		return errors.Wrap(err, "failed to complete GPG command execution")
	}

	return nil
}
