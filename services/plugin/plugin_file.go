package plugin

import (
	"archive/zip"
	"fmt"
	"io"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"path"
	"os"
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
	readCloser     *zip.ReadCloser
	PluginManifest PluginManifest
}

func Open(name string) (*File, error) {
	ret := new(File)
	var err error
	// Open a zip archive for reading.
	ret.readCloser, err = zip.OpenReader(name)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	for _, f := range ret.readCloser.File {
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
			break
		}
	}
	return ret, nil
}

func (f *File) CheckArchitecture() error {
	return nil
}

func (f *File) CheckOS() error {
	return nil
}

func (f *File) CheckSum() error {
	return nil
}

func (f *File) CheckSysDeps() error {
	return nil
}

func (f *File) CheckDeps() error {
	return nil
}

func (f *File) ReleaseToDirectory(dir string) error {
	// Open a zip archive for reading.
	for _, f := range f.readCloser.File {
		if f.Name == "manifest.yml" {
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
	return f.readCloser.Close()
}

