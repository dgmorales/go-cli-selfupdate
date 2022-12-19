/*
Copyright © 2022 Diego Morales <dgmorales@gmail.com>

*/
package version

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"runtime"
	"strings"

	"github.com/dgmorales/go-cli-selfupdate/logger"
	"github.com/google/go-github/v48/github"
	semver "github.com/hashicorp/go-version"
)

type GitHubChecker struct {
	minimal   *semver.Version
	current   *semver.Version
	latest    *semver.Version
	repoOwner string
	repoName  string
	assetName string
	assetID   int64
	assetUrl  string
}

// DiscoverLatest discovers what is the latest version from GitHub Releases
//
// It already saves asset information, leaving everything ready for calling Download()
func NewGithubChecker(owner string, repo string, minimalReq string, current string) (*GitHubChecker, error) {
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

	gh := github.NewClient(nil)

	latest, _, err := gh.Repositories.GetLatestRelease(context.Background(), owner, repo)
	if err != nil {
		return nil, fmt.Errorf("error getting latest github release from repo %s/%s: %w",
			owner, repo, err)
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
			ghc.assetUrl = asset.GetBrowserDownloadURL()
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
	if c == nil {
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

// Download downloads the saved GitHub Release Asset to a temporary file
func (c *GitHubChecker) DownloadLatest() (filename string, err error) {
	var httpClient http.Client

	if c.assetName == "" || c.assetID == 0 || c.assetUrl == "" {
		return "", errors.New("in Download: github release asset information is unavailable")
	}

	filename = path.Join(logger.WorkDir, c.assetName)

	f, err := os.Create(filename)
	if err != nil {
		return "", fmt.Errorf("error creating file for release download: %w", err)
	}
	defer f.Close()

	gh := github.NewClient(nil)

	data, redirectUrl, err := gh.Repositories.DownloadReleaseAsset(context.Background(),
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

// Check checks if current version can or must be updated
//
// It is OK for the version information on cliInfo to be unset (nil),
// maybe because of previous errors parsing it. That is a noop and
// will return IsLatest and nil error on that case.
//
// It does return error on some very abnormal cases (gross programming errors)
func (c *GitHubChecker) Check() (Assertion, error) {
	// current, err := semver.NewVersion(Version)
	// if err != nil {
	// 	return IsLatest, errors.New("in checkVersion: error parsing our current version. This should never happen")
	// }

	if c.minimal != nil && c.current.LessThan(c.minimal) {
		return MustUpdate, nil
	} // ignore if minimal required version is unset

	if c.latest != nil && c.current.LessThan(c.latest) {
		return CanUpdate, nil
	} // ignore if latest version is unset

	return IsLatest, nil
}
