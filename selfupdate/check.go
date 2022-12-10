/*
Copyright © 2022 Diego Morales <dgmorales@gmail.com>

*/
package selfupdate

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"log"
	"os"
	"reflect"
	"runtime"
	"strings"

	"github.com/dgmorales/go-cli-selfupdate/kube"
	"github.com/dgmorales/go-cli-selfupdate/logger"
	"github.com/dgmorales/go-cli-selfupdate/version"
	"github.com/google/go-github/v48/github"
	semver "github.com/hashicorp/go-version"
	"github.com/mitchellh/mapstructure"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type cliInfo struct {
	RepositoryOwner        string
	RepositoryName         string
	MinimalRequiredVersion *semver.Version
	latestVersion          *semver.Version
	assetName              string
	assetUrl               string
}

type updateDecision int

const (
	IsLatest updateDecision = iota
	CanUpdate
	MustUpdate
)

// StringToVersionHookFunc is a mapstructure HookFunc to convert
// a string to a hashicorp/go-version Version pointer
//
// This will be passed to mapstructure.Decoder and called for every field
// in the source structure being mapped, so we must "passthrough" if the
// field types don't match
func StringToVersionHookFunc(from reflect.Type, to reflect.Type, data interface{}) (interface{}, error) {
	if to != reflect.TypeOf(semver.Version{}) {
		return data, nil
	}
	if from.Kind() != reflect.String {
		return data, nil
	}

	// Convert it by parsing
	v, err := semver.NewVersion(data.(string))
	if err != nil {
		return &semver.Version{}, fmt.Errorf("failed parsing version %v", data)
	}

	return v, nil
}

// readFromCfgMap gets version information from a Kubernetes ConfigMap
//
// We want to continue if some of the fields of the versions ConfigMap are missing
// We also want to continue on if there are errors on the remote specification of versions,
// but we don't want to fully ignore the errors. It can be useful to know about them and
// report them somwhere, so we build an error list with any error we find.
//
// versions returned may be a struct with only nil values if errors are found.
func (i *cliInfo) readFromCfgMap(kc *kubernetes.Clientset, ns string, name string) error {
	if i == nil {
		return errors.New("in readFromCfgMap: called with a nil receiver")
	}

	cm, err := kc.CoreV1().ConfigMaps(ns).Get(context.Background(), name,
		metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("error reading configmap %s/%s: %w", ns, name, err)
	}

	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		DecodeHook: StringToVersionHookFunc,
		Result:     &i,
	})
	if err != nil {
		return fmt.Errorf("error parsing configmap %s/%s: %w", ns, name, err)
	}

	err = decoder.Decode(cm.Data)
	if err != nil {
		return fmt.Errorf("error parsing configmap %s/%s: %w", ns, name, err)
	}

	return nil
}

// readFromGitHub reads the latest version from GitHub Releases
//
// It also finds the correct asset to use from the release assets
func (i *cliInfo) readFromGitHub(gh *github.Client) error {
	if i == nil {
		return errors.New("in readFromGitHub: called with a nil receiver")
	}

	if i.RepositoryOwner == "" || i.RepositoryName == "" {
		return errors.New("error getting latest github release, owner or repo are unset")
	}

	rel, _, err := gh.Repositories.GetLatestRelease(context.Background(),
		i.RepositoryOwner, i.RepositoryName)
	if err != nil {
		return fmt.Errorf("error getting latest github release from repo %s/%s: %w",
			i.RepositoryOwner, i.RepositoryName, err)
	}

	i.latestVersion, err = semver.NewVersion(rel.GetTagName())
	if err != nil {
		return fmt.Errorf("error parsing github release tag %s as version, from repo %s/%s: %w",
			rel.GetTagName(), i.RepositoryOwner, i.RepositoryName, err)
	}

	// asset must contain goos string (linux|windows|darwin) on filename
	for _, asset := range rel.Assets {
		if strings.Contains(asset.GetName(), runtime.GOOS) {
			i.assetName = asset.GetName()
			i.assetUrl = asset.GetBrowserDownloadURL()
		}
	}
	return nil
}

// checkVersion checks if current version can or must be updated
//
// check is internal and meant to be easily testable
// It is OK for the version information on cliInfo to be unset (nil),
// maybe because of previous errors parsing it. That is a noop and
// will return IsLatest and nil error on that case.
//
// It does return error on some very abnormal cases (gross programming errors)
func (i *cliInfo) CheckVersion() (updateDecision, error) {
	if i == nil {
		return IsLatest, errors.New("in checkVersion: called with a nil receiver")
	}

	current, err := semver.NewVersion(version.Version)
	if err != nil {
		return IsLatest, errors.New("in checkVersion: error parsing our current version. This should never happen")
	}

	err = i.readFromCfgMap(kube.Client, kube.SYSTEM_NS, "cli-info")
	if err != nil {
		return IsLatest, err
	}

	err = i.readFromGitHub(github.NewClient(nil))
	if err != nil {
		return IsLatest, err
	}

	if i.MinimalRequiredVersion != nil && current.LessThan(i.MinimalRequiredVersion) {
		return MustUpdate, nil
	} // ignore if minimal required version is unset

	if i.latestVersion != nil && current.LessThan(i.latestVersion) {
		return CanUpdate, nil
	} // ignore if latest version is unset

	return IsLatest, nil
}

// Check checks if current version can or must be update, and interacts with the user
// about it
func Check() {
	i := cliInfo{}
	decision, err := i.CheckVersion()
	// We will just log errors below and continue, without disturbing user interaction
	// flow. Version check and update is a non essential feature.
	if err != nil {
		fmt.Printf("Info: we are having some trouble checking for a new version of the CLI. Check details on logfile %s\n", logger.LogFile)
		log.Println(err)
	}
	log.Printf("debug: cli information dump: %v\n", i)

	switch decision {
	case MustUpdate:
		fmt.Printf("Fatal: your current version (%s) is not supported anymore (minimal: %s). You need to upgrade.\n",
			version.Version, i.MinimalRequiredVersion.String())
		//doSelfUpdate(versions.repository)
		// TODO: Ask permission to selfupdate instead of exiting
		os.Exit(0)
	case CanUpdate:
		fmt.Printf("Warning: there's a newer version (%s), but this version (%s) is still usable.\n",
			i.latestVersion.String(), version.Version)
	}
}

// func doSelfUpdate(repo string) {
// }