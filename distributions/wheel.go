package distributions

import (
	"bytes"
	"fmt"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/rstudio/python-distribution-parser/archiver"
)

var wheelFileRe = regexp.MustCompile(`^(?P<namever>(?P<name>.+?)(-(?P<ver>\d.+?))?)(?:(-(?P<build>\d.*?))?-(?P<pyver>.+?)-(?P<abi>.+?)-(?P<plat>.+?)\.whl|\.dist-info)$`)

type Wheel struct {
	BaseDistribution
	Filename     string `json:"file_name"`
	BaseFilename string `json:"base_filename"`
}

func NewWheel(filename string) (Distribution, error) {
	wheel := &Wheel{}
	wheel.Filename = filename
	wheel.BaseFilename = filepath.Base(filename)
	err := wheel.ExtractMetadata()
	if err != nil {
		return nil, err
	}
	return wheel, nil
}

func (whl *Wheel) MetadataMap() map[string][]string {
	result := StructToMap(*whl)
	return result

	// metadataMap := make(map[string][]string, 0)
	// v := reflect.ValueOf(whl).Elem()
	// for i := 0; i < v.NumField(); i++ {
	// 	field := v.Field(i)
	// 	fieldName := v.Type().Field(i).Name
	// 	fmt.Println(fieldName)
	// 	fieldValue := field.Interface()
	// 	switch val := fieldValue.(type) {
	// 	case string:
	// 		metadataMap[fieldName] = []string{val}
	// 	case []string:
	// 		metadataMap[fieldName] = val
	// 	interface{}:
	// 		fmt.Printf("type: %T\n", val)
	// 		metadataMap[fieldName] = []string{fmt.Sprintf("%v", val)}
	// }
	// return metadataMap
}

func (whl *Wheel) ExtractMetadata() error {
	data, err := whl.read()
	if err != nil {
		return err
	}
	err = whl.Parse(data)
	if err != nil {
		return err
	}
	return nil
}

func (whl *Wheel) read() ([]byte, error) {
	filename := whl.Filename
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
		if strings.Contains(name, "METADATA") {
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
	return nil, fmt.Errorf("no METADATA in archive: %s", fqn)
}

func (whl *Wheel) GetPythonVersion() string {
	wheelInfo := wheelFileRe.FindStringSubmatch(whl.BaseFilename)
	if wheelInfo == nil {
		return "any"
	}

	for i, name := range wheelFileRe.SubexpNames() {
		if name == "pyver" {
			return wheelInfo[i]
		}
	}

	return "any"
}
