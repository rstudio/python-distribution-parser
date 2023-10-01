package distributions

import (
	"bytes"
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/rstudio/python-distribution-parser/archiver"
)

type SDist struct {
	BaseDistribution
	Filename string `json:"filename"`
}

func NewSDist(filename string) (Distribution, error) {
	sdist := &SDist{}
	sdist.Filename = filename
	err := sdist.ExtractMetadata()
	if err != nil {
		return nil, err
	}
	return sdist, nil
}

func (sd *SDist) ExtractMetadata() error {
	data, err := sd.read()
	if err != nil {
		return err
	}
	err = sd.Parse(data)
	if err != nil {
		return err
	}
	return nil
}

func (sd *SDist) read() ([]byte, error) {
	filename := sd.Filename
	fqn, err := filepath.Abs(filepath.Clean(filename))
	if err != nil {
		return nil, fmt.Errorf("error normalizing path: %w", err)
	}

	archiveReader, err := archiver.NewArchiveReader(fqn)
	if err != nil {
		return nil, fmt.Errorf("error getting archive: %w", err)
	}
	defer archiveReader.Close() // Ensure the archive is closed after reading

	fileNames, err := archiveReader.FileNames()
	if err != nil {
		return nil, err
	}

	var tuples [][]string
	for _, name := range fileNames {
		if strings.Contains(name, "PKG-INFO") {
			tuples = append(tuples, strings.Split(name, "/"))
		}
	}

	sort.Slice(tuples, func(i, j int) bool {
		return len(tuples[i]) < len(tuples[j])
	})

	for _, path := range tuples {
		candidate := strings.Join(path, "/")
		data, err := archiveReader.ReadFile(candidate)
		if err != nil {
			return nil, fmt.Errorf("error reading file %s from archive: %v", candidate, err)
		}
		if bytes.Contains(data, []byte("Metadata-Version")) {
			return data, nil
		}
	}
	return nil, fmt.Errorf("no PKG-INFO in archive: %s", fqn)
}
