package distributions

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
)

var DistExtensions = map[string]string{
	".whl":     "bdist_wheel",
	".tar.bz2": "sdist",
	".tar.gz":  "sdist",
	".zip":     "sdist",

	// TODO maybe eventually add support for other distribution extensions
	// ".exe":     "bdist_wininst",
	// ".egg":     "bdist_egg",
}

func getFullExtension(filename string) string {
	parts := strings.Split(filename, ".")
	if len(parts) < 2 {
		return ""
	}

	// Check for tar.gz, tar.bz2, etc.
	if parts[len(parts)-1] == "gz" || parts[len(parts)-1] == "bz2" {
		if len(parts) < 3 {
			return "." + parts[len(parts)-1]
		}
		return "." + parts[len(parts)-2] + "." + parts[len(parts)-1]
	}

	return "." + parts[len(parts)-1]
}

func NewDistributionMetadata(filename string) (Distribution, string, string, error) {
	var metadata Distribution
	var fileType string

	for ext, dt := range DistExtensions {
		if getFullExtension(filename) == ext {
			fileType = dt
			break
		}
	}

	if fileType == "" {
		return nil, "", "", errors.New("unknown distribution format: " + filepath.Base(filename))
	}

	var err error
	switch fileType {
	case "bdist_wheel":
		metadata, err = NewWheel(filename)
	case "sdist":
		metadata, err = NewSDist(filename)
	default:
		return nil, "", "", errors.New("invalid distribution type: " + fileType)
		// TODO maybe eventually add support for other distribution extensions
		// "bdist_wininst": WinInst,
		// "bdist_egg":     BDist,
	}
	if err != nil {
		return nil, "", "", fmt.Errorf("invalid distribution file: %s, err: %v", filepath.Base(filename), err)
	}

	// If this encounters a metadata version it doesn't support, it may give us
	// back empty metadata. At the very least, we should have a name and version,
	// which could also be empty if, for example, a MANIFEST.in doesn't include
	// setup.cfg.
	if metadata.GetName() == "" || metadata.GetVersion() == "" {
		return nil, "", "", errors.New("metadata is missing required fields")
	}

	var pythonVersion string
	switch fileType {
	// TODO maybe eventually add support for other distribution extensions
	// case "bdist_egg":
	case "bdist_wheel", "bdist_wininst":
		pythonVersion = metadata.GetPythonVersion()
	default:
		pythonVersion = ""
	}

	return metadata, pythonVersion, fileType, nil
}
