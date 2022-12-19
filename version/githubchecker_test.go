package version_test

import (
	"testing"

	"github.com/dgmorales/go-cli-selfupdate/version"
)

func TestGitHubCheckerImplementsChecker(t *testing.T) {
	var i interface{} = new(version.GitHubChecker)
	if _, ok := i.(version.Checker); !ok {
		t.Fatalf("expected %T to implement version2.Checker", i)
	}
}
