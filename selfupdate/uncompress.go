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

	// Create a new tar reader
	tarReader := tar.NewReader(gzipReader)

	// Iterate through the files in the tar archive
	header, err := tarReader.Next()
	if err == io.EOF {
		return nil, closer, errors.New("in untarGzSingleFileReader: empty tar file")
	}
	if err != nil {
		return nil, closer, err
	}

	// Print the file name
	fmt.Println(header.Name)
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

	if len(zipReader.File) != 1 {
		f.Close()
		return nil, nil, fmt.Errorf("in unzipSingleFileReader: zip file should contain exactly 1 file. It contains %d", len(zipReader.File))
	}

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
