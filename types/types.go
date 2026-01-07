package types

import (
	"errors"
	"io"
	"log"
	"os"

	"github.com/rstudio/python-distribution-parser/internal/distributions"
	"github.com/samber/lo"
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
	SHA2Digest         string                      `json:"sha256_digest"`
	Blake2_256Digest   string                      `json:"blake2_256_digest"`
}

type Signature struct {
	Filename string `json:"signed_filename"`
	Bytes    []byte `json:"signed_bytes"`
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

	result["name"] = result["safe_name"]
	// This makes the request look more like Twine
	result["protocol_version"] = []string{"1"}
	result[":action"] = []string{"file_upload"}

	ignoredKeys := []string{
		"base_filename",
		"file_name",
		"safe_name",
		"signed_base_filename",
		"signed_filename",
		"metadata",
		"gpg_signature",
	}

	for _, key := range ignoredKeys {
		delete(result, key)
	}

	// remove any keys that are an empty value, unless twine expects them
	result = lo.OmitBy(result, func(key string, value []string) bool {
		return value == nil || len(value) == 1 && (value[0] == "" || value[0] == "<nil>")
	})

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
	defer func(gpg *os.File) {
		err := gpg.Close()
		if err != nil {
			log.Printf("error closing file: %v", err)
		}
	}(gpg)

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
