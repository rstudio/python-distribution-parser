package parse

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/rstudio/python-distribution-parser/packages"
)

func endsWith(str, suffix string) bool {
	return len(str) >= len(suffix) && str[len(str)-len(suffix):] == suffix
}

func groupWheelFilesFirst(files []string) []string {
	// Check if there's any wheel file
	hasWheel := false
	for _, fname := range files {
		if endsWith(fname, ".whl") {
			hasWheel = true
			break
		}
	}

	if !hasWheel {
		return files
	}

	sort.Slice(files, func(i, j int) bool {
		return endsWith(files[i], ".whl") && !endsWith(files[j], ".whl")
	})

	return files
}

func findDistributions(dists []string) ([]string, error) {
	var packages []string
	for _, filename := range dists {
		if _, err := os.Stat(filename); err == nil {
			packages = append(packages, filename)
			continue
		}

		files, err := filepath.Glob(filename)
		if err != nil {
			return nil, err
		}
		if files == nil || len(files) == 0 {
			return nil, fmt.Errorf("cannot find file (or expand pattern): %s", filename)
		}

		packages = append(packages, files...)
	}

	return groupWheelFilesFirst(packages), nil
}

func makePackage(filename string, signatures map[string]string) (*packages.PackageFile, error) {
	packageFile, err := packages.NewPackageFile(filename)
	if err != nil {
		return nil, fmt.Errorf("unable to create packageFile from %s: %v\n", filename, err)
	}

	signedName := packageFile.SignedBaseFilename
	if signature, exists := signatures[signedName]; exists {
		packageFile.AddGPGSignature(signature, signedName)
	}

	_, err = packages.GetFileSize(packageFile.Filename)
	if err != nil {
		return nil, fmt.Errorf("%s is not a real file\n", packageFile.Filename)
	}
	return packageFile, nil
}

func parse(dists []string) ([]*packages.PackageFile, error) {
	dists, err := findDistributions(dists)
	if err != nil {
		return nil, fmt.Errorf("error finding distributions: %v\n", err)
	}

	// Initialize maps for signatures and a slice for distributions
	signatures := make(map[string]string)
	var distributions []string

	// Separate signatures and distributions
	for _, d := range dists {
		base := filepath.Base(d)
		if strings.HasSuffix(d, ".asc") {
			signatures[base] = d
		} else {
			distributions = append(distributions, d)
		}
	}

	var packages []*packages.PackageFile
	for _, filename := range distributions {
		p, err := makePackage(filename, signatures)
		if err != nil {
			return nil, err
		}
		packages = append(packages, p)
	}

	return packages, nil
}

func Parse(path string) ([]*packages.PackageFile, error) {
	var files []string
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("%s does not exist: %v\n", path, err)
		}
		return nil, err
	}

	if info.IsDir() {
		dirFiles, err := os.ReadDir(path)
		if err != nil {
			return nil, err
		}
		for _, entry := range dirFiles {
			// Don't recursively go in directories, only go one level deep.
			if !entry.IsDir() {
				fullPath := filepath.Join(path, entry.Name())
				files = append(files, fullPath)
			}
		}
	} else {
		files = append(files, path)
	}

	if len(files) == 0 {
		return nil, errors.New("no files to parse")
	}

	return parse(files)
}
