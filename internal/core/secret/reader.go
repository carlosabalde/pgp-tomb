package secret

import (
	"compress/gzip"
	"encoding/json"
	"os"
	"sort"

	"github.com/pkg/errors"
)

type Reader struct {
	file *os.File
	gzip *gzip.Reader
}

func (self *Reader) Read(p []byte) (int, error) {
	return self.gzip.Read(p)
}

func (self *Reader) Close() error {
	if err := self.gzip.Close(); err != nil {
		self.file.Close()
		return errors.Wrap(err, "failed to close gzip reader")
	}

	if err := self.file.Close(); err != nil {
		return errors.Wrap(err, "failed to close file reader")
	}

	return nil
}

func (self *Secret) NewReader() (*Reader, error) {
	fileReader, err := os.Open(self.path)
	if err != nil {
		return nil, errors.Wrap(err, "failed to open file")
	}

	gzipReader, err := gzip.NewReader(fileReader)
	if err != nil {
		fileReader.Close()
		return nil, errors.Wrap(err, "failed to gunzip secret")
	}

	if err := json.Unmarshal(gzipReader.Extra, &self.tags); err != nil {
		gzipReader.Close()
		fileReader.Close()
		return nil, errors.Wrap(err, "failed to unserialize tags")
	}

	sort.Slice(self.tags, func(i, j int) bool {
		return self.tags[i].Name < self.tags[j].Name
	})

	return &Reader{
		file: fileReader,
		gzip: gzipReader,
	}, nil
}
