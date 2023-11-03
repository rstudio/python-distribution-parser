package parse_test

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/bradleyjkemp/cupaloy"
	"github.com/google/go-cmp/cmp"
	"github.com/rstudio/python-distribution-parser"
	"github.com/rstudio/python-distribution-parser/types"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"log"
	"maps"
	"math"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path"

	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// testdata is the path that we should store cloned repositories
var testdata = "testdata/repositories/"

// repositoryUrls is the list of repositories that we should test
var repositoryUrls = []string{
	"https://github.com/ActiveState/appdirs",
	"https://github.com/pallets/click",
	"https://github.com/python/importlib_metadata",
	"https://github.com/matplotlib/matplotlib",
	"https://github.com/pytest-dev/pytest",
	"https://github.com/tkem/cachetools/",
	"https://github.com/certifi/python-certifi",
	"https://github.com/chardet/chardet",
	"https://github.com/jaraco/configparser/",
	"https://github.com/nedbat/coveragepy",
	"https://github.com/micheles/decorator",
	"https://github.com/tiran/defusedxml",
}

// toRepositoryName converts a repository name to the name of the folder the repository will be cloned in
// for example, https://github.com/ActiveState/appdirs => appdirs
func toRepositoryName(repositoryUrl string) string {
	result, err := url.Parse(repositoryUrl)
	if err != nil {
		log.Fatalf("%v", err)
	}
	return path.Base(result.Path)
}

// getRepositoryPath returns the path the a repository is cloned at
func getRepositoryPath(repository string) string {
	return fmt.Sprintf("%s%s/", testdata, repository)
}

// getDistributionPath returns the path that distribution tarballs are kept
func getDistributionPath(repository string) string {
	return fmt.Sprintf("%sdist/", getRepositoryPath(repository))
}

// getArtifactPath returns the path to a built tarball/wheel for a repository
func getArtifactPath(path string, extension string) (string, error) {
	files, err := os.ReadDir(path)
	if err != nil {
		return "", err
	}

	var artifacts = lo.Filter(files, func(file os.DirEntry, _ int) bool {
		return filepath.Ext(file.Name()) == extension
	})

	if len(artifacts) != 1 {
		return "", fmt.Errorf("expected exactly 1 ParserData with extension %s in %s, but found %d", extension, path, len(artifacts))
	}

	return fmt.Sprintf("%s%s", path, artifacts[0].Name()), nil
}

// clone will clone a Git repository to disk if it does not already exist
func clone(repositoryUrl string, destination string) error {
	_, err := os.Stat(destination)
	if os.IsNotExist(err) {
		cmd := exec.Command("git", "clone", repositoryUrl, destination)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		return err
	}

	return nil
}

// buildDistribution builds a Python package with python -m build
func buildDistribution(directory string) error {
	_, err := os.Stat(fmt.Sprintf("%s/dist", directory))
	if os.IsNotExist(err) {
		cmd := exec.Command("python", "-m", "build")
		cmd.Dir = directory
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		return err
	}

	return nil
}

// signDistribution will use gpg to sign a ParserData, and return the path to the resulting `.asc` ParserData
func signDistribution(file string) (string, error) {
	signatureFile := fmt.Sprintf("%s.asc", file)

	_, err := os.Stat(signatureFile)
	if os.IsNotExist(err) {
		cmd := exec.Command("gpg", "--detach-sign", "-a", file)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		if err != nil {
			return "", err
		}
		return signatureFile, nil
	}

	return signatureFile, nil
}

// getTwineFile returns the Metadata that Twine generates
func getTwineFile(file string, signatureFile string) (ParserData, error) {
	var metadata map[string][]string
	var gpgSignature []byte
	var err error

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseMultipartForm(math.MaxInt64)
		metadata = r.MultipartForm.Value
		file, _, err := r.FormFile("gpg_signature")
		if err != nil {
			if errors.Is(err, http.ErrMissingFile) {
				return
			} else {
				log.Fatalf("error while reading multipart file: %v", err)
			}
		}
		defer func(file multipart.File) {
			err := file.Close()
			if err != nil {
				log.Fatalf("error while closing multipart file: %v", err)
			}
		}(file)
		buf := bytes.NewBuffer(nil)
		if _, err := io.Copy(buf, file); err != nil {
			log.Fatalf("error while reading multipart file: %v", err)
		}
		gpgSignature = buf.Bytes()
	}))
	defer ts.Close()

	var cmd *exec.Cmd

	if signatureFile == "" {
		cmd = exec.Command("twine", "upload", file)
	} else {
		cmd = exec.Command("twine", "upload", file, signatureFile)
	}

	cmd.Env = append(cmd.Env, fmt.Sprintf("TWINE_REPOSITORY_URL=%s", ts.URL))
	// twine requires these variable to be set
	cmd.Env = append(cmd.Env, "TWINE_USERNAME=user")
	cmd.Env = append(cmd.Env, "TWINE_PASSWORD=password")

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		return ParserData{}, err
	}

	return ParserData{
		Metadata:     metadata,
		GpgSignature: gpgSignature,
	}, nil
}

// getParserFile runs the Go parser and returns the resulting Metadata
func getParserFile(file string, signatureFile string) (ParserData, error) {
	var packageFiles []*types.PackageFile
	var err error

	if signatureFile == "" {
		packageFiles, err = parse.Parse(file)
	} else {
		packageFiles, err = parse.Parse(file, signatureFile)
	}

	if err != nil {
		return ParserData{}, err
	}

	if len(packageFiles) != 1 {
		return ParserData{}, fmt.Errorf("unexpected length: %d", len(packageFiles))
	}

	distribution := packageFiles[0]
	metadata := distribution.MetadataMap()

	if distribution.GPGSignature == nil {
		return ParserData{
			Metadata:     metadata,
			GpgSignature: nil,
		}, nil
	} else {
		return ParserData{
			Metadata:     metadata,
			GpgSignature: distribution.GPGSignature.Bytes,
		}, nil
	}
}

// checkRequirements ensures that all test requirements are installed
func checkRequirements() error {

	_, err := exec.LookPath("python")
	if err != nil {
		return errors.New("python is not installed")
	}

	_, err = exec.LookPath("twine")
	if err != nil {
		return errors.New("twine is not installed")
	}

	_, err = exec.LookPath("cargo")
	if err != nil {
		return errors.New("rust is not installed")
	}

	_, err = exec.LookPath("git")
	if err != nil {
		return errors.New("git is not installed")
	}

	_, err = exec.LookPath("gpg")
	if err != nil {
		return errors.New("gpg is not installed")
	}

	cmd := exec.Command("pip", "show", "build")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		return errors.New("build package is not installed")
	}

	cmd = exec.Command("pip", "show", "wheel")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		return errors.New("wheel package is not installed")
	}

	return nil
}

// createDistribution builds a Python distribution and optionally signs it.
// The method returns the location of the directory, built artifact, and signature ParserData if any.
func createDistribution(repositoryUrl string, format string, isSigned bool) (string, string, error) {
	repoUrl := repositoryUrl
	repositoryName := toRepositoryName(repoUrl)

	repositoryPath := getRepositoryPath(repositoryName)

	err := clone(repoUrl, repositoryPath)
	if err != nil {
		return "", "", err
	}

	err = buildDistribution(repositoryPath)
	if err != nil {
		return "", "", err
	}

	distributionPath := getDistributionPath(repositoryName)

	artifact, err := getArtifactPath(distributionPath, format)
	if err != nil {
		return "", "", err
	}

	var signature string

	if isSigned {
		signature, err = signDistribution(artifact)
		if err != nil {
			return "", "", err
		}
	}

	return artifact, signature, nil
}

// signedString returns the correct form of "signed", so that unit test names look appropriate
func signedString(value bool) string {
	if value {
		return "signed"
	} else {
		return "unsigned"
	}
}

// formatString returns a string for unit test names
func formatString(value string) string {
	if value == ".gz" {
		return "tarball"
	} else if value == ".whl" {
		return "wheel"
	} else {
		return ""
	}
}

type testCase struct {
	repositoryUrl string
	isSigned      bool
	format        string
}

type ParserData struct {
	Metadata     map[string][]string
	GpgSignature []byte
}

// filterNondeterministicData replaces some generated fields with static strings
func filterNondeterministicData(data ParserData) ParserData {
	copiedData := maps.Clone(data.Metadata)

	fields := []string{
		"blake2_256_digest",
		"md5_digest",
		"sha256_digest",
	}

	for _, field := range fields {
		if _, ok := copiedData[field]; ok {
			copiedData[field] = []string{fmt.Sprintf("%s exists", field)}
		}
	}

	if data.GpgSignature == nil {
		return ParserData{
			Metadata: copiedData,
		}
	} else {
		return ParserData{
			Metadata:     copiedData,
			GpgSignature: []byte("GPG signature exists"),
		}
	}
}

func TestParse(t *testing.T) {
	err := checkRequirements()
	require.NoError(t, err)

	testCases := lo.FlatMap(repositoryUrls, func(repositoryUrl string, _ int) []testCase {
		return lo.FlatMap([]bool{true, false}, func(isSigned bool, _ int) []testCase {
			return lo.FlatMap([]string{".gz", ".whl"}, func(format string, _ int) []testCase {
				repositoryName := toRepositoryName(repositoryUrl)
				noWheels := []string{}
				if lo.Contains(noWheels, repositoryName) && format == ".whl" {
					return []testCase{}
				}
				return []testCase{
					{
						repositoryUrl: repositoryUrl,
						isSigned:      isSigned,
						format:        format,
					},
				}
			})
		})
	})

	for _, testCase := range testCases {
		repositoryName := toRepositoryName(testCase.repositoryUrl)
		require.NoError(t, err)

		repositoryUrl := testCase.repositoryUrl
		isSigned := testCase.isSigned
		format := testCase.format

		t.Run(fmt.Sprintf("%s %s %s", repositoryName, signedString(isSigned), formatString(format)), func(t *testing.T) {
			artifact, signature, err := createDistribution(repositoryUrl, format, isSigned)
			require.NoError(t, err)

			expectedMetadata, err := getTwineFile(artifact, signature)
			require.NoError(t, err)

			actualMetadata, err := getParserFile(artifact, signature)
			require.NoError(t, err)

			// compare against the normalized outputs to account for expects differences between the two parsers
			assert.Empty(t, cmp.Diff(expectedMetadata, actualMetadata))

			cupaloy.SnapshotT(t, filterNondeterministicData(actualMetadata))
		})
	}
}
