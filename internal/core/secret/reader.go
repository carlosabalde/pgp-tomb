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

	if err := self.unserializeTags(gzipReader.Extra); err != nil {
		gzipReader.Close()
		fileReader.Close()
		return nil, errors.Wrap(err, "failed to unserialize tags")
	}

	return &Reader{
		file: fileReader,
		gzip: gzipReader,
	}, nil
}

func (self *Secret) unserializeTags(data []byte) error {
	tagsMap := make(map[string]string)
	if err := json.Unmarshal(data, &tagsMap); err != nil {
		return err
	}

	var tags []Tag
	for name, value := range tagsMap {
		tags = append(tags, Tag{
			Name:  name,
			Value: value,
		})
	}
	self.SetTags(tags)

	return nil
}
