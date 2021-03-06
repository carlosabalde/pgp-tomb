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

func (self *Writer) Write(p []byte) (int, error) {
	return self.gzip.Write(p)
}

func (self *Writer) Close() error {
	if err := self.gzip.Close(); err != nil {
		self.file.Close()
		return errors.Wrap(err, "failed to close gzip writer")
	}

	if err := self.file.Close(); err != nil {
		return errors.Wrap(err, "failed to close file writer")
	}

	return nil
}

func (self *Secret) NewWriter() (*Writer, error) {
	if err := os.MkdirAll(filepath.Dir(self.path), os.ModePerm); err != nil {
		return nil, errors.Wrap(err, "failed to create path to secret")
	}

	fileWriter, err := os.Create(self.path)
	if err != nil {
		return nil, errors.Wrap(err, "failed to open file")
	}

	gzipWriter := gzip.NewWriter(fileWriter)

	gzipWriter.Comment = fmt.Sprintf("Generated by PGP Tomb %s", config.GetVersion())

	gzipWriter.Extra, err = self.serializeTags()
	if err != nil {
		gzipWriter.Close()
		fileWriter.Close()
		return nil, errors.Wrap(err, "failed to serialize tags")
	}

	return &Writer{
		file: fileWriter,
		gzip: gzipWriter,
	}, nil
}

func (self *Secret) serializeTags() ([]byte, error) {
	tagsMap := make(map[string]string)
	for _, tag := range self.GetTags() {
		tagsMap[tag.Name] = tag.Value
	}
	return json.Marshal(tagsMap)
}
