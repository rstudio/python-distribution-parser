package parse_test

import (
	"fmt"
	"github.com/google/go-cmp/cmp"
	"github.com/rstudio/python-distribution-parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"math"
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
	"https://github.com/sdispater/pendulum",
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
func toRepositoryName(repositoryUrl string) (string, error) {
	result, err := url.Parse(repositoryUrl)
	if err != nil {
		return "", err
	}
	return path.Base(result.Path), nil
}

// getRepositoryPath returns the path the a repository is cloned at
func getRepositoryPath(repository string) string {
	return fmt.Sprintf("%s%s/", testdata, repository)
}

// getDistributionPath returns the path that distribution tarballs are kept
func getDistributionPath(repository string) string {
	return fmt.Sprintf("%sdist/", getRepositoryPath(repository))
}

// getTarballPath returns the path to a built tarball for a repository
func getTarballPath(repository string) (string, error) {
	distributionPath := getDistributionPath(repository)
	files, err := os.ReadDir(distributionPath)
	if err != nil {
		return "", err
	}

	var tarballs []string

	for _, file := range files {
		if filepath.Ext(file.Name()) == ".gz" {
			tarballs = append(tarballs, file.Name())
		}
	}

	if len(tarballs) != 1 {
		return "", fmt.Errorf("unexpected number of .gz files in %s: %d", distributionPath, len(tarballs))
	}
	return fmt.Sprintf("%s%s", distributionPath, tarballs[0]), nil
}

// clone will clone a Git repository to disk if it does not already exist
func clone(repositoryUrl string) error {
	repositoryName, err := toRepositoryName(repositoryUrl)
	if err != nil {
		return err
	}

	_, err = os.Stat(getRepositoryPath(repositoryName))
	if os.IsNotExist(err) {
		cmd := exec.Command("git", "clone", repositoryUrl)
		cmd.Dir = testdata
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		return err
	}

	return nil
}

// buildDistribution builds a Python package with python -m build
func buildDistribution(repository string) error {
	_, err := os.Stat(getDistributionPath(repository))
	if os.IsNotExist(err) {
		cmd := exec.Command("python", "-m", "build")
		cmd.Dir = getRepositoryPath(repository)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		return err
	}

	return nil
}

// getTwineMetadata returns the metadata that Twine generates
func getTwineMetadata(repository string) (map[string][]string, error) {
	var metadata map[string][]string

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseMultipartForm(math.MaxInt64)
		metadata = r.MultipartForm.Value
	}))
	defer ts.Close()

	tarball, err := getTarballPath(repository)
	if err != nil {
		return nil, err
	}

	cmd := exec.Command("twine", "upload", tarball)
	cmd.Env = append(cmd.Env, fmt.Sprintf("TWINE_REPOSITORY_URL=%s", ts.URL))
	// twine requires these variable to be set
	cmd.Env = append(cmd.Env, "TWINE_USERNAME=user")
	cmd.Env = append(cmd.Env, "TWINE_PASSWORD=password")

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		return nil, err
	}

	return metadata, nil
}

// getParserMetadata runs the Go parser and returns the resulting metadata
func getParserMetadata(repository string) (map[string][]string, error) {
	tarball, err := getTarballPath(repository)
	if err != nil {
		return nil, err
	}

	result, err := parse.Parse(tarball)
	if err != nil {
		return nil, err
	}

	if len(result) != 1 {
		return nil, fmt.Errorf("unexpected length: %d", len(result))
	}

	distribution := result[0]
	metadata := distribution.MetadataMap()

	return metadata, nil
}

// checkRequirements ensures that all test requirements are installed
func checkRequirements() error {
	_, err := exec.LookPath("twine")
	if err != nil {
		return err
	}

	_, err = exec.LookPath("python")
	if err != nil {
		return err
	}

	_, err = exec.LookPath("cargo")
	if err != nil {
		return err
	}

	_, err = exec.LookPath("git")
	if err != nil {
		return err
	}

	cmd := exec.Command("pip", "show", "build")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		return err
	}

	cmd = exec.Command("pip", "show", "wheel")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		return err
	}

	return nil
}

func TestParse(t *testing.T) {
	err := checkRequirements()
	require.NoError(t, err)

	for _, repositoryUrl := range repositoryUrls {
		t.Run(repositoryUrl, func(t *testing.T) {
			url := repositoryUrl
			repositoryName, err := toRepositoryName(url)
			assert.NoError(t, err)

			t.Parallel()
			err = clone(url)
			assert.NoError(t, err)

			err = buildDistribution(repositoryName)
			assert.NoError(t, err)

			expectedMetadata, err := getTwineMetadata(repositoryName)
			assert.NoError(t, err)

			actualMetadata, err := getParserMetadata(repositoryName)
			assert.NoError(t, err)

			// compare against the normalized outputs to account for expects differences between the two parsers
			assert.Empty(t, cmp.Diff(expectedMetadata, actualMetadata))
		})
	}
}
