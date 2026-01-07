package packages

import (
	"path/filepath"

	"github.com/rstudio/python-distribution-parser/internal/distributions"
	"github.com/rstudio/python-distribution-parser/types"
)

func NewPackageFile(filename string) (*types.PackageFile, error) {
	metadata, pythonVersion, fileType, err := distributions.NewDistributionMetadata(filename)
	if err != nil {
		return nil, err
	}

	baseFilename := filepath.Base(filename)
	safeName := distributions.SafeName(metadata.GetName())

	signedFilename := filename + ".asc"
	signedBaseFilename := baseFilename + ".asc"

	hashManager, err := NewHashManager(filename)
	if err != nil {
		return nil, err
	}
	err = hashManager.Hash()
	if err != nil {
		return nil, err
	}
	hexdigest := hashManager.HexDigest()

	return &types.PackageFile{
		Filename:           filename,
		Comment:            "", // Adding a comment isn't currently possible
		BaseFilename:       baseFilename,
		Metadata:           &metadata,
		PythonVersion:      pythonVersion,
		FileType:           fileType,
		SafeName:           safeName,
		SignedFilename:     signedFilename,
		SignedBaseFilename: signedBaseFilename,
		SHA2Digest:         hexdigest.sha2,
		Blake2_256Digest:   hexdigest.blake2,
	}, nil
}
