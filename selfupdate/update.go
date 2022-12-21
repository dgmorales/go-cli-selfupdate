package selfupdate

import (
	minioSelfUpdate "github.com/minio/selfupdate"
)

func Apply(filename string) error {

	reader, closer, err := Uncompress(filename)
	if err != nil {
		if closer != nil {
			closer()
		}
		return err
	}
	defer closer()

	err = minioSelfUpdate.Apply(reader, minioSelfUpdate.Options{})
	if err != nil {
		return err
	}

	return nil
}
