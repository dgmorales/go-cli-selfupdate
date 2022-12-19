package selfupdate

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// untarGzSingleFileReader uncompress the first file of a tar.gz file (expected to be the only one in the archive)
//
// It returns an io.Reader for consuming the content, and a closer function to cleanup after read.
// WARNING: **The closer function may be nil** when errors were found opening the files, caller should
// test for that before calling it. If returned err is nil, close won't be nil.
func untarGzSingleFileReader(filename string) (r io.Reader, closer func(), err error) {

	f, err := os.Open(filename)
	if err != nil {
		return nil, nil, err
	}

	gzipReader, err := gzip.NewReader(f)
	if err != nil {
		f.Close()
		return nil, nil, err
	}

	closer = func() {
		gzipReader.Close()
		f.Close()
	}

	tarReader := tar.NewReader(gzipReader)

	// Get the first file in the tar archive, and ignore the rest
	// we could improve this by making it sure there's only 1 file as expected, erroring
	// out if not.
	_, err = tarReader.Next()
	if err == io.EOF {
		return nil, closer, errors.New("in untarGzSingleFileReader: empty tar file")
	}
	if err != nil {
		return nil, closer, err
	}

	return tarReader, closer, nil
}

func unzipSingleFileReader(filename string) (r io.Reader, closer func(), err error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, nil, err
	}

	// zip.NewReader needs the file size
	zipStat, err := f.Stat()
	if err != nil {
		f.Close()
		return nil, nil, err
	}

	zipReader, err := zip.NewReader(f, zipStat.Size())
	if err != nil {
		f.Close()
		return nil, nil, err
	}

	// We ensure there's only one file in the archive, as expected
	if len(zipReader.File) != 1 {
		f.Close()
		return nil, nil, fmt.Errorf("in unzipSingleFileReader: zip file should contain exactly 1 file. It contains %d", len(zipReader.File))
	}

	// Get the first file in the zip archive, and ignore the rest
	zippedFile, err := zipReader.File[0].Open()
	if err != nil {
		f.Close()
		return nil, nil, err
	}

	closer = func() {
		zippedFile.Close()
		f.Close()
	}

	return zippedFile, closer, nil
}

func Uncompress(filename string) (io.Reader, func(), error) {
	fn := strings.ToLower(filename)

	// using HasSuffix instead of a switch on filepath.Ext() because Ext("a.tar.gz") gives
	// only gz, and not tar.gz (correctly, of course)
	if strings.HasSuffix(fn, ".tar.gz") || strings.HasSuffix(fn, ".tgz") {

		r, closer, err := untarGzSingleFileReader(filename)
		if err != nil {
			return nil, closer, err
		}
		return r, closer, nil

	} else if strings.HasSuffix(fn, ".zip") {

		r, closer, err := unzipSingleFileReader(filename)
		if err != nil {
			return nil, closer, err
		}
		return r, closer, nil

	}

	return nil, nil, fmt.Errorf("unknown file format: %s", filepath.Ext(filename))
}
