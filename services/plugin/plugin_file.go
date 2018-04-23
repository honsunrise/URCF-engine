package plugin

import (
	"archive/zip"
	"errors"
	"fmt"
	"gopkg.in/yaml.v2"
	"io"
	"io/ioutil"
	"os"
	"path"
)

var (
	ErrCannotFindManifestFile = errors.New("can't find manifest.yml")
)

type PositionError struct {
	off int64
	msg string
	val interface{}
}

func (e *PositionError) Error() string {
	msg := e.msg
	if e.val != nil {
		msg += fmt.Sprintf(" '%v' ", e.val)
	}
	msg += fmt.Sprintf("in record at byte %#x", e.off)
	return msg
}

type File struct {
	io.Closer
	close          func() error
	reader         *zip.Reader
	PluginManifest PluginManifest
}

func OpenReader(readCloser io.ReaderAt, size int64) (*File, error) {
	ret := new(File)
	var err error
	// Open a zip archive for reading.
	ret.reader, err = zip.NewReader(readCloser, size)
	if err != nil {
		return nil, err
	}

	found := false

	for _, f := range ret.reader.File {
		if f.Name == "manifest.yml" {
			rc, err := f.Open()
			if err != nil {
				return nil, err
			}
			buf, err := ioutil.ReadAll(rc)
			if err != nil {
				return nil, err
			}
			err = yaml.Unmarshal(buf, &ret.PluginManifest)
			if err != nil {
				return nil, err
			}
			found = true
			break
		}
	}
	if !found {
		return nil, ErrCannotFindManifestFile
	}
	return ret, nil
}

func (f *File) ReleaseToDirectory(dir string) error {
	err := os.MkdirAll(dir, 0770)
	if err != nil {
		return err
	}
	// Open a zip archive for reading.
	for _, f := range f.reader.File {
		if f.FileInfo().IsDir() {
			err = os.MkdirAll(path.Join(dir, f.Name), 0770)
			if err != nil {
				return err
			}
		} else {
			rc, err := f.Open()
			if err != nil {
				return err
			}
			releasePath := path.Join(dir, f.Name)
			file, err := os.Create(releasePath)
			if err != nil {
				return err
			}
			_, err = io.Copy(file, rc)
			if err != nil {
				return err
			}
			err = file.Sync()
			if err != nil {
				return err
			}
			err = file.Close()
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (f *File) Close() error {
	return f.close()
}
