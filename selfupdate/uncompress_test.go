package selfupdate_test

import (
	"io/ioutil"
	"testing"

	"github.com/dgmorales/go-cli-selfupdate/selfupdate"
)

const (
	fakeAssetContent = "yeah-of-course-i-was-downloaded\n"
)

func TestUncompress(t *testing.T) {
	testCases := []struct {
		desc       string
		filename   string
		shouldFail bool
	}{
		{
			desc:       "WorksWithTarGZFiles",
			filename:   "testdata/test.tar.gz",
			shouldFail: false,
		},
		{
			desc:       "WorksWithTGZFiles",
			filename:   "testdata/test.tgz",
			shouldFail: false,
		},
		{
			desc:       "WorksWithZipFiles",
			filename:   "testdata/test.zip",
			shouldFail: false,
		},
		{
			desc:       "DoesNotWorkWithOtherExtensions",
			filename:   "testdata/test.tar.bz2", // it exists on testdata
			shouldFail: true,
		},
		{
			desc:       "ErrorsWhenFileDoesNotExist",
			filename:   "testdata/doesnotexist.tar.gz",
			shouldFail: true,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {

			fd, closer, err := selfupdate.Uncompress(tC.filename)
			if err != nil {
				if tC.shouldFail {
					return
				}
				t.Fatalf("expected nil error got %s", err)
			}

			if tC.shouldFail {
				t.Fatal("expected error, got nil error")
			}

			data, err := ioutil.ReadAll(fd)
			if err != nil {
				t.Fatalf("got error reading uncompressed file: %s", err)

			}
			defer closer()

			if string(data) != fakeAssetContent {
				t.Errorf("== expected file content:\n%s\n== got:\n%s\n", fakeAssetContent, string(data))
			}

		})
	}
}
