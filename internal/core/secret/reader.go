package secret

import (
	"compress/gzip"
	"encoding/json"
	"os"

	"github.com/pkg/errors"
)

type Reader struct {
	file *os.File
	gzip *gzip.Reader
}

func (reader *Reader) Read(p []byte) (int, error) {
	return reader.gzip.Read(p)
}

func (reader *Reader) Close() error {
	if err := reader.gzip.Close(); err != nil {
		reader.file.Close()
		return errors.Wrap(err, "failed to close gzip reader")
	}

	if err := reader.file.Close(); err != nil {
		return errors.Wrap(err, "failed to close file reader")
	}

	return nil
}

func (secret *Secret) NewReader() (*Reader, error) {
	fileReader, err := os.Open(secret.path)
	if err != nil {
		return nil, errors.Wrap(err, "failed to open file")
	}

	gzipReader, err := gzip.NewReader(fileReader)
	if err != nil {
		fileReader.Close()
		return nil, errors.Wrap(err, "failed to gunzip secret")
	}

	if err := json.Unmarshal(gzipReader.Extra, &secret.headers); err != nil {
		gzipReader.Close()
		fileReader.Close()
		return nil, errors.Wrap(err, "failed to unserialize headers")
	}

	return &Reader{
		file: fileReader,
		gzip: gzipReader,
	}, nil
}
