package secret

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/pkg/errors"

	"github.com/carlosabalde/pgp-tomb/internal/core/config"
)

type Writer struct {
	file *os.File
	gzip *gzip.Writer
}

func (writer *Writer) Write(p []byte) (int, error) {
	return writer.gzip.Write(p)
}

func (writer *Writer) Close() error {
	if err := writer.gzip.Close(); err != nil {
		writer.file.Close()
		return errors.Wrap(err, "failed to close gzip writer")
	}

	if err := writer.file.Close(); err != nil {
		return errors.Wrap(err, "failed to close file writer")
	}

	return nil
}

func (secret *Secret) NewWriter() (*Writer, error) {
	if err := os.MkdirAll(filepath.Dir(secret.path), os.ModePerm); err != nil {
		return nil, errors.Wrap(err, "failed to create path to secret")
	}

	fileWriter, err := os.Create(secret.path)
	if err != nil {
		return nil, errors.Wrap(err, "failed to open file")
	}

	gzipWriter := gzip.NewWriter(fileWriter)

	gzipWriter.Comment = fmt.Sprintf("Generated by PGP Tomb %s", config.GetVersion())

	gzipWriter.Extra, err = json.Marshal(secret.headers)
	if err != nil {
		gzipWriter.Close()
		fileWriter.Close()
		return nil, errors.Wrap(err, "failed to serialize headers")
	}

	return &Writer{
		file: fileWriter,
		gzip: gzipWriter,
	}, nil
}
