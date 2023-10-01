package packages

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"regexp"

	"github.com/rstudio/python-distribution-parser/distributions"
)

type PackageFile struct {
	Filename           string                      `json:"filename"`
	BaseFilename       string                      `json:"base_filename"`
	Metadata           *distributions.Distribution `json:"metadata"`
	PythonVersion      string                      `json:"python_version"`
	FileType           string                      `json:"file_type"`
	SafeName           string                      `json:"safe_name"`
	SignedFilename     string                      `json:"signed_filename"`
	SignedBaseFilename string                      `json:"signed_base_filename"`
	GPGSignature       *Signature                  `json:"gpg_signature"`
	MD5Digest          string                      `json:"md5_digest"`
	SHA2Digest         string                      `json:"sha2_digest"`
	Blake2_256Digest   string                      `json:"blake2_256_digest"`
}

type Signature struct {
	Filename string `json:"filename"`
	Bytes    []byte `json:"bytes"`
}

// Convert an arbitrary string to a standard distribution name.
// Any runs of non-alphanumeric/. characters are replaced with a single '-'.
// Copied from pkg_resources.safe_name for compatibility with warehouse.
// See https://github.com/pypa/twine/issues/743.
func safeName(name string) string {
	reg := regexp.MustCompile("[^A-Za-z0-9.]+")
	return reg.ReplaceAllString(name, "-")
}

func NewPackageFile(filename string) (*PackageFile, error) {
	metadata, pythonVersion, fileType, err := distributions.NewDistributionMetadata(filename)
	if err != nil {
		return nil, err
	}

	baseFilename := filepath.Base(filename)
	safeName := safeName(metadata.GetName())

	signedFilename := filename + ".asc"
	signedBaseFilename := baseFilename + ".asc"

	hashManager, err := NewHashManager(filename)
	if err != nil {
		return nil, err
	}
	hashManager.Hash()
	hexdigest := hashManager.HexDigest()

	return &PackageFile{
		Filename:           filename,
		BaseFilename:       baseFilename,
		Metadata:           &metadata,
		PythonVersion:      pythonVersion,
		FileType:           fileType,
		SafeName:           safeName,
		SignedFilename:     signedFilename,
		SignedBaseFilename: signedBaseFilename,
		MD5Digest:          hexdigest.md5,
		SHA2Digest:         hexdigest.sha2,
		Blake2_256Digest:   hexdigest.blake2,
	}, nil
}

func (pf *PackageFile) AddGPGSignature(signatureFilepath string, signatureFilename string) error {
	if pf.GPGSignature != nil {
		return errors.New("GPG Signature can only be added once")
	}

	gpg, err := os.Open(signatureFilepath)
	if err != nil {
		return err
	}
	defer gpg.Close()

	bytes, err := io.ReadAll(gpg)
	if err != nil {
		return err
	}

	pf.GPGSignature = &Signature{
		Filename: signatureFilename,
		Bytes:    bytes,
	}
	return nil
}
