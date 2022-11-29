/*
Copyright © 2022 Diego Morales <dgmorales@gmail.com>

*/
package version

import (
	"context"
	"fmt"
	"os"

	"github.com/dgmorales/go-cli-selfupdate/kube"
	semver "github.com/hashicorp/go-version"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

var Version = "0.2.6"

type versions struct {
	minimal *semver.Version
	latest  *semver.Version
}

type updateDecision int

const (
	IsLatest updateDecision = iota
	CanUpdate
	MustUpdate
)

//func SelfUpdate(kc *kubernetes.Clientset) error {
// cliInfo, err := kc.CoreV1().ConfigMaps(SYSTEM_NS).Get(context.Background(), "cli-info", metav1.GetOptions{})
// resp, err := http.Get(url)
// if err != nil {
// 	return err
// }
// defer resp.Body.Close()

// err = selfupdate.Apply(resp.Body, update.Options{})
// selfupdate.
// if err != nil {
// 	// error handling
// }
// return err
//	return nil
//}

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

	if val, ok := cm.Data["minimal-required-version"]; ok {
		res.minimal, err = semver.NewVersion(val)
		if err != nil {
			errorList = append(errorList, err)
		}
	}

	if val, ok := cm.Data["latest-version"]; ok {
		res.latest, err = semver.NewVersion(val)
		if err != nil {
			errorList = append(errorList, err)
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

	curVer, err := semver.NewVersion(Version)
	if err != nil {
		panic("This should never happen. Wrong version set internally.")
	}

	if v.minimal != nil && curVer.LessThan(v.minimal) {
		return MustUpdate, nil
	} // ignore if minimal required version is unset

	if v.latest != nil && curVer.LessThan(v.latest) {
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
		fmt.Printf("Fatal: your current version (%s) is not supported anymore (minimal: %s). You need to upgrade.\n", Version, versions.minimal.String())
		os.Exit(1)
		// TODO: Ask permission to selfupdate instead of exiting
	case CanUpdate:
			fmt.Printf("Warning: there's a newer version (%s), but this version (%s) is still usable.\n", versions.latest.String(), Version)
	}
}
