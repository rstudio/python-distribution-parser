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
	Filename           string                      `json:"file_name"`
	Comment            string                      `json:"comment"`
	BaseFilename       string                      `json:"base_filename"`
	Metadata           *distributions.Distribution `json:"metadata"`
	PythonVersion      string                      `json:"pyversion"`
	FileType           string                      `json:"filetype"`
	SafeName           string                      `json:"safe_name"`
	SignedFilename     string                      `json:"signed_filename"`
	SignedBaseFilename string                      `json:"signed_base_filename"`
	GPGSignature       *Signature                  `json:"gpg_signature"`
	MD5Digest          string                      `json:"md5_digest"`
	SHA2Digest         string                      `json:"sha256_digest"`
	Blake2_256Digest   string                      `json:"blake2_256_digest"`
}

type Signature struct {
	Filename string `json:"signed_filename"`
	Bytes    []byte `json:"signed_bytes"`
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
		Comment:            "", // Adding a comment isn't currently possible
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

func (pf *PackageFile) MetadataMap() map[string][]string {
	pkgMap := distributions.StructToMap(pf)
	metadata := *pf.Metadata
	metadataMap := metadata.MetadataMap()
	result := make(map[string][]string, len(pkgMap)+len(metadataMap))
	for pk, pv := range pkgMap {
		result[pk] = pv
	}
	for mk, mv := range metadataMap {
		result[mk] = mv
	}

	// This makes the request look more like Twine
	pkgMap["protocol_version"] = []string{"1"}
	delete(pkgMap, "metadata")

	return result
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
