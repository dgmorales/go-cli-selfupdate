package version_test

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/dgmorales/go-cli-selfupdate/logger"
	"github.com/dgmorales/go-cli-selfupdate/version"
	"github.com/google/go-github/v48/github"
	"github.com/migueleliasweb/go-github-mock/src/mock"
)

const (
	fakeOrg          = "someorg"
	fakeRepo         = "somerepo"
	fakeAssetContent = "yeah-of-course-i-was-downloaded"
)

type versionsCaseSpec struct {
	desc      string
	min       string
	cur       string
	latest    string
	expMin    string // expected min, cur, max version string
	expCur    string // (if blank, ignore this and use originals )
	expLatest string
	vAssert   version.Assertion
}

// unPrefixV removes v prefix from input string (common prefix for version strings)
func unPrefixV(s string) string {
	return strings.TrimLeft(s, "v")
}

func strp(s string) *string {
	return &s
}

func int64p(i int64) *int64 {
	return &i
}

func assertStr(a version.Assertion) string {
	switch a {
	case version.IsLatest:
		return "is latest"
	case version.CanUpdate:
		return "can update"
	case version.MustUpdate:
		return "must update"
	case version.IsBeyond:
		return "is beyond"
	}

	return "is unknown"
}

// newReleasesGitHubMock simulates a scenario where repo has latest
// release set to latestV
func newReleasesGitHubMock(latestV string) *github.Client {
	m := mock.NewMockedHTTPClient(
		mock.WithRequestMatch(
			mock.GetReposReleasesLatestByOwnerByRepo,
			github.RepositoryRelease{
				TagName: strp(latestV),
				Assets: []*github.ReleaseAsset{
					{
						ID:   int64p(100),
						Name: strp(fmt.Sprintf("test-v%s-darwin.tar.gz", unPrefixV(latestV))),
					},
					{
						ID:   int64p(101),
						Name: strp(fmt.Sprintf("test-v%s-linux.tar.gz", unPrefixV(latestV))),
					},
					{
						ID:   int64p(102),
						Name: strp(fmt.Sprintf("test-v%s-windows.zip", unPrefixV(latestV))),
					},
				},
			},
		),
		mock.WithRequestMatch(
			mock.GetReposReleasesAssetsByOwnerByRepoByAssetId,
			fakeAssetContent,
		),
	)

	return github.NewClient(m)
}

// newNotReleasedGitHubMock simulates a scenario where latest
// release can't be found
func newNotReleasedGitHubMock() *github.Client {
	// We should error Getting the Release, as if not found
	m := mock.NewMockedHTTPClient(
		mock.WithRequestMatchHandler(
			mock.GetReposReleasesLatestByOwnerByRepo,
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				mock.WriteError(
					w,
					http.StatusNotFound,
					"Not Found",
				)
			}),
		),
	)
	return github.NewClient(m)
}

func newGitHubMock(latestV string) *github.Client {
	if strings.TrimSpace(latestV) == "" {
		return newNotReleasedGitHubMock()
	} else {
		return newReleasesGitHubMock(latestV)
	}

}

// verifyExpectedVersions checks that the versions in the input GitHubChecker exactly match min, cur and latest
func verifyExpectedVersions(t *testing.T, gc *version.GitHubChecker, tc versionsCaseSpec) {

	t.Helper()

	c := func(ver string, exp string) string {
		if strings.TrimSpace(exp) == "" {
			return unPrefixV(ver)
		}
		return unPrefixV(exp)
	}

	min := c(tc.min, tc.expMin)
	cur := c(tc.cur, tc.expCur)
	latest := c(tc.latest, tc.expLatest)

	if gc.Minimal() != min {
		t.Errorf("expected minimal version to be %s, got %s", min, gc.Minimal())
	}

	if gc.Current() != cur {
		t.Errorf("expected current version to be %s, got %s", cur, gc.Current())
	}

	if gc.Latest() != latest {
		t.Errorf("expected latest version to be %s, got %s", latest, gc.Latest())
	}
}

func TestMain(m *testing.M) {
	// disable output of log during testing to not pollute test output
	logger.SetUp(false)
	os.Exit(m.Run())
}

func TestGitHubCheckerImplementsChecker(t *testing.T) {
	var i interface{} = new(version.GitHubChecker)
	if _, ok := i.(version.Checker); !ok {
		t.Fatalf("expected %T to implement version.Checker", i)
	}
}

func TestNewGitHubCheckerWorksWithValidVersions(t *testing.T) {
	testCases := []versionsCaseSpec{
		{
			desc:   "With non-prefixed, v-prefixed, and suffixed versions",
			min:    "1.1.9",
			cur:    "v2.0.0",
			latest: "v2.3.4-test.1",
		},
		{
			desc:   "With unset minimal required version",
			min:    "",
			cur:    "v2.2.1",
			latest: "v2.5.1",
		},
		{
			desc:      "With only major or major+minor versions set",
			min:       "1",
			cur:       "v2.2",
			latest:    "v3",
			expMin:    "1.0.0",
			expCur:    "2.2.0",
			expLatest: "3.0.0",
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			gc, err := version.NewGithubChecker(newGitHubMock(tC.latest), fakeOrg, fakeRepo, tC.min, tC.cur)
			if err != nil {
				t.Fatalf("expected nil error, got %s", err.Error())
			}

			verifyExpectedVersions(t, gc, tC)
		})
	}
}

func TestNewGitHubCheckerFailsWithInvalidVersions(t *testing.T) {
	testCases := []versionsCaseSpec{
		{
			desc:   "Invalid minimal required version",
			min:    "banana",
			cur:    "2.2.1",
			latest: "2.5.1",
		},
		{
			desc:   "Invalid current version",
			min:    "v1.0.0",
			cur:    "gold",
			latest: "v2.5.1",
		},
		{
			desc:   "Invalid current version (blank)",
			min:    "v1.0.0",
			cur:    "",
			latest: "v2.5.1",
		},
		{
			desc:   "Invalid latest version",
			min:    "v1.0.0",
			cur:    "v2.2.1",
			latest: "silver",
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			_, err := version.NewGithubChecker(newGitHubMock(tC.latest), fakeOrg, fakeRepo, tC.min, tC.cur)
			if err == nil {
				t.Errorf("expected error, got nil")
			}

		})
	}
}

func TestGithubCheckerCheck(t *testing.T) {
	testCases := []versionsCaseSpec{
		{
			desc:    "ReturnsLatestWhenCurrentEqualsLatest",
			min:     "2.1.0",
			cur:     "2.6.0",
			latest:  "2.6.0",
			vAssert: version.IsLatest,
		},
		{
			desc:    "ReturnsMustUpdateWhenCurrentIsBellowMinimal",
			min:     "2.1.0",
			cur:     "2.0.9",
			latest:  "2.6.0",
			vAssert: version.MustUpdate,
		},
		{
			desc:    "ReturnsMustUpdateWhenCurrentIsBellowMinimalEvenIfLatestIsUnknown",
			min:     "1.0.0",
			cur:     "0.4.0",
			latest:  "",
			vAssert: version.MustUpdate,
		},
		{
			desc:    "ReturnsCanUpdateWhenCurrentIsBehindOnMajor",
			min:     "2.0.0",
			cur:     "2.5.0",
			latest:  "3.0.0",
			vAssert: version.CanUpdate,
		},
		{
			desc:    "ReturnsCanUpdateWhenCurrentIsBehindOnMinor",
			min:     "2.0.0",
			cur:     "2.0.9",
			latest:  "2.6.0",
			vAssert: version.CanUpdate,
		},
		{
			desc:    "ReturnsCanUpdateWhenCurrentIsBehindOnPatch",
			min:     "2.0.0",
			cur:     "2.6.0",
			latest:  "2.6.1",
			vAssert: version.CanUpdate,
		},
		{
			desc:    "ReturnsIsBeyondWhenCurrentIsAboveLatest",
			min:     "2.1.0",
			cur:     "2.6.1",
			latest:  "2.6.0",
			vAssert: version.IsBeyond,
		},
		{
			desc:    "ReturnsIsLatestWhenAllVersionsAreEqual",
			min:     "2.6.0",
			cur:     "2.6.0",
			latest:  "2.6.0",
			vAssert: version.IsLatest,
		},
		{
			desc:    "ReturnsCanUpdateWhenCurrentIsEqualToMinimal",
			min:     "2.6.0",
			cur:     "2.6.0",
			latest:  "2.6.1",
			vAssert: version.CanUpdate,
		},
		{
			desc:    "TreatsSuffixedVersionsAsOlderThanNonSuffixedVersions",
			min:     "2.1.0",
			cur:     "2.6.0-blah",
			latest:  "2.6.0",
			vAssert: version.CanUpdate,
		},
		{
			desc:    "CanCheckIfLatestWhenMinimalReqIsEmpty(VersionIsLatest)",
			min:     "",
			cur:     "2.6.0",
			latest:  "2.6.0",
			vAssert: version.IsLatest,
		},
		{
			desc:    "CanCheckIfLatestWhenMinimalReqIsEmpty(VersionCanUpdate)",
			min:     "",
			cur:     "0.0.1",
			latest:  "2.6.0",
			vAssert: version.CanUpdate,
		},
		{
			desc:    "ReturnsUnknownWhenLatestIsUnknownAndVersionIsAboveMinimal",
			min:     "1.0.0",
			cur:     "1.4.0",
			latest:  "",
			vAssert: version.IsUnknown,
		},
		{
			desc:    "ReturnsUnknownWhenMinAndLatestAreUnset",
			min:     "",
			cur:     "0.0.1",
			latest:  "",
			vAssert: version.IsUnknown,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			var ghErr *github.ErrorResponse

			gc, err := version.NewGithubChecker(newGitHubMock(tC.latest), fakeOrg, fakeRepo, tC.min, tC.cur)
			if err != nil {
				if errors.As(err, &ghErr) {
					t.Fatalf("expected nil error, got %s", ghErr.Message)
				} else {
					t.Fatalf("expected nil error, got %s", err)
				}
			}

			ans := gc.Check()
			if ans != tC.vAssert {
				t.Errorf("expected '%d/%s', got '%d/%s'", tC.vAssert, assertStr(tC.vAssert), ans, assertStr(ans))
			}

		})
	}
}

func TestDownloadLatest(t *testing.T) {
	var ghErr *github.ErrorResponse

	gc, err := version.NewGithubChecker(newGitHubMock("v3"), fakeOrg, fakeRepo, "v1", "v2")
	if err != nil {
		if errors.As(err, &ghErr) {
			t.Fatalf("expected nil error, got %s", ghErr.Message)
		} else {
			t.Fatalf("expected nil error, got %s", err)
		}
	}

	filename, err := gc.DownloadLatest()
	if err != nil {
		t.Fatalf("expected nil error on download, got %s", err)
	}

	fd, err := os.Open(filename)
	if err != nil {
		t.Fatalf("error opening downloaded file: %s", err)
	}
	defer fd.Close()

	text, err := ioutil.ReadAll(fd)
	if err != nil {
		t.Fatalf("error reading downloaded file: %s", err)
	}

	// For some reason the mock returns the fake content wrapped in "",
	// so we use a contains match. That's good enough for this.
	if !strings.Contains(string(text), fakeAssetContent) {
		t.Errorf("== expected file content:\n%s\n== got:\n%s\n", fakeAssetContent, string(text))
	}

}
