package pgp

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"
	"strings"

	"github.com/pkg/errors"
	"golang.org/x/crypto/openpgp"
)

func Decrypt(agent string, input io.Reader, output io.Writer, key *PrivateKey) error {
	promptError := false
	prompt := func(keys []openpgp.Key, symmetric bool) ([]byte, error) {
		if symmetric {
			return getPassphrase(agent, promptError)
		}

		for _, k := range keys {
			if k.PrivateKey.Encrypted {
				passphrase, err := getPassphrase(agent, promptError)
				if err != nil {
					return nil, errors.Wrap(err, "failed to get passphrase for private key")
				}
				if passphrase == nil {
					return nil, errors.New("no passphrase for private key")
				}
				if err := k.PrivateKey.Decrypt(passphrase); err != nil {
					promptError = true
				}
			}
		}

		return nil, nil
	}

	message, err := openpgp.ReadMessage(input, openpgp.EntityList{key.Entity}, prompt, nil)
	if err != nil {
		return errors.Wrap(err, "failed to read message")
	}

	if _, err := io.Copy(output, message.UnverifiedBody); err != nil {
		return errors.Wrap(err, "failed to copy decrypted message")
	}

	return nil
}

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

func getPassphrase(agent string, error bool) ([]byte, error) {
	// See:
	//   - https://www.gnupg.org/documentation/manuals/gnupg/Agent-GET_005fPASSPHRASE.html

	cacheId := "pgp-tomb:private-key-passphrase"
	errorMessage := "X"

	if error {
		if err := exec.Command(
			agent,
			fmt.Sprintf("CLEAR_PASSPHRASE %s", cacheId),
			"/bye").Run(); err != nil {
			return nil, errors.Wrap(err, "failed to clear GPG Connect Agent cache")
		}
		errorMessage = "Failed+to+decrypt+private+key!"
	}

	cmd := exec.Command(
		agent,
		fmt.Sprintf(
			"GET_PASSPHRASE --data %s %s Passphrase PGP+Tomb+Private+Key",
			cacheId, errorMessage),
		"/bye")

	cmd.Stderr = nil

	cmd.Stdin = nil

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, errors.Wrap(err, "failed to create GPG Connect Agent stdout pipe")
	}

	if err := cmd.Start(); err != nil {
		return nil, errors.Wrap(err, "failed to start GPG Connect Agent command execution")
	}

	passphrase := ""
	cancelled := false
	reader := bufio.NewReader(stdout)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			cancelled = true
			break
		}
		if strings.HasPrefix(line, "D ") {
			passphrase = strings.TrimSuffix(strings.TrimPrefix(line, "D "), "\n")
		} else if strings.HasPrefix(line, "OK") {
			break
		}
	}
	if err := cmd.Wait(); err != nil {
		return nil, errors.Wrap(err, "failed to complete GPG Connect Agent command execution")
	}

	if !cancelled {
		return []byte(passphrase), nil
	} else {
		return nil, nil
	}
}
