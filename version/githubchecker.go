package version

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/google/go-github/v48/github"
	semver "github.com/hashicorp/go-version"
)

type GitHubChecker struct {
	client    *github.Client
	minimal   *semver.Version
	current   *semver.Version
	latest    *semver.Version
	repoOwner string
	repoName  string
	assetName string
	assetID   int64
}

// DiscoverLatest discovers what is the latest version from GitHub Releases
//
// It already saves asset information, leaving everything ready for calling Download()
func NewGithubChecker(client *github.Client, owner string, repo string, minimalReq string, current string) (*GitHubChecker, error) {
	var err error
	ghc := GitHubChecker{}

	if owner == "" || repo == "" {
		return nil, errors.New("error getting latest github release, owner or repo are unset")
	}
	ghc.repoOwner = owner
	ghc.repoName = repo

	ghc.current, err = semver.NewSemver(current)
	if err != nil {
		return nil, err
	}

	if strings.TrimSpace(minimalReq) != "" {
		ghc.minimal, err = semver.NewSemver(minimalReq)
		if err != nil {
			return nil, err
		}
	}

	if client == nil {
		ghc.client = github.NewClient(nil)
	} else {
		ghc.client = client
	}

	latest, _, err := ghc.client.Repositories.GetLatestRelease(context.Background(), owner, repo)
	if err != nil {
		// This special handling bellow with error.As is necessary because we want to log
		// the error message, and that triggers a weird bug with github.ErrorResponse
		// wrapping and go-github-mock.
		//
		// There is an issue open there (with no solution at the time of this writing):
		// https://github.com/migueleliasweb/go-github-mock/issues/6
		var ghErr *github.ErrorResponse
		var errStr string

		if errors.As(err, &ghErr) {
			errStr = ghErr.Message
		} else {
			errStr = err.Error()
		}

		// We won't error if we can't get the latest version.
		// Instead we will proceed without the latest version (and assets) info
		log.Printf("error getting latest github release from repo %s/%s: %s. Will ignore and continue with latest version as unknown",
			owner, repo, errStr)
		return &ghc, nil
	}

	ghc.latest, err = semver.NewSemver(latest.GetTagName())
	if err != nil {
		return nil, err
	}

	// asset must contain goos string (linux|windows|darwin) on filename
	for _, asset := range latest.Assets {
		if strings.Contains(asset.GetName(), runtime.GOOS) {
			ghc.assetName = asset.GetName()
			ghc.assetID = asset.GetID()
		}
	}

	return &ghc, err
}

// Current returns current version as string
func (c *GitHubChecker) Current() string {
	if c == nil {
		return ""
	}
	return c.current.String()
}

// Minimal returns minimal required version as string
func (c *GitHubChecker) Minimal() string {
	if c == nil || c.minimal == nil {
		return ""
	}
	return c.minimal.String()
}

// Latest returns latest version as string
func (c *GitHubChecker) Latest() string {
	if c == nil {
		return ""
	}
	return c.latest.String()
}

func openFileForDownload(name string) (filename string, fd *os.File, err error) {
	filename = path.Join(os.TempDir(), fmt.Sprintf("%s-%s", filepath.Base(os.Args[0]), name))

	fd, err = os.Create(filename)
	if err != nil {
		return filename, nil, fmt.Errorf("error creating file for release download: %w", err)
	}

	return filename, fd, nil
}

// DownloadLatest downloads the saved GitHub Release Asset to a temporary file
func (c *GitHubChecker) DownloadLatest() (filename string, err error) {
	var httpClient http.Client

	if c == nil {
		return "", fmt.Errorf("in GitHubChecker.DownloadLatest: called with nil receiver")
	}

	if c.assetName == "" || c.assetID == 0 {
		return "", errors.New("in Download: github release asset information is unavailable")
	}

	filename, f, err := openFileForDownload(c.assetName)
	if err != nil {
		return filename, err
	}
	defer f.Close()

	data, redirectUrl, err := c.client.Repositories.DownloadReleaseAsset(context.Background(),
		c.repoOwner, c.repoName, c.assetID, &httpClient)
	if redirectUrl != "" {
		return filename, fmt.Errorf("in Download: got unsupported asset redirect URL (%s)", redirectUrl)
	}
	if err != nil {
		return filename, fmt.Errorf("in Download: %w", err)
	}

	_, err = io.Copy(f, data)
	if err != nil {
		return filename, fmt.Errorf("in Download: %w", err)
	}

	return filename, nil
}

// Check discovers if the current version can or must be updated.
//
// More precisely, it checks in which version.Assertion case the current version falls
// in.
//
// It tries to be resilient and always return the best assertion it can about the
// current version. So for example, if the minimal version is unknown it will ignore
// that check and check against the latest.
//
// If it cannot tell anything, it returns is IsUnknown.
func (c *GitHubChecker) Check() Assertion {
	if c == nil || c.current == nil {
		// return now otherwise we would panic trying the comparisons bellow
		return IsUnknown
	}

	if c.latest != nil && c.current.Equal(c.latest) {
		return IsLatest
	}

	if c.minimal != nil && c.current.LessThan(c.minimal) {
		return MustUpdate
	}

	if c.latest != nil && c.current.LessThan(c.latest) {
		return CanUpdate
	}

	if c.latest != nil && c.current.GreaterThan(c.latest) {
		return IsBeyond
	}

	return IsUnknown
}
