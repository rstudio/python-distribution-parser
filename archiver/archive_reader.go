package archiver

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

// archiveReader is an interface to abstract the behavior of different archive types.
type archiveReader interface {
	FileNames() ([]string, error)
	ReadFile(name string) ([]byte, error)
	Close() error
}

type zipReader struct {
	*zip.ReadCloser
}

func (z *zipReader) FileNames() ([]string, error) {
	var names []string
	for _, f := range z.File {
		names = append(names, f.Name)
	}
	return names, nil
}

func (z *zipReader) ReadFile(name string) ([]byte, error) {
	for _, f := range z.File {
		if f.Name == name {
			rc, err := f.Open()
			if err != nil {
				return nil, err
			}
			defer rc.Close()
			return io.ReadAll(rc)
		}
	}
	return nil, fmt.Errorf("file not found: %s", name)
}

type tarReader struct {
	filename string
	*tar.Reader
	closer io.Closer
}

func (t *tarReader) resetReader() error {
	t.Close()

	// Reopen the file
	f, err := os.Open(t.filename)
	if err != nil {
		return err
	}

	if strings.HasSuffix(t.filename, ".tar.gz") || strings.HasSuffix(t.filename, ".tgz") {
		gzr, err := gzip.NewReader(f)
		if err != nil {
			f.Close()
			return err
		}
		t.Reader = tar.NewReader(gzr) // Reset the tar reader with new gzip reader
	}

	return nil
}

func (t *tarReader) FileNames() ([]string, error) {
	defer t.resetReader()
	var names []string
	for {
		hdr, err := t.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		names = append(names, hdr.Name)
	}
	return names, nil
}

func (t *tarReader) ReadFile(name string) ([]byte, error) {
	for {
		hdr, err := t.Next()
		if err == io.EOF {
			return nil, fmt.Errorf("file not found: %s", name)
		}
		if err != nil {
			return nil, err
		}
		if hdr.Name == name {
			return io.ReadAll(t)
		}
	}
}

func (t *tarReader) Close() error {
	return t.closer.Close()
}

func NewArchiveReader(fqn string) (archiveReader, error) {
	_, err := os.Stat(fqn)
	if errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("no such file: %s", fqn)
	} else if err != nil {
		return nil, err
	}

	if r, err := zip.OpenReader(fqn); err == nil {
		return &zipReader{r}, nil
	}

	f, err := os.Open(fqn)
	if err != nil {
		return nil, err
	}

	// check if file is a gzipped tarball
	if strings.HasSuffix(fqn, ".tar.gz") || strings.HasSuffix(fqn, ".tgz") {
		gzr, err := gzip.NewReader(f)
		if err != nil {
			f.Close()
			return nil, err
		}
		r := tar.NewReader(gzr)
		err = tarReadCheck(r)
		if err != nil {
			return nil, err
		}
		return &tarReader{fqn, r, f}, nil
	}

	f.Close()
	return nil, fmt.Errorf("not a known archive format: %s", fqn)
}

// tarReadCheck will attempt to read a header from the tar.Reader to
// determine if it is a valid tar file. It will return nil if it can read
// a header successfully; otherwise, it will return an error.
func tarReadCheck(r *tar.Reader) error {
	_, err := r.Next()
	return err
}
