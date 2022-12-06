/*
Copyright © 2022 Diego Morales <dgmorales@gmail.com>

*/
package selfupdate

import (
	"context"
	_ "embed"
	"fmt"
	"log"
	"os"

	"github.com/blang/semver"
	"github.com/dgmorales/go-cli-selfupdate/kube"
	"github.com/dgmorales/go-cli-selfupdate/version"
	"github.com/rhysd/go-github-selfupdate/selfupdate"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type versions struct {
	repository string
	minimal    *semver.Version
	latest     *semver.Version
}

type updateDecision int

const (
	IsLatest updateDecision = iota
	CanUpdate
	MustUpdate
)

func doSelfUpdate(repo string) {
	if repo == "" {
		fmt.Println("Can't update")
		return
	}
	v := semver.MustParse(version.Version)
	latest, err := selfupdate.UpdateSelf(v, repo)
	if err != nil {
		log.Println("Binary update failed:", err)
		return
	}
	if latest.Version.Equals(v) {
		// latest version is the same as current version. It means current binary is up to date.
		log.Println("Current binary is the latest version", version.Version)
	} else {
		log.Println("Successfully updated to version", latest.Version)
		log.Println("Release note:\n", latest.ReleaseNotes)
	}
}

// getVersionsFromConfigMap get version information from a Kubernetes ConfigMap
//
// We want to continue if some of the fields of the versions ConfigMap are missing
// We also want to continue on if there are errors on the remote specification of versions,
// but we don't want to fully ignore the errors. It can be useful to know about them and
// report them somwhere, so we build an error list with any error we find.
//
// versions returned may be a struct with only nil values if errors are found.
func getVersionsFromConfigMap(kc *kubernetes.Clientset, ns string, name string) (versions, []error) {
	var res versions
	var errorList []error

	cm, err := kc.CoreV1().ConfigMaps(ns).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return res, append(errorList, err)
	}

	if val, ok := cm.Data["repository"]; ok {
		res.repository = val
	}

	if val, ok := cm.Data["minimal-required-version"]; ok {
		v, err := semver.Make(val)
		if err != nil {
			errorList = append(errorList, err)
		} else {
			res.minimal = &v
		}
	}

	if val, ok := cm.Data["latest-version"]; ok {
		v, err := semver.Make(val)
		if err != nil {
			errorList = append(errorList, err)
		} else {
			res.latest = &v
		}
	}

	return res, errorList
}

// check checks if current version can or must be updated
//
// check is internal and meant to be easily testable
// If v contains only nils, this is a noop and will return IsLatest
func check(v versions) (updateDecision, error) {
	res := IsLatest

	curVer, err := semver.Make(version.Version)
	if err != nil {
		panic("This should never happen. Wrong version set internally.")
	}

	if v.minimal != nil && curVer.Compare(*v.minimal) == -1 {
		return MustUpdate, nil
	} // ignore if minimal required version is unset

	if v.latest != nil && curVer.Compare(*v.latest) == -1 {
		return CanUpdate, nil
	} // ignore if latest version is unset

	return res, nil
}

// Check checks if current version can or must be update, and interacts with the user about it
func Check() {
	versions, _ := getVersionsFromConfigMap(kube.Client, kube.SYSTEM_NS, "cli-info")
	decision, _ := check(versions)
	// For now, we will ignore errors reading the config map
	// And that's safe:
	// versions won't be nil, but may contain only nil values
	// and check won't mind that.

	switch decision {
	case MustUpdate:
		fmt.Printf("Fatal: your current version (%s) is not supported anymore (minimal: %s). You need to upgrade.\n",
			version.Version, versions.minimal.String())
		doSelfUpdate(versions.repository)
		// TODO: Ask permission to selfupdate instead of exiting
		os.Exit(0)
	case CanUpdate:
		fmt.Printf("Warning: there's a newer version (%s), but this version (%s) is still usable.\n",
			versions.latest.String(), version.Version)
	}
}
